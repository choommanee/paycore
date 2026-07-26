-- internal/repository/queries/payment_link.sql
-- Payment links CRUD. Reads/updates are ALWAYS scoped by merchant_id to prevent
-- one merchant reading/altering another's link (IDOR); the public_id lookup is
-- the only unscoped read and is used by the (Phase 3) public checkout page.

-- name: CreatePaymentLink :one
INSERT INTO payment_links (
    id, merchant_id, public_id, title, description, amount_minor, currency,
    allowed_methods, link_type, status, reference, image_url, expires_at, created_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
RETURNING *;

-- name: GetPaymentLink :one
SELECT * FROM payment_links WHERE id = $1 AND merchant_id = $2;

-- name: GetPaymentLinkByPublicID :one
SELECT * FROM payment_links WHERE public_id = $1;

-- name: ListPaymentLinksByMerchant :many
SELECT * FROM payment_links
WHERE merchant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePaymentLinkStatus :one
UPDATE payment_links
SET status = $3, updated_at = NOW()
WHERE id = $1 AND merchant_id = $2
RETURNING *;

-- name: ConsumePaymentLinkIfActive :one
-- Atomically flips an active link to 'paid'. Returns no row if the link is not
-- active (already paid/disabled/expired) — the caller uses that to prevent
-- double payment of a single_use link.
UPDATE payment_links SET status = 'paid', updated_at = NOW()
WHERE id = $1 AND merchant_id = $2 AND status = 'active'
RETURNING *;
