-- 007_coupons：用户优惠券（M5）
CREATE TABLE IF NOT EXISTS user_coupons (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    merchant_id varchar(26) NOT NULL DEFAULT '',
    title varchar(64) NOT NULL,
    amount double precision NOT NULL DEFAULT 0,
    min_spend double precision NOT NULL DEFAULT 0,
    status varchar(16) NOT NULL DEFAULT 'unused',
    expires_at bigint NOT NULL DEFAULT 0,
    used_order_id varchar(32) NOT NULL DEFAULT '',
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS user_coupons_user_status_idx ON user_coupons (user_id, status, expires_at);
CREATE INDEX IF NOT EXISTS user_coupons_order_idx ON user_coupons (used_order_id);
