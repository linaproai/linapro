-- ------------------------------------------------------------
-- 001-plugin-demo-source-records.sql
-- plugin-demo-source 示例记录表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `plugin_demo_source_record` (
    `id`              BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    `title`           VARCHAR(128) NOT NULL DEFAULT '' COMMENT '记录标题',
    `content`         VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '记录内容',
    `attachment_name` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '附件原始文件名',
    `attachment_path` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '附件相对存储路径',
    `created_at`      DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='源码插件示例记录表';
