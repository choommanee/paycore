-- internal/repository/queries/merchant.sql
-- Merchant onboarding and lookup. The raw API key is never stored — only its
-- hash (api_key_hash). Auth resolves a merchant by hashing the presented key and
-- matching it here. All parameters are bound ($1, $2, ...).

-- name: CreateMerchant :one
INSERT INTO merchants (
    id, name, mcc, settlement_currency, api_key_hash
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMerchant :one
SELECT * FROM merchants WHERE id = $1;

-- name: GetMerchantByAPIKeyHash :one
SELECT * FROM merchants
WHERE api_key_hash = $1 AND status = 'active';

-- name: RotateMerchantAPIKey :one
-- Issues a new API key by replacing the stored hash. The previous hash no
-- longer resolves, so the old key is immediately invalidated. The raw key is
-- generated and returned by the service layer; only its hash is persisted here.
UPDATE merchants
SET api_key_hash = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SetMerchantWebhook :one
-- Sets the merchant's outbound webhook URL and rotates its delivery signing
-- secret. Only the hash of the signing secret is stored; the raw secret is
-- returned once by the service layer.
UPDATE merchants
SET webhook_url = $2,
    webhook_secret_hash = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
