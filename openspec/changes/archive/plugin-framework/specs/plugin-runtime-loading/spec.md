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

### Requirement: Dynamic Plugin Lifecycle Contract Auto-Discovery at Build Time

The system SHALL auto-discover dynamic plugin lifecycle handler contracts at build time from guest controller methods. The `build-wasm` tool MUST identify methods matching source-plugin lifecycle naming conventions (`BeforeInstall`, `AfterInstall`, `BeforeUpgrade`, `AfterUpgrade`, `BeforeDisable`, `AfterDisable`, `BeforeUninstall`, `AfterUninstall`, `BeforeTenantDisable`, `AfterTenantDisable`, `BeforeTenantDelete`, `AfterTenantDelete`, `BeforeInstallModeChange`, `AfterInstallModeChange`) that satisfy guest dispatcher bridge handler signatures, and generate `LifecycleContract` entries written to the WASM lifecycle custom section. The host runtime MUST continue reading only artifact-embedded lifecycle contracts and must not probe guest methods at runtime.

#### Scenario: Build-time discovers lifecycle methods
- **WHEN** a dynamic plugin controller exposes a valid bridge handler method `BeforeInstall`
- **AND** the plugin does not provide a `backend/lifecycle` override declaration
- **THEN** `build-wasm` auto-generates a lifecycle contract with `operation=BeforeInstall`
- **AND** the generated contract is written to the dynamic plugin WASM artifact lifecycle custom section

#### Scenario: Host runtime does not probe lifecycle methods
- **WHEN** the host loads a dynamic plugin artifact
- **THEN** the host only reads lifecycle contracts from the artifact
- **AND** the host must not probe `Before*` or `After*` paths to determine whether the dynamic plugin implements lifecycle handlers

#### Scenario: Unimplemented lifecycle method does not generate contract
- **WHEN** a dynamic plugin controller does not expose a `BeforeUninstall` method
- **THEN** `build-wasm` does not generate a lifecycle contract for `operation=BeforeUninstall`
- **AND** the host must not call the dynamic plugin's `BeforeUninstall` handler when executing the corresponding lifecycle scenario

### Requirement: Dynamic Plugin Lifecycle Declaration Must Be Optional Override for Auto-Discovery

The system SHALL treat `backend/lifecycle/*.yaml` as optional overrides for lifecycle auto-discovery results. An override MAY customize `requestType`, `internalPath`, or `timeoutMs` for a discovered operation, but MUST NOT create new contracts for lifecycle handlers that do not exist in the plugin. The build tool MUST fail for duplicate operations, invalid operations, invalid timeouts, and overrides that cannot match auto-discovered handlers.

#### Scenario: Override customizes lifecycle timeout
- **WHEN** a dynamic plugin controller exposes a valid `BeforeInstall` lifecycle method
- **AND** `backend/lifecycle/001-before-install.yaml` declares `operation=BeforeInstall` and `timeoutMs=3000`
- **THEN** `build-wasm` generates a `BeforeInstall` lifecycle contract
- **AND** the contract timeout uses the override-declared `3000` milliseconds

#### Scenario: Override declares non-existent method
- **WHEN** `backend/lifecycle/001-before-install.yaml` declares `operation=BeforeInstall`
- **AND** the dynamic plugin controller does not expose a valid `BeforeInstall` handler
- **THEN** `build-wasm` build fails
- **AND** the error points to the lifecycle override not finding a corresponding handler

#### Scenario: Override duplicates operation
- **WHEN** two YAML files under `backend/lifecycle` declare `operation=BeforeInstall`
- **THEN** `build-wasm` build fails
- **AND** the error points to the duplicate lifecycle operation

### Requirement: Official Dynamic Sample Plugin Must Declare Lifecycle via Auto-Discovery

The official dynamic sample plugin SHALL rely on controller method auto-discovery to generate lifecycle contracts and no longer require maintaining duplicate `backend/lifecycle/*.yaml` files. The sample plugin's artifact after packaging MUST still contain lifecycle contracts matching source-plugin naming conventions, covering install, upgrade, disable, uninstall, tenant disable, tenant delete, and install mode change pre/post handlers.

#### Scenario: Sample plugin generates complete contracts without handwritten lifecycle YAML
- **WHEN** building `plugin-demo-dynamic`
- **AND** the sample plugin does not maintain `backend/lifecycle/*.yaml`
- **THEN** the build artifact contains `BeforeInstall`, `AfterInstall`, `BeforeUpgrade`, `AfterUpgrade`, `BeforeDisable`, `AfterDisable`, `BeforeUninstall`, `AfterUninstall`, `BeforeTenantDisable`, `AfterTenantDisable`, `BeforeTenantDelete`, `AfterTenantDelete`, `BeforeInstallModeChange`, and `AfterInstallModeChange` lifecycle contracts
- **AND** the host runtime can call these handlers following existing lifecycle flow after parsing the artifact

### Requirement: Lifecycle Manifest Snapshot Must Use Shared Typed Bridge Contract

The system SHALL use a typed manifest snapshot DTO from `pluginbridge/contract` as the sole manifest snapshot publishing contract for dynamic plugin lifecycle requests and source plugin upgrade callbacks. Dynamic plugin `LifecycleRequest.fromManifest` and `LifecycleRequest.toManifest` MUST use the typed DTO and must not construct via hand-written `map[string]interface{}` field names. The source plugin side manifest snapshot wrapper MUST reuse the same DTO to avoid maintaining two sets of field names for source and dynamic plugins.

#### Scenario: Dynamic lifecycle request publishes typed manifest snapshot
- **WHEN** the host builds a lifecycle request for a dynamic plugin `BeforeUpgrade`, `Upgrade`, or `AfterUpgrade`
- **THEN** `fromManifest` and `toManifest` use the shared typed manifest snapshot DTO serialization
- **AND** manifest snapshot fields are defined by the DTO's JSON tags
- **AND** the runtime code building the request must not hand-write manifest snapshot map keys

#### Scenario: Source plugin and dynamic plugin reuse the same manifest snapshot contract
- **WHEN** the host builds a `ManifestSnapshot` for a source plugin upgrade callback
- **THEN** the source plugin wrapper reuses the same typed manifest snapshot DTO as the dynamic plugin lifecycle request
- **AND** adding, removing, or renaming manifest snapshot publishing fields must expose all unsynchronized call sites through compile-time field references

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
