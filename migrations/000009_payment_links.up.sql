-- Shareable payment links. A merchant creates a link (fixed amount + allowed
-- methods); the public_id is the slug in the hosted-checkout URL (/pay/<public_id>).
-- created_by is nullable: links created via API key have no dashboard user.
CREATE TABLE payment_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    public_id       TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    amount_minor    BIGINT NOT NULL CHECK (amount_minor > 0),
    currency        TEXT NOT NULL DEFAULT 'THB',
    allowed_methods TEXT[] NOT NULL DEFAULT '{}',
    link_type       TEXT NOT NULL DEFAULT 'single_use',  -- single_use | reusable
    status          TEXT NOT NULL DEFAULT 'active',       -- active | paid | expired | disabled
    reference       TEXT NOT NULL DEFAULT '',
    image_url       TEXT NOT NULL DEFAULT '',
    expires_at      TIMESTAMPTZ,
    created_by      UUID REFERENCES merchant_users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The slug is globally unique (it addresses a checkout).
CREATE UNIQUE INDEX payment_links_public_id_idx ON payment_links (public_id);
-- Merchant's link list, newest first.
CREATE INDEX payment_links_merchant_idx ON payment_links (merchant_id, created_at DESC);
