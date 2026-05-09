## 1. Database and Dictionary

- [x] 1.1 Create `014-scheduled-job-management.sql` in `apps/lina-core/manifest/sql/` with `sys_job_group / sys_job / sys_job_log` three table DDLs (with indexes), foreign key constraints, idempotent `CREATE TABLE IF NOT EXISTS`
- [x] 1.2 In the same SQL, insert default group (`code=default, is_default=1`), system parameters `cron.shell.enabled=false` and `cron.log.retention={"mode":"days","value":30}`
- [x] 1.3 In the same SQL, insert menus (Task Management / Group Management / Execution Log) and button permissions (`system:job:add/edit/remove/status/trigger/reset`, `system:jobgroup:add/edit/remove`, `system:joblog:remove/cancel`, `system:job:shell`), and register menu permissions `system:job:list / system:jobgroup:list / system:joblog:list`
- [x] 1.4 In the same SQL, insert dictionary types and data: `cron_job_status / cron_job_task_type / cron_job_scope / cron_job_concurrency / cron_job_trigger / cron_job_log_status / cron_log_retention_mode`
- [x] 1.5 Execute `make init` to verify SQL idempotency and re-entrancy; confirm repeated execution has no errors
- [x] 1.6 Verify that delivery SQL does NOT seed built-in job rows into `sys_job`; built-in jobs are projected from code and plugin declarations at runtime

## 2. Backend DAO and API Skeleton

- [x] 2.1 Update configuration in `apps/lina-core/internal/cmd/` for relevant tables, then run `make dao` to generate `dao/sys_job.go`, `dao/sys_job_group.go`, `dao/sys_job_log.go` and DO/Entity
- [x] 2.2 Create DTOs in `api/job/v1/` split by interface purpose: `job_list.go / job_detail.go / job_create.go / job_update.go / job_delete.go / job_status.go / job_trigger.go / job_reset.go`
- [x] 2.3 Create group DTOs in `api/jobgroup/v1/`: `group_list.go / group_create.go / group_update.go / group_delete.go`
- [x] 2.4 Create log DTOs in `api/joblog/v1/`: `log_list.go / log_detail.go / log_cancel.go / log_clear.go`
- [x] 2.5 Create handler registry query DTOs in `api/jobhandler/v1/`: `handler_list.go / handler_detail.go`
- [x] 2.6 Add cron expression preview DTO in `api/job/v1/job_cron_preview.go` (input `expr / timezone`, output next 5 trigger times)
- [x] 2.7 All DTO `g.Meta` with `dc / permission` tags, all fields with `dc / eg / json`, meeting interface documentation standards; run `make ctrl` to generate controller skeletons

## 3. Backend Handler Registry

- [x] 3.1 Create `internal/service/jobhandler/jobhandler.go` (main file): define `Registry / HandlerDef / HandlerInfo / HandlerSource` interfaces and structs, `serviceImpl`, `New()`; main file carries component package comment
- [x] 3.2 Create `internal/service/jobhandler/jobhandler_host.go`: host built-in handler registration entry `RegisterHostHandlers()`, registering first batch of handlers under `host:xxx` naming (cleanup logs, cleanup session logs, regenerate session fingerprint, etc.)
- [x] 3.3 Create `internal/service/jobhandler/jobhandler_plugin.go`: define plugin lifecycle observer, synchronously calling `Register / Unregister` in `service/plugin` enable/disable/uninstall success paths
- [x] 3.4 Create `internal/service/jobhandler/jobhandler_schema.go`: wrap `ValidateParams(schemaText, paramsJSON) error` based on `JSON Schema draft-07` restricted scalar subset compatible validation library, explicitly rejecting keywords beyond this iteration's support scope
- [x] 3.5 Create `internal/service/jobhandler/jobhandler_test.go`: Register conflict, Lookup hit/miss, Unregister cascade notification callback

## 4. Backend Task Persistence and Scheduling Core

- [x] 4.1 Create `internal/service/jobmgmt/jobmgmt.go` (main file): define `Service / JobMgmt`, `New()`, base dependencies (DAO, handler registry, scheduler)
- [x] 4.2 Create `internal/service/jobmgmt/jobmgmt_group.go`: group CRUD (default group not deletable, tasks migrate to default group on deletion)
- [x] 4.3 Create `internal/service/jobmgmt/jobmgmt_job_crud.go`: task CRUD (unique within group, built-in task field lock validation, enable/disable coordination, notify scheduler to refresh on task change)
- [x] 4.4 Create `internal/service/jobmgmt/jobmgmt_job_status.go`: enable/disable/reset count; validate handler availability on enable
- [x] 4.5 Create `internal/service/jobmgmt/jobmgmt_log.go`: log query, clear, cleanup policy calculation (global + task-level override)
- [x] 4.6 Create `internal/service/jobmgmt/jobmgmt_cron_preview.go`: calculate next trigger time based on `robfig/cron` parser (parsed by timezone)
- [x] 4.7 Create `internal/service/jobmgmt/jobmgmt_test.go`: unit tests covering built-in task locking, group deletion migration, shared `timeout_seconds` validation, concurrency policy validation

## 5. Backend Scheduler Component (on top of gcron)

- [x] 5.1 Create `internal/service/jobmgmt/scheduler/scheduler.go` (main file): define `Scheduler` interface, `New()`, startup `LoadAndRegister(ctx)`, per-job mutex lock
- [x] 5.2 Create `internal/service/jobmgmt/scheduler/scheduler_register.go`: `Add / Remove / Refresh` wrapping gcron, with small LRU cache and proactive invalidation
- [x] 5.3 Create `internal/service/jobmgmt/scheduler/scheduler_runner.go`: tick wrapper `runJob(jobID, trigger)` — check scope / concurrency / max_executions / timeout_seconds, dispatch to handler or shell executor, capture result and write log
- [x] 5.4 Create `internal/service/jobmgmt/scheduler/scheduler_cancel.go`: maintain `runningInstances map[logID]cancelFn`, support `CancelLog(logID)`
- [x] 5.5 Create `internal/service/jobmgmt/scheduler/scheduler_test.go`: cover Add/Remove race conditions, scope guard skip, singleton/parallel counting, max_executions auto-disable
- [x] 5.6 Adjust `LoadAndRegister` to load only user-defined jobs with `is_builtin=0 AND status=enabled`
- [x] 5.7 Preserve `Refresh` remove-then-register semantics for user-defined job CRUD and enable/disable paths; confirm regular job dynamic refresh is unaffected
- [x] 5.8 Remove or reposition the `gcron.Remove(jobEntryName(job.Id))` patch in `registerJob` that only existed for startup duplicate registration of built-in jobs
- [x] 5.9 Adjust startup degradation for missing plugin handlers so only user-defined plugin-handler jobs loaded from persistence degrade there; plugin built-in jobs are projected as unavailable by plugin lifecycle paths

## 6. Backend Declaration-Driven Built-in Job Registration

- [x] 6.1 Modify built-in job synchronization in `apps/lina-core/internal/service/jobmgmt` so stable `sys_job.id` values are available after synchronization
- [x] 6.2 Modify `apps/lina-core/internal/service/cron` so host built-in jobs are registered directly from code definitions after projection sync and use projected `sys_job.id` for log linkage
- [x] 6.3 Modify plugin built-in job synchronization so source-plugin and dynamic-plugin cron declarations register scheduler entries after plugin enablement without depending on `LoadAndRegister` scanning `sys_job`
- [x] 6.4 Unregister all built-in scheduler entries for a plugin on disable or uninstall, and project related `sys_job` rows as `paused_by_plugin` or `plugin_unavailable`
- [x] 6.5 Confirm built-in job projections preserve log linkage, list display, detail display, i18n projection, and source markers

## 7. Backend Shell Executor

- [x] 7.1 Create `internal/service/jobmgmt/shellexec/shellexec.go` (main file): define `Executor` interface, `New()`, `cron.shell.enabled` and Windows platform guards
- [x] 7.2 Create `internal/service/jobmgmt/shellexec/shellexec_process.go`: `/bin/sh -c` launch subprocess, `Setpgid`, work_dir/env merge, stdout/stderr `LimitReader` 64KB truncation
- [x] 7.3 Create `internal/service/jobmgmt/shellexec/shellexec_lifecycle.go`: timeout `kill -- -<pgid>`, manual termination SIGTERM -> 5 seconds -> SIGKILL escalation path, avoid writing duplicate `oper_log`
- [x] 7.4 Create `internal/service/jobmgmt/shellexec/shellexec_test.go`: output truncation, timeout termination, manual cancellation

## 8. Backend Controller Implementation

- [x] 8.1 `controller/job/v1_new.go` initialize dependency fields (`jobmgmt.Service / jobhandler.Registry / scheduler.Scheduler`), `NewV1()` one-time injection
- [x] 8.2 `controller/job/v1_*.go` implement business by interface: list/detail/create/update/delete/enable-disable/trigger/reset/cron-preview
- [x] 8.3 `controller/jobgroup/v1_*.go` implement group CRUD
- [x] 8.4 `controller/joblog/v1_*.go` implement log query/detail/clear/cancel
- [x] 8.5 `controller/jobhandler/v1_*.go` implement handler list/detail
- [x] 8.6 Declare `g.Meta.permission` by resource: task interfaces use `system:job:*`, group interfaces use `system:jobgroup:*`, log interfaces use `system:joblog:*`; shell create/modify/trigger interfaces add `system:job:shell`, shell log cancel interface combines `system:joblog:cancel + system:job:shell`
- [x] 8.7 Declare `operLog` meta tags for shell create/modify/trigger/cancel interfaces and return necessary correlation identifiers (e.g., `log_id`), reusing host `OperLog` middleware for single audit record writing
- [x] 8.8 Extend host audit request parameter desensitization rules to mask `env` payloads in shell-related interfaces, ensuring `oper_log` does not record raw environment variable values
- [x] 8.9 Confirm `TriggerJob` allows manual trigger for `is_builtin=1` jobs and validates executability through the current handler registry or built-in declaration
- [x] 8.10 Confirm manual trigger for `paused_by_plugin` or handler-unavailable built-in jobs returns a stable business error
- [x] 8.11 Confirm the backend continues rejecting edit, delete, enable/disable, reset, and other execution-definition mutations for built-in jobs

## 9. Backend Integration into Startup Flow

- [x] 9.1 Modify `internal/service/cron/cron.go`: call `jobmgmtScheduler.LoadAndRegister(ctx)` at the end of `Start(ctx)`; inject through constructor, not create within tick
- [x] 9.2 Modify `internal/service/plugin` enable/disable/uninstall success paths to synchronously notify `jobhandler` observer through explicit lifecycle callbacks
- [x] 9.3 Modify `internal/service/config` to expose typed read entries for `cron.shell.enabled` and `cron.log.retention` and reuse existing runtime param refresh mechanism; `sysconfig` continues to handle parameter data management
- [x] 9.4 Complete dependency injection for `jobhandler / jobmgmt / scheduler / shellexec` at `cmd` startup assembly

## 10. Frontend API and Adapters

- [x] 10.1 Add `job.ts / jobGroup.ts / jobLog.ts / jobHandler.ts` in `apps/lina-vben/apps/web-antd/src/api/system/`, covering CRUD/trigger/cancel/log/cron-preview/handler-list interfaces
- [x] 10.2 Add handler dynamic parameter sub-form adapter in `src/adapter/form/`: generate Vben form `schema` array from `JSON Schema draft-07` restricted scalar subset (supporting string/integer/number/boolean/enum/date/date-time/textarea)
- [x] 10.3 Register task list, log list, group list column definitions in `src/adapter/vxe-table/`

## 11. Frontend Page: Group Management

- [x] 11.1 Create `src/views/system/job-group/index.vue`: `Page + useVbenVxeGrid` list + `GhostButton + Popconfirm` action column
- [x] 11.2 Create `src/views/system/job-group/modal.vue`: `useVbenModal + useVbenForm` add/edit dialog
- [x] 11.3 Default group row disables delete button, displays "Default Group" label

## 12. Frontend Page: Task Management

- [x] 12.1 Create `src/views/system/job/index.vue`: list page, top filters include group, status, task type, keyword
- [x] 12.2 Create `src/views/system/job/form.vue`: `useVbenModal` main dialog; top Tab switches handler / shell type (shell tab hidden when `cron.shell.enabled=false` or no `system:job:shell` permission)
- [x] 12.3 Create `src/views/system/job/form-handler.vue`: handler selection dropdown + dynamic parameter sub-form rendering (based on schema) + cron expression + timezone + scope + concurrency + max_concurrency + timeout_seconds + max_executions + log_retention_override
- [x] 12.4 Create `src/views/system/job/form-shell.vue`: shell_cmd multi-line textarea + work_dir + env (KV table, masks existing values in edit mode) + timeout_seconds + warning color prompt
- [x] 12.5 Action column: enable/disable, run now, edit, delete, reset count; built-in jobs only show partial buttons
- [x] 12.6 `paused_by_plugin` status explicitly highlighted in red with "Plugin handler unavailable" tooltip
- [x] 12.7 Built-in jobs hide or disable edit, delete, enable/disable, and reset; runnable built-ins keep Run Now; unavailable plugin built-ins hide or disable Run Now
- [x] 12.8 List status column remains read-only display; enable/disable changes through edit dialog task status field
- [x] 12.9 Add `Host Built-in / Plugin Built-in / User Created` source labels as primary display in list; stop using `Handler / Shell` as primary list column
- [x] 12.10 Run Now action shows confirmation modal before triggering; confirmation uses execution-specific styling
- [x] 12.11 Hide empty "More" button when a row has no secondary action items

## 13. Frontend Page: Execution Logs

- [x] 13.1 Create `src/views/system/job-log/index.vue`: log list, top filters include task, status, node, time range
- [x] 13.2 Create `src/views/system/job-log/detail.vue`: `useVbenModal` detail dialog, showing trigger / params_snapshot / result_json (shell tasks with stdout/stderr code highlighting)
- [x] 13.3 Running log rows show explicit "Terminate" button, calls cancel interface on confirmation
- [x] 13.4 Log list batch clear function (requires `system:joblog:remove` permission), terminate button requires `system:joblog:cancel`; terminating shell instance also requires `system:job:shell`

## 14. Frontend Navigation and Menu

- [x] 14.1 Register `/system/job / /system/job-group / /system/job-log` in `src/router/routes/modules/system.ts` or equivalent route module
- [x] 14.2 Menu icons use `IconifyIcon` (`ant-design:clock-circle-outlined` or similar icons)
- [x] 14.3 Button permissions correspond to `system:job:*`, `system:jobgroup:*`, `system:joblog:*`, shell create/modify/trigger additionally require `system:job:shell`
- [x] 14.4 Add "Scheduled Jobs" directory menu under System Management, with Task Management, Group Management, Execution Log as child menus
- [x] 14.5 Directory entry defaults to Task Management page; existing `/system/job`, `/system/job-group`, `/system/job-log` routes remain compatible
- [x] 14.6 All directory and menu icons must be globally unique and non-repeating; menu management save rejects duplicate icons
- [x] 14.7 "System Monitoring" directory icon adjusted to be more semantically appropriate for monitoring scenarios

## 15. E2E Test Cases

- [x] 15.1 TC0081 Scheduled job group CRUD (add, edit, delete non-default group, default group not deletable)
- [x] 15.2 TC0082 Handler type task CRUD + dynamic form rendering
- [x] 15.3 TC0083 Shell type task creation (with `cron.shell.enabled` prerequisite)
- [x] 15.4 TC0084 Shell global switch off: frontend hides shell type option, backend rejects write
- [x] 15.5 TC0085 Task enable/disable, status switch takes effect immediately
- [x] 15.6 TC0086 Manual trigger execution & log trigger=manual, not counted in executed_count
- [x] 15.7 TC0087 Long task manual termination -> log status=cancelled
- [x] 15.8 TC0088 `max_executions` reaching limit auto-disables + `stop_reason` display
- [x] 15.9 TC0089 Execution log list filtering, detail, clear
- [x] 15.10 TC0090 Plugin disable causing task paused_by_plugin highlighted red, enable button disabled
- [x] 15.11 TC0091 Built-in task cron modifiable, handler_ref locked, delete rejected
- [x] 15.12 TC0092 Deleting non-default group migrates tasks to default group
- [x] 15.13 TC0093 Timezone field persistence and next execution time preview
- [x] 15.14 TC0094 Shell task stdout/stderr truncation viewable
- [x] 15.15 TC0095 Handler task timeout log `status=timeout`, `err_msg` contains timeout duration
- [x] 15.16 TC0096 No `system:job:shell` permission prohibits terminating running shell instance
- [x] 15.17 TC0097 Navigation, old entry redirect, and help copy improvements
- [x] 15.18 TC0158 Built-in job execution boundary: read-only actions, Run Now for runnable built-ins, unavailable plugin built-in trigger-entry state
- [x] 15.19 All new test cases automatically run `pnpm test -- TC008x` / `pnpm test -- TC009x` / `pnpm test -- TC015x` during execution; test file naming strictly corresponds to TC IDs

## 16. Documentation and Closeout

- [x] 16.1 Update `apps/lina-core/README.md / README.zh-CN.md`: add "Scheduled Job Management" section in module list (Chinese and English synchronized)
- [x] 16.2 Update `apps/lina-vben/apps/web-antd/README.md / README.zh-CN.md`: add route and page entry description
- [x] 16.3 Verify `make init / make dao / make ctrl / make dev / make test` full flow passes
- [x] 16.4 Call `/lina-review` skill for code and specification review, handle all critical issues
- [x] 16.5 After user confirms feature completion, execute `/opsx:archive scheduled-jobs`

## Feedback

- [x] **FB-1**: Unified task management permission matrix and removed undefined export permissions
- [x] **FB-2**: Solidified `job / jobgroup / joblog / jobhandler` resource split and controller ownership
- [x] **FB-3**: Clarified handler `ParamsSchema` `JSON Schema draft-07` restricted scalar subset and validation boundary
- [x] **FB-4**: Unified shell audit to reuse host `OperLog` middleware and avoid duplicate `oper_log`
- [x] **FB-5**: Clarified `cron.shell.enabled` and `cron.log.retention` runtime read ownership
- [x] **FB-6**: Changed plugin handler coordination from loose event bus to explicit lifecycle callbacks
- [x] **FB-7**: Clarified `timeout_seconds` as shared field for all task types and supplemented test planning
- [x] **FB-8**: Clarified shell `env` audit desensitization boundary and added implementation task
- [x] **FB-9**: Clarified combined permissions for terminating shell instances and test coverage
- [x] **FB-10**: Changed admin user permission bypass and menu query logic from role-specific to user-specific, removed `sys_role_menu` super-admin dependency
- [x] **FB-11**: Unified SQL seed syntax, removed `ON DUPLICATE KEY UPDATE` and explicit auto-increment primary key writes; prioritized fixing scheduled job SQL, supplemented full-repo SQL audit task
- [x] **FB-12**: Removed redundant `sys_role_menu` writes in admin role menu seeds and plugin menu synchronization
- [x] **FB-13**: Audited and synchronized other SQL copies, removed residual admin role menu bindings and legacy explicit auto-increment `id` seed writes
- [x] **FB-14**: Moved built-in admin account strategy from `pkg` to `service/user` component internal, eliminated misleading cross-component public dependencies
- [x] **FB-15**: Supplemented new controller skeleton comments and fixed review omissions, ensuring untracked controller files also included in current review round
- [x] **FB-16**: Improved `lina-review` skill scope identification flow, included untracked files in specification review scope and prohibited relying solely on `git diff`
- [x] **FB-17**: Elevated SQL seed prohibition of `ON DUPLICATE KEY UPDATE` and explicit auto-increment `id` writes to project specification, synchronized review skill and active design documents
- [x] **FB-18**: Treated scheduled job scheduling capability as current iteration core feature, supplemented critical execution path unit tests and E2E coverage with pass verification
- [x] **FB-19**: Added "Scheduled Jobs" directory menu under System Management, maintaining existing page entry compatibility
- [x] **FB-20**: Clarified `paused_by_plugin` status meaning, improved task status filtering and list tooltip copy
- [x] **FB-21**: Clarified Cron expression support for both 5-segment and 6-segment, supplemented form help copy
- [x] **FB-22**: Changed task list scheduling scope and concurrency policy to understandable Chinese display
- [x] **FB-23**: Added question-mark tooltips for scheduling scope, concurrency policy, and log retention, displayed current system log retention policy in "Follow System" explanation
- [x] **FB-24**: Supplemented TC0097 E2E coverage for navigation, old entry redirect, and help copy improvements
- [x] **FB-25**: Fixed execution log "clear" triggering ORM delete protection error when no `jobId` condition present
- [x] **FB-26**: Supplemented batch delete entry for execution log list, reusing consistent multi-select delete interaction with operation log
- [x] **FB-27**: Changed 5-segment Cron expression normalized second placeholder from `0` to `#`, synchronized validation, help copy, and tests
- [x] **FB-28**: Changed "Cron Expression" label in add/edit dialog to "Scheduled Expression", eliminated field title line break issue
- [x] **FB-29**: Improved scheduled expression input to code-input style, evaluated highlight support boundary under current component capabilities
- [x] **FB-30**: Added segmented line breaks to scheduling scope, concurrency policy, and log retention help tooltips, improving long copy readability
- [x] **FB-31**: Unified `all_node` Chinese display text to `All Nodes Execute`
- [x] **FB-32**: Changed task timezone field to support common dropdown and custom input, defaulting to host current system timezone
- [x] **FB-33**: Allowed disabled tasks to still be manually "Run Now" from list for testing, while keeping `paused_by_plugin` restriction
- [x] **FB-34**: Changed task list scheduled expression column to code-highlight display for Cron readability
- [x] **FB-35**: Added help tooltips for "Timeout (seconds)" and "Max Executions", clarified `0` means no execution limit
- [x] **FB-36**: Adjusted vertical spacing between locked prompt blocks and form areas in built-in task edit page, ensuring no less than `5px`
- [x] **FB-37**: Removed unnecessary upgrade backfill menu/dictionary/role-association SQL from `014-scheduled-job-management.sql` under "new project with no legacy debt" premise
- [x] **FB-38**: Unified `014-scheduled-job-management.sql` scheduling scope, concurrency policy and other dictionary seed copy to match current final UI copy
- [x] **FB-39**: Moved scheduler and shell executor sub-components used internally by `jobmgmt` component into `internal` directory, preventing external direct references
- [x] **FB-40**: Strictly validated scheduled expression, timezone, status and shared scheduling fields in add and edit tasks; illegal input returns clear errors
- [x] **FB-41**: Changed scheduled expression input in add and edit pages to code-style input box without segmented code highlighting
- [x] **FB-42**: Adjusted vertical spacing between Shell task warning prompt block and form area, ensuring no less than `5px`
- [x] **FB-43**: Unified host service and plugin source-code registered scheduled jobs projection into `sys_job` with full display in task management
- [x] **FB-44**: Converged public task create/edit entry to only support Shell tasks; Handler tasks changed to source-code registered read-only display
- [x] **FB-45**: Added `Host Built-in / Plugin Built-in / User Created` source display in task list and details, tightened source-code registered tasks to read-only operations
- [x] **FB-46**: Adjusted System Management scheduled task directory and child menu icons, ensuring globally unique and non-repeating left menu icons
- [x] **FB-47**: Supplemented and updated scheduled job E2E coverage for source-code registered task visibility, Handler creation entry removal, and read-only behavior
- [x] **FB-48**: Fixed remaining E2E TypeScript compilation errors in the repository, ensuring related test page objects and cases pass `tsc --noEmit`
- [x] **FB-49**: Converged scheduled expression display in task list to simple code block style without segmented highlighting
- [x] **FB-50**: Fixed `executed_count` not incrementing after scheduled trigger execution, supplemented regression test
- [x] **FB-51**: Hid empty "More" button when task list row has no secondary actions
- [x] **FB-52**: Moved task enable/disable entry from action column to status column click toggle, supplemented interaction regression test
- [x] **FB-53**: Restored task list status column to read-only display, moved enable/disable back to edit dialog task status field
- [x] **FB-54**: Supplemented dynamic plugin scheduled job declaration, projection, and execution paths, ensuring plugin built-in jobs enter unified task management and support maintenance
- [x] **FB-55**: Organized guest-side runtime host service SDK into independent interface object entry consistent with `DataHostService`, unified calling method
- [x] **FB-56**: Clarified scheduled job governance contract should maintain independent boundary, not merge plugin scheduled job declaration directly into runtime host service
- [x] **FB-57**: Converged hard-coded enumerations and protocol strings in dynamic plugin scheduled jobs to shared bridge constants and helper functions
- [x] **FB-58**: Changed dynamic plugin scheduled jobs to register through independent cron host service code, removed YAML/custom section declaration paths
- [x] **FB-59**: Fixed plugin lifecycle sync mistakenly executing dynamic cron discovery for source plugins during scheduled job sync, avoiding unrelated plugin errors disrupting dynamic plugin install authorization flow
- [x] **FB-60**: Improved dynamic plugin authorization page host service copy and cron service display, unified with equal-length labels like "Data Table Name / Storage Path / Access URL", showed registered scheduled job name, expression, scheduling scope and concurrency policy in "Task Service" card
- [x] **FB-61**: Adjusted dynamic plugin authorization page host service sorting and cron card style, moved runtime service to bottom, removed "Scheduled Job" summary label, changed task property titles to bold display
- [x] **FB-62**: Unified plugin authorization page and details page host service label background semantics, unified details page "Current Effective Scope" text to "Effective Scope"
- [x] **FB-63**: Fixed and constrained left menu directory/menu icon global uniqueness, supplemented menu management save validation and navigation regression test
- [x] **FB-64**: Adjusted left menu "System Monitoring" directory icon to be more semantically appropriate for monitoring scenarios

## Verification

- `go test ./internal/service/jobhandler ./internal/service/jobmgmt/... ./internal/service/cron` passed.
- `go test -p 1 ./...` passed under `apps/lina-core`.
- `npx playwright test e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts` passed: 2 tests.
- Startup SQL before `http server started`: `startup_sql_statements=48`, `startup_select=26`, `startup_show=8`, `startup_insert=0`, `startup_update=0`, `startup_delete=0`, `startup_sys_job_writes=0`.
- Startup evidence: persistent scheduler query uses `FROM sys_job WHERE status='enabled' AND is_builtin=0`; built-in projection snapshot query remains `WHERE is_builtin=1` for display/log linkage only.
