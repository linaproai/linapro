## ADDED Requirements

### Requirement: Plugin bizctx Read-Only Context Snapshot
Host published `bizctx.Service` for source plugins SHALL provide read-only context snapshot method, returning plugin-visible current user, tenant and impersonation metadata in one call.

### Requirement: Plugin Tenant Filtering Public Component
Host SHALL publish reusable tenant filtering helper in `lina-core/pkg/pluginservice/tenantfilter` for source plugins to read plugin-visible business context, inject query conditions by `tenant_id` column, and derive audit fields.

### Requirement: Host Service Auto-Forwards TenantId
All host service methods exposed to plugins SHALL auto-read `bizctx.TenantId` from caller ctx and inject into service internal context; plugin handler does not need to read manually.

### Requirement: Host Service Rejects Cross-Tenant Calls
When plugin ctx `TenantId` does not match host service operation target, SHALL reject; only platform mode (`TenantId=0`) with all data permissions can through explicit `PlatformXxx` host service cross-tenant.
