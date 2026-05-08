## Requirements

### Requirement: Plugin Pages Support Multiple Host Integration Modes

The system SHALL support iframe, new-tab, and embedded-mount plugin page integration modes.

#### Scenario: iframe mode
- **WHEN** a plugin menu declares iframe mode
- **THEN** the host loads the plugin page in an iframe within the unified layout
- **AND** uses Lina's menu, permission, and navigation system

#### Scenario: New-tab mode
- **WHEN** a plugin menu declares new-tab mode
- **THEN** clicking the menu opens the plugin page in a new tab
- **AND** the host backend navigation state is unaffected

#### Scenario: Embedded-mount mode
- **WHEN** a plugin page declares embedded-mount mode
- **THEN** the host loads the plugin's frontend mount entry in a container
- **AND** the plugin can use any frontend framework internally
- **AND** the entry must satisfy the host's mount contract

### Requirement: Host Manages Plugin Frontend Resources

The system SHALL host plugin frontend resources and provide stable access paths.

#### Scenario: Source plugins participate in host build
- **WHEN** a source plugin provides frontend resources
- **THEN** resources participate in host frontend build
- **AND** pages are real source files, not JSON declarations
- **AND** resources are placed in `frontend/pages/` by convention
- **AND** final artifacts embed with the host binary

#### Scenario: Dynamic plugins provide frontend resources
- **WHEN** a runtime WASM plugin with resources is installed
- **THEN** the host extracts and hosts static resources
- **AND** generates stable static resource access paths
- **AND** unprepared plugin pages cannot be enabled

#### Scenario: Dynamic plugin frontend resources can come from embedded declaration
- **WHEN** a dynamic plugin author uses `go:embed` for `frontend` resources
- **THEN** the builder extracts from the embedded filesystem into the runtime snapshot
- **AND** the host continues hosting from the snapshot at `/plugin-assets/<plugin-id>/<version>/...`

### Requirement: Plugin Pages Enter Host Routing with Generation Awareness

The system SHALL include enabled plugin pages in the dynamic route system and protect user experience during generation changes.

#### Scenario: Plugin page generation refresh prompt
- **WHEN** a user is on a plugin page when that plugin hot-upgrades
- **THEN** the host shows "Plugin updated, please refresh current page"
- **AND** non-plugin-page users are not affected

#### Scenario: Refresh switches to new generation without forced navigation
- **WHEN** a user on a plugin page clicks refresh after generation change
- **THEN** the host rebuilds menus and dynamic routes
- **AND** does not force-navigate to another workspace page
- **AND** the plugin page remounts to the new generation

### Requirement: Plugin Management Page Shows Key Governance Information

The system SHALL prioritize key governance fields in the plugin management page.

#### Scenario: Source plugin type display
- **WHEN** plugin type is `source`
- **THEN** the management page shows "Source Plugin"

#### Scenario: Dynamic plugin type display
- **WHEN** plugin type is `dynamic`
- **THEN** the management page shows "Dynamic Plugin"
