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
- `choose-your-path/PLAN.md` — the design spec, only if the user asks "what's the architecture?"

You do **not** need to load the per-path references at this stage. Each path skill loads its own.

## Step 1 — Ask one question

Print:

```
choose-your-path — pick a difficulty:

  1. beginner       — short script, local-only, no cloud account.
                       (~1 afternoon · langchain-oracledb + Ollama + Oracle 26ai Free)
  2. intermediate   — RAG chatbot with a UI and persistent chat history.
                       (~1-2 days · langchain-oracledb + OCI GenAI or Ollama + Gradio)
  3. advanced       — agent system where Oracle is the only state store.
                       (~3-5 days · multi-feature: vector + memory + JSON Duality / graph / ONNX)

Which one?
```

Wait for a number. If the user describes their experience instead of picking a number ("I've built RAG before"), suggest the matching path and ask them to confirm.

## Step 2 — Dispatch

| Answer | Hand off to |
| --- | --- |
| 1 / "beginner" / "easy" / "first time with Oracle" | `choose-your-path/beginner/SKILL.md` |
| 2 / "intermediate" / "RAG" / "chatbot" | `choose-your-path/intermediate/SKILL.md` |
| 3 / "advanced" / "agent" / "DB as only store" | `choose-your-path/advanced/SKILL.md` |

Pass through `target_dir` and `topic` if the user supplied them.

After handoff, the path skill takes over completely. Don't second-guess it.

## Stop conditions

- User wants something not on this list ("can I do this in TypeScript?", "I want a Postgres backend", "what about Mongo?"). Print: "v1 of choose-your-path is Python + Oracle only. See `PLAN.md` for the rationale and the v2 backlog." Stop.
- User can't decide between two paths after 2 nudges. Default to **intermediate** and tell them so — it's the most common shape.

## What you must NOT do

- Don't try to scaffold from this file. Hand off.
- Don't load all three path skills "to compare." Read `README.md` once; pick.
- Don't invent a fourth path.
