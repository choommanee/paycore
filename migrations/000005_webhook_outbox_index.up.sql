-- 000005_webhook_outbox_index.up.sql
-- The outbound webhook outbox is drained on a hot polling loop:
--   ListPendingWebhooks ->
--     SELECT ... FROM webhook_events
--     WHERE delivered = FALSE AND failed = FALSE
--       AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
--     ORDER BY created_at
--     LIMIT $1
--
-- The existing idx_webhook_undelivered (delivered, next_attempt_at) does not
-- cover the `failed` predicate and does not support the ORDER BY created_at, so
-- as the table grows with delivered rows the planner falls back to scanning +
-- sorting. A PARTIAL index over only the still-pending rows (delivered = FALSE
-- AND failed = FALSE), ordered by created_at, keeps the outbox drain an
-- index-ordered range scan whose size is bounded by the backlog, not the whole
-- (mostly-delivered) table. next_attempt_at is included so the time predicate is
-- evaluated from the index.
--
-- Not created CONCURRENTLY so it runs inside golang-migrate's per-file
-- transaction.
CREATE INDEX IF NOT EXISTS idx_webhook_outbox_pending
    ON webhook_events (created_at, next_attempt_at)
    WHERE delivered = FALSE AND failed = FALSE;
