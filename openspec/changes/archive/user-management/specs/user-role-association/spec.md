## ADDED Requirements

### Requirement: sys_user_role Adds Tenant Dimension
`sys_user_role` SHALL add `tenant_id INT NOT NULL DEFAULT 0` field; UNIQUE constraint changed to `(user_id, role_id, tenant_id)`.

### Requirement: Role Binding Validated by Tenant
Binding SHALL validate `sys_role.tenant_id` matches request `tenant_id`; tenant admin can only bind own tenant roles; platform context roles only bindable to platform users.

### Requirement: Permission Resolution Filters by Current Tenant
Permission/menu resolution SHALL only take `sys_user_role.tenant_id = bizctx.TenantId` associations; platform admin context (`bizctx.TenantId=0`) takes platform roles.

### Requirement: User Logout/Kick Cleanup
When user removed from tenant (membership deleted) or tenant deleted, SHALL cascade delete `sys_user_role.tenant_id = T` related rows, but retain other tenant associations.
