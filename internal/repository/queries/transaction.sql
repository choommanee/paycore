-- internal/repository/queries/transaction.sql
-- Unified transaction feed for the merchant dashboard: a UNION ALL over the
-- three payment-activity tables (card, PromptPay, wallet), each already
-- storing money as amount_minor BIGINT + currency (no decimal conversion
-- needed). Avoiding double-counting:
--   * payments        -> source 'card'      (every card payment).
--   * qr_payments     -> source 'promptpay' (every QR payment).
--   * checkout_sessions -> source 'wallet', ONLY rows that are paid AND whose
--     selected_method is a wallet method. Card/PromptPay checkout sessions
--     already surface via payments/qr_payments above; wallet mocks never
--     create a payments/qr_payments row, so this is the only place they
--     appear.
-- All parameters are bound ($1..$5) so there is no SQL-injection surface. The
-- merchant id is repeated as $1/$2/$3 (one per UNIONed branch, all bound to
-- the SAME value by the caller) rather than reused positionally: sqlc's
-- analyzer resolves a shared parameter against every branch's FROM clause at
-- once and misreports merchant_id as ambiguous when the same ordinal is reused
-- across a UNION, even though plain Postgres accepts it (verified via
-- PREPARE). Distinct ordinals per branch sidesteps that false positive.

-- name: ListTransactionsByMerchant :many
SELECT payments.id, 'card'::text AS source, 'card'::text AS method,
       payments.amount_minor, payments.currency, payments.status, payments.reference, payments.created_at
FROM payments
WHERE payments.merchant_id = $1

UNION ALL

SELECT qr_payments.id, 'promptpay'::text AS source, qr_payments.method,
       qr_payments.amount_minor, qr_payments.currency, qr_payments.status, qr_payments.reference, qr_payments.created_at
FROM qr_payments
WHERE qr_payments.merchant_id = $2

UNION ALL

SELECT checkout_sessions.id, 'wallet'::text AS source, checkout_sessions.selected_method AS method,
       checkout_sessions.amount_minor, checkout_sessions.currency, checkout_sessions.status, ''::text AS reference, checkout_sessions.created_at
FROM checkout_sessions
WHERE checkout_sessions.merchant_id = $3
  AND checkout_sessions.status = 'paid'
  AND checkout_sessions.selected_method IN ('mobile_banking', 'truemoney', 'shopeepay', 'alipay', 'wechat', 'card_installment')

ORDER BY created_at DESC
LIMIT $4 OFFSET $5;
