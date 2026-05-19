## MODIFIED Requirements

### Requirement: Host must execute startup bootstrap before plugin wiring

The system SHALL advance the lifecycle state of plugins listed in `plugin.autoEnable` before plugin HTTP route registration, plugin Cron wiring, and dynamic frontend bundle prewarm. Within one host startup orchestration, the startup bootstrap and subsequent plugin wiring, enabled snapshot refresh, and dynamic frontend prewarm MUST reuse the same set of plugin governance startup snapshots, avoiding repeated reads of equivalent plugin registry, release snapshot, menu, and resource reference full-table data.

#### Scenario: Source plugins reach enabled state before startup wiring

- **WHEN** a discovered source plugin appears in `plugin.autoEnable`
- **THEN** the host installs and enables the source plugin before route and Cron registration
- **AND** subsequent enabled snapshot reads treat the plugin as enabled
- **AND** subsequent plugin route registration and dynamic frontend prewarm reuse the plugin governance snapshot created or updated during startup bootstrap

#### Scenario: Plugins not in auto-enable list remain under manual governance

- **WHEN** a plugin is discovered but not in `plugin.autoEnable`
- **THEN** the host only performs normal manifest synchronization and registry refresh
- **AND** the host MUST NOT automatically install or enable it due to startup bootstrap
- **AND** subsequent startup wiring phases MUST NOT construct equivalent governance snapshots for that plugin
