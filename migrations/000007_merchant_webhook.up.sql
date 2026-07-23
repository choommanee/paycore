-- 000007_merchant_webhook.up.sql
-- The Merchant Dashboard lets a merchant self-manage its outbound webhook
-- endpoint and rotate its delivery signing secret. Both live on the merchant
-- profile so /v1/me can echo the URL and PUT /v1/me/webhook can rotate the
-- secret. Only the SHA-256 HASH of the signing secret is stored — the raw
-- secret is returned exactly once and never persisted in clear (mirrors the
-- api_key_hash discipline).
ALTER TABLE merchants
    ADD COLUMN IF NOT EXISTS webhook_url         TEXT,
    ADD COLUMN IF NOT EXISTS webhook_secret_hash TEXT;
