# Advanced — project ideas

Five projects where Oracle AI DB is the **only** state store. The advanced skill *refuses* to scaffold Redis, Postgres, SQLite, ChromaDB, FAISS, Qdrant, Pinecone, or filesystem state — that constraint is the whole point.

The user has built agents before. They know what episodic memory is. The skill picks one of these (or accepts a custom pitch that fits the constraint) and builds it real.

---

## 1. Personal research agent

**Pitch.** Ingest papers, build a citation graph, answer questions episodically (it remembers what you've already read).

**Features used.** Vector search · Hybrid search · Property graph · Agent memory tables.

**Shape.**
- Ingestion pipeline: PDFs → chunks → vectors + metadata.
- Graph: papers ↔ citations (entity + edge tables, Python BFS for n-hop).
- Memory layer: 6 memory tables from `apps/finance-ai-agent-demo/backend/memory/manager.py`.
- Agent loop: retrieve (vector + 1-hop graph expansion) → reason → write summary back to `summary_memory`.
- Notebook + Gradio UI showing: chat with the agent, browse the citation graph, inspect what's in memory.

**~2000 lines. Notebook mandatory.**

---

## 2. Code-review agent with auditable history (the JSON Duality showcase)

**Pitch.** Watches a repo, reviews PRs, remembers prior reviews. Agent persists nested JSON; humans run SQL analytics on the same data.

**Features used.** Vector search · **JSON Duality views (the headline)** · Agent memory.

**Shape.**
- Schema: `review` + `review_file` + `review_finding` (per `json-duality.md`).
- View: `review_dv` exposes them as nested JSON.
- Agent: produces a JSON review per PR; writes through `review_dv`; reads taste-history through the same view.
- Dashboard tab in Gradio: runs `SELECT severity, COUNT(*) FROM review_finding GROUP BY severity` and a "most-flagged files" query against the same data.
- Notebook proves duality: write JSON in cell 4, see it as relational rows in cell 5, in the same transaction.

**~1500 lines. Notebook mandatory.**

---

## 3. Email triage agent

**Pitch.** Read an inbox export, classify by intent, remember sender context (preferences, prior asks, relationships).

**Features used.** Vector search · Entity memory · Hybrid search · Optional ONNX in-DB embeddings (sensitivity).

**Shape.**
- Ingestion: `.mbox` or Gmail Takeout export → embed bodies, extract entities (sender, mentioned-people, asks).
- Entity memory: per-person row with vector embedding of their "voice" (concatenated past messages).
- Triage agent: for new email, retrieve sender context + similar past emails + relationship-graph neighbors, classify intent, draft reply.
- Optional: register an in-DB ONNX embedding model so email content never leaves the DB.

**~1800 lines. Notebook mandatory.**

---

## 4. "Translate-this-toy-agent-to-Oracle" path

**Pitch.** User points at an existing toy agent (smolagent, langgraph demo, autogen quickstart). The skill reads it, identifies its storage layer, replaces it with Oracle.

**Features used.** Whatever the original used — translated. Vector store → `OracleVS`. Conversation history → `OracleChatMessageHistory`. Tool log → DB table. Entity memory → entity-memory pattern.

**Shape.**
- The skill reads the user-supplied source, builds a translation plan, asks for approval, then patches.
- Result: same agent, Oracle backing it. Notebook shows the *before* and *after* working identically from the user's perspective.

**Variable size — depends on source. Notebook mandatory.**

This idea is also a *recruiting tool*: it's how someone with a popular agent demo on GitHub gets to brag "now it runs on Oracle AI DB" with one PR.

---

## 5. DevDay-style demo (pick-3-features)

**Pitch.** Pick any 3 features from `shared/references/visual-oracledb-features.md` and build a demo that genuinely uses all 3. Optimized to look great in a 5-minute live demo.

**Features used.** User picks. Skill enforces "all 3 features must actually be exercised in the demo path" (verify checks each).

**Shape.** Variable. The skill scaffolds an outline based on the user's 3 picks, citing the reference docs for each, and the Gradio app has a tab per feature.

**~2000-2500 lines. Notebook mandatory.**

---

## What the skill won't scaffold

- Anything storing state outside Oracle. Hard refusal — the verify step greps for forbidden imports.
- Multi-user auth / ACLs / role-based access. Out of scope.
- Production hardening (rate limits, autoscaling, observability stack). The user can layer that on after.
- Agentic systems with browser automation / OS-level tool use. Tool *registry* is in scope; tool *execution* of arbitrary OS calls is not.
