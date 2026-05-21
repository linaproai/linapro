## ADDED Requirements

### Requirement: org-center Post Table Tenantization
`plugin_org_center_post` and `plugin_org_center_user_post` SHALL add `tenant_id` column; post CRUD filtered by `tenant_id = bizctx.TenantId`.

### Requirement: Intra-Tenant Post Code Unique
Post `code` uniqueness constraint SHALL be on `(tenant_id, code)`; different tenants can reuse same code.

### Requirement: Tenant Delete Does Not Cascade Clean Posts
`org-center` SHALL NOT depend on unimplemented `tenant.deleted` event bus to clean tenant posts.
