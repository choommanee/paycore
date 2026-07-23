-- internal/repository/queries/payment.sql
-- sqlc generates type-safe Go from these queries. All parameters are bound
-- ($1, $2, ...) so there is no SQL-injection surface.

-- name: CreatePayment :one
INSERT INTO payments (
    id, merchant_id, amount_minor, currency, status,
    card_brand, card_last4, reference, idempotency_key
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = $1;

-- name: GetPaymentForMerchant :one
-- Merchant-scoped fetch: only returns the row when it belongs to the
-- authenticated merchant, avoiding a cross-tenant existence oracle (returns
-- ErrNoRows -> 404 when the id belongs to another merchant).
SELECT * FROM payments WHERE id = $1 AND merchant_id = $2;

-- name: GetByIdempotencyKey :one
SELECT * FROM payments
WHERE merchant_id = $1 AND idempotency_key = $2;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    captured_amount_minor = $3,
    refunded_amount_minor = $4,
    acquirer_ref = $5,
    auth_code = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InsertLedgerEntry :exec
INSERT INTO ledger_entries (
    id, payment_id, entry_type, amount_minor, currency, created_at
) VALUES ($1, $2, $3, $4, $5, NOW());

-- name: WriteAuditLog :exec
INSERT INTO audit_log (
    id, actor, action, entity_type, entity_id, metadata, created_at
) VALUES ($1, $2, $3, $4, $5, $6, NOW());

-- name: ListPaymentsByMerchant :many
SELECT * FROM payments
WHERE merchant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
