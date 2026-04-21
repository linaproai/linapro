-- 001: monitor-server schema

CREATE TABLE IF NOT EXISTS plugin_monitor_server (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '记录ID',
    node_name   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '节点名称（hostname）',
    node_ip     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '节点IP地址',
    data        JSON         NOT NULL             COMMENT '监控数据（JSON格式，包含CPU、内存、磁盘、网络、Go运行时等指标）',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '采集时间',
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE INDEX uk_plugin_monitor_server_node (node_name, node_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='服务器监控表';
