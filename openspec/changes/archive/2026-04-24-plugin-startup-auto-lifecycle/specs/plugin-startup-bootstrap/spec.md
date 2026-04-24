## ADDED Requirements

### Requirement: The host must provide a simplified plugin auto-enable config in the main config file
The system SHALL provide a `plugin.autoEnable` setting in the host service main config file and use a list of plugin IDs to declare which plugins must auto-enable during system startup. The config MUST NOT require users to provide complex structures such as `desiredState`, `required`, or authorization details.

#### Scenario: Parse a valid auto-enable list
- **WHEN** the host starts and reads the `plugin.autoEnable` array from the main config file
- **THEN** the host builds a valid set of plugin IDs that must auto-enable
- **AND** each item in the set is parsed according to plugin-ID semantics

#### Scenario: Reject invalid auto-enable config
- **WHEN** `plugin.autoEnable` is not a string array, or the array contains an empty plugin ID
- **THEN** the host MUST refuse to continue startup
- **AND** the error message MUST clearly identify `plugin.autoEnable` as invalid

### Requirement: The host must execute startup bootstrap before plugin wiring
The system SHALL advance the lifecycle state of plugins listed in `plugin.autoEnable` before plugin HTTP route registration, plugin cron wiring, and dynamic frontend bundle warm-up.

#### Scenario: A source plugin reaches enabled state before startup wiring
- **WHEN** a discovered source plugin appears in the `plugin.autoEnable` list
- **THEN** the host installs and enables that source plugin before plugin routes and plugin cron registration
- **AND** later reads of the enabled snapshot see that plugin in the enabled state

#### Scenario: Plugins not in the auto-enable list remain under manual governance
- **WHEN** a plugin has been discovered by the host but is not present in `plugin.autoEnable`
- **THEN** the host only performs routine manifest sync and registry refresh for that plugin
- **AND** the host MUST NOT auto-install or auto-enable it because of startup bootstrap

### Requirement: The auto-enable list must implicitly include install and enable semantics
The system SHALL interpret plugins in `plugin.autoEnable` as plugins that must be enabled during host startup. If a listed plugin is not installed yet, the host MUST install it first and then continue to enable it.

#### Scenario: Auto-enable a newly discovered source plugin
- **WHEN** a source plugin appears in `plugin.autoEnable` and is still in the not-installed, not-enabled state
- **THEN** the host installs the source plugin first
- **AND** then advances it to enabled

#### Scenario: An already enabled plugin appears again in the auto-enable list
- **WHEN** a plugin is already enabled and its plugin ID is still present in `plugin.autoEnable`
- **THEN** the host keeps that plugin enabled
- **AND** the host MUST NOT downgrade it during repeated auto-enable processing

### Requirement: Any failure for a listed auto-enable plugin must block host startup
The system SHALL treat `plugin.autoEnable` as an explicit list of required boot-time plugins. If any listed plugin is missing, fails to install, fails to enable, or does not reach enabled state inside the wait window, the host MUST fail fast.

#### Scenario: A missing auto-enable plugin causes startup failure
- **WHEN** `plugin.autoEnable` declares a plugin ID that the host cannot discover during startup
- **THEN** the host MUST stop the startup flow
- **AND** the returned error message MUST include the missing plugin ID

#### Scenario: An auto-enable plugin fails and causes startup failure
- **WHEN** a plugin listed in `plugin.autoEnable` fails during installation, enablement, or convergence waiting
- **THEN** the host MUST stop the startup flow
- **AND** the error message MUST include the plugin identifier and the stage that failed

### Requirement: Startup bootstrap must separate shared lifecycle side effects from local convergence in cluster mode
The system SHALL allow only the primary node to execute shared plugin lifecycle actions in cluster mode, such as install SQL, menu writes, release switches, and shared-state advancement. Follower nodes only wait for shared-state results and refresh their local projections.

#### Scenario: The primary node executes shared plugin actions
- **WHEN** a plugin appears in `plugin.autoEnable` in cluster mode and installation or enablement must be advanced
- **THEN** only the primary node executes the shared install, enable, or reconcile actions for that plugin
- **AND** follower nodes MUST NOT repeat the same shared side effects

#### Scenario: Follower nodes refresh local views after shared convergence
- **WHEN** a follower node starts in cluster mode and finds a plugin in `plugin.autoEnable`
- **THEN** the follower waits for the primary node to write a shared stable state or for the wait window to time out
- **AND** then refreshes its local enabled snapshot and runtime projection from that shared result

### Requirement: Startup auto-enable of dynamic plugins must reuse existing authorization snapshots
The system SHALL reuse the host-approved authorization snapshot of the current release when a dynamic plugin that declares governed host services appears in `plugin.autoEnable`. The host MUST NOT require users to fill complex authorization details in the main config file.

#### Scenario: Reuse an existing authorization snapshot during dynamic-plugin auto-enable
- **WHEN** a dynamic plugin appears in `plugin.autoEnable` and its current release already has a host-approved authorization snapshot
- **THEN** the host reuses that snapshot to drive startup auto-enable for the dynamic plugin
- **AND** the host MUST NOT require authorization details from the main config file again

#### Scenario: Reject startup auto-enable when no authorization snapshot exists
- **WHEN** a dynamic plugin appears in `plugin.autoEnable`, declares governed host services, and the current release still has no authorization snapshot
- **THEN** the host MUST stop startup
- **AND** the error message MUST clearly say that a normal reviewed flow is required first to produce the authorization snapshot

### Requirement: The plugin-management UI must label startup auto-enabled plugins and warn about temporary governance actions
The system SHALL show whether the current plugin is matched by `plugin.autoEnable` in the host main config file through read-only indicators in the plugin-management list and detail views. When administrators disable or uninstall those plugins from the UI, the interface MUST warn before submission that the action takes effect immediately but the host will reinstall and re-enable the plugin after restart unless the config changes.

#### Scenario: The list and detail views show the startup auto-enable indicator
- **WHEN** a plugin ID currently exists in `plugin.autoEnable` in the host main config file
- **THEN** the plugin-management list SHALL show a read-only indicator that the plugin is managed by `plugin.autoEnable`
- **AND** the plugin detail view SHALL show the same read-only meaning

#### Scenario: Disabling a startup auto-enabled plugin warns about restart behavior
- **WHEN** an administrator attempts to disable a plugin that is matched by `plugin.autoEnable`
- **THEN** the UI MUST show a risk-confirmation prompt before sending the disable request
- **AND** the prompt MUST clearly state that the disable action takes effect immediately but the host will enable the plugin again after restart if `plugin.autoEnable` remains unchanged
- **AND** the prompt MUST clearly state that permanently disabling the plugin requires editing `plugin.autoEnable` in the host main config file first

#### Scenario: Uninstalling a startup auto-enabled plugin warns about restart behavior
- **WHEN** an administrator attempts to uninstall a plugin that is matched by `plugin.autoEnable`
- **THEN** the uninstall confirmation UI MUST show a risk warning
- **AND** the warning MUST clearly state that the uninstall takes effect immediately but the host will install and enable the plugin again after restart if `plugin.autoEnable` remains unchanged
- **AND** the warning MUST clearly state that permanently disabling the plugin requires editing `plugin.autoEnable` in the host main config file first
