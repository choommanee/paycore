-- internal/repository/queries/qr.sql
-- QR payment data access. QR is an async push model: mint on create, then the
-- bank/PSP confirms via a signed webhook. All parameters are bound ($1, $2, ...)
-- so there is no SQL-injection surface.

-- name: CreateQRPayment :one
INSERT INTO qr_payments (
    id, merchant_id, method, amount_minor, currency, status,
    reference, qr_payload, qr_image_url, provider_ref, correlation_ref, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetQRPayment :one
SELECT * FROM qr_payments WHERE id = $1;

-- name: GetQRPaymentWithMerchant :one
-- Sandbox display fetch: a single QR payment joined to the owning merchant's
-- name so the payer simulator can show "who is charging". Unauthenticated
-- (sandbox only); returns display fields plus the correlation/provider handles
-- the server needs to sign a confirmation, never any secret.
SELECT q.id, q.merchant_id, m.name AS merchant_name, q.method,
       q.amount_minor, q.currency, q.status, q.reference,
       q.provider_ref, q.correlation_ref, q.expires_at, q.paid_at, q.created_at
FROM qr_payments q
JOIN merchants m ON m.id = q.merchant_id
WHERE q.id = $1;

-- name: GetQRPaymentForMerchant :one
-- Merchant-scoped fetch: only returns the row when it belongs to the
-- authenticated merchant (returns ErrNoRows -> 404 for another merchant's id).
SELECT * FROM qr_payments WHERE id = $1 AND merchant_id = $2;

-- name: GetQRByProviderRef :one
SELECT * FROM qr_payments WHERE provider_ref = $1;

-- name: ListPendingQRPayments :many
-- Recent QR payments in a given status, joined to the owning merchant's name,
-- most recent first. Powers the sandbox simulator's "incoming requests" list.
-- Unauthenticated (sandbox only); returns display fields, never secrets.
SELECT q.id, q.merchant_id, m.name AS merchant_name, q.method,
       q.amount_minor, q.currency, q.status, q.reference, q.created_at
FROM qr_payments q
JOIN merchants m ON m.id = q.merchant_id
WHERE q.status = $1
ORDER BY q.created_at DESC
LIMIT $2;

-- name: GetQRByCorrelationRef :one
-- Correlates an inbound webhook back to the minted QR via the stable
-- correlation_ref. Used for PromptPay, whose webhook echoes the reference we
-- embedded in the EMVCo tag-62 payload at creation.
SELECT * FROM qr_payments WHERE correlation_ref = $1;

-- name: UpdateQRStatus :one
UPDATE qr_payments
SET status = $2,
    payer_bank = $3,
    provider_ref = $4,
    paid_at = $5
WHERE id = $1
RETURNING *;

-- name: CreateWebhookEvent :exec
INSERT INTO webhook_events (
    id, merchant_id, event_type, payload, next_attempt_at, created_at
) VALUES ($1, $2, $3, $4, NOW(), NOW());
