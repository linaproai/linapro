## ADDED Requirements

### Requirement: Scheduled Job Management Data Model

The system SHALL provide a user-manageable scheduled job data model with at least `sys_job_group`, `sys_job`, `sys_job_log` three tables, satisfying the following constraints.

#### Scenario: Flat group structure

- **WHEN** an administrator adds, modifies, or queries groups
- **THEN** groups SHALL be flat (no parent hierarchy)
- **AND** the system SHALL guarantee exactly one `is_default=1` default group that cannot be deleted

#### Scenario: Task migration on group deletion

- **WHEN** an administrator deletes a non-default group
- **THEN** the system SHALL migrate all tasks in that group to the default group before deletion
- **AND** the UI SHALL show a confirmation dialog explaining impact scope and irreversibility

#### Scenario: Task name unique within group

- **WHEN** creating or modifying a task name within the same group
- **THEN** the system SHALL reject names duplicate within that group
- **AND** different groups may have tasks with the same name

#### Scenario: Task type distinction

- **WHEN** creating a task specifying `task_type=handler`
- **THEN** the task SHALL only be writable through host or plugin source-code registration paths
- **AND** public UI/API SHALL reject direct creation or editing of `handler` type tasks
- **AND** source-code registered tasks SHALL still require `handler_ref` non-empty, `params` validated against handler's JSON Schema, `timeout_seconds in [1, 86400]`
- **AND** `shell_cmd / work_dir / env` SHALL be ignored on write

#### Scenario: Shell task fields

- **WHEN** creating a task specifying `task_type=shell`
- **THEN** the system SHALL require `shell_cmd` non-empty, `timeout_seconds in [1, 86400]`
- **AND** `handler_ref / params` SHALL be ignored on write
- **AND** `work_dir` may be empty; when non-empty, must be an existing directory accessible by the host process

### Requirement: Task Execution Timeout

The system SHALL store a shared `timeout_seconds` field for all task types, enforced uniformly by the scheduling executor.

#### Scenario: Validate timeout on create or modify

- **WHEN** an administrator creates or modifies any `handler` or `shell` task
- **THEN** the system SHALL require `timeout_seconds in [1, 86400]`
- **AND** reject save on missing or out-of-range values

#### Scenario: Handler task timeout

- **WHEN** `task_type=handler` task execution exceeds `timeout_seconds`
- **THEN** the system SHALL cancel the execution's `context`
- **AND** the log final status SHALL be `timeout`
- **AND** `err_msg` SHALL record the timeout duration

### Requirement: Scheduled Job CRUD Interface

The system SHALL provide RESTful CRUD for scheduled jobs, conforming to project HTTP method semantics.

#### Scenario: List query

- **WHEN** calling `GET /job?groupId=&status=&taskType=&keyword=&page=&pageSize=`
- **THEN** the system SHALL return a paginated list of matching tasks
- **AND** returned fields SHALL include task basic info, group name, handler display name (handler type), and most recent execution status

#### Scenario: Create

- **WHEN** calling `POST /job` with a schema-compliant task
- **THEN** the system SHALL persist the record
- **AND** the public create interface SHALL only accept `task_type=shell` user-created tasks
- **AND** if `status=enabled`, immediately register in the local node scheduler
- **AND** `executed_count=0`, `stop_reason=null`

#### Scenario: Update

- **WHEN** calling `PUT /job/{id}` to modify fields
- **THEN** the system SHALL only allow updating user-created tasks
- **AND** SHALL unregister the old record from the scheduler and re-register with new configuration
- **AND** preserve `executed_count` unchanged (unless explicitly calling reset)

#### Scenario: Strict validation of shared scheduling fields on add and edit

- **WHEN** an administrator submits `cron_expr`, `status`, `scope`, `concurrency`, `max_concurrency`, `max_executions`, `timeout_seconds`, `timezone` and other shared fields during add or edit
- **THEN** frontend and backend SHALL strictly validate field types, allowed values, numeric ranges, and required constraints
- **AND** `cron_expr` SHALL only accept 5-segment or 6-segment Cron text format, rejecting empty, overlong, or other segment counts
- **AND** 5-segment expressions SHALL have the system automatically prepend `#` as second placeholder at runtime; user-submitted 6-segment expressions SHALL use real second values, not `#`
- **AND** `status` write values SHALL only allow `enabled` and `disabled`; `paused_by_plugin` SHALL only be written by the system when plugin handler becomes unavailable
- **AND** validation failures SHALL return clear error messages identifying whether the expression format, timezone, enum value, or numeric range is invalid

#### Scenario: Batch delete

- **WHEN** calling `DELETE /job` with a list of ids
- **THEN** the system SHALL unregister these tasks from the scheduler
- **AND** reject deletion of `is_builtin=1` tasks with a clear error

### Requirement: Source-Code Registered Task Projection and Read-Only Governance

The system SHALL project host, source-plugin, and dynamic-plugin declared built-in scheduled jobs into unified scheduled job management, avoiding the split experience where "some tasks can only be seen in code, invisible in the management page."

#### Scenario: Host built-in tasks unified visibility

- **WHEN** the host service registers built-in scheduling capabilities such as `session-cleanup`, `server-monitor-collector`, `server-monitor-cleanup`, `access-topology-sync`, `runtime-param-sync`, `cleanup-job-logs` at startup
- **THEN** the system SHALL synchronize these tasks into `sys_job`
- **AND** administrators SHALL be able to view, search, see details, view execution logs, and manually trigger these tasks in the task management list

#### Scenario: Plugin built-in tasks unified visibility

- **WHEN** a plugin declares scheduled jobs through the host's public scheduled job registration path
- **THEN** the system SHALL synchronize that task into `sys_job`
- **AND** list and detail SHALL identify its source as `Plugin Built-in`

#### Scenario: Dynamic plugin built-in tasks unified visibility

- **WHEN** a dynamic plugin declares independent scheduled job contracts in runtime artifacts
- **THEN** the host SHALL parse these declarations during install or enable path and synchronize into `sys_job`
- **AND** uniformly use `plugin:<pluginID>/cron:<name>` handler reference form to join the shared handler registry
- **AND** administrators SHALL be able to view and maintain these tasks in task management list, detail, and log pages

#### Scenario: Dynamic plugin not yet enabled remains visible but not executable

- **WHEN** a dynamic plugin's built-in tasks have been projected into `sys_job`, but the corresponding plugin is not yet enabled or has been disabled/uninstalled
- **THEN** the system SHALL mark the task status as `paused_by_plugin`
- **AND** the list SHALL continue to retain the task and historical logs
- **AND** the frontend SHALL prohibit Run Now and enable-type operations and clearly indicate the plugin is currently unavailable

#### Scenario: Source-code registered tasks are read-only

- **WHEN** an administrator views an `is_builtin=1` source-code registered task
- **THEN** frontend and backend SHALL treat it as read-only
- **AND** the administrator may only view details, view logs, and trigger once
- **AND** deletion, editing, enable/disable, or resetting execution count through public UI/API SHALL NOT be allowed

#### Scenario: Source-code registered task expressions preserved as-is

- **WHEN** the system synchronizes source-code registered task scheduling expressions
- **THEN** it SHALL preserve the expression text as registered by the host
- **AND** if the source task uses GoFrame scheduler natively supported `@every ...` expressions, public create/edit validation rules are not affected
- **AND** this capability is only for source-code registered task projection, not opened for user-created task forms

#### Scenario: Source labels replace task type as primary display

- **WHEN** an administrator views the task management list
- **THEN** the frontend SHALL use `Host Built-in / Plugin Built-in / User Created` as primary source labels
- **AND** SHALL NOT use `Handler / Shell` as the primary list display column
- **AND** users can still view the task's real execution type and handler reference in details

### Requirement: Built-in Job Projection-Only Execution Boundary

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

### Requirement: Scheduled Job Navigation and Explanatory Copy

The system SHALL present scheduled job capabilities in the management console with a more understandable navigation structure and copy, reducing first-use cost.

#### Scenario: System Management menu grouping

- **WHEN** an administrator views the System Management menu
- **THEN** the frontend SHALL provide a directory menu named "Scheduled Jobs" under System Management
- **AND** Task Management, Group Management, Execution Log SHALL appear as child menus under that directory

#### Scenario: Menu icons globally unique

- **WHEN** the system initializes the left navigation menu, or an administrator subsequently adds/edits directory and menu icons in menu management
- **THEN** all directory and menu icons in the left navigation SHALL be globally unique and non-repeating
- **AND** the Scheduled Jobs directory, Task Management, Group Management, Execution Log icons SHALL all be different
- **AND** System Management, System Monitoring, System Info and other existing built-in menus SHALL also use non-repeating icons
- **AND** menu management save SHALL reject submission and show a clear error when the icon duplicates an existing directory or menu icon

#### Scenario: Directory entry defaults to Task Management

- **WHEN** a user accesses the Scheduled Jobs directory entry route
- **THEN** the frontend SHALL default to the Task Management page
- **AND** existing `/system/job`, `/system/job-group`, `/system/job-log` page routes SHALL remain compatible and accessible

#### Scenario: Plugin unavailable status explanation

- **WHEN** the task list returns `status=paused_by_plugin`
- **THEN** the frontend SHALL use clear Chinese status copy indicating "Plugin handler unavailable"
- **AND** SHALL explain through tooltip that this status means the task's dependent plugin handler is currently unregistered, disabled, or uninstalled

#### Scenario: Cron expression help copy

- **WHEN** a user views the "Scheduled Expression" field during add or edit
- **THEN** the frontend SHALL clearly state support for both 5-segment and 6-segment Cron expressions
- **AND** SHALL explain that 5-segment expressions are automatically padded with `#` second placeholder to 6 segments before being submitted to the scheduler

#### Scenario: Scheduling scope and concurrency policy Chinese display

- **WHEN** the task list renders `scope` and `concurrency` fields
- **THEN** the frontend SHALL use understandable Chinese labels for scheduling scope and concurrency policy
- **AND** SHALL NOT directly display `master_only / all_node / singleton / parallel` raw English values
- **AND** the Chinese label for `all_node` SHALL be "All Nodes Execute"

#### Scenario: Shared scheduling field help tooltips

- **WHEN** a user views `Scheduling Scope`, `Concurrency Policy`, `Log Retention` fields during add or edit
- **THEN** the frontend SHALL provide help tooltip entry points next to field labels
- **AND** help copy SHALL explain each option's meaning, differences, and applicable scenarios
- **AND** overly long explanations SHALL use explicit line breaks for segmented display, avoiding hard-to-read long lines in tooltip overlays

#### Scenario: Follow-system log retention explanation

- **WHEN** a user views the "Follow System" option in the Log Retention field
- **THEN** the frontend SHALL display the current system-level log retention policy explanation
- **AND** SHALL explain that this task will follow the system default policy for cleanup when no override is set

#### Scenario: Scheduled expression field styling

- **WHEN** a user inputs a scheduled expression during add or edit
- **THEN** the frontend SHALL use a single-line style closer to a code input box for this field
- **AND** SHALL at least provide monospace font, code-box visual styling, and a more expression-editing-friendly input experience
- **AND** SHALL NOT require segmented syntax highlighting of expression fragments in add or edit pages

#### Scenario: Scheduled expression code-style display in task list

- **WHEN** the task management list renders the `cron_expr` column
- **THEN** the frontend SHALL display scheduled expressions using inline code style
- **AND** SHALL maintain a simple, stable code-block visual style without segmented highlighting of `* / # / digits`
- **AND** SHALL NOT change the raw expression text content returned by the API

#### Scenario: Empty action dropdown not displayed

- **WHEN** a task list row has no secondary action items
- **THEN** the frontend SHALL NOT display the "More" button for that row
- **AND** SHALL NOT render a dropdown action entry containing only an empty menu

#### Scenario: Timeout and max executions help tooltips

- **WHEN** a user views "Timeout (seconds)" and "Max Executions" fields during add or edit
- **THEN** the frontend SHALL provide help tooltip entry points next to field labels
- **AND** help copy SHALL explain the timeout's effect, the impact after task timeout, and the max execution limit rules
- **AND** `Max Executions=0` SHALL be explicitly stated as "No execution limit"

#### Scenario: Timezone field supports common options and custom input

- **WHEN** a user configures timezone during add or edit
- **THEN** the frontend SHALL provide searchable common timezone dropdown options
- **AND** SHALL default to the host's current system timezone
- **AND** SHALL allow users to enter custom timezone strings, e.g., `Asia/Shanghai` or `UTC`

#### Scenario: Shell task warning prompt and form spacing

- **WHEN** a user switches to the Shell task configuration tab during add or edit
- **THEN** the frontend SHALL continue to display "Shell tasks execute directly on the host node. Please strictly control command content and environment variables." warning
- **AND** the vertical spacing between that prompt block and the form area below SHALL be no less than `5px`

### Requirement: Scheduled Job Enable and Disable

The system SHALL allow administrators to toggle task start/stop status and ensure scheduler registry consistency with database state.

#### Scenario: List status column remains read-only display

- **WHEN** an administrator views task status in the task list
- **THEN** the Status column SHALL only be responsible for displaying current status
- **AND** SHALL NOT directly bind enable/disable toggle interaction to the status label

#### Scenario: Modify enable/disable through editing task

- **WHEN** an administrator needs to modify a user-created task's enable or disable status
- **THEN** the frontend SHALL complete the modification through the Task Status field in the edit dialog
- **AND** the list page SHALL NOT additionally provide status column click-toggle capability
- **AND** source-code registered tasks SHALL remain read-only detail, not opening status modification

#### Scenario: Enable task

- **WHEN** calling `PUT /job/{id}/status` setting `status` to `enabled`
- **THEN** the system SHALL register the task in the scheduler
- **AND** reset `stop_reason` to `null`

#### Scenario: Disable task

- **WHEN** `status` is set to `disabled`
- **THEN** the system SHALL unregister the task from the scheduler
- **AND** running instances SHALL NOT be forcibly terminated, but no new ticks SHALL be accepted

#### Scenario: Plugin unavailable blocks enable

- **WHEN** attempting to enable a task whose `handler_ref` source plugin is disabled/uninstalled
- **THEN** the system SHALL reject the enable operation with a clear error

### Requirement: Execution Count Policy

The system SHALL support "exit after N executions" policy, automatically disabling the task upon reaching the limit.

#### Scenario: Default unlimited execution

- **WHEN** task `max_executions=0`
- **THEN** the system SHALL have no execution count limit
- **AND** scheduled trigger executions SHALL still continue to increment `executed_count`

#### Scenario: Scheduled trigger counts executions

- **WHEN** the scheduler actually starts an execution with `trigger=cron`
- **THEN** the system SHALL count that execution toward `executed_count`
- **AND** this counting rule SHALL be independent of whether `max_executions` is `0`

#### Scenario: Auto-disable on reaching limit

- **WHEN** task `max_executions=N` and `executed_count` reaches N
- **THEN** the system SHALL immediately set the task `status` to `disabled`
- **AND** write `stop_reason=max_executions_reached`
- **AND** unregister from the scheduler
- **AND** preserve the complete record on the corresponding execution log

#### Scenario: Reset execution count

- **WHEN** calling `POST /job/{id}/reset`
- **THEN** the system SHALL zero out `executed_count`
- **AND** clear `stop_reason`

#### Scenario: Manual trigger does not count

- **WHEN** starting an execution through the manual trigger interface
- **THEN** the system SHALL NOT count this execution toward `executed_count`
- **AND** the log SHALL show `trigger=manual`

### Requirement: Cluster Scheduling Scope

The system SHALL support per-job configuration of `scope in {master_only, all_node}`, executing according to scope semantics during scheduling.

#### Scenario: master-only job skipped on non-primary node

- **WHEN** `scope=master_only` and current node `cluster.IsPrimary()=false`
- **THEN** the system SHALL skip this execution
- **AND** write a log with `status=skipped_not_primary`

#### Scenario: all-node job executes independently per node

- **WHEN** `scope=all_node`
- **THEN** each running node SHALL execute independently
- **AND** the log `node_id` field SHALL record the execution node

### Requirement: Concurrency Policy

The system SHALL support `concurrency in {singleton, parallel}`, with `max_concurrency` hard cap for parallel.

#### Scenario: Singleton execution skip

- **WHEN** `concurrency=singleton` and this node already has an instance running
- **THEN** the system SHALL skip this tick
- **AND** write a log with `status=skipped_singleton`

#### Scenario: Parallel over-limit skip

- **WHEN** `concurrency=parallel` and this node running instance count >= `max_concurrency`
- **THEN** the system SHALL skip this tick
- **AND** write a log with `status=skipped_max_concurrency`

#### Scenario: Parallel within quota triggers

- **WHEN** `concurrency=parallel` and running instance count < `max_concurrency`
- **THEN** the system SHALL start a new execution instance, running in parallel with existing instances

### Requirement: Timezone Handling

The system SHALL store an independent `timezone` field per task, with the scheduler parsing `cron_expr` according to that timezone.

#### Scenario: Task triggers in declared timezone

- **WHEN** task `timezone=Asia/Shanghai` and `cron_expr` expresses daily 8 AM
- **THEN** the system SHALL trigger the task at 8 AM in UTC+8
- **AND** independent of the server's local timezone

#### Scenario: Timezone default value

- **WHEN** creating a task without specifying `timezone`
- **THEN** the system SHALL default to the server process timezone
- **AND** the UI form SHALL default to the server timezone

### Requirement: Manual Trigger and Manual Termination

The system SHALL provide immediate trigger of one execution and termination of running instances.

#### Scenario: Immediate trigger

- **WHEN** calling `POST /job/{id}/trigger`
- **THEN** the system SHALL bypass the scheduling clock and start one execution
- **AND** the log SHALL show `trigger=manual`
- **AND** this execution SHALL NOT count toward `executed_count`
- **AND** this execution SHALL still be subject to `concurrency / max_concurrency / scope / timeout` constraints

#### Scenario: Run Now action asks for confirmation

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

#### Scenario: Disabled task can still be manually triggered

- **WHEN** task currently has `status=disabled`
- **THEN** the administrator SHALL still be able to manually trigger the task via "Run Now"
- **AND** this operation SHALL NOT automatically switch the task status to `enabled`
- **AND** if the task is in `paused_by_plugin` or pre-run validation fails, the system SHALL continue to reject the trigger with a clear reason

#### Scenario: Shell trigger remains available when shell editing is blocked

- **WHEN** a shell job cannot be created or edited because of environment switches or missing shell permissions
- **THEN** rows for runnable shell jobs still show a clickable Run Now action
- **AND** clicking it shows the confirmation modal

#### Scenario: Built-in job manual trigger is allowed

- **WHEN** an administrator confirms Run Now for an `is_builtin=1` job whose handler is currently available
- **THEN** the backend SHALL execute the job through the current host code or plugin handler declaration
- **AND** the backend SHALL create a `sys_job_log` record linked to the projected `sys_job.id`
- **AND** the execution snapshot SHALL preserve the projected job metadata for audit display

#### Scenario: Built-in job manual trigger is blocked when handler unavailable

- **WHEN** an administrator confirms Run Now for an `is_builtin=1` plugin job whose handler is unavailable
- **THEN** the backend SHALL reject the trigger with handler-unavailable semantics
- **AND** the frontend SHALL not show a clickable Run Now action for `paused_by_plugin` rows

#### Scenario: Terminate running instance

- **WHEN** calling `POST /job/log/{logId}/cancel` and the target log `status=running`
- **THEN** the system SHALL send a cancellation signal to this execution's context
- **AND** for shell tasks, kill the process group
- **AND** the log final status SHALL be set to `cancelled`
- **AND** `end_at` SHALL record the cancellation time

#### Scenario: Terminate already-ended instance rejected

- **WHEN** the target log `status` is no longer `running`
- **THEN** the system SHALL return a clear error

### Requirement: Execution Log

The system SHALL record one execution log for each trigger (including skipped triggers), capturing execution snapshot and results.

#### Scenario: Clear all logs

- **WHEN** an administrator performs "Clear All Logs" on the execution log page
- **THEN** the backend SHALL safely delete all execution logs when no `jobId` condition is present
- **AND** SHALL NOT error due to ORM missing `WHERE` clause protection

#### Scenario: Batch delete selected logs

- **WHEN** an administrator selects multiple logs in the execution log list and performs batch delete
- **THEN** the frontend SHALL provide a batch delete entry consistent with Monitoring Management - Operation Log
- **AND** the backend SHALL only delete the selected execution log records
- **AND** after success, list and selection state SHALL refresh synchronously

#### Scenario: Log fields

- **WHEN** an execution produces a log
- **THEN** the log SHALL include `job_id / job_snapshot / node_id / trigger / params_snapshot / start_at / end_at / duration_ms / status / err_msg / result_json`

#### Scenario: Shell task output

- **WHEN** task `task_type=shell` finishes execution
- **THEN** the log `result_json` SHALL contain `stdout / stderr / exit_code`
- **AND** stdout and stderr each truncate at first 64KB, appending `...[truncated]` marker on overflow

#### Scenario: Handler task return

- **WHEN** task `task_type=handler` executes successfully and returns structured result
- **THEN** the log `result_json` SHALL serialize the handler return value

#### Scenario: Execution failure

- **WHEN** task execution throws an error
- **THEN** the log SHALL show `status=failed`
- **AND** `err_msg` SHALL record the error summary

### Requirement: Log Cleanup Policy

The system SHALL provide a global default cleanup policy with task-level overrides, executed periodically by a built-in system cron job. The built-in cleanup task SHALL be registered through host source code and projected into `sys_job` during service startup; delivery SQL SHALL NOT write initialization seed data for `host:cleanup-job-logs` into `sys_job`. The `sys_job` row for this built-in task SHALL be a governance projection for display, logs, and audit linkage, not the execution-definition source used by startup persistent loading.

#### Scenario: Global default policy

- **WHEN** the task `log_retention_override` is empty
- **THEN** the task log SHALL be cleaned according to system parameter `cron.log.retention`

#### Scenario: Task-level override

- **WHEN** the task `log_retention_override` is configured as `{mode: days, value: 60}` or `{mode: count, value: 500}`
- **THEN** the system SHALL clean that task's logs by the task-level policy, ignoring the global default

#### Scenario: No cleanup policy

- **WHEN** the policy has `mode=none`
- **THEN** the system SHALL NOT clean logs for that task

#### Scenario: Built-in cleanup task

- **WHEN** the host starts and synchronizes built-in jobs
- **THEN** it SHALL project `host:cleanup-job-logs` into `sys_job`
- **AND** the default `cron_expr` SHALL run daily at midnight
- **AND** the task SHALL have `is_builtin=1`
- **AND** delivery SQL SHALL NOT seed this job into `sys_job`
- **AND** the cleanup task runtime registration uses the host code definition rather than a later persistent scan of `sys_job`

### Requirement: Built-in Task Partial Read-Only

The system SHALL protect `is_builtin=1` task critical fields from modification while opening operational parameter adjustment.

#### Scenario: Modifiable fields

- **WHEN** an administrator modifies an `is_builtin=1` task's `cron_expr / timezone / status / timeout_seconds / max_executions / log_retention_override`
- **THEN** the system SHALL accept and apply the modification

#### Scenario: Locked fields

- **WHEN** requesting to modify any of `task_type / handler_ref / params / scope / concurrency / group_id / name`
- **THEN** the system SHALL reject the modification with a clear error

#### Scenario: Edit page locked prompt layout

- **WHEN** an administrator opens a built-in task's edit dialog
- **THEN** the frontend SHALL preserve clear top and bottom spacing for "Shared scheduling field locked" and "Handler reference and parameter locked" prompt blocks
- **AND** the vertical gap between prompt blocks and adjacent form areas SHALL be no less than `5px`

#### Scenario: Delete prohibited

- **WHEN** requesting to delete an `is_builtin=1` task
- **THEN** the system SHALL reject the operation

#### Scenario: Upgrade does not overwrite user changes

- **WHEN** new version seed SQL re-executes and the task already exists
- **THEN** the system SHALL only update locked fields (when `seed_version` is newer)
- **AND** SHALL NOT overwrite user-modified open fields

### Requirement: Permission and Audit

The system SHALL constrain task management operations through menu and button permissions, and audit sensitive operations.

#### Scenario: Menu and button permissions

- **WHEN** an administrator accesses task management related pages or operations
- **THEN** the system SHALL validate the following permissions: menu `system:job:list / system:jobgroup:list / system:joblog:list`
- **AND** task button permissions: `system:job:add / edit / remove / status / trigger / reset`
- **AND** group button permissions: `system:jobgroup:add / edit / remove`
- **AND** log button permissions: `system:joblog:remove / cancel`

#### Scenario: Shell combined permissions

- **WHEN** a user creates, modifies, or manually triggers a `task_type=shell` task
- **THEN** the system SHALL simultaneously validate the corresponding base task permission (`system:job:add / edit / trigger`) and additional permission `system:job:shell`

#### Scenario: Shell terminate combined permissions

- **WHEN** a user terminates a running `task_type=shell` instance
- **THEN** the system SHALL simultaneously validate `system:joblog:cancel` and additional permission `system:job:shell`

#### Scenario: Operation audit

- **WHEN** executing task create/modify/delete/enable-disable/manual-trigger/manual-terminate
- **THEN** the system SHALL write to `oper_log`, recording operator, operation type, task ID, and key field snapshots
