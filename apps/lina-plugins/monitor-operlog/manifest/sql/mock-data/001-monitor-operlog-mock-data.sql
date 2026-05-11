-- Mock data: operation log records for monitoring demos.
-- 模拟数据：监控演示使用的操作日志记录。
-- Static operation log rows use exact existence checks so mock loading is idempotent.

INSERT INTO plugin_monitor_operlog (
    "tenant_id",
    "title",
    "oper_summary",
    "route_owner",
    "route_method",
    "route_path",
    "route_doc_key",
    "oper_type",
    "method",
    "request_method",
    "oper_name",
    "oper_url",
    "oper_ip",
    "oper_param",
    "json_result",
    "status",
    "error_msg",
    "cost_time",
    "oper_time"
)
SELECT
    0,
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
    1,
    '2026-04-20 09:30:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog existing
    WHERE existing."tenant_id" = 0
      AND existing."route_owner" = 'core'
      AND existing."route_method" = 'POST'
      AND existing."route_path" = '/api/v1/user'
      AND existing."oper_name" = 'admin'
      AND existing."oper_time" = '2026-04-20 09:30:00'
);

INSERT INTO plugin_monitor_operlog (
    "tenant_id",
    "title",
    "oper_summary",
    "route_owner",
    "route_method",
    "route_path",
    "route_doc_key",
    "oper_type",
    "method",
    "request_method",
    "oper_name",
    "oper_url",
    "oper_ip",
    "oper_param",
    "json_result",
    "status",
    "error_msg",
    "cost_time",
    "oper_time"
)
SELECT
    0,
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
    0,
    '2026-04-20 10:05:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog existing
    WHERE existing."tenant_id" = 0
      AND existing."route_owner" = 'core'
      AND existing."route_method" = 'PUT'
      AND existing."route_path" = '/api/v1/config/{id}'
      AND existing."oper_name" = 'admin'
      AND existing."oper_time" = '2026-04-20 10:05:00'
);

INSERT INTO plugin_monitor_operlog (
    "tenant_id",
    "title",
    "oper_summary",
    "route_owner",
    "route_method",
    "route_path",
    "route_doc_key",
    "oper_type",
    "method",
    "request_method",
    "oper_name",
    "oper_url",
    "oper_ip",
    "oper_param",
    "json_result",
    "status",
    "error_msg",
    "cost_time",
    "oper_time"
)
SELECT
    0,
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
    4,
    '2026-04-20 11:20:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog existing
    WHERE existing."tenant_id" = 0
      AND existing."route_owner" = 'core'
      AND existing."route_method" = 'POST'
      AND existing."route_path" = '/api/v1/plugins/{id}/install'
      AND existing."oper_name" = 'admin'
      AND existing."oper_time" = '2026-04-20 11:20:00'
);

INSERT INTO plugin_monitor_operlog (
    "tenant_id",
    "title",
    "oper_summary",
    "route_owner",
    "route_method",
    "route_path",
    "route_doc_key",
    "oper_type",
    "method",
    "request_method",
    "oper_name",
    "oper_url",
    "oper_ip",
    "oper_param",
    "json_result",
    "status",
    "error_msg",
    "cost_time",
    "oper_time"
)
SELECT
    0,
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
    0,
    '2026-04-21 15:40:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_operlog existing
    WHERE existing."tenant_id" = 0
      AND existing."route_owner" = 'core'
      AND existing."route_method" = 'DELETE'
      AND existing."route_path" = '/api/v1/file/{id}'
      AND existing."oper_name" = 'user023'
      AND existing."oper_time" = '2026-04-21 15:40:00'
);
