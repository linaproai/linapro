## ADDED Requirements

### Requirement: Session Storage Abstraction Layer

The system SHALL define a `SessionStore` abstract interface for session management, currently implemented using MySQL MEMORY engine. The interface MUST support the following operations: create session, query session, delete session, list query (supporting filtering by username and IP), TouchOrValidate (update active time and determine session existence), CleanupInactive (clean up timed-out sessions).

#### Scenario: Create Session on Login
- **WHEN** user successfully logs in via `POST /api/v1/auth/login`
- **THEN** the system creates a session record in `sys_online_session` table, containing token_id (UUID), user_id, username, dept_name, ip, browser, os, login_time, last_active_time, and other fields

#### Scenario: Delete Session on Logout
- **WHEN** logged-in user calls `POST /api/v1/auth/logout`
- **THEN** the system deletes the user's corresponding session record from `sys_online_session` table

#### Scenario: Validate Session on Request
- **WHEN** user accesses a protected API with a valid JWT Token
- **THEN** the middleware MUST, in addition to verifying JWT signature, check via TouchOrValidate whether a corresponding session record exists in `sys_online_session` table; if not (forced offline or timed out), return 401

### Requirement: Online User List Query

The system SHALL provide administrators with the ability to query all currently online users, supporting filtering by username and IP address. The online user list SHALL integrate host data permission governance: all-data scope can query all online sessions; department-data scope only queries online sessions of users within the current user's department scope; self-only scope only queries the current user's own online sessions.

#### Scenario: Query Online User List
- **WHEN** admin calls `GET /api/v1/monitor/online/list`
- **THEN** the system returns all online session record list, each containing: token_id, username, dept_name (department name), ip (login IP), login_location (login location), browser, os, login_time

#### Scenario: Filter by Username
- **WHEN** admin calls `GET /api/v1/monitor/online/list?username=admin`
- **THEN** the system only returns online session records whose username contains "admin"

#### Scenario: Filter by IP Address
- **WHEN** admin calls `GET /api/v1/monitor/online/list?ip=192.168`
- **THEN** the system only returns online session records whose IP address contains "192.168"

#### Scenario: Department Scope Restricts Online User List
- **WHEN** a normal user's role data scope is department data
- **AND** queries the online user list
- **THEN** the system only returns online sessions of users within the current user's visible department scope

#### Scenario: Self-Only Scope Restricts Online User List
- **WHEN** a normal user's role data scope is self-only data
- **AND** queries the online user list
- **THEN** the system only returns the current logged-in user's own online sessions

### Requirement: Forced Offline

The system SHALL support administrators forcing specified online users offline. Subsequent requests from the forced-offline user MUST return 401. Forced offline SHALL integrate host data permission governance; the caller can only force offline online sessions within their data permission scope.

#### Scenario: Successful Forced Offline
- **WHEN** admin calls `DELETE /api/v1/monitor/online/{tokenId}`
- **THEN** the system deletes the session record corresponding to that tokenId, returns success response

#### Scenario: Forced-Offline User Requests Again
- **WHEN** a user who has been forced offline accesses any protected API with the original Token
- **THEN** the middleware detects that the session does not exist, returns 401 status code

#### Scenario: Force Offline Non-Existent tokenId
- **WHEN** admin calls `DELETE /api/v1/monitor/online/{tokenId}` but the tokenId does not exist
- **THEN** the system returns success response (idempotent operation)

#### Scenario: Reject Forced Offline of Out-of-Scope Session
- **WHEN** a normal user's role data scope is department data
- **AND** the target `tokenId` belongs to a user outside the department scope
- **THEN** the system rejects the forced offline operation
- **AND** the target session remains valid until the user logs out or times out

### Requirement: Online User Frontend Page

The system SHALL provide an online user management page displaying the current online user list and supporting forced offline operations.

#### Scenario: Page Displays Online User List
- **WHEN** admin visits the online user page
- **THEN** the page displays a VXE-Grid table with columns: login account, department name, IP address, login location, browser (with icon), OS (with icon), login time, operations (forced offline button); toolbar shows online user count

#### Scenario: Search and Filter
- **WHEN** admin enters username or IP address in the search bar and searches
- **THEN** the table data refreshes based on filter conditions

#### Scenario: Forced Offline Interaction
- **WHEN** admin clicks the "Force Offline" button on a user row
- **THEN** a confirmation dialog appears; after confirmation, the forced offline API is called, and the table data refreshes on success

### Requirement: Session Active Time Tracking

The system SHALL track each online user's last active time to determine session timeout.

#### Scenario: Initialize Active Time on Login
- **WHEN** user successfully logs in and system creates `sys_online_session` session record
- **THEN** the `last_active_time` field MUST be set to the current time

#### Scenario: Update Active Time on Each Request
- **WHEN** logged-in user accesses a protected API with a valid Token
- **THEN** the authentication middleware MUST update the session's `last_active_time` to the current time via UPDATE operation, and determine session existence by affected row count (>0 exists, =0 does not exist or has been cleared)

### Requirement: Inactive Session Auto-Cleanup

The system SHALL provide a scheduled task to automatically clean up long-inactive online sessions, preventing unlimited session table growth. The timeout threshold and cleanup frequency MUST support adjustment through configuration file.

#### Scenario: Scheduled Cleanup of Timed-Out Sessions
- **WHEN** the scheduled cleanup task executes (default every 5 minutes)
- **THEN** the system MUST query `sys_online_session` table for records where `last_active_time` exceeds the timeout threshold (default 24 hours) from current time, and delete them

#### Scenario: Configurable Timeout Threshold
- **WHEN** admin sets `session.timeoutHour` in `config.yaml`
- **THEN** the system MUST use that config value as session timeout threshold; default is 24 hours if not set

#### Scenario: Configurable Cleanup Frequency
- **WHEN** admin sets `session.cleanupMinute` in `config.yaml`
- **THEN** the system MUST use that config value as cleanup task execution interval; default is 5 minutes if not set

### Requirement: System Monitor Menu

The system SHALL add a "System Monitor" top-level menu in the navigation menu, containing "Online Users" and "Server Monitor" sub-menu items.

#### Scenario: Menu Display
- **WHEN** admin logs in and views the left navigation
- **THEN** the "System Monitor" top-level menu is visible; expanding it shows "Online Users" and "Server Monitor" sub-menus

#### Scenario: Menu Navigation
- **WHEN** admin clicks "Online Users" or "Server Monitor" menu item
- **THEN** the page navigates to the corresponding function page
