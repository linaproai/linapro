## ADDED Requirements

### Requirement: 演示只读模式由 linapro-ops-demo-guard 的启用状态控制

系统 SHALL 将 `linapro-ops-demo-guard` 的已安装并启用状态视为演示保护的运行时开关。`plugin.autoEnable` 仅控制启动安装和启用；启动后不得将其视为单独的运行时开关。

#### Scenario: 默认配置下演示保护保持禁用
- **WHEN** 宿主以默认交付配置启动且 `plugin.autoEnable` 不包含 `linapro-ops-demo-guard`
- **THEN** 宿主不安装或启用 `linapro-ops-demo-guard`
- **AND** 从未启用该插件的部署默认不阻断写操作

#### Scenario: 手动启用激活演示保护
- **WHEN** 管理员安装并启用 `linapro-ops-demo-guard`
- **THEN** 演示守卫中间件对后续请求生效
- **AND** 写请求被只读演示规则阻断

### Requirement: 宿主必须随源码树交付 linapro-ops-demo-guard 源码插件

系统 SHALL 交付名为 `linapro-ops-demo-guard` 的官方源码插件，使部署可通过启动配置或插件治理启用该能力。宿主不得再将旧 ID `demo-control` 作为官方演示只读保护插件 ID 暴露。

#### Scenario: 宿主发现 linapro-ops-demo-guard 源码插件
- **WHEN** 宿主扫描源码插件并同步注册表数据
- **THEN** 发现 `linapro-ops-demo-guard`
- **AND** 运维人员可决定是否启用
- **AND** 插件列表不得再出现 `demo-control` 官方插件项

### Requirement: linapro-ops-demo-guard 插件启用时必须阻断系统写操作

启用时，`linapro-ops-demo-guard` SHALL 按 HTTP 方法语义阻断系统写请求，同时允许读式请求。

#### Scenario: 禁用时无写拦截
- **WHEN** `linapro-ops-demo-guard` 未启用
- **THEN** `POST`、`PUT` 和 `DELETE` 请求不被演示守卫拒绝

#### Scenario: 查询式请求保持允许
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求使用 `GET`、`HEAD` 或 `OPTIONS`
- **THEN** 演示守卫允许请求继续

#### Scenario: 写请求被拒绝
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求使用 `POST`、`PUT` 或 `DELETE`
- **THEN** 演示守卫以清晰的只读演示消息拒绝请求
- **AND** 请求不继续进入业务处理

### Requirement: linapro-ops-demo-guard 插件必须保留最小会话白名单

系统 SHALL 在 `linapro-ops-demo-guard` 启用时保留登录、令牌刷新、租户选择、租户切换和退出行为，使演示环境保持可用。

#### Scenario: 登录保持允许
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `POST /api/v1/auth/login`
- **THEN** 演示守卫允许请求继续

#### Scenario: 退出保持允许
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `POST /api/v1/auth/logout`
- **THEN** 演示守卫允许请求继续

#### Scenario: 多租户会话切换保持允许
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `POST /api/v1/auth/select-tenant` 或 `POST /api/v1/auth/switch-tenant`
- **THEN** 演示守卫允许请求继续

### Requirement: linapro-ops-demo-guard 插件启用时必须拒绝插件治理写操作

`linapro-ops-demo-guard` 启用时，系统 SHALL 拒绝插件治理写操作，包括插件同步、动态包上传、安装、卸载、启用和禁用。插件管理的 `GET`、`HEAD` 和 `OPTIONS` 请求作为只读操作保持允许。

#### Scenario: 启用 linapro-ops-demo-guard 时拒绝插件安装
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `POST /api/v1/plugins/{id}/install`
- **THEN** 演示守卫以只读演示消息拒绝请求

#### Scenario: 拒绝插件启用和禁用请求
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `PUT /api/v1/plugins/{id}/enable` 或 `PUT /api/v1/plugins/{id}/disable`
- **THEN** 演示守卫以只读演示消息拒绝请求

#### Scenario: 拒绝插件卸载
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `DELETE /api/v1/plugins/{id}`
- **THEN** 演示守卫以只读演示消息拒绝请求

#### Scenario: 拒绝插件同步和上传写操作
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为 `POST /api/v1/plugins/sync` 或 `POST /api/v1/plugins/dynamic/package`
- **THEN** 演示守卫以只读演示消息拒绝请求

#### Scenario: 插件管理读取保持允许
- **WHEN** `linapro-ops-demo-guard` 已启用
- **AND** 请求为使用 `GET`、`HEAD` 或 `OPTIONS` 的插件管理查询
- **THEN** 演示守卫允许请求继续

## REMOVED Requirements

### Requirement: 演示只读模式由 demo-control 的启用状态控制系统

**Reason**: 该要求绑定旧官方插件 ID `demo-control`。本变更将该插件破坏式重命名为 `linapro-ops-demo-guard`，不保留旧 ID 兼容。

**Migration**: 使用新增要求 `演示只读模式由 linapro-ops-demo-guard 的启用状态控制`。

### Requirement: 宿主必须随源码树交付 demo-control 源码插件

**Reason**: 官方演示只读保护插件 ID 从 `demo-control` 改为 `linapro-ops-demo-guard`。

**Migration**: 使用 `apps/lina-plugins/linapro-ops-demo-guard` 和 manifest ID `linapro-ops-demo-guard`。

### Requirement: demo-control 插件启用时必须阻断系统写操作

**Reason**: 守卫行为保留，但旧插件 ID `demo-control` 不再是有效运行时身份。

**Migration**: 使用新增要求 `linapro-ops-demo-guard 插件启用时必须阻断系统写操作`。

### Requirement: demo-control 插件必须保留最小会话白名单

**Reason**: 会话白名单行为保留，但旧插件 ID `demo-control` 不再是有效运行时身份。

**Migration**: 使用新增要求 `linapro-ops-demo-guard 插件必须保留最小会话白名单`。

### Requirement: demo-control 插件启用时必须拒绝插件治理写操作

**Reason**: 插件治理写操作拒绝行为保留，但旧插件 ID `demo-control` 不再是有效运行时身份。

**Migration**: 使用新增要求 `linapro-ops-demo-guard 插件启用时必须拒绝插件治理写操作`。
