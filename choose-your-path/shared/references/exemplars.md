# Exemplar Citation Index

When a skill needs to show a pattern, it cites a real file from this index. No abstract pseudocode, no invented SQL.

Paths starting with `~/` live outside this repo (jasperan's personal/work tree on the local machine). Paths starting with `apps/` are in this repo. The skill's `target_dir` may not have access to `~/` paths — when that's the case, the skill copies the relevant snippet into `shared/snippets/` first, then cites the local copy. (The snippets directory will be added during build step 7.)

---

## ORACLE CONNECTION

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| beginner | `~/git/personal/oracle-aidev-template/app/db.py` | 1-50 | Lazy-init pool singleton; minimal `oracledb.create_pool()` |
| intermediate | `~/git/personal/cAST-efficient-ollama/src/cast_ollama/oracle_db/connection.py` | 1-70 | Class-based pool with retry + graceful-degradation fallback |
| advanced | `apps/limitless-workflow/src/limitless/db/pool.py` | 1-50 | Singleton pool with mTLS wallet auth |

## VECTOR INSERT + SIMILARITY (raw oracledb)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| beginner | `~/git/personal/oracle-aidev-template/app/vector_search.py` | 30-80 | Insert with RETURNING INTO + `VECTOR_DISTANCE(... COSINE)` |
| beginner | `~/git/personal/cAST-efficient-ollama/src/cast_ollama/oracle_db/operations.py` | 30-100 | Batch insert + JSON metadata + `VECTOR INDEX INMEMORY NEIGHBOR GRAPH ... TARGET ACCURACY 95` |
| beginner | `~/git/personal/cAST-efficient-ollama/src/cast_ollama/oracle_db/schema.py` | 10-45 | `CREATE TABLE ... VECTOR` + `CREATE VECTOR INDEX` |

## LANGCHAIN ORACLEVS (the load-bearing one for beginner + intermediate)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| beginner | `~/git/work/ai-solutions/apps/langflow-agentic-ai-oracle-mcp-vector-nl2sql/components/vectorstores/oracledb_vectorstore.py` | full | Minimal `add_texts` + `similarity_search` — easiest read |
| intermediate | `apps/agentic_rag/src/OraDBVectorStore.py` | 1-100 | Multi-collection wrapper (`PDFCOLLECTION`, `WEBCOLLECTION`, `REPOCOLLECTION`, `GENERALCOLLECTION`) |
| intermediate | `apps/agentic_rag/src/OraDBVectorStore.py` | (same file) | The `_read_similarity_output` JSON-metadata monkeypatch — copy verbatim, otherwise filtered retrievals silently fail |
| intermediate | `apps/limitless-workflow/src/limitless/research/vector_store.py` | full | Hybrid retriever blending vector with keyword filters |

## OCI GENAI EMBEDDINGS (Cohere via `oci` SDK)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| intermediate | `~/git/personal/oci-genai-service/src/oci_genai_service/inference/embeddings.py` | 1-60 | `GenerativeAiInferenceClient.embed_text()` + clean-empty-strings filter |
| intermediate | `~/git/work/ai-solutions/apps/langgraph_agent_with_genai/src/jlibspython/oci_embedding_utils.py` | 1-80 | Falls back to `InstancePrincipalsSecurityTokenSigner` if no `~/.oci/config` |

## OCI GENAI CHAT (OpenAI-compatible endpoint)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| intermediate | `~/git/personal/oci-genai-service/src/oci_genai_service/inference/chat.py` | 1-95 | Dual auth (api_key vs OciOpenAI), streaming, model-agnostic — defaults to Grok 4 |
| intermediate | `apps/agentic_rag/src/openai_compat.py` | 54+ | Exposes OracleVS as OpenAI-compatible `/chat/completions` endpoint |

## OLLAMA (local inference)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| beginner | (use `langchain_ollama.OllamaEmbeddings` directly — no exemplar needed) | — | — |
| under-the-hood | `~/git/personal/cAST-efficient-ollama/src/cast_ollama/embedding/embedder.py` | 1-50 | HTTP POST to `/api/embed` (cite when explaining what LangChain hides) |
| config | `~/git/personal/cAST-efficient-ollama/src/cast_ollama/config.py` | 1-80 | Env-override + YAML fallback; `LOCAL_ONLY` flag |

## ONNX IN-DB EMBEDDINGS (advanced)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| advanced | `~/git/personal/onnx2oracle/src/onnx2oracle/pipeline.py` | 1-100 | HF sentence-transformer → ONNX (tokenizer + transformer + L2-norm), opset ≤ 14 |
| advanced | `~/git/personal/onnx2oracle/src/onnx2oracle/loader.py` | 15-70 | `DBMS_VECTOR.LOAD_ONNX_MODEL` registration, idempotent upload |
| advanced | `~/git/personal/onnx2oracle/tests/test_loader_integration.py` | full | `VECTOR_EMBEDDING(text)` SQL after registration |

## HYBRID SEARCH (intermediate when dropping below LangChain; advanced for full control)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| intermediate | `apps/finance-ai-agent-demo/backend/retrieval/hybrid_search.py` | 1-80 | Pre-filter, post-filter, RRF — three SQL templates |
| intermediate | `apps/finance-ai-agent-demo/backend/retrieval/vector_search.py` | full | Pure vector distance with similarity normalization |
| canonical | `~/git/work/demoapp/api/app/routers/vector.py` | full | REST endpoint with configurable distance metrics |

## AGENT MEMORY (DB-only, advanced)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| advanced | `apps/finance-ai-agent-demo/backend/memory/manager.py` | 1-100 | 6 memory types (conversational, KB, workflow, toolbox, entity, summary) — all Oracle-backed |
| advanced | `apps/finance-ai-agent-demo/backend/memory/sprawl_manager.py` | full | Entity memory with vector embeddings |
| advanced | `apps/agentic_rag/src/OraDBEventLogger.py` | full | Tool execution + intermediate reasoning logs |

## JSON DUALITY VIEWS (advanced)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| advanced | `~/git/work/demoapp/api/app/routers/json_views.py` | 1-80 | `inspection_report_dv` view; `@insert @update @delete` annotations on parent + child; nested `findings` sub-document |

## PROPERTY GRAPH (advanced)

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| advanced | `~/git/work/demoapp/api/app/routers/graph.py` | 1-80 | Adjacency-table edges + Python-side BFS (avoids the recursive-WITH cycle bug on bidirectional graphs) |

## DOCKER COMPOSE — Oracle 26ai Free

| Tier | Path | Range | What it shows |
| --- | --- | --- | --- |
| all paths | `~/git/personal/oracle-aidev-template/docker-compose.yml` | 1-80 | `container-registry.oracle.com/database/free`; healthcheck via sqlplus; init-schema mount; ports 1521 + 5500 |

---

## Gaps the skills must not paper over

- **No clean ONNX in-DB exemplar inside `oracle-ai-developer-hub` itself.** Advanced path cites `~/git/personal/onnx2oracle/`. If the user has only this repo, the skill copies the loader snippet into `shared/snippets/` during build step 7.
- **No JSON Duality / property-graph exemplars in `~/git/personal/`.** Advanced path cites `~/git/work/demoapp/`. Same snippet-copy fallback applies.
- **`langchain-oracledb` chat-history class** is referenced in intermediate but not yet exemplified in any of the surveyed code. The skill must point at the package's own README/source for that one until we have a first-party exemplar.

---

## Excluded from public-facing skills

- `~/git/personal/oci-genai-service/control-plane/genai-control-plane-python-sample/` — internal CLI scripting, not user-facing template material.
- `~/git/personal/orahermes-agent/` — client-restricted state patterns.
- `~/git/personal/jasperan/` — image generation + GitHub stats; off-topic.
- `~/git/work/ai-solutions/apps/agentic_rag/old/` — archived non-canonical code.
- `~/git/work/leagueoflegends-optimizer/`, `~/git/work/foosball-synthesia-api/` — domain-specific demos, not Oracle-AI patterns.
