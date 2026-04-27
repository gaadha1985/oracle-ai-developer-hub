"""verify.py — the gate every choose-your-path project must pass before "done".

Run this after `docker compose up -d --wait`. It exits non-zero on any failure
and prints exactly one line on success: `verify: OK (...)`.

PLACEHOLDERS the scaffolding skill replaces (in order they appear):
  {{embedder_init}}      — full embedder constructor expression, e.g.
                           OllamaEmbeddings(model="nomic-embed-text")
                           OciCohereEmbeddings(model="cohere.embed-english-v3.0", ...)
  {{llm_init}}           — chat-LLM constructor, e.g.
                           ChatOllama(model="llama3.1:8b", temperature=0)
                           OciOpenAI-backed wrapper (see oci-genai-openai.md)
  {{prompt}}             — short, deterministic test prompt as a Python string,
                           e.g. "Reply with the single word OK."
  {{inference_enabled}}  — Python literal True or False; True if the project
                           uses a chat LLM, False if it's vector-only.

The skill ALSO injects the right `import` lines at the top of the file for
whichever embedder/LLM it picked, plus the metadata-as-string monkeypatch
from shared/references/langchain-oracledb.md when the project uses OracleVS
filtered retrieval.

Do not edit by hand unless you know what you're doing.
"""
from __future__ import annotations

import os
import sys

import oracledb


def _connect():
    return oracledb.connect(
        user=os.environ["ORACLE_USER"],
        password=os.environ["ORACLE_PWD"],
        dsn=os.environ["ORACLE_DSN"],
    )


def check_db() -> str:
    with _connect() as conn, conn.cursor() as cur:
        cur.execute("SELECT 1 FROM DUAL")
        if cur.fetchone()[0] != 1:
            raise RuntimeError("DB connect succeeded but SELECT 1 returned wrong value")
    return "db"


def check_vector() -> str:
    """Round-trip one known string through OracleVS and assert it comes back.

    The skill replaces {{embedder_init}} with the concrete embedder.
    """
    from langchain_oracledb import OracleVS
    from langchain_oracledb.utils.distance_strategy import DistanceStrategy

    embedder = {{embedder_init}}  # e.g. OllamaEmbeddings(model="nomic-embed-text")

    with _connect() as conn:
        vs = OracleVS.from_texts(
            texts=["choose-your-path verify smoke test"],
            embedding=embedder,
            client=conn,
            table_name="CYP_VERIFY_SMOKE",
            distance_strategy=DistanceStrategy.COSINE,
        )
        hits = vs.similarity_search("verify smoke test", k=1)
        if not hits or "verify smoke" not in hits[0].page_content.lower():
            raise RuntimeError(f"vector round-trip failed; got {hits!r}")

        # Best-effort cleanup so re-running verify stays idempotent.
        try:
            with conn.cursor() as cur:
                cur.execute("DROP TABLE CYP_VERIFY_SMOKE PURGE")
            conn.commit()
        except Exception:
            pass
    return "vector"


def check_inference() -> str:
    """One deterministic LLM call. Skill replaces {{llm_init}} + {{prompt}}."""
    llm = {{llm_init}}  # e.g. ChatOllama(model="llama3.1:8b", temperature=0)
    out = llm.invoke({{prompt}})  # e.g. "Reply with the single word OK."
    text = getattr(out, "content", out)
    if not text or not str(text).strip():
        raise RuntimeError(f"inference returned empty: {out!r}")
    return "inference"


def main() -> int:
    checks = []
    try:
        checks.append(check_db())
        checks.append(check_vector())
        if {{inference_enabled}}:
            checks.append(check_inference())
    except Exception as e:  # noqa: BLE001 — verify must surface anything
        print(f"verify: FAIL ({checks[-1] if checks else 'startup'}): {e}", file=sys.stderr)
        return 1
    print(f"verify: OK ({', '.join(checks)})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
