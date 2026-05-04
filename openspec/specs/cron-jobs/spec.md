# Cron Jobs Specification

## Purpose

Define host scheduled-job categories, master-only and all-node execution rules, and persistent scheduler registration boundaries for user-manageable jobs.

## Requirements

### Requirement: Scheduled job categories

The system SHALL support two scheduled-job categories: master-only jobs and all-node jobs. In single-node mode, master-only jobs execute directly on the current node. In cluster mode, master-only jobs execute only on the primary node.

#### Scenario: Define master-only job

- **WHEN** a scheduled job is registered as master-only
- **THEN** the job executes on the current node in single-node mode
- **AND** executes only on the primary node in cluster mode

#### Scenario: Define all-node job

- **WHEN** a scheduled job is registered as all-node
- **THEN** the job executes on every running node

### Requirement: Master-only jobs must check primary-node eligibility

The system SHALL decide whether the current node should execute a master-only job before running it.

#### Scenario: Cluster primary executes master-only job

- **WHEN** `cluster.enabled=true` and the current node is primary when a master-only job triggers
- **THEN** the job executes normally

#### Scenario: Cluster secondary skips master-only job

- **WHEN** `cluster.enabled=true` and the current node is secondary when a master-only job triggers
- **THEN** the job returns immediately
- **AND** no business logic is executed

#### Scenario: Single-node mode executes master-only job

- **WHEN** `cluster.enabled=false` and a master-only job triggers
- **THEN** the current node executes the job directly

### Requirement: Existing scheduled jobs use deployment-aware categories

The system SHALL apply unified scheduling rules to existing scheduled jobs according to deployment mode.

#### Scenario: Session Cleanup is master-only

- **WHEN** the Session Cleanup scheduled job is registered
- **THEN** the system marks it as master-only
- **AND** it executes on the current node in single-node mode

#### Scenario: Server Monitor Collector is all-node

- **WHEN** the Server Monitor Collector scheduled job is registered
- **THEN** the system marks it as all-node
- **AND** it executes on every node

#### Scenario: Server Monitor Cleanup is master-only

- **WHEN** the Server Monitor Cleanup scheduled job is registered
- **THEN** the system marks it as master-only
- **AND** it executes on the current node in single-node mode

### Requirement: Scheduling-scope consistency for user-manageable jobs

The system SHALL apply the existing master-only and all-node scheduling semantics consistently to user-manageable scheduled jobs. User-manageable jobs are jobs with `sys_job.is_builtin=0`. Built-in jobs MUST also follow the same scope execution rules, but their execution definition comes from host code or plugin declarations rather than from `sys_job` records.

#### Scenario: User creates a master-only job

- **WHEN** a user creates a scheduled job with `scope=master_only`
- **THEN** the job SHALL execute only on the primary node in cluster mode
- **AND** execute directly on the current node in single-node mode

#### Scenario: User creates an all-node job

- **WHEN** a user creates a scheduled job with `scope=all_node`
- **THEN** the job SHALL execute once on every running node
- **AND** each execution SHALL record the triggering node in `sys_job_log.node_id`

#### Scenario: Non-primary master-only skip record

- **WHEN** `cluster.enabled=true` and a `scope=master_only` job is triggered on a non-primary node
- **THEN** the system SHALL return immediately without executing business logic
- **AND** write a `sys_job_log` record with `status=skipped_not_primary`

#### Scenario: Built-in jobs follow the unified scope rules

- **WHEN** a built-in job declared by host code or a plugin is registered with the scheduler
- **THEN** the system SHALL apply master-only or all-node execution rules according to the declared `scope`
- **AND** the `sys_job.is_builtin=1` projection row MUST NOT become the execution-definition source for that built-in job

### Requirement: Scheduler registration for user-manageable jobs

The system SHALL maintain the gcron registry during startup and CRUD operations so it stays consistent with user-defined jobs in `sys_job` where `status=enabled` and `is_builtin=0`. Built-in jobs with `sys_job.is_builtin=1` MUST NOT be registered by persistent scheduler startup scanning. Built-in jobs SHALL be registered by host code or plugin declaration synchronization paths.

#### Scenario: Startup loading

- **WHEN** the host process starts and `service/cron` starts
- **THEN** the system SHALL scan `sys_job where status=enabled and is_builtin=0`
- **AND** register each user-defined job into gcron with its `scope / concurrency / timezone / cron_expr`
- **AND** MUST NOT register `is_builtin=1` built-in jobs through that persistent scan

#### Scenario: Dynamic CRUD refresh

- **WHEN** a user-defined job is created, updated, or deleted
- **THEN** the system SHALL atomically unregister the old gcron entry and register the new entry when applicable
- **AND** hold a per-job mutex during refresh to avoid races with scheduler ticks

#### Scenario: Enable/disable refresh

- **WHEN** a user-defined job `status` changes from `disabled` to `enabled` or the reverse
- **THEN** the system SHALL register or unregister that job in the scheduler accordingly

#### Scenario: Built-in jobs do not participate in persistent loading

- **WHEN** `sys_job` contains a built-in job projection with `is_builtin=1 and status=enabled`
- **AND** the host process starts and runs persistent scheduler loading
- **THEN** the persistent scheduler MUST NOT use that record as a registration source
- **AND** the built-in job may only be registered by the corresponding host code definition or plugin cron declaration
