---
name: choose-your-path-beginner
description: Scaffold a small, runnable Oracle-AI-DB project using langchain-oracledb + Ollama + Oracle 26ai Free in Docker. For users who haven't touched Oracle before and want to ship something in an afternoon.
inputs:
  - target_dir: where to scaffold (default ~/git/personal/<slug>)
  - topic: optional; one of the ideas in beginner/project-ideas.md, or a free-text pitch
---

The user picked the **beginner** path. Your job is to interrogate them, scaffold a real project, run verify, and stop.

## Step 0 — Read these references first (mandatory)

Load and keep at hand:

- `shared/references/sources.md`
- `shared/references/oracle-26ai-free-docker.md`
- `shared/references/langchain-oracledb.md`  ← load-bearing
- `shared/references/ollama-local.md`
- `shared/references/oracledb-python.md` (skim — beginner only touches `oracledb.connect`)
- `shared/references/exemplars.md`
- `beginner/project-ideas.md`

You may not write SQL, embedder calls, or table-creation code that contradicts these files. If the user asks for something not covered, say so and stop — don't invent.

## Step 1 — Interview

Run the questions from `shared/interview.md`. For beginner specifically:

- Q3 (DB target) — default to "Local Docker" without re-asking unless the user wants otherwise.
- Q4 (Inference) — Ollama. Confirm:
  - chat model: `llama3.1:8b` (default) or `qwen2.5:7b` (with thinking-mode mitigations from `ollama-local.md`).
  - embed model: `nomic-embed-text` (always; 768 dims).
- Q5 (Topic) — pick one of the eight from `beginner/project-ideas.md`. Map free-text pitches to the closest. If none fits, default to idea 5 (smoke-only) and tell the user why.
- Q6 (Notebook) — default **no**. Beginner projects don't get notebooks unless the user explicitly asks.

Print the confirmation block from `interview.md`. Do not proceed without an explicit `y`.

## Step 2 — Resolve choices

Build a scaffold spec dict from interview answers:

| Variable | Source |
| --- | --- |
| `project_slug` | derived from topic, kebab-case (`bookmarks-search`, `recipe-finder`, ...) |
| `target_dir` | answer to Q2, or `~/git/personal/<project_slug>` |
| `embedder_pkg` | `langchain_ollama.OllamaEmbeddings` |
| `embedder_init` | `OllamaEmbeddings(model="nomic-embed-text")` |
| `embedding_dim` | `768` |
| `llm_pkg` | `langchain_ollama.ChatOllama` |
| `llm_init` | `ChatOllama(model="<chat_model>", temperature=0)` |
| `inference_enabled` | `False` for ideas 1-4 (no chat); `False` for idea 5 |
| `qwen_mitigations` | True iff chat model starts with `qwen` |
| `entrypoint` | one of the scripts named in the chosen idea (e.g. `search.py`) |

If `qwen_mitigations`: add `OLLAMA_NUM_THREAD=1` to `.env`, and inject the strip-`<think>` helper into the project's main module per `ollama-local.md`.

## Step 3 — Scaffold

Create files in this exact order. **Cite the exemplar in a comment at the top of any file you write that copies a pattern.**

1. `target_dir/` — create. Refuse if non-empty (per `interview.md`).
2. `.gitignore` — `.env`, `__pycache__/`, `.venv/`, `*.pyc`.
3. `docker-compose.yml` — copy from `shared/templates/docker-compose.oracle-free.yml` verbatim.
4. `.env.example` — copy from `shared/templates/env.example`. Trim sections the project doesn't use (OCI, BYO).
5. `.env` — `.env.example` filled in. Generate `ORACLE_PWD` per `oracle-26ai-free-docker.md`. Do not commit (already in .gitignore).
6. `pyproject.toml` — from `shared/templates/pyproject.toml.template`, with `dependencies = ["oracledb>=2.4", "langchain-oracledb>=0.1", "langchain-ollama>=0.2", "python-dotenv>=1.0"]`. No Gradio, no FastAPI.
7. `store.py` — the one-and-only DB module. Pattern from `~/git/work/ai-solutions/apps/langflow-agentic-ai-oracle-mcp-vector-nl2sql/components/vectorstores/oracledb_vectorstore.py` (cited at top). Exposes:
   ```python
   def get_store(table_name: str) -> OracleVS: ...
   ```
   The `get_store` either calls `OracleVS.from_texts(["__bootstrap__"], ...)` if the table doesn't exist, or returns `OracleVS(client=conn, embedding_function=..., table_name=...)` if it does.
8. **Project files** for the chosen idea, per `beginner/project-ideas.md` shape. Each file ≤ 80 lines. No abstractions. No FastAPI. No UI.
9. `verify.py` — copy `shared/templates/verify.template.py`, fill placeholders. For beginner ideas without an LLM call, set `inference_enabled = False`.
10. `README.md` — copy `shared/templates/readme.template.md`, fill all `{{...}}` placeholders. The "What I built" section stays as a TODO with a one-line comment for the user.

## Step 4 — Verify

1. Print: "Bringing up Oracle 26ai Free; first boot ~90s."
2. Run `docker compose up -d --wait` in `target_dir`.
3. If `--wait` fails after 3 minutes, print the most recent `docker compose logs oracle | tail -20` and stop.
4. Run `python verify.py`. Expect `verify: OK (db, vector)`.
5. On failure, follow the recovery loop in `shared/verify.md` (max 3 retries, then stop and report).
6. Do NOT proceed to Step 5 until verify is green.

## Step 5 — Polish for sharing

1. README — confirm all placeholders are filled.
2. Add `docs/` directory; leave a note: "drop a 30s demo GIF as `docs/demo.gif`."
3. Print a final report:
   ```
   Done.
     project at: <target_dir>
     run with:   cd <target_dir> && python <entrypoint>
     verify:     OK
     next:       record a 30s demo, fill the "What I built" section, push to GitHub.
   ```

## Stop conditions

Stop and ask the user — don't barrel through — when:

- Verify fails 3 times.
- The user's topic doesn't fit any idea and idea 5 (smoke) seems wrong.
- The target dir is non-empty.
- The user picks a non-Oracle DB or non-Python language. Print "out of scope for v1" and stop.

## What you must NOT do

- Don't write raw `CREATE TABLE ... VECTOR(...)` DDL — let `OracleVS.from_texts` do it.
- Don't introduce FastAPI, Gradio, Flask, or any UI framework — beginner is CLI-only.
- Don't introduce Redis, ChromaDB, FAISS, or any non-Oracle store — even as fallback.
- Don't write more than ~200 lines of project code total. If you find yourself going past that, stop and reduce.
- Don't generate ORACLE_PWD with a weak pattern. Use the generator in `oracle-26ai-free-docker.md`.
- Don't claim "done" before verify is green.
