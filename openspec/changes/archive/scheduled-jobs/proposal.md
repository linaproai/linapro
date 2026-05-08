## Why

The `apps/lina-core/internal/service/cron` component originally only hosted code-embedded governance tasks (session cleanup, monitoring collection, runtime parameter sync, etc.) with no user-facing entry point for scheduled job management. In practice, administrators need to create business-specific scheduled jobs (data cleanup, data reconciliation, external system health probes, etc.) with capabilities including: visual CRUD, group management, enable/disable control, execution logs, execution-count policies, cluster scheduling scope selection, concurrency policies, manual trigger and cancellation, and support for custom shell commands.

A critical architectural refinement clarifies the execution source boundary for built-in jobs. The startup path initially used `sys_job` as both the governance projection and the execution registration source for built-in jobs, which created duplicate registration paths and forced the scheduler to use an idempotency patch for same-name gcron entries. The refined boundary states: built-in jobs are executed from host code or plugin declarations; `sys_job.is_builtin=1` only stores console display data, log linkage, audit snapshots, and governance state. User-defined jobs continue to be executed from persistent `sys_job.is_builtin=0` records.

This change introduces a complete user-manageable scheduled job subsystem on top of the existing code-embedded cron capability, extends the handler registration mechanism to reuse host and plugin code capabilities, and enforces a clear execution-source boundary between built-in and user-defined jobs.

## What Changes

- **Data model**: New `sys_job_group`, `sys_job`, `sys_job_log` tables for user-defined scheduled jobs, groups, and execution logs.
- **Task types**: Two execution forms supported:
  - `handler` type: Calls a host- or plugin-registered named handler; parameters validated against a `JSON Schema draft-07` restricted scalar subset.
  - `shell` type: Executes user-defined `/bin/sh -c` multi-line scripts with mandatory timeout, working directory, environment variables, process group kill, and stdout/stderr truncation.
- **Scheduling semantics**: Each job independently configures `scope` (master-only / all-node), `concurrency` (singleton / parallel + max_concurrency), `timezone`, `cron_expr`, `timeout_seconds`, `max_executions`, and `log_retention`.
- **Handler registry**: New host-side handler registration and plugin handler subscription. Plugin disable/uninstall automatically pauses associated jobs with `paused_by_plugin` and highlights them in UI.
- **Built-in job execution boundary**: Built-in jobs are registered directly from host code definitions or plugin declarations. `sys_job.is_builtin=1` rows serve only as governance projections for display, log linkage, and audit snapshots. The persistent scheduler startup loading scans and registers only user-defined jobs where `is_builtin=0 AND status=enabled`. Delivery SQL does not write initialization seed data for built-in jobs into `sys_job`; projections are upserted from code and plugin declarations at startup.
- **Built-in job read-only governance**: Fields like `task_type`, `handler_ref`, `params`, `scope`, `concurrency`, `group_id`, `name` are locked for built-in jobs. `cron_expr`, `timezone`, `status`, `timeout_seconds`, `max_executions`, and `log_retention_override` remain modifiable by administrators. Manual trigger is available for executable built-in jobs but blocked for `paused_by_plugin` or handler-unavailable jobs.
- **Shell security governance**: New permission point `system:job:shell`, system parameter `cron.shell.enabled` global switch, and reuse of the host `oper_log` middleware for shell job create/modify/trigger/cancel audit logging. Audit preserves `shell_cmd / work_dir / timeout_seconds` snapshots but does not record raw `env` values.
- **Human interaction**: UI provides manual trigger once (`trigger=manual`, not counted in `executed_count`) and manual termination of running instances (`ctx.Cancel` + process group kill). Manual trigger shows a confirmation modal before execution.
- **Execution count policy**: When `max_executions>0`, reaching the limit auto-disables the job and records the reason in `stop_reason`; manual reset re-enables it.
- **Log cleanup**: Global default cleanup policy (retention by count or days) plus task-level overrides, executed by a built-in system cron job.
- **Frontend**: Built on Vben5 (`useVbenForm / useVbenModal / useVbenVxeGrid / GhostButton + Popconfirm / IconifyIcon`) with task management, group management, and execution log pages. Handler parameter area dynamically renders forms from schema. Scheduled job navigation uses a "Scheduled Jobs" directory menu under System Management.
- **E2E**: New Playwright test cases for task CRUD, groups, manual trigger/cancellation, shell switch, cluster scheduling scope, and built-in job execution boundary behavior.

## Capabilities

### New Capabilities
- `cron-job-management`: User-manageable scheduled job CRUD, groups, enable/disable, execution logs, count policies, manual trigger and cancellation, log cleanup policies, built-in job read-only governance, navigation structure, and explanatory copy.
- `cron-handler-registry`: Handler registry contract covering host registration, plugin subscription, JSON Schema parameter declaration, plugin lifecycle impact on job state, dynamic plugin cron host service declaration, and built-in job projection synchronization.
- `cron-shell-execution`: Security boundary for shell-type tasks, execution context, process lifecycle, output capture rules, global switch, independent permission, and audit logging.

### Modified Capabilities
- `cron-jobs`: Extends existing master-only/all-node scheduling semantics to user-manageable tasks; built-in jobs follow the same scope rules but their execution definition comes from code or plugin declarations rather than from `sys_job` records. Persistent scheduler startup loading is limited to user-defined jobs.

## Impact

- **Database**: New `manifest/sql/<NNN>-scheduled-job-management.sql` with three new table DDLs, seed data (default group, system parameters, permission points and menu items, dictionaries), and idempotent protection. Built-in job rows are NOT seeded into `sys_job`; they are upserted from code and plugin declarations at runtime.
- **Backend new modules**:
  - Existing `internal/service/cron` `Service` preserved and extended with registration capabilities; built-in job synchronization now registers runtime scheduler entries from code definitions and upserts `sys_job` projections.
  - New `internal/service/jobmgmt` (job/group/log domain service) and `internal/service/jobhandler` (handler registry), organized with sub-packages and main files per convention.
  - New `api/job/v1/*.go`, `api/jobgroup/v1/*.go`, `api/joblog/v1/*.go`, `api/jobhandler/v1/*.go` REST interfaces split by resource.
  - New `internal/controller/job/*`, `internal/controller/jobgroup/*`, `internal/controller/joblog/*`, `internal/controller/jobhandler/*` controller skeletons with implementations.
- **Backend modifications**:
  - `service/cron/cron.go` startup phase loads persistent jobs (`is_builtin=0 AND status=enabled`) and registers built-in jobs from code/plugin declarations via unified scheduler.
  - `service/plugin` enable/disable/uninstall paths synchronize handler availability through explicit lifecycle callbacks, unregister plugin built-in scheduler entries, and project jobs as `paused_by_plugin`.
  - `service/config` adds typed runtime read entries for `cron.shell.enabled` and `cron.log.retention`.
- **Frontend**: `apps/lina-vben/apps/web-antd` adds routes, views, APIs, adapters. Menu items and button permissions follow module decoupling principles. Built-in jobs display with read-only actions. Scheduled Jobs directory menu added under System Management.
- **Permissions**: Menu permissions `system:job:list / system:jobgroup:list / system:joblog:list`, button permissions `system:job:add/edit/remove/status/trigger/reset`, `system:jobgroup:add/edit/remove`, `system:joblog:remove/cancel`, and shell additional permission `system:job:shell`.
- **E2E**: New test cases under `hack/tests/e2e/system/job/` and `hack/tests/e2e/scheduler/job/` covering task/group CRUD, shell operations, plugin cascade, built-in job boundary, and navigation.
- **Dictionaries**: Task status, trigger type, result status, shell mode, cleanup policy type and other enumerations unified into the dictionary module.
