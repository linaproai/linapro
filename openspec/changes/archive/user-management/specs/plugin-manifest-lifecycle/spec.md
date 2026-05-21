## ADDED Requirements

### Requirement: plugin.yaml Multi-Tenant Fields
plugin.yaml SHALL add: `scope_nature` (required), `supports_multi_tenant` (required boolean), `default_install_mode` (optional when scope_nature=tenant_aware).

### Requirement: Install-Time Consistency Validation
Install SHALL validate: `scope_nature=platform_only` only `install_mode=global`; `scope_nature=platform_only` `supports_multi_tenant` must be `false`.

### Requirement: sys_plugin Adds Governance Columns
`sys_plugin` SHALL add `scope_nature VARCHAR(32)`, `install_mode VARCHAR(32)` and platform strategy column `auto_enable_for_new_tenants BOOL`; `sys_plugin_state` SHALL keep `id` auto-increment primary key, add `tenant_id INT NOT NULL DEFAULT 0`, use `(plugin_id, tenant_id, state_key)` unique index for business uniqueness.

### Requirement: Uninstall Protected by LifecycleGuard Veto
Uninstall flow SHALL call all plugins' `CanUninstall` hooks before actual uninstall steps.

### Requirement: Dynamic Plugin Orphan Uninstall
When dynamic plugin installed but staging and active release artifacts unavailable, host SHALL allow platform admin through `force=true` restricted orphan uninstall.
