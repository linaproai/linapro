-- ------------------------------------------------------------
-- 001 media schema uninstall SQL file
-- Purpose: Removes media plugin-owned tables.
-- ------------------------------------------------------------

DROP TABLE IF EXISTS media_strategy_device_tenant;
DROP TABLE IF EXISTS media_strategy_device;
DROP TABLE IF EXISTS media_strategy_tenant;
DROP TABLE IF EXISTS media_stream_alias;
DROP TABLE IF EXISTS media_device_node;
DROP TABLE IF EXISTS media_tenant_stream_config;
DROP TABLE IF EXISTS media_node;
DROP TABLE IF EXISTS media_tenant_white;
DROP TABLE IF EXISTS media_strategy;
DROP FUNCTION IF EXISTS media_touch_update_time();
