## MODIFIED Requirements

### Requirement: Plugin lifecycle state machine must be governable

The system SHALL provide an auditable lifecycle state machine for plugins, distinguishing lifecycle semantics between source plugins and dynamic plugins.

#### Scenario: Source plugin compiles with host

- **WHEN** the host compiles the source tree containing source plugins and generates the LinaPro binary
- **THEN** the source plugin's backend Go code is compiled together with the host source
- **AND** the source plugin remains in a "discoverable but default not installed" governance form in the plugin registry
- **AND** administrators need to explicitly install the source plugin before deciding whether to enable it

#### Scenario: Source plugin defaults to not installed after first sync

- **WHEN** the host first discovers a source plugin and writes it to the plugin registry
- **THEN** the source plugin defaults to "not installed and not enabled" state
- **AND** the host sync only updates manifest, release snapshot, and basic governance metadata

### Requirement: Plugin menus must be governed through manifest metadata

The system SHALL use `menus` metadata in `plugin.yaml` or dynamic artifact embedded manifests to manage plugin menus and button permissions, rather than requiring plugins to directly operate `sys_menu` and `sys_role_menu` through SQL.

#### Scenario: Source plugin syncs menus during installation
- **WHEN** an administrator installs a source plugin
- **THEN** the host idempotently writes or updates corresponding `sys_menu` based on its `menus` metadata after executing install SQL
- **AND** the host resolves real `parent_id` from `parent_key`
- **AND** the host adds default admin role authorization for these menus

#### Scenario: Dynamic plugin installs and registers menus
- **WHEN** an administrator installs a dynamic plugin
- **THEN** the host continues to idempotently write or update `sys_menu` based on manifest `menus` metadata after executing install SQL
- **AND** plugin install SQL handles business tables and seed data, but no longer handles menu registration

### Requirement: Plugin installation must show unified review dialog

The system SHALL show a single install review dialog before executing installation for both source plugins and dynamic plugins, allowing administrators to review plugin details before deciding to proceed.

#### Scenario: Install source plugin shows details first

- **WHEN** an administrator clicks "Install" for an uninstalled source plugin
- **THEN** the host first shows a source plugin install detail dialog
- **AND** the dialog shows at least plugin name, plugin ID, plugin type, version, and description
- **AND** the host begins the install flow only after administrator confirmation

#### Scenario: Install dynamic plugin uses single review dialog

- **WHEN** an administrator clicks "Install" for an uninstalled dynamic plugin
- **THEN** the host directly shows the same install review dialog with plugin details and host service authorization scope
- **AND** no longer shows a generic install confirmation first, then a second authorization confirmation

### Requirement: Resource-type host service authorization must be confirmed at install time

The system SHALL display all resource-type host service permission requests during dynamic plugin installation, and persist the result as an authorization snapshot. For releases that have already formed authorization snapshots, subsequent enables MUST reuse the snapshot directly.

#### Scenario: Install shows host service permission requests

- **WHEN** the host prepares to install a dynamic plugin declaring resource-type hostServices
- **THEN** the host shows the plugin's requested services, methods, and resources in the install review dialog
- **AND** authorization items are displayed in order: data service, storage service, network service, runtime service
- **AND** the review dialog is read-only, showing the complete service manifest

#### Scenario: Confirmed authorization snapshot reused on enable

- **WHEN** a dynamic plugin release has already formed an authorization snapshot at install time, and the administrator later enables the plugin
- **THEN** the host enables the plugin directly using that snapshot
- **AND** no longer shows an authorization confirmation dialog

### Requirement: Plugins must deliver i18n resources with versions and participate in lifecycle management

Plugins SHALL deliver locale resources through a standard `manifest/i18n/<locale>/` resource directory. The host SHALL manage those resources during plugin discovery, installation, upgrade, enablement, disablement, and uninstallation.

#### Scenario: Dynamic plugin uninstall removes i18n resources
- **WHEN** an administrator uninstalls an installed dynamic plugin
- **THEN** the host removes the plugin's i18n resources from runtime translation aggregation
- **AND** the plugin's menus and metadata no longer expose its localized messages

### Requirement: Plugin manifest and lifecycle must support zero-code extension when adding new languages

The plugin manifest and lifecycle SHALL automatically cover new language resources when the host adds a new built-in language, without modifying host code or individual plugin source code.

#### Scenario: Plugin resources auto-integrate after enabling Traditional Chinese
- **WHEN** the host enables `zh-TW` as a built-in language
- **AND** source plugin `manifest/i18n/zh-TW/*.json` exists
- **THEN** when the plugin is enabled, its `zh-TW` translation resources automatically join runtime translation aggregation

### Requirement: Plugin resource interfaces must use plugin resource permissions

When a logged-in user accesses plugin-owned resource interfaces like `/plugins/{pluginId}/resources/{resource}`, the host SHALL verify access by the plugin's resource permissions or default derived plugin resource permissions. The host MUST NOT additionally require plugin management governance permissions like `plugin:query`.

#### Scenario: Plugin resource interface checked by plugin permissions
- **WHEN** a user requests a plugin-owned resource interface
- **THEN** the host checks access by the plugin's declared menu/button permissions
- **AND** if the user has the corresponding permissions, the request continues
- **AND** the host does not additionally require `plugin:query` governance permissions
