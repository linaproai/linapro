## ADDED Requirements

### Requirement: User-Created Cron Jobs Must Follow Host Data Permission

The system SHALL apply host data permission governance to `sys_job.is_builtin=0` user-created cron jobs. All-data scope can access all user-created tasks; department-data scope only accesses user-created tasks whose creator belongs to the current user's visible department scope; self-only scope only accesses user-created tasks where `created_by` equals the current user ID. `sys_job.is_builtin=1` built-in task projections are system governance data and SHALL maintain visibility, edit protection, and trigger boundaries controlled by existing built-in task rules, without filtering by role data permission.

#### Scenario: All-Data Scope Queries User-Created Tasks

- **WHEN** caller has all-data scope and queries the cron job list
- **THEN** the system returns all user-created tasks matching the query conditions
- **AND** built-in task projections are handled per existing built-in task display rules

#### Scenario: Department Scope Queries User-Created Tasks

- **WHEN** a normal user's role data scope is department data
- **AND** queries the cron job list
- **THEN** the system only returns user-created tasks whose creator belongs to the current user's visible department scope
- **AND** does not return tasks created by users from other departments

#### Scenario: Self-Only Scope Queries User-Created Tasks

- **WHEN** a normal user's role data scope is self-only data
- **AND** queries the cron job list
- **THEN** the system only returns user-created tasks where `created_by` equals the current user ID

#### Scenario: Built-in Tasks Not Filtered by Data Permission

- **WHEN** the cron job list contains `is_builtin=1` host or plugin built-in task projections
- **THEN** the system does not filter these built-in task projections because the current user's role data scope is self-only or department
- **AND** built-in tasks continue to control operability per function permission, handler availability, and built-in task protection rules

### Requirement: Cron Job Detail and Changes Must Validate Data Permission

The system SHALL verify whether the target task is within the current caller's data permission scope before reading, updating, deleting, enabling, disabling, resetting, or triggering a user-created cron job. When the target task is not within scope, the system SHALL return a structured data-not-visible error, and MUST NOT modify the task, scheduler registration state, or create execution logs.

#### Scenario: Reject Viewing Out-of-Scope Task Detail

- **WHEN** a normal user's role data scope is self-only data
- **AND** requests to view another user's created task detail
- **THEN** the system returns a structured data-not-visible error
- **AND** does not return the task detail

#### Scenario: Reject Editing Out-of-Scope Task

- **WHEN** a normal user's role data scope is department data
- **AND** the target task's creator is not within the current user's visible department scope
- **THEN** the system rejects edit, delete, enable, or disable for that task
- **AND** does not refresh that task's scheduler registration state

#### Scenario: Reject Triggering Out-of-Scope Task

- **WHEN** a normal user's role data scope is self-only data
- **AND** requests to manually trigger another user's created task
- **THEN** the system rejects the trigger
- **AND** does not create a new `sys_job_log` execution record

### Requirement: Cron Job Logs Must Inherit Task Data Permission Boundary

The system SHALL apply data permission boundaries consistent with the parent task to user-created task execution logs. The caller can only query, view, export, or terminate logs produced by tasks within their data permission scope. Built-in task logs continue to be controlled by built-in task governance and function permission.

#### Scenario: Department Scope Queries Task Logs

- **WHEN** a normal user's role data scope is department data
- **AND** queries the cron job log list
- **THEN** the system only returns logs produced by tasks created by users within the visible department scope

#### Scenario: Reject Terminating Out-of-Scope Task Log

- **WHEN** a normal user's role data scope is self-only data
- **AND** requests to terminate a running log of another user's created task
- **THEN** the system rejects the termination
- **AND** the target task execution state does not change due to this request
