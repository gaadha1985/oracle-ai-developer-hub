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
> All three tiers use **OCI Generative AI** for the LLM (`grok-4` in `us-chicago-1`, OpenAI-compatible endpoint, Pattern 1 SigV1 auth). What differs is the embedder:
>   - **beginner** → OCI `cohere.embed-english-v3.0` (1024 dim).
>   - **intermediate / advanced** → in-database ONNX (`MY_MINILM_V1`, 384 dim) — embeddings happen inside Oracle via `VECTOR_EMBEDDING(MODEL USING text)`, no external embedder process.
>
> *Why: this picks the embedder dim, the chat client, and which env vars get filled in. All three tiers require `~/.oci/config` or instance principal.*

The skill confirms:
- `~/.oci/config` exists (or instance principal env vars present).
- `OCI_COMPARTMENT_ID` is captured.
- The user is OK with non-zero OCI cost (Grok 4 is not on the always-free list).
- Region is `us-chicago-1` (warn + offer Cohere/Llama same-region fallback if not).

Older versions of this interview offered Ollama and BYO endpoints. Those flows live in `archive/` ideas; the active tiers don't surface them.

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
  target_dir:  ~/git/personal/nl2sql-explorer
  database:    local docker (26ai Free)
  inference:   OCI GenAI Grok 4 (us-chicago-1)
               + in-DB ONNX embeddings (MY_MINILM_V1, 384d)
  mcp:         oracle-database-mcp-server (read_only)
  project:     NL2SQL data explorer
  notebook:    yes
  references:  shared/references/{langchain-oracledb,oci-genai-openai,onnx-in-db-embeddings,...}.md

Proceed? (y/n)
```

The skill does **not** scaffold without an explicit `y`. For advanced idea 3 (conversational schema designer), the skill *additionally* requires an explicit `y` for `mcp_sql_mode=read_write`.

## Stop conditions

The interview halts (and the skill exits with a status message, not a half-built project) if:
- The user says they want a database other than Oracle.
- The user has no OCI tenancy / `~/.oci/config` and won't set one up.
- The user picks an embedder/dim the skill can't validate against `OracleVS`.
- The user wants a language other than Python (out of scope v1 — point them at the plan).
- The target dir is non-empty and the user declines overwrite.
