-- Mock data: 参数设置示例数据

DELETE FROM `sys_config` WHERE `key` = 'sys.user.initPassword';
DELETE FROM `sys_config` WHERE `key` IN ('sys.index.skinName', 'sys.index.sideTheme', 'sys.account.registerUser');

INSERT IGNORE INTO `sys_config` (`name`, `key`, `value`, `remark`, `created_at`, `updated_at`) VALUES
('演示-支持邮箱', 'demo.support.email', 'support@example.com', '仅用于演示自定义参数能力，不被宿主运行时直接消费。', NOW(), NOW()),
('演示-首页公告文案', 'demo.notice.banner', '欢迎使用 LinaPro', '仅用于演示自定义参数能力，不被宿主运行时直接消费。', NOW(), NOW());
