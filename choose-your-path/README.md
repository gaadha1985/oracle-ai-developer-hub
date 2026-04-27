# choose-your-path

A skill set that interrogates you, picks a project at your level, and scaffolds a real, runnable Oracle-AI-DB project you can ship to social media in an afternoon (beginner) or a week (advanced).

Three paths. Pick by complexity, not access — any of them are open to anyone.

| Path | What you build | Stack | Time |
| --- | --- | --- | --- |
| [beginner](./beginner/) | A small CLI that does semantic search on Oracle. | `langchain-oracledb` + Ollama + Oracle 26ai Free in Docker. | ~1 afternoon |
| [intermediate](./intermediate/) | A RAG chatbot with a UI, persistent chat history, hybrid retrieval. | + OCI Generative AI (Grok 4 / Cohere) or Ollama, Gradio. | ~1-2 days |
| [advanced](./advanced/) | An agent system where Oracle is the **only** state store. | + JSON Duality, property graph, ONNX in-DB embeddings, 6 memory types. | ~3-5 days |

## How it works

1. You point your agent (Claude Code, Cursor, Aider, or any agent that reads markdown skills) at this directory.
2. The agent reads `SKILL.md`, asks you one question (which path), then hands off to the path's own skill.
3. The path skill interviews you on the rest (where to scaffold, which inference backend, which project topic), confirms, and builds.
4. The skill brings up an Oracle 26ai Free container in Docker, runs `verify.py` end-to-end, and only declares done when verify is green.
5. You get a polished, social-media-ready repo at the target dir of your choice.

The skills are **agent-agnostic markdown** — they don't depend on a specific harness. Each `SKILL.md` is a step-by-step the agent must follow. The skill cites real exemplar files (from this repo and from jasperan's other projects) so the model copies known-good patterns instead of inventing Oracle SQL.

## What you'll need

- Docker. The skills use the official Oracle 26ai Free image; no Oracle install on your host.
- Python 3.11+.
- (Beginner) Ollama installed. The skills tell you which model to pull.
- (Intermediate / Advanced, optional) An OCI tenancy + `~/.oci/config` if you want Grok 4 / Cohere. Otherwise stick to Ollama and you'll never leave your laptop.

## What gets scaffolded

Every project comes with:

- A working `docker-compose.yml` for Oracle 26ai Free.
- A `verify.py` that proves the whole stack runs end-to-end.
- A README built from a template, with the "Why Oracle" paragraph auto-assembled from the features your project actually uses.
- A `.gitignore` that's already tuned for the project's deps.

Intermediate adds a Gradio UI and a Jupyter notebook. Advanced adds the notebook and a feature-tab dashboard.

## Why three paths

Because building "your first Oracle vector query" and "an agent system using Oracle as the only state store" are not the same task. Lumping them under one tutorial means everyone gets something wrong-sized — too much for the beginner, too thin for the agent-builder. Three sized doors.

## Where projects land

By default: `~/git/personal/<project-slug>` — outside this repo, so what you build is **yours**, not a fork of the developer hub. You take the repo, you push it to your own GitHub, you post the demo. The skill credits the developer hub in the generated README footer; that's the only attribution.

## See also

- [`PLAN.md`](./PLAN.md) — the full design spec. Read this if you want to understand why each path is shaped the way it is, or if you want to contribute a new project idea.
- [`shared/references/`](./shared/references/) — the canonical docs the skills cite. Read these if you want to learn the underlying tech without going through a project.
- [`shared/references/visual-oracledb-features.md`](./shared/references/visual-oracledb-features.md) — frozen catalog of Oracle AI Database features, mirrored from https://jasperan.github.io/visual-oracledb/.

## License

MIT. Built by the [oracle-ai-developer-hub](https://github.com/oracle-devrel/oracle-ai-developer-hub).
