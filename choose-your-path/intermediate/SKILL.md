---
name: choose-your-path-intermediate
description: Scaffold a Grok-4 tool-calling agent over an Oracle schema using langchain-oracledb + oracle-database-mcp-server + in-DB ONNX embeddings (registered MiniLM model, no external embedding API) + Open WebUI. For users who've built RAG before and want to rebuild it on the production-feeling Oracle stack.
inputs:
  - target_dir: where to scaffold (default = current working directory; ask if it isn't empty)
  - topic: optional; one of intermediate/project-ideas.md, or a free-text pitch
---

The user picked the **intermediate** path. They've built RAG and chatbots before. Your job is to introduce them to **two** new ideas at once: **(a)** an LLM agent that calls live SQL via `oracle-database-mcp-server`, and **(b)** embeddings that happen *inside the database* via a registered ONNX model. The stack is production-shaped: OCI GenAI Grok 4, in-DB ONNX, Open WebUI. No Ollama, no external embedding API.

## Step 0 — Read these references first

- `shared/references/sources.md`
- `shared/references/oracle-26ai-free-docker.md`
- `shared/references/langchain-oracledb.md`
- `shared/references/oci-genai-openai.md`  ← Pattern 1 SigV1 auth
- `shared/references/onnx-in-db-embeddings.md`  ← load-bearing for embeddings
- `shared/references/oracledb-python.md`
- `shared/references/ai-vector-search.md`
- `shared/references/hybrid-search.md` (idea 3 specifically)
- `shared/references/exemplars.md`
- `intermediate/project-ideas.md`
- `skills/oracle-aidb-docker-setup/SKILL.md`
- `skills/langchain-oracledb-helper/SKILL.md`
- `skills/oracle-mcp-server-helper/SKILL.md`

## Step 1 — Interview

Run `shared/interview.md`. For intermediate specifically:

- **Q3 (DB target)** — default to local Docker. Allow "already-running container" if user says so.
- **Q4 (Inference)** — *not optional at this tier*. **OCI GenAI** for the LLM (`xai.grok-4` via the OpenAI-compat bearer-token endpoint at `us-phoenix-1`). **In-DB ONNX** for embeddings. Confirm:
  - `OCI_GENAI_API_KEY` (a `sk-...` value) is set or about to be added to project `.env`. If absent, stop and ask the user to generate one in the OCI GenAI service console. **No `~/.oci/config` / no compartment OCID needed.**
  - Default endpoint is `https://inference.generativeai.us-phoenix-1.oci.oraclecloud.com`; override via `OCI_GENAI_BASE_URL`.
  - **In-DB ONNX model:** default = `sentence-transformers/all-MiniLM-L6-v2`, registered as `MY_MINILM_V1` (384 dim). The user does *not* need to download this themselves — the skill scaffolds via `onnx2oracle` CLI (one command).
- **Q5 (Topic)** — one of the three from `intermediate/project-ideas.md`. Map free-text pitches; default to idea 1 (NL2SQL).
- **Q6 (Notebook)** — default **yes**.
- **Q7 (intermediate-only) — sql_mode for MCP?** — `read_only` (default — covers all three idea shapes safely) or `read_write`. Idea 1 and idea 2 are read-only. Idea 3 can be either. Capture an explicit `y` if `read_write` selected.

Print confirmation block. Wait for `y`.

## Step 2 — Resolve choices

| Variable | Value |
| --- | --- |
| `project_slug` | derived from topic |
| `package_slug` | snake_case |
| `embedder` | `in-db-onnx` |
| `embedding_dim` | 384 |
| `onnx_model_local_id` | `sentence-transformers/all-MiniLM-L6-v2` |
| `onnx_model_db_name` | `MY_MINILM_V1` |
| `llm_model` | `grok-4` (or chosen fallback) |
| `oci_base_url` | `https://inference.generativeai.us-chicago-1.oci.oraclecloud.com/20231130/actions/openai` |
| `collections` | per-idea: idea 1 → `["CONVERSATIONS"]` only; idea 2 → `["SCHEMA_DOCS_DOCUMENTS", "CONVERSATIONS"]`; idea 3 → `["INVOICES_DOCS", "CONVERSATIONS"]` |
| `mcp_sql_mode` | `read_only` (default) |
| `mcp_allowed_tools` | per-idea (see below) |
| `notebook` | yes |

## Step 3 — Scaffold

Order matters: building-block skills first, then project code.

### 3a — Foundation via building-block skills

1. Refuse if `target_dir` is non-empty.
2. **Invoke `skills/oracle-aidb-docker-setup`.** Block until OK.
3. Append the **Open WebUI** service to the generated `docker-compose.yml` (same as beginner SKILL step 3a-3).
4. **Register the in-DB ONNX model via `onnx2oracle` CLI** *before* invoking the langchain helper, since the helper's dim assertion needs the model registered:
   - Add `onnx2oracle` to the project's `pyproject.toml` deps.
   - Install: `~/miniconda3/envs/<env>/bin/pip install onnx2oracle`.
   - Run: `onnx2oracle load all-MiniLM-L6-v2 --name MY_MINILM_V1 --dsn "$DB_USER/$DB_PASSWORD@$DB_DSN" --force` — outputs `MY_MINILM_V1` registered in the DB.
   - Smoke: `SELECT VECTOR_EMBEDDING(MY_MINILM_V1 USING 'test' AS data) FROM dual` returns a 384-vector. If not, stop and surface the loader error (most common: missing GRANTs — see `shared/references/onnx-in-db-embeddings.md` "Required GRANTs").
   - The required GRANTs (`CREATE MINING MODEL`, `EXECUTE ON SYS.DBMS_VECTOR`) are issued by `oracle-aidb-docker-setup` Step 6. If they're missing, the docker-setup didn't run fully — fix that first.
5. **Invoke `skills/langchain-oracledb-helper`.** Pass `target_dir`, `package_slug`, `embedder=in-db-onnx` (the helper writes the `InDBEmbeddings` subclass), `collections=...`, `has_chat_history=True`. Block until OK.
6. **Invoke `skills/oracle-mcp-server-helper`.** Pass `target_dir`, `package_slug`, `sql_mode=...`, `allowed_tools=...`. Block until OK. Tool list per idea:
   - Idea 1: `[list_tables, describe_table, run_sql]`
   - Idea 2: `[list_tables, describe_table, describe_schema, run_sql, vector_search]`
   - Idea 3: `[run_sql, vector_search]` (the agent doesn't need to discover tables — they're known)

### 3b — Per-idea seeding

7. **Idea 1 (NL2SQL with seeded fake data).** Generate `migrations/100_seed_dummy.sql` — 10 tables (customers, orders, products, employees, suppliers, invoices, payments, regions, categories, returns), populated via `Faker` from `scripts/seed_faker.py`. ~50K rows. Run during bootstrap.
8. **Idea 2 (Schema doc Q&A).** Reuse the seed schema from idea 1 if the user wants; otherwise expect them to point at their real schema.
9. **Idea 3 (Hybrid retrieval).** Generate `INVOICE_PDFS/` folder via `scripts/seed_invoice_pdfs.py` (uses `reportlab` to make 20 fake invoice PDFs). Run `ingest.py` once at bootstrap to embed them into `INVOICES_DOCS` via in-DB embeddings. Plus the seed schema from idea 1.

### 3c — Project-specific code (the only files this skill writes itself)

10. `target_dir/.gitignore` — extend with `data/`, `INVOICE_PDFS/`, `*.onnx`, `scripts/__pycache__/`.
11. `target_dir/pyproject.toml` — extend deps:
    - Always: `fastapi>=0.110`, `uvicorn[standard]>=0.27`, `langchain-core>=0.3`, `langchain-community>=0.3`, `openai>=1.40`, `onnx2oracle`, `Faker>=24`, `python-multipart`.
    - Idea 1: + (no extras).
    - Idea 2: + (no extras).
    - Idea 3: + `reportlab>=4`, `pypdf>=4`.
    - **Do NOT add** `oci-openai`, `openai`, `oracle-database-mcp-server`, or hand-rolled ONNX deps. Those are friction P0-1 / P0-2 / P0-4 — superseded.
12. `src/<package_slug>/inference.py` — copy `shared/snippets/oci_chat_factory.py` verbatim. It uses `oci.generative_ai_inference.GenerativeAiInferenceClient` directly (the OpenAI-compat path is unstable; see friction P0-2). Model id is `xai.grok-4`.
13. **Per-idea agent module:**
    - **Idea 1** → `src/<package_slug>/agent.py`:
      ```python
      tools = get_tools()  # from tool_registry.py
      llm = make_llm()  # Grok 4 via OCI
      agent = create_tool_calling_agent(llm.bind_tools(tools), tools, prompt)
      executor = AgentExecutor(agent=agent, tools=tools)
      with_history = RunnableWithMessageHistory(executor, get_history_factory(...), ...)
      ```
      The prompt teaches Grok to: list_tables → describe_table → emit run_sql; return both the answer and the SQL it ran.
    - **Idea 2** → `src/<package_slug>/generate.py` (one-shot script that walks the schema and INSERTs rows into `SCHEMA_DOCS_DOCUMENTS` with embeddings via `VECTOR_EMBEDDING(MY_MINILM_V1 USING :description)`) + `src/<package_slug>/agent.py` (RAG over the generated docs via `vector_search` MCP tool).
    - **Idea 3** → `src/<package_slug>/agent.py` with a system prompt that explicitly teaches the agent the two-modality choice (vector for "find similar invoices to this PDF", run_sql for "sum unpaid amounts", both for "find unpaid invoices similar to X").
14. `src/<package_slug>/adapter.py` — FastAPI `/v1/chat/completions` wrapping the agent (same shape as beginner; differences: handles tool-call streaming events from the agent executor, surfaces them as OpenAI-compatible "function_call" deltas).

    **SQLcl-tee logging (folded in by default at this tier — friction-pass decision).** Wrap the `run_sql` BaseTool with `shared/snippets/sqlcl_tee.py` so every SQL the agent emits gets teed through SQLcl into `<target>/logs/sqlcl_<ts>.log`. The wrapper appends `[sqlcl_log: <path>]` to the streamed response. Setup:
    - In `src/<package_slug>/tool_registry.py`, import `from shared.snippets.sqlcl_tee import wrap_with_sqlcl_tee` and wrap the `run_sql` tool that comes back from `mcp_client.list_tools()`.
    - Document SQLcl install in the project's README (link to `shared/references/sqlcl-tee.md`).
    - Why MCP+SQLcl: MCP shows the SQL the agent emits; SQLcl shows what the DB actually did (rows, errors, plan). Together you can debug an agent turn end-to-end.
15. `verify.py` — fill template:
    - Round-trip: `len(get_embedder().embed_query("dim check")) == 384`.
    - Smoke: query the registered ONNX model directly via SQL.
    - Smoke: list MCP tools — assert at least the per-idea allowed list is present.
    - Smoke: a single chain call asking a simple question of the seeded data.
16. `notebook.ipynb` — 8 cells:
    1. Setup (load `.env`, smoke `verify`).
    2. Show the registered ONNX model (`SELECT * FROM USER_MINING_MODELS WHERE MODEL_NAME='MY_MINILM_V1'`).
    3. Show MCP tools list.
    4. One direct `vector_search` MCP call.
    5. One `run_sql` MCP call.
    6. One full agent turn (idea-specific question).
    7. Show the chat history table populated.
    8. "Now run `python -m <pkg>.adapter` and open `http://localhost:3000`."
17. `README.md` — fill placeholders. "Why Oracle" paragraph names: in-DB ONNX embeddings, AI Vector Search, oracle-database-mcp-server, JSON Duality (idea 3), persistent chat history. **Include the "Why in-DB embeddings?" callout from `intermediate/project-ideas.md`** verbatim — it's the load-bearing pitch.

## Step 4 — Verify

1. DB is up (skill 1).
2. ONNX model registered (step 3a-4).
3. From `target_dir`: `python -m pip install -e .`.
4. `python verify.py`. Expect `verify: OK (db, vector, inference, mcp)`.
5. Run notebook end-to-end: `jupyter nbconvert --to notebook --execute notebook.ipynb`. Must complete clean.
6. Bring Open WebUI up. Boot adapter, hit `/v1/models`, kill it. Don't keep it running.
7. On any failure, follow `shared/verify.md` recovery loop, max 3 retries.

## Step 5 — Polish for sharing

1. README placeholders filled.
2. `docs/` — note: drop a 60s demo GIF showing tool-call traces.
3. Final report:
   ```
   Done.
     project at:    <target_dir>
     features used: in-DB ONNX (MY_MINILM_V1), oracle-database-mcp-server, OracleVS, OracleChatHistory
     run with:      cd <target_dir>
                    docker compose up -d
                    python -m <pkg>.adapter   # blocks; Open WebUI on :3000
     verify:        OK
     notebook:      <target_dir>/notebook.ipynb (executed clean)
     ui:            http://localhost:3000
     next:          record a 60s tool-call demo, push to GitHub.
   ```

## Stop conditions

- OCI selected but `~/.oci/config` missing — stop, point at `oci setup config`.
- ONNX export fails (BertTokenizer model only — SentencePiece will fail). Surface error, stop.
- ONNX model registers but its `embed_query` returns dim ≠ 384. Drop the model, surface error.
- MCP server fails to initialize within 30s — stop and surface stderr.
- `sql_mode=read_write` without explicit user `y`.
- Verify fails 3 times.

## What you must NOT do

- Don't bypass the metadata-as-string monkeypatch (`langchain-oracledb-helper` includes it; just don't remove the import).
- Don't write raw `VECTOR_DISTANCE` SQL when `OracleVS.similarity_search` covers it.
- Don't introduce non-Oracle vector stores anywhere.
- Don't introduce Ollama as a fallback. OCI GenAI only at this tier.
- Don't introduce Cohere embeddings as a fallback. In-DB ONNX is the contract — the whole pedagogical point is "no external embedder."
- Don't pin a model that doesn't exist in the user's region without warning.
- Don't ship without the executed notebook.
- Don't claim done before verify is green AND the notebook runs clean AND the adapter boots.
