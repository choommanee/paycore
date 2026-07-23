-- internal/repository/queries/admin.sql
-- Platform (admin) read models behind X-Admin-Key. These are intentionally
-- unscoped by merchant — they power the operator console. All parameters are
-- bound ($1, $2, ...) so there is no SQL-injection surface.

-- name: ListAllMerchants :many
SELECT * FROM merchants
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLog :many
SELECT * FROM audit_log
ORDER BY created_at DESC
LIMIT $1;

-- name: ListAllDisputes :many
SELECT * FROM disputes
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: PlatformStats :one
-- Platform-wide key risk indicators across every merchant. Volume is captured
-- (net-of-refund) minor units. merchant_count is a cheap correlated count.
SELECT
    COUNT(*)::BIGINT AS payment_count,
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
    ), 0)::BIGINT AS failed_count,
    (SELECT COUNT(*) FROM merchants)::BIGINT AS merchant_count,
    (SELECT COUNT(*) FROM disputes)::BIGINT  AS dispute_count
FROM payments;
