## ADDED Requirements

### Requirement: 插件 bizctx 只读上下文快照
宿主发布给源码插件的 `bizctx.Service` SHALL 提供只读上下文快照方法,一次性返回插件可见的当前用户、租户与 impersonation 元数据;该快照不得暴露宿主 `internal/model.Context` 可变指针。

#### Scenario: 插件读取当前业务上下文
- **WHEN** 源码插件调用 `bizctx.Service.Current(ctx)`
- **THEN** 宿主返回插件可见的只读上下文快照
- **AND** 快照包含 `UserID`、`Username`、`TenantID`、`ActingUserID`、`ActingAsTenant`、`IsImpersonation` 与 `PlatformBypass`
- **AND** 插件不能通过该快照修改宿主请求上下文

### Requirement: 插件租户过滤公共组件
宿主 SHALL 在 `lina-core/pkg/pluginservice/tenantfilter` 发布源码插件可复用的租户过滤 helper,用于读取插件可见业务上下文、按约定 `tenant_id` 列注入查询条件,并派生审计所需的代操作租户字段;源码插件不得为相同职责各自复制反射读取 `BizCtx` 的本地实现。

#### Scenario: 插件业务表查询按当前租户过滤
- **WHEN** 源码插件调用 `tenantfilter.Apply(ctx, model)` 或 `tenantfilter.ApplyColumn(ctx, model, column)`
- **THEN** 公共组件从 `bizctx.Service.Current(ctx)` 读取当前 `TenantID`
- **AND** 查询模型追加对应 `tenant_id` 条件
- **AND** join 查询可通过显式列名注入限定条件

#### Scenario: 插件审计字段派生
- **WHEN** 源码插件调用 `tenantfilter.CurrentContext(ctx)` 生成审计元数据
- **THEN** 普通租户请求返回 `OnBehalfOfTenantID=0`
- **AND** impersonation 或代操作请求返回 `OnBehalfOfTenantID=TenantID`
- **AND** `ActingUserID` 优先使用 bizctx 中的真实代操作用户,缺省时回退为当前 `UserID`

### Requirement: host service 自动透传 TenantId
宿主对插件暴露的所有 host service 方法 SHALL 自动从调用方 ctx 中读取 `bizctx.TenantId` 并注入到 service 内部上下文;插件 handler 无需自行读取。

#### Scenario: 透传链路
- **WHEN** 插件 handler 接收请求,调用 `hostsvc.User.List(ctx, ...)`
- **THEN** host service 内部读 `bizctx.TenantId(ctx)` 自动过滤
- **AND** 插件无需手动加 tenant_id 参数

### Requirement: host service 拒绝跨租户调用
插件传入的 ctx `TenantId` 与 host service 操作目标不一致时 SHALL 拒绝;只有管理平台模式(`TenantId=0`)、有效数据权限为全部数据权限且具备对应 `system:*` 功能权限的调用方可通过显式 `PlatformXxx` host service 跨租户。impersonation 模式不允许全量跨租户 host service 调用。

#### Scenario: 跨租户调用被拒
- **WHEN** 插件在租户 A 上下文中调用 `hostsvc.User.GetById(ctx, userIdInTenantB)`
- **THEN** host service 返回 `bizerr.CodeCrossTenantNotAllowed`
- **AND** 不返回数据

### Requirement: 平台 host service 接口分离
跨租户操作的 host service 方法 SHALL 命名为 `PlatformXxx`(如 `hostsvc.User.PlatformList(ctx)`);只有平台权限可调用,审计自动记录。

#### Scenario: 平台 host 调用审计
- **WHEN** 平台管理员的 ctx 调用 `hostsvc.User.PlatformList(ctx)`
- **THEN** 接口正常返回全租户数据
- **AND** operlog 以 `oper_type='other'` 记录平台 host 调用,摘要或载荷包含 `platform_host_call`
