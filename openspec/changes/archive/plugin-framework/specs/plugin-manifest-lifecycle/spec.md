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
