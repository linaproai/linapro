## ADDED Requirements

### Requirement: Plugins must deliver i18n resources with versions and participate in lifecycle management
The system SHALL allow plugins to deliver locale resources through a standard plugin resource directory, and the host SHALL manage those resources together with plugin discovery, installation, upgrade, enablement, disablement, and uninstallation.

#### Scenario: Source plugin sync registers i18n resources
- **WHEN** the host discovers a source plugin with a standard i18n resource directory
- **THEN** the host registers the plugin's available locale resources
- **AND** those resources can participate in localized projection for menus, plugin name, and plugin description

#### Scenario: Dynamic plugin uninstall removes i18n resources
- **WHEN** an administrator uninstalls an installed dynamic plugin
- **THEN** the host removes the plugin's i18n resources from runtime translation aggregation
- **AND** the plugin's menus and metadata no longer expose its localized messages

### Requirement: Plugin metadata and plugin menus must support current-language localization
The system SHALL localize plugin name, plugin description, and plugin-declared menu titles according to the current request language while keeping plugin ID, menu key, route path, and permission semantics unchanged.

#### Scenario: Plugin list returns localized plugin names
- **WHEN** an administrator views the plugin management list or plugin detail with `en-US`
- **THEN** plugin name and plugin description use English localized results
- **AND** plugin ID, version, status, and governance fields keep their original semantics

#### Scenario: Plugin menus return titles by language
- **WHEN** an enabled plugin's declared menus have translation resources for the current language
- **THEN** plugin menu titles in left navigation, menu management, and role authorization trees use that language result
- **AND** the plugin does not need to directly change multilingual field structures in `sys_menu`
