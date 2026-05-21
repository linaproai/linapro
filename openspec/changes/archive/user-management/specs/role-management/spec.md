## ADDED Requirements

### Requirement: sys_role Adds Tenant Ownership and Data Permissions
`sys_role` SHALL add: `tenant_id INT NOT NULL DEFAULT 0` (platform context = 0, tenant context = owning tenant); `data_scope SMALLINT NOT NULL DEFAULT 2` (`1=all data, 2=tenant data, 3=dept data, 4=personal data`); no platform role boolean field.

### Requirement: Role Query Isolated by Hierarchy
Tenant context `GET /role/list` SHALL only return `tenant_id = current_tenant` roles; platform context can query all via platform governance interface.

### Requirement: Unified system Permission Namespace
Permission point strings SHALL only express "resource + action", not platform/tenant boundary. Host and multi-tenant plugin built-in management permissions SHALL unified use `system:*` form.

### Requirement: Tenant Context Prohibits All Data
`data_scope=1` SHALL represent platform global all data, only allowed in `tenant_id=0` platform context roles. Tenant context roles can only configure `data_scope` as `2`, `3` or `4`.

### Requirement: Role Authorization Must Limit to Current Context Assignable Permissions
System SHALL use same assignable permissions set in role add, edit and role menu authorization tree; tenant context can only assign current tenant business-allowed permissions; platform tenant management, plugin governance, global menu governance write permissions and other platform-only permissions must not appear in tenant role authorization tree.

### Requirement: Abnormal Historical Authorization Must Not Elevate Tenant Access Boundary
System SHALL ensure abnormal historical authorization does not elevate tenant context to platform context at user permission resolution and protected API boundaries.
