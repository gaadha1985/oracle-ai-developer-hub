---
name: langchain-oracledb-helper
description: Scaffold a langchain-oracledb store layer — multi-collection OracleVS wrapper, metadata-as-string monkeypatch, embedder-dim assertion, OracleChatHistory subclass (langchain-oracledb does not ship one). Use when a project needs Oracle as its LangChain vector store and chat-history backend.
inputs:
  - target_dir: project root (must already have a working DB — invoke oracle-aidb-docker-setup first if not)
  - package_slug: snake_case Python package name (e.g. "pdf_chat", "nl2sql_agent")
  - embedder: one of "minilm-py" (384 dim, sentence-transformers), "in-db-onnx" (384 dim, registered MY_MINILM_V1), "oci-cohere" (1024 dim, opt-in alt)
  - collections: list[str] — collection names like ["DOCUMENTS", "CONVERSATIONS"]. Always include CONVERSATIONS if the project has chat.
  - has_chat_history: bool — if True, scaffold the OracleChatHistory class + DDL.
outputs:
  - target_dir/src/<package_slug>/store.py        (multi-collection wrapper)
  - target_dir/src/<package_slug>/_monkeypatch.py (metadata-as-string fix)
  - target_dir/src/<package_slug>/history.py      (only if has_chat_history)
  - target_dir/migrations/001_chat_history.sql    (only if has_chat_history)
---

You write the Oracle data-layer modules. You do not write app code, chains, or UI.

## Step 0 — References (mandatory)

- `shared/references/langchain-oracledb.md` — load-bearing.
- `shared/snippets/metadata_monkeypatch.py` — copy verbatim into `_monkeypatch.py`.
- `shared/snippets/oracle_chat_history.py` — copy verbatim into `history.py`.
- `shared/references/onnx-in-db-embeddings.md` — only if `embedder == "in-db-onnx"`.

## Step 1 — Validate inputs

- `target_dir/.env` exists and has `DB_DSN`, `DB_USER`, `DB_PASSWORD`. If not, stop — tell the user to run `oracle-aidb-docker-setup` first.
- `package_slug` matches `[a-z][a-z0-9_]*`. Reject otherwise.
- `collections` non-empty. Naming: `<PROJECT_PREFIX>_<KIND>` enforced — e.g. for slug `pdf_chat` and kind `DOCUMENTS`, the actual table name is `PDF_CHAT_DOCUMENTS`. Document this in the file's docstring.
- For `embedder == "in-db-onnx"`, confirm the user has a registered ONNX model name. If not, stop — point them at `shared/references/onnx-in-db-embeddings.md` step 3.

## Step 2 — Pick embedding dim

| Embedder | Dim | Module |
| --- | --- | --- |
| `minilm-py` | 384 | `langchain_huggingface.HuggingFaceEmbeddings(model_name="sentence-transformers/all-MiniLM-L6-v2")` — Python-side inference, no DB registration. Default for **beginner** tier. |
| `in-db-onnx` | 384 | Custom `Embeddings` subclass calling `VECTOR_EMBEDDING(MODEL_NAME USING :t AS data) FROM dual`. Default for **intermediate / advanced** tiers. Same MiniLM model as `minilm-py`, just registered inside Oracle. |
| `oci-cohere` | 1024 | `shared/snippets/oci_cohere_embeddings.py` (LangChain `Embeddings` subclass over `GenerativeAiInferenceClient`). Opt-in alternate for users who specifically want Cohere quality + multilingual. Different dim, so re-bootstrap required when swapping. |

Hard-code `EXPECTED_DIM` in `store.py`. `verify.py` (written by the tier skill) asserts `len(embedder.embed_query("dim check")) == EXPECTED_DIM` — runtime check, not import-time.

The whole point of defaulting to `minilm-py` in beginner and `in-db-onnx` in intermediate/advanced: **same model, same dim, same chunk-size sweet spot across tiers**. A corpus ingested at tier 1 can be re-ingested at tier 2 against the same embedding space — only the inference location changes.

## Step 3 — Write `_monkeypatch.py`

Copy `shared/snippets/metadata_monkeypatch.py` verbatim. Add a docstring at the top:

```python
"""
langchain-oracledb stores metadata as JSON strings, but its similarity_search
return path doesn't always parse them back. This monkeypatch makes the parsing
consistent. MUST be imported before any OracleVS instantiation in your app.

Source: shared/references/langchain-oracledb.md § "Metadata-as-string fix"
"""
```

Tier skills then add `from <package_slug>._monkeypatch import *` at the top of `store.py`, `app.py`, and any other entry point. **Don't skip this** — categorical metadata filtering breaks silently without it.

## Step 4 — Write `store.py`

Skeleton:

```python
"""
Oracle vector store layer for <package_slug>.

Owns:
- Multi-collection OracleVS wrapper (one logical collection per kind)
- Embedder factory (selected at scaffold time: <embedder>)
- Connection lifecycle (one shared connection, lazy)

Cites:
- shared/references/langchain-oracledb.md
- apps/agentic_rag/src/OraDBVectorStore.py:1-100
"""
from . import _monkeypatch  # noqa: F401  -- must be first

import os
import oracledb
from langchain_oracledb.vectorstores.oraclevs import OracleVS
# ... embedder import per choice ...

PROJECT_PREFIX = "<UPPER_PACKAGE_SLUG>"
EXPECTED_DIM = <384 | 768 | 1024>

_conn = None
_embedder = None

def get_connection() -> oracledb.Connection:
    global _conn
    if _conn is None or not _conn.ping():
        _conn = oracledb.connect(
            user=os.environ["DB_USER"],
            password=os.environ["DB_PASSWORD"],
            dsn=os.environ["DB_DSN"],
        )
    return _conn

def get_embedder():
    global _embedder
    if _embedder is None:
        _embedder = <embedder factory call>
    return _embedder

def get_store(kind: str) -> OracleVS:
    table = f"{PROJECT_PREFIX}_{kind.upper()}"
    return OracleVS(
        client=get_connection(),
        embedding_function=get_embedder(),
        table_name=table,
        distance_strategy="COSINE",
    )

def bootstrap() -> None:
    """Idempotent: ensure each collection table exists by inserting a
    no-op then immediately deleting it. Cheaper than checking USER_TABLES."""
    for kind in <collections list>:
        store = get_store(kind)
        ids = store.add_texts(["__bootstrap__"], metadatas=[{"_skip": True}])
        store.delete(ids)
```

Replace placeholders with concrete values from inputs. The bootstrap dance is the load-bearing trick — `OracleVS.from_texts` creates the table on first call; subsequent calls are no-ops.

## Step 5 — Write `history.py` (if `has_chat_history`)

Copy `shared/snippets/oracle_chat_history.py` verbatim into `target_dir/src/<package_slug>/history.py`. Add the LangChain glue at the bottom:

```python
def get_history_factory(conn):
    """Return a callable suitable for RunnableWithMessageHistory."""
    def _factory(session_id: str):
        return OracleChatHistory(conn, session_id)
    return _factory
```

Then `migrations/001_chat_history.sql`:

```sql
CREATE TABLE IF NOT EXISTS chat_history (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id VARCHAR2(128) NOT NULL,
    role VARCHAR2(16) NOT NULL,
    content CLOB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chat_history_session ON chat_history(session_id, created_at);
```

The tier skill is responsible for running this migration during bootstrap. Document that requirement in the docstring at the top of `history.py`.

## Step 6 — Smoke

After writing, run from `target_dir`:

```python
from <package_slug>.store import bootstrap, get_store, get_embedder, EXPECTED_DIM
bootstrap()
v = get_embedder().embed_query("dim check")
assert len(v) == EXPECTED_DIM, f"dim mismatch: got {len(v)} expected {EXPECTED_DIM}"
print("langchain-oracledb-helper: OK")
```

If dim mismatches: drop the offending tables (`DROP TABLE <PREFIX>_<KIND> CASCADE CONSTRAINTS`), fix the embedder, re-bootstrap. Mismatches happen most when the user changes embedder mid-project.

## Stop conditions

- `langchain-oracledb` not in `pyproject.toml`. Tell the user to add `langchain-oracledb>=0.1` and stop.
- Embedder choice doesn't match what the project already uses (existing tables have a different dim). Surface mismatch, don't silently re-bootstrap.
- `has_chat_history=True` but no `chat_history` table after migration. Stop — the migration didn't run.

## What you must NOT do

- Don't skip `_monkeypatch.py`. Filtered retrievals break silently.
- Don't `from langchain_oracledb.chat_message_histories import ...` — it doesn't exist.
- Don't write SQL DDL for vector tables manually. `OracleVS.from_texts` (via the bootstrap dance) handles it.
- Don't pin a different distance strategy unless the user asks. COSINE is the default and matches the rest of the skill set.

## Final report

```
langchain-oracledb-helper: OK
  store:        target_dir/src/<package_slug>/store.py
  monkeypatch:  target_dir/src/<package_slug>/_monkeypatch.py
  history:      target_dir/src/<package_slug>/history.py        (if scaffolded)
  migrations:   target_dir/migrations/001_chat_history.sql      (if scaffolded)
  embedder:     <minilm-py|in-db-onnx|oci-cohere> (dim=<384|384|1024>)
  collections:  <list>
  next:         hand off to the tier skill — it writes app code that imports `store`.
```
