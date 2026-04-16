-- 009: Server Monitor - Add updated_at field for tracking latest report time

-- ============================================================
-- 服务器监控表添加 updated_at 字段
-- ============================================================
ALTER TABLE sys_server_monitor
ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间' AFTER created_at;
