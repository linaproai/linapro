## ADDED Requirements

### Requirement: Tenant Admin Plugin Enable/Disable Interface
For `install_mode = tenant_scoped` `tenant_aware` plugins, tenant admin SHALL control own tenant's enable state through `POST /tenant/plugins/{plugin_id}/enable` and `POST /tenant/plugins/{plugin_id}/disable`; cannot affect other tenants.

### Requirement: Tenant View Visibility Constraints
`GET /tenant/plugins` SHALL only return: `scope_nature = tenant_aware` and `install_mode = tenant_scoped` plugins; installed status; not show `platform_only` plugins.

### Requirement: tenant_scoped Plugin Route Hit Strategy
`tenant_scoped` plugin HTTP routes SHALL be globally mounted at startup; middleware decides by `(tenant_id, plugin_id)` enable state whether to enter handler.

### Requirement: Enable State Cache and Invalidation
`(plugin_id, tenant_id, state_key='__tenant_enabled__') -> enabled` mapping SHALL be cached in `pluginruntimecache`, key carries tenant dimension; enable/disable change triggers that `(plugin_id, tenant_id)` dimension invalidation, cluster mode broadcast.
