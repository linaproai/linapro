-- 001: monitor-server schema
-- 001：monitor-server 数据结构

CREATE TABLE IF NOT EXISTS plugin_monitor_server (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Record ID',
    node_name   VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Node name (hostname)',
    node_ip     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Node IP address',
    data        JSON         NOT NULL             COMMENT  'Monitoring data in JSON format, including CPU, memory, disk, network, Go runtime, and other metrics',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Collection time',
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time',
    UNIQUE INDEX uk_plugin_monitor_server_node (node_name, node_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Server monitoring table';
