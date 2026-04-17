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
-- 参数初始化数据：宿主内置运行时参数
-- ============================================================
INSERT INTO `sys_config` (`name`, `key`, `value`, `remark`, `created_at`, `updated_at`) VALUES
('认证管理-JWT Token 有效期', 'sys.jwt.expire', '24h', '控制新签发 JWT Token 的有效期，支持 Go duration 格式，如 12h、24h。', NOW(), NOW()),
('在线用户-会话超时时间', 'sys.session.timeout', '24h', '控制在线会话无活动超时时长，支持 Go duration 格式，如 30m、24h。', NOW(), NOW()),
('文件管理-上传大小上限', 'sys.upload.maxSize', '10', '控制单个上传文件大小上限，单位为 MB，必须为正整数。', NOW(), NOW()),
('用户登录-IP 黑名单列表', 'sys.login.blackIPList', '', '禁止登录的 IP 或 CIDR 网段，多个值以英文分号分隔，例如 127.0.0.1;10.0.0.0/8。', NOW(), NOW()),
('用户管理-账号初始密码', 'sys.user.initPassword', '123456', '用户重置密码弹窗默认回填值，长度必须为 5-20 个字符。', NOW(), NOW())
ON DUPLICATE KEY UPDATE
`name` = VALUES(`name`),
`value` = VALUES(`value`),
`remark` = VALUES(`remark`),
`updated_at` = VALUES(`updated_at`);

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
