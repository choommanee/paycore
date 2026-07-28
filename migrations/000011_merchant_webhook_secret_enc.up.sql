-- 000011_merchant_webhook_secret_enc.up.sql
-- Per-merchant webhook signing (Stripe-style key isolation). The outbound
-- delivery worker must sign each delivery with the MERCHANT'S OWN secret so the
-- merchant can verify with the whsec_… returned from SetWebhook. That requires
-- storing the secret RETRIEVABLY, not as a one-way hash — so add an encrypted
-- column holding the envelope-encrypted secret (SecretBox: per-record DEK
-- wrapped by the KMS). The existing webhook_secret_hash column is kept for now.
ALTER TABLE merchants
    ADD COLUMN IF NOT EXISTS webhook_secret_enc BYTEA;
