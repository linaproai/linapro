## ADDED Requirements

### Requirement: plugin.yaml Required scope_nature Field
All source plugins and dynamic plugins' `plugin.yaml` SHALL contain `scope_nature` field, value only `platform_only` or `tenant_aware`; missing or illegal value causes plugin install failure.

### Requirement: scope_nature Semantics
`scope_nature` values SHALL be interpreted as: `platform_only` means plugin only runs at platform level affecting all tenants, tenant admin MUST NOT see or control its enablement, `install_mode` forced to `global`; `tenant_aware` means plugin supports platform-level or tenant-level operation, `install_mode` selected by platform admin at install time.

### Requirement: scope_nature Immutable
Plugin once installed, its `scope_nature` SHALL not be modifiable at runtime; only through plugin version upgrade with different manifest scope_nature is change allowed, requiring migration script.

### Requirement: Startup Consistency Validation
Framework startup SHALL validate `sys_plugin.scope_nature` and `sys_plugin.install_mode` combination is legal: `platform_only` must be `install_mode = global`, otherwise panic startup failure.
