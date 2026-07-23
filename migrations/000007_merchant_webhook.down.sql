-- 000007_merchant_webhook.down.sql
ALTER TABLE merchants
    DROP COLUMN IF EXISTS webhook_url,
    DROP COLUMN IF EXISTS webhook_secret_hash;
