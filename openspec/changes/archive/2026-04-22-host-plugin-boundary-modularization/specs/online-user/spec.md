## MODIFIED Requirements

### Requirement: Session Storage Abstraction Layer
The system SHALL treats online session storage, session validity verification and session active time maintenance as host authentication session core; `monitor-online` only consumes the session projection and management capabilities provided by the host.

#### Scenario: Online user plugin not installed
- **WHEN** `monitor-online` is not installed or not enabled
- **THEN** The host is still creating, deleting, verifying and cleaning session records in `sys_online_session`
- **AND** Login, logout, protected interface authentication and timeout determination continue to work normally

#### Scenario: Online User Plugin Enabled
- **WHEN** `monitor-online` is installed and enabled
- **THEN** plugin queries online users through the session projection provided by the host and performs forced offline management
- **AND** The plugin does not have a JWT validation, `last_active_time` maintenance or cleanup task truth source

### Requirement: Online user list query
The system SHALL provides administrators with the ability to query current online users when `monitor-online` is installed and enabled, and supports filtering by username and IP address.

#### Scenario: Query online user list
- **WHEN** `monitor-online` is installed and enabled and the administrator requests a list of online users
- **THEN** plugin returns a list of online session records in the host session projection
- **AND** Each record still contains governance fields such as token_id, username, dept_name, ip, login_location, browser, os, login_time, etc.

### Requirement: Force offline
The system SHALL supports administrators to force offline specified online users when `monitor-online` is installed and enabled. Subsequent requests by offline users MUST return 401.

#### Scenario: Plug-in execution is forced offline
- **WHEN** The administrator uses `monitor-online` to force the specified `tokenId` offline
- **THEN** The host session kernel fails the session record
- **AND** Subsequent requests carrying the Token are rejected by the host authentication middleware

### Requirement: System monitoring menu
The system SHALL When `monitor-online` is installed and enabled, the `online user` menu is mounted as a plugin menu to the host `system monitoring` directory, instead of requiring it to appear as a fixed built-in submenu together with `service monitoring`.

#### Scenario: Menu display
- **WHEN** `monitor-online` is installed, enabled and the current user has access to its menu
- **THEN** The `Online Users` submenu is displayed under `System Monitoring`
- **AND** This rule does not require that `Service Monitoring` also exists

#### Scenario: Plug-in missing or disabled
- **WHEN** `monitor-online` is not installed, not enabled, or the current user does not have access to its menu
- **THEN** The host hides the `Online User` menu entry
- **AND** Host authentication session kernel continues to run independently
