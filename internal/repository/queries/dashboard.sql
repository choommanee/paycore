-- internal/repository/queries/dashboard.sql
-- Merchant Dashboard read models: KPI aggregation over payments, the merchant's
-- own settlement/payout rows, and its full dispute list. Everything is scoped to
-- the authenticated merchant id ($1) so there is no cross-tenant leakage. All
-- parameters are bound ($1, $2, ...) so there is no SQL-injection surface.

-- name: MerchantStats :one
-- One-row KPI rollup for a merchant over an inclusive [from, to] window. Volume
-- is captured (net-of-refund) money in minor units; the by-status counts drive
-- the dashboard tiles and the success / refund ratios. authorized here means
-- "reached authorized or beyond" (authorized|captured|partial_refunded|refunded)
-- so the success rate reflects payments that were approved by the acquirer.
SELECT
    COUNT(*)::BIGINT AS count,
    COALESCE(SUM(
        CASE WHEN status IN ('captured', 'partial_refunded', 'refunded')
             THEN captured_amount_minor ELSE 0 END
    ), 0)::BIGINT AS volume_minor,
    COALESCE(SUM(
        CASE WHEN status IN ('authorized', 'captured', 'partial_refunded', 'refunded')
             THEN 1 ELSE 0 END
    ), 0)::BIGINT AS authorized_count,
    COALESCE(SUM(
        CASE WHEN status IN ('captured', 'partial_refunded', 'refunded')
             THEN 1 ELSE 0 END
    ), 0)::BIGINT AS captured_count,
    COALESCE(SUM(
        CASE WHEN status IN ('refunded', 'partial_refunded')
             THEN 1 ELSE 0 END
    ), 0)::BIGINT AS refunded_count,
    COALESCE(SUM(
        CASE WHEN status = 'failed' THEN 1 ELSE 0 END
    ), 0)::BIGINT AS failed_count
FROM payments
WHERE merchant_id = $1
  AND created_at >= $2
  AND created_at <= $3;

-- name: CountDisputesByMerchant :one
-- Chargeback count for a merchant over the same window, used to derive the
-- chargeback ratio against total payment count.
SELECT COUNT(*)::BIGINT AS dispute_count
FROM disputes
WHERE merchant_id = $1
  AND created_at >= $2
  AND created_at <= $3;

-- name: ListPayoutsByMerchant :many
-- Settlement / payout rows for a merchant, most recent first.
SELECT * FROM payouts
WHERE merchant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllPayouts :many
-- Platform-wide settlement / payout rows for the admin console, most recent
-- first. Unscoped (admin-only) — authorization is enforced by AdminAuth.
SELECT * FROM payouts
ORDER BY created_at DESC
LIMIT $1;
