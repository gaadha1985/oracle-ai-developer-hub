# Getting started — two worked walk-throughs

Pick a path, point Claude Code (or any agent that follows SKILL.md) at the right `SKILL.md`, answer six questions, get a runnable project. This doc walks you through one beginner build and one intermediate build end-to-end so you don't have to remember anything tomorrow.

---

## Once-only setup (do this before either walk-through)

Three things on your machine:

1. **Docker** — to run Oracle 26ai Free locally. Verify: `docker --version`.
2. **Python 3.11+** with `conda` (or `venv`). Verify: `python --version`.
3. **Ollama** running and reachable at `http://localhost:11434`. Verify:
   ```bash
   ollama list
   # If empty, pull the two defaults the skills assume:
   ollama pull nomic-embed-text
   ollama pull llama3.1:8b
   ```

That's it. Oracle is started by the skill (it scaffolds `docker-compose.yml`); no separate install.

> Skip if you already used CC against this skill set yesterday — you have everything.

---

## Walk-through 1 — Beginner: Personal bookmarks search (idea 1)

What you'll have at the end: two CLI scripts. `add.py` to save URLs, `search.py` to find them by natural-language query. ~80 LOC. No web UI, no chat, no notebook.

### 1. Open Claude Code in an empty project directory

```bash
mkdir -p ~/projects/bookmarks-poc && cd ~/projects/bookmarks-poc
claude
```

Anything outside `~/git/work/oracle-ai-developer-hub` is fine — projects scaffold *into* your chosen directory, not into the hub repo.

### 2. Tell Claude to follow the skill

Paste this exactly (replace the path if your hub repo lives elsewhere):

> Read `/home/ubuntu/git/work/oracle-ai-developer-hub/choose-your-path/SKILL.md` and follow it.

CC will read it, then ask you the six interview questions one by one (or in a batch — depends on the model's mood).

### 3. Answer the interview

| Q | Answer |
| --- | --- |
| Q1 — Path | `beginner` |
| Q2 — Target dir | `.` (or `~/projects/bookmarks-poc` — same thing) |
| Q3 — Database | `local Docker` |
| Q4 — Inference | `Ollama` — embed `nomic-embed-text`, chat `llama3.1:8b` |
| Q5 — Topic | `1` (Personal bookmarks search) |
| Q6 — Notebook | `no` |

Confirm with `y` when CC prints the summary block.

### 4. Wait for the scaffold

CC writes ~10 files: `docker-compose.yml`, `pyproject.toml`, `.env.example`, `src/<project>/{store,add,search}.py`, `migrations/001_core.sql`, `verify.py`, `README.md`. Takes 1-3 minutes.

### 5. Bring up Oracle and verify

Three commands. Run them yourself — the skill won't (it never starts services on your machine):

```bash
cp .env.example .env                       # tweak ORACLE_PWD if you want
docker compose up -d                       # ~60s for the container to be healthy
python -m venv .venv && source .venv/bin/activate
pip install -e .
python verify.py
```

Expected last line: `verify: OK (...)`. If it says `FAIL`, see the troubleshooting section.

### 6. Use it

```bash
python -m bookmarks.add "https://oracle.com/database/ai-vector-search/" "Oracle Vector Search" "the docs page"
python -m bookmarks.add "https://python.langchain.com" "LangChain docs" "framework for LLM apps"
python -m bookmarks.search "vector search documentation"
```

You should see your first bookmark with a similarity score near the top. Done.

---

## Walk-through 2 — Intermediate: PDF-RAG chatbot (idea 1)

What you'll have: a Gradio chat UI on `localhost:7860` that answers questions about PDFs you drop in a folder, with citations and chat history that survives a restart. ~600 LOC.

### 1. New empty directory

```bash
mkdir -p ~/projects/pdf-rag-poc && cd ~/projects/pdf-rag-poc
claude
```

### 2. Invoke the skill

Same paste as before:

> Read `/home/ubuntu/git/work/oracle-ai-developer-hub/choose-your-path/SKILL.md` and follow it.

### 3. Answer the interview

| Q | Answer |
| --- | --- |
| Q1 — Path | `intermediate` |
| Q2 — Target dir | `.` |
| Q3 — Database | `local Docker` |
| Q4 — Inference | `Ollama` (simpler) **or** `OCI GenAI` if you have a tenancy with Grok 4 in `us-chicago-1` |
| Q5 — Topic | `1` (PDF-RAG chatbot) |
| Q6 — Notebook | `yes` (intermediate defaults to yes — gives you a Jupyter walkthrough alongside the app) |

If you pick OCI for Q4, also answer:
- Region (e.g. `us-chicago-1`)
- Compartment OCID
- Auth: `oci-config` (default — uses `~/.oci/config`) or `instance_principal`

Then confirm with `y`.

### 4. Drop PDFs

After the scaffold finishes, put 1-3 PDFs in `data/pdfs/` (any PDFs — Oracle whitepapers, your old slides, anything).

### 5. Bring up the stack

```bash
cp .env.example .env                       # set OCI_* fields here if you picked OCI
docker compose up -d
python -m venv .venv && source .venv/bin/activate
pip install -e .
python verify.py                           # should print verify: OK
python -m <project>.ingest                 # chunks + embeds the PDFs (one-time)
python -m <project>.app                    # boots Gradio on http://localhost:7860
```

Open the URL in your browser. Ask a question about your PDFs. Citations link back to filename + page.

Kill the app, restart it, ask a follow-up — your chat history is still there. That's `OracleChatHistory` doing its job.

### 6. Optional — open the notebook

```bash
jupyter lab notebook.ipynb
```

It walks the same flow cell-by-cell: connect, ingest, query the vector store, hit the chat chain, inspect chat-history rows in Oracle. The last cell launches the Gradio app (and is a "terminator" — kills the kernel after you exit the UI).

---

## When something goes wrong

| Symptom | What to check |
| --- | --- |
| `verify: FAIL — connect: ORA-12541` | Oracle container isn't up. `docker compose ps` should show `healthy`. Wait 60s after first `up -d`. |
| `verify: FAIL — connect: ORA-01017` | Wrong password. `.env` `ORACLE_PWD` must match what you set in `docker-compose.yml`. Default is fine if you didn't change it. |
| `pip install -e .` fails with "no `[build-system]`" | Old scaffold. Re-run the skill — fix `93cc6ff7` added the missing build-system block. |
| Gradio answers but citations are weird strings like `'{"filename": ...}'` | The metadata-as-string monkeypatch didn't get scaffolded. Verify `src/<project>/_monkeypatch.py` exists and is imported at the top of `app.py`. |
| OCI returns 401 | OCI's endpoint needs Signature V1, not bearer auth. The scaffold uses the dual-auth client from `shared/snippets/oci_chat_factory.py` — make sure your `.env` has `OCI_GENAI_BASE_URL` and `OCI_COMPARTMENT_ID`, and `~/.oci/config` exists. |
| `ollama` not reachable | `ollama serve` in another terminal, or `brew services start ollama` on Mac. |

If you're really stuck: re-read the `README.md` the scaffold wrote in your project — it's tailored to your choices and has the exact commands.

---

## What if I want a different idea?

Same flow. At Q5, pick a different number from `choose-your-path/{beginner,intermediate}/project-ideas.md` (each has 8 ideas now). Or pitch your own in free text — the skill will map it to the closest of the eight and confirm.

---

## What if I want the advanced path?

Same shape, but the skill *refuses* to scaffold anything that uses Redis, Postgres, SQLite, ChromaDB, FAISS, Qdrant, or Pinecone — Oracle has to be the only state store. The `verify.py` it generates greps for forbidden imports and fails the build if any sneak in. Pick advanced when you want to feel the constraint, not when you want it to be easy.

---

## TL;DR cheat sheet (rip this out)

```
1. mkdir empty-dir && cd empty-dir && claude
2. "Read /home/ubuntu/git/work/oracle-ai-developer-hub/choose-your-path/SKILL.md and follow it."
3. Answer 6 questions (path, target_dir=., db=local Docker, inference=Ollama, topic=1, notebook=no/yes)
4. cp .env.example .env && docker compose up -d
5. python -m venv .venv && source .venv/bin/activate && pip install -e .
6. python verify.py        → expect: verify: OK
7. python -m <project>.<entrypoint>
```

That's the whole thing.
