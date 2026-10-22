-- 001_init.up.sql — 曙光跑腿 SaaS 初始化表结构（对齐 ent schema，PostgreSQL 16）。
-- 由 cmd/migrate 执行（幂等，通过 schema_migrations 记录）。
BEGIN;

CREATE TABLE IF NOT EXISTS users (
    id varchar(26) NOT NULL PRIMARY KEY,
    email varchar(255) NOT NULL,
    phone varchar(32) NOT NULL DEFAULT '',
    name varchar(64) NOT NULL,
    avatar varchar(512) NOT NULL DEFAULT '',
    status varchar(16) NOT NULL DEFAULT 'active',
    email_verified boolean NOT NULL DEFAULT false,
    role varchar(32) NOT NULL DEFAULT 'user',
    region_id varchar(26) NOT NULL DEFAULT '',
    merchant_id varchar(26) NOT NULL DEFAULT '',
    rider_id varchar(26) NOT NULL DEFAULT '',
    registered_at bigint NOT NULL,
    updated_at bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (email);
CREATE INDEX IF NOT EXISTS users_phone_idx ON users (phone);
CREATE INDEX IF NOT EXISTS users_status_idx ON users (status);
CREATE INDEX IF NOT EXISTS users_region_id_idx ON users (region_id);
CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

CREATE TABLE IF NOT EXISTS credentials (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    password_hash varchar(255) NOT NULL,
    session_version bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS credentials_user_id_unique ON credentials (user_id);

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    token_hash varchar(64) NOT NULL,
    expires_at bigint NOT NULL DEFAULT 0,
    used_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS email_verification_tokens_token_hash_unique ON email_verification_tokens (token_hash);
CREATE INDEX IF NOT EXISTS email_verification_tokens_user_id_idx ON email_verification_tokens (user_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    token_hash varchar(64) NOT NULL,
    expires_at bigint NOT NULL DEFAULT 0,
    used_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS password_reset_tokens_token_hash_unique ON password_reset_tokens (token_hash);
CREATE INDEX IF NOT EXISTS password_reset_tokens_user_id_idx ON password_reset_tokens (user_id);

CREATE TABLE IF NOT EXISTS regions (
    id varchar(26) NOT NULL PRIMARY KEY,
    name varchar(64) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'open',
    villages bigint NOT NULL DEFAULT 0,
    radius_km double precision NOT NULL DEFAULT 0,
    start_fee double precision NOT NULL DEFAULT 0,
    free_km double precision NOT NULL DEFAULT 0,
    notice varchar(255) NOT NULL DEFAULT '',
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS regions_name_unique ON regions (name);

CREATE TABLE IF NOT EXISTS system_settings (
    id varchar(26) NOT NULL PRIMARY KEY,
    key varchar(64) NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'disabled',
    value_json jsonb,
    updated_by varchar(26) NOT NULL DEFAULT '',
    updated_at bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS system_settings_key_unique ON system_settings (key);

CREATE TABLE IF NOT EXISTS user_profiles (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    level varchar(64) NOT NULL DEFAULT 'Lv.1 新用户',
    nickname varchar(64) NOT NULL DEFAULT '',
    real_name varchar(64) NOT NULL DEFAULT '',
    verified varchar(16) NOT NULL DEFAULT '未实名',
    region varchar(128) NOT NULL DEFAULT '',
    address varchar(255) NOT NULL DEFAULT '',
    register_channel varchar(128) NOT NULL DEFAULT '',
    biz_types varchar(128) NOT NULL DEFAULT '',
    balance double precision NOT NULL DEFAULT 0,
    red_packet_balance double precision NOT NULL DEFAULT 0,
    recent_active_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS user_profiles_user_id_unique ON user_profiles (user_id);

CREATE TABLE IF NOT EXISTS merchants (
    id varchar(26) NOT NULL PRIMARY KEY,
    region_id varchar(26) NOT NULL,
    owner_user_id varchar(26) NOT NULL DEFAULT '',
    name varchar(128) NOT NULL,
    category varchar(64) NOT NULL DEFAULT '',
    address varchar(255) NOT NULL DEFAULT '',
    address_sub varchar(128) NOT NULL DEFAULT '',
    contact varchar(64) NOT NULL DEFAULT '',
    phone varchar(32) NOT NULL DEFAULT '',
    open_hours varchar(64) NOT NULL DEFAULT '',
    delivery_range varchar(64) NOT NULL DEFAULT '',
    commission double precision NOT NULL DEFAULT 0,
    bank_account varchar(128) NOT NULL DEFAULT '',
    delivery_method varchar(64) NOT NULL DEFAULT '',
    status varchar(16) NOT NULL DEFAULT 'open',
    audit_status varchar(16) NOT NULL DEFAULT 'approved',
    apply_no varchar(32) NOT NULL DEFAULT '',
    operator varchar(64) NOT NULL DEFAULT '',
    applied_at bigint NOT NULL DEFAULT 0,
    cert_labels jsonb,
    reject_reason varchar(255) NOT NULL DEFAULT '',
    rating double precision NOT NULL DEFAULT 0,
    monthly_orders bigint NOT NULL DEFAULT 0,
    monthly_revenue double precision NOT NULL DEFAULT 0,
    avg_order double precision NOT NULL DEFAULT 0,
    good_rate double precision NOT NULL DEFAULT 0,
    joined_at bigint NOT NULL DEFAULT 0,
    last_updated_at bigint NOT NULL,
    completeness bigint NOT NULL DEFAULT 0,
    review_count bigint NOT NULL DEFAULT 0,
    last_operation varchar(128) NOT NULL DEFAULT '',
    today_orders bigint NOT NULL DEFAULT 0,
    today_revenue double precision NOT NULL DEFAULT 0,
    on_time_rate double precision NOT NULL DEFAULT 0,
    pending_abnormal bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS merchants_region_status_idx ON merchants (region_id, status);
CREATE INDEX IF NOT EXISTS merchants_audit_status_idx ON merchants (audit_status);
CREATE INDEX IF NOT EXISTS merchants_name_idx ON merchants (name);

CREATE TABLE IF NOT EXISTS riders (
    id varchar(26) NOT NULL PRIMARY KEY,
    region_id varchar(26) NOT NULL,
    user_id varchar(26) NOT NULL DEFAULT '',
    name varchar(64) NOT NULL,
    phone varchar(32) NOT NULL DEFAULT '',
    area varchar(128) NOT NULL DEFAULT '',
    tool varchar(32) NOT NULL DEFAULT '',
    level varchar(16) NOT NULL DEFAULT 'Lv.1',
    id_card varchar(64) NOT NULL DEFAULT '',
    vehicle varchar(64) NOT NULL DEFAULT '',
    account varchar(128) NOT NULL DEFAULT '',
    status varchar(16) NOT NULL DEFAULT 'offline',
    audit_status varchar(16) NOT NULL DEFAULT 'pending',
    apply_no varchar(32) NOT NULL DEFAULT '',
    submitted_at bigint NOT NULL DEFAULT 0,
    order_types varchar(128) NOT NULL DEFAULT '',
    id_check varchar(16) NOT NULL DEFAULT '待补充',
    notice_type varchar(16) NOT NULL DEFAULT 'none',
    notice varchar(255) NOT NULL DEFAULT '',
    doc_count bigint NOT NULL DEFAULT 0,
    docs varchar(255) NOT NULL DEFAULT '',
    service_type varchar(64) NOT NULL DEFAULT '',
    joined_at bigint NOT NULL DEFAULT 0,
    service_months bigint NOT NULL DEFAULT 0,
    rating double precision NOT NULL DEFAULT 0,
    today_orders bigint NOT NULL DEFAULT 0,
    today_orders_delta varchar(32) NOT NULL DEFAULT '',
    online_hours double precision NOT NULL DEFAULT 0,
    month_income double precision NOT NULL DEFAULT 0,
    on_time_rate double precision NOT NULL DEFAULT 0,
    complaints_today bigint NOT NULL DEFAULT 0,
    avg_minutes double precision NOT NULL DEFAULT 0,
    last_action varchar(64) NOT NULL DEFAULT '',
    last_action_at bigint NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS riders_region_status_idx ON riders (region_id, status);
CREATE INDEX IF NOT EXISTS riders_audit_status_idx ON riders (audit_status);
CREATE INDEX IF NOT EXISTS riders_phone_idx ON riders (phone);

CREATE TABLE IF NOT EXISTS orders (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_no varchar(32) NOT NULL,
    region_id varchar(26) NOT NULL,
    merchant_id varchar(26) NOT NULL DEFAULT '',
    user_id varchar(26) NOT NULL,
    rider_id varchar(26) NOT NULL DEFAULT '',
    type varchar(16) NOT NULL,
    status varchar(16) NOT NULL,
    route varchar(255) NOT NULL DEFAULT '',
    user_name varchar(64) NOT NULL DEFAULT '',
    rider_name varchar(64) NOT NULL DEFAULT '',
    amount double precision NOT NULL DEFAULT 0,
    created_at bigint NOT NULL,
    channel varchar(64) NOT NULL DEFAULT '',
    estimated_arrival bigint NOT NULL DEFAULT 0,
    remark varchar(512) NOT NULL DEFAULT '',
    goods jsonb,
    fees jsonb,
    steps jsonb,
    merchant jsonb,
    "user" jsonb,
    rider jsonb,
    pay_amount double precision NOT NULL DEFAULT 0,
    pay_method varchar(64) NOT NULL DEFAULT '',
    fund_flow varchar(255) NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX IF NOT EXISTS orders_order_no_unique ON orders (order_no);
CREATE INDEX IF NOT EXISTS orders_region_status_created_idx ON orders (region_id, status, created_at);
CREATE INDEX IF NOT EXISTS orders_merchant_created_idx ON orders (merchant_id, created_at);
CREATE INDEX IF NOT EXISTS orders_user_created_idx ON orders (user_id, created_at);
CREATE INDEX IF NOT EXISTS orders_rider_idx ON orders (rider_id);

CREATE TABLE IF NOT EXISTS complaints (
    id varchar(26) NOT NULL PRIMARY KEY,
    order_id varchar(32) NOT NULL DEFAULT '',
    region_id varchar(26) NOT NULL DEFAULT '',
    user_id varchar(26) NOT NULL DEFAULT '',
    category varchar(32) NOT NULL,
    title varchar(128) NOT NULL,
    status varchar(16) NOT NULL DEFAULT '待处理',
    priority varchar(16) NOT NULL DEFAULT '普通',
    complainant varchar(128) NOT NULL DEFAULT '',
    target varchar(128) NOT NULL DEFAULT '',
    extra varchar(128) NOT NULL DEFAULT '',
    description varchar(1024) NOT NULL DEFAULT '',
    desc_tone varchar(16) NOT NULL DEFAULT 'yellow',
    deadline varchar(64) NOT NULL DEFAULT '',
    handler varchar(64) NOT NULL DEFAULT '',
    progress varchar(255) NOT NULL DEFAULT '',
    initiator varchar(64) NOT NULL DEFAULT '',
    initiator_extra varchar(128) NOT NULL DEFAULT '',
    pay_amount double precision NOT NULL DEFAULT 0,
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL,
    flow jsonb,
    notes jsonb,
    order_items jsonb
);
CREATE INDEX IF NOT EXISTS complaints_status_created_idx ON complaints (status, created_at);
CREATE INDEX IF NOT EXISTS complaints_order_id_idx ON complaints (order_id);

CREATE TABLE IF NOT EXISTS withdrawals (
    id varchar(26) NOT NULL PRIMARY KEY,
    region_id varchar(26) NOT NULL DEFAULT '',
    party_type varchar(16) NOT NULL DEFAULT 'rider',
    party_name varchar(64) NOT NULL DEFAULT '',
    party_no varchar(32) NOT NULL DEFAULT '',
    bank_info varchar(128) NOT NULL DEFAULT '',
    amount double precision NOT NULL DEFAULT 0,
    fee double precision NOT NULL DEFAULT 0,
    fee_note varchar(64) NOT NULL DEFAULT '',
    status varchar(16) NOT NULL DEFAULT '待审核',
    kind varchar(32) NOT NULL DEFAULT '骑手提现',
    applicant varchar(64) NOT NULL DEFAULT '',
    account_holder varchar(64) NOT NULL DEFAULT '',
    withdraw_type varchar(64) NOT NULL DEFAULT '',
    arrive_expect varchar(128) NOT NULL DEFAULT '',
    purpose varchar(255) NOT NULL DEFAULT '',
    balance double precision NOT NULL DEFAULT 0,
    remain_balance double precision NOT NULL DEFAULT 0,
    month_orders bigint NOT NULL DEFAULT 0,
    month_income double precision NOT NULL DEFAULT 0,
    risk_items jsonb,
    risk_warning varchar(255) NOT NULL DEFAULT '',
    reject_reason varchar(255) NOT NULL DEFAULT '',
    created_at bigint NOT NULL,
    timeline jsonb
);
CREATE INDEX IF NOT EXISTS withdrawals_status_created_idx ON withdrawals (status, created_at);
CREATE INDEX IF NOT EXISTS withdrawals_region_id_idx ON withdrawals (region_id);

CREATE TABLE IF NOT EXISTS settlements (
    id varchar(26) NOT NULL PRIMARY KEY,
    region_id varchar(26) NOT NULL DEFAULT '',
    party varchar(64) NOT NULL DEFAULT '',
    type varchar(32) NOT NULL DEFAULT '',
    amount double precision NOT NULL DEFAULT 0,
    status varchar(16) NOT NULL DEFAULT '已到账',
    period varchar(64) NOT NULL DEFAULT '',
    settled_at bigint NOT NULL,
    action varchar(64) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS settlements_settled_at_idx ON settlements (settled_at);
CREATE INDEX IF NOT EXISTS settlements_type_idx ON settlements (type);

CREATE TABLE IF NOT EXISTS announcements (
    id varchar(26) NOT NULL PRIMARY KEY,
    title varchar(128) NOT NULL,
    type varchar(16) NOT NULL DEFAULT '提醒',
    type_tone varchar(16) NOT NULL DEFAULT 'blue',
    audience varchar(64) NOT NULL DEFAULT '全部用户',
    status varchar(16) NOT NULL DEFAULT '待发布',
    published_at bigint NOT NULL DEFAULT 0,
    planned_at bigint NOT NULL DEFAULT 0,
    action_name varchar(16) NOT NULL DEFAULT '发布',
    content varchar(2048) NOT NULL DEFAULT '',
    tags jsonb,
    stats jsonb,
    created_by varchar(26) NOT NULL DEFAULT '',
    created_at bigint NOT NULL,
    updated_at bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS announcements_status_planned_idx ON announcements (status, planned_at);
CREATE INDEX IF NOT EXISTS announcements_type_idx ON announcements (type);

CREATE TABLE IF NOT EXISTS fee_rules (
    id varchar(26) NOT NULL PRIMARY KEY,
    category varchar(16) NOT NULL,
    name varchar(64) NOT NULL,
    status varchar(16) NOT NULL DEFAULT '已启用',
    value varchar(64) NOT NULL DEFAULT '',
    "desc" varchar(128) NOT NULL DEFAULT '',
    accent boolean NOT NULL DEFAULT false,
    highlight boolean NOT NULL DEFAULT false,
    rule varchar(128) NOT NULL DEFAULT '',
    example varchar(128) NOT NULL DEFAULT '',
    fee varchar(64) NOT NULL DEFAULT '',
    rider_share varchar(64) NOT NULL DEFAULT '',
    icon_color varchar(32) NOT NULL DEFAULT '',
    version varchar(16) NOT NULL DEFAULT '',
    updated_at bigint NOT NULL,
    next_effective_at bigint NOT NULL DEFAULT 0,
    enabled boolean NOT NULL DEFAULT true
);
CREATE INDEX IF NOT EXISTS fee_rules_category_idx ON fee_rules (category);

CREATE TABLE IF NOT EXISTS mail_outboxes (
    id varchar(26) NOT NULL PRIMARY KEY,
    to_email varchar(255) NOT NULL,
    subject varchar(255) NOT NULL DEFAULT '',
    body varchar(4096) NOT NULL DEFAULT '',
    created_at bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS mail_outboxes_created_at_idx ON mail_outboxes (created_at);

CREATE TABLE IF NOT EXISTS casbin_rules (
    id varchar(26) NOT NULL PRIMARY KEY,
    ptype varchar(16) NOT NULL,
    v0 varchar(64) NOT NULL DEFAULT '',
    v1 varchar(64) NOT NULL DEFAULT '',
    v2 varchar(64) NOT NULL DEFAULT '',
    v3 varchar(64) NOT NULL DEFAULT '',
    v4 varchar(64) NOT NULL DEFAULT '',
    v5 varchar(64) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS casbin_rules_ptype_v0_idx ON casbin_rules (ptype, v0);

CREATE TABLE IF NOT EXISTS audit_logs (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL DEFAULT '',
    action varchar(64) NOT NULL,
    object varchar(255) NOT NULL DEFAULT '',
    result varchar(16) NOT NULL DEFAULT 'success',
    request_id varchar(32) NOT NULL DEFAULT '',
    operator varchar(64) NOT NULL DEFAULT '',
    ip varchar(64) NOT NULL DEFAULT '',
    created_at bigint NOT NULL
);
CREATE INDEX IF NOT EXISTS audit_logs_user_created_idx ON audit_logs (user_id, created_at);
CREATE INDEX IF NOT EXISTS audit_logs_action_idx ON audit_logs (action);

COMMIT;
