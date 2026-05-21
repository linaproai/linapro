## ADDED Requirements

### Requirement: Login Flow Supports Tenant Awareness
Login interface `POST /auth/login` SHALL in multi-tenant mode execute "authenticate first, select tenant second" two-phase; see `tenant-aware-authentication` capability spec.

### Requirement: JWT Claims Must Carry TenantId
JWT issuance logic SHALL include `TenantId` in Claims; issuance forces writing `bizctx.TenantId` current value.

### Requirement: Tenant Switch Interface
System SHALL provide `POST /auth/switch-tenant {target_tenant_id}` interface; success MUST immediately invalidate old token and issue new token.

### Requirement: Platform Admin Impersonation Interface
System SHALL provide `POST /platform/tenants/{id}/impersonate` interface; only platform admin can call; non-platform admin MUST return 403.
