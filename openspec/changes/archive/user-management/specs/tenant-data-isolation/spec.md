## ADDED Requirements

### Requirement: File Storage Path Tenant Prefix
Local and object storage SHALL include tenant prefix in path: local path `/storage/t/{tenant_id}/yyyy/mm/dd/{file_id}`; object storage key `t/{tenant_id}/yyyy/mm/dd/{file_id}`. `tenant_id=0` represents platform shared files.

#### Scenario: Tenant user uploads file
- **WHEN** Tenant A user uploads file
- **THEN** File lands at `/storage/t/A/yyyy/mm/dd/{file_id}`
- **AND** `sys_file.tenant_id = A`

### Requirement: File Read Tenant Validation
File download/preview interface SHALL validate request `bizctx.TenantId` matches `sys_file.tenant_id`; mismatch returns 403. Platform admin in management platform mode can only cross-tenant access through explicit `/platform/*` read-only interface; impersonation mode performs same tenant match validation.

### Requirement: Cache Keys Must Carry Tenant Dimension
All runtime cache keys in tenant-sensitive scenarios SHALL carry `tenant_id` dimension; cache key construction unified through helper `CacheKey(tenant, scope, key)`, prohibit scattered concatenation.

### Requirement: Cluster Invalidation Broadcast Carries Tenant Scope
`distributed-cache-coordination` invalidation messages SHALL carry `tenant_id` field; invalidation must explicitly declare scope:
- Single tenant: `tenant_id = T` (only invalidate that tenant's scope cache).
- Platform default change: `tenant_id = 0` + `cascade_to_tenants = true` (invalidate all tenants' corresponding scope).
- Global fallback (rare): `tenant_id = -1` (explicit "all tenants clear" flag, must have audit log).

### Requirement: Audit Log Tenant and Impersonation Dual-Track
`monitor-operlog` and `monitor-loginlog` SHALL contain:
- `tenant_id`: Operation tenant context (= `bizctx.TenantId`).
- `acting_user_id`: Actual operator's global user_id.
- `on_behalf_of_tenant_id`: Only has value during impersonation.
- `is_impersonation`: Boolean flag.

### Requirement: Cross-Tenant Operations Must Be Explicit
Any "cross-tenant read/write" path SHALL only be available to platform admin, and must go through explicit API (prefixed `/platform/*`) or dedicated platform service; prohibit implicit cross-tenant access in normal business APIs.

### Requirement: Business Table Isolation Test Coverage
e2e test suite SHALL include "cross-tenant isolation" anti-example cases for each tenancy-aware `sys_*` table and plugin business table; at least one case per table verifying "tenant A cannot see tenant B data".
