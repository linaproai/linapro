-- 006: System Monitoring (Online Users + Server Monitor)

-- ============================================================
-- 在线会话表（MEMORY 引擎，用于跟踪当前在线用户）
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_online_session (
    token_id    VARCHAR(64)  NOT NULL COMMENT '会话Token ID（UUID）',
    user_id     INT          NOT NULL DEFAULT 0  COMMENT '用户ID',
    username    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录账号',
    dept_name   VARCHAR(100) NOT NULL DEFAULT '' COMMENT '部门名称',
    ip          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录IP',
    browser     VARCHAR(100) NOT NULL DEFAULT '' COMMENT '浏览器',
    os          VARCHAR(100) NOT NULL DEFAULT '' COMMENT '操作系统',
    login_time       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    last_active_time DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后活跃时间',
    PRIMARY KEY (token_id),
    INDEX idx_user_id (user_id),
    INDEX idx_username (username)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='在线会话表';

-- 服务监控能力已下沉到源码插件 monitor-server。
