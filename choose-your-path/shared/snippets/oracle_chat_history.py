"""Oracle-backed chat message history for LangChain.

Source: shared/references/langchain-oracledb.md (this repo)

WHY THIS EXISTS
---------------
`langchain-oracledb` does not ship a chat-history class as of this writing —
its top-level submodules are only `document_loaders`, `embeddings`,
`retrievers`, `utilities`, `vectorstores`. So we roll a small one ourselves on
top of `oracledb` + LangChain's `BaseChatMessageHistory`.

DDL
---
Run this once during project bootstrap (the skill emits it as part of
`store.py` or a `migrations/` SQL file):

    CREATE TABLE __TABLE__ (
        session_id VARCHAR2(120) NOT NULL,
        seq        NUMBER GENERATED ALWAYS AS IDENTITY,
        payload    CLOB CHECK (payload IS JSON),
        created_at TIMESTAMP DEFAULT SYSTIMESTAMP NOT NULL,
        PRIMARY KEY (session_id, seq)
    )

USAGE
-----
    history = OracleChatHistory(conn, session_id="user-42")
    history.add_user_message("hi")
    history.add_ai_message("hello")
    print(history.messages)  # reload-safe across kernel restarts

    # In a chain:
    from langchain_core.runnables.history import RunnableWithMessageHistory
    chain_with_history = RunnableWithMessageHistory(
        chain,
        lambda sid: OracleChatHistory(conn, session_id=sid),
        input_messages_key="question",
        history_messages_key="history",
    )
"""

from __future__ import annotations

import json

import oracledb
from langchain_core.chat_history import BaseChatMessageHistory
from langchain_core.messages import BaseMessage, messages_from_dict, messages_to_dict


class OracleChatHistory(BaseChatMessageHistory):
    def __init__(
        self,
        conn: oracledb.Connection,
        session_id: str,
        table_name: str = "chat_history",
    ):
        self.conn = conn
        self.session_id = session_id
        self.table = table_name

    @property
    def messages(self) -> list[BaseMessage]:
        with self.conn.cursor() as cur:
            cur.execute(
                f"SELECT payload FROM {self.table} "
                f"WHERE session_id = :sid ORDER BY seq",
                sid=self.session_id,
            )
            rows = []
            for (payload,) in cur.fetchall():
                raw = payload.read() if hasattr(payload, "read") else payload
                rows.append(json.loads(raw))
        return messages_from_dict(rows)

    def add_message(self, message: BaseMessage) -> None:
        payload = json.dumps(messages_to_dict([message])[0])
        with self.conn.cursor() as cur:
            cur.execute(
                f"INSERT INTO {self.table} (session_id, payload) VALUES (:sid, :p)",
                sid=self.session_id,
                p=payload,
            )
        self.conn.commit()

    def clear(self) -> None:
        with self.conn.cursor() as cur:
            cur.execute(
                f"DELETE FROM {self.table} WHERE session_id = :sid",
                sid=self.session_id,
            )
        self.conn.commit()
