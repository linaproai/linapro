# Plugin Startup Bootstrap

## Purpose

Define how plugins listed in `plugin.autoEnable` are installed, optionally seeded with mock data, enabled, and converged during host startup.

## Requirements

### Requirement: The host must provide structured plugin auto-enable entries in the main config file

The host SHALL provide `plugin.autoEnable` in `apps/lina-core/manifest/config/config.yaml` as a list of structured entries. Each entry MUST be an object with required `id` and optional `withMockData` fields. `withMockData` defaults to `false`; only entries with `withMockData=true` load plugin mock data during first-time startup installation. Bare string entries MUST be rejected.

#### Scenario: Parse a valid structured auto-enable list
- **WHEN** `plugin.autoEnable` contains `{id: "demo-control", withMockData: false}`, `{id: "plugin-demo-source", withMockData: true}`, and `{id: "plugin-demo-dynamic"}`
- **THEN** the host parses those entries as `[(demo-control, false), (plugin-demo-source, true), (plugin-demo-dynamic, false)]`
- **AND** startup loads mock data only for `plugin-demo-source`

#### Scenario: Reject invalid auto-enable config
- **WHEN** config contains `{id: ""}`, `{withMockData: true}`, `{id: "x", withMockData: "yes"}`, or a bare string entry
- **THEN** the host fails during config loading or startup
- **AND** the error identifies the invalid entry position or key

### Requirement: The host must execute startup bootstrap before plugin wiring

The system SHALL advance lifecycle state for plugins listed in `plugin.autoEnable` before plugin HTTP route registration, plugin cron wiring, and dynamic frontend bundle warm-up.

#### Scenario: A source plugin reaches enabled state before startup wiring
- **WHEN** a discovered source plugin appears in `plugin.autoEnable`
- **THEN** the host installs and enables that source plugin before route and cron registration
- **AND** later enabled-snapshot reads see that plugin as enabled

#### Scenario: Plugins not in the auto-enable list remain under manual governance
- **WHEN** a plugin is discovered but not present in `plugin.autoEnable`
- **THEN** the host only performs routine manifest sync and registry refresh
- **AND** the host MUST NOT auto-install or auto-enable it because of startup bootstrap

### Requirement: The auto-enable list must implicitly include install and enable semantics

For each `plugin.autoEnable` entry, `BootstrapAutoEnable(ctx)` SHALL execute implicit install-if-needed plus enable semantics. If `withMockData=true`, first-time installation MUST reuse the manual install path's transactional mock SQL execution. If `withMockData=false`, startup MUST NOT scan or execute the mock-data directory. Already installed plugins MUST NOT reload mock data even when their entry has `withMockData=true`; the option applies only to first-time installation.

#### Scenario: Auto-enable a newly discovered plugin without mock data
- **WHEN** `plugin.autoEnable` contains `{id: "plugin-demo-source"}` and the plugin is not installed
- **AND** the host executes `BootstrapAutoEnable`
- **THEN** the host executes install SQL, registration, menu synchronization, and enablement
- **AND** it does not scan `manifest/sql/mock-data/`
- **AND** no mock data rows from that plugin are created

#### Scenario: Auto-enable a newly discovered plugin with mock data opt-in
- **WHEN** `plugin.autoEnable` contains `{id: "plugin-demo-source", withMockData: true}` and the plugin is not installed
- **AND** the host executes `BootstrapAutoEnable`
- **THEN** the host transactionally executes all plugin `manifest/sql/mock-data/*.sql` files after install SQL succeeds
- **AND** the plugin is enabled after the mock phase succeeds

#### Scenario: An installed plugin reappears with mock-data opt-in
- **WHEN** `plugin.autoEnable` contains `{id: "x", withMockData: true}` and the plugin is already installed
- **AND** the host executes `BootstrapAutoEnable`
- **THEN** the host only ensures the plugin is enabled
- **AND** it does not re-run install SQL or mock-data SQL

### Requirement: Any failure for a listed auto-enable plugin must block host startup

Any `BootstrapAutoEnable` stage failure MUST block host startup. Mock phase failures for entries with `withMockData=true` SHALL also block startup; after the mock transaction rolls back, the host MUST surface an error that includes the plugin ID, failed SQL file, and rollback fact so operations can fix the issue and restart.

#### Scenario: A missing auto-enable plugin causes startup failure
- **WHEN** a plugin ID listed in `plugin.autoEnable` does not exist in the catalog
- **THEN** startup fails
- **AND** the error includes the plugin ID

#### Scenario: Install failure causes startup failure
- **WHEN** install SQL fails for an auto-enabled plugin
- **THEN** startup fails
- **AND** the error includes the failure cause

#### Scenario: Mock SQL failure during auto-enable causes startup failure
- **WHEN** `plugin.autoEnable` contains `{id: "x", withMockData: true}`
- **AND** install SQL succeeds
- **AND** any SQL file under `manifest/sql/mock-data/` fails
- **THEN** the host rolls back the mock transaction and fails startup
- **AND** the error includes the plugin ID, failed mock SQL file, and failure cause

### Requirement: Startup bootstrap must separate shared lifecycle side effects from local convergence in cluster mode

The system SHALL allow only the primary node to execute shared plugin lifecycle actions in cluster mode, such as install SQL, menu writes, release switches, and shared-state advancement. Follower nodes only wait for shared-state results and refresh their local projections.

#### Scenario: The primary node executes shared plugin actions
- **WHEN** a plugin appears in `plugin.autoEnable` in cluster mode and installation or enablement must advance
- **THEN** only the primary node executes shared install, enable, or reconcile actions
- **AND** follower nodes MUST NOT repeat those shared side effects

#### Scenario: Follower nodes refresh local views after shared convergence
- **WHEN** a follower starts in cluster mode and finds a plugin in `plugin.autoEnable`
- **THEN** the follower waits for the primary node to write shared stable state or for the wait window to time out
- **AND** it refreshes its local enabled snapshot and runtime projection from that shared result

### Requirement: Startup auto-enable of dynamic plugins must reuse existing authorization snapshots

The system SHALL reuse the host-approved authorization snapshot of the current release when a dynamic plugin that declares governed host services appears in `plugin.autoEnable`. The host MUST NOT require authorization details in the main config file.

#### Scenario: Reuse an existing authorization snapshot during dynamic-plugin auto-enable
- **WHEN** a dynamic plugin appears in `plugin.autoEnable` and its current release already has an approved authorization snapshot
- **THEN** the host reuses that snapshot to drive startup auto-enable
- **AND** the host does not require authorization details from the main config file again

#### Scenario: Reject startup auto-enable when no authorization snapshot exists
- **WHEN** a dynamic plugin appears in `plugin.autoEnable`, declares governed host services, and has no authorization snapshot
- **THEN** the host stops startup
- **AND** the error clearly says a normal reviewed flow is required first

### Requirement: The plugin-management UI must label startup auto-enabled plugins and warn about temporary governance actions

The system SHALL show whether a plugin is matched by `plugin.autoEnable` through read-only indicators in plugin-management list and detail views. When administrators disable or uninstall those plugins, the UI MUST warn that the action is immediate but the host will reinstall or re-enable the plugin after restart unless the config changes.

#### Scenario: The list and detail views show the startup auto-enable indicator
- **WHEN** a plugin ID exists in `plugin.autoEnable`
- **THEN** plugin-management list and detail views show a read-only auto-enable indicator

#### Scenario: Disabling a startup auto-enabled plugin warns about restart behavior
- **WHEN** an administrator attempts to disable an auto-enabled plugin
- **THEN** the UI shows a risk-confirmation prompt
- **AND** the prompt states that permanent disablement requires editing `plugin.autoEnable`

#### Scenario: Uninstalling a startup auto-enabled plugin warns about restart behavior
- **WHEN** an administrator attempts to uninstall an auto-enabled plugin
- **THEN** the uninstall confirmation shows a risk warning
- **AND** the warning states that startup will reinstall and enable the plugin if the config remains unchanged

