## Requirements

### Requirement: Unified Plugin Directory and Manifest Contract

The system SHALL provide a unified directory structure and manifest contract for all plugins. Source plugins MUST reside under `apps/lina-plugins/<plugin-id>/`; dynamic WASM plugins MUST be discoverable from `plugin.dynamic.storagePath` and parseable to equivalent manifest information.

#### Scenario: Discover source plugin directories
- **WHEN** the host scans `apps/lina-plugins/` for plugin directories
- **THEN** only directories containing a valid manifest file are recognized as plugins
- **AND** each plugin's `plugin-id` is unique within the host scope
- **AND** the manifest only requires basic information and first-level plugin type

#### Scenario: Manifest remains minimal with menu declaration
- **WHEN** the host parses `plugin.yaml`
- **THEN** the manifest only requires `id`, `name`, `version`, `type` as mandatory fields
- **AND** `schemaVersion`, `compatibility`, `entry` are not required
- **AND** plugins declaring menus or button permissions must use `menus` metadata
- **AND** frontend pages, slots, and SQL locations follow directory and code conventions

#### Scenario: First-level type retains only source and dynamic
- **WHEN** the host parses `plugin.yaml` `type`
- **THEN** `type` only allows `source` or `dynamic`
- **AND** `wasm` is a runtime artifact semantic under `dynamic`, not a first-level type

#### Scenario: Install dynamic plugin artifacts
- **WHEN** an administrator uploads a `wasm` file to install a dynamic plugin
- **THEN** the host parses plugin ID, name, version, and type from embedded manifest
- **AND** rejects installation if basic fields are missing
- **AND** writes the artifact to `plugin.dynamic.storagePath/<plugin-id>.wasm`

#### Scenario: Dynamic plugin uses embedded resource declaration for manifest and SQL snapshots
- **WHEN** a dynamic plugin author uses `go:embed` for `plugin.yaml`, `manifest/sql`, and `manifest/sql/uninstall`
- **THEN** the builder reads resources from the embedded filesystem
- **AND** the runtime artifact's manifest and SQL snapshots remain the source of truth for host governance
- **AND** the host does not switch to guest runtime methods for these resources

#### Scenario: Dynamic plugin artifacts use independent storage directory
- **WHEN** the host discovers, uploads, or syncs a dynamic WASM plugin artifact
- **THEN** the artifact uses `plugin.dynamic.storagePath/<plugin-id>.wasm` as the canonical path
- **AND** the host does not rely on `apps/lina-plugins/<plugin-id>/plugin.yaml` for runtime discovery
- **AND** the readable source directory continues to maintain `backend/`, `frontend/`, and `manifest/` structure

#### Scenario: Active release reloads from stable archive
- **WHEN** a dynamic plugin has an active release and the host needs to reload its manifest
- **THEN** the host reloads from the stable archive path (e.g., `plugin.dynamic.storagePath/releases/<plugin-id>/<version>/<plugin-id>.wasm`)
- **AND** staging directory updates do not immediately replace the active release
- **AND** the reloaded manifest includes embedded hooks, resource contracts, and menu metadata

### Requirement: Plugin Manifest Must Validate Plugin ID Safety Boundary

Plugin manifest lifecycle SHALL uniformly validate plugin ID basic safety boundary when loading source plugin manifests, dynamic plugin artifact manifests, and plugin dependency declarations. The validation rules MUST reuse the non-empty, kebab-case, and 64-character length limits defined by the plugin ID governance capability. Runtime validation SHALL NOT enforce `<author>-<domain>-<capability>` recommended structure, domain whitelists, official capability reservations, or old official ID rejection tables. Any plugin ID basic safety validation failure MUST reject the manifest or artifact and return a diagnosable error.

#### Scenario: Source plugin manifest ID validation failure
- **WHEN** a source plugin `plugin.yaml` declares `id: Monitor_Server`
- **THEN** the host rejects loading the source plugin manifest
- **AND** the error states the plugin ID must use kebab-case lowercase letters and digits

#### Scenario: Source plugin registration ID inconsistent with manifest ID
- **WHEN** a source plugin registration ID is `linapro-monitor-server`
- **AND** `plugin.yaml` declares `id: linapro-monitor-loginlog`
- **THEN** the host rejects loading the source plugin
- **AND** the error states the source plugin registration ID is inconsistent with the manifest ID

#### Scenario: Dynamic plugin artifact manifest ID validation failure
- **WHEN** an administrator uploads a dynamic plugin artifact
- **AND** the artifact embedded manifest declares `id: plugin_demo_dynamic`
- **THEN** the host rejects the artifact
- **AND** the error states the dynamic plugin ID does not conform to plugin ID basic safety rules

#### Scenario: Plugin dependency declaration uses non-three-segment ID
- **WHEN** a plugin manifest's `dependencies.plugins[].id` declares `plugin-demo-source`
- **THEN** the host accepts the dependency declaration
- **AND** the host must not reject the manifest because it does not satisfy the recommended structure

### Requirement: Plugin Manifest Resource Namespace Must Use Current Plugin ID

Plugin manifest lifecycle SHALL require menus, permissions, runtime i18n, apidoc i18n, cron, and dynamic frontend asset entries declared in the manifest to use the current plugin ID-derived namespace. The host must not automatically derive or backfill these resources from old official IDs.

#### Scenario: Menu key uses old plugin ID
- **WHEN** plugin `linapro-content-notice` manifest declares menu key `plugin:content-notice:notice`
- **THEN** the host rejects the manifest
- **AND** the error states the menu key must use the `plugin:linapro-content-notice:` prefix

#### Scenario: Dynamic plugin menu path uses old asset path
- **WHEN** dynamic plugin `linapro-demo-dynamic` manifest declares menu path `/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js`
- **THEN** the host rejects or governance verification blocks the resource
- **AND** the error states the dynamic asset path must use `/plugin-assets/linapro-demo-dynamic/`

### Requirement: Plugin Lifecycle State Machine

The system SHALL provide an auditable plugin lifecycle state machine with distinct semantics for source and dynamic plugins, and allow `plugin.autoEnable` to advance plugins during startup.

#### Scenario: Source plugin compiled and integrated
- **WHEN** the host compiles a source tree containing source plugins
- **THEN** the plugin enters lifecycle scope as a discovered governable plugin
- **AND** administrators or `plugin.autoEnable` can advance it to installed and enabled

#### Scenario: Source plugin stays discovered-only after first sync
- **WHEN** the host discovers a source plugin for the first time
- **THEN** the plugin remains in discovered-only state by default
- **AND** routine sync does not auto-upgrade to installed or enabled

#### Scenario: Auto-enable installs and enables source plugins during startup
- **WHEN** `plugin.autoEnable` matches a discovered source plugin
- **THEN** the host installs then enables the plugin during startup
- **AND** routes, menus, cron, and hooks only become effective after enablement succeeds

#### Scenario: Install dynamic plugins
- **WHEN** an administrator installs a valid WASM dynamic plugin or `plugin.autoEnable` requires it
- **THEN** the host creates installation records, processes migrations, registers resources, and prepares loading
- **AND** normal users do not see the plugin until explicitly enabled

#### Scenario: Disable plugin
- **WHEN** an administrator disables an enabled plugin
- **THEN** the host stops exposing hooks, slots, pages, and menus
- **AND** preserves business data, role authorizations, and installation record

#### Scenario: Uninstall dynamic plugins
- **WHEN** an administrator uninstalls a dynamic plugin
- **THEN** the host removes menus, resource references, runtime artifacts, and mount info
- **AND** does not delete plugin business data by default

#### Scenario: Upgrade plugin
- **WHEN** an administrator upgrades a plugin to a new release
- **THEN** the host creates a new release record with generation info
- **AND** the old release remains rollback-capable until the new one is stable

#### Scenario: Failed release remains isolated
- **WHEN** a dynamic plugin upgrade fails and triggers rollback
- **THEN** the host marks the failed release as `failed`
- **AND** restores the registry to the stable release
- **AND** failed release assets do not continue serving publicly

#### Scenario: Source plugins do not expose install/uninstall actions
- **WHEN** an administrator views source plugin management actions
- **THEN** the host does not show install or uninstall for source plugins
- **AND** only exposes sync, enable, and disable

### Requirement: Plugin Manifest Lifecycle Must Recognize Dependency Declaration

Plugin manifest discovery, explicit sync, release snapshot, and read-only governance query SHALL preserve and expose plugin `dependencies` declaration. Source plugin and dynamic plugin dependency declarations must use the same structured semantics during manifest validation, sync to release snapshot, and plugin list projection.

#### Scenario: Sync source plugin dependency declaration
- **WHEN** a source plugin `plugin.yaml` contains `dependencies`
- **AND** an administrator performs explicit plugin sync
- **THEN** the system validates the dependency declaration
- **AND** the system preserves the dependency declaration in the plugin release snapshot
- **AND** plugin list or detail query can return dependency summary

#### Scenario: Sync dynamic plugin dependency declaration
- **WHEN** a dynamic plugin artifact manifest contains `dependencies`
- **AND** the system parses the dynamic plugin artifact
- **THEN** the system uses the same dependency validation rules as source plugins
- **AND** the dynamic plugin release snapshot preserves the dependency declaration

### Requirement: Explicit Plugin Install Lifecycle Must Execute Dependency Orchestration

Explicit plugin install requests SHALL call dependency resolution and install orchestration before target plugin install side effects. The system must first execute framework version and plugin dependency checks, then install dependency plugins according to the auto-dependency install plan, and finally install the target plugin.

#### Scenario: Explicit install processes dependencies first
- **WHEN** an administrator requests installing plugin `x`
- **AND** `x` has auto-install hard dependency `a`
- **THEN** the system installs `a` before executing `x`'s install SQL or dynamic runtime coordination
- **AND** `a` installation succeeds before continuing to install `x`

#### Scenario: Dependency check failure means install has no side effects
- **WHEN** plugin `x`'s dependency check fails
- **THEN** the system must not execute `x`'s install SQL
- **AND** the system must not sync `x`'s menus, permissions, resource references, or install state

### Requirement: Plugin Install Interface Must Return Dependency Plan and Results

Plugin management install interface SHALL support callers obtaining dependency check results, auto-install plans, and auto-install results. On install failure, the error response must express dependency blocker reasons as structured business errors, avoiding reliance on free text alone.

#### Scenario: Install success returns auto-dependency results
- **WHEN** installing a target plugin auto-installs dependency plugins
- **THEN** the install response contains the target plugin ID
- **AND** the install response or subsequent detail query contains the list of successfully auto-installed dependency plugins

#### Scenario: Dependency blocker returns structured error
- **WHEN** installing a target plugin fails due to dependency version mismatch
- **THEN** the HTTP response contains a stable business error code
- **AND** the response contains the target plugin ID, dependency plugin ID, current version, and required version range

### Requirement: Plugin Uninstall Lifecycle Must Check Reverse Dependencies

Plugin uninstall requests SHALL check installed plugins' hard dependencies before executing uninstall side effects. If downstream installed plugins depend on the target, the uninstall lifecycle must block.

#### Scenario: Reverse dependency found before uninstall
- **WHEN** an administrator requests uninstalling plugin `base`
- **AND** installed plugin `consumer` hard-depends on `base`
- **THEN** the system refuses to uninstall `base`
- **AND** the system returns `consumer` as the downstream dependency

#### Scenario: Reverse dependency reads from release snapshot
- **WHEN** the currently installed source plugin's workspace manifest is unreadable
- **THEN** the system preferentially uses the installed release snapshot's dependency declaration for reverse dependency checking
- **AND** when dependency safety cannot be confirmed, the system adopts a conservative blocking strategy

### Requirement: Plugin Menu Governance via Manifest Metadata

The system SHALL use `menus` metadata in `plugin.yaml` or embedded manifest for plugin menu and button permission management.

#### Scenario: Source plugin syncs menus
- **WHEN** the host syncs a source plugin manifest
- **THEN** it idempotently writes menus based on `menus` metadata
- **AND** resolves `parent_id` via `parent_key`
- **AND** grants default admin role authorization

#### Scenario: Install dynamic plugin registers menus
- **WHEN** an administrator installs a dynamic plugin
- **THEN** after install SQL, the host writes menus from manifest `menus` metadata
- **AND** install SQL handles business tables and seed data, not menu registration

#### Scenario: Uninstall dynamic plugin deletes menus
- **WHEN** an administrator uninstalls a dynamic plugin
- **THEN** after uninstall SQL, the host deletes menus by `menu_key` from manifest
- **AND** cleanup is scoped to declared menu keys only

### Requirement: Plugin Resource Ownership and Migration Tracking

The system SHALL record plugin ownership of host resources and migration execution for audit, rollback, and recovery.

#### Scenario: Plugin registers host resources
- **WHEN** a plugin creates menus, permissions, configs, dicts, files, or other resources during install
- **THEN** the host records the resource-to-plugin-to-release ownership

#### Scenario: Execute plugin migrations
- **WHEN** a plugin install or upgrade requires SQL or other migration steps
- **THEN** the host records execution order, version, checksum, result, and timestamp
- **AND** the same migration item for the same release is not re-executed

#### Scenario: Plugin SQL naming and directory constraints
- **WHEN** a plugin provides install SQL under `manifest/sql/`
- **THEN** files use `{序号}-{迭代名称}.sql` naming
- **AND** install SQL is in `manifest/sql/` root
- **AND** uninstall SQL is in `manifest/sql/uninstall/`
- **AND** mock-data SQL is in `manifest/sql/mock-data/`

#### Scenario: Plugin menu governance uses stable identifiers
- **WHEN** the host syncs menus from manifest metadata
- **THEN** `menu_key` is the stable menu identifier
- **AND** parent relationships use `parent_key` to resolve `parent_id`
- **AND** governance does not depend on fixed integer `id`

#### Scenario: Partial install failure
- **WHEN** a plugin fails during migration, resource registration, or artifact preparation
- **THEN** the host marks the plugin as failed or pending manual intervention
- **AND** rolls back uncommitted governance resources
- **AND** preserves failure context for diagnosis

### Requirement: Plugin Install/Enable Shortcut

The system SHALL allow administrators to trigger enablement directly from the installation review flow while preserving the existing `install -> enable` lifecycle order.

#### Scenario: Choose install and enable from the dialog
- **WHEN** an administrator chooses "Install and Enable" in the installation review dialog
- **THEN** the host runs install first, then enable
- **AND** when both succeed, the plugin ends in installed and enabled state

#### Scenario: Dynamic plugin composite action reuses authorization snapshot
- **WHEN** a dynamic plugin completes authorization confirmation and the administrator continues with "Install and Enable"
- **THEN** the authorization snapshot persists during install
- **AND** enable reuses that snapshot without a second confirmation dialog

#### Scenario: Enablement failure does not roll back install
- **WHEN** install succeeds but enable fails in the composite action
- **THEN** the plugin stays in `installed but disabled` state
- **AND** the administrator can retry enablement later

### Requirement: Plugin Mock Data Installation

The manual plugin install request SHALL expose `installMockData` for optional mock-data loading.

#### Scenario: User opts in and mock data installs
- **WHEN** the user checks mock-data checkbox and install SQL succeeds
- **THEN** the host executes mock SQL files in order
- **AND** marks plugin installed after all mock SQL succeeds

#### Scenario: Mock SQL failure rolls back mock data only
- **WHEN** install SQL succeeds but a mock SQL file fails
- **THEN** the host rolls back mock data and ledger rows
- **AND** the plugin remains installed without mock data

#### Scenario: Source and dynamic plugins share mock mechanism
- **WHEN** source and dynamic plugins use `manifest/sql/mock-data/`
- **THEN** same scanning, transactional execution, error format, and frontend behavior apply

### Requirement: Plugin List Query is Side-Effect Free

The system SHALL treat plugin list queries as read-only. Synchronization is triggered only by explicit sync actions.

#### Scenario: Query plugin list
- **WHEN** an administrator calls `GET /api/v1/plugins`
- **THEN** the system returns the plugin list without writing governance tables

#### Scenario: Explicit sync
- **WHEN** an administrator triggers `POST /api/v1/plugins/sync`
- **THEN** the system scans and may synchronize governance data
