-- 007: Config Management and Job Scheduling Module
-- 007：参数设置与任务调度模块
-- Includes config parameter table, login-status dictionary, file-scene dictionary, and job-status dictionary.
-- 包含：参数设置表、登录状态字典、文件场景字典、任务状态字典

-- ----------------------------
-- 1. Config parameter table
-- 1. 参数设置表
-- ----------------------------
CREATE TABLE IF NOT EXISTS `sys_config` (
    `id`         BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT COMMENT  'Config parameter ID',
    `name`       VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT  'Config parameter name',
    `key`        VARCHAR(100)     NOT NULL DEFAULT ''     COMMENT  'Config parameter key',
    `value`      VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT  'Config parameter value',
    `is_builtin` TINYINT          NOT NULL DEFAULT 0      COMMENT  'Built-in record flag: 1=yes, 0=no',
    `remark`     VARCHAR(500)     NOT NULL DEFAULT ''     COMMENT  'Remark',
    `created_at` DATETIME         DEFAULT NULL            COMMENT  'Creation time',
    `updated_at` DATETIME         DEFAULT NULL            COMMENT  'Modification time',
    `deleted_at` DATETIME         DEFAULT NULL            COMMENT  'Deletion time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT= 'Config parameter table';

-- ============================================================
-- Config seed data: host built-in runtime parameters and public frontend display parameters
-- 参数初始化数据：宿主内置运行时参数与公开前端展示参数
-- ============================================================
INSERT IGNORE INTO `sys_config` (`name`, `key`, `value`, `is_builtin`, `remark`, `created_at`, `updated_at`) VALUES
('品牌展示-应用名称', 'sys.app.name', 'LinaPro.AI', 1, '控制浏览器标题、登录页品牌名称和工作台Logo文案展示，建议填写简洁的产品名称。', NOW(), NOW()),
('品牌展示-应用 Logo', 'sys.app.logo', '/logo.webp', 1, '控制登录页与工作台默认 Logo 图片地址，支持 http(s) 或站内绝对路径。', NOW(), NOW()),
('品牌展示-深色 Logo', 'sys.app.logoDark', '/logo.webp', 1, '控制深色主题下的 Logo 图片地址，支持 http(s) 或站内绝对路径。', NOW(), NOW()),
('用户管理-默认头像', 'sys.user.defaultAvatar', '/avatar.webp', 1, '控制用户未设置头像时的默认头像地址，支持 http(s) 或站内绝对路径。', NOW(), NOW()),
('登录展示-页面标题', 'sys.auth.pageTitle', '面向可持续交付的 AI 原生全栈框架', 1, '控制登录页顶部主标题文案。', NOW(), NOW()),
('登录展示-页面说明', 'sys.auth.pageDesc', '帮助团队快速交付生产级应用，同时保持架构、测试与治理的可持续演进', 1, '控制登录页顶部说明文案。', NOW(), NOW()),
('登录展示-登录副标题', 'sys.auth.loginSubtitle', '请输入您的帐户信息以开始管理您的项目', 1, '控制登录表单上方的提示说明文案。', NOW(), NOW()),
('登录展示-登录框位置', 'sys.auth.loginPanelLayout', 'panel-right', 1, '控制登录框默认布局，可选值：panel-left、panel-center、panel-right。', NOW(), NOW()),
('认证管理-JWT Token 有效期', 'sys.jwt.expire', '24h', 1, '控制新签发 JWT Token 的有效期，支持 Go duration 格式如 12h、24h。', NOW(), NOW()),
('在线用户-会话超时时间', 'sys.session.timeout', '24h', 1, '控制在线会话无活动超时时长，支持 Go duration 格式，如 30m、24h。', NOW(), NOW()),
('文件管理-上传大小上限', 'sys.upload.maxSize', '20', 1, '控制单个上传文件大小上限，单位为 MB，必须为正整数。', NOW(), NOW()),
('用户登录-IP 黑名单列表', 'sys.login.blackIPList', '', 1, '禁止登录的 IP 或 CIDR 网段，多个值以英文分号分隔，例如 127.0.0.1;10.0.0.0/8。', NOW(), NOW()),
('界面风格-主题模式', 'sys.ui.theme.mode', 'light', 1, '控制默认主题模式，可选值：light、dark、auto。', NOW(), NOW()),
('界面风格-工作台布局', 'sys.ui.layout', 'sidebar-nav', 1, '控制后台默认布局，可选值：sidebar-nav、sidebar-mixed-nav、header-nav、header-sidebar-nav、header-mixed-nav、mixed-nav、full-content。', NOW(), NOW()),
('界面风格-是否启用水印', 'sys.ui.watermark.enabled', 'false', 1, '控制工作台是否启用水印，可选值：true、false。', NOW(), NOW()),
('界面风格-水印文案', 'sys.ui.watermark.content', 'LinaPro', 1, '控制工作台水印文案内容。', NOW(), NOW());

-- ============================================================
-- Dictionary seed data: login status
-- 字典初始化数据：登录状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('登录状态', 'sys_login_status', 1, 1, '登录日志状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_login_status', '成功', '0', 1, 'success', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_login_status', '失败', '1', 2, 'danger', 1, 1, NOW(), NOW());

-- ============================================================
-- Dictionary seed data: file business scenes
-- 字典初始化数据：文件业务场景
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('文件业务场景', 'sys_file_scene', 1, 1, '文件管理业务场景列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_file_scene', '用户头像', 'avatar', 1, 'primary', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告图片', 'notice_image', 2, 'success', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_file_scene', '通知公告附件', 'notice_attachment', 3, 'warning', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_file_scene', '其他', 'other', 4, 'default', 1, 1, NOW(), NOW());
