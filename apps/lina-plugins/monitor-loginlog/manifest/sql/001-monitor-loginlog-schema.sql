-- 001: monitor-loginlog schema

CREATE TABLE IF NOT EXISTS plugin_monitor_loginlog (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    user_name   VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '登录账号',
    status      TINYINT      NOT NULL DEFAULT 0  COMMENT '登录状态（0成功 1失败）',
    ip          VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '登录IP地址',
    browser     VARCHAR(200) NOT NULL DEFAULT '' COMMENT '浏览器类型',
    os          VARCHAR(200) NOT NULL DEFAULT '' COMMENT '操作系统',
    msg         VARCHAR(500) NOT NULL DEFAULT '' COMMENT '提示消息',
    login_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='系统登录日志表';
