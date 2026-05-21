## ADDED Requirements

### Requirement: Scheduled Job Table Tenantized
`sys_job` and `sys_job_log` SHALL add `tenant_id`; `tenant_id=0` represents platform-level built-in tasks; `tenant_id>0` represents tenant business tasks.

### Requirement: Job Execution Tenant Context Binding
Scheduler SHALL construct `bizctx.Context` with `TenantId` before triggering job execution; handler reads `bizctx.TenantId(ctx)`.

### Requirement: Job Group Tenant Isolation
`sys_job_group` SHALL be treated as tenant-scoped resource. Job group list, detail, create, update, delete, task count and migration MUST use current tenant context limited to corresponding `tenant_id`.

### Requirement: Job Query/Create/Modify Tenant Isolated
Tenant admin can only see/modify own tenant jobs; platform admin manages cross-tenant/platform jobs through `/platform/jobs/*` management interface.
