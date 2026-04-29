-- Distributed lock table (MEMORY engine)
-- 分布式锁表（MEMORY 引擎）
-- Used for leader election and distributed lock management.
-- 用于领导选举和分布式锁管理
-- Lock state is automatically cleared after service restart, matching the temporary-state nature of distributed locks.
-- 服务重启后锁状态自动清空，符合分布式锁的临时状态特性

CREATE TABLE IF NOT EXISTS `sys_locker` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT  'Primary key ID',
    `name` VARCHAR(64) NOT NULL COMMENT  'Lock name, unique identifier',
    `reason` VARCHAR(255) DEFAULT '' COMMENT  'Reason for acquiring the lock',
    `holder` VARCHAR(64) DEFAULT '' COMMENT  'Lock holder identifier (node name)',
    `expire_time` DATETIME NOT NULL COMMENT  'Lock expiration time',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT  'Update time',
    UNIQUE KEY `uk_name` (`name`),
    INDEX `idx_expire_time` (`expire_time`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT= 'Distributed lock table';
