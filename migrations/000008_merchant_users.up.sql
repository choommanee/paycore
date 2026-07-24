-- Human identities that log into the merchant dashboard. Distinct from the
-- API-key credential on merchants: one merchant may later have several users.
-- Login is via OAuth (Google) or 'dev' (sandbox only); we store the provider
-- subject, never a password.
CREATE TABLE merchant_users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id    UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    email          TEXT NOT NULL,
    name           TEXT NOT NULL DEFAULT '',
    avatar_url     TEXT NOT NULL DEFAULT '',
    oauth_provider TEXT NOT NULL,            -- 'google' | 'dev'
    oauth_subject  TEXT NOT NULL,            -- provider 'sub'
    role           TEXT NOT NULL DEFAULT 'owner',
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A provider identity maps to exactly one user (idempotent login/signup).
CREATE UNIQUE INDEX merchant_users_oauth_idx
    ON merchant_users (oauth_provider, oauth_subject);

-- Fast lookup of all users for a merchant.
CREATE INDEX merchant_users_merchant_idx ON merchant_users (merchant_id);
