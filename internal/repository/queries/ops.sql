-- internal/repository/queries/ops.sql
-- Data access for the background ops workers: settlement, reconciliation and
-- outbound webhook delivery. All parameters are bound ($1, $2, ...).

-- Settlement: aggregate captured (net of refunded) payments per merchant/currency
-- within a time window. Only captured/partial_refunded/refunded rows have moved
-- money and are eligible for payout.

-- name: AggregateCapturedForSettlement :many
SELECT
    merchant_id,
    currency,
    SUM(captured_amount_minor)::BIGINT AS gross_minor,
    SUM(refunded_amount_minor)::BIGINT AS refunded_minor,
    COUNT(*)::BIGINT                    AS payment_count
FROM payments
WHERE status IN ('captured', 'partial_refunded', 'refunded')
  AND updated_at >= $1
  AND updated_at <  $2
GROUP BY merchant_id, currency;

-- name: CreatePayout :one
INSERT INTO payouts (
    id, merchant_id, currency, gross_minor, refunded_minor, fee_minor, net_minor,
    payment_count, status, period_start, period_end
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- Reconciliation: ledger-side captured totals per merchant/currency. capture and
-- settlement entries are additive; refund/void reduce the net.

-- name: LedgerCapturedTotals :many
SELECT
    p.merchant_id                                    AS merchant_id,
    l.currency                                       AS currency,
    COALESCE(SUM(
        CASE WHEN l.entry_type = 'capture' THEN l.amount_minor
             WHEN l.entry_type = 'refund'  THEN -l.amount_minor
             ELSE 0 END
    ), 0)::BIGINT                                    AS net_minor
FROM ledger_entries l
JOIN payments p ON p.id = l.payment_id
WHERE l.created_at >= $1 AND l.created_at < $2
GROUP BY p.merchant_id, l.currency;

-- name: CreateReconMismatch :exec
INSERT INTO recon_mismatches (
    id, merchant_id, currency, ledger_minor, acquirer_minor, delta_minor, report_date
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- Outbound webhook delivery.

-- name: ListPendingWebhooks :many
SELECT * FROM webhook_events
WHERE delivered = FALSE
  AND failed = FALSE
  AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())
ORDER BY created_at
LIMIT $1;

-- name: MarkWebhookDelivered :exec
UPDATE webhook_events
SET delivered = TRUE,
    delivered_at = NOW(),
    attempts = attempts + 1,
    last_error = NULL
WHERE id = $1;

-- name: MarkWebhookRetry :exec
UPDATE webhook_events
SET attempts = attempts + 1,
    next_attempt_at = $2,
    last_error = $3
WHERE id = $1;

-- name: MarkWebhookFailed :exec
UPDATE webhook_events
SET failed = TRUE,
    attempts = attempts + 1,
    last_error = $2
WHERE id = $1;
