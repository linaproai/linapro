## ADDED Requirements

### Requirement: Startup (plugin, tenant) State Cache Assembly
Framework startup SHALL one-time read all `state_key='__tenant_enabled__'` `(plugin_id, tenant_id, enabled)` rows from `sys_plugin_state`, assemble into `pluginruntimecache`; subsequent runtime only invalidation on demand.

### Requirement: Startup Consistency Validation
Startup SHALL execute:
1. `sys_plugin.scope_nature` and `install_mode` consistency (platform_only <-> global).
2. `sys_role` must not have platform role boolean field; `tenant_id>0` roles cannot configure `data_scope=1`.
3. `sys_user.tenant_id=0` must have no active membership.
4. `multi-tenant` plugin state and `tenantcap.Provider` registration consistency.

Any consistency check failure SHALL prevent startup and print clear error.

### Requirement: New Tenant Creation Triggers Platform Plugin Enable Strategy
After startup, `multi-tenant` plugin SHALL in tenant creation flow explicitly call plugin governance service for subsequent new tenant creation to auto-initialize platform plugin system's `auto_enable_for_new_tenants=true` plugins' `sys_plugin_state` rows.
