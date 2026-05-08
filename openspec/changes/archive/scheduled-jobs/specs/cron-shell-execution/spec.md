## ADDED Requirements

### Requirement: Shell Task Global Switch

The system SHALL control whether Shell-type tasks can be created, modified, or executed through system parameter `cron.shell.enabled`, defaulting to enabled.

#### Scenario: Switch off rejects creation

- **WHEN** `cron.shell.enabled=false` and a user attempts to create a `task_type=shell` task
- **THEN** the system SHALL reject creation and return a clear error

#### Scenario: Switch off rejects modification

- **WHEN** `cron.shell.enabled=false` and a user attempts to modify any field of an existing Shell task
- **THEN** the system SHALL reject the modification

#### Scenario: Switch off rejects execution

- **WHEN** `cron.shell.enabled=false` and a Shell task reaches tick or is manually triggered
- **THEN** the system SHALL reject this execution
- **AND** the log SHALL show `status=failed` with `err_msg` indicating shell switch is disabled

#### Scenario: UI hides shell option

- **WHEN** the frontend reads `cron.shell.enabled=false` from the system parameters API
- **THEN** the UI SHALL hide the `shell` option in task type selection
- **AND** existing shell tasks SHALL only allow read-only viewing

#### Scenario: Windows runtime forced disable

- **WHEN** the host runs on Windows platform
- **THEN** the system SHALL treat `cron.shell.enabled` as false
- **AND** the UI SHALL display "Current platform does not support shell mode"

### Requirement: Shell Task Independent Permission Point

The system SHALL provide an independent permission point `system:job:shell` for Shell task management and triggering, separate from normal task CRUD permissions.

#### Scenario: Create Shell task permission

- **WHEN** a user lacks `system:job:shell` permission and attempts to create a `task_type=shell` task
- **THEN** the system SHALL return a permission error
- **AND** the UI SHALL hide the shell option in task type selection

#### Scenario: Modify/manual-trigger Shell task permission

- **WHEN** a user lacks `system:job:shell` permission and attempts to modify or manually trigger an existing Shell task
- **THEN** the system SHALL reject the operation

#### Scenario: Manual terminate Shell task permission

- **WHEN** a user lacks `system:job:shell` permission and attempts to terminate a running Shell instance
- **THEN** the system SHALL reject the operation

#### Scenario: Default permission assignment

- **WHEN** the system initializes seed data
- **THEN** the `system:job:shell` permission SHALL only be assigned by default to the built-in `admin` role

### Requirement: Shell Execution Context

The system SHALL provide Shell tasks with four dimensions of execution context: `shell_cmd / work_dir / env / timeout_seconds`, ensuring predictable behavior.

#### Scenario: Environment variable storage boundary

- **WHEN** a Shell task persists `env`
- **THEN** the system SHALL store the KV collection as plaintext JSON in `sys_job.env`
- **AND** the UI SHALL mask existing values when editing an existing Shell task
- **AND** audit logs SHALL NOT record raw `env` payloads

#### Scenario: Default shell interpreter

- **WHEN** the system executes a Shell task
- **THEN** the system SHALL use `/bin/sh -c <shell_cmd>` as the fixed launch command
- **AND** support multi-line scripts (UI uses multi-line text input)

#### Scenario: Working directory

- **WHEN** the task `work_dir` is non-empty
- **THEN** the system SHALL validate that the directory exists and the process has access before launching the subprocess
- **AND** the subprocess SHALL use that directory as CWD
- **AND** when `work_dir` is empty, the host process's current CWD SHALL be used

#### Scenario: Environment variables

- **WHEN** the task `env` is non-empty (KV JSON)
- **THEN** the system SHALL use the host process environment variables as base
- **AND** task-level `env` SHALL override same-named process-level variables
- **AND** the final merged result SHALL be passed to the subprocess

#### Scenario: Timeout required

- **WHEN** creating or modifying a Shell task
- **THEN** the system SHALL require `timeout_seconds in [1, 86400]`
- **AND** reject save on missing or out-of-range values

### Requirement: Shell Process Lifecycle

The system SHALL ensure Shell task process groups are properly managed, supporting timeout and manual termination without leaving orphan processes.

#### Scenario: Independent process group launch

- **WHEN** launching a Shell subprocess
- **THEN** the system SHALL set `SysProcAttr.Setpgid=true` on Unix platforms
- **AND** record PGID for subsequent termination

#### Scenario: Timeout forced termination

- **WHEN** execution time exceeds `timeout_seconds`
- **THEN** the system SHALL `kill -- -<pgid>` to terminate the entire process group
- **AND** the log SHALL show `status=timeout`
- **AND** `err_msg` SHALL record the timeout duration

#### Scenario: Manual termination

- **WHEN** an administrator calls `POST /job/log/{logId}/cancel` to terminate a running Shell instance
- **THEN** the system SHALL send SIGTERM to the process group, and if not exited within 5 seconds, send SIGKILL
- **AND** the log SHALL show `status=cancelled`

#### Scenario: Normal exit

- **WHEN** the Shell subprocess exits normally
- **THEN** the system SHALL determine result by `exit_code`: 0 = `success`, non-0 = `failed`
- **AND** `err_msg` SHALL contain a tail summary of stderr (on failure)

### Requirement: Shell Output Capture and Truncation

The system SHALL capture Shell subprocess stdout and stderr, truncating to preserve the prefix when exceeding limits.

#### Scenario: Truncation policy

- **WHEN** the subprocess produces stdout or stderr
- **THEN** the system SHALL retain the first 64KB for each
- **AND** overflow SHALL be discarded with `...[truncated]` marker appended

#### Scenario: Log storage

- **WHEN** the Shell task finishes execution
- **THEN** the system SHALL write `stdout / stderr / exit_code` to `sys_job_log.result_json`
- **AND** SHALL NOT create additional log tables

### Requirement: Shell Operation Audit

The system SHALL record Shell task sensitive operations in the operation log, supporting post-hoc accountability.

#### Scenario: Reuse host audit middleware

- **WHEN** creating, modifying, manually triggering, or manually terminating a Shell task through the host HTTP API
- **THEN** the system SHALL reuse the existing host `OperLog` middleware to write to `oper_log`
- **AND** for the same request, SHALL write exactly one semantically equivalent audit record
- **AND** the implementation SHALL NOT additionally hand-write a second duplicate `oper_log` record

#### Scenario: Create/modify audit

- **WHEN** creating or modifying a Shell task
- **THEN** the system SHALL write one record to `oper_log`
- **AND** record operator, IP, operation type, shell_cmd snapshot, work_dir, timeout_seconds
- **AND** SHALL NOT record raw `env` values

#### Scenario: Manual trigger audit

- **WHEN** an administrator manually triggers a Shell task execution
- **THEN** the system SHALL write a trigger record to `oper_log`
- **AND** associate it with the generated `sys_job_log` record ID

#### Scenario: Manual termination audit

- **WHEN** an administrator terminates a running Shell instance
- **THEN** the system SHALL write a termination record to `oper_log`
- **AND** record the terminated log_id and target task ID
- **AND** SHALL NOT record raw `env` values
