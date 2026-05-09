## ADDED Requirements

### Requirement: host service 自动透传 TenantId
宿主对插件暴露的所有 host service 方法 SHALL 自动从调用方 ctx 中读取 `bizctx.TenantId` 并注入到 service 内部上下文;插件 handler 无需自行读取。

#### Scenario: 透传链路
- **WHEN** 插件 handler 接收请求,调用 `hostsvc.User.List(ctx, ...)`
- **THEN** host service 内部读 `bizctx.TenantId(ctx)` 自动过滤
- **AND** 插件无需手动加 tenant_id 参数

### Requirement: host service 拒绝跨租户调用
插件传入的 ctx `TenantId` 与 host service 操作目标不一致时 SHALL 拒绝;只有管理平台模式(`TenantId=0`)且具备 `platform:*` 权限的调用方可通过显式 `PlatformXxx` host service 跨租户。impersonation 模式不允许全量跨租户 host service 调用。

#### Scenario: 跨租户调用被拒
- **WHEN** 插件在租户 A 上下文中调用 `hostsvc.User.GetById(ctx, userIdInTenantB)`
- **THEN** host service 返回 `bizerr.CodeCrossTenantNotAllowed`
- **AND** 不返回数据

### Requirement: 平台 host service 接口分离
跨租户操作的 host service 方法 SHALL 命名为 `PlatformXxx`(如 `hostsvc.User.PlatformList(ctx)`);只有平台权限可调用,审计自动记录。

#### Scenario: 平台 host 调用审计
- **WHEN** 平台管理员的 ctx 调用 `hostsvc.User.PlatformList(ctx)`
- **THEN** 接口正常返回全租户数据
- **AND** operlog 记录 `action_kind='platform_host_call'`
