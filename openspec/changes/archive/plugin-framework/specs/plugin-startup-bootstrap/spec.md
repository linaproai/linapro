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

### Requirement: Startup Auto-Enable Must Synchronize Lifecycle Writes to the Startup Snapshot

The system SHALL maintain consistency between plugin lifecycle writes and the shared startup snapshot within a single host startup orchestration. When `plugin.autoEnable` performs an on-demand install for a source plugin, the subsequent enable check, status inspection, route wiring, and warmup phases within the same startup context MUST read the updated `installed`, `status`, `desiredState`, and `currentState` projections.

#### Scenario: Source plugin auto-installs then enables immediately

- **WHEN** the host startup context already carries a plugin governance startup snapshot
- **AND** `plugin.autoEnable` contains a source plugin that is not yet installed
- **THEN** the auto-install must synchronize the current startup snapshot's plugin registry projection
- **AND** the subsequent enable check must recognize the plugin as installed
- **AND** the host startup must not fail with `Plugin is not installed` for that plugin

#### Scenario: Already-installed source plugin auto-enables

- **WHEN** the host startup context already carries a plugin governance startup snapshot
- **AND** `plugin.autoEnable` contains a source plugin that is installed but not enabled
- **THEN** the enable phase must reuse the installed state from the current startup snapshot
- **AND** the enable must synchronize the current startup snapshot's enabled-state projection after completion

### Requirement: Startup Auto-Enable Must Resolve and Install Auto Dependencies

`BootstrapAutoEnable(ctx)` SHALL execute dependency checks for plugins listed in `plugin.autoEnable`. For dependencies that are discovered, version-satisfied, not yet installed, and declared as `required: true` with `install: auto`, the startup auto-enable must complete dependency installation in deterministic topological order before installing the target plugin.

#### Scenario: Install dependencies before auto-enabling target plugin
- **WHEN** `plugin.autoEnable` contains plugin `x`
- **AND** `x` declares auto-install hard dependency `a`
- **AND** `a` is not yet installed
- **THEN** startup bootstrap first installs `a`
- **AND** startup bootstrap then installs and enables `x`

#### Scenario: Startup dependency version not satisfied blocks startup
- **WHEN** `plugin.autoEnable` contains plugin `x`
- **AND** `x`'s hard dependency version is not satisfied
- **THEN** host startup fails
- **AND** the error contains the target plugin, dependency plugin, and version requirement

### Requirement: Startup Auto-Enable Must Not Implicitly Enable Dependency Plugins

The startup auto-enable flow SHALL only enable plugins explicitly listed in `plugin.autoEnable`. Plugins installed via dependency relationships must not be automatically enabled because they were installed as dependencies, unless the dependency plugin itself also appears in `plugin.autoEnable`.

#### Scenario: Dependency plugin not in auto-enable list
- **WHEN** plugin `a` is installed as an auto dependency of plugin `x`
- **AND** `a` is not in `plugin.autoEnable`
- **THEN** startup bootstrap only ensures `a` is installed
- **AND** startup bootstrap must not enable `a`

#### Scenario: Dependency plugin also in auto-enable list
- **WHEN** plugin `a` is installed as an auto dependency of plugin `x`
- **AND** `a` is also in `plugin.autoEnable`
- **THEN** startup bootstrap ensures `a` is enabled after dependency installation completes

### Requirement: Cluster Mode Startup Dependency Installation Must Respect Primary Node Side-Effect Boundary

In cluster mode, dependency installation triggered by startup auto-enable SHALL respect existing plugin lifecycle primary node boundaries. Shared installation, menu writes, release switches, and state advancement can only be executed by the primary node; follower nodes must wait for shared state and refresh local projections.

#### Scenario: Primary node installs auto dependencies
- **WHEN** in cluster mode the primary node executes `BootstrapAutoEnable`
- **AND** the auto-enable target plugin requires installing dependency plugins
- **THEN** the primary node executes dependency plugin installation side effects
- **AND** the primary node publishes runtime revision or equivalent events for affected plugins

#### Scenario: Follower node waits for dependency installation result
- **WHEN** in cluster mode a follower node executes `BootstrapAutoEnable`
- **AND** the auto-enable target plugin's dependency has not completed installation in shared state
- **THEN** the follower node waits for primary node convergence or wait window timeout
- **AND** the follower node must not duplicate dependency installation SQL or shared state writes
