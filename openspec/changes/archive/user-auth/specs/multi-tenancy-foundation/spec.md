## ADDED Requirements

### Requirement: 多租户能力接缝(tenantcap)
宿主 SHALL 在 `pkg/tenantcap` 与 `internal/service/tenantcap` 中维护稳定的多租户能力接缝,使租户主体的具体语义由 `multi-tenant` 插件提供,而宿主仅持有契约与默认 no-op 行为。

#### Scenario: 插件未安装时的默认行为
- **WHEN** `multi-tenant` 插件未安装或未启用
- **THEN** `tenantcap.Service.Enabled(ctx)` 返回 `false`
- **AND** `tenantcap.Service.Apply(ctx, model, col)` 不修改入参 `model`,等价于 no-op
- **AND** `tenantcap.Service.Current(ctx)` 返回 `TenantID = 0`(PLATFORM)
- **AND** 系统行为与未引入多租户能力前等价

#### Scenario: 插件已启用且注册了 Provider
- **WHEN** `multi-tenant` 插件 enabled 并通过 `tenantcap.RegisterProvider` 注册了 Provider
- **THEN** `tenantcap.Service.Enabled(ctx)` 返回 `true`
- **AND** `tenantcap.Service.Apply(ctx, model, col)` 在非平台管理员上下文中追加 `WHERE col = current_tenant_id`
- **AND** 平台管理员管理平台上下文(`TenantId=0` 且 `PlatformBypass(ctx)=true`)下 `Apply` 不注入过滤
- **AND** 平台管理员 impersonation 某租户时不视为全量 bypass,仍追加目标租户过滤

### Requirement: bizctx 租户身份字段
`bizctx.Context` SHALL 增加 `TenantId int` 字段,沿请求链路完整传递,所有依赖租户身份的代码路径必须从 `bizctx` 读取该值,禁止通过其他方式重新解析租户。

#### Scenario: 中间件注入租户身份
- **WHEN** 一个 HTTP 请求经过 `tenancy` 中间件
- **THEN** 中间件将解析得到的 `TenantId` 写入 `bizctx.Context.TenantId`
- **AND** 后续 service / DAO / 日志 / 缓存调用从 `bizctx` 读取该 `TenantId`

#### Scenario: 后台/定时任务上下文
- **WHEN** 定时任务、消息消费者、后台 worker 创建新的 `context.Context`
- **THEN** 必须从任务元数据中取出 `TenantId` 并通过 `bizctx.SetTenant(ctx, id)` 注入
- **AND** 严禁在没有显式 `TenantId` 的情况下访问 `sys_*` 业务数据

### Requirement: Pool 隔离模型与 schema 总则
所有租户敏感 `sys_*` 业务表、租户作用域运行时状态表与插件持有的业务表 SHALL 包含 `tenant_id INT NOT NULL DEFAULT 0` 列,原租户相关索引必须升级为 `(tenant_id, ...)` 联合索引;`tenant_id = 0` 表示 PLATFORM(平台默认),正整数表示具体租户。平台控制面或全局配置表 SHALL NOT 机械增加 `tenant_id`,包括 `sys_locker`、`sys_menu`、`sys_plugin`、`sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state` 与 `sys_notify_channel`。

#### Scenario: 单租户开箱场景
- **WHEN** 用户使用 LinaPro 但未启用多租户能力
- **THEN** 所有租户敏感数据落在 `tenant_id = 0` 上
- **AND** 平台控制面数据保持全局唯一
- **AND** 租户敏感索引性能不下降(联合索引前导列为 0 的常量,等价于原索引)

#### Scenario: 多租户启用后跨租户隔离
- **WHEN** 多租户能力启用,租户 A 的用户查询 `sys_user`
- **THEN** 查询自动追加 `WHERE tenant_id = A`
- **AND** 租户 A 的用户不能查询/创建/修改/删除 `tenant_id != A` 的任意租户敏感 `sys_*` 行(平台管理员除外)

### Requirement: DAO 注入纪律
凡读取或写入带 `tenant_id` 的租户敏感 `sys_*` 表、租户作用域运行时状态表或插件业务表的 service 层代码,SHALL 通过 `tenantcap.Apply(ctx, model, col)`(读)与 service 层 helper(写)注入租户上下文;直接访问此类 DAO 而不经接缝的代码视为违反规范,必须在 `lina-review` 中被拒。平台控制面表不经 `tenantcap.Apply`,但必须通过宿主/平台治理 service 访问。

#### Scenario: 读取查询的注入
- **WHEN** service 调用 `dao.SysUser.Ctx(ctx).Where(...).Scan(&list)`
- **THEN** service 必须先 `tenantcap.Apply(ctx, model, "tenant_id")` 包装 model
- **AND** 不允许散点写 `model.Where("tenant_id", id)`

#### Scenario: 写入数据的注入
- **WHEN** service 调用 `dao.SysUser.Ctx(ctx).Data(do.SysUser{...}).Insert()`
- **THEN** DO 字段必须填入 `TenantId = bizctx.TenantId(ctx)`
- **AND** 平台管理员跨租户写时,必须通过 impersonation 或专用平台 service/API 显式指定目标 `tenant_id`,且记审计

### Requirement: tenancy bypass 与平台管理员
`tenantcap.Service.PlatformBypass(ctx)` SHALL 仅在当前请求处于平台上下文(`bizctx.TenantId = 0`)、未 impersonation、且有效数据权限为 `全部数据权限(data_scope=1)` 时返回 `true`,bypass 后查询不注入 `tenant_id` 过滤。平台管理员显式 impersonation 某租户(`bizctx.TenantId > 0` 且 `bizctx.ActingAsTenant = true`)时 SHALL 返回 `false`,查询/写入必须按目标租户过滤,仅在审计中记录 `on_behalf_of_tenant_id`。

#### Scenario: 平台管理员管理平台
- **WHEN** 平台管理员 `bizctx.TenantId = 0` 查询 `sys_user`
- **THEN** `Apply` 不注入过滤,返回所有租户的用户列表
- **AND** 操作日志 `acting_user_id = 当前用户`,`tenant_id = 0`

#### Scenario: 平台管理员 impersonation 某租户
- **WHEN** 平台管理员切换为"以租户 T 视角操作"
- **THEN** `bizctx.TenantId = T`,`bizctx.ActingAsTenant = true`
- **AND** `PlatformBypass(ctx)=false`
- **AND** 查询/写入按租户 T 视角执行并注入 `tenant_id = T`
- **AND** 操作日志 `acting_user_id = 平台管理员 user_id`,`on_behalf_of_tenant_id = T`

### Requirement: 隔离模型代码默认值
系统 SHALL 在代码中以明确常量定义租户隔离模型默认值,首版固定为 `pool`;宿主 `config.template.yaml` SHALL NOT 提供 `tenant.isolation.mode` 配置项。未来新增 schema-per-tenant 或 db-per-tenant 模式时,应先通过受控管理入口开放设置,业务代码仍通过统一 tenancy/tenantcap 接缝消费有效模式。

#### Scenario: 默认隔离模型
- **WHEN** 系统启动且未安装或未启用 `multi-tenant` 插件
- **THEN** 租户敏感数据按 `tenant_id = 0` 的 PLATFORM 租户处理
- **AND** 隔离模型默认值在代码中记录为 `pool`,不依赖宿主配置文件
