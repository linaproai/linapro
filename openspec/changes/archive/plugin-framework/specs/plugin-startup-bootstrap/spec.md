## Requirements

### Requirement: Simplified Plugin Auto-Enable Config in Main Config File

The system SHALL provide `plugin.autoEnable` as a list of structured objects with required `id` and optional `withMockData`. Bare string entries are also accepted. The config declares which plugins must auto-enable during startup.

#### Scenario: Parse valid auto-enable list
- **WHEN** the host reads `plugin.autoEnable` from the main config
- **THEN** it builds a valid set of plugin IDs for auto-enable

#### Scenario: Reject invalid config
- **WHEN** `plugin.autoEnable` is invalid or contains empty IDs
- **THEN** the host refuses to continue startup

### Requirement: Startup Bootstrap Runs Before Plugin Wiring

The system SHALL advance auto-enable plugins before route registration, cron wiring, and bundle warmup.

#### Scenario: Source plugin reaches enabled before wiring
- **WHEN** a source plugin is in `plugin.autoEnable`
- **THEN** the host installs and enables it before route/cron registration

#### Scenario: Non-listed plugins remain manual
- **WHEN** a plugin is not in `plugin.autoEnable`
- **THEN** the host only performs routine sync, not auto-install/enable

### Requirement: Auto-Enable Includes Install Semantics

Plugins in `plugin.autoEnable` are interpreted as "ensure enabled during startup." If not installed, the host installs first, then enables.

### Requirement: Auto-Enable Failure Blocks Startup

If any listed plugin is missing, fails to install, fails to enable, or does not converge within the wait window, the host MUST fail fast.

### Requirement: Cluster Mode Separates Shared and Local Actions

In cluster mode, only the primary node executes shared lifecycle actions. Followers wait for convergence and refresh local projections.

### Requirement: Dynamic Plugin Auto-Enable Reuses Authorization Snapshots

When a dynamic plugin with governed host services appears in `plugin.autoEnable`, the host reuses the existing authorization snapshot. Missing snapshots block startup.

### Requirement: Management UI Labels Auto-Enabled Plugins

The plugin management UI SHALL show read-only indicators for auto-enabled plugins and warn before disable/uninstall that the host will restore on restart unless config changes.
