## ADDED Requirements

### Requirement: org-center Dept Table Tenantization
`plugin_org_center_dept` and `plugin_org_center_user_dept` SHALL add `tenant_id` column; dept tree build/query/modify all filtered by `tenant_id = bizctx.TenantId`.

### Requirement: Tenant Create Default Dept Not Auto-Modeled
`org-center` plugin SHALL NOT depend on unimplemented `tenant.created` event bus to auto-create tenant root department.

### Requirement: orgcap.Provider Filters by Tenant View
`org-center` plugin's `orgcap.Provider` interface methods SHALL internally read `bizctx.TenantId` and filter own tenant data.
