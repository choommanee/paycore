-- 000006_qr_correlation_ref.down.sql
DROP INDEX IF EXISTS idx_qr_correlation_ref;
ALTER TABLE qr_payments DROP COLUMN IF EXISTS correlation_ref;
