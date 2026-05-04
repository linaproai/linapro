# Online User

## Purpose

Define online session storage abstraction, login status tracking, list query, and forced offline capabilities to ensure that the system can steadily manage current online users and their session lifecycles.
## Requirements
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

### Requirement: Online User Frontend Page

The system SHALL provides an online user management page that displays the current list of online users and supports forced offline operations.

#### Scenario: Page shows a list of online users
- **WHEN** Admin visits online user page
- **THEN** page displays the VXE-Grid form, including the following: login account, department name, IP address, login location, browser (with icon), operating system (with icon), login time, operation (forced offline button); the toolbar shows the online number of people statistics

#### Scenario: Search Filter
- **WHEN** Admin enters username or IP address in search bar and searches
- **THEN** table data refreshes based on filters

#### Scenario: Enforce offline interactions
- **WHEN** Admin clicks "Force Offline" button on a user line
- **THEN** pops up a confirmation dialog box, calls the forced downline API after confirmation, refreshes the table data after success

### Requirement: Session Activity Time Tracking

The system SHALL tracks the last active time of each online user and is used to determine if a session has timed out. The authentication middleware MUST validate the session and timeout on every protected request, but MAY skip writing `last_active_time` when the persisted active time is still within the configured short update window.

#### Scenario: Initial active time at login
- **WHEN** User successfully logged in, system created `sys_online_session` session record
- **THEN** `last_active_time` field MUST be set to the current time

#### Scenario: Validate active session on every protected request
- **WHEN** logged in user with a valid token accessing the protected API
- **THEN** The authentication middleware MUST verify that the session record exists
- **AND** The authentication middleware MUST reject the request if the persisted `last_active_time` has exceeded the effective timeout threshold

#### Scenario: Refresh active time after update window
- **WHEN** logged in user with a valid token accesses a protected API and the persisted `last_active_time` is older than the short update window
- **THEN** The authentication middleware MUST update `last_active_time` to the current time
- **AND** The request is processed normally when the update succeeds

#### Scenario: Skip duplicate active time write within update window
- **WHEN** logged in user with a valid token repeatedly accesses protected APIs within the short update window
- **THEN** The authentication middleware MAY skip the duplicate `last_active_time` update
- **AND** The request is still processed normally if the session has not timed out

### Requirement: Automatic cleanup of inactive sessions
System SHALL provides scheduled tasks to automatically clean up online sessions that have been inactive for a long time, preventing the session table from growing indefinitely.Timeout threshold and cleanup frequency MUST support adjustment via the duration string profile.

#### Scenario: Scheduled cleanup of timeout sessions
- **WHEN** When scheduled cleanup tasks are performed (default every 5 minutes)
- **THEN** System MUST query `last_active_time` in the `sys_online_session` table for records whose current time exceeds the timeout threshold (default 24 hours) and delete them

#### Scenario: Timeout thresholds can be adjusted with new configurations
- **WHEN** Admin set `session.timeout = 24h` in `config.yaml`
- **THEN** The system MUST uses this duration value as the session timeout threshold, which defaults to 24 hours when not set

#### Scenario: Cleaning frequency can be adjusted with the new configuration
- **WHEN** Admin set `session.cleanupInterval = 5m` in `config.yaml`
- **THEN** The system MUST uses this duration value as the cleanup task execution interval, which defaults to 5 minutes when not set

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

### Requirement: Runtime-Configured Session Timeout
The system SHALL allow `sys.session.timeout` to control the online-session timeout threshold at runtime and SHALL fall back to static configuration when no runtime override exists.

#### Scenario: Runtime timeout value becomes effective
- **WHEN** an administrator maintains `sys.session.timeout=24h`
- **THEN** the host uses that duration as the effective online-session timeout threshold

### Requirement: Authentication Checks Session Timeout on Every Protected Request
The system SHALL evaluate session timeout during authentication instead of relying only on scheduled cleanup.

#### Scenario: Expired session is rejected during authentication
- **WHEN** a protected request carries a token whose session `last_active_time` exceeds the effective timeout threshold
- **THEN** the authentication chain rejects the request with `401`
- **AND** the host cleans up the corresponding online-session record

### Requirement: sys_online_session must include last_active_time index

The system SHALL maintain `KEY idx_last_active_time (last_active_time)` on the `sys_online_session` table to support inactive-session cleanup queries by `last_active_time` range and avoid full table scans.

#### Scenario: Index exists

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_online_session` MUST include `idx_last_active_time` on column `last_active_time`

#### Scenario: Inactive-session cleanup uses the index

- **WHEN** the service executes cleanup queries of the form `WHERE last_active_time < ?`
- **THEN** the database MUST select `idx_last_active_time` to avoid a full table scan
