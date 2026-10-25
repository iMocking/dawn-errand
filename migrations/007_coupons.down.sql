-- 回滚 007_coupons
DROP INDEX IF EXISTS user_coupons_order_idx;
DROP INDEX IF EXISTS user_coupons_user_status_idx;
DROP TABLE IF EXISTS user_coupons;
