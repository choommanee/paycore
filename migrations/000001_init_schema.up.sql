-- 000001_init_schema.up.sql
-- Core schema for the acquiring payment gateway.
-- Design notes:
--   * Money is stored as BIGINT minor units (satang) + currency, never float.
--   * No full PAN / CVV is ever stored. Only card_brand + card_last4 for display.
--   * ledger_entries is append-only (double-entry style) and is the source of truth
--     for settlement & reconciliation.
--   * Cardholder data lives in a separate PCI-scoped vault schema/service, encrypted
--     with an HSM/KMS-managed key — not in this operational database.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Merchants onboarded to the gateway.
CREATE TABLE merchants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active',   -- active | suspended
    api_key_hash  TEXT NOT NULL,                    -- store hash, never the raw key
    mcc           TEXT,                             -- merchant category code
    settlement_currency TEXT NOT NULL DEFAULT 'THB',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payment aggregate (auth/capture lifecycle).
CREATE TABLE payments (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id           UUID NOT NULL REFERENCES merchants(id),
    amount_minor          BIGINT NOT NULL CHECK (amount_minor > 0),
    captured_amount_minor BIGINT NOT NULL DEFAULT 0,
    refunded_amount_minor BIGINT NOT NULL DEFAULT 0,
    currency              TEXT NOT NULL,
    status                TEXT NOT NULL,            -- requires_action|authorized|captured|partial_refunded|refunded|voided|failed
    card_brand            TEXT,
    card_last4            CHAR(4),
    acquirer_ref          TEXT,
    auth_code             TEXT,
    reference             TEXT,
    idempotency_key       TEXT NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (merchant_id, idempotency_key)
);
CREATE INDEX idx_payments_merchant ON payments (merchant_id, created_at DESC);
CREATE INDEX idx_payments_status   ON payments (status);

-- Append-only ledger (double-entry). Never UPDATE/DELETE rows here.
CREATE TABLE ledger_entries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id   UUID NOT NULL REFERENCES payments(id),
    entry_type   TEXT NOT NULL,                     -- authorize|capture|refund|void|fee|settlement
    amount_minor BIGINT NOT NULL,
    currency     TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ledger_payment ON ledger_entries (payment_id);

-- Refunds (a payment can have many).
CREATE TABLE refunds (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id   UUID NOT NULL REFERENCES payments(id),
    amount_minor BIGINT NOT NULL CHECK (amount_minor > 0),
    currency     TEXT NOT NULL,
    reason       TEXT,
    acquirer_ref TEXT,
    status       TEXT NOT NULL DEFAULT 'pending',   -- pending|succeeded|failed
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_refunds_payment ON refunds (payment_id);

-- Outbound webhook delivery (at-least-once, with retry).
CREATE TABLE webhook_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id    UUID NOT NULL REFERENCES merchants(id),
    event_type     TEXT NOT NULL,                   -- payment.captured, payment.refunded, ...
    payload        JSONB NOT NULL,
    delivered      BOOLEAN NOT NULL DEFAULT FALSE,
    attempts       INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_webhook_undelivered ON webhook_events (delivered, next_attempt_at);

-- Immutable audit trail (required for BOT/PCI). Append-only.
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor       TEXT NOT NULL,                      -- merchant id / user / system
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   TEXT NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_entity ON audit_log (entity_type, entity_id);
