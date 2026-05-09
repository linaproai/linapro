## ADDED Requirements

### Requirement: 租户管理员的插件启用/禁用接口
对于 `install_mode = tenant_scoped` 的 `tenant_aware` 插件,租户管理员 SHALL 可通过 `POST /tenant/plugins/{plugin_id}/enable` 与 `POST /tenant/plugins/{plugin_id}/disable` 控制本租户的启用状态;不能影响其他租户。

#### Scenario: 租户管理员启用 monitor-loginlog
- **WHEN** 租户 A 管理员调用 `POST /tenant/plugins/monitor-loginlog/enable`
- **AND** 该插件 `install_mode = tenant_scoped`
- **THEN** 系统 upsert `sys_plugin_state(plugin_id, tenant_id=A, enabled=true)`
- **AND** 租户 A 内的用户立即可见登录日志菜单
- **AND** 租户 B/C 不受影响

#### Scenario: 租户管理员禁用插件
- **WHEN** 租户 A 管理员调用 `POST /tenant/plugins/monitor-loginlog/disable`
- **THEN** 系统调用插件 `CanTenantDisable(ctx, tenant_id=A)` 否决钩子
- **AND** 全部通过后写 `enabled=false`,触发缓存失效
- **AND** 租户 A 内菜单/路由立即不可见

### Requirement: 租户视图的可见性约束
`GET /tenant/plugins` SHALL 仅返回:
- `scope_nature = tenant_aware` 且 `install_mode = tenant_scoped` 的插件;
- 安装状态为 `installed`(未卸载);
- 不显示 `platform_only` 插件,即便它们 enabled。

#### Scenario: 租户管理员查询插件列表
- **WHEN** 租户 A 管理员调用 `GET /tenant/plugins`
- **THEN** 返回结果按上述规则过滤
- **AND** 每条目附带本租户的当前 `enabled` 状态(从 `sys_plugin_state.tenant_id=A` 读取)

### Requirement: tenant_scoped 插件的路由命中策略
`tenant_scoped` 插件的 HTTP 路由 SHALL 在框架启动期全局挂载;请求到达时,中间件根据 `(tenant_id, plugin_id)` 启用状态决定:
- 已启用:正常路由到插件 handler。
- 未启用:返回 404 `bizerr.CodePluginNotEnabledForTenant`。
- 平台管理员请求:即使租户未启用,平台管理员的运维请求(以 `X-Platform-Probe: true`)允许通过(为运维查看)。

#### Scenario: 租户未启用时请求被拒
- **WHEN** 租户 B 用户请求 `/api/loginlog/list`
- **AND** `monitor-loginlog` 在租户 B 中 `enabled=false`
- **THEN** 返回 404 `bizerr.CodePluginNotEnabledForTenant`
- **AND** 不进入插件 handler

#### Scenario: 已启用时正常路由
- **WHEN** 租户 A 用户请求 `/api/loginlog/list`
- **AND** `monitor-loginlog` 在租户 A 中 `enabled=true`
- **THEN** 路由进入插件 handler,正常返回数据

### Requirement: 启用状态缓存与失效
`(plugin_id, tenant_id) → enabled` 映射 SHALL 缓存于 `pluginruntimecache`,key 携带租户维度;启用/禁用变更触发该 `(plugin_id, tenant_id)` 维度失效,集群模式下广播。

#### Scenario: 启用后立即生效
- **WHEN** 租户管理员 enable 某插件
- **THEN** 缓存失效广播在 1s 内传播到所有节点
- **AND** 立即可见效果(无需重启)

### Requirement: 已启用插件的菜单/权限投影
租户级 `enabled = true` 的插件,其声明的菜单与权限点 SHALL 进入该租户的菜单/权限解析结果;`enabled = false` 时不参与解析,即使用户被分配了该插件的权限点也无法看到对应菜单。

#### Scenario: 启用后菜单出现
- **WHEN** 租户 A enable `monitor-online`
- **AND** 租户 A 中存在被分配 `system:online:list` 权限的用户
- **THEN** 该用户登录后菜单中出现"在线用户"项

#### Scenario: 禁用后菜单消失
- **WHEN** 租户 A disable `monitor-online`
- **THEN** 该用户菜单中"在线用户"项立即消失
- **AND** 其权限分配数据保留(下次启用恢复)
