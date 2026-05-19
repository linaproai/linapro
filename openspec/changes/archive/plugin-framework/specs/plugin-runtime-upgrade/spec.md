## Requirements

### Requirement: Plugin Startup Scan Must Mark Runtime Upgrade State

The system SHALL scan source plugin and dynamic plugin file metadata during host startup and compare the database effective plugin version with the file-discovered target version. The startup scan must not automatically execute upgrades, must not switch the effective release, must not run upgrade SQL, and must not block host startup due to discovering an upgradable version.

#### Scenario: Discovered file version higher than database effective version
- **WHEN** at host startup the database records plugin `plugin-demo` effective version as `v0.1.0`
- **AND** the plugin file manifest version is `v0.2.0`
- **THEN** the system marks the plugin runtime state as `pending_upgrade`
- **AND** `sys_plugin.version` and effective `release_id` remain `v0.1.0`
- **AND** the host service continues starting

#### Scenario: Discovered file version lower than database effective version
- **WHEN** at host startup the database records plugin `plugin-demo` effective version as `v0.2.0`
- **AND** the plugin file manifest version is `v0.1.0`
- **THEN** the system marks the plugin runtime state as `abnormal`
- **AND** the system records the abnormal reason as file version lower than database effective version
- **AND** the host service continues starting and requires administrator manual fix

#### Scenario: Discovered version matches effective version
- **WHEN** at host startup the database records plugin `plugin-demo` effective version as `v0.2.0`
- **AND** the plugin file manifest version is `v0.2.0`
- **THEN** the system marks the plugin runtime state as `normal`
- **AND** the plugin continues participating in runtime loading as installed and enabled

### Requirement: Pending Upgrade Plugin Business Entry Must Be Controlled

The system SHALL protect plugin business entries when the plugin is in `pending_upgrade`, `abnormal`, or `upgrade_failed` state, to prevent target version code from running directly on old database or abnormal metadata. The host basic plugin management interface and upgrade interface MUST remain available so the administrator can complete the upgrade or fix.

#### Scenario: Pending upgrade source plugin has business routes
- **WHEN** plugin `plugin-demo` is in `pending_upgrade`
- **AND** a user accesses the plugin's declared business API or page route
- **THEN** the system blocks the plugin business entry from normal execution
- **AND** the system returns a stable upgrade-required status or hides the corresponding menu entry
- **AND** the plugin management page can still display and trigger the upgrade

#### Scenario: Abnormal plugin has cron tasks
- **WHEN** plugin `plugin-demo` is in `abnormal`
- **AND** the plugin declares cron tasks
- **THEN** the system must not schedule the plugin's cron tasks
- **AND** the plugin management page displays the abnormal reason and manual fix prompt

### Requirement: Plugin List Must Expose Effective Version and Discovered Version

The system SHALL expose plugin runtime upgrade state, database effective version, file discovered version, whether upgradable, abnormal reason, and recent upgrade failure information in plugin management list and detail responses. This state MUST be independent of plugin install state and enable state.

#### Scenario: Management page queries pending upgrade plugin
- **WHEN** the management page requests the plugin list
- **AND** plugin `plugin-demo` effective version is `v0.1.0` and discovered version is `v0.2.0`
- **THEN** the response includes `runtimeState=pending_upgrade`
- **AND** the response includes `effectiveVersion=v0.1.0`
- **AND** the response includes `discoveredVersion=v0.2.0`
- **AND** the response includes `upgradeAvailable=true`

#### Scenario: Management page queries abnormal plugin
- **WHEN** the management page requests the plugin list
- **AND** plugin `plugin-demo` runtime state is `abnormal`
- **THEN** the response includes a stable abnormal reason code
- **AND** the response must not mix abnormal state into `installed` or `enabled` fields

### Requirement: Plugin Management Page Must Provide Runtime Upgrade Operations

The system SHALL display an upgrade action for `pending_upgrade` plugins on the plugin management page, and display an upgrade content confirmation dialog before executing the upgrade. After upgrade completion, the plugin state SHALL become `normal`; on upgrade failure, the system SHALL display failure state and diagnosable errors.

#### Scenario: Pending upgrade plugin shows upgrade button
- **WHEN** an administrator opens the plugin management page
- **AND** plugin `plugin-demo` is installed with runtime state `pending_upgrade`
- **THEN** the original install action position displays as an upgrade action
- **AND** clicking upgrade opens the upgrade confirmation dialog

#### Scenario: Upgrade confirmation dialog shows upgrade content
- **WHEN** an administrator prepares to upgrade plugin `plugin-demo`
- **THEN** the dialog displays before/after version, manifest change summary, dependency check results, hostServices authorization changes, and upgrade risk warnings
- **AND** the system must not execute upgrade side effects before administrator confirmation

#### Scenario: Upgrade success restores normal state
- **WHEN** an administrator confirms upgrading plugin `plugin-demo`
- **AND** the upgrade flow completes successfully
- **THEN** the plugin runtime state becomes `normal`
- **AND** the plugin effective version becomes the target manifest version
- **AND** the management page after refresh no longer displays the upgrade button

#### Scenario: Abnormal plugin prompts manual fix
- **WHEN** an administrator opens the plugin management page
- **AND** plugin `plugin-demo` runtime state is `abnormal`
- **THEN** the management page displays abnormal state and fix instructions
- **AND** the management page must not display a normal upgrade confirmation action

### Requirement: Plugin Runtime Upgrade Must Execute via REST API

The system SHALL provide a read-only upgrade preview API and a side-effecting upgrade execution API. The upgrade preview MUST use `GET`, the upgrade execution MUST use `POST /plugins/{id}/upgrade`, and the execution request must pass permission check, confirmation check, and server-side state re-read.

#### Scenario: Get plugin upgrade preview
- **WHEN** the management page requests `GET /plugins/{id}/upgrade/preview`
- **AND** the plugin is in `pending_upgrade`
- **THEN** the system returns before/after versions, manifest diff summary, dependency check, upgrade SQL summary, authorization changes, and risk warnings
- **AND** the system does not modify database state

#### Scenario: Execute plugin upgrade
- **WHEN** an administrator requests `POST /plugins/{id}/upgrade`
- **AND** the request passes plugin management permission check
- **AND** the plugin current state is still `pending_upgrade`
- **THEN** the system executes the runtime upgrade orchestration
- **AND** after upgrade success updates effective version and runtime state

#### Scenario: Reject non-pending-upgrade plugin upgrade request
- **WHEN** an administrator requests `POST /plugins/{id}/upgrade`
- **AND** the plugin current state is not `pending_upgrade`
- **THEN** the system rejects the request
- **AND** the response contains a stable business error code and localizable message key

### Requirement: Runtime Upgrade Must Execute Plugin Custom Upgrade Callback

The system SHALL provide optional upgrade execution-phase callback interfaces for source plugins and dynamic plugins. When the host triggers plugin upgrade, it passes the pre-upgrade manifest snapshot and target manifest snapshot, enabling the plugin to perform custom data migration, state cleanup, and compatibility handling. The dynamic plugin upgrade execution-phase operation name MUST be `Upgrade` and MUST form a complete upgrade lifecycle with `BeforeUpgrade` and `AfterUpgrade`; when the `Upgrade` callback is missing, the host SHALL skip the custom upgrade step and continue executing standard upgrade SQL and governance resource synchronization.

#### Scenario: Target version plugin implements upgrade callback
- **WHEN** plugin `plugin-demo` upgrades from `v0.1.0` to `v0.2.0`
- **AND** the target version source plugin implements the upgrade callback
- **THEN** the host calls the upgrade callback
- **AND** the callback request contains the `v0.1.0` manifest snapshot
- **AND** the callback request contains the `v0.2.0` manifest snapshot

#### Scenario: Plugin does not implement upgrade callback
- **WHEN** plugin `plugin-demo` is in `pending_upgrade`
- **AND** the target version plugin does not implement the upgrade callback
- **THEN** the host skips the plugin custom upgrade step
- **AND** the host continues executing standard upgrade SQL and governance resource synchronization

#### Scenario: Dynamic plugin implements upgrade execution-phase callback
- **WHEN** dynamic plugin `plugin-demo` upgrades from `v0.1.0` to `v0.2.0`
- **AND** the target artifact declares an `Upgrade` lifecycle handler
- **THEN** the host calls `Upgrade` after `BeforeUpgrade` allows
- **AND** the callback request contains the `v0.1.0` manifest snapshot
- **AND** the callback request contains the `v0.2.0` manifest snapshot
- **AND** the host only continues upgrade SQL, governance sync, and release switch after `Upgrade` succeeds

#### Scenario: Upgrade callback failure
- **WHEN** the plugin upgrade callback returns an error
- **THEN** the host stops subsequent upgrade steps
- **AND** the plugin runtime state becomes `upgrade_failed`
- **AND** the system records the failure phase and error details

### Requirement: Dynamic Plugin Uninstall Must Support Custom Cleanup Callback

The system SHALL provide dynamic plugins with an uninstall execution-phase callback symmetric to source plugin `RegisterUninstallHandler`. The dynamic plugin uninstall execution-phase operation name MUST be `Uninstall`; the host SHALL only execute this callback when the administrator chooses to purge plugin storage and data. `Uninstall` SHALL execute after `BeforeUninstall` allows and before uninstall SQL and authorized storage cleanup; on callback failure, the host SHALL stop the uninstall and keep the plugin in installed state.

#### Scenario: Dynamic plugin implements uninstall cleanup callback
- **WHEN** an administrator uninstalls dynamic plugin `plugin-demo`
- **AND** the request selects purging plugin storage and data
- **AND** the active release declares an `Uninstall` lifecycle handler
- **THEN** the host calls `Uninstall` after `BeforeUninstall` allows
- **AND** the callback request contains `purgeStorageData=true`
- **AND** the host only continues uninstall SQL and authorized storage cleanup after `Uninstall` succeeds

#### Scenario: Dynamic plugin uninstall preserves data
- **WHEN** an administrator uninstalls dynamic plugin `plugin-demo`
- **AND** the request selects preserving plugin storage and data
- **THEN** the host must not call the dynamic plugin `Uninstall` execution-phase callback
- **AND** the host may still call `AfterUninstall` notification after successful uninstall

### Requirement: Lifecycle Pre-Callbacks Must Replace Old Guard/Can* Contract

The system SHALL provide a unified lifecycle callback model enabling source plugins and dynamic plugins to return allow or deny decisions before install, upgrade, disable, uninstall, tenant disable, tenant delete, and install mode change operations. The same lifecycle capability MUST use the same `Before*` operation names across source and dynamic plugins; upgrade and uninstall execution phases MUST use `Upgrade` and `Uninstall` operation names; the system must not introduce parallel `Can*`, guard, or pre-* naming for the same capability. `BeforeUninstall` and `AfterUninstall` callback requests MUST include the `purgeStorageData` strategy indicating whether the uninstall purges plugin storage and data. The system MUST delete the old Lifecycle Guard and `Can*` plugin contracts and must not simultaneously retain old Guard registration, execution, or compatibility adapter entries.

#### Scenario: Plugin blocks upgrade
- **WHEN** a plugin registers a `BeforeUpgrade` pre-callback
- **AND** an administrator requests upgrading the plugin
- **AND** the callback returns a deny decision with a reason key
- **THEN** the host refuses the upgrade
- **AND** the response contains the deny reason
- **AND** the plugin effective version must not change

#### Scenario: Plugin blocks uninstall
- **WHEN** a plugin registers a `BeforeUninstall` pre-callback
- **AND** an administrator requests uninstalling the plugin
- **AND** the callback returns a deny decision with a reason key
- **THEN** the host refuses the uninstall
- **AND** the response contains `PLUGIN_LIFECYCLE_PRECONDITION_VETOED`
- **AND** the plugin install state must not change

#### Scenario: Plugin decides based on uninstall purge strategy
- **WHEN** a plugin registers a `BeforeUninstall` pre-callback
- **AND** an administrator requests uninstalling the plugin
- **THEN** the callback request contains `purgeStorageData`
- **AND** the plugin can deny data-preserving uninstall when `purgeStorageData=false`
- **AND** the plugin can allow data-purging uninstall when `purgeStorageData=true` to continue

#### Scenario: Dynamic plugin blocks install
- **WHEN** a dynamic plugin artifact declares a `BeforeInstall` lifecycle pre-handler
- **AND** an administrator requests installing the dynamic plugin
- **AND** the dynamic plugin handler returns a deny decision with a reason key
- **THEN** the host refuses the install
- **AND** the response contains `PLUGIN_LIFECYCLE_PRECONDITION_VETOED`
- **AND** the host must not execute the plugin's install SQL, governance resource sync, or install state write

#### Scenario: Dynamic plugin blocks upgrade
- **WHEN** a dynamic plugin artifact declares a `BeforeUpgrade` lifecycle pre-handler
- **AND** an administrator requests upgrading the dynamic plugin
- **AND** the dynamic plugin handler returns a deny decision with a reason key
- **THEN** the host refuses the upgrade
- **AND** the response contains the deny reason
- **AND** the plugin effective version must not change

#### Scenario: Event hooks do not carry pre-blocking semantics
- **WHEN** a dynamic plugin listens to `plugin.installed`, `plugin.enabled`, `plugin.disabled`, `plugin.uninstalled`, or `plugin.upgraded` events
- **THEN** these event hooks are only for post-lifecycle event notification or follow-up actions
- **AND** the system must not treat these event hooks as `Before*` pre-blocking mechanisms

#### Scenario: Tenant disable triggers lifecycle callbacks
- **WHEN** a tenant administrator disables tenant-level plugin `plugin-demo`
- **THEN** the host calls source plugin and dynamic plugin `BeforeTenantDisable` before writing the tenant plugin enable state
- **AND** when any plugin returns a deny decision, the system refuses the tenant disable and preserves the original tenant plugin state
- **AND** after disable success, the system calls `AfterTenantDisable` as a best-effort notification

#### Scenario: Tenant delete triggers lifecycle callbacks
- **WHEN** a platform administrator deletes a tenant
- **THEN** the system calls source plugin and dynamic plugin `BeforeTenantDelete` before deleting the tenant through the unified host plugin lifecycle service
- **AND** when any plugin returns a deny decision, the system refuses to delete the tenant
- **AND** after delete success, the system calls `AfterTenantDelete` as a best-effort notification

### Requirement: Plugin Upgrade Must Guarantee Cache and Cluster Consistency

The system SHALL invalidate runtime caches by plugin ID and resource scope after plugin runtime upgrade success or failure, and notify other nodes through the host unified cluster coordination mechanism in cluster mode. When `cluster.enabled=true`, the system must not rely solely on current node memory state to determine upgrade completion.

#### Scenario: Standalone mode upgrade success
- **WHEN** `cluster.enabled=false`
- **AND** plugin upgrade succeeds
- **THEN** the system invalidates local plugin state, menu, permission, route, cron, i18n, and apidoc caches by plugin ID
- **AND** the system does not force dependency on distributed coordination components

#### Scenario: Cluster mode upgrade success
- **WHEN** `cluster.enabled=true`
- **AND** plugin upgrade succeeds
- **THEN** the system writes shared revision state or publishes cluster events
- **AND** other nodes refresh the corresponding plugin's runtime state and caches after receiving notification
- **AND** the same plugin must not concurrently execute upgrade on multiple nodes

### Requirement: Plugin Upgrade Failure Must Be Diagnosable and Retriable

The system SHALL retain failure state, failure phase, error code, error message key, pre-upgrade manifest snapshot, and target manifest snapshot on upgrade failure. The administrator MUST be able to view the failure reason on the plugin management page and retry the upgrade after fixing the issue or perform manual repair.

#### Scenario: Upgrade SQL failure
- **WHEN** plugin upgrade fails during upgrade SQL execution
- **THEN** the system marks the plugin runtime state as `upgrade_failed`
- **AND** the system records the failure phase as the SQL execution phase
- **AND** the system preserves the effective version without switching to the target version

#### Scenario: Administrator retries failed upgrade
- **WHEN** the plugin state is `upgrade_failed`
- **AND** the administrator fixes the failure cause and re-initiates the upgrade
- **THEN** the system re-executes upgrade pre-checks
- **AND** the system skips already-completed migration steps or safely retries based on idempotent migration records
