-- Mock data: 参数设置示例数据

INSERT INTO `sys_config` (`name`, `key`, `value`, `remark`, `created_at`, `updated_at`) VALUES
('主框架页-默认皮肤样式名称', 'sys.index.skinName', 'skin-blue', '蓝色 skin-blue、绿色 skin-green、紫色 skin-purple、红色 skin-red、黄色 skin-yellow', NOW(), NOW()),
('用户管理-账号初始密码',      'sys.user.initPassword', '123456', '初始化密码 123456', NOW(), NOW()),
('主框架页-侧边栏主题',       'sys.index.sideTheme', 'theme-dark', '深色主题 theme-dark、浅色主题 theme-light', NOW(), NOW()),
('账号自助-是否开启用户注册功能', 'sys.account.registerUser', 'false', '是否开启注册用户功能（true开启，false关闭）', NOW(), NOW()),
('用户登录-黑名单列表',        'sys.login.blackIPList', '', '设置登录IP黑名单限制，多个匹配项以;分隔', NOW(), NOW());
