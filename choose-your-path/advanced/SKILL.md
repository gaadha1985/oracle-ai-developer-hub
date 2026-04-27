---
name: choose-your-path-advanced
description: Scaffold an agent system where Oracle AI DB is the *only* state store. Multi-feature (vector + memory + at least one of JSON Duality / property graph / ONNX in-DB) on Oracle 26ai Free + OCI GenAI (or Ollama) + LangChain + Gradio. For users who've built agents before and want a real DB-as-only-store demo.
inputs:
  - target_dir: where to scaffold (default ~/git/personal/<slug>)
  - topic: optional; one of advanced/project-ideas.md, or a free-text pitch within the constraint
---

The user picked the **advanced** path. The defining constraint is **Oracle AI DB is the only state store** — no Redis, no Postgres, no SQLite, no Chroma/FAISS/Qdrant/Pinecone, no JSON / pickle on disk for runtime state. The skill enforces this in scaffolding *and* in the verify step.

## Step 0 — Read these references first

- All of `shared/references/`. Yes, all of them. The advanced path can touch every feature.
- Specifically required: `langchain-oracledb.md`, `oci-genai-openai.md`, `ai-vector-search.md`, `hybrid-search.md`, `json-duality.md`, `property-graph.md`, `onnx-in-db-embeddings.md`, `visual-oracledb-features.md`.
- `advanced/project-ideas.md`.

## Step 1 — Interview

Run `shared/interview.md` plus the advanced-only questions below.

- Q4 (Inference) — recommend OCI GenAI for the "polished demo" feel; allow Ollama for the air-gapped variant. Discourage BYO unless the user has a reason.
- Q6 (Notebook) — yes, **mandatory**. Reject "no" — advanced is where notebook payoff lives.
- **Q7 (advanced-only) — Which features?** From `visual-oracledb-features.md`. The user must pick at least:
  - Vector search (always — it's how everything talks).
  - Agent memory tables (always — it's the "DB-as-only-store" core).
  - **At least one of**: JSON Duality, Property Graph, ONNX in-DB embeddings.
- **Q8 (advanced-only) — Demo focus?** "Polished UI demo" / "Notebook deep-dive" / "Both" — affects how much Gradio polish vs notebook narrative the skill produces.

For idea 4 ("Translate-this-toy-agent"): also ask for the source repo path / URL. The skill reads it before continuing.

## Step 2 — Resolve choices

Spec dict has the standard fields plus:

| Variable | Source |
| --- | --- |
| `feature_set` | answers to Q7; subset of {`vector`, `hybrid`, `memory`, `json_duality`, `property_graph`, `onnx_in_db`} |
| `memory_types` | always all 6 (conversational, KB, workflow, toolbox, entity, summary) |
| `forbidden_imports` | hardcoded — verify greps for these |
| `notebook_focus` | `polished_ui` / `deep_dive` / `both` from Q8 |

Embedding-dim consistency check — same as intermediate, but additionally validate against the column dim if ONNX in-DB is selected (must match the registered ONNX model's output, e.g. 384 for MiniLM-L6-v2).

## Step 3 — Scaffold

Order matters; later modules depend on earlier ones.

### 3a — Foundation (always)

1. `target_dir/` — refuse if non-empty.
2. `.gitignore` — `.env`, `__pycache__/`, `.venv/`, `*.pyc`, `data/`, `notebook-checkpoints/`.
3. `docker-compose.yml` — from template.
4. `.env.example` + `.env` — generate ORACLE_PWD; OCI section enabled by default.
5. `pyproject.toml` — deps:
   - Always: `oracledb>=2.4`, `langchain-oracledb>=0.1`, `langchain>=0.3`, `gradio>=4.0`, `python-dotenv>=1.0`, `pydantic>=2`.
   - OCI: `oci>=2.130`, `langchain-openai>=0.2`.
   - ONNX in-DB (if selected): `optimum[onnxruntime]`, `onnxruntime>=1.18`, `onnxruntime-extensions`, `transformers`, `torch`.
6. `migrations/` — numbered `.sql` files. Run them in order on first boot.
   - `001_core.sql` — entity/document tables, vector columns.
   - `002_memory.sql` — 6 memory tables from `apps/finance-ai-agent-demo/backend/memory/manager.py`. Cite that file at the top.
   - `003_<feature>.sql` per selected feature:
     - `003_json_duality.sql` — schema + view from `json-duality.md`.
     - `003_property_graph.sql` — entity + edge tables.
     - `003_onnx_dir.sql` — `CREATE DIRECTORY` + grants for ONNX model loading.

### 3b — Inference and storage

7. `src/<package>/inference.py` — embedder + LLM factories per `oci-genai-openai.md` and `langchain-oracledb.md`. If ONNX in-DB selected, also expose `InDBEmbeddings` from `onnx-in-db-embeddings.md`.
8. `src/<package>/store.py` — `OracleVS` multi-collection wrapper. Metadata-as-string monkeypatch. Cite `apps/agentic_rag/src/OraDBVectorStore.py:1-100`.

### 3c — Memory layer (always)

9. `src/<package>/memory/manager.py` — port the 6-memory-type pattern from `apps/finance-ai-agent-demo/backend/memory/manager.py:1-100`. Each memory type gets `write_*`, `read_*`, `search_*` methods. All vector-searchable where it makes sense.
10. `src/<package>/memory/entity.py` — entity-memory pattern from `sprawl_manager.py` (people/places/topics with embeddings).
11. `src/<package>/memory/event_log.py` — tool-execution + reasoning log from `apps/agentic_rag/src/OraDBEventLogger.py`.

### 3d — Per-feature modules (one per Q7 selection)

- **JSON Duality** → `src/<package>/duality.py`. Write-through JSON helpers + relational read helpers. Cite `~/git/work/demoapp/api/app/routers/json_views.py:1-80`.
- **Property Graph** → `src/<package>/graph.py`. Entity/edge writers + Python BFS for n-hop. Cite `~/git/work/demoapp/api/app/routers/graph.py:1-80`.
- **ONNX in-DB** → `src/<package>/onnx_loader.py` — wraps the pipeline + loader from `onnx2oracle/`. Skill copies `pipeline.py` and `loader.py` from `~/git/personal/onnx2oracle/src/onnx2oracle/` into the user's project (with attribution comment).

### 3e — Agent loop and UI

12. `src/<package>/tools.py` — tool registry. Each tool registered, every call logged via `event_log`. No filesystem state.
13. `src/<package>/agent.py` — main loop: retrieve (vector + optional graph hop) → reason (LLM) → act (tool call) → reflect → write to memory. Pure Python, no LangGraph (keeps the dependency surface honest at this scope).
14. `src/<package>/app.py` — Gradio with multiple tabs:
    - **Chat** — main agent interface.
    - **Memory** — browse the 6 memory types.
    - **Per-feature tabs** — for each Q7 selection (Duality dashboard, Graph viewer, etc).
15. `gradio_app.py` — entrypoint.

### 3f — Tests, notebook, README

16. `verify.py` — from template, with the **advanced extension** from `shared/verify.md`:
    - Round-trip each of the 6 memory tables.
    - Grep the project for `forbidden_imports` (`redis`, `psycopg`, `psycopg2`, `sqlite3`, `chromadb`, `qdrant_client`, `pinecone`, `faiss`). Fail if any found.
    - For each Q7 feature, run a feature-specific smoke (e.g. JSON Duality → write-then-read-relationally; graph → 2-hop BFS).
17. `notebook.ipynb` — mandatory. Structure depends on Q8:
    - `polished_ui` focus → 8 cells, last cell launches Gradio.
    - `deep_dive` focus → 12-15 cells, walks every feature with prose.
    - `both` → 12-15 cells, last cell launches Gradio.
18. `README.md` — from template. The "Why Oracle" paragraph is assembled from `visual-oracledb-features.md` entries matching the user's `feature_set`. Include a "DB-as-only-store proof" callout pointing at the verify forbidden-import grep.

## Step 4 — Verify

1. `docker compose up -d --wait`.
2. Apply `migrations/*.sql` in order. Stop on first error.
3. `python verify.py` — must print `verify: OK (db, vector, inference, memory, <features>)`.
4. **Advanced extra**: `verify.py` greps for forbidden imports. If found, fail and tell the user which file violates the rule.
5. Run the notebook end-to-end (`jupyter nbconvert --to notebook --execute notebook.ipynb`). It must complete without errors.
6. On any failure, follow `shared/verify.md` recovery loop. Max 3 retries before stopping.

## Step 5 — Polish for sharing

1. README placeholders filled.
2. `docs/` directory with placeholders for the demo GIF, screenshot of the Gradio UI's per-feature tabs, and a "DB-as-only-store" architecture diagram (skill leaves a note: use mermaid/excalidraw).
3. Boot Gradio once in background, confirm `localhost:7860` responds, kill it.
4. Final report:
   ```
   Done.
     project at:    <target_dir>
     features used: vector, memory, <feature_set>
     run with:      cd <target_dir> && python gradio_app.py
     verify:        OK
     notebook:      <target_dir>/notebook.ipynb (executed clean)
     proof:         no Redis/Postgres/SQLite/Chroma/etc — verified by grep.
     next:          record demo, fill "What I built", architecture diagram, push.
   ```

## Stop conditions

- User declines the "Oracle is the only store" constraint. The advanced path doesn't make sense without it; offer them the intermediate path instead.
- User picks fewer features than the minimum set (vector + memory + ≥1 advanced feature).
- Notebook execution fails after 3 retries.
- Verify fails after 3 retries.
- ONNX selected but the chosen model uses SentencePiece. Stop with a clear error.
- Idea 4 selected but the source repo doesn't have a clear storage layer to translate. Stop and ask.

## What you must NOT do

- Don't add Redis. Don't add Postgres. Don't add SQLite. Don't add Chroma / FAISS / Qdrant / Pinecone. Don't write to a filesystem JSON file as state. Verify will catch you.
- Don't make memory ephemeral (in-process dicts). All state in DB.
- Don't pick a feature the user didn't ask for. The Q7 set is the contract.
- Don't ship without the executed notebook. Mandatory.
- Don't claim done before verify *and* notebook execution are both green.
- Don't use recursive WITH for bidirectional graphs. Use Python BFS over an adjacency table per `property-graph.md`.
- Don't try to register a SentencePiece-tokenized ONNX model. It loads then fails on inference. Stay on BertTokenizer-family models.
