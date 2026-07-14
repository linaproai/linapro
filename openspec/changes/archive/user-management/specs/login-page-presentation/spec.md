# login-page-presentation Specification

## Purpose

定义标准登录页公开入口展示、已交付账号辅助子流程可达性、登录面板布局与描述文案配置。

## Requirements

### Requirement: 标准登录页暴露用户名密码与已交付账号辅助入口

系统 SHALL 将用户名/密码登录作为默认公开登录能力，并允许在系统开关开启时展示已交付的忘记密码与创建账号入口。手机号登录、扫码登录不得作为正式公开能力展示。

#### Scenario: 标准登录页显示用户名密码与账号辅助入口
- **当** 未认证用户访问 `/auth/login` 且忘记密码、创建账号开关均开启时
- **则** 页面显示用户名、密码、记住我和登录控件
- **且** 页面显示忘记密码入口与创建账号入口
- **且** 页面不显示手机号登录或扫码登录入口

#### Scenario: 用户访问未交付的认证子路由
- **当** 用户访问 `/auth/code-login` 或 `/auth/qrcode-login` 时
- **则** 系统重定向回标准登录页 `/auth/login`
- **且** 页面仍不显示手机号登录或扫码登录入口

#### Scenario: 用户访问已交付的账号辅助子路由
- **当** 忘记密码与创建账号开关开启且用户访问 `/auth/forget-password` 或 `/auth/register` 时
- **则** 系统渲染对应认证子页，而不是重定向到 `/auth/login`

### Requirement: 登录面板默认右侧布局并支持位置配置

系统 SHALL 默认以右侧布局渲染登录面板，并允许宿主公共前端配置在左、中、右布局间切换。

#### Scenario: 无覆盖时默认右侧布局
- **当** 浏览器加载登录页且宿主未提供登录面板位置覆盖时
- **则** 登录页使用 `panel-right` 布局
- **且** 登录面板显示在主内容区右侧

#### Scenario: 宿主配置覆盖登录面板位置
- **当** 宿主公共前端配置返回 `auth.panelLayout` 为 `panel-left`、`panel-center` 或 `panel-right` 时
- **则** 登录页按对应模式渲染
- **且** 登录页工具栏布局切换器仍可在三种布局间切换

### Requirement: 默认登录页描述支持宿主配置

系统 SHALL 在宿主未提供覆盖时展示默认登录页描述，并在公共前端配置提供覆盖值时展示配置内容。

#### Scenario: 无覆盖时展示默认描述
- **当** 浏览器加载登录页且宿主未提供 `auth.pageDesc` 覆盖时
- **则** 登录页展示描述 `Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model`

#### Scenario: 宿主配置覆盖登录页描述
- **当** 宿主公共前端配置返回非空 `auth.pageDesc` 时
- **则** 登录页展示返回的描述
