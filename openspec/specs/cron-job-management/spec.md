# Cron Job Management

## Purpose

Define scheduled job management behavior, including log cleanup policy governance, built-in job projection, manual trigger confirmation, and shell-job action availability.

## Requirements

### Requirement: Log Cleanup Policy

The system SHALL provide a global default cleanup policy with task-level overrides, executed periodically by a built-in system cron job. The built-in cleanup task SHALL be registered through host source code and projected into `sys_job` during service startup; delivery SQL SHALL NOT write initialization seed data for `host:cleanup-job-logs` into `sys_job`. The `sys_job` row for this built-in task SHALL be a governance projection for display, logs, and audit linkage, not the execution-definition source used by startup persistent loading.

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
- **AND** the cleanup task runtime registration uses the host code definition rather than a later persistent scan of `sys_job`

### Requirement: Manual job trigger must require confirmation

The Run Now action for scheduled jobs SHALL show a confirmation modal before triggering execution so administrators do not accidentally run operational tasks. Built-in scheduled jobs SHALL remain manually triggerable when their current code-owned or plugin-owned handler is available, but the manual trigger MUST NOT imply that administrators can edit the built-in job execution definition stored in `sys_job`.

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

#### Scenario: Built-in job manual trigger is allowed
- **WHEN** an administrator confirms Run Now for an `is_builtin=1` job whose handler is currently available
- **THEN** the backend SHALL execute the job through the current host code or plugin handler declaration
- **AND** the backend SHALL create a `sys_job_log` record linked to the projected `sys_job.id`
- **AND** the execution snapshot SHALL preserve the projected job metadata for audit display

#### Scenario: Built-in job manual trigger is blocked when handler unavailable
- **WHEN** an administrator confirms Run Now for an `is_builtin=1` plugin job whose handler is unavailable
- **THEN** the backend SHALL reject the trigger with handler-unavailable semantics
- **AND** the frontend SHALL not show a clickable Run Now action for `paused_by_plugin` rows

### Requirement: Built-in job projection-only execution boundary

The system SHALL treat `sys_job.is_builtin=1` rows as governance projections for framework-owned or plugin-owned scheduled jobs. These rows SHALL support list display, detail display, localization projection, log association, and manual trigger lookup, but administrators MUST NOT modify their execution definition through scheduled-job management APIs or frontend forms.

#### Scenario: Built-in job projection is synchronized from declarations
- **WHEN** the host starts or plugin lifecycle synchronization runs
- **THEN** the system SHALL upsert built-in job projection rows from host code definitions and plugin cron declarations
- **AND** each projected row SHALL keep `is_builtin=1`
- **AND** delivery SQL SHALL NOT be the source of built-in job seed rows

#### Scenario: Built-in job definition changes are denied
- **WHEN** a caller attempts to edit, delete, enable, disable, reset, or otherwise mutate execution-definition fields of an `is_builtin=1` job through scheduled-job management APIs
- **THEN** the backend SHALL reject the request with a stable business error
- **AND** the frontend SHALL hide or disable those mutation actions for built-in rows

#### Scenario: Built-in job projection remains visible
- **WHEN** an administrator opens scheduled-job management
- **THEN** built-in rows SHALL remain visible with localized display metadata
- **AND** their source SHALL indicate whether they are host built-ins or plugin built-ins

### Requirement: Scheduled job default timezone must be configurable

The system SHALL read the default timezone for built-in cron jobs from configuration key `scheduler.defaultTimezone`, defaulting to `UTC`. Source code MUST NOT keep hard-coded constants such as `defaultManagedJobTimezone = "Asia/Shanghai"`.

#### Scenario: Missing configuration uses UTC

- **WHEN** the configuration file does not declare `scheduler.defaultTimezone`
- **AND** the service starts and registers built-in jobs
- **THEN** built-in jobs MUST use `UTC` as the default timezone

#### Scenario: Custom timezone takes effect

- **WHEN** the configuration file sets `scheduler.defaultTimezone: "Asia/Shanghai"`
- **AND** the service starts and registers built-in jobs
- **THEN** built-in jobs MUST use `Asia/Shanghai` as the default timezone

### Requirement: sys_job table must not use foreign key constraints

The system SHALL remove the `fk_sys_job_group_id` foreign key constraint from the `sys_job` table, maintain `group_id` to `sys_job_group` reference consistency in the application layer, and keep `KEY idx_group_id (group_id)` on `sys_job` for group-based query and cleanup paths. Other association tables in this repository rely on application-level consistency, and this table MUST follow that convention to avoid extra foreign-key lock overhead in high-concurrency scheduler scenarios.

#### Scenario: sys_job table no longer contains foreign key constraints

- **WHEN** `make init` completes database initialization
- **THEN** the `sys_job` table MUST NOT contain `fk_sys_job_group_id` or any `FOREIGN KEY` constraint pointing to `sys_job_group`

#### Scenario: sys_job keeps group_id index

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_job` MUST include `idx_group_id` on column `group_id`

#### Scenario: Write path validates group_id reference consistency

- **WHEN** an upper-layer caller creates or updates a `sys_job` record with `group_id`
- **THEN** the service layer MUST validate that the referenced group exists
- **AND** validation failure MUST return a `bizerr` business error instead of relying on database foreign-key interception
