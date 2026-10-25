-- 006_finance：钱包账本、评价、售后、结算/提现主体化（M4）
CREATE TABLE IF NOT EXISTS wallet_accounts (
    id varchar(26) NOT NULL PRIMARY KEY,
    party_type varchar(16) NOT NULL,
    party_id varchar(26) NOT NULL,
    balance double precision NOT NULL DEFAULT 0,
    frozen double precision NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS wallet_accounts_party_unique ON wallet_accounts (party_type, party_id);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    id varchar(26) NOT NULL PRIMARY KEY,
    account_id varchar(26) NOT NULL,
    party_type varchar(16) NOT NULL DEFAULT '',
    party_id varchar(26) NOT NULL DEFAULT '',
    order_id varchar(32) NOT NULL DEFAULT '',
    type varchar(32) NOT NULL,
    amount double precision NOT NULL DEFAULT 0,
    balance_after double precision NOT NULL DEFAULT 0,
    note varchar(255) NOT NULL DEFAULT '',
    created_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS wallet_transactions_account_idx ON wallet_transactions (account_id, created_at);
CREATE INDEX IF NOT EXISTS wallet_transactions_party_idx ON wallet_transactions (party_type, party_id, created_at);
CREATE INDEX IF NOT EXISTS wallet_transactions_order_idx ON wallet_transactions (order_id);

CREATE TABLE IF NOT EXISTS reviews (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_id varchar(32) NOT NULL,
    user_id varchar(26) NOT NULL,
    merchant_id varchar(26) NOT NULL DEFAULT '',
    rider_id varchar(26) NOT NULL DEFAULT '',
    rating integer NOT NULL DEFAULT 5,
    tags jsonb,
    content varchar(1024) NOT NULL DEFAULT '',
    images jsonb,
    reply varchar(1024) NOT NULL DEFAULT '',
    replied_at bigint NOT NULL DEFAULT 0,
    created_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS reviews_order_unique ON reviews (order_id);
CREATE INDEX IF NOT EXISTS reviews_merchant_idx ON reviews (merchant_id, created_at);
CREATE INDEX IF NOT EXISTS reviews_rider_idx ON reviews (rider_id, created_at);

CREATE TABLE IF NOT EXISTS refunds (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_id varchar(32) NOT NULL,
    user_id varchar(26) NOT NULL DEFAULT '',
    merchant_id varchar(26) NOT NULL DEFAULT '',
    amount double precision NOT NULL DEFAULT 0,
    reason varchar(512) NOT NULL DEFAULT '',
    status varchar(16) NOT NULL DEFAULT 'requested',
    prev_status varchar(16) NOT NULL DEFAULT '',
    operator varchar(26) NOT NULL DEFAULT '',
    reject_reason varchar(255) NOT NULL DEFAULT '',
    created_at bigint NOT NULL DEFAULT 0,
    handled_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS refunds_order_idx ON refunds (order_id);
CREATE INDEX IF NOT EXISTS refunds_status_idx ON refunds (status, created_at);

ALTER TABLE withdrawals ADD COLUMN IF NOT EXISTS party_id varchar(26) NOT NULL DEFAULT '';
ALTER TABLE settlements ADD COLUMN IF NOT EXISTS party_type varchar(16) NOT NULL DEFAULT '';
ALTER TABLE settlements ADD COLUMN IF NOT EXISTS party_id varchar(26) NOT NULL DEFAULT '';

-- 历史提现单：骑手编号即 rider_id；商家暂无法回填，后续由管理端补。
UPDATE withdrawals SET party_id = party_no WHERE party_id = '' AND party_type = 'rider' AND party_no <> '';
