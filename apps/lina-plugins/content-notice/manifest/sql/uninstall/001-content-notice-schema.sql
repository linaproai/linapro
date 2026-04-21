-- 001: content-notice schema uninstall

DELETE FROM sys_dict_data WHERE dict_type IN ('sys_notice_type', 'sys_notice_status');
DELETE FROM sys_dict_type WHERE type IN ('sys_notice_type', 'sys_notice_status');
DROP TABLE IF EXISTS sys_notice;
