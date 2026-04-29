-- Mock data: operation log records for monitoring demos.
-- 模拟数据：监控演示使用的操作日志记录。

INSERT IGNORE INTO plugin_monitor_operlog (
    title,
    oper_summary,
    route_owner,
    route_method,
    route_path,
    route_doc_key,
    oper_type,
    method,
    request_method,
    oper_name,
    oper_url,
    oper_ip,
    oper_param,
    json_result,
    status,
    error_msg,
    cost_time,
    oper_time
)
SELECT
    '用户管理',
    'Create demo user',
    'core',
    'POST',
    '/api/v1/user',
    'core.user.create',
    'create',
    'user.Create',
    'POST',
    'admin',
    '/api/v1/user',
    '192.168.10.11',
    '{"username":"demo_user"}',
    '{"code":0,"message":"ok"}',
    0,
    '',
    86,
    '2026-04-20 09:30:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog
    WHERE title = '用户管理'
      AND oper_summary = 'Create demo user'
      AND oper_time = '2026-04-20 09:30:00'
);

INSERT IGNORE INTO plugin_monitor_operlog (
    title,
    oper_summary,
    route_owner,
    route_method,
    route_path,
    route_doc_key,
    oper_type,
    method,
    request_method,
    oper_name,
    oper_url,
    oper_ip,
    oper_param,
    json_result,
    status,
    error_msg,
    cost_time,
    oper_time
)
SELECT
    '参数设置',
    'Update public runtime config',
    'core',
    'PUT',
    '/api/v1/config/{id}',
    'core.config.update',
    'update',
    'config.Update',
    'PUT',
    'admin',
    '/api/v1/config/12',
    '192.168.10.11',
    '{"key":"sys.ui.theme.mode","value":"light"}',
    '{"code":0,"message":"ok"}',
    0,
    '',
    42,
    '2026-04-20 10:05:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog
    WHERE title = '参数设置'
      AND oper_summary = 'Update public runtime config'
      AND oper_time = '2026-04-20 10:05:00'
);

INSERT IGNORE INTO plugin_monitor_operlog (
    title,
    oper_summary,
    route_owner,
    route_method,
    route_path,
    route_doc_key,
    oper_type,
    method,
    request_method,
    oper_name,
    oper_url,
    oper_ip,
    oper_param,
    json_result,
    status,
    error_msg,
    cost_time,
    oper_time
)
SELECT
    '插件管理',
    'Install source plugin',
    'core',
    'POST',
    '/api/v1/plugins/{id}/install',
    'core.plugin.install',
    'create',
    'plugin.Install',
    'POST',
    'admin',
    '/api/v1/plugins/org-center/install',
    '192.168.10.11',
    '{"id":"org-center"}',
    '{"code":0,"message":"ok"}',
    0,
    '',
    318,
    '2026-04-20 11:20:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog
    WHERE title = '插件管理'
      AND oper_summary = 'Install source plugin'
      AND oper_time = '2026-04-20 11:20:00'
);

INSERT IGNORE INTO plugin_monitor_operlog (
    title,
    oper_summary,
    route_owner,
    route_method,
    route_path,
    route_doc_key,
    oper_type,
    method,
    request_method,
    oper_name,
    oper_url,
    oper_ip,
    oper_param,
    json_result,
    status,
    error_msg,
    cost_time,
    oper_time
)
SELECT
    '文件管理',
    'Delete locked demo file',
    'core',
    'DELETE',
    '/api/v1/file/{id}',
    'core.file.delete',
    'delete',
    'file.Delete',
    'DELETE',
    'user023',
    '/api/v1/file/9001',
    '203.0.113.24',
    '{"id":9001}',
    '{"code":500,"message":"permission denied"}',
    1,
    'Permission denied for demo file deletion',
    64,
    '2026-04-21 15:40:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog
    WHERE title = '文件管理'
      AND oper_summary = 'Delete locked demo file'
      AND oper_time = '2026-04-21 15:40:00'
);
