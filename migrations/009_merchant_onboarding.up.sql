-- 009_merchant_onboarding：商家入驻完整资料与认证信息（地推闭环）
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS application_json jsonb;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS submitted_at bigint NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS verified_at bigint NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS verified_by varchar(26) NOT NULL DEFAULT '';
