---
name: choose-your-path
description: Top-level router for the choose-your-path skill set. Asks the user one question (which path?), then dispatches to beginner/, intermediate/, or advanced/. Use when the user wants to scaffold an Oracle-AI-DB project but hasn't picked a difficulty yet.
inputs:
  - target_dir: optional; passed through to the path-specific skill
  - topic: optional; passed through
---

You are the entry point. Your only job is to pick which path's `SKILL.md` to hand off to. Do not scaffold anything yourself.

## Step 0 — Read these references

- `choose-your-path/README.md` — what each path is, in plain English.
- `choose-your-path/skills/README.md` — the building-block skills the higher tiers compose.
- `choose-your-path/PLAN.md` — only if the user asks "what's the architecture?"

You do **not** need to load the per-path references at this stage. Each path skill loads its own.

## Step 1 — Ask one question

Print:

```
choose-your-path — pick a difficulty:

  1. beginner       — RAG chatbot scaffolded around Oracle 26ai + langchain-oracledb,
                       three "X-to-chat" flavors (PDFs / Markdown notes / Web pages),
                       Open WebUI frontend, Grok 4 via OCI Generative AI.
                       (~1 afternoon · 3 ideas to pick from)

  2. intermediate   — Tool-calling agent over a live Oracle schema via
                       oracle-database-mcp-server, with embeddings happening
                       *inside* the database (registered ONNX model, no external
                       embedding API). Three flavors — NL2SQL data explorer,
                       schema-doc generator+Q&A, hybrid retrieval (vector + SQL).
                       Open WebUI + Grok 4.
                       (~1-2 days · 3 ideas to pick from)

  3. advanced       — Agent system where Oracle is the *only* state store,
                       composed from the choose-your-path/skills/ building blocks
                       (oracle-aidb-docker-setup + langchain-oracledb-helper +
                       oracle-mcp-server-helper). Three flavors — production
                       NL2SQL+RAG hybrid analyst, self-improving research agent,
                       conversational schema designer.
                       (~3-5 days · 3 ideas to pick from)

Which one?
```

Wait for a number. If the user describes their experience instead of picking a number ("I've built RAG before"), suggest the matching path and ask them to confirm.

## Step 2 — Dispatch

| Answer | Hand off to |
| --- | --- |
| 1 / "beginner" / "easy" / "first time with Oracle" | `choose-your-path/beginner/SKILL.md` |
| 2 / "intermediate" / "RAG" / "MCP" / "tool-calling" | `choose-your-path/intermediate/SKILL.md` |
| 3 / "advanced" / "agent" / "DB as only store" | `choose-your-path/advanced/SKILL.md` |

Pass through `target_dir` and `topic` if the user supplied them.

After handoff, the path skill takes over completely. Don't second-guess it.

## Step 3 — One-time prerequisites check (before handoff)

The new (post-restructure) skill set requires:

1. **Docker** — verify `docker --version` works.
2. **Python 3.11+** — verify `python --version`.
3. **OCI tenancy with `~/.oci/config`** — *all three tiers* now require this. Grok 4 lives at `inference.generativeai.us-chicago-1.oci.oraclecloud.com`. If the user has no OCI config:
   - Ask if they want to do `oci setup config` first (yes → stop here, point them at the OCI docs).
   - If no — tell them this skill set needs OCI for the LLM and stop. Earlier tiers had Ollama fallback; that's been moved to `archive/` ideas.
4. **`OCI_COMPARTMENT_ID`** — capture from env if set, otherwise prompt.

These prerequisites apply to every tier. Don't re-ask them inside each path.

## Stop conditions

- User wants something not on this list ("can I do this in TypeScript?", "I want a Postgres backend", "what about Mongo?"). Print: "v1 of choose-your-path is Python + Oracle only. See `PLAN.md` for the rationale and the v2 backlog." Stop.
- User can't decide between two paths after 2 nudges. Default to **intermediate** and tell them so — it's the most representative shape.
- User has no OCI tenancy. Stop. Point at `archive/` for Ollama-flavored older ideas, or at the OCI free trial signup.

## What you must NOT do

- Don't try to scaffold from this file. Hand off.
- Don't load all three path skills "to compare." Read `README.md` once; pick.
- Don't invent a fourth path.
- Don't suggest Ollama as a fallback. The post-restructure tiers are OCI-GenAI-only on purpose.
