## ADDED Requirements

### Requirement: Host-embedded plugin pages must participate in host locale context and message refresh
The system SHALL let host-managed plugin pages participate in the host locale context and runtime message refresh flow. When the active language changes, host-embedded plugin pages and their route metadata MUST refresh to the new language without requiring plugin reinstallation.

#### Scenario: Plugin route titles refresh after language switch
- **WHEN** a user switches workspace language in an authenticated session
- **THEN** the host rebuilds plugin route titles and menu titles accessible to the current user
- **AND** enabled plugin pages switch their navigation and tab display copy to the new language

#### Scenario: Host-embedded plugin pages load runtime message bundles
- **WHEN** the host loads a plugin frontend page in embedded mode
- **THEN** the plugin page can access the host current-locale context and matching runtime translation resources
- **AND** the plugin does not need to reimplement language resolution rules detached from the host

### Requirement: Plugin language resource changes must be observed by the host promptly
The system SHALL refresh plugin-related runtime messages after plugin enablement, disablement, or upgrade so that the host UI does not continue showing stale plugin translations.

#### Scenario: Enabled plugin immediately displays plugin translations
- **WHEN** an administrator enables a plugin that has i18n resources in the current session
- **THEN** the host refreshes runtime message bundles and dynamic menus
- **AND** the plugin's navigation title and page copy can immediately display in the current language

#### Scenario: Plugin upgrade switches to new-version translation resources
- **WHEN** a plugin is upgraded to a release containing new translation resources and the switch becomes effective
- **THEN** subsequent runtime message bundles provided by the host use the new release resources
- **AND** old release translation messages are no longer consumed by the frontend
