-- 001_init.down.sql — 回滚 001_init.up.sql（逆序删表）。
BEGIN;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS casbin_rules;
DROP TABLE IF EXISTS mail_outboxes;
DROP TABLE IF EXISTS fee_rules;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS settlements;
DROP TABLE IF EXISTS withdrawals;
DROP TABLE IF EXISTS complaints;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS riders;
DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS regions;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS users;

COMMIT;
