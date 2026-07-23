-- 000003_disputes.up.sql
-- Ops tables added for the Merchant/Admin API and background workers:
--   * disputes       — chargeback / dispute lifecycle state machine
--   * payouts        — settlement worker aggregates captured payments per merchant
--   * recon_mismatches — reconciliation worker flags ledger vs acquirer report deltas
-- Also extends webhook_events with delivery bookkeeping columns used by the
-- outbound webhook delivery worker (delivered_at / failed / last_error).

-- Chargebacks / disputes. State machine:
--   opened -> under_review -> won | lost | accepted
-- A dispute is always attached to exactly one payment.
CREATE TABLE disputes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id    UUID NOT NULL REFERENCES payments(id),
    merchant_id   UUID NOT NULL REFERENCES merchants(id),
    amount_minor  BIGINT NOT NULL CHECK (amount_minor > 0),
    currency      TEXT NOT NULL,
    reason        TEXT,                            -- network reason code / free text
    status        TEXT NOT NULL DEFAULT 'opened',  -- opened|under_review|won|lost|accepted
    evidence      JSONB,                           -- merchant-submitted evidence bundle
    network_ref   TEXT,                            -- card-network case id
    opened_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_disputes_payment  ON disputes (payment_id);
CREATE INDEX idx_disputes_merchant ON disputes (merchant_id, created_at DESC);
CREATE INDEX idx_disputes_status   ON disputes (status);

-- Payouts. The settlement worker aggregates captured (net of refunded) payment
-- amounts per merchant/currency into a single payout row per run window.
CREATE TABLE payouts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id),
    currency        TEXT NOT NULL,
    gross_minor     BIGINT NOT NULL,               -- sum captured
    refunded_minor  BIGINT NOT NULL DEFAULT 0,      -- sum refunded in window
    fee_minor       BIGINT NOT NULL DEFAULT 0,
    net_minor       BIGINT NOT NULL,               -- gross - refunded - fee
    payment_count   INT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',-- pending|paid|failed
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payouts_merchant ON payouts (merchant_id, created_at DESC);

-- Reconciliation mismatches. The reconciliation worker compares internal ledger
-- totals against a (mock) acquirer settlement report and records deltas here for
-- ops follow-up.
CREATE TABLE recon_mismatches (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id    UUID REFERENCES merchants(id),
    currency       TEXT NOT NULL,
    ledger_minor   BIGINT NOT NULL,               -- our captured total
    acquirer_minor BIGINT NOT NULL,               -- acquirer-reported total
    delta_minor    BIGINT NOT NULL,               -- ledger - acquirer
    report_date    DATE NOT NULL,
    resolved       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_recon_open ON recon_mismatches (resolved, report_date);

-- Outbound webhook delivery bookkeeping used by the delivery worker.
ALTER TABLE webhook_events
    ADD COLUMN IF NOT EXISTS delivered_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS failed       BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS last_error   TEXT,
    ADD COLUMN IF NOT EXISTS target_url   TEXT;
