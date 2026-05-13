CREATE TABLE `media_strategy` (
`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '策略ID',
`name` varchar(255) DEFAULT '' COMMENT '策略名称',
`strategy` text COMMENT 'yaml格式策略内容',
`global` int DEFAULT '2' COMMENT '为1则是全局策略，只能有一个是1，2关闭',
`enable` int DEFAULT '1' COMMENT '1开启，2关闭',
`creator_id` int DEFAULT '0' COMMENT '创建人Id',
`create_time` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
`updater_id` int DEFAULT '0' COMMENT '修改人Id',
`update_time` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '修改时间',
PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='策略记录表';

CREATE TABLE `media_strategy_device` (
`device_id` varchar(255) NOT NULL COMMENT '设备国标id',
`strategy_id` bigint unsigned NOT NULL COMMENT '策略id',
UNIQUE KEY `uk_device` (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='设备策略表';

CREATE TABLE `media_strategy_device_tenant` (
`tenant_id` varchar(255) NOT NULL COMMENT '租户id',
`device_id` varchar(255) NOT NULL COMMENT '设备国标id',
`strategy_id` bigint unsigned NOT NULL COMMENT '策略ID',
UNIQUE KEY `uk_tenant_device` (`tenant_id`,`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户设备策略表';

CREATE TABLE `media_strategy_tenant` (
`tenant_id` varchar(255) NOT NULL COMMENT '租户id',
`strategy_id` bigint unsigned NOT NULL COMMENT '策略ID',
UNIQUE KEY `uk_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户策略表';

CREATE TABLE `media_stream_alias` (
`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
`alias` varchar(255) NOT NULL COMMENT '流别名（主键）',
`auto_remove` tinyint(1) DEFAULT '0' COMMENT '是否自动移除',
`stream_path` varchar(255) DEFAULT '' COMMENT '真实流路径',
`create_time` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
PRIMARY KEY (`id`),
UNIQUE KEY `uk_alias` (`alias`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='流别名表';

CREATE TABLE `hg_tenant_white` (
`tenant_id` varchar(64) NOT NULL COMMENT '租户id',
`ip` varchar(32) NOT NULL COMMENT '白名单地址',
`description` varchar(32) DEFAULT NULL COMMENT '白名单描述',
`enable` tinyint(1) NOT NULL COMMENT '1开启，0关闭',
`creator_id` int DEFAULT NULL COMMENT '创建人Id',
`create_time` datetime NOT NULL COMMENT '创建时间',
`updater_id` int DEFAULT NULL COMMENT '修改人Id',
`update_time` datetime DEFAULT NULL COMMENT '修改时间',
UNIQUE KEY `uk_tenant_ip` (`tenant_id`,`ip`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='租户白名单表';
