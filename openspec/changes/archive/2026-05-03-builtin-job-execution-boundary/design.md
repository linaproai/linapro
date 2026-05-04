## Context

At startup, `service/cron` currently synchronizes built-in job projections and then `jobmgmt/internal/scheduler.LoadAndRegister` scans `sys_job where status=enabled` and registers every job. This model makes `sys_job` both a governance projection and an execution source for built-in jobs, which creates duplicate registration paths for host and plugin built-ins and forces the scheduler to use an idempotency patch for same-name gcron entries.

The new boundary is: built-in jobs are executed from host code or plugin declarations. `sys_job.is_builtin=1` only stores console display data, log linkage, audit snapshots, and governance state. User-defined jobs continue to be executed from persistent `sys_job.is_builtin=0` records.

## Goals / Non-Goals

**Goals:**

- Move the built-in job execution source back to code and plugin declarations, avoiding reverse execution registration from `sys_job`.
- Make persistent scheduler startup loading handle only user-defined jobs.
- Preserve built-in job visibility, manual trigger support, and log linkage in the management console.
- Preserve non-runnable projection state when plugins are disabled or uninstalled, so disabled plugins no longer execute their jobs.
- Delete or reposition scheduler patches and tests that only existed for duplicate built-in registration.
- Keep SQL debug evidence for startup database work explainable and measurable.

**Non-Goals:**

- Do not allow the console to modify cron expressions, timeouts, concurrency policies, status, or execution definitions for built-in jobs.
- Do not introduce new database tables or external scheduler dependencies.
- Do not change CRUD, enable/disable, manual trigger, or shell-job permission semantics for user-defined jobs.
- Do not change `sys_job_log` as the unified execution log table.

## Decisions

### Decision 1: Built-in jobs are registered directly from code or plugin declarations

`cron.syncBuiltinScheduledJobs` SHALL synchronize the `sys_job` projection first, then use the resulting `sys_job.id` to register the corresponding gcron entry. Cron, scope, concurrency, timeout, handler reference, and other execution fields SHALL come from the current in-memory model converted from code or plugin declarations, not from a later scan of built-in rows in `sys_job`.

The alternative was to let `LoadAndRegister` continue scanning all `enabled` jobs and use `gcron.Remove` or same-name overwrite for idempotency. That approach keeps the wrong boundary because built-in execution is still driven backward from the table and startup still needs extra queries and duplicate-registration protection.

### Decision 2: The persistent scheduler only loads user-defined jobs

`jobmgmt/internal/scheduler.LoadAndRegister` SHALL query `is_builtin=0 AND status=enabled`. CRUD, enable/disable, and edit refresh paths continue to serve user-defined jobs. Runtime registration for built-in jobs is owned by the built-in job synchronization path in the cron component.

This keeps the persistent scheduler responsibility clear: restore user jobs from user-owned persistent data. Built-in jobs belong to host and plugin runtime capabilities.

### Decision 3: Manual trigger remains available, but uses declaration-owned execution data

Manual trigger still enters through `sys_job.id`, so permissions, confirmation modals, log linkage, and audit snapshots can be reused. When triggering a built-in job, the backend MUST use the current handler registry and built-in projection to verify executability. If the plugin handler is unavailable or the job is `paused_by_plugin`, the trigger MUST return handler-unavailable semantics.

The implementation may still read `sys_job` as the log snapshot source, but it must not let console writes change built-in execution definitions. If stricter declaration consistency is needed, trigger execution parameters can be rebuilt from the current built-in declaration index by `handler_ref`.

### Decision 4: Plugin lifecycle controls plugin built-in scheduler entries

When a plugin is enabled, the system SHALL discover and register declared plugin cron handlers and scheduler entries. When a plugin is disabled or uninstalled, the system SHALL unregister all cron handlers and scheduler entries for that plugin and project related `sys_job` rows as `paused_by_plugin` or `plugin_unavailable`. Historical logs and task projections remain for administrator audit.

### Decision 5: The console keeps built-in definitions read-only and still allows runnable triggers

The frontend continues to hide or disable edit, delete, enable/disable, reset, and other definition-changing actions for built-in jobs. Run Now remains available for executable built-in jobs. For `paused_by_plugin` rows or unavailable handlers, the frontend must not show a clickable Run Now action.

## Risks / Trade-offs

- [Risk] Built-in job registration needs `sys_job.id` for log linkage. Mitigation: upsert the projection first and return or reload the projection before runtime registration.
- [Risk] Plugin lifecycle and cron registration order can create a short unavailable window. Mitigation: keep lifecycle callbacks in the same request chain and complete handler registration, projection sync, and scheduler refresh together.
- [Risk] Manual trigger code that reads `sys_job` can be misread as table-driven execution. Mitigation: specs and code comments clarify that `sys_job` is only a governance projection and log snapshot, while executability is determined by the handler registry and declaration index.
- [Risk] Removing the `gcron.Remove` startup idempotency patch exposes real duplicate registrations again. Mitigation: reposition duplicate-registration tests to verify that built-in jobs do not enter `LoadAndRegister`, while preserving explicit remove-then-register behavior for CRUD refresh.

## Migration Plan

1. Adjust the startup loading filter so `LoadAndRegister` only loads user-defined jobs.
2. Extend built-in job synchronization to return projection rows or registration input, ensuring every built-in job has a stable `sys_job.id` during registration.
3. Register host and plugin built-in scheduler entries in the built-in job synchronization path.
4. Adjust plugin disable and uninstall paths so plugin built-in scheduler entries are unregistered and projections become unavailable.
5. Update backend and frontend tests for non-editable built-in jobs, enable/disable protections, and manual trigger behavior.
6. Verify startup logs to confirm built-in jobs are no longer registered through `LoadAndRegister`.

Rollback strategy: if declaration-driven built-in registration fails, temporarily restore the old `LoadAndRegister` behavior for built-in jobs, but keep duplicate-registration protection and re-evaluate startup query count.

## Open Questions

None. The user confirmed that built-in jobs may be manually triggered, and the remaining boundaries follow this design.
