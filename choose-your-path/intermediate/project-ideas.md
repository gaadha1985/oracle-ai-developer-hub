# Intermediate — project ideas

Five RAG-shaped projects. Each lands a working chatbot with a UI, persistent chat history, and hybrid retrieval — all on Oracle. The user has built RAG before (probably with FAISS or Chroma); this is them rebuilding it on Oracle and OCI GenAI.

The skill asks the user to pick one. Free-text pitches get mapped to the closest. If nothing maps well, the skill picks idea 1 (PDF-RAG) as the safest default.

---

## 1. PDF-RAG chatbot

**Pitch.** Drop PDFs in a folder, get a chat UI that answers questions about them with citations.

**What the user does.**
- `python ingest.py` — chunks + embeds the PDFs into one collection per file.
- `gradio_app.py` — chat UI on `localhost:7860`. Conversation history persists across restarts.

**LangChain primitives forced.**
- Multi-collection `OracleVS` (one collection per PDF, one shared `CONVERSATIONS` for chat history).
- `EnsembleRetriever` for hybrid search.
- `OracleChatMessageHistory` + `RunnableWithMessageHistory` for stateful chat.
- Metadata = `{"filename": ..., "page": ...}` so citations point at exact pages.

**Shape.** `src/<project>/{ingest,store,chains,app}.py`. ~600 lines total.

**Demo.** GIF: drop a PDF, ask 3 questions, see citations + persistent history after a kill/restart.

---

## 2. Codebase Q&A

**Pitch.** Index a Git repo's source. "Where is auth implemented? How does the rate limiter work?"

**What the user does.**
- `python ingest.py /path/to/repo` — walks the tree, embeds source files (chunked by symbol, not by line count).
- `gradio_app.py` — chat UI. Filter by language or directory at retrieval time.

**LangChain primitives forced.**
- One `OracleVS` collection, with metadata = `{"path": ..., "lang": ..., "symbol": ...}`.
- `vs.as_retriever(search_kwargs={"k": 5, "filter": {"lang": "python"}})` — metadata filtering at retrieval.
- The metadata-as-string monkeypatch (filtered retrieval *requires* it).

**Shape.** Same skeleton as PDF-RAG. ~700 lines.

**Demo.** Index this very repo. Ask "where does the agentic_rag verify step run?"; get `apps/agentic_rag/tests/test_smoke_reasoning.py` cited correctly.

---

## 3. Web-page librarian

**Pitch.** Personal Pocket replacement with semantic search and answers-with-citations.

**What the user does.**
- `python add.py "https://..."` — fetches, extracts text (use `trafilatura`), chunks, embeds.
- `gradio_app.py` — search + chat. Citations link back to the source URL.

**LangChain primitives forced.**
- `OracleVS.add_documents()` per page (vs `from_texts` once).
- `similarity_search_with_score()` to surface confidence.
- Persistent chat history scoped per user session.

**Shape.** ~600 lines.

**Demo.** Save 10 pages, ask cross-page questions, see source URLs in answers.

---

## 4. Slack-thread digest

**Pitch.** Paste an exported Slack export; get a chat UI that knows your team's history.

**What the user does.**
- `python ingest.py slack_export.zip` — one collection per channel.
- `gradio_app.py` — pick channel(s), chat. Summarize chains over retrieved threads.

**LangChain primitives forced.**
- Multi-collection (one per channel) with cross-collection search.
- Summarization chain reading from the retriever, writing to a second collection (pattern that prefigures the advanced "second brain" idea).

**Shape.** ~700 lines.

**Demo.** "What did we decide about the migration?" → summary + thread links.

---

## 5. Personal Wikipedia (markdown notes)

**Pitch.** RAG over your second brain. Two collections — `RAW_NOTES` and `SYNTHESIZED_SUMMARIES` — and the agent writes back into the second one.

**What the user does.**
- Point the skill at a folder of `.md` files (Obsidian, Logseq, plain notes — whatever).
- `gradio_app.py` — chat. Each conversation produces a summary that gets embedded into `SYNTHESIZED_SUMMARIES` so future questions can recall *prior conversations*, not just notes.

**LangChain primitives forced.**
- Two-collection pattern with cross-collection retrieval.
- Agent that *writes* to a vector store, not just reads.
- Persistent chat history.

**Shape.** ~800 lines. The most ambitious of the five.

**Demo.** Show how asking a question once and asking it again three days later surfaces the prior conversation as context.

---

## What the skill won't scaffold

- **Multi-user with auth.** Out of scope. Single-user only.
- **Production deployment.** No Docker for the *app*, no nginx, no TLS. The user can dockerize after the project works.
- **Anything where Oracle isn't the vector store.** No FAISS / Chroma / Qdrant fallback.
- **Voice input, image RAG, agent tool-calling.** All advanced-tier features.
