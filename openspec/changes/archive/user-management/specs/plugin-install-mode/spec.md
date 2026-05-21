## ADDED Requirements

### Requirement: install_mode Field Semantics
`sys_plugin.install_mode` SHALL take values `global` or `tenant_scoped`:
- `global`: Plugin effective for all tenants, enable/disable controlled by platform admin (`sys_plugin_state.tenant_id=0 AND state_key='__tenant_enabled__'` single row).
- `tenant_scoped`: Plugin enabled per tenant independently, enable/disable controlled by each tenant admin (`sys_plugin_state.tenant_id>0 AND state_key='__tenant_enabled__'` multiple rows).

### Requirement: Install-Time install_mode Selection
Platform admin installing `tenant_aware` plugin SHALL explicitly select `install_mode` through UI or API; `platform_only` plugin install_mode forced to `global`, no selection allowed.

### Requirement: New Tenant Auto-Enable Strategy
`tenant_aware` plugin must not declare new tenant default enable strategy in plugin.yaml; platform plugin system SHALL maintain this strategy via `sys_plugin.auto_enable_for_new_tenants`. When plugin declares `supports_multi_tenant=true`, installed, host layer enabled, `install_mode = tenant_scoped` and `auto_enable_for_new_tenants = true`, system SHALL auto-initialize `sys_plugin_state` on new tenant creation.

### Requirement: install_mode Switching Rules
- `global -> tenant_scoped`: Allowed, platform admin confirms, system auto-initializes all active tenants with `enabled=current global state`.
- `tenant_scoped -> global`: Platform admin only, MUST second confirmation + forced audit, immediately force-enable all tenants.
- `scope_nature` change: Runtime prohibited, only through plugin version upgrade.
