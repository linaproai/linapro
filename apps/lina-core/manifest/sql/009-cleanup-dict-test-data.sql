-- ============================================================
-- 009-cleanup-dict-test-data.sql
-- Clean up orphaned test data from system dict types
-- ============================================================

-- Remove test data entries that were mistakenly added to sys_normal_disable
-- These were created by TC0013 tests that used sys_normal_disable instead of a dedicated test type
DELETE FROM sys_dict_data
WHERE dict_type = 'sys_normal_disable'
  AND label NOT IN ('正常', '停用');

-- Remove test data entries that were mistakenly added to sys_show_hide
DELETE FROM sys_dict_data
WHERE dict_type = 'sys_show_hide'
  AND label NOT IN ('显示', '隐藏');
