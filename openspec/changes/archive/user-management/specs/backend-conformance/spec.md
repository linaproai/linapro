## ADDED Requirements

### Requirement: DAO Must Pass tenantcap.Apply
All service layer code reading tenant-sensitive `sys_*` tables, tenant-scoped runtime state tables, or plugin business tables with `tenant_id` SHALL inject tenant filtering through `tenantcap.Apply(ctx, model, "tenant_id")`; direct DAO access without seam is code violation.

### Requirement: DO Write Must Fill tenant_id
All `do.*` data operations corresponding to tables with `tenant_id` SHALL explicitly fill `TenantId = bizctx.TenantId(ctx)` at service layer; prohibit relying on default 0 implicit write.

### Requirement: Veto Hook reason Uses i18n Key
Plugin implementing `LifecycleGuard.*` interfaces, reason string SHALL be i18n key, prohibit hardcoded text.

### Requirement: Cross-Tenant Operations Must Use Dedicated API/Service
Normal business paths SHALL not allow cross-tenant read/write; cross-tenant operations must go through `/platform/*` interfaces or `*.PlatformXxx` host service methods.

### Requirement: Startup Tenant Context Construction
Background workers, scheduled tasks, message consumers SHALL when creating new ctx explicitly construct tenant context from task metadata; prohibit accessing tenant-sensitive `sys_*` business data with empty ctx.
