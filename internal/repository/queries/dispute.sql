-- internal/repository/queries/dispute.sql
-- Chargeback / dispute data access. State machine transitions are enforced in
-- the service layer; these queries persist the resulting rows. All parameters
-- are bound ($1, $2, ...).

-- name: CreateDispute :one
INSERT INTO disputes (
    id, payment_id, merchant_id, amount_minor, currency, reason, status, network_ref
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetDispute :one
SELECT * FROM disputes WHERE id = $1;

-- name: ListDisputesByPayment :many
SELECT * FROM disputes
WHERE payment_id = $1
ORDER BY created_at DESC;

-- name: ListDisputesByMerchant :many
SELECT * FROM disputes
WHERE merchant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateDisputeStatus :one
UPDATE disputes
SET status = $2,
    evidence = COALESCE($3, evidence),
    resolved_at = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
