-- A single hosted-checkout payment attempt. Created from a payment_link (or
-- directly via API); the opaque session token is given to the browser and only
-- its SHA-256 hash is stored here (like api_key_hash). The token is the sole
-- credential for the public /v1/checkout routes; merchant context is derived
-- from this row, never from the request. Short-lived (expires_at ~30 min).
CREATE TABLE checkout_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    payment_link_id     UUID REFERENCES payment_links(id) ON DELETE SET NULL,
    session_token_hash  TEXT NOT NULL,
    amount_minor        BIGINT NOT NULL CHECK (amount_minor > 0),
    currency            TEXT NOT NULL DEFAULT 'THB',
    status              TEXT NOT NULL DEFAULT 'open',  -- open | processing | requires_action | paid | failed | expired
    selected_method     TEXT NOT NULL DEFAULT '',
    payment_id          UUID REFERENCES payments(id) ON DELETE SET NULL,
    qr_payment_id       UUID REFERENCES qr_payments(id) ON DELETE SET NULL,
    customer_email      TEXT NOT NULL DEFAULT '',
    return_url          TEXT NOT NULL DEFAULT '',
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The token hash is the lookup key for every public checkout request.
CREATE UNIQUE INDEX checkout_sessions_token_hash_idx ON checkout_sessions (session_token_hash);
-- A merchant's recent sessions (ops / debugging), newest first.
CREATE INDEX checkout_sessions_merchant_idx ON checkout_sessions (merchant_id, created_at DESC);
