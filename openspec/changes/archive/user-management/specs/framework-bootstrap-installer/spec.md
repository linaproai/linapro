## ADDED Requirements

### Requirement: Startup Assembly tenantcap.Provider
Framework startup SHALL detect if `multi-tenant` enabled during plugin enable phase; if enabled, wait for plugin to register `tenantcap.Provider`, then activate tenancy middleware; if not enabled, assemble no-op Service.

### Requirement: Startup Consistency Validation
Startup SHALL validate:
1. `sys_plugin.scope_nature` and `install_mode` combination legal.
2. `sys_role` does not contain platform role boolean field; tenant roles cannot configure `data_scope=1`.
3. `sys_user.tenant_id=0` must have no active membership.
4. `multi-tenant` enabled must have Provider registered.

Validation failure SHALL prevent startup and print clear error.
