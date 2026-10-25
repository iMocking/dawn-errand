-- 005_delivery：骑手实时位置与站内通知（M3）
CREATE TABLE IF NOT EXISTS rider_locations (
    id varchar(26) NOT NULL PRIMARY KEY,
    rider_id varchar(26) NOT NULL,
    region_id varchar(26) NOT NULL DEFAULT '',
    lat double precision NOT NULL DEFAULT 0,
    lng double precision NOT NULL DEFAULT 0,
    heading double precision NOT NULL DEFAULT 0,
    speed double precision NOT NULL DEFAULT 0,
    accuracy double precision NOT NULL DEFAULT 0,
    updated_at bigint NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS rider_locations_rider_id_unique ON rider_locations (rider_id);
CREATE INDEX IF NOT EXISTS rider_locations_region_idx ON rider_locations (region_id);
CREATE INDEX IF NOT EXISTS rider_locations_updated_idx ON rider_locations (updated_at);

CREATE TABLE IF NOT EXISTS notifications (
    id varchar(26) NOT NULL PRIMARY KEY,
    user_id varchar(26) NOT NULL,
    title varchar(128) NOT NULL,
    content varchar(512) NOT NULL DEFAULT '',
    type varchar(16) NOT NULL DEFAULT 'order',
    ref_id varchar(32) NOT NULL DEFAULT '',
    read_at bigint NOT NULL DEFAULT 0,
    created_at bigint NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS notifications_user_created_idx ON notifications (user_id, created_at);
CREATE INDEX IF NOT EXISTS notifications_user_read_idx ON notifications (user_id, read_at);
