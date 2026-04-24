## MODIFIED Requirements

### Requirement: The plugin life cycle state machine can be managed

The system SHALL provide an auditable plugin lifecycle state machine, distinguish the lifecycle semantics of source plugins and dynamic plugins, and allow the host to advance plugins to the enabled state during startup through `plugin.autoEnable` in the host main config file.

#### Scenario: Source code plugin is compiled and integrated with the host
- **WHEN** the host compiles a source tree that contains source plugins and builds the LinaPro binary
- **THEN** the backend Go code of that source plugin is compiled together with the host source
- **AND** the source plugin enters the lifecycle scope as a discovered governable plugin instead of being treated as installed or enabled automatically just because it was compiled in
- **AND** administrators or `plugin.autoEnable` can still explicitly advance it later to installed and enabled states

#### Scenario: Source plugin stays discovered-only after the first synchronization
- **WHEN** the host discovers a source plugin for the first time and writes it into the plugin registry
- **THEN** that source plugin remains in a discovered-only, not-installed, not-enabled state by default
- **AND** later routine synchronization does not automatically upgrade it to installed or enabled

#### Scenario: Auto-enable list installs and enables a source plugin during host startup
- **WHEN** `plugin.autoEnable` in the host main config file matches a discovered source plugin
- **THEN** the host installs that source plugin during startup and then advances it to enabled
- **AND** the source plugin's routes, menus, cron jobs, and hooks only become externally effective after enablement succeeds

#### Scenario: Install dynamic plugins
- **WHEN** an administrator installs a valid `wasm` dynamic plugin, or `plugin.autoEnable` in the host main config requires a discovered dynamic plugin to be enabled during startup
- **THEN** the host creates plugin installation records and current-version records
- **AND** the host processes migrations, resource registration, permission access, and frontend/backend loading preparation in order
- **AND** normal users do not see the plugin capability until it has been explicitly advanced to enabled

#### Scenario: Auto-enable list can request enabled state for dynamic plugins
- **WHEN** `plugin.autoEnable` in the host main config file matches a discovered dynamic plugin
- **THEN** the host advances that dynamic plugin through the shared lifecycle flow required to reach the enabled target state
- **AND** the plugin only becomes visible to normal users after installation, authorization, and reconcile all succeed

#### Scenario: Disable plugin
- **WHEN** an administrator switches an enabled plugin to disabled
- **THEN** the host stops exposing that plugin's hooks, slots, pages, and menus
- **AND** the host keeps the plugin's business data, role authorization relationships, and installation record
- **AND** the governance relationship can be restored when the plugin is enabled again

#### Scenario: Uninstall dynamic plugins
- **WHEN** an administrator uninstalls a dynamic plugin
- **THEN** the host removes the host-side menus, resource references, runtime artifacts, and mount information registered by that plugin
- **AND** the host does not delete the plugin's business tables or business data by default
- **AND** the host keeps uninstall audit information
