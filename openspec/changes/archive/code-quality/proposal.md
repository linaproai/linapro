## Why

The LinaPro backend accumulated systematic deviations across GoFrame v2 ORM conventions, REST API contract consistency, production `panic` discipline, transactional correctness, SQL performance, API documentation quality, module decoupling, and host runtime operability. These deviations degraded development velocity, introduced runtime risks (unnecessary panics, swallowed transaction errors, full table scans), and created friction during OpenSpec validation and archival. A consolidated code-quality hardening iteration was needed to establish a stable backend baseline before further feature work.

The project is new and has no legacy burden. SQL can be modified in place and verified by `make init`. Internal function signatures, call chains, and tests can be adjusted directly without backward compatibility concerns.

## What Changes

### Data consistency and security

- User deletion (`internal/service/user/user.go` `Delete`) must soft-delete the user, clean organization associations, and clean user-role associations inside one transaction. Swallowed cleanup `Warningf` calls become returned errors that roll back the transaction. `NotifyAccessTopologyChanged` runs after transaction commit.
- Role deletion (`internal/service/role/role.go` `Delete`) currently logs and continues when role-menu or user-role association cleanup fails inside the transaction. Those failures must return errors and roll back the transaction.
- Role user assignment (`internal/service/role/role.go` `AssignUsers`) must use one transaction plus batch insert instead of per-row inserts with swallowed `Warningf` failures.
- Menu deletion (`internal/service/menu/menu.go`) must return errors for role-menu cleanup failures inside the transaction instead of logging and continuing.
- Upload file access route `GET /api/v1/uploads/*` must be declared by the file API/controller module, mounted under the protected route group, and guarded by unified Auth plus Permission middleware. It must read files through the file service and storage backend, and must not directly concatenate local file paths in `cmd_http.go`.

### Database structure and performance

- Add query indexes: `idx_status`, `idx_phone`, `idx_created_at` on `sys_user`; `idx_role_id` on `sys_user_role`; `idx_menu_id` on `sys_role_menu`; `idx_last_active_time` on `sys_online_session`.
- Remove `sys_job` foreign key constraint `fk_sys_job_group_id`, replacing with application-level consistency and `KEY idx_group_id`.
- Add `deleted_at DATETIME DEFAULT NULL` to `sys_dict_type` and `sys_dict_data` to align dictionary tables with other business tables.
- Rewrite menu `isDescendant` from per-level SQL to in-memory BFS traversal.
- All SQL changes remain idempotent.

### Batch operations and frontend performance

- Add RESTful batch delete endpoints `DELETE /api/v1/user?ids=...` and `DELETE /api/v1/role?ids=...`. DTO fields use `json` tags and English `dc` / `eg`; `g.Meta` carries the corresponding permission tag. Service `BatchDelete` methods reuse existing protection rules inside a single transaction.
- Change batch delete in `views/system/user/index.vue` and `views/system/role/index.vue` from loops over single-item delete APIs to one batch API call.
- Add 30 second automatic refresh to the server monitor page with `useIntervalFn` plus page visibility awareness; polling pauses while the tab is hidden.
- Make user-message polling visibility-aware: pause the interval while hidden and refresh once immediately when visible again.
- Change router guard `loadedPaths` to a bounded LRU with a default size of 50 to prevent unbounded growth in long-running SPA sessions.
- Keep public config sync and dict cache reset during language switching, but stop reloading the full permission/menu/route state; menu titles must update through reactive `$t()`.

### Host runtime observability and operations

- Add public `GET /api/v1/health` endpoint with lightweight database probe, returning `200 {status:"ok", mode:"<single|master|slave>"}` or `503`.
- Use GoFrame `Server.Run()` for HTTP graceful shutdown with ordered cleanup (cron, cluster, database) bounded by `shutdown.timeout`.
- Move upload file access `GET /api/v1/uploads/*` into the file module under protected routes with unified auth and permission middleware.
- Replace hard-coded `defaultManagedJobTimezone` with configurable `scheduler.defaultTimezone` defaulting to `UTC`.
- Split `config.Service` and `middleware.Service` interfaces by responsibility through embedded category interfaces.
- Delete empty placeholder packages `pkg/auditi18n/` and `pkg/audittype/`.

### Documentation and module decoupling

- Standardize OpenSpec main spec structure to require `## Purpose` and `## Requirements` sections.
- Define module enable/disable configuration with graceful backend degradation.
- Ensure exported methods, structs, and key fields carry proper Go doc comments.

### Out of scope

- Real audit-log modeling and persistence.
- API rate limiting, TraceID middleware, request cancellation, Vue global error boundary, and similar cross-cutting infrastructure.
- DI containerization and `cmd_http.go` controller assembly refactoring.
- Dictionary-management spec changes; the spec is already correct and only SQL implementation needs alignment.

## Capabilities

### New Capabilities

- `host-runtime-operations`: Health probes, graceful shutdown, protected static-resource routing, configurable scheduler timezone, service interface decomposition, and stale package cleanup.
- `cron-job-management`: Configurable default timezone and removal of foreign key constraints from the scheduled job table.
- `framework-i18n-runtime-performance`: Language switching optimization that avoids reloading full permissions, menus, and routes.
- `user-management`: Transactional deletion, batch delete endpoint, and query indexes.
- `role-management`: Transactional deletion, transactional `AssignUsers`, and batch delete endpoint.
- `user-role-association`: Reverse index on `sys_user_role.role_id` and transactional association cleanup.
- `menu-management`: Transactional deletion, in-memory `isDescendant`, and reverse index.
- `server-monitor`: Visibility-aware automatic frontend polling.
- `user-message`: Visibility-aware unread-message polling.
- `online-user`: Session activity index for timeout cleanup.
- `spec-governance`: OpenSpec main spec structure standardization and archive residual cleanup.
- `backend-conformance`: GoFrame v2 ORM/soft-delete conformance, controller/service layer constraints, documentation completeness, and production panic governance.
- `api-contract-consistency`: REST semantics, path parameter binding, API documentation tags, and batch delete endpoints.
- `module-decoupling`: Module enable/disable configuration with graceful backend degradation.

### Modified Capabilities

None (this is the initial establishment of these capabilities).

## Impact

- **Backend code**: `internal/service/{user,role,menu,cron,config,middleware,file}/`, `internal/cmd/cmd_http*.go`, `internal/controller/{user,role,file,health}/`, `api/{user,role,file,health}/v1/`, `pkg/{excelutil,closeutil,pluginbridge}`, and related plugin export logic.
- **SQL**: `001-project-init.sql`, `002-dict-dept-post.sql`, `008-menu-role-management.sql`, `014-scheduled-job-management.sql`, and the SQL file containing `sys_online_session`.
- **Configuration**: `config.template.yaml` adds `scheduler.defaultTimezone`, `health.timeout`, and `shutdown.timeout`.
- **Frontend**: `views/system/{user,role}/index.vue`, `views/monitor/server/index.vue`, `store/message.ts`, `router/guard.ts`, `bootstrap.ts`, and `api/system/{user,role}/index.ts`.
- **Tests**: New Go unit tests for transactional rollback, batch delete, panic allowlist, Excel helpers, and `isDescendant` boundaries; new E2E tests for health endpoint, batch delete, upload route authorization, server monitor polling, and language switching.
- **No i18n resource changes**: This change does not add, modify, or remove user-visible copy, API DTO documentation source text, or manifest/apidoc i18n resources.
- **No database schema migration**: All SQL changes are applied via `make init` with idempotent scripts; the project has no legacy burden.
- **Operational impact**: `/health` and graceful shutdown let Kubernetes and container orchestrators use standard probes and termination flows. Removing the foreign key reduces extra locking in high-concurrency scheduler paths.
- **API compatibility**: Batch delete endpoints are additive. Existing single-record `DELETE /api/v1/user/{id}` and `DELETE /api/v1/role/{id}` remain unchanged.
