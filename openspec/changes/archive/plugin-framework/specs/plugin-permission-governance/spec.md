## Requirements

### Requirement: Plugin Menu and Permission Reuse Lina Governance Modules

The system SHALL require plugin menus, button permissions, and role authorizations to reuse Lina's existing menu and role management systems.

#### Scenario: Plugin registers menus and permissions
- **WHEN** a plugin declares menus, button permissions, and page entries and completes installation
- **THEN** the host registers these into Lina's menu and role authorization system
- **AND** permission identifiers use plugin namespace prefix

#### Scenario: Plugin menus maintain authorization by menu_key
- **WHEN** a plugin installs, upgrades, or uninstalls a menu resource
- **THEN** the host locates `sys_menu` by `menu_key` and maintains `sys_role_menu` authorization

### Requirement: Plugin Role Authorization Persists Across Disable/Enable Cycles

The system SHALL preserve role authorization relationships when a plugin is disabled and restore them when re-enabled.

#### Scenario: Plugin disabled preserves authorization
- **WHEN** a plugin with role authorizations is disabled
- **THEN** the host stops the authorizations from taking effect
- **AND** does not delete the authorization relationships

#### Scenario: Plugin re-enabled restores authorization
- **WHEN** a previously disabled plugin is re-enabled
- **THEN** the host restores menu, button permission, and role authorization to active state

### Requirement: Plugin Runtime Can Access Host Permission Context

The system SHALL provide standardized host permission context for plugins to reuse Lina's user, role, department, and data permission scope.

#### Scenario: Plugin backend processes a request
- **WHEN** a plugin API, hook, or task needs the current user permission context
- **THEN** the host provides user ID, role IDs, menu permission codes, and data permission scope

### Requirement: Plugin Uninstall Only Cleans Governance Resources

The system SHALL remove host governance resources on plugin uninstall but preserve plugin business data by default.

#### Scenario: Uninstall an enabled dynamic plugin
- **WHEN** an administrator uninstalls an enabled dynamic plugin
- **THEN** the host removes menus, resource references, runtime artifacts, and mount info
- **AND** cleans up role-menu relationships
- **AND** preserves plugin business tables and data
