-- 007: 参数设置与任务调度模块
-- 包含：参数设置表、登录状态字典、文件场景字典、任务状态字典

-- ----------------------------
-- 1. 参数设置表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `sys_config` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT COMMENT '参数ID',
    `name`       VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT '参数名称',
    `key`        VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT '参数键名',
    `value`      VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT '参数键值',
    `remark`     VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT '备注',
    `created_at` DATETIME         DEFAULT NULL            COMMENT '创建时间',
    `updated_at` DATETIME         DEFAULT NULL            COMMENT '修改时间',
    `deleted_at` DATETIME         DEFAULT NULL            COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='参数设置表';

-- ============================================================
-- 字典初始化数据：登录状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('登录状态', 'sys_login_status', 1, '登录日志状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_login_status', '成功', '0', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_login_status', '失败', '1', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典初始化数据：文件业务场景
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('文件业务场景', 'sys_file_scene', 1, '文件管理业务场景列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '用户头像', 'avatar', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告图片', 'notice_image', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告附件', 'notice_attachment', 3, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_file_scene', '其他', 'other', 4, 'default', 1, NOW(), NOW());

