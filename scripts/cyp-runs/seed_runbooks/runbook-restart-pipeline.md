# Runbook: Restart the ingest pipeline

When the ingest pipeline alarms with `BACKLOG > 10000`:

1. Check the dead-letter queue length: `oci queue list-messages --queue-id $DLQ_OCID`.
2. If the DLQ is non-empty, page #data-platform; the upstream is likely emitting malformed events.
3. If the DLQ is empty, restart workers: `kubectl rollout restart deployment/ingest-worker -n data`.
4. Watch backlog for 5 minutes. If it does not drain, escalate.
