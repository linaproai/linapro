## Why

After verifying `temp/codebase-review-report.md` against the current code, several confirmed gaps remain in host correctness, security, database performance, and observability. These items are not covered by the active `backend-hardcoded-chinese-i18n-governance`, `plugin-api-query-performance`, or `regression-feedback-localization-ui` changes, and they have not been captured in capability specs. A focused host hardening iteration is needed so future feature work and performance work are built on a stable host baseline.

## What Changes

### Data consistency and security

- User deletion (`internal/service/user/user.go` `Delete`) must soft-delete the user, clean organization associations, and clean user-role associations inside one transaction. Swallowed cleanup `Warningf` calls become returned errors that roll back the transaction. `NotifyAccessTopologyChanged` runs after transaction commit.
- Role deletion (`internal/service/role/role.go` `Delete`) currently logs and continues when role-menu or user-role association cleanup fails inside the transaction. Those failures must return errors and roll back the transaction.
- Role user assignment (`internal/service/role/role.go` `AssignUsers`) must use one transaction plus batch insert instead of per-row inserts with swallowed `Warningf` failures.
- Menu deletion (`internal/service/menu/menu.go`) must return errors for role-menu cleanup failures inside the transaction instead of logging and continuing.
- Upload file access route `GET /api/v1/uploads/*` must be declared by the file API/controller module, mounted under the protected route group, and guarded by unified Auth plus Permission middleware. It must read files through the file service and storage backend, and must not directly concatenate local file paths in `cmd_http.go`.

### Database structure and performance

The project has no legacy burden, so SQL can be modified in place and verified by `make init`. All changes remain idempotent.

- `manifest/sql/008-menu-role-management.sql`: add `KEY idx_role_id (role_id)` to `sys_user_role` and `KEY idx_menu_id (menu_id)` to `sys_role_menu`.
- `manifest/sql/001-project-init.sql`: add `KEY idx_status (status)`, `KEY idx_phone (phone)`, and `KEY idx_created_at (created_at)` to `sys_user`.
- The SQL file containing `sys_online_session`: add `KEY idx_last_active_time (last_active_time)` for timeout cleanup.
- `manifest/sql/002-dict-dept-post.sql`: add `deleted_at DATETIME DEFAULT NULL` to `sys_dict_type` and `sys_dict_data`, letting GoFrame soft delete align those tables with other business tables; rerun `make dao`.
- `manifest/sql/014-scheduled-job-management.sql`: remove `CONSTRAINT fk_sys_job_group_id` and use `KEY idx_group_id (group_id)`, with application-level consistency.

### Batch operations and frontend performance

- Add RESTful batch delete endpoints `DELETE /api/v1/user?ids=...` and `DELETE /api/v1/role?ids=...`. DTO fields use `json` tags and English `dc` / `eg`; `g.Meta` carries the corresponding permission tag. Service `BatchDelete` methods reuse existing protection rules inside a single transaction.
- Change batch delete in `views/system/user/index.vue` and `views/system/role/index.vue` from loops over single-item delete APIs to one batch API call.
- Change menu `isDescendant` (`internal/service/menu/menu.go`) to load the parent-child map once and perform in-memory BFS, eliminating per-level SQL round trips.
- Add 30 second automatic refresh to the server monitor page (`views/monitor/server/index.vue`) with `useIntervalFn` plus page visibility awareness; polling pauses while the tab is hidden.
- Make user-message polling (`store/message.ts`) visibility-aware: pause the interval while hidden and refresh once immediately when visible again.
- Change router guard `loadedPaths` (`router/guard.ts`) to a bounded LRU with a default size of 50 to prevent unbounded growth in long-running SPA sessions.
- Keep public config sync and dict cache reset during language switching (`bootstrap.ts`), but stop reloading the full permission/menu/route state; menu titles must update through reactive `$t()`.

### Host runtime observability and operations

- Add public `GET /api/v1/health`: exposed through standard API/controller flow, runs a lightweight DB probe, returns `{status:"ok"}` when healthy, and returns 503 with a stable redacted reason when the DB is unavailable.
- Use GoFrame `Server.Run()` for built-in signal listening and HTTP graceful shutdown. `cmd_http.go` no longer registers `os/signal`; after `Server.Run()` returns, cleanup runs in order: Cron scheduler, cluster service, database pool, all bounded by `shutdown.timeout`.
- Split host foundation service interfaces by responsibility: `config.Service` becomes a composition of category reader/syncer interfaces; `middleware.Service` becomes a composition of HTTP middleware and non-middleware runtime support interfaces.
- Delete empty packages `apps/lina-core/pkg/auditi18n/` and `apps/lina-core/pkg/audittype/` to avoid implying that audit capability already exists. Real audit logging remains a separate future iteration.
- Replace the hard-coded `defaultManagedJobTimezone` in `internal/service/cron/cron_managed_jobs.go` with configuration key `scheduler.defaultTimezone`, defaulting to `UTC`.

### Out of scope

- Backend hard-coded Chinese cleanup is owned by `backend-hardcoded-chinese-i18n-governance`.
- English frontend copy/layout regressions and server-side default interval regressions are owned by `regression-feedback-localization-ui`.
- `GET /plugins` side effects and `sys_online_session.last_active_time` write throttling are owned by `plugin-api-query-performance`.
- Real audit logging, API rate limiting middleware, TraceID middleware, Vue global error boundary, and request cancellation infrastructure each require separate iterations.
- The `NewV1()` self-constructed service issue in `cmd_http.go`, multi-instance behavior, and DI containerization are intentionally left for a future host assembly refactor.

## Capabilities

### New Capabilities

- `host-runtime-operations`: general host runtime operations covering health probes, graceful shutdown, protected static-resource routing, and configurable runtime defaults, including scheduler default timezone and removal of stale placeholder packages.

### Modified Capabilities

- `user-management`: transactional deletion, batch delete endpoint and frontend integration, and key `sys_user` query indexes.
- `role-management`: deletion rollback for transaction failures, transactional `AssignUsers`, batch delete endpoint and frontend integration.
- `user-role-association`: add `role_id` reverse index to support role-based user queries.
- `menu-management`: rollback on menu deletion association cleanup failures, in-memory `isDescendant`, and `sys_role_menu` reverse index.
- `dict-management`: add `deleted_at` to dictionary type and data tables to align soft-delete behavior.
- `cron-job-management`: configurable default timezone and replacement of the `sys_job` foreign key with a reverse index.
- `server-monitor`: visibility-aware automatic frontend polling.
- `user-message`: visibility-aware unread-message polling.
- `online-user`: add `sys_online_session.last_active_time` index for session timeout cleanup.
- `framework-i18n-runtime-performance`: language switching no longer reloads full permissions, menus, and routes.

The existing `dict-management` spec already declares `deleted_at` for dictionary tables. This iteration only aligns SQL implementation with that existing spec, so it is not listed as a modified capability.

## Impact

- Code impact:
  - Backend: `internal/service/{user,role,menu,cron,config,middleware,file}/`, `internal/cmd/cmd_http*.go`, `internal/controller/{user,role,file,health}/`, `api/{user,role,file,health}/v1/`, `pkg/auditi18n`, and `pkg/audittype`.
  - SQL: `manifest/sql/001-project-init.sql`, `002-dict-dept-post.sql`, `008-menu-role-management.sql`, `014-scheduled-job-management.sql`, and the SQL file containing `sys_online_session`.
  - Config: `apps/lina-core/manifest/config/config.template.yaml` adds `scheduler.defaultTimezone`, `health.timeout`, and `shutdown.timeout`.
  - Frontend: `views/system/user/index.vue`, `views/system/role/index.vue`, `views/monitor/server/index.vue`, `store/message.ts`, `router/guard.ts`, `bootstrap.ts`, and `api/system/{user,role}/index.ts`.
- Operational impact: `/health` and graceful shutdown let Kubernetes and container orchestrators use standard probes and termination flows. Removing the foreign key reduces extra locking in high-concurrency scheduler paths.
- Test impact: add or extend unit tests for batch delete and transactional rollback; extend `hack/tests/e2e/` for batch delete frontend flow, anonymous health probe access, and server monitor auto-refresh.
- Dependency impact: no new third-party dependency; `@vueuse/core` is already available in the repository.
- API compatibility: batch delete endpoints are additive. Existing single-record `DELETE /api/v1/user/{id}` and `DELETE /api/v1/role/{id}` remain unchanged.
