# Interview — the questions every path asks

The interview is the first interaction. The skill asks these in order, **does not assume defaults silently**, and waits for an answer before continuing. Path-specific skills extend this list with their own questions.

The skill should print the question, the choices, and a one-line note about why it matters. No multi-paragraph explanations.

## Q1 — Path

> **Which path are you on?**
>   1. **beginner** — short script, local-only, no cloud account needed.
>   2. **intermediate** — RAG chatbot with a UI, Oracle + LangChain + (Ollama or OCI GenAI).
>   3. **advanced** — multi-feature agent system, Oracle as the only state store.
>
> *Why: this picks the project shape, the LangChain features taught, and roughly how long this'll take.*

Skip Q1 if the user invoked the skill path-specifically (`beginner/SKILL.md` directly).

## Q2 — Where should the project live?

> **Where on disk should I scaffold this?**
> Default: `~/git/personal/<slug>` where `<slug>` is auto-derived from the project topic (Q5).
>
> *Why: I don't want to drop files in the wrong place. Bail out and ask if `<slug>` already exists.*

If the target dir exists and isn't empty, the skill asks before proceeding (don't overwrite, don't merge silently).

## Q3 — Database target

> **Where's your Oracle database?**
>   1. **Local Docker (default)** — I'll start a 26ai Free container for you.
>   2. **Already-running container** — I'll use the DSN from your `.env`.
>   3. **Autonomous DB on OCI** — paste your wallet path; I'll wire mTLS auth.
>
> *Why: option 1 is the path of least resistance; 3 is for users who already have an Oracle tenancy.*

For v1 the skill only fully supports option 1. Option 2 = the user has done the docker step themselves; the skill skips the compose copy. Option 3 = print "v2 — please use option 1 for now," abort.

## Q4 — Inference

> **What should generate text and embeddings?**
> Beginner default: **Ollama (local)**.
> Intermediate / advanced: pick one.
>   1. **Ollama (local)** — free, no account, no internet.
>   2. **OCI Generative AI** (OpenAI-compatible endpoint) — needs `~/.oci/config` or instance principal; default chat model = Grok 4 (`us-chicago-1` only).
>   3. **Bring your own OpenAI-compatible URL** — paste base_url + key.
>
> *Why: this picks the embedder dim, the chat client, and which env vars get filled in.*

The interview captures the choice and the embedding model in the same breath:
- Ollama → embedder = `nomic-embed-text` (768), chat = `llama3.1:8b` (default) or `qwen2.5:7b` (with thinking-mode mitigations).
- OCI → embedder = `cohere.embed-english-v3.0` (1024), chat = `grok-4`.
- BYO → ask the user for embed model + dim explicitly. The skill *will not guess*.

## Q5 — Project topic

> **What are you building?** Pick from `<path>/project-ideas.md` or describe your own.
>
> *Why: this picks which exemplars get cited and what `verify.py` smoke-tests.*

If the user goes off-script ("I want to do X"), the skill maps X to the closest idea and confirms before scaffolding. If nothing matches well, the skill says so and offers the generic "first vector query" smoke instead of hallucinating a project shape.

## Q6 — Notebook?

> **Want a Jupyter notebook that demonstrates the project?**
>   - beginner default: **no**
>   - intermediate default: **yes**
>   - advanced: **yes (mandatory)** — the notebook is how the project shows off
>
> *Why: notebooks are great for social-media demos but add scaffolding cost. Advanced makes them mandatory because that's where the visual payoff lives.*

## Confirmation gate

After all questions, the skill prints back:

```
About to scaffold:
  path:        intermediate
  target_dir:  ~/git/personal/codebase-qa
  database:    local docker (26ai Free)
  inference:   OCI Generative AI (Grok 4 in us-chicago-1, Cohere embeddings 1024d)
  project:     codebase Q&A
  notebook:    yes
  references:  shared/references/{langchain-oracledb,oci-genai-openai,hybrid-search,...}.md

Proceed? (y/n)
```

The skill does **not** scaffold without an explicit `y`.

## Stop conditions

The interview halts (and the skill exits with a status message, not a half-built project) if:
- The user says they want a database other than Oracle.
- The user picks an embedder/dim the skill can't validate against `OracleVS`.
- The user wants a language other than Python (out of scope v1 — point them at the plan).
- The target dir is non-empty and the user declines overwrite.
