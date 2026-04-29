-- ------------------------------------------------------------
-- 012 plugin host-call SQL file
-- 012 插件宿主调用 SQL 文件
-- Plugin host call: key-value state storage table
-- 插件宿主调用：键值状态存储表
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_plugin_state (
    id           INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Primary key ID',
    plugin_id    VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'Plugin unique identifier (kebab-case)',
    state_key    VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'State key',
    state_value  LONGTEXT                          COMMENT  'State value with JSON support',
    created_at   DATETIME                          COMMENT  'Creation time',
    updated_at   DATETIME                          COMMENT  'Update time',
    UNIQUE KEY uk_plugin_state (plugin_id, state_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Plugin key-value state storage table';
