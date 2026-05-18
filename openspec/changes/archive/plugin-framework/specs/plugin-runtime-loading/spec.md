## Requirements

### Requirement: Source Plugin Resource Discovery via Directory Convention

The system SHALL discover source plugin resources by directory convention and load backends through a centralized explicit registry.

#### Scenario: Scan source plugin directory resources
- **WHEN** the host runs backend or frontend build
- **THEN** it scans all valid source plugins under `apps/lina-plugins/`
- **AND** discovers manifest, SQL, frontend pages, and slot resources by convention

#### Scenario: Source plugin Go backend compiles via explicit registry
- **WHEN** a source plugin provides backend Go code in its directory
- **THEN** the developer adds a blank import in `apps/lina-plugins/lina-plugins.go`
- **AND** the plugin's Go package compiles into the same binary as the host

### Requirement: Dynamic WASM Plugin Validation and Loading

The system SHALL support installing dynamic WASM plugin artifacts with integrity and compatibility validation.

#### Scenario: Install single-file WASM plugin
- **WHEN** an administrator uploads a single `.wasm` file
- **THEN** the host reads embedded metadata and optional resources
- **AND** backend-only plugins need no extra frontend resources
- **AND** plugins with frontend resources require correct extraction for enablement

#### Scenario: Builder prioritizes embedded resource declaration
- **WHEN** the builder generates a dynamic plugin WASM artifact
- **THEN** it reads manifest, frontend, and SQL from the plugin's embedded filesystem
- **AND** converts them to host-recognized custom section snapshots
- **AND** the host continues consuming snapshots, not guest resource reads

### Requirement: Plugin Enable/Disable/Upgrade Without Host Restart

The system SHALL support enabling, disabling, and upgrading dynamic plugins without restarting the host process.

#### Scenario: Hot-enable plugin
- **WHEN** an administrator enables an installed dynamic plugin
- **THEN** the host loads the release and updates the registry in-process
- **AND** new requests immediately access the plugin's pages, hooks, and resources

#### Scenario: Hot-upgrade plugin
- **WHEN** an administrator upgrades a dynamic plugin to a new release
- **THEN** new requests switch to the new release
- **AND** in-flight old requests complete naturally
- **AND** users on the plugin page receive a refresh prompt

#### Scenario: Staged upload does not immediately replace active release
- **WHEN** an administrator uploads a higher-version WASM
- **THEN** the artifact is written to staging
- **AND** the active release continues serving
- **AND** the new release only becomes active after Reconciler generation switch

#### Scenario: Upgrade failure serves stable release
- **WHEN** a dynamic plugin upgrade fails
- **THEN** the host rolls back to the stable release
- **AND** the failed release's assets do not continue serving

### Requirement: Multi-Node Generation-Based Convergence

The system SHALL propagate plugin changes via generation sync in multi-node deployments.

#### Scenario: Primary node executes upgrade
- **WHEN** a multi-node environment triggers plugin install/enable/disable/upgrade
- **THEN** only the primary node executes shared migrations and release switches
- **AND** other nodes converge local state from the latest generation

#### Scenario: Nodes report convergence state
- **WHEN** the primary switches a plugin's active release
- **THEN** each node updates its `sys_plugin_node_state` based on `generation/release_id`
- **AND** nodes that fail to load mark their projection as failed with diagnostics
