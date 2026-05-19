## Requirements

### Requirement: Source Plugin Must Separate Effective Version and Discovered Source Version

The system SHALL distinguish the currently effective source plugin version and the discovered source version in the source tree. `sys_plugin.version` and `release_id` only represent the effective version; newly discovered versions are stored as release records or discovery snapshots and must not overwrite the effective version until runtime upgrade completes.

#### Scenario: Installed source plugin discovers higher version
- **WHEN** source plugin `plugin-demo` effectively runs `v0.1.0` and its source `plugin.yaml` has been upgraded to `v0.5.0`
- **THEN** `sys_plugin.version` remains `v0.1.0`
- **AND** the system records a `v0.5.0` source plugin release snapshot
- **AND** the new release is not considered the current effective version until runtime upgrade completes

#### Scenario: Installed source plugin discovers lower version
- **WHEN** source plugin `plugin-demo` effectively runs `v0.5.0` and its source `plugin.yaml` is `v0.1.0`
- **THEN** `sys_plugin.version` remains `v0.5.0`
- **AND** the system marks the plugin as abnormal state
- **AND** the system requires the administrator to manually fix the file or database state before recovery

### Requirement: Source Plugin Upgrade Must Be Explicit Runtime Operation

The system SHALL require source plugin upgrade to be explicitly executed through the runtime management API after host startup, not through a development-time upgrade command, and not automatically fixed during host startup. The development stage can only let the host discover a new version through file overwrites.

#### Scenario: Explicit upgrade of single source plugin
- **WHEN** an administrator confirms upgrading `plugin-demo` on the plugin management page
- **AND** the host receives a `POST /plugins/{id}/upgrade` request
- **THEN** the system only executes the runtime upgrade flow for `plugin-demo`
- **AND** does not trigger upgrades for other source plugins or dynamic plugins

#### Scenario: Source plugin file overwrite waits for runtime upgrade
- **WHEN** a developer overwrites source plugin files under `apps/lina-plugins/plugin-demo`
- **AND** the plugin `plugin.yaml` version is higher than the database effective version
- **THEN** the host marks the plugin as pending upgrade after startup
- **AND** the system does not require running the old development-time upgrade command

### Requirement: Host Startup Must Mark Source Plugin Upgrade State

The host SHALL scan source plugins during startup, then compare the discovered version and effective version of installed source plugins. If the discovered version is higher than the effective version, the host must mark the plugin as pending upgrade and continue startup; if the discovered version is lower than the effective version, the host must mark the plugin as abnormal and continue startup.

#### Scenario: Pending source plugin upgrade does not block startup
- **WHEN** the host starts and discovers `plugin-demo` effectively runs `v0.1.0` while source discovery reports `v0.5.0`
- **THEN** startup continues
- **AND** the plugin runtime state becomes pending upgrade
- **AND** the plugin management page can display effective version, discovered version, and upgrade action

#### Scenario: Source plugin discovered version lower than effective version
- **WHEN** the host starts and discovers `plugin-demo` effectively runs `v0.5.0` while source discovery reports `v0.1.0`
- **THEN** startup continues
- **AND** the plugin runtime state becomes abnormal
- **AND** the plugin management page prompts the administrator to manually intervene and fix

### Requirement: Source Plugin Upgrade Must Record phase=upgrade and Synchronize Governance Resources

Source plugin runtime upgrade SHALL execute upgrade-phase migration accounting and synchronize governance resources including menus, permissions, resource references, i18n, apidoc, routes, and cron. After successful execution, the new release becomes the effective release.

#### Scenario: Source plugin upgrade success
- **WHEN** an administrator upgrades an installed source plugin and all upgrade callbacks, SQL, and governance sync steps succeed
- **THEN** `sys_plugin.version` and `release_id` update to the new release
- **AND** `sys_plugin_migration` records a `phase=upgrade` entry
- **AND** the new release becomes the effective release
- **AND** the plugin runtime state becomes normal

#### Scenario: Source plugin upgrade failure
- **WHEN** during source plugin upgrade a plugin callback, upgrade SQL statement, or governance sync step fails
- **THEN** the runtime upgrade flow immediately stops
- **AND** the failed upgrade record and error information are preserved
- **AND** the plugin runtime state becomes upgrade failed
- **AND** the system does not automatically perform rollback

### Requirement: Dynamic Plugin Upgrade Must Enter Unified Runtime Upgrade Model

The system SHALL keep dynamic plugin upgrades in the runtime model. After a dynamic plugin new artifact is uploaded or file-overwritten, if the discovered version is higher than the database effective version, the system must mark the plugin as pending upgrade and complete effective release switching and governance resource synchronization through the same plugin management page upgrade flow.

#### Scenario: Dynamic plugin discovers higher version
- **WHEN** dynamic plugin `plugin-demo-dynamic` effectively runs `v0.1.0`
- **AND** the locally discovered or uploaded dynamic plugin artifact version is `v0.2.0`
- **THEN** the system marks the plugin as pending upgrade
- **AND** the effective release remains `v0.1.0`
- **AND** the administrator must explicitly confirm the upgrade through the plugin management page

#### Scenario: Dynamic plugin file version lower than effective version
- **WHEN** dynamic plugin `plugin-demo-dynamic` effectively runs `v0.2.0`
- **AND** the locally discovered or uploaded artifact version is `v0.1.0`
- **THEN** the system marks the plugin as abnormal
- **AND** the system must not automatically downgrade the effective release

### Requirement: Plugin Upgrade Must Validate New Version Dependency Constraints

Source plugin upgrade commands and dynamic plugin install/upgrade paths SHALL validate new version manifest dependency constraints before switching the effective release. New framework version constraints, hard dependency existence, and hard dependency version ranges must be satisfied, otherwise the upgrade or release switch must fail.

#### Scenario: Source plugin upgrade validates dependencies before upgrade
- **WHEN** a developer upgrades source plugin `x` to a new version
- **AND** the new version declares hard dependency `a >=0.2.0`
- **AND** the currently installed or available `a` version does not satisfy
- **THEN** source plugin upgrade fails
- **AND** `x`'s effective version remains the pre-upgrade version

#### Scenario: Dynamic plugin same-version refresh validates dependencies
- **WHEN** a dynamic plugin refreshes with a same-version new artifact
- **AND** the new artifact manifest declares framework version constraints not satisfied by the current environment
- **THEN** dynamic plugin refresh fails
- **AND** the current active release continues pointing to the pre-refresh artifact

### Requirement: Plugin Upgrade Must Not Destroy Reverse Dependencies of Installed Plugins

Plugin upgrade SHALL validate that the post-upgrade effective version does not destroy other installed plugins' hard dependency version ranges on this plugin. If the upgrade result makes downstream plugin dependencies unsatisfied, the system must refuse to switch the effective release.

#### Scenario: Target plugin upgrade fails to satisfy downstream dependency
- **WHEN** installed plugin `consumer` hard-depends on `base <0.3.0`
- **AND** an administrator attempts to upgrade `base` to `v0.3.0`
- **THEN** the upgrade request fails
- **AND** the error contains the downstream plugin `consumer` and its dependency version range

### Requirement: Plugin Upgrade Must Not Auto-Upgrade Dependency Plugins

The plugin upgrade process SHALL not auto-upgrade dependency plugins. If the new version requires dependency versions higher than current dependency plugin versions, the system must block the upgrade and return the dependency list requiring manual upgrade first.

#### Scenario: New version requires higher dependency version
- **WHEN** plugin `x` new version requires `a >=0.2.0`
- **AND** the current `a` effective version is `v0.1.0`
- **THEN** upgrading `x` fails
- **AND** the error prompts to upgrade `a` first
- **AND** the system must not auto-upgrade `a`

## REMOVED Requirements

### Requirement: Source Plugin Upgrade Must Be Explicit Development-Time Operation

**Reason**: This requirement depends on an unimplemented development-time upgrade command and incorrectly places runtime database state, governance resources, and plugin data migration into the development stage. Plugin file overwrites can only produce discovered versions, not complete runtime upgrades.

**Migration**: Use runtime plugin management page and `POST /plugins/{id}/upgrade` for explicit upgrades. Development stage only updates local plugin files through `plugins.update` or direct file overwrite.

### Requirement: Host Startup Must Verify Source Plugin Upgrade Completion

**Reason**: Startup blocking would prevent administrators from accessing the plugin management page to execute runtime upgrades, and cannot handle recovery flows for inconsistent database and plugin file states.

**Migration**: Host startup changed to marking `pending_upgrade`, `abnormal`, or `normal` state and keeping management entry accessible. Pending upgrade plugin business entries are controlled by runtime state.
