-- ------------------------------------------------------------
-- 001 media schema uninstall SQL file
-- Purpose: Removes media plugin-owned tables.
-- ------------------------------------------------------------

DROP TABLE IF EXISTS media_strategy_device_tenant;
DROP TABLE IF EXISTS media_strategy_device;
DROP TABLE IF EXISTS media_strategy_tenant;
DROP TABLE IF EXISTS media_stream_alias;
DROP TABLE IF EXISTS hg_tenant_white;
DROP TABLE IF EXISTS media_strategy;
DROP FUNCTION IF EXISTS media_touch_update_time();
