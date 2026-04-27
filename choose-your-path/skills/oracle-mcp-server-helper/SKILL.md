---
name: oracle-mcp-server-helper
description: Wire the oracle-database-mcp-server into a Python project so an LLM agent can call list_tables / describe_table / run_sql / vector_search at inference time. Handles install, stdio launch, and LangChain tool conversion via langchain-mcp-adapters. Use whenever a project needs a Grok 4 / GPT-class agent that talks to a live Oracle schema.
inputs:
  - target_dir: project root (must already have a working DB and store layer)
  - package_slug: snake_case Python package name
  - allowed_tools: subset of {list_tables, describe_table, describe_schema, run_sql, vector_search} — default = all
  - sql_mode: "read_only" (default) or "read_write" — read_write enables INSERT/UPDATE/DELETE/DDL
  - tool_prefix: optional namespace for emitted LangChain tools (default = "oracle_")
outputs:
  - target_dir/src/<package_slug>/mcp_client.py    (server launcher + LangChain tool factory)
  - target_dir/src/<package_slug>/tool_registry.py (cached list[BaseTool] for the agent loop)
  - additions to target_dir/pyproject.toml
  - additions to target_dir/.env
---

You wire MCP. You do not write the agent loop, the chain, or the UI — those are tier-skill responsibilities.

## Step 0 — References

- `shared/references/sources.md` — links to oracle-database-mcp-server.
- `shared/references/oracledb-python.md` — `oracledb.connect()` shape.
- `shared/references/langchain-oracledb.md` — vector_search semantics if exposed.

## Step 1 — Validate inputs

- `target_dir/.env` has `DB_DSN`, `DB_USER`, `DB_PASSWORD`.
- A working `store.py` exists at `target_dir/src/<package_slug>/store.py` (langchain-oracledb-helper output). If not, stop — order matters.
- `sql_mode == "read_write"` requires the user to confirm explicitly during interview. Refuse to proceed silently — destructive SQL via an LLM agent is a footgun.

## Step 2 — Add deps

In `target_dir/pyproject.toml` under `[project] dependencies`:

```
"oracle-database-mcp-server>=0.1",
"langchain-mcp-adapters>=0.1",
"mcp>=0.9",
```

Run `pip install -e .` from `target_dir`. If the user's environment is conda, prefer `python -m pip install -e .` over `pip` directly (per CLAUDE.md note).

## Step 3 — Add env keys

Append to `target_dir/.env.example` (and `.env` if it exists):

```
# Oracle MCP server — connection inherited from DB_DSN/DB_USER/DB_PASSWORD
ORACLE_MCP_SQL_MODE=<read_only|read_write>
ORACLE_MCP_ALLOWED_TOOLS=<comma-separated allowed_tools>
```

## Step 4 — Write `mcp_client.py`

```python
"""
Oracle MCP client wrapper.

Spawns oracle-database-mcp-server over stdio, attaches a single client session
to it, and converts the server's tools into LangChain BaseTool instances via
langchain-mcp-adapters.

Lifecycle: the session is module-scoped and lazy. First call spawns; subsequent
calls reuse. The session shuts down on process exit (atexit).

Cites:
- shared/references/sources.md → oracle-database-mcp-server
"""
import asyncio
import atexit
import os
from contextlib import asynccontextmanager
from typing import Optional

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from langchain_mcp_adapters.tools import load_mcp_tools


_session: Optional[ClientSession] = None
_session_cm = None  # keep the context manager alive
_loop: Optional[asyncio.AbstractEventLoop] = None


def _server_params() -> StdioServerParameters:
    return StdioServerParameters(
        command="oracle-database-mcp-server",
        args=[
            "--dsn", os.environ["DB_DSN"],
            "--user", os.environ["DB_USER"],
            "--mode", os.environ.get("ORACLE_MCP_SQL_MODE", "read_only"),
        ],
        env={"DB_PASSWORD": os.environ["DB_PASSWORD"]},
    )


async def _ensure_session() -> ClientSession:
    global _session, _session_cm
    if _session is not None:
        return _session
    _session_cm = stdio_client(_server_params())
    read, write = await _session_cm.__aenter__()
    session = ClientSession(read, write)
    await session.__aenter__()
    await session.initialize()
    _session = session
    return _session


def _shutdown():
    if _session_cm is not None:
        try:
            asyncio.get_event_loop().run_until_complete(_session_cm.__aexit__(None, None, None))
        except Exception:
            pass


atexit.register(_shutdown)


def get_loop() -> asyncio.AbstractEventLoop:
    global _loop
    if _loop is None:
        _loop = asyncio.new_event_loop()
    return _loop


async def _list_tools_async():
    session = await _ensure_session()
    tools = await load_mcp_tools(session)
    allowed = set(os.environ.get("ORACLE_MCP_ALLOWED_TOOLS", "").split(","))
    if allowed:
        tools = [t for t in tools if t.name in allowed or f"oracle_{t.name}" in allowed]
    return tools


def list_tools():
    """Synchronous wrapper — returns list[BaseTool] usable by LangChain agents."""
    loop = get_loop()
    return loop.run_until_complete(_list_tools_async())
```

Notes for the tier skill that uses this:
- Tools come back already shaped as `langchain_core.tools.BaseTool`. Bind them to the LLM via `llm.bind_tools(list_tools())` — no manual `@tool` decoration needed.
- The async-from-sync bridge is the simplest possible. If the tier skill builds an async agent, expose `_list_tools_async` directly instead.

## Step 5 — Write `tool_registry.py`

```python
"""
Cached list of available MCP tools for the agent.

Why a registry: building the tool list spawns a subprocess. Cache it once
per process so the agent loop doesn't pay that cost per turn.
"""
from functools import lru_cache
from typing import List
from langchain_core.tools import BaseTool

from .mcp_client import list_tools


@lru_cache(maxsize=1)
def get_tools() -> List[BaseTool]:
    return list_tools()


def get_tool(name: str) -> BaseTool:
    for t in get_tools():
        if t.name == name:
            return t
    raise KeyError(f"no MCP tool named {name!r}")
```

## Step 6 — Smoke

```python
from <package_slug>.tool_registry import get_tools

tools = get_tools()
names = [t.name for t in tools]
print(f"oracle-mcp-server-helper: OK (tools: {', '.join(names)})")

# call list_tables to prove the server actually works
list_tables = next(t for t in tools if "list_tables" in t.name)
result = list_tables.invoke({})
print(f"  found {len(result)} tables")
```

If the smoke hangs > 30s, the MCP server didn't initialize — usually because `DB_DSN` is wrong or the container isn't healthy. Stop and report.

## Stop conditions

- `oracle-database-mcp-server` binary not on PATH after `pip install -e .`. Show the install path and stop.
- `sql_mode=read_write` without explicit user confirmation. Refuse.
- The MCP server fails to initialize within 30s. Print the stderr from the subprocess and stop.

## What you must NOT do

- Don't expose `run_sql` in `read_write` mode without surfacing the risk in the tier README.
- Don't share one MCP session across processes (Gunicorn workers, etc.) — each process spawns its own. Document this.
- Don't convert MCP tools manually. Use `load_mcp_tools` — it preserves the JSON schema for tool calls.

## Final report

```
oracle-mcp-server-helper: OK
  client:    target_dir/src/<package_slug>/mcp_client.py
  registry:  target_dir/src/<package_slug>/tool_registry.py
  tools:     <comma list of tool names>
  sql_mode:  <read_only|read_write>
  next:      hand off to the tier skill — it builds the agent loop using these tools.
```
