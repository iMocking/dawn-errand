-- 回滚 009_merchant_onboarding
ALTER TABLE merchants DROP COLUMN IF EXISTS verified_by;
ALTER TABLE merchants DROP COLUMN IF EXISTS verified_at;
ALTER TABLE merchants DROP COLUMN IF EXISTS submitted_at;
ALTER TABLE merchants DROP COLUMN IF EXISTS application_json;
