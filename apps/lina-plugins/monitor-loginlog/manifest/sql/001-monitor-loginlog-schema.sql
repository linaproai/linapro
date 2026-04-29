-- 001: monitor-loginlog schema
-- 001：monitor-loginlog 数据结构

CREATE TABLE IF NOT EXISTS plugin_monitor_loginlog (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Log ID',
    user_name   VARCHAR(50)  NOT NULL DEFAULT '' COMMENT  'Login account',
    status      TINYINT      NOT NULL DEFAULT 0  COMMENT  'Login status: 0=succeeded, 1=failed',
    ip          VARCHAR(50)  NOT NULL DEFAULT '' COMMENT  'Login IP address',
    browser     VARCHAR(200) NOT NULL DEFAULT '' COMMENT  'Browser type',
    os          VARCHAR(200) NOT NULL DEFAULT '' COMMENT  'Operating system',
    msg         VARCHAR(500) NOT NULL DEFAULT '' COMMENT  'Prompt message',
    login_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Login time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'System login log table';
