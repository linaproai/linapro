## ADDED Requirements

### Requirement: Built-in cron job display metadata must be localized by the backend
The system SHALL return display names, descriptions, and remarks for built-in scheduler data in cron job management, job group management, and execution log APIs according to the current request language. Built-in cron jobs registered by the host, source plugins, and dynamic plugins MUST use stable `handlerRef`, plugin ID, job name, or group code values as backend translation anchors. The frontend MUST NOT maintain job-name or group-name translation mappings based on Chinese source text, `handlerRef`, or group codes.

#### Scenario: Code-registered built-in jobs use English source copy
- **WHEN** the host or a plugin registers a built-in cron job, job handler, or dynamic plugin cron contract
- **THEN** source copy in `Name`, `DisplayName`, and `Description` uses readable English
- **AND** Chinese and other non-English display values are returned through backend runtime i18n resources or plugin i18n resources
- **AND** English display directly uses code-registered English source copy and does not require duplicate translations in `en-US` runtime i18n JSON
- **AND** when the current language lacks a translation, the backend falls back to code-registered English source copy and MUST NOT implicitly mix in default-language Chinese copy
- **AND** `handlerRef`, plugin ID, job name, and group code remain stable and are not directly displayed as user-visible translation results

#### Scenario: Job list returns localized names
- **WHEN** an administrator requests `GET /job` with `en-US`
- **THEN** built-in job `name`, `description`, and `groupName` values in the response have been projected to English
- **AND** user-created jobs keep their database values for `name`, `description`, and user-created group name
- **AND** the frontend job list renders API response values directly and no longer calls frontend seed-job mapping helpers

#### Scenario: Job group query returns the localized default group
- **WHEN** an administrator requests `GET /job-group` with `en-US`
- **THEN** the default group's `name` and `remark` have been projected to English
- **AND** user-created groups continue to return database values
- **AND** group display remains consistent between the group list and job list

#### Scenario: Execution logs return localized job names
- **WHEN** an administrator requests `GET /job/log` or `GET /job/log/{id}` with `en-US`
- **THEN** built-in job log `jobName` values have been projected to English
- **AND** the backend can use stable `handlerRef` or job anchors from the log snapshot to resolve the current-language display value
- **AND** the frontend log list and detail views no longer parse snapshots and translate locally
