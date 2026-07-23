-- 000003_disputes.down.sql
ALTER TABLE webhook_events
    DROP COLUMN IF EXISTS target_url,
    DROP COLUMN IF EXISTS last_error,
    DROP COLUMN IF EXISTS failed,
    DROP COLUMN IF EXISTS delivered_at;

DROP TABLE IF EXISTS recon_mismatches;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS disputes;
