-- 000004_merchant_api_key_hash_index.up.sql
-- API-key auth runs on the hot path of EVERY payment/QR request:
--   GetMerchantByAPIKeyHash -> SELECT ... WHERE api_key_hash = $1 AND status = 'active'
-- Without a supporting index this seq-scans merchants on every request, which is
-- O(n) per request and becomes a real bottleneck at merchant scale.
--
-- A partial UNIQUE index on active merchants both accelerates the lookup to an
-- index scan and enforces that no two active merchants share an api_key_hash.
-- (Not created CONCURRENTLY so it runs inside golang-migrate's per-file
-- transaction; the merchants table is small enough that the brief lock is fine.)
CREATE UNIQUE INDEX IF NOT EXISTS idx_merchants_api_key_hash_active
    ON merchants (api_key_hash)
    WHERE status = 'active';
