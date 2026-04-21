-- 001: monitor-operlog schema uninstall

DELETE FROM sys_dict_data WHERE dict_type IN ('sys_oper_type', 'sys_oper_status');
DELETE FROM sys_dict_type WHERE type IN ('sys_oper_type', 'sys_oper_status');
DROP TABLE IF EXISTS plugin_monitor_operlog;
