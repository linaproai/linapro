## ADDED Requirements

### Requirement: Built-in cron job display metadata must be localized by the backend
The system SHALL return display names, descriptions, and remarks for built-in scheduler data in cron job management, job group management, and execution log APIs according to the current request language. Built-in cron jobs registered by the host, source plugins, and dynamic plugins MUST use stable `handlerRef`, plugin ID, job name, or group code values as backend translation anchors. The frontend MUST NOT maintain job-name or group-name translation mappings based on Chinese source text, `handlerRef`, or group codes.

#### Scenario: Code-registered built-in jobs use English source copy
- **WHEN** the host or a plugin registers a built-in cron job
- **THEN** source copy in `Name`, `DisplayName`, and `Description` uses readable English
- **AND** Chinese and other non-English display values are returned through backend runtime i18n resources
- **AND** English display directly uses code-registered English source copy without duplicate translations in `en-US` runtime i18n JSON

#### Scenario: Job list returns localized names
- **WHEN** an administrator requests `GET /job` with `en-US`
- **THEN** built-in job `name`, `description`, and `groupName` values in the response have been projected to English
- **AND** user-created jobs keep their database values

#### Scenario: Job group query returns the localized default group
- **WHEN** an administrator requests `GET /job-group` with `en-US`
- **THEN** the default group's `name` and `remark` have been projected to English
- **AND** user-created groups continue to return database values

#### Scenario: Execution logs return localized job names
- **WHEN** an administrator requests `GET /job/log` with `en-US`
- **THEN** built-in job log `jobName` values have been projected to English

### Requirement: Manual job trigger must require confirmation

The Run Now action for scheduled jobs SHALL show a confirmation modal before triggering execution so administrators do not accidentally run operational tasks.

#### Scenario: Trigger action asks for confirmation
- **WHEN** an administrator clicks Run Now in the scheduled-job list
- **THEN** the frontend displays a confirmation modal
- **AND** no trigger API is called before confirmation

#### Scenario: Shell trigger remains available when shell editing is blocked
- **WHEN** a runnable shell job cannot be edited because of environment or permission limits
- **THEN** the row still shows a clickable Run Now action
