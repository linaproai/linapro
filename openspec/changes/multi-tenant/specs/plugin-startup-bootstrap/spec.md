## ADDED Requirements

### Requirement: 启动期按 (plugin, tenant) 装配状态缓存
框架启动期 SHALL 一次性从 `sys_plugin_state` 读取所有 `(plugin_id, tenant_id, enabled)` 行,装配到 `pluginruntimecache`;后续运行时仅按需失效。

#### Scenario: 启动期初始装配
- **WHEN** 框架冷启动
- **THEN** 一次 `SELECT * FROM sys_plugin_state` 装配缓存
- **AND** 后续 `IsEnabled` 命中缓存,无 DB 调用

### Requirement: 启动期一致性校验
启动期 SHALL 执行:
1. `sys_plugin.scope_nature` 与 `install_mode` 一致性(platform_only ↔ global)。
2. `sys_role.is_platform_role=true` 必须 `tenant_id=0`。
3. `sys_user.tenant_id=0` 必须无 active membership。
4. `multi-tenant` 插件状态与 `tenantcap.Provider` 注册一致(已 enabled 则 Provider 必须存在)。

任何一致性检查失败 SHALL 阻止启动并打印明确错误。

#### Scenario: 不一致状态阻止启动
- **WHEN** 直接 SQL 修改导致 `(scope_nature=platform_only, install_mode=tenant_scoped)`
- **THEN** 启动失败,日志明确指出具体插件 id 与建议
- **AND** 服务进程退出

### Requirement: multi-tenant 插件未启用时短路
若 `multi-tenant` 插件未启用或未注册 Provider,启动期 SHALL 装配 no-op `tenantcap.Service`;所有 `Apply` 调用 no-op,系统行为等价于单租户。

#### Scenario: 短路模式
- **WHEN** multi-tenant 插件未启用
- **THEN** `tenantcap.Service.Enabled() = false`
- **AND** 中间件链路存在但仅注入 `TenantId=0` 后放行

### Requirement: tenant.created 事件触发 default_for_new_tenants 处理
启动期完成后,`multi-tenant` 插件 SHALL 注册自己监听 `tenant.created` 的处理器,以便后续创建租户时自动初始化 `default_for_new_tenants=true` 插件的 `sys_plugin_state` 行。

#### Scenario: 新租户创建自动初始化插件状态
- **WHEN** 平台管理员创建租户 T
- **AND** `org-center` 安装为 `tenant_scoped` + `default_for_new_tenants=true`
- **THEN** 自动 insert `sys_plugin_state(plugin_id=org-center, tenant_id=T, enabled=true)`
