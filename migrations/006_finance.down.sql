-- 回滚 006_finance
DROP INDEX IF EXISTS refunds_status_idx;
DROP INDEX IF EXISTS refunds_order_idx;
DROP TABLE IF EXISTS refunds;

DROP INDEX IF EXISTS reviews_rider_idx;
DROP INDEX IF EXISTS reviews_merchant_idx;
DROP INDEX IF EXISTS reviews_order_unique;
DROP TABLE IF EXISTS reviews;

DROP INDEX IF EXISTS wallet_transactions_order_idx;
DROP INDEX IF EXISTS wallet_transactions_party_idx;
DROP INDEX IF EXISTS wallet_transactions_account_idx;
DROP TABLE IF EXISTS wallet_transactions;

DROP INDEX IF EXISTS wallet_accounts_party_unique;
DROP TABLE IF EXISTS wallet_accounts;

ALTER TABLE settlements DROP COLUMN IF EXISTS party_id;
ALTER TABLE settlements DROP COLUMN IF EXISTS party_type;
ALTER TABLE withdrawals DROP COLUMN IF EXISTS party_id;
