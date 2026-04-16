-- ------------------------------------------------------------
-- 012-plugin-host-call.sql
-- 插件宿主调用：键值状态存储表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id           INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    state_key    VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '状态键',
    state_value  LONGTEXT                          COMMENT '状态值（支持JSON）',
    created_at   DATETIME                          COMMENT '创建时间',
    updated_at   DATETIME                          COMMENT '更新时间',
    UNIQUE KEY uk_plugin_state (plugin_id, state_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件键值状态存储表';
