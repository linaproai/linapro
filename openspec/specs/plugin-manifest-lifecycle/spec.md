# plugin-manifest-lifecycle Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: Unify the plugin directory and manifest contract

The system SHALL provides a unified directory structure and manifest contract for all plugins. The source plugin MUST be placed in the `apps/lina-plugins/<plugin-id>/` directory; the current `dynamic` dynamic `wasm` plugin MUST be discovered from `plugin.dynamic.storagePath` and parse the manifest information equivalent to the source plugin.

#### Scenario: Discover the source plugin directory

- **WHEN** The host scans the plugin directory under `apps/lina-plugins/`
- **THEN** Only identify directories containing legitimate manifest files as plugins
- **AND** Each plugin's `plugin-id` is unique within the host scope
- **AND** The list only needs to contain plugin basic information and first-level plugin type

#### Scenario: `plugin.yaml` keeps the plugin menu simple and declarative

- **WHEN** Host parses `plugin.yaml`
- **THEN** The manifest only requires basic fields such as `id`, `name`, `version`, `type` etc.
- **AND** The host no longer requires extended metadata such as `schemaVersion`, `compatibility`, `entry`, etc.
- **AND** If the plugin needs to register menu or button permissions with the host, it MUST be declared in the manifest `menus` metadata
- **AND** The locations of frontend pages, `Slot` and SQL files are still deduced according to directory and code conventions instead of being configured repeatedly in the manifest.

#### Scenario: Only source and dynamic types are reserved for the list-level type.

- **WHEN** The host parses the `type` in `plugin.yaml`
- **THEN** `type` only allows `source` or `dynamic`
- **AND** Currently only `wasm` is used as the product semantics of dynamic plugins, and is no longer used as a first-level plugin type.
- **AND** For historical `wasm` first-level type values, the host will treat them as `dynamic` from the governance perspective.

#### Scenario: Install dynamic plugin products

- **WHEN** The administrator uploads a `wasm` file to install dynamic plugins
- **THEN** The host can parse the plugin ID, name, version and first-level plugin type that are consistent with the source mode.
- **AND** deny installation of dynamic plugins missing these basic fields
- **AND** The host writes the uploaded product to `plugin.dynamic.storagePath/<plugin-id>.wasm`

#### Scenario: Dynamic plugin generates manifests and SQL snapshots by embedding resource declarations

- **WHEN** Dynamic plugin authors use `go:embed` to declare `plugin.yaml`, `manifest/sql` and `manifest/sql/uninstall`
- **THEN** The builder MUST read these resources from the embedded file system
- **AND** Manifests and SQL snapshots embedded in runtime artifacts MUST continue to be the source of truth for host installation, upload, and lifecycle management
- **AND** The host MUST NOT dynamically obtain these governance resources via guest runtime methods instead

#### Scenario: Dynamic plugin products use independent storage directories

- **WHEN** The host discovers, uploads or synchronizes a `dynamic` `wasm` dynamic plugin product
- **THEN** The runtime product MUST use `plugin.dynamic.storagePath/<plugin-id>.wasm` as the host-side standard disk path
- **AND** The host MUST no longer rely on `apps/lina-plugins/<plugin-id>/plugin.yaml` as the runtime discovery entry
- **AND** The readable source directory of the runtime sample plugin SHOULD continues to be maintained under `backend/`, `frontend/` and `manifest/` like the source plugin

#### Scenario: The current effective release is reloaded from the stable archive.

- **WHEN** A dynamic plugin already has a currently effective release, and the host needs to load its active manifest again
- **THEN** The host reloads the release from a stable archive path such as `plugin.dynamic.storagePath/releases/<plugin-id>/<version>/<plugin-id>.wasm`
- **AND** The host will not immediately replace the active release in the current service just because an updated `wasm` file appears in the staging directory.
- **AND** active manifest still contains hooks, universal resource contracts and menu metadata declared inline in this release after reloading

### Requirement: The plugin life cycle state machine can be managed

The system SHALL provides an auditable life cycle state machine for plugins, and distinguishes life cycle semantics between source plugins and dynamic plugins.

#### Scenario: Source code plugin is compiled and integrated with the host

- **WHEN** The host compiles the source tree where the source plugin is located and generates the LinaPro binary
- **THEN** The backend Go code of the source plugin is compiled together with the host source.
- **AND** Source code plugins are considered integrated in the plugin registry and do not require additional installation steps.
- **AND** Administrators only need to manage the enabled and disabled status of source plugins

#### Scenario: The source plugin is enabled by default after the first synchronization.

- **WHEN** The host discovers a source plugin for the first time and writes it to the plugin registry
- **THEN** This source plugin is in the "Integrated and Enabled" state by default
- **AND** Subsequent synchronization of the host will not overwrite the administrator's explicit disabling of the source plugin.

#### Scenario: Install dynamic plugins

- **WHEN** The administrator installs a valid `wasm` dynamic plugin
- **THEN** The host creates plugin installation records and current version records
- **AND** The host handles migration, resource registration, permission access and frontend and backend loading preparation in sequence according to the list
- **AND** Plugins will not be visible to normal users until explicitly enabled

#### Scenario: Disable plugin

- **WHEN** Administrator switches enabled plugin to disabled state
- **THEN** The host stops the plugin’s Hook, Slot, page and menu exposure
- **AND** The host retains plugin business data, role authorization relationships and installation records
- **AND** The existing governance relationship can be restored after the plugin is re-enabled

#### Scenario: Uninstall dynamic plugins

- **WHEN** The administrator uninstalls a dynamic plugin
- **THEN** The host removes the menu, resource references, runtime products and mounting information registered by the plugin on the host side.
- **AND** The host does not delete the plugin's own business data table or business data by default
- **AND** Host retains uninstall audit information

### Requirement: Plug-in menu is managed through manifest metadata

The system SHALL uses `plugin.yaml` or `menus` metadata embedded in the manifest of dynamic products to manage plugin menu and button permissions, instead of requiring the plugin to directly operate `sys_menu` and `sys_role_menu` through SQL.

#### Scenario: Source plugin synchronization menu

- **WHEN** The host synchronizes a source plugin list
- **THEN** The host writes or updates the corresponding `sys_menu` idempotently based on the plugin `menus` metadata.
- **AND** The host parses the real `parent_id` based on `parent_key`
- **AND** The host completes the default administrator role authorization for these menus without requiring the plugin SQL to manually write `sys_role_menu`

#### Scenario: Install dynamic plugin registration menu

- **WHEN** Administrator installs a dynamic plugin
- **THEN** After executing the plugin installation SQL, the host continues to write or update the corresponding `sys_menu` based on the manifest `menus` metadata idempotently.
- **AND** After plugin installation, SQL can continue to be responsible for business tables and business seed data, but it will no longer be responsible for plugin menu registration.

#### Scenario: Uninstall dynamic plugin delete menu

- **WHEN** The administrator uninstalls a dynamic plugin
- **THEN** After the plugin uninstall SQL is successfully executed, the host deletes the corresponding `sys_role_menu` association and `sys_menu` based on the manifest `menus` metadata.
- **AND** The deletion scope is limited to the menu keys declared in the plugin manifest, and does not rely on the plugin SQL to manually maintain the deletion statement.
- **AND** If the plugin does not declare any menu, the host skips the menu deletion step

#### Scenario: Upgrade plugin

- **WHEN** The administrator installs a higher release version for the installed plugin
- **THEN** The host creates new release records and generation information for the plugin
- **AND** The old release remains reversible before the new release takes effect
- **AND** When the upgrade fails, the host can roll back to the last available release.

#### Scenario: Upgrade failed release remains isolated

- **WHEN** A dynamic plugin fails and triggers a rollback during an upgrade, migration, or frontend resource switch
- **THEN** The host will mark the failed release as `failed`
- **AND** The host restores the registry pointer to the last stable release
- **AND** The public frontend resources of failed release will no longer be provided externally through `/plugin-assets/<plugin-id>/<version>/...`

#### Scenario: The source plugin does not expose the installation and uninstallation operations

- **WHEN** The administrator views the plugin management operation items of the source plugin
- **THEN** The host will not display installation or uninstallation operations for source plugins.
- **AND** The source plugin only exposes applicable operations such as sync discovery, enablement and disabling

### Requirement: Plug-in resource ownership and migration records can be tracked

The system SHALL records the plugin's occupation of host resources and migration to support uninstallation, reinstallation, upgrade, auditing and fault recovery.

#### Scenario: Plug-in registration host resource

- **WHEN** Plugins create or declare menus, permissions, configurations, dictionaries, static resources or other host-managed resources during installation
- **THEN** The host records the ownership relationship between the resource, plugin and release
- **AND** These reference relationships can be queried, audited and used for uninstall cleanup

#### Scenario: Perform plugin migration

- **WHEN** Plug-in installation or upgrade requires SQL or other migration steps
- **THEN** The host records the execution order, version, verification summary, execution results and time of each migration item
- **AND** The same migration item of the same release will not be executed repeatedly

#### Scenario: Plug-in version SQL naming and directory constraints

- **WHEN** plugin provides installation phase SQL in the `manifest/sql/` directory
- **THEN** Install SQL files MUST use the naming format consistent with the host `{serial number}-{current iteration name}.sql`
- **AND** These installation SQL files MUST be placed in the `manifest/sql/` root directory of the plugin for the host to scan and execute in sequence
- **AND** Plug-in uninstallation SQL MUST be placed independently in the `manifest/sql/uninstall/` directory
- **AND** Host initialization sequence execution process MUST only scan the `manifest/sql/` root directory, and MUST not mistakenly execute the uninstall SQL under `manifest/sql/uninstall/`

#### Scenario: Plug-in menu management does not rely on integer menu ID

- **WHEN** The host synchronizes the host menu and button permissions based on the plugin manifest `menus` metadata.
- **THEN** Menu records MUST use `menu_key` as menu stable identifier
- **AND** The parent-child relationship MUST parse the real `parent_id` through `parent_key` instead of hard-coding the fixed integer `parent_id`
- **AND** The plugin installation, upgrade and uninstall process MUST not rely on the fixed integer `id`

#### Scenario: Partial failure of installation process

- **WHEN** The plugin failed at any step during migration, resource registration, or product preparation.
- **THEN** The host marks the plugin status as failed or pending manual intervention
- **AND** Host rollback of host management resources that have not yet taken effect
- **AND** The host retains the failure context for subsequent diagnostics

### Requirement: Dynamic plugin manifest can declare structured hosting service strategy

The system SHALL allows dynamic plugins to declare only the structured `hostServices` policy in `plugin.yaml`, which is used to describe the required host services, methods, resource application and governance parameters; the host's internal capability classification MUST be automatically derived based on these declarations, rather than requiring the author to repeatedly maintain the top-level `capabilities` field. The `storage` service currently declares logical path requests through `resources.paths`, and the `data` service currently declares data table requests through `resources.tables`.

#### Scenario: Plug-in declares host service strategy

- **WHEN** Developers write dynamic plugin lists
- **THEN** Manifest can declare `hostServices` metadata
- **AND** Each statement contains at least service, method collection and resource application or policy parameters
- **AND** Manifest no longer needs to declare top-level `capabilities` separately
- **AND** The builder directly reports errors for unknown services, unknown methods and illegal strategies.

#### Scenario: Host reads host service policy snapshot

- **WHEN** The host views the manifest snapshot or release snapshot of a dynamic plugin
- **THEN** The host can restore the host service policy declared by the plugin
- **AND** Administrators can use this to review the scope of host capabilities that plugins plan to access.

#### Scenario: Plug-in declares resource request instead of hosting underlying connection

- **WHEN** Developer declares host service dependencies in the manifest
- **THEN** For the `storage` service, the plugin only declares stable logical paths or path prefixes `resources.paths`
- **AND** For the `network` service, the plugin only declares a list of URL patterns
- **AND** For the `data` service, the plugin declares the list of table names `tables` that needs to be accessed under the `resources` node.
- **AND** For low-priority services such as `cache`, `lock` and `notify`, you can still continue to use logical `resourceRef` planning, which respectively represent the cache namespace, logical lock name and notification channel identifier.
- **AND** Plug-in manifest MUST not solidify database connection, host file absolute path, cache address or key plain text
- **AND** Real resource binding is completed by the host installation process or administrator configuration

### Requirement: Host service resource application is included in the plugin management resource index

The system SHALL integrates the host service resource applications declared by dynamic plugins into the `sys_plugin_resource_ref` governance resource index; this table is used to carry release-level plugin governance resource projections, rather than just mirroring an author-side field named `resourceRef`. Record logical path requests for `storage`, record URL pattern requests for `network`, record table name requests for `data`, and continue to record logical resource references for low-priority services such as `cache`, `lock`, and `notify`.

#### Scenario: Install or upgrade dynamic plugins to synchronize management resource indexes

- **WHEN** The host installs or upgrades a dynamic plugin that declares the host service resources
- **THEN** The host synchronizes these resource applications into plugin resource ownership records
- **AND** resource types can distinguish between `host-storage`, `host-upstream`, `host-data-table`, `host-cache`, `host-lock` and `host-notify-channel`
- **AND** These records can participate in audit, offload and rollback governance

#### Scenario: Uninstall or rollback dynamic plugin update management resource index

- **WHEN** The host uninstalls a dynamic plugin or rolls it back to an old release
- **THEN** The host synchronizes and updates the corresponding host service resource application record
- **AND** Logical paths, URL patterns, low-priority service logic `resourceRef` or data table declarations that are no longer used in the current release MUST not remain in the valid state

#### Scenario: Restore logical reference binding when release is activated

- **WHEN** The host activates a dynamic plugin release
- **THEN** The host restores the final authorization status of the resource application based on the release snapshot
- **AND** At runtime, host service calls will only be interpreted based on this snapshot.

### Requirement: Resource hosting service application requires host confirmation authorization when installing or enabling it.

The system SHALL displays all resource-based host service permission applications during the dynamic plugin installation phase, and the host administrator confirms the final authorization result of the current release; for a release that has already formed a confirmed authorization snapshot, the snapshot MUST be directly reused during subsequent activation, and the authorization window will not pop up repeatedly.

#### Scenario: Display host service permission application during installation

- **WHEN** The host is preparing to install a dynamic plugin that declares resource type hostServices
- **THEN** The host displays the service, method, resource identifier (such as `path`, URL pattern, `resourceRef` or `table`) applied by the plugin and its management parameter summary in the installation review window
- **AND** Authorization items are displayed in the order of "data service → storage service → network service → runtime service"; other services that are not hit are ranked after the known order
- **AND** The service-level method summary is displayed in the corresponding service title area, and the same set of methods is not repeatedly displayed under each data table or storage path entry.
- **AND** When the application item belongs to the `data` service and the host can parse the table-level description, the host will display the corresponding human-readable description directly after the table name, preventing the administrator from having to rely on the bare table name to determine the purpose.
- **AND** The inspection window no longer provides item-by-item check-cut interaction, but instead displays the complete service list declared by the current plugin in a read-only manner.
- **AND** Administrators can review the scope of host resources that the plugin plans to access based on this list
- **AND** For services that have declared subdivided resources, the host continues to use an unordered list to display the corresponding resource list item by item, but no longer provides a check control.
- **AND** For services that are managed only by service-level method summary and do not have subdivided resource entries, the host will only display the service title and method summary, and will no longer display additional prompt copy such as "No additional confirmation required"

#### Scenario: Persist the final authorized snapshot during installation

- **WHEN** Administrator confirms dynamic plugin installation review window
- **THEN** The host will persist the final confirmation result as an authorized snapshot of the current release
- **AND** By default, all host service resource applications currently declared by the plugin will be written into the authorization snapshot.
- **AND** The runtime subsequently interprets host service calls according to this complete authorization snapshot

#### Scenario: Directly reuse the snapshot when the authorized release is enabled

- **WHEN** A dynamic plugin release has taken a confirmed authorization snapshot during the installation phase and the administrator subsequently enables the plugin
- **THEN** The host directly activates the plugin according to the snapshot
- **AND** No more repeated authorization confirmation windows popping up
- **AND** The runtime continues to interpret host service calls only according to this final snapshot

#### Scenario: History needs to be confirmed, and confirmation will be made when release is enabled.

- **WHEN** A dynamic plugin release has not yet formed a confirmed authorization snapshot, but the administrator attempted to enable the plugin
- **THEN** The host still allows an additional authorization confirmation before activation.
- **AND** Enable the review window to continue to read-only display the complete service list declared by this release
- **AND** The host will write the confirmation result back to the authorization snapshot of the current release for direct reuse in subsequent activations.

### Requirement: Uniformly display the details confirmation window before plugin installation

The system SHALL displays a single installation review window before installing source plugins and dynamic plugins, allowing administrators to view plugin details before deciding whether to continue the installation.

#### Scenario: View details before installing the source plugin

- **WHEN** The administrator clicks the "Install" operation on the plugin management page for an uninstalled source plugin.
- **THEN** The host first pops up the source plugin installation details window instead of directly executing the installation.
- **AND** The window displays at least the plugin name, plugin ID, plugin type, version and description
- **AND** After the administrator confirms, the host will start the source plugin installation process.

#### Scenario: Use a single review window when installing dynamic plugins

- **WHEN** The administrator clicks the "Install" operation on the plugin management page that does not have a dynamic plugin installed.
- **THEN** The host directly pops up the same installation review window, and displays the plugin details and the host service authorization scope that need to be confirmed in the window.
- **AND** No longer pops up a general installation confirmation first, and then pops up a second authorization confirmation
- **AND** The host will start the dynamic plugin installation process only after the administrator confirms

### Requirement: `plugin.yaml` Remains Minimal and May Declare Menus
The system SHALL keep `plugin.yaml` focused on stable plugin metadata and SHALL not require source plugins to declare backend route inventories in the manifest.

#### Scenario: Source plugin backend routes are not duplicated in the manifest
- **WHEN** the host parses a source plugin `plugin.yaml`
- **THEN** the manifest does not need to list backend routes
- **AND** backend route registration code plus DTO `g.Meta` remains the only source of truth for source-plugin routes
- **AND** the host captures route ownership and documentation metadata during registration instead of reading a second route declaration model from `plugin.yaml`

### Requirement: Official source plugin usage field-capability plugin ID

The system SHALL uses domain-capability style `kebab-case` flags without the `plugin-` prefix for official source plugins to improve readability and avoid semantic duplication.

#### Scenario: Define the official source plugin identifier
- **WHEN** team named the official source plugin in the open source stage
- **THEN** Plugin ID uses `org-center`, `content-notice`, `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog`
- **AND** does not require the `plugin-` prefix

#### Scenario: Verify the validity of the plugin ID
- **WHEN** The host parses the `plugin.yaml` of the above official plugin
- **THEN** These plugin IDs only need to satisfy globally unique and `kebab-case` rules
- **AND** not illegal due to missing `plugin-` prefix

### Requirement: The source plugin menu MUST be mounted to the host stable directory

The system SHALL requires the official source plugin to point to the host stable directory key through `parent_key` in the manifest menu statement to ensure the long-term stability of the background navigation structure.

#### Scenario: Organization and content plugin declares parent directory
- **WHEN** `org-center` or `content-notice` declare menu metadata
- **THEN** Its top-level menu `parent_key` points to the host directory keys `org` and `content` respectively
- **AND** The internal submenu of the plugin can still continue to refer to the parent menu key declared by the same plugin.

#### Scenario: Monitoring plugin declares parent directory
- **WHEN** `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog` declare menu metadata
- **THEN** Its top menu `parent_key` points to the host directory key `monitor`
- **AND** The host presses the parent key to complete menu synchronization and start-stop linkage visibility management

#### Scenario: Official plugin uses fixed parent directory key mapping
- **WHEN** The host verifies the official source plugin manifest
- **THEN** The top-level `parent_key` of `org-center` MUST be `org`
- **AND** The top-level `parent_key` of `content-notice` MUST be `content`
- **AND** The top-level `parent_key` of `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog` MUST be `monitor`

#### Scenario: The official plugin declares an unsupported top-level mount key
- **WHEN** The above official source plugin uses `parent_key` that is inconsistent with the convention in its top-level menu declaration.
- **THEN** The host refuses to synchronize the plugin menu
- **AND** Provide administrators with diagnosable mount verification errors

### Requirement: The source plugin backend directory structure MUST converge to backend/internal

The system SHALL requires the source plugin to converge the backend business implementation under `backend/internal/`, avoid directly exposing the business service directory in the `backend/` root directory, and ensure that the private implementation boundary of the plugin is clear and consistent with the host agreement.

#### Scenario: Planning the standard directory of source plugins
- **WHEN** Team creates or refactors a source plugin
- **THEN** The plugin backend is organized by at least `backend/api/`, `backend/plugin.go`, `backend/internal/controller/`, `backend/internal/service/`
- **AND** Plugin frontend pages remain in `frontend/pages/`
- **AND** Plugin manifests and embedded resources are kept in `plugin.yaml`, `plugin_embed.go`, `manifest/sql/` and `manifest/sql/uninstall/`

#### Scenario: Place the plugin service component
- **WHEN** team adds or migrates business services for source plugins
- **THEN** All service components MUST be placed in `backend/internal/service/<component>/`
- **AND** MUST NOT CREATE `backend/service/<component>/`
- **AND** Non-`internal` directories such as `backend/provider/` are only used for stable capability provider / adapter and do not carry main business orchestration

#### Scenario: Plugin requires local ORM artifacts
- **WHEN** Source code plugin needs to access the database
- **THEN** `backend/hack/config.yaml` serves as the plugin’s local `gf gen dao` configuration entry
- **AND** The generated results fall into `backend/internal/dao/`, `backend/internal/model/do/` and `backend/internal/model/entity/`
- **AND** Access to the host shared table also continues to use the plugin's local generation of artifacts

### Requirement: Plugin installation review flow supports direct chained enablement

The system SHALL allow administrators to trigger enablement directly from the plugin installation review flow, while the host still follows the existing `install -> enable` lifecycle order instead of collapsing both actions into a new implicit state transition.

#### Scenario: Choose install and enable from the installation review dialog
- **WHEN** an administrator chooses `Install and Enable` in the installation review dialog for a plugin that is not installed
- **THEN** the host runs the plugin install lifecycle first
- **AND** after installation succeeds, the host continues with the enable lifecycle
- **AND** when both steps succeed, the plugin ends in the `installed and enabled` state

#### Scenario: Dynamic plugin composite action reuses the authorization snapshot captured during install
- **WHEN** a dynamic plugin completes host-service authorization confirmation in the installation review dialog and the administrator continues with `Install and Enable`
- **THEN** the host persists the authorization snapshot for the current release during install
- **AND** the subsequent enable step reuses that snapshot directly
- **AND** the composite action MUST NOT open a second authorization-confirmation dialog

#### Scenario: Enablement failure in the composite action does not roll back a successful install
- **WHEN** an administrator runs `Install and Enable`, the install lifecycle succeeds, and the enable lifecycle fails
- **THEN** the host keeps the plugin in the real `installed but disabled` state
- **AND** the host MUST NOT automatically undo the completed installation because enablement failed
- **AND** the administrator can still trigger enablement again later from the existing installed state

