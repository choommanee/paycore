-- 000002_qr_payments.up.sql
-- QR payments (PromptPay dynamic/static, card-scheme QR, cross-border).
-- QR is an async push model: the payer pushes funds and the bank/PSP confirms
-- via webhook. There is no authorize/capture split.

CREATE TABLE qr_payments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id   UUID NOT NULL REFERENCES merchants(id),
    method        TEXT NOT NULL,            -- promptpay_dynamic|promptpay_static|card_scheme|cross_border
    amount_minor  BIGINT NOT NULL CHECK (amount_minor > 0),
    currency      TEXT NOT NULL,
    status        TEXT NOT NULL,            -- awaiting_payment|paid|expired|failed
    reference     TEXT NOT NULL,            -- merchant ref, reconciled against the webhook
    qr_payload    TEXT NOT NULL,            -- EMVCo string rendered as the QR image
    qr_image_url  TEXT,
    payer_bank    TEXT,                     -- filled on confirmation
    provider_ref  TEXT,                     -- bank/PSP txn id (idempotency)
    expires_at    TIMESTAMPTZ NOT NULL,
    paid_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (merchant_id, reference)
);
CREATE INDEX idx_qr_status  ON qr_payments (status, expires_at);
CREATE INDEX idx_qr_ref     ON qr_payments (reference);

-- Idempotency guard for inbound confirmations: a provider_ref is applied once.
CREATE UNIQUE INDEX idx_qr_provider_ref
    ON qr_payments (provider_ref) WHERE provider_ref IS NOT NULL;
