## 1. Backend scheduler boundary adjustment

- [x] 1.1 Modify `apps/lina-core/internal/service/jobmgmt/internal/scheduler` so `LoadAndRegister` loads only user-defined jobs with `is_builtin=0 AND status=enabled`
- [x] 1.2 Preserve `Refresh` remove-then-register semantics for user-defined job CRUD and enable/disable paths, and confirm regular job dynamic refresh is unaffected
- [x] 1.3 Remove or reposition the `gcron.Remove(jobEntryName(job.Id))` patch in `registerJob` that only existed for startup duplicate registration of built-in jobs
- [x] 1.4 Adjust startup degradation for missing plugin handlers so only user-defined plugin-handler jobs loaded from persistence degrade there; plugin built-in jobs are projected as unavailable by plugin lifecycle paths

## 2. Declaration-driven built-in job registration

- [x] 2.1 Modify built-in job synchronization in `apps/lina-core/internal/service/jobmgmt` so stable `sys_job.id` values are available after synchronization
- [x] 2.2 Modify `apps/lina-core/internal/service/cron` so host built-in jobs are registered directly from code definitions after projection sync and use projected `sys_job.id` for log linkage
- [x] 2.3 Modify plugin built-in job synchronization so source-plugin and dynamic-plugin cron declarations register scheduler entries after plugin enablement without depending on `LoadAndRegister` scanning `sys_job`
- [x] 2.4 Unregister all built-in scheduler entries for a plugin on disable or uninstall, and project related `sys_job` rows as `paused_by_plugin` or `plugin_unavailable`
- [x] 2.5 Confirm built-in job projections preserve log linkage, list display, detail display, i18n projection, and source markers

## 3. Manual trigger and management protections

- [x] 3.1 Confirm `TriggerJob` allows manual trigger for `is_builtin=1` jobs and validates executability through the current handler registry or built-in declaration
- [x] 3.2 Confirm manual trigger for `paused_by_plugin` or handler-unavailable built-in jobs returns a stable business error
- [x] 3.3 Confirm the backend continues rejecting edit, delete, enable/disable, reset, and other execution-definition mutations for built-in jobs
- [x] 3.4 Check the frontend `system/job` page: built-in jobs hide or disable edit, delete, enable/disable, and reset; runnable built-ins keep Run Now; unavailable plugin built-ins hide or disable Run Now
- [x] 3.5 Evaluate i18n impact; this change did not add or modify runtime UI/API copy, menus, buttons, status labels, or caller-visible backend error codes, so no i18n JSON resources were required

## 4. Backend tests

- [x] 4.1 Update scheduler unit tests so `LoadAndRegister` does not register `is_builtin=1` jobs and only registers enabled user-defined jobs
- [x] 4.2 Update or delete duplicate-registration tests so they verify built-in jobs no longer enter persistent scanning instead of depending on `registerJob` same-name overwrite
- [x] 4.3 Add host built-in job synchronization tests covering scheduler entry registration after sync and execution-log linkage through projected `sys_job.id`
- [x] 4.4 Add plugin lifecycle tests covering plugin enable registration and plugin disable/uninstall unregistering plus `paused_by_plugin` projection
- [x] 4.5 Add manual trigger tests covering runnable built-in trigger allowed, unavailable plugin built-in trigger rejected, and user-defined job behavior unchanged

## 5. E2E and startup log verification

- [x] 5.1 Create `hack/tests/e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts` to verify built-in job read-only actions, Run Now for runnable built-ins, and unavailable plugin built-in trigger-entry state
- [x] 5.2 Run `npx playwright test e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts`; 2 subtests passed
- [x] 5.3 Run `go test -p 1 ./...`; full backend tests passed
- [x] 5.4 Run `openspec validate builtin-job-execution-boundary --strict`
- [x] 5.5 Start services and analyze SQL debug in `temp/lina-core.log` before `http server started`; confirm no startup duplicate-registration writes for built-in jobs and confirm `LoadAndRegister` does not scan/register `is_builtin=1` jobs

## 6. Review closeout

- [x] 6.1 Run `/lina-review` for this change, covering GoFrame constraints, OpenSpec compliance, i18n impact, frontend interaction, and test results
- [x] 6.2 Fix review findings and rerun affected tests
- [x] 6.3 Update final verification records and prepare for archive confirmation

## Verification

- `go test ./internal/service/jobhandler ./internal/service/jobmgmt/... ./internal/service/cron` passed.
- `go test -p 1 ./...` passed under `apps/lina-core`.
- `openspec validate builtin-job-execution-boundary --strict` passed.
- `npx playwright test e2e/scheduler/job/TC0158-builtin-job-execution-boundary.ts` passed: 2 tests.
- Startup SQL before `http server started`: `startup_sql_statements=48`, `startup_select=26`, `startup_show=8`, `startup_insert=0`, `startup_update=0`, `startup_delete=0`, `startup_sys_job_writes=0`.
- Startup evidence: persistent scheduler query uses `FROM sys_job WHERE status='enabled' AND is_builtin=0`; built-in projection snapshot query remains `WHERE is_builtin=1` for display/log linkage only.
