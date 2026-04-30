# Plugin Manifest Lifecycle

## Purpose

Define plugin manifest discovery, lifecycle resource synchronization, read-only governance queries, SQL asset classification, and language-resource extension behavior for source and dynamic plugins.

## Requirements

### Requirement: Plugin manifest and lifecycle must support zero-code extension when adding new languages

The plugin manifest and lifecycle SHALL automatically cover new language runtime UI translation resources and apidoc translation resources when the host adds a new built-in language, without modifying host code or individual plugin source code. Source plugins SHALL append that language's resources in their own `manifest/i18n/<locale>/*.json` and `manifest/i18n/<locale>/apidoc/**/*.json`; dynamic plugins SHALL write that language's resources into release custom sections during packaging; the host automatically discovers, loads, and cleans up these resources during loading, enable, disable, upgrade, and uninstall flows.

#### Scenario: Plugin resources auto-integrate after enabling a new language
- **WHEN** the host enables an additional built-in language
- **AND** a source plugin provides `manifest/i18n/<locale>/*.json`
- **THEN** enabling the plugin adds that language's resources to runtime translation aggregation
- **AND** disabling or uninstalling the plugin removes those resources from aggregation
- **AND** the flow requires no host code modifications and no unrelated plugin code modifications

#### Scenario: Dynamic plugins carry new language resources via release
- **WHEN** a dynamic plugin adds `manifest/i18n/<locale>/*.json` in a new version and repackages
- **AND** the host upgrades the plugin to that release
- **THEN** the new language resources take effect and old release resources are no longer used
- **AND** cache invalidation is scoped to the affected plugin sector

### Requirement: Plugin list query is side-effect free

The system SHALL treat plugin list queries as side-effect-free read operations. The list query may read discovered source manifests, dynamic plugin registry data, release snapshots, and governance projections, but MUST NOT create, update, or delete plugin governance table data. Plugin scanning and governance synchronization MUST be triggered only by explicit synchronization actions.

#### Scenario: Query plugin list from management page
- **WHEN** an administrator opens plugin management and calls `GET /api/v1/plugins`
- **THEN** the system returns the plugin list and current governance state
- **AND** the GET request does not write `sys_plugin`, `sys_plugin_release`, `sys_plugin_resource_ref`, `sys_menu`, or `sys_role_menu`

#### Scenario: Synchronize plugins explicitly
- **WHEN** an administrator triggers plugin synchronization with `POST /api/v1/plugins/sync`
- **THEN** the system scans source plugins and dynamic plugin artifacts
- **AND** the system may synchronize registry, release snapshot, resource index, menu, and permission governance data from manifests

### Requirement: Plugin host-service metadata lookup must avoid schema probing errors

The system SHALL read host database metadata for plugin list host-service projections through read-only metadata queries. This lookup MUST NOT trigger incorrect business-table schema probing for `information_schema.TABLES`; if the database does not support the metadata lookup or the lookup fails, the plugin list API SHALL degrade to returning raw table names.

#### Scenario: Resolve data table comments for dynamic plugin permissions
- **WHEN** a plugin list item declares `data.resources.tables`
- **THEN** the system attempts to read table comments for permission review display
- **AND** the lookup does not emit `SHOW FULL COLUMNS FROM TABLES` errors

#### Scenario: Metadata lookup unavailable
- **WHEN** the current database dialect does not support host table comment lookup or the lookup fails
- **THEN** the plugin list API still returns successfully
- **AND** hostServices permission display uses raw table names as fallback information

### Requirement: Plugin manifest SQL resources must classify mock-data as a separate asset type

When scanning a plugin `manifest/sql/` directory, the host SHALL distinguish install, uninstall, and mock-data SQL assets without overlap. `manifest/sql/*.sql` belongs to install assets, `manifest/sql/uninstall/*.sql` belongs to uninstall assets, and `manifest/sql/mock-data/*.sql` belongs to mock assets. Mock SQL files MUST NOT appear in install or uninstall asset lists. Source plugins and dynamic plugins MUST use the same scanning logic.

#### Scenario: Install asset list excludes mock-data files
- **WHEN** the host resolves install SQL assets for a plugin containing `manifest/sql/001-schema.sql` and `manifest/sql/mock-data/001-mock.sql`
- **THEN** the returned install asset list contains only `001-schema.sql`
- **AND** it does not contain `mock-data/001-mock.sql` or any variant of that path

#### Scenario: Mock asset scan returns mock-data files only
- **WHEN** the host resolves mock SQL assets for the same plugin
- **THEN** the returned asset list contains only files under `manifest/sql/mock-data/`
- **AND** the files are sorted by file name ascending

### Requirement: Dynamic plugin packaging must preserve the mock-data directory convention

Dynamic plugin packaging SHALL preserve `manifest/sql/mock-data/` in the artifact file-system view and use the same runtime scanning method as source plugins. Packaging tools and artifact schema MUST NOT introduce a different mock-data path or additional manifest fields for this purpose.

#### Scenario: Dynamic plugin upgrade preserves mock-data visibility
- **WHEN** a dynamic plugin adds or modifies `manifest/sql/mock-data/*.sql` in a new version
- **AND** the host upgrades to the new artifact
- **THEN** mock SQL scanning reflects the new version contents
- **AND** the mock-data directory remains visible through the artifact file-system view

