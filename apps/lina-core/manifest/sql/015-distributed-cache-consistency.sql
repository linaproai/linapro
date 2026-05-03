-- ------------------------------------------------------------
-- 015 distributed cache consistency SQL file
-- 015 分布式缓存一致性 SQL 文件
-- Persistent cache revision coordination for critical host caches
-- 关键宿主缓存的持久修订号协调
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS `sys_cache_revision` (
    `id`         BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Primary key ID',
    `domain`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'Cache domain, such as runtime-config, permission-access, or plugin-runtime',
    `scope`      VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'Explicit invalidation scope, such as global, plugin:<id>, locale:<locale>, or user:<id>',
    `revision`   BIGINT       NOT NULL DEFAULT 0 COMMENT 'Monotonic cache revision for this domain and scope',
    `reason`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Latest change reason used for diagnostics',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Update time',
    UNIQUE KEY `uk_domain_scope` (`domain`, `scope`),
    KEY `idx_domain_updated_at` (`domain`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Persistent cache revision coordination table';
