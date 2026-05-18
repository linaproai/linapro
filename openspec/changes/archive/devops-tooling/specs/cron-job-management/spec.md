## MODIFIED Requirements

### Requirement: Log Cleanup Policy

The system SHALL provide a global default cleanup policy with task-level overrides, executed periodically by a built-in system cron job. The built-in cleanup task SHALL be registered through host source code and projected into `sys_job` during service startup; delivery SQL SHALL NOT write initialization seed data for `host:cleanup-job-logs` into `sys_job`.

#### Scenario: Global Default Policy

- **WHEN** the task `log_retention_override` is empty
- **THEN** the task logs SHALL be cleaned up according to the system parameter `cron.log.retention` policy

#### Scenario: Task-level Override

- **WHEN** the task `log_retention_override` is configured as `{mode: days, value: 60}` or `{mode: count, value: 500}`
- **THEN** the system SHALL clean up that task's logs according to the task-level policy, ignoring the global default

#### Scenario: No Cleanup Policy

- **WHEN** the policy `mode=none`
- **THEN** the system SHALL NOT clean up logs for that task

#### Scenario: Built-in Cleanup Task

- **WHEN** the host service starts and executes built-in task synchronization
- **THEN** the system SHALL project the `host:cleanup-job-logs` handler task into `sys_job`
- **AND** the default `cron_expr` triggers daily at midnight
- **AND** the task has `is_builtin=1`
- **AND** delivery SQL does not contain initialization seed data for this task in `sys_job`
