-- 008_payment_hardening：支付流水号唯一（M5）
-- 支付回调幂等兜底：非空 trade_no 唯一。
CREATE UNIQUE INDEX IF NOT EXISTS payments_trade_no_unique
    ON payments (trade_no) WHERE trade_no <> '';
