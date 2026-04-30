-- 006: System Monitoring (Online Users + Server Monitor)
-- 006：系统监控（在线用户与服务监控）

-- ============================================================
-- Online session table (MEMORY engine, tracks currently online users)
-- 在线会话表（MEMORY 引擎，用于跟踪当前在线用户）
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_online_session (
    token_id    VARCHAR(64)  NOT NULL COMMENT  'Session token ID (UUID)',
    user_id     INT          NOT NULL DEFAULT 0  COMMENT  'User ID',
    username    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Login account',
    dept_name   VARCHAR(100) NOT NULL DEFAULT '' COMMENT  'Department name',
    ip          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Login IP',
    browser     VARCHAR(100) NOT NULL DEFAULT '' COMMENT  'Browser',
    os          VARCHAR(100) NOT NULL DEFAULT '' COMMENT  'Operating system',
    login_time       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Login time',
    last_active_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Last active time',
    PRIMARY KEY (token_id),
    INDEX idx_user_id (user_id),
    INDEX idx_username (username),
    INDEX idx_last_active_time (last_active_time)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Online session table';

-- Server monitoring capability has moved down to the source plugin monitor-server.
-- 服务监控能力已下沉到源码插件 monitor-server。
