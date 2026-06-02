## ADDED Requirements

### Requirement: 匿名 `/auth/providers` 响应必须只暴露登录按钮元数据

系统 SHALL 通过 `/auth/providers` 公开接口向匿名调用方仅返回登录按钮渲染所需字段：`providerId`、`pluginId`、`kind`、`name`、`description`、`icon`、`entryUrl`、`displayOrder`、`enabled`。系统 MUST NOT 在该接口的 DTO、service 投影或前端类型上携带 `backendRedirectEnabled`、`backendRedirectDefault`、`backendRedirectRules`、SSO 投递接收方 URL、redirect 规则字典或任何受管员管理的 redirect 行为开关，因为这些字段不属于登录入口渲染语义，又是匿名可见接口不允许暴露的运维配置。

#### Scenario: 匿名调用方请求 provider 列表

- **WHEN** 未登录浏览器请求 `GET /auth/providers`
- **THEN** 响应里每个 provider 条目只包含登录按钮元数据字段
- **AND** 响应里不出现 `backendRedirectEnabled`、`backendRedirectDefault`、`backendRedirectRules` 或等价语义字段
- **AND** 响应里不出现 SSO 接收方 URL 或 redirect 规则字典

#### Scenario: 已登录管理员仍能配置 SSO 投递规则

- **WHEN** 已认证管理员请求 OIDC 插件的 `/plugin/<id>/settings` GET
- **THEN** SSO 投递规则、默认后端跳转、redirect 规则字典仍通过受保护的 settings API 返回给管理员
- **AND** 该 API 受 `Auth+Tenancy+Permission` 中间件保护
- **AND** 这些字段不会被 `/auth/providers` 公开镜像

### Requirement: 公开 provider 列表必须按宿主 provider enablement 过滤

系统 SHALL 在 `auth.ListProviders` 内部使用宿主 `PluginEnabledChecker` 调用插件能力专用的 `pluginSvc.IsProviderEnabled` 来决定 provider 是否出现在 `/auth/providers` 响应里。系统 MUST NOT 使用业务入口可见性语义的 `pluginSvc.IsEnabled` 来过滤公开登录入口，也 MUST NOT 在公开列表中显示「插件已安装但未启用」的 provider 登录按钮。

#### Scenario: 已安装但未启用的认证插件不显示按钮

- **WHEN** `linapro-oidc-google` 插件已经被安装但当前 platform plugin enabled snapshot 中 `enabled=false`
- **AND** 调用 `GET /auth/providers`
- **THEN** 响应中不包含 Google provider 入口
- **AND** 公开列表中不出现该 provider 的按钮元数据

#### Scenario: 启用插件后按钮立即可见

- **WHEN** 管理员在插件管理界面把 `linapro-oidc-google` 切换为启用
- **AND** 平台 enabled snapshot 完成刷新
- **AND** 调用 `GET /auth/providers`
- **THEN** 响应中出现 Google provider 入口
- **AND** Google provider 入口字段保持登录按钮元数据形态

### Requirement: 公开 provider 列表必须避免按 provider 数量的 sys_config 读取

系统 SHALL 将 `/auth/providers` 路径上的 provider `LoginEntry(ctx)` 实现为静态投影，仅返回固定的按钮元数据。源码插件 MUST NOT 在 `LoginEntry(ctx)` 实现中读取 `sys_config` 或访问 `PluginSettingsService`，也 MUST NOT 在该路径上引入按 provider 数量线性增长的数据库访问。

#### Scenario: 登录页聚合不读 sys_config

- **WHEN** 匿名调用方请求 `GET /auth/providers`
- **AND** 系统已注册 N 个源码 provider
- **THEN** 处理路径中不发起按 provider 数量线性增长的 `sys_config` 查询
- **AND** 不通过 provider 的 `LoginEntry(ctx)` 进入 `PluginSettingsService.List` 或类似接口

#### Scenario: SSO 投递配置仍可通过受保护接口读取

- **WHEN** 已认证管理员请求 OIDC 插件的 `/plugin/<id>/settings` GET
- **THEN** SSO 投递规则、默认后端跳转、redirect 规则字典通过 `PluginSettingsService.List` 一次性读取
- **AND** 该路径与公开登录入口路径互不重叠
