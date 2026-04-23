## MODIFIED Requirements

### Requirement: Automatic logging of login logs
The system SHALL automatically emits unified login events at authentication life cycle nodes such as successful login, failed login, and successful logout. When `monitor-loginlog` is installed and enabled, the plugin subscribes to events and persists logs to the `plugin_monitor_loginlog` table; when the plugin is unavailable, the host authentication link still executes normally.

#### Scenario: Login log plugin is enabled
- **WHEN** The user has successfully logged in, failed to log in, or successfully logged out, and `monitor-loginlog` has been installed and enabled
- **THEN** The host launches a unified login event
- **AND** `monitor-loginlog` writes the corresponding login log record after subscribing to the event

#### Scenario: Login log plugin is missing or disabled
- **WHEN** The user has successfully logged in, failed to log in, or successfully logged out, but `monitor-loginlog` is not installed, not enabled, or fails to initialize
- **THEN** The host still returns the authentication result normally.
- **AND** The host does not return an error due to lack of specific login log implementation.

## MODIFIED Requirements

### Requirement: The login log management interface is delivered by the source plugin

The The system SHALL deliver login log query, details, export, cleaning and page capabilities as the `monitor-loginlog` source plugin.

#### Scenario: Expose the management entrance when the plugin is enabled
- **WHEN** `monitor-loginlog` is installed and enabled
- **THEN** The host exposes login log query, details, export, cleanup interface and frontend page
- **AND** The plugin menu is mounted to the host `system monitoring` directory, and the top-level `parent_key` is `monitor`

#### Scenario: Hide the management entrance when the plugin is missing
- **WHEN** `monitor-loginlog` is not installed or not enabled
- **THEN** The host does not display the login log menu and page entry
- **AND** Login and logout processes continue to function normally
