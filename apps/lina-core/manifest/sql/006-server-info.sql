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

-- ============================================================
-- 服务器监控表（存储各节点的定时采集指标数据）
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_server_monitor (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '记录ID',
    node_name   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '节点名称（hostname）',
    node_ip     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '节点IP地址',
    data        JSON         NOT NULL             COMMENT '监控数据（JSON格式，包含CPU、内存、磁盘、网络、Go运行时等指标）',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '采集时间',
    UNIQUE INDEX uk_node (node_name, node_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='服务器监控表';
