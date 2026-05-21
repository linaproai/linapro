## ADDED Requirements

### Requirement: Multi-Tenant Capability Seam (tenantcap)
Host SHALL maintain stable multi-tenant capability seam in `pkg/tenantcap` and `internal/service/tenantcap`, enabling tenant entity semantics to be provided by `multi-tenant` plugin while host only holds contracts and default no-op behavior.

#### Scenario: Default behavior when plugin not installed
- **WHEN** `multi-tenant` plugin not installed or not enabled
- **THEN** `tenantcap.Service.Enabled(ctx)` returns `false`
- **AND** `tenantcap.Service.Apply(ctx, model, col)` does not modify input `model`, equivalent to no-op
- **AND** `tenantcap.Service.Current(ctx)` returns `TenantID = 0` (PLATFORM)
- **AND** System behavior equivalent to pre-multi-tenant capability

#### Scenario: Plugin enabled and Provider registered
- **WHEN** `multi-tenant` plugin enabled and registered Provider via `tenantcap.RegisterProvider`
- **THEN** `tenantcap.Service.Enabled(ctx)` returns `true`
- **AND** `tenantcap.Service.Apply(ctx, model, col)` appends `WHERE col = current_tenant_id` in non-platform admin context
- **AND** Platform admin management platform context (`TenantId=0` and `PlatformBypass(ctx)=true`) does not inject filtering
- **AND** Platform admin impersonation of a tenant is not full bypass, still appends target tenant filtering

### Requirement: bizctx Tenant Identity Field
`bizctx.Context` SHALL add `TenantId int` field, propagated along request chain; all code paths depending on tenant identity must read from `bizctx`, prohibit re-resolving tenant through other means.

#### Scenario: Middleware injects tenant identity
- **WHEN** HTTP request passes through `tenancy` middleware
- **THEN** Middleware writes resolved `TenantId` to `bizctx.Context.TenantId`
- **AND** Subsequent service/DAO/log/cache calls read `TenantId` from `bizctx`

#### Scenario: Background/scheduled task context
- **WHEN** Scheduled task, message consumer, background worker creates new `context.Context`
- **THEN** Must extract `TenantId` from task metadata and inject via `bizctx.SetTenant(ctx, id)`
- **AND** Prohibit accessing `sys_*` business data without explicit `TenantId`

### Requirement: Pool Isolation Model and Schema General Principles
All tenant-sensitive `sys_*` business tables, tenant-scoped runtime state tables, and plugin-owned business tables SHALL contain `tenant_id INT NOT NULL DEFAULT 0` column; original tenant-related indexes must be upgraded to `(tenant_id, ...)` composite indexes; `tenant_id = 0` represents PLATFORM (platform default), positive integers represent specific tenants. Platform control plane or global config tables SHALL NOT mechanically add `tenant_id`, including `sys_locker`, `sys_menu`, `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, and `sys_notify_channel`.

#### Scenario: Single-tenant out-of-box scenario
- **WHEN** User uses LinaPro but does not enable multi-tenant capability
- **THEN** All tenant-sensitive data falls on `tenant_id = 0`
- **AND** Platform control plane data remains globally unique
- **AND** Tenant-sensitive index performance does not degrade (composite index leading column is constant 0, equivalent to original index)

#### Scenario: Cross-tenant isolation after multi-tenant enabled
- **WHEN** Multi-tenant capability enabled, tenant A user queries `sys_user`
- **THEN** Query auto-appends `WHERE tenant_id = A`
- **AND** Tenant A user cannot query/create/modify/delete any tenant-sensitive `sys_*` row where `tenant_id != A` (platform admin exempted)

### Requirement: DAO Injection Discipline
All service layer code reading or writing tenant-sensitive `sys_*` tables, tenant-scoped runtime state tables, or plugin business tables with `tenant_id` SHALL inject tenant context through `tenantcap.Apply(ctx, model, col)` (read) and service layer helper (write); direct DAO access without seam is code violation, must be rejected in `lina-review`. Platform control plane tables not injected through tenantcap, but only accessible through host/platform governance service.

#### Scenario: Read query injection
- **WHEN** Service calls `dao.SysUser.Ctx(ctx).Where(...).Scan(&list)`
- **THEN** Service must first `tenantcap.Apply(ctx, model, "tenant_id")` wrap model
- **AND** Not allowed scattered `model.Where("tenant_id", id)`

#### Scenario: Write data injection
- **WHEN** Service calls `dao.SysUser.Ctx(ctx).Data(do.SysUser{...}).Insert()`
- **THEN** DO fields must fill `TenantId = bizctx.TenantId(ctx)`
- **AND** Platform admin cross-tenant write must go through impersonation or dedicated platform service/API explicitly specifying target `tenant_id`, with audit

### Requirement: tenancy Bypass and Platform Admin
`tenantcap.Service.PlatformBypass(ctx)` SHALL only return `true` when current request is in platform context (`bizctx.TenantId = 0`), not impersonation, and effective data permission is `all data permissions (data_scope=1)`; bypassed queries do not inject `tenant_id` filtering. Platform admin explicitly impersonating a tenant (`bizctx.TenantId > 0` and `bizctx.ActingAsTenant = true`) SHALL return `false`; queries/writes must filter by target tenant, only recording `on_behalf_of_tenant_id` in audit.

#### Scenario: Platform admin managing platform
- **WHEN** Platform admin `bizctx.TenantId = 0` queries `sys_user`
- **THEN** `Apply` does not inject filtering, returns all tenant user lists
- **AND** Operation log `acting_user_id = current user`, `tenant_id = 0`

#### Scenario: Platform admin impersonation of tenant
- **WHEN** Platform admin switches to "operate as tenant T perspective"
- **THEN** `bizctx.TenantId = T`, `bizctx.ActingAsTenant = true`
- **AND** `PlatformBypass(ctx) = false`
- **AND** Queries/writes execute as tenant T perspective and inject `tenant_id = T`
- **AND** Operation log `acting_user_id = platform admin user_id`, `on_behalf_of_tenant_id = T`

### Requirement: Isolation Model Code Default
System SHALL define tenant isolation model default in code with explicit constant, first version fixed as `pool`; host `config.template.yaml` SHALL NOT provide `tenant.isolation.mode` config item. When adding schema-per-tenant or db-per-tenant mode in future, should first open through controlled management entry, business code still consumes effective mode through unified tenancy/tenantcap seam.

#### Scenario: Default isolation model
- **WHEN** System starts and `multi-tenant` plugin not installed or not enabled
- **THEN** Tenant-sensitive data processed as `tenant_id = 0` PLATFORM tenant
- **AND** Isolation model default recorded as `pool` in code, not dependent on host config file
