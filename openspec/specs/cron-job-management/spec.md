# Cron Job Management

## Purpose

Define scheduled job management behavior, including log cleanup policy governance, built-in job projection, manual trigger confirmation, and shell-job action availability.

## Requirements

### Requirement: Log Cleanup Policy

The system SHALL provide a global default cleanup policy with task-level overrides, executed periodically by a built-in system cron job. The built-in cleanup task SHALL be registered through host source code and projected into `sys_job` during service startup; delivery SQL SHALL NOT write initialization seed data for `host:cleanup-job-logs` into `sys_job`.

#### Scenario: Global default policy
- **WHEN** the task `log_retention_override` is empty
- **THEN** task logs are cleaned according to system parameter `cron.log.retention`

#### Scenario: Task-level override
- **WHEN** the task `log_retention_override` is configured as `{mode: days, value: 60}` or `{mode: count, value: 500}`
- **THEN** the system cleans that task's logs by the task-level policy and ignores the global default

#### Scenario: No cleanup policy
- **WHEN** the policy has `mode=none`
- **THEN** the system does not clean logs for that task

#### Scenario: Built-in cleanup task
- **WHEN** the host starts and synchronizes built-in jobs
- **THEN** it projects `host:cleanup-job-logs` into `sys_job`
- **AND** the default `cron_expr` runs daily at midnight
- **AND** the task has `is_builtin=1`
- **AND** delivery SQL does not seed this job into `sys_job`

### Requirement: Manual job trigger must require confirmation

The Run Now action for scheduled jobs SHALL show a confirmation modal before triggering execution so administrators do not accidentally run operational tasks.

#### Scenario: Trigger action asks for confirmation
- **WHEN** an administrator clicks Run Now in the scheduled-job list
- **THEN** the frontend displays a confirmation modal
- **AND** no trigger API is called before the administrator confirms

#### Scenario: Trigger confirmation uses execution styling
- **WHEN** the confirmation modal is shown for Run Now
- **THEN** it reuses the existing confirmation component pattern
- **AND** the confirm action uses execution-specific styling and wording rather than delete styling

#### Scenario: Canceling trigger does nothing
- **WHEN** an administrator cancels the Run Now confirmation
- **THEN** the frontend does not call `POST /job/{id}/trigger`
- **AND** the job list state remains unchanged

#### Scenario: Shell trigger remains available when shell editing is blocked
- **WHEN** a shell job cannot be created or edited because of environment switches or missing shell permissions
- **THEN** rows for runnable shell jobs still show a clickable Run Now action
- **AND** clicking it shows the confirmation modal
- **AND** the action column shows only one edit entry

#### Scenario: Shell jobs are enabled by default
- **WHEN** the system initializes `cron.shell.enabled` or falls back because the parameter is missing
- **THEN** the default value is `true`
- **AND** platform-level unsupported-shell protection can still disable shell jobs safely

