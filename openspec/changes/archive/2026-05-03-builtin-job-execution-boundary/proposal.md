## Why

The current scheduled-job startup path uses `sys_job` as both the governance projection for built-in jobs and the execution registration source. Built-in jobs are first declared by host code or plugin metadata, synchronized into the table, and then scanned again by the persistent scheduler, which can register them twice. The boundary between built-in jobs and user-defined jobs must be explicit so startup database work and duplicate-registration patches are reduced, while the behavior better matches the architecture: built-in capabilities are declaration-driven, and tables provide governance views.

## What Changes

- Limit the execution source for built-in jobs to host code definitions or plugin cron declarations; `sys_job.is_builtin=1` is only for console display, log linkage, audit snapshots, and governance projection.
- Make persistent scheduler startup loading scan and register only user-defined jobs where `is_builtin=0 AND status=enabled`.
- Make built-in job synchronization register runtime scheduler entries from code or plugin declarations while synchronizing or updating `sys_job` projection rows.
- Keep manual trigger support for built-in jobs, but make the execution definition come from code or plugin declarations rather than mutable management-table fields.
- Keep backend and console protections that prevent enabling, disabling, editing, resetting, or deleting built-in job definitions.
- When a plugin is disabled or uninstalled, retain governance projections for plugin built-in jobs, unregister scheduler entries, and project the job into a non-runnable state.
- Remove the startup duplicate-registration patch and reposition the related tests around the new boundary.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cron-jobs`: Adjust startup scheduler registration rules and distinguish the real execution source of user-defined jobs and built-in jobs.
- `cron-job-management`: Clarify built-in job projection, manual trigger, status display, and non-editable boundaries.
- `cron-handler-registry`: Clarify plugin cron declarations as the execution source for plugin built-in jobs, including scheduler entry unregistering and projection state after plugin disable or uninstall.

## Impact

- Backend scheduler: startup registration, refresh, trigger, and tests under `apps/lina-core/internal/service/jobmgmt/internal/scheduler/`.
- Built-in job synchronization: host built-in jobs, plugin built-in job projection, and registration paths under `apps/lina-core/internal/service/cron/`.
- Job management service: built-in job state protection, manual trigger, log linkage, and i18n projection under `apps/lina-core/internal/service/jobmgmt/`.
- Plugin cron integration: managed cron job lifecycle for source-plugin and dynamic-plugin declarations.
- Frontend console: built-in job action visibility, manual trigger entry point, and read-only display under `apps/lina-vben/apps/web-antd/src/views/system/job/`.
- Tests: backend scheduler, job management, plugin lifecycle, and required E2E coverage for startup and console behavior.
