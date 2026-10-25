-- 003_app_surface：商家菜单、用户地址簿、门店经营设置（M1）
-- 说明：只增表/增列，不改既有列语义；down 可回滚。

-- ---------- 菜品分类 ----------
CREATE TABLE IF NOT EXISTS dish_categories (
    id varchar(26) NOT NULL PRIMARY KEY,
    merchant_id varchar(26) NOT NULL,
    name varchar(64) NOT NULL,
    sort integer NOT NULL DEFAULT 0,
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS dish_categories_merchant_sort_idx ON dish_categories (merchant_id, sort);

-- ---------- 菜品（软删除） ----------
CREATE TABLE IF NOT EXISTS dishes (
    id varchar(26) NOT NULL PRIMARY KEY,
    merchant_id varchar(26) NOT NULL,
    category_id varchar(26) NOT NULL,
    name varchar(128) NOT NULL,
    description varchar(512) NOT NULL DEFAULT '',
    image_url varchar(512) NOT NULL DEFAULT '',
    price double precision NOT NULL DEFAULT 0,
    stock integer NOT NULL DEFAULT 0,
    on_sale boolean NOT NULL DEFAULT true,
    monthly_sales bigint NOT NULL DEFAULT 0,
    sort integer NOT NULL DEFAULT 0,
    deleted_at bigint NOT NULL DEFAULT 0,
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS dishes_merchant_on_sale_idx ON dishes (merchant_id, on_sale);
CREATE INDEX IF NOT EXISTS dishes_category_id_idx ON dishes (category_id);
CREATE INDEX IF NOT EXISTS dishes_name_idx ON dishes (name);

-- ---------- 用户收货地址 ----------
CREATE TABLE IF NOT EXISTS user_addresses (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    contact_name varchar(64) NOT NULL DEFAULT '',
    phone varchar(32) NOT NULL DEFAULT '',
    region_id varchar(26) NOT NULL DEFAULT '',
    detail varchar(255) NOT NULL,
    tag varchar(16) NOT NULL DEFAULT '',
    lat double precision NOT NULL DEFAULT 0,
    lng double precision NOT NULL DEFAULT 0,
    is_default boolean NOT NULL DEFAULT false,
    created_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS user_addresses_user_created_idx ON user_addresses (user_id, created_at);
-- 每个用户最多一个默认地址（部分唯一索引）。
CREATE UNIQUE INDEX IF NOT EXISTS user_addresses_one_default_idx
    ON user_addresses (user_id) WHERE is_default = true;

-- ---------- 门店经营设置 ----------
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS lat double precision NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS lng double precision NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS auto_accept boolean NOT NULL DEFAULT false;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS min_price double precision NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS delivery_fee double precision NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS packing_fee double precision NOT NULL DEFAULT 0;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS announcement varchar(512) NOT NULL DEFAULT '';
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS service_phone varchar(32) NOT NULL DEFAULT '';

-- ---------- 账号与主体一对一约束（防止并发重复开通） ----------
CREATE UNIQUE INDEX IF NOT EXISTS merchants_owner_user_id_unique
    ON merchants (owner_user_id) WHERE owner_user_id <> '';
CREATE UNIQUE INDEX IF NOT EXISTS riders_user_id_unique
    ON riders (user_id) WHERE user_id <> '';
