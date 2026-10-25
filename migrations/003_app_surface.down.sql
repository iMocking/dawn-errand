-- 回滚 003_app_surface
DROP INDEX IF EXISTS riders_user_id_unique;
DROP INDEX IF EXISTS merchants_owner_user_id_unique;

ALTER TABLE merchants DROP COLUMN IF EXISTS service_phone;
ALTER TABLE merchants DROP COLUMN IF EXISTS announcement;
ALTER TABLE merchants DROP COLUMN IF EXISTS packing_fee;
ALTER TABLE merchants DROP COLUMN IF EXISTS delivery_fee;
ALTER TABLE merchants DROP COLUMN IF EXISTS min_price;
ALTER TABLE merchants DROP COLUMN IF EXISTS auto_accept;
ALTER TABLE merchants DROP COLUMN IF EXISTS lng;
ALTER TABLE merchants DROP COLUMN IF EXISTS lat;

DROP INDEX IF EXISTS user_addresses_one_default_idx;
DROP INDEX IF EXISTS user_addresses_user_created_idx;
DROP TABLE IF EXISTS user_addresses;

DROP INDEX IF EXISTS dishes_name_idx;
DROP INDEX IF EXISTS dishes_category_id_idx;
DROP INDEX IF EXISTS dishes_merchant_on_sale_idx;
DROP TABLE IF EXISTS dishes;

DROP INDEX IF EXISTS dish_categories_merchant_sort_idx;
DROP TABLE IF EXISTS dish_categories;
