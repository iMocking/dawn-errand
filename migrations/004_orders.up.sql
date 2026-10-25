-- 004_orders：订单生命周期、状态事件、支付单（M2）
-- 只增列/增表；down 可回滚。

ALTER TABLE orders ADD COLUMN IF NOT EXISTS client_order_no varchar(64) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS address_id varchar(26) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS address_snapshot jsonb;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS pickup_lat double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS pickup_lng double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_lat double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_lng double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_distance_km double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_fee double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS packing_fee double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_amount double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS coupon_amount double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS coupon_id varchar(26) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_status varchar(16) NOT NULL DEFAULT 'unpaid';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS accepted_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS ready_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS picked_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivered_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS canceled_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS confirmed_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS reviewed_at bigint NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS cancel_reason varchar(255) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS rider_earning double precision NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS pickup_code varchar(8) NOT NULL DEFAULT '';
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_code varchar(8) NOT NULL DEFAULT '';

-- 下单幂等：同一用户同一 client_order_no 只允许一单。
CREATE UNIQUE INDEX IF NOT EXISTS orders_user_client_no_unique
    ON orders (user_id, client_order_no) WHERE client_order_no <> '';
-- 抢单池查询。
CREATE INDEX IF NOT EXISTS orders_pool_idx ON orders (region_id, status, rider_id);

-- 状态事件。
CREATE TABLE IF NOT EXISTS order_events (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_id varchar(32) NOT NULL,
    from_status varchar(16) NOT NULL DEFAULT '',
    to_status varchar(16) NOT NULL DEFAULT '',
    actor_type varchar(16) NOT NULL DEFAULT 'system',
    actor_id varchar(26) NOT NULL DEFAULT '',
    note varchar(255) NOT NULL DEFAULT '',
    created_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS order_events_order_created_idx ON order_events (order_id, created_at);

-- 支付单。
CREATE TABLE IF NOT EXISTS payments (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_id varchar(32) NOT NULL,
    user_id varchar(26) NOT NULL DEFAULT '',
    channel varchar(16) NOT NULL DEFAULT '',
    amount double precision NOT NULL DEFAULT 0,
    status varchar(16) NOT NULL DEFAULT 'created',
    trade_no varchar(64) NOT NULL DEFAULT '',
    paid_at bigint NOT NULL DEFAULT 0,
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS payments_order_id_idx ON payments (order_id);
CREATE INDEX IF NOT EXISTS payments_trade_no_idx ON payments (trade_no);

-- 历史订单支付状态回填：已有订单视为已支付（后台演示数据）。
UPDATE orders SET payment_status = 'paid' WHERE pay_amount > 0 AND payment_status = 'unpaid';
