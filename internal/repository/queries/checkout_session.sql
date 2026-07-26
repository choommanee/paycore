-- internal/repository/queries/checkout_session.sql
-- Hosted-checkout sessions. A session is ALWAYS resolved by its token hash (the
-- token is the credential); the merchant scope comes from the resolved row, not
-- from the caller. Updates are by primary key (already resolved from the hash).

-- name: CreateCheckoutSession :one
INSERT INTO checkout_sessions (
    id, merchant_id, payment_link_id, session_token_hash, amount_minor, currency,
    status, selected_method, payment_id, qr_payment_id, customer_email, return_url, expires_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING *;

-- name: GetCheckoutSessionByTokenHash :one
SELECT * FROM checkout_sessions WHERE session_token_hash = $1;

-- name: UpdateCheckoutSession :one
UPDATE checkout_sessions
SET status = $2, selected_method = $3, payment_id = $4, qr_payment_id = $5,
    customer_email = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;
