-- 000011_merchant_webhook_secret_enc.down.sql
ALTER TABLE merchants
    DROP COLUMN IF EXISTS webhook_secret_enc;
