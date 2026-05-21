## ADDED Requirements

### Requirement: sys_config Adds tenant_id and Override Semantics
`sys_config` SHALL add `tenant_id INT NOT NULL DEFAULT 0`; UNIQUE constraint changed to `(tenant_id, config_key)`; read path platform fallback, write path defaults to current tenant.

### Requirement: Tenant Config Fallback Row Metadata
Tenant context querying config list SHALL for platform default fallback rows return source and action metadata.

### Requirement: Config Fallback Actions Must Avoid Must-Fail Detail Requests
Frontend SHALL use config row action metadata to determine operation buttons.

### Requirement: Config Cache Invalidation by Tenant
Config cache key SHALL carry `tenant_id`; tenant write MUST trigger own tenant config cache invalidation; platform default write MUST trigger all-tenant cascade invalidation.
