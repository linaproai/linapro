-- Mock data: online sessions used by the online-user monitoring plugin.
-- 模拟数据：在线用户监控插件使用的在线会话。

INSERT INTO sys_online_session (
    "tenant_id",
    "token_id",
    "user_id",
    "username",
    "dept_name",
    "ip",
    "browser",
    "os",
    "login_time",
    "last_active_time"
)
SELECT
    0,
    'mock-online-admin-session',
    u."id",
    u."username",
    'LinaPro.AI',
    '192.168.10.11',
    'Chrome 124.0',
    'macOS 14',
    '2026-04-20 08:45:00',
    '2026-04-20 10:45:00'
FROM sys_user u
WHERE u."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO sys_online_session (
    "tenant_id",
    "token_id",
    "user_id",
    "username",
    "dept_name",
    "ip",
    "browser",
    "os",
    "login_time",
    "last_active_time"
)
SELECT
    0,
    'mock-online-user002-session',
    u."id",
    u."username",
    '研发部门',
    '192.168.10.42',
    'Edge 124.0',
    'Windows 11',
    '2026-04-20 09:12:00',
    '2026-04-20 10:40:00'
FROM sys_user u
WHERE u."username" = 'user002'
ON CONFLICT DO NOTHING;

INSERT INTO sys_online_session (
    "tenant_id",
    "token_id",
    "user_id",
    "username",
    "dept_name",
    "ip",
    "browser",
    "os",
    "login_time",
    "last_active_time"
)
SELECT
    0,
    'mock-online-user060-session',
    u."id",
    u."username",
    '研发部门',
    '198.51.100.18',
    'Safari 17.4',
    'iOS 17',
    '2026-04-21 14:35:00',
    '2026-04-21 15:05:00'
FROM sys_user u
WHERE u."username" = 'user060'
ON CONFLICT DO NOTHING;
