-- 001: monitor-server schema
-- 001：monitor-server 数据结构

-- Purpose: Stores platform-level server monitoring snapshots collected from runtime nodes.
-- 用途：存储从运行时节点采集的平台级服务器监控快照。
CREATE TABLE IF NOT EXISTS plugin_monitor_server (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "node_name"   VARCHAR(128) NOT NULL DEFAULT '',
    "node_ip"     VARCHAR(64)  NOT NULL DEFAULT '',
    "data"        TEXT                            NOT NULL,
    "created_at"  TIMESTAMP                       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP                       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_monitor_server IS 'Server monitoring table';
COMMENT ON COLUMN plugin_monitor_server."id" IS 'Record ID';
COMMENT ON COLUMN plugin_monitor_server."node_name" IS 'Node name (hostname)';
COMMENT ON COLUMN plugin_monitor_server."node_ip" IS 'Node IP address';
COMMENT ON COLUMN plugin_monitor_server."data" IS 'Monitoring data in structured text format, including CPU, memory, disk, network, Go runtime, and other metrics';
COMMENT ON COLUMN plugin_monitor_server."created_at" IS 'Collection time';
COMMENT ON COLUMN plugin_monitor_server."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_monitor_server_node ON plugin_monitor_server ("node_name", "node_ip");
