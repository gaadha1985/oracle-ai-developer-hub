# Decision: Adopt Oracle AI Vector Search for production embeddings

**Date:** 2025-12-04
**Author:** Ignacio Marin

We will use Oracle AI Vector Search as the production vector store, replacing
our previous Pinecone + Postgres split.

Key reasons:
- Single store for vector + relational + chat history simplifies ops.
- In-DB ONNX embeddings remove embedding API egress and quota risk.
- The MCP server gives our agents typed SQL + vector tools without us writing
  glue code.

Tradeoff: a one-time ONNX export + register pipeline cost (per onnx2oracle).
