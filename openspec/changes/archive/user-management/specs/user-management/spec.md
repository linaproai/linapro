## ADDED Requirements

### Requirement: sys_user Adds Tenant Identity Field
`sys_user` SHALL add `tenant_id INT NOT NULL DEFAULT 0` field: `tenant_id = 0` represents platform user (platform admin, all users in single-tenant mode); `tenant_id > 0` represents user's "primary tenant/default login tenant"; coexists with 1:N membership model.

### Requirement: User Query Tenant Isolation
When multi-tenant enabled, user list/detail queries SHALL use `plugin_multi_tenant_user_membership` as tenant visibility authoritative boundary; `sys_user.tenant_id` only represents primary tenant, not sole filter condition. Tenant A admin only sees `membership.tenant_id = A AND status = active` users; platform admin queries full users with tenant ownership filter.

### Requirement: User Create/Import Tenant-Scoped Write
Tenant admin creating users via `POST /user` and `POST /user/import` SHALL auto-write `tenant_id = bizctx.TenantId`; auto-create membership; cross-tenant write not allowed.

### Requirement: User Batch Edit
User management page SHALL support batch editing selected users for status, roles, and tenant ownership changes; backend processes in single transaction; any target user invisible, unauthorized, or illegal causes overall rejection.

### Requirement: Username Globally Unique
`sys_user.username` SHALL remain globally unique; different tenants cannot have same username.

### Requirement: Invite Existing User to Join Tenant
`POST /tenant/members/invite` SHALL allow tenant admin to invite existing global user to join own tenant (only creates membership, no new sys_user row).
