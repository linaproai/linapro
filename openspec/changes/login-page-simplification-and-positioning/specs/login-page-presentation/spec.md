## ADDED Requirements

### Requirement: 当前阶段仅暴露账号密码登录入口

系统在当前阶段 MUST 只暴露账号密码登录能力，不得继续展示或保留未实现的认证入口作为正式可访问能力。

#### Scenario: 标准登录页仅显示账号密码登录表单
- **WHEN** 未登录用户访问 `/auth/login`
- **THEN** 页面显示用户名、密码、记住我和登录按钮
- **AND** 页面不显示忘记密码、注册账号、手机号登录、扫码登录和第三方登录入口

#### Scenario: 访问未实现的认证子路由
- **WHEN** 用户访问 `/auth/code-login`、`/auth/qrcode-login`、`/auth/forget-password` 或 `/auth/register`
- **THEN** 系统回退到标准登录页 `/auth/login`
- **AND** 页面仍然只暴露账号密码登录能力

### Requirement: 登录框默认居右且支持位置配置

系统 MUST 让登录页默认以居右布局展示登录框，并支持通过宿主公共前端配置切换为左侧、居中或右侧布局。

#### Scenario: 未配置覆盖值时默认居右展示
- **WHEN** 浏览器加载登录页且宿主未提供登录框位置覆盖值
- **THEN** 登录页使用 `panel-right` 布局
- **AND** 登录框在页面主区域右侧展示

#### Scenario: 宿主配置覆盖登录框位置
- **WHEN** 宿主公共前端配置返回 `auth.panelLayout` 为 `panel-left`、`panel-center` 或 `panel-right`
- **THEN** 登录页使用对应的布局模式渲染登录框
- **AND** 登录页工具栏中的布局切换入口仍然可用于在三种布局之间切换

### Requirement: 登录页默认说明文案支持宿主配置

系统 MUST 在宿主未提供说明文案覆盖值时展示默认登录页说明文案，并在宿主提供公共前端配置覆盖值时显示对应内容。

#### Scenario: 未配置覆盖值时显示默认说明文案
- **WHEN** 浏览器加载登录页且宿主未提供 `auth.pageDesc` 覆盖值
- **THEN** 登录页显示说明文案 `面向业务演进，提供开箱即用的管理入口与灵活可插拔的扩展机制`

#### Scenario: 宿主配置覆盖登录页说明文案
- **WHEN** 宿主公共前端配置返回非空 `auth.pageDesc`
- **THEN** 登录页显示返回的说明文案
