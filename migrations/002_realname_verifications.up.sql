CREATE TABLE IF NOT EXISTS real_name_verifications (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    real_name varchar(64) NOT NULL,
    id_card_cipher varchar(512) NOT NULL,
    status varchar(16) NOT NULL DEFAULT '待审核',
    reject_reason varchar(255) NOT NULL DEFAULT '',
    reviewed_by varchar(26) NOT NULL DEFAULT '',
    applied_at bigint NOT NULL DEFAULT 0,
    reviewed_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS real_name_verifications_user_id_unique ON real_name_verifications (user_id);
CREATE INDEX IF NOT EXISTS real_name_verifications_status_idx ON real_name_verifications (status);
