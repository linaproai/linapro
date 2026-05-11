-- Mock data: login log records for monitoring demos.
-- 模拟数据：监控演示使用的登录日志记录。
-- Static login log rows use exact existence checks so mock loading is idempotent.

INSERT INTO plugin_monitor_loginlog ("tenant_id", "user_name", "status", "ip", "browser", "os", "msg", "login_time")
SELECT 0, 'admin', 0, '192.168.10.11', 'Chrome 124.0', 'macOS 14', 'Login succeeded', '2026-04-20 08:45:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog existing
    WHERE existing."tenant_id" = 0
      AND existing."user_name" = 'admin'
      AND existing."status" = 0
      AND existing."ip" = '192.168.10.11'
      AND existing."login_time" = '2026-04-20 08:45:00'
);

INSERT INTO plugin_monitor_loginlog ("tenant_id", "user_name", "status", "ip", "browser", "os", "msg", "login_time")
SELECT 0, 'user002', 0, '192.168.10.42', 'Edge 124.0', 'Windows 11', 'Login succeeded', '2026-04-20 09:12:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog existing
    WHERE existing."tenant_id" = 0
      AND existing."user_name" = 'user002'
      AND existing."status" = 0
      AND existing."ip" = '192.168.10.42'
      AND existing."login_time" = '2026-04-20 09:12:00'
);

INSERT INTO plugin_monitor_loginlog ("tenant_id", "user_name", "status", "ip", "browser", "os", "msg", "login_time")
SELECT 0, 'user023', 1, '203.0.113.24', 'Firefox 125.0', 'Ubuntu 24.04', 'Password verification failed', '2026-04-20 10:05:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog existing
    WHERE existing."tenant_id" = 0
      AND existing."user_name" = 'user023'
      AND existing."status" = 1
      AND existing."ip" = '203.0.113.24'
      AND existing."login_time" = '2026-04-20 10:05:00'
);

INSERT INTO plugin_monitor_loginlog ("tenant_id", "user_name", "status", "ip", "browser", "os", "msg", "login_time")
SELECT 0, 'user060', 0, '198.51.100.18', 'Safari 17.4', 'iOS 17', 'Login succeeded', '2026-04-21 14:35:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog existing
    WHERE existing."tenant_id" = 0
      AND existing."user_name" = 'user060'
      AND existing."status" = 0
      AND existing."ip" = '198.51.100.18'
      AND existing."login_time" = '2026-04-21 14:35:00'
);
