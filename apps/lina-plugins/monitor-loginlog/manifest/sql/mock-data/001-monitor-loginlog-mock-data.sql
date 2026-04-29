-- Mock data: login log records for monitoring demos.
-- 模拟数据：监控演示使用的登录日志记录。

INSERT IGNORE INTO plugin_monitor_loginlog (user_name, status, ip, browser, os, msg, login_time)
SELECT 'admin', 0, '192.168.10.11', 'Chrome 124.0', 'macOS 14', 'Login succeeded', '2026-04-20 08:45:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog
    WHERE user_name = 'admin'
      AND ip = '192.168.10.11'
      AND login_time = '2026-04-20 08:45:00'
);

INSERT IGNORE INTO plugin_monitor_loginlog (user_name, status, ip, browser, os, msg, login_time)
SELECT 'user002', 0, '192.168.10.42', 'Edge 124.0', 'Windows 11', 'Login succeeded', '2026-04-20 09:12:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog
    WHERE user_name = 'user002'
      AND ip = '192.168.10.42'
      AND login_time = '2026-04-20 09:12:00'
);

INSERT IGNORE INTO plugin_monitor_loginlog (user_name, status, ip, browser, os, msg, login_time)
SELECT 'user023', 1, '203.0.113.24', 'Firefox 125.0', 'Ubuntu 24.04', 'Password verification failed', '2026-04-20 10:05:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog
    WHERE user_name = 'user023'
      AND ip = '203.0.113.24'
      AND login_time = '2026-04-20 10:05:00'
);

INSERT IGNORE INTO plugin_monitor_loginlog (user_name, status, ip, browser, os, msg, login_time)
SELECT 'user060', 0, '198.51.100.18', 'Safari 17.4', 'iOS 17', 'Login succeeded', '2026-04-21 14:35:00'
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_monitor_loginlog
    WHERE user_name = 'user060'
      AND ip = '198.51.100.18'
      AND login_time = '2026-04-21 14:35:00'
);
