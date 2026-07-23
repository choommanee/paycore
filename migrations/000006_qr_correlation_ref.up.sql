-- 000006_qr_correlation_ref.up.sql
-- Stable correlation reference for inbound QR webhooks.
--
-- Card-scheme / cross-border QR receive a provider_ref from the PSP at mint time,
-- so their webhook correlates via provider_ref. PromptPay dynamic/static QR are
-- built LOCALLY (no upstream mint), so they had no correlation key at creation —
-- their inbound webhook could not be matched back to the qr_payment row.
--
-- correlation_ref is a stable, method-agnostic key minted at creation:
--   * PromptPay  -> a generated reference embedded in the EMVCo tag-62 payload,
--                   which the bank echoes back in the webhook.
--   * card/cross-border -> the PSP provider_ref.
-- It is UNIQUE so a confirmation applies exactly once (idempotency).

ALTER TABLE qr_payments ADD COLUMN correlation_ref TEXT;

-- Backfill existing rows so the new unique index can be built: fall back to the
-- provider_ref, then the id, for any pre-existing row.
UPDATE qr_payments
   SET correlation_ref = COALESCE(provider_ref, id::text)
 WHERE correlation_ref IS NULL;

CREATE UNIQUE INDEX idx_qr_correlation_ref
    ON qr_payments (correlation_ref) WHERE correlation_ref IS NOT NULL;
