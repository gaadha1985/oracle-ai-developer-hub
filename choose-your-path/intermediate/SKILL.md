---
name: choose-your-path-intermediate
description: Scaffold a real RAG chatbot on Oracle 26ai Free + langchain-oracledb (multi-collection, persistent chat history, hybrid retrieval) + OCI Generative AI or Ollama + Gradio UI. For users who've built RAG before and want to rebuild it on Oracle.
inputs:
  - target_dir: where to scaffold (default ~/git/personal/<slug>)
  - topic: optional; one of intermediate/project-ideas.md, or a free-text pitch
---

The user picked the **intermediate** path. They've built RAG before. Don't over-explain; do drop them on the rails of `langchain-oracledb` + OCI GenAI + Oracle 26ai Free.

## Step 0 — Read these references first

- `shared/references/sources.md`
- `shared/references/oracle-26ai-free-docker.md`
- `shared/references/langchain-oracledb.md` ← the centerpiece
- `shared/references/oci-genai-openai.md`
- `shared/references/ollama-local.md` (in case the user picks Ollama or wants a fallback embedder)
- `shared/references/oracledb-python.md`
- `shared/references/ai-vector-search.md`
- `shared/references/hybrid-search.md`
- `shared/references/exemplars.md`
- `intermediate/project-ideas.md`

## Step 1 — Interview

Run `shared/interview.md`. For intermediate specifically:

- Q3 (DB target) — default to "Local Docker" but allow "already-running container" if the user says so.
- Q4 (Inference) — surface all three options. Defaults the skill should suggest:
  - **OCI GenAI (Grok 4 / Cohere embeddings)** — recommended.
  - Ollama (local) — fallback when the user lacks an OCI tenancy.
  - BYO OpenAI-compat — only if asked.
- Q5 (Topic) — one of the five ideas; map free-text pitches.
- Q6 (Notebook) — default **yes**.

When the user picks OCI:
- Confirm region. If they pick `grok-4` and aren't in `us-chicago-1`, warn and offer Cohere or Llama as a same-region alternative (per `oci-genai-openai.md`).
- Check `~/.oci/config` exists; if not, point them at `oci setup config` and stop.
- Confirm `OCI_COMPARTMENT_ID` is available (env or interview answer).

Print confirmation block. Wait for `y`.

## Step 2 — Resolve choices

Spec dict:

| Variable | Source |
| --- | --- |
| `project_slug` | derived from topic |
| `embedder_pkg` / `embedder_init` / `embedding_dim` | OCI: Cohere wrapper / 1024; Ollama: `OllamaEmbeddings(model="nomic-embed-text")` / 768 |
| `llm_pkg` / `llm_init` | OCI: `ChatOpenAI` with OCI base_url; Ollama: `ChatOllama(model=...)` |
| `collections` | per-idea: PDF-RAG → one per file + `CONVERSATIONS`; Codebase → one + `CONVERSATIONS`; etc. |
| `ui_stack` | `gradio` |
| `notebook` | yes by default |
| `qwen_mitigations` | only if user picked an Ollama Qwen chat model |

Validate **embedding-dim consistency before writing any file** by calling `embedder.embed_query("dim check")` and checking the length matches `embedding_dim`. If mismatch: stop, ask the user.

## Step 3 — Scaffold

Create files in this order. Cite exemplars in headers.

1. `target_dir/` — refuse if non-empty.
2. `.gitignore` — `.env`, `__pycache__/`, `.venv/`, `*.pyc`, `data/` (where the user puts their PDFs / repos / etc.).
3. `docker-compose.yml` — from template.
4. `.env.example` + `.env` — keep both Ollama + OCI sections, comment out the unused one. Generate `ORACLE_PWD`.
5. `pyproject.toml` — from template, with deps:
   - Always: `oracledb>=2.4`, `langchain-oracledb>=0.1`, `langchain>=0.3`, `gradio>=4.0`, `python-dotenv>=1.0`, `pydantic>=2`.
   - OCI: `oci>=2.130`, `langchain-openai>=0.2`.
   - Ollama: `langchain-ollama>=0.2`.
   - Per-idea: PDF-RAG → `pypdf`, `unstructured`; Web → `trafilatura`, `httpx`; Codebase → `tree-sitter` (or `pygments` for v1 simpler chunking).
6. `src/<package>/__init__.py`
7. `src/<package>/store.py` — multi-collection wrapper.
   - Cite `apps/agentic_rag/src/OraDBVectorStore.py:1-100` at top.
   - Class `ProjectStore` exposing `add(kind, texts, metadatas)`, `search(kind, query, k, filter)`, `as_retriever(kind, **kwargs)`.
   - **Always include the metadata-as-string monkeypatch** (verbatim from `langchain-oracledb.md`).
   - Naming convention `<PROJECT_SLUG>_<KIND>` enforced.
8. `src/<package>/inference.py` — embedder + LLM init. One factory per backend (OCI / Ollama / BYO). The factory checks env vars on import and raises a clear error if unset.
   - For OCI embeddings: write the LangChain `Embeddings` subclass that wraps `GenerativeAiInferenceClient.embed_text` per `oci-genai-openai.md`. Filter empties, batch by 96, cache the client.
9. `src/<package>/history.py` — persistent chat history.
   - Use `OracleChatMessageHistory` from `langchain-oracledb`.
   - Wrap in `RunnableWithMessageHistory` exposed as `with_history(chain)`.
10. `src/<package>/retrieval.py` — hybrid retriever.
    - Default = `EnsembleRetriever([vector, BM25])` per `hybrid-search.md` Pattern A.
    - Add a `pure_vector_retriever()` for the user to swap in if BM25 memory is a concern.
11. `src/<package>/ingest.py` — per-idea ingestion script. Chunking strategy hard-coded per idea (PDFs by page, repos by file or symbol, web pages with trafilatura, slack by message-thread, markdown by H2).
12. `src/<package>/chains.py` — the RAG chain. LCEL, taking `retriever` and `llm`, returning `RunnableWithMessageHistory`.
13. `src/<package>/app.py` — Gradio UI. Single-page chat with file upload (where applicable). Hooks the chain to `gr.ChatInterface`.
14. `gradio_app.py` (root) — thin entrypoint: `from <package>.app import demo; demo.launch()`.
15. `verify.py` — from template; `inference_enabled = True`. Round-trip check uses the chosen embedder + LLM. The vector smoke uses a non-conflicting table (`CYP_VERIFY_SMOKE`).
16. `notebook.ipynb` — copy `shared/templates/notebook.template.ipynb` if present, otherwise generate a 6-cell notebook:
    1. Setup (load .env, smoke `verify`).
    2. Ingest a tiny corpus.
    3. One vector search.
    4. One retrieved-augmented question.
    5. Show conversation history persistence.
    6. "Now run `python gradio_app.py` to use the UI."
17. `README.md` — from template. Fill all placeholders. Include the screenshot slot for the Gradio UI.

## Step 4 — Verify

1. `docker compose up -d --wait`.
2. `python verify.py`. Expect `verify: OK (db, vector, inference)`.
3. **Then** run an end-to-end smoke:
   - For ideas 1, 2, 4, 5: ingest a tiny known corpus the skill includes (e.g. 3 lines of test text).
   - Ask one question through the chain.
   - Assert the answer references the known content.
4. On verify failure, follow the `shared/verify.md` recovery loop.

## Step 5 — Polish for sharing

1. README placeholders filled.
2. Open `gradio_app.py` once (background) so the user sees it boot to `http://localhost:7860`. Don't keep it running — just confirm it starts.
3. Final report:
   ```
   Done.
     project at: <target_dir>
     run with:   cd <target_dir> && python gradio_app.py
     verify:     OK
     notebook:   <target_dir>/notebook.ipynb
     next:       record a 30s demo of the Gradio UI, fill the "What I built" section, push.
   ```

## Stop conditions

- OCI selected but `~/.oci/config` missing — stop, point at `oci setup config`.
- User wants Grok 4 but isn't in `us-chicago-1` — stop unless they accept Cohere / Llama fallback.
- Embedder dim doesn't match what the embedder *says* it returns — stop, surface the mismatch.
- Verify fails 3 times.
- Topic doesn't fit any idea well.

## What you must NOT do

- Don't bypass the metadata-as-string monkeypatch. Filtered retrievals will silently break.
- Don't write raw `VECTOR_DISTANCE` SQL when `OracleVS.similarity_search` covers it. Hybrid search is the only place SQL is allowed at this tier (Pattern A in `hybrid-search.md` keeps it out of view anyway).
- Don't introduce non-Oracle vector stores anywhere — not even as "for comparison."
- Don't pin a model that doesn't exist in the user's region without warning.
- Don't ship without a notebook (intermediate default is yes — only skip if the user explicitly said no during interview).
- Don't claim done before verify is green AND the e2e smoke passes.
