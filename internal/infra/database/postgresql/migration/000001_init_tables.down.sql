DROP INDEX IF EXISTS idx_property_ads_deleted_at;
DROP INDEX IF EXISTS idx_property_ads_city_type;
DROP INDEX IF EXISTS idx_users_deleted_at;

DROP TABLE IF EXISTS property_ads;
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pgcrypto";