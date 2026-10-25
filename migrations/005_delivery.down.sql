-- 回滚 005_delivery
DROP INDEX IF EXISTS notifications_user_read_idx;
DROP INDEX IF EXISTS notifications_user_created_idx;
DROP TABLE IF EXISTS notifications;

DROP INDEX IF EXISTS rider_locations_updated_idx;
DROP INDEX IF EXISTS rider_locations_region_idx;
DROP INDEX IF EXISTS rider_locations_rider_id_unique;
DROP TABLE IF EXISTS rider_locations;
