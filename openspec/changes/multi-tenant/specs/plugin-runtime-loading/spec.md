## ADDED Requirements

### Requirement: 路由全局挂载 + 请求时按租户过滤
`tenant_scoped` 插件的 HTTP 路由 SHALL 在框架启动期一次性全局挂载;请求到达时由中间件按 `(tenant_id, plugin_id)` 启用状态决定是否进入 handler。

#### Scenario: 启动期挂载所有插件路由
- **WHEN** 框架启动
- **THEN** 所有已安装插件的路由(无论 install_mode)都被挂载到 GoFrame router
- **AND** 不为每个租户单独构造路由表

#### Scenario: 请求时按租户过滤
- **WHEN** 租户 B 用户访问 `/api/loginlog/list`
- **AND** `monitor-loginlog` 在租户 B 中 disabled
- **THEN** 中间件在 handler 之前返回 404 `bizerr.CodePluginNotEnabledForTenant`
- **AND** 不进入插件代码

### Requirement: 菜单解析按租户启用过滤
`menu-management` 解析当前用户的菜单时 SHALL 排除 `(plugin_id, tenant_id)` 未启用的插件菜单项;即使用户被分配了对应权限。

#### Scenario: 未启用插件菜单不出现
- **WHEN** 用户在租户 A 持有 `system:online:list` 权限
- **AND** `monitor-online` 在租户 A 中 disabled
- **THEN** 菜单中不出现"在线用户"项

### Requirement: 启用状态变更立即生效
插件启用/禁用变更 SHALL 通过 `pluginruntimecache` 失效广播 + 集群消息使所有节点立即感知;无需重启服务,期望延迟 < 1s。

#### Scenario: 集群环境下立即生效
- **WHEN** 集群节点 N1 上租户管理员 enable 某插件
- **THEN** 失效广播在 1s 内到达 N2、N3
- **AND** N2、N3 立即按新状态过滤路由

### Requirement: 启用状态读取的缓存 key
`IsEnabled(ctx, pluginID)` 缓存 key SHALL 为 `(plugin_id, tenant_id)`;查询逻辑遵循 `plugin-install-mode` 中定义的解析顺序。

#### Scenario: 缓存隔离
- **WHEN** 租户 A 与租户 B 同一 plugin 启用状态不同
- **THEN** 缓存中存在两个独立条目
- **AND** 两租户读取互不影响
