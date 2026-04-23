# Login Log

## Purpose

Define the automatic logging, querying, cleaning, and exporting behavior undertaken by the `monitor-loginlog` source plugin to ensure that the system can track, audit, and analyze authentication successes and failures.
## Requirements
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

### Requirement: Login Log Contents
The system SHALL record the following login information: user name (user_name), login status (status), login IP (ip), browser (browser), operating system (os), login message (msg), login time (login_time).

#### Scenario: parsing browsers and operating systems
- **WHEN** Logging
- **THEN** Resolves the browser name and operating system name from the HTTP request header `User-Agent` field

### Requirement: Login Log List Query
The system SHALL provides the login log paging query interface `get/api/v1/loginlog`, which supports filtering by user name, IP, status, and time range.

#### Scenario: Paging Log Logs
- **WHEN** Admin request `get/api/v1/loginlog? pageNum = 1 & pageSize = 10`
- **THEN** Return to the login log pagination list, arranged in reverse order by login time

#### Scenario: Filter by criteria
- **WHEN** Admin requests a query with filters (e.g. `userName = admin&status = 0 & beginTime = 2026-01-01 & endTime = 2026-03-15`)
- **THEN** Returns login log records that meet all conditions

### Requirement: Login Log Details View
The system SHALL provides the login log details interface `get/api/v1/loginlog/{id}` and returns the complete login log information.

#### Scenario: View login log details
- **WHEN** admin request `get/api/v1/loginlog/1`
- **THEN** returns all field information for this login log

### Requirement: Login logs are cleaned up by time range
The system SHALL provides the login log cleaning interface `delete/api/v1/loginlog/clean`, which supports hard deletion of log records by time range.

#### Scenario: Clean up logs by timeframe
- **WHEN** Admin request `delete/api/v1/loginlog/clean? beginTime = 2026-01-01 & endTime = 2026-01-31`
- **THEN** Hard delete all login log records within this timeframe, returning the number of deleted records

#### Scenario: Clean up all login logs
- **WHEN** Admin request `delete/api/v1/loginlog/clean` (without time parameter)
- **THEN** hard delete all login log records

### Requirement: Login log batch deletion
The system SHALL provides the login log batch deletion interface `delete/api/v1/loginlog/{ids}`, which supports hard deletion of log records by ID list.

#### Scenario: Bulk delete by ID
- **WHEN** admin request `delete/api/v1/loginlog/1,2,3`
- **THEN** Hard delete the login log record with the specified ID, and return the number of deleted records

#### Scenario: frontend bulk delete operation
- **WHEN** Admin checks one or more login logs and clicks the "Delete" button
- **THEN** pops up a confirmation dialog box showing the selected number, confirming and performing batch deletion

### Requirement: Login Log Export
The system SHALL provides the login log export interface `get/api/v1/loginlog/export`, which is exported to the xlsx format according to the current filters.

#### Scenario: Export by filter
- **WHEN** Admin request `get/api/v1/loginlog/export? userName = admin&status = 0`
- **THEN** returns an xlsx file containing all eligible login logs

### Requirement: Login log frontend page
The system SHALL provides a login log management page under the frontend system monitoring menu through the `monitor-loginlog` source plugin.

#### Scenario: Login Log List Page
- **WHEN** Admin accesses login log page
- **THEN** displays the login log form, including the filter area (username, IP address, status, time range), table columns (username, IP address, browser, operating system, login result, prompt message, login time), toolbar (empty, export, delete buttons), line actions (view details button).The delete button is grayed out when no record is checked.

#### Scenario: View details popup
- **WHEN** Admin clicks "Details" button on a log
- **THEN** Open Modal with full login log information

#### Scenario: Cleanup actions
- **WHEN** Admin clicks "Clean" button
- **THEN** pops up a confirmation dialog box with a time range selector, confirmation and cleanup

#### Scenario: Export Actions
- **WHEN** Admin clicks "Export" button
- **THEN** Export the login log as an xlsx file with the current filters and download it

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

