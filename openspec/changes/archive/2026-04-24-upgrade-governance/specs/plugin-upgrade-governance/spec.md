## ADDED Requirements

### Requirement: Source plugins must separate the effective version from discovered source versions

The system SHALL distinguish the current effective source-plugin version from higher versions discovered in the source tree. `sys_plugin.version` and `release_id` represent only the effective version, while newly discovered source versions are stored as release records and must not overwrite the effective version before an explicit upgrade.

#### Scenario: An installed source plugin discovers a higher version
- **WHEN** source plugin `plugin-demo` is effectively running `v0.1.0` and its `plugin.yaml` in source has been bumped to `v0.5.0`
- **THEN** `sys_plugin.version` remains `v0.1.0`
- **AND** the system records a `v0.5.0` source-plugin release snapshot
- **AND** that new release is not treated as the current effective version until an explicit upgrade completes

### Requirement: Source-plugin upgrades must be explicit development-time operations

The system SHALL require source-plugin upgrades to be executed through the shared development-time upgrade command instead of being repaired automatically during host startup. The command must support both single-plugin and bulk source-plugin upgrades.

#### Scenario: Upgrade one source plugin explicitly
- **WHEN** a developer runs `make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo`
- **THEN** the system generates and executes an upgrade plan only for `plugin-demo`
- **AND** it does not trigger upgrades for other source plugins or any dynamic plugin

#### Scenario: Upgrade all source plugins in one run
- **WHEN** a developer runs `make upgrade confirm=upgrade scope=source-plugin plugin=all`
- **THEN** the system scans all source plugins and processes pending upgrades in a deterministic order
- **AND** it prints explicit skip results for plugins that are not installed or do not require upgrades

### Requirement: Host startup must verify that source-plugin upgrades are complete

The host SHALL scan source plugins during startup and then validate whether any installed source plugin has a higher discovered version than the effective version. If such a plugin exists, the host MUST refuse to start and print the matching development-time upgrade command.

#### Scenario: A pending source-plugin upgrade blocks startup
- **WHEN** the host starts and discovers that `plugin-demo` is effectively running `v0.1.0` while source discovery reports `v0.5.0`
- **THEN** the startup flow fails
- **AND** the error message includes the plugin ID, the effective version, the discovered version, and the recommended `make upgrade` command

### Requirement: Source-plugin upgrades must record `phase=upgrade` and synchronize governance resources

The source-plugin upgrade command SHALL execute upgrade-phase migration bookkeeping and synchronize menus, permissions, and governance resource references. After a successful run, the new release becomes the effective release.

#### Scenario: A source-plugin upgrade succeeds
- **WHEN** a developer upgrades an installed source plugin and all SQL and governance synchronization steps succeed
- **THEN** `sys_plugin.version` and `release_id` are updated to the new release
- **AND** `sys_plugin_migration` records an entry with `phase=upgrade`
- **AND** the new release becomes the effective release

#### Scenario: A source-plugin upgrade fails
- **WHEN** an upgrade SQL statement or a governance synchronization step fails during a source-plugin upgrade
- **THEN** the command stops immediately
- **AND** it preserves failed upgrade records and error information
- **AND** the iteration does not perform rollback automatically

### Requirement: Dynamic-plugin upgrades stay on the runtime model

The system SHALL keep dynamic-plugin upgrades on the existing runtime upload plus install/reconcile model. The development-time `make upgrade` command must not scan, migrate, or switch dynamic-plugin releases.

#### Scenario: The development-time upgrade command ignores dynamic plugins
- **WHEN** a developer runs `make upgrade confirm=upgrade scope=source-plugin plugin=plugin-demo`
- **THEN** the system does not scan or switch any dynamic-plugin release
- **AND** dynamic plugins continue to upgrade only through upload plus install/reconcile
