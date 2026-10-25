-- 回滚 004_orders
DROP INDEX IF EXISTS payments_trade_no_idx;
DROP INDEX IF EXISTS payments_order_id_idx;
DROP TABLE IF EXISTS payments;

DROP INDEX IF EXISTS order_events_order_created_idx;
DROP TABLE IF EXISTS order_events;

DROP INDEX IF EXISTS orders_pool_idx;
DROP INDEX IF EXISTS orders_user_client_no_unique;

ALTER TABLE orders DROP COLUMN IF EXISTS delivery_code;
ALTER TABLE orders DROP COLUMN IF EXISTS pickup_code;
ALTER TABLE orders DROP COLUMN IF EXISTS rider_earning;
ALTER TABLE orders DROP COLUMN IF EXISTS cancel_reason;
ALTER TABLE orders DROP COLUMN IF EXISTS reviewed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS confirmed_at;
ALTER TABLE orders DROP COLUMN IF EXISTS canceled_at;
ALTER TABLE orders DROP COLUMN IF EXISTS delivered_at;
ALTER TABLE orders DROP COLUMN IF EXISTS picked_at;
ALTER TABLE orders DROP COLUMN IF EXISTS ready_at;
ALTER TABLE orders DROP COLUMN IF EXISTS accepted_at;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS coupon_id;
ALTER TABLE orders DROP COLUMN IF EXISTS coupon_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS packing_fee;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_fee;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_distance_km;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_lng;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_lat;
ALTER TABLE orders DROP COLUMN IF EXISTS pickup_lng;
ALTER TABLE orders DROP COLUMN IF EXISTS pickup_lat;
ALTER TABLE orders DROP COLUMN IF EXISTS address_snapshot;
ALTER TABLE orders DROP COLUMN IF EXISTS address_id;
ALTER TABLE orders DROP COLUMN IF EXISTS client_order_no;
