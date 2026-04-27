---
name: oracle-aidb-docker-setup
description: Bring up Oracle 26ai Free in Docker. Generates docker-compose.yml + .env, waits for healthy, smoke-tests connect. Idempotent — safe to re-invoke. Use when any project needs a fresh local Oracle 26ai instance.
inputs:
  - target_dir: where to write docker-compose.yml + .env (default = current dir)
  - port: host port to expose on (default = 1521)
  - oracle_pwd: optional; if not set, generate a strong one
  - volume_name: docker named volume for persistence (default = oracle_<project_slug>_data)
outputs:
  - target_dir/docker-compose.yml
  - target_dir/.env  (contains ORACLE_PWD, DB_DSN)
  - a healthy `oracle` container reachable on host:port
---

You scaffold the Oracle 26ai Free container. Nothing else. No Python, no app code.

## Step 0 — References

- `shared/references/oracle-26ai-free-docker.md` — image, healthcheck, password rules.
- `shared/templates/docker-compose.oracle-free.yml` — the canonical compose file. Copy it; don't reinvent.

## Step 1 — Detect existing setup

In `target_dir`:
- If `docker-compose.yml` exists AND it contains `image: container-registry.oracle.com/database/free:`, **don't overwrite**. Read its `ORACLE_PWD` from the existing `.env` and skip to Step 4.
- If `docker-compose.yml` exists but is for something else, stop and ask the user. Don't clobber.
- If neither, proceed.

## Step 2 — Generate password

If the user passed `oracle_pwd`, use it (validate first: 12+ chars, mixed case, digit, no `$/@/"`).

Otherwise generate one:

```python
import secrets, string
alphabet = string.ascii_letters + string.digits
pwd = "Or" + "".join(secrets.choice(alphabet) for _ in range(18))
```

Why no symbols: Oracle's `ORACLE_PWD` env var splits on `$`, breaks on `@` (interpreted as connect string), and chokes on quotes. Letters + digits avoid the whole class.

## Step 3 — Write compose + env

`target_dir/docker-compose.yml` — copied verbatim from `shared/templates/docker-compose.oracle-free.yml`, with `${ORACLE_PORT}`, `${ORACLE_PWD}`, and `${ORACLE_VOLUME}` placeholders left in (compose reads them from `.env`).

`target_dir/.env`:

```
ORACLE_PWD=<generated>
ORACLE_PORT=<port>
ORACLE_VOLUME=<volume_name>
DB_DSN=localhost:<port>/FREEPDB1
DB_USER=SYSTEM
DB_PASSWORD=<same as ORACLE_PWD>
```

Don't commit `.env`. The skill confirms `target_dir/.gitignore` contains `.env` — adds the line if missing.

## Step 4 — Bring it up

```bash
cd target_dir
docker compose up -d --wait
```

`--wait` blocks until the healthcheck passes. First boot: ~90s. Subsequent boots: ~15s.

If `--wait` exits non-zero:
1. Print `docker compose logs oracle | tail -30`.
2. Common causes: port collision (someone else has 1521 — ask the user to pick a different port and re-run), insufficient memory (Oracle needs ~2GB; tell the user), corrupt volume (rare, suggest `docker compose down -v` AFTER confirming with the user).
3. Stop. Don't retry blindly.

## Step 5 — Smoke connect

```python
import oracledb, os
from dotenv import load_dotenv

load_dotenv(f"{target_dir}/.env")
conn = oracledb.connect(
    user=os.environ["DB_USER"],
    password=os.environ["DB_PASSWORD"],
    dsn=os.environ["DB_DSN"],
)
with conn.cursor() as cur:
    cur.execute("SELECT 'ok' FROM dual")
    assert cur.fetchone()[0] == "ok"
print("oracle-aidb-docker-setup: OK")
```

If this fails with `ORA-12541` (no listener), wait another 30s and retry once — the container claims healthy a few seconds before the listener is ready in some kernels. If it still fails, stop and surface the error.

If `oracledb` isn't installed in the current env, install it: `pip install oracledb`.

## Stop conditions

- `docker` not on PATH. Tell the user to install Docker Desktop / engine and stop.
- Port already in use AND user hasn't supplied a different one. Ask which port to use.
- The user's `target_dir` is a different project's repo and already has a `docker-compose.yml`. Don't clobber.

## What you must NOT do

- Don't generate weak passwords (no defaults like `Welcome123`).
- Don't run `docker system prune` or anything that touches other containers.
- Don't expose Oracle on `0.0.0.0` — bind to `127.0.0.1:<port>:1521`.
- Don't mount the data volume on a path inside the project repo. Use a docker named volume.

## Final report

```
oracle-aidb-docker-setup: OK
  compose:  target_dir/docker-compose.yml
  env:      target_dir/.env  (ORACLE_PWD generated)
  dsn:      localhost:<port>/FREEPDB1
  status:   healthy (smoke connect OK)
  next:     hand off to a higher-level skill, or run your own oracledb.connect(...)
```
