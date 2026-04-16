-- 分布式锁表（MEMORY 引擎）
-- 用于领导选举和分布式锁管理
-- 服务重启后锁状态自动清空，符合分布式锁的临时状态特性

CREATE TABLE IF NOT EXISTS `sys_locker` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    `name` VARCHAR(64) NOT NULL COMMENT '锁名称，唯一标识',
    `reason` VARCHAR(255) DEFAULT '' COMMENT '获取锁的原因',
    `holder` VARCHAR(64) DEFAULT '' COMMENT '锁持有者标识（节点名）',
    `expire_time` DATETIME NOT NULL COMMENT '锁过期时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_name` (`name`),
    INDEX `idx_expire_time` (`expire_time`)
) ENGINE=MEMORY DEFAULT CHARSET=utf8mb4 COMMENT='分布式锁表';
