## Requirements

### Requirement: Official Plugin Workspace Must Be Optional

The system SHALL treat `apps/lina-plugins` as the default location for official source plugin workspace, but host basic development, build, and test entry points must not require the directory to exist or contain plugin code. The directory can be absent, an empty directory, or mounted as a `git submodule` containing official plugin code.

#### Scenario: Host commands executable when plugin workspace absent
- **WHEN** `apps/lina-plugins` does not exist in the workspace
- **THEN** host backend compilation and host backend unit tests remain executable
- **AND** failures must not come from Go workspace loading missing plugin modules or host compile-time importing official plugin aggregate package

#### Scenario: Host commands executable when plugin workspace empty
- **WHEN** an empty directory `apps/lina-plugins` exists in the workspace
- **THEN** host backend compilation and host backend unit tests remain executable
- **AND** source plugin discovery results in an empty set
- **AND** host startup must not fail due to absence of source plugin manifests

#### Scenario: Official plugins restored via submodule
- **WHEN** `apps/lina-plugins` submodule initialization is complete
- **THEN** the host can discover official source plugin manifests
- **AND** official source plugin backend, frontend, manifest, SQL, and E2E continue using the `apps/lina-plugins/<plugin-id>/` directory convention

### Requirement: Host and Official Source Plugin Compile-Time Dependency Must Be Decoupled

The host SHALL be able to complete default compilation without official source plugin Go modules. Official source plugin backend registration must be enabled by an explicit plugin-full build path, build tag, generated aggregate file, or equivalent mechanism; the host default entry must not unconditionally depend on the `lina-plugins` module.

#### Scenario: Default host build does not import official plugin aggregate module
- **WHEN** a developer builds `apps/lina-core` in a workspace where `apps/lina-plugins` is not initialized
- **THEN** the build does not resolve the `lina-plugins` module
- **AND** the build does not require any `lina-plugin-*` module to exist

#### Scenario: Full plugin build enables official source plugin registration
- **WHEN** a developer explicitly runs a full build or test entry including official plugins
- **THEN** the build path enables official source plugin backend registration
- **AND** if `apps/lina-plugins` submodule is not initialized, the command fails fast with a prompt to initialize the submodule

### Requirement: Plugin Workspace State Must Be Diagnosable

The developer tool SHALL distinguish four workspace states: missing, empty directory, initialized submodule, and path exists but structure invalid, and output actionable diagnostics in commands that need plugin content.

#### Scenario: Submodule not initialized when plugin content needed
- **WHEN** a developer runs official plugin unit tests, plugin E2E, dynamic plugin wasm build, or full plugin build
- **AND** `apps/lina-plugins` does not exist or is empty
- **THEN** the command fails and explains the current lack of official plugin workspace
- **AND** the error prompt includes `git submodule update --init --recursive`

#### Scenario: Host command encounters missing plugin workspace
- **WHEN** a developer runs a host-only host command
- **AND** `apps/lina-plugins` does not exist or is empty
- **THEN** the command continues executing
- **AND** diagnostics at most output at info level that the current mode is host-only

### Requirement: Host-only and Plugin-full Verification Must Be Separated

The system SHALL provide repeatable host-only verification and plugin-full verification. Host-only verification must cover host backend, host frontend, and host E2E in both absent and empty plugin workspace states; plugin-full verification must cover official plugin backend unit tests and plugin E2E after submodule initialization.

#### Scenario: Host-only verification covers missing plugin directory state
- **WHEN** `apps/lina-plugins` is temporarily removed from the workspace
- **THEN** host-only verification runs host backend unit tests
- **AND** host-only verification runs host frontend type check or build
- **AND** host-only verification runs host E2E tests not dependent on official source plugin content

#### Scenario: Host-only verification covers empty plugin directory state
- **WHEN** `apps/lina-plugins` is an empty directory
- **THEN** host-only verification runs the same set of host verifications
- **AND** plugin management or plugin discovery related host tests assert correct behavior when source plugin set is empty

#### Scenario: Plugin-full verification covers submodule state
- **WHEN** `apps/lina-plugins` submodule initialization is complete
- **THEN** plugin-full verification runs all official plugin Go unit tests
- **AND** plugin-full verification runs all official plugin-owned E2E tests
- **AND** plugin-full verification runs dynamic plugin wasm build or equivalent plugin artifact verification

### Requirement: Official Plugin Submodule Migration Must Preserve Default Path

The official plugin repository SHALL be mounted as a single submodule to `apps/lina-plugins` unless a subsequent OpenSpec change explicitly approves a path migration. Main repository documentation, CI, and developer commands must describe initialization and verification flows around this default mount path.

#### Scenario: Clone main repository then initialize official plugins
- **WHEN** a developer clones the main repository and needs full official plugin capabilities
- **THEN** documentation provides the submodule initialization command
- **AND** after initialization, official plugin directories appear under `apps/lina-plugins`
- **AND** existing plugin directory structure specifications continue to apply

#### Scenario: Main repository only runs host capabilities
- **WHEN** a developer clones the main repository but does not initialize the submodule
- **THEN** documentation explains that host-only development and test entry points can be run
- **AND** documentation explains that plugin-specific commands require submodule initialization
