## ADDED Requirements

### Requirement: Platform Default + Tenant Override Read Path
For resources supporting tenant override (`sys_dict_type` / `sys_dict_data` / `sys_config`), read SHALL resolve in order:
1. First query `tenant_id = bizctx.TenantId`, if result found return directly.
2. Otherwise fallback to `tenant_id = 0` (PLATFORM default), if found return.
3. Both empty return empty.

Read logic encapsulated in `tenantcap.ReadWithPlatformFallback` helper, business service does not re-implement.

#### Scenario: Tenant override priority
- **WHEN** Tenant A sets `sys_config[key=upload.max_size, tenant_id=A] = 50MB`
- **AND** Platform default `sys_config[key=upload.max_size, tenant_id=0] = 10MB`
- **THEN** Tenant A request reads `50MB`
- **AND** Other tenant requests fallback to `10MB`

#### Scenario: Tenant fallback when no override
- **WHEN** Tenant B has no `sys_config[key=upload.max_size, tenant_id=B]`
- **THEN** Fallback reads `tenant_id = 0` platform default
- **AND** Returns `10MB`

### Requirement: Write Path Defaults to Current Tenant
Tenant admin SHALL only write own tenant's override records; platform admin SHALL explicitly call `WritePlatformDefault(...)` to write `tenant_id = 0` platform default.

#### Scenario: Tenant admin writes override
- **WHEN** Tenant A admin calls `PUT /tenant/config { key: upload.max_size, value: 50MB }`
- **THEN** Writes `sys_config(tenant_id=A, key=upload.max_size, value=50MB)` (insert or update)
- **AND** Does not affect `tenant_id = 0` platform default

### Requirement: Dict Type "Allow Tenant Override" Switch
`sys_dict_type` SHALL add `allow_tenant_override BOOL DEFAULT false` field; when `false` only platform admin can modify and all tenants share (read path does not fallback, directly reads `tenant_id = 0`).

#### Scenario: Non-overridable dict type
- **WHEN** Dict type `sys_user_status` (system enum) `allow_tenant_override = false`
- **AND** Tenant A admin attempts to modify its dict data
- **THEN** Returns `bizerr.CodeDictTypeNotOverridable`
- **AND** Modification rejected

### Requirement: Cache Invalidation by Tenant Override Dimension
Dict cache / config cache keys SHALL carry `tenant_id`; tenant write override triggers that `(tenant_id, dict_type|config_key)` dimension invalidation; platform write default triggers all tenants' corresponding dimension invalidation (because fallback affects all tenants).

#### Scenario: Tenant override triggers own tenant invalidation
- **WHEN** Tenant A modifies `sys_config[key=upload.max_size]`
- **THEN** Only `(tenant_id=A, key=upload.max_size)` cache invalidated
- **AND** Other tenants' cache unaffected

#### Scenario: Platform default change triggers all-tenant invalidation
- **WHEN** Platform admin modifies `sys_config[tenant_id=0, key=upload.max_size]`
- **THEN** All `(tenant_id=*, key=upload.max_size)` cache invalidated
- **AND** Propagated via cluster broadcast
