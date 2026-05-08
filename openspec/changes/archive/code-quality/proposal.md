## Why

The LinaPro backend accumulated systematic deviations across GoFrame v2 ORM conventions, REST API contract consistency, production `panic` discipline, transactional correctness, SQL performance, API documentation quality, module decoupling, and host runtime operability. These deviations degraded development velocity, introduced runtime risks (unnecessary panics, swallowed transaction errors, full table scans), and created friction during OpenSpec validation and archival. A consolidated code-quality hardening iteration was needed to establish a stable backend baseline before further feature work.

## What Changes

### GoFrame ORM and soft-delete conformance

- Eliminate hand-written `WhereNull(deleted_at)` filters and manual `created_at`/`updated_at`/`deleted_at` maintenance, relying on GoFrame automatic soft-delete and timestamp behavior instead.
- Use DO objects for database write operations instead of `g.Map`.
- Add `deleted_at DATETIME DEFAULT NULL` to `sys_dict_type` and `sys_dict_data` to align dictionary tables with other business tables.

### REST API contract consistency

- Unify all path parameters to `{param}` syntax in `g.Meta` with `json:"param"` tags on input DTO fields; eliminate mixed `p:`/`json:` tag usage.
- Ensure all read operations use `GET`, write operations use `POST`/`PUT`, and deletions use `DELETE` with resource-based URL naming.
- Add comprehensive `dc` and `eg` documentation tags to all API DTO fields for OpenAPI documentation completeness.
- Add RESTful batch delete endpoints for users and roles (`DELETE /api/v1/user?ids=...`, `DELETE /api/v1/role?ids=...`).

### Production panic governance

- Define the allowed `panic` boundary: startup, initialization, unrecoverable critical paths, `Must*` semantic constructors, and unknown panic rethrow.
- Convert unnecessary runtime panics in Excel helpers, resource closing, runtime configuration parsing, and dynamic plugin hostServices normalization into explicit `error` returns.
- Add a static check with an allowlist that blocks new `panic` calls outside approved categories.

### Transactional correctness

- Wrap user, role, and menu deletion with their association cleanup inside single `dao.Xxx.Transaction` closures; any cleanup failure rolls back the entire operation.
- Refactor `AssignUsers` to use one transaction with batch insert instead of per-row inserts with swallowed warnings.
- Replace `Warningf` log-and-continue patterns inside transactions with explicit error returns.

### SQL performance and structure

- Add query indexes: `idx_status`, `idx_phone`, `idx_created_at` on `sys_user`; `idx_role_id` on `sys_user_role`; `idx_menu_id` on `sys_role_menu`; `idx_last_active_time` on `sys_online_session`.
- Remove `sys_job` foreign key constraint `fk_sys_job_group_id`, replacing with application-level consistency and `KEY idx_group_id`.
- Rewrite menu `isDescendant` from per-level SQL to in-memory BFS traversal.

### Host runtime operations

- Add public `GET /api/v1/health` endpoint with lightweight database probe, returning `200 {status:"ok", mode:"..."}` or `503`.
- Use GoFrame `Server.Run()` for HTTP graceful shutdown with ordered cleanup (cron, cluster, database) bounded by `shutdown.timeout`.
- Move upload file access `GET /api/v1/uploads/*` into the file module under protected routes with unified auth and permission middleware.
- Replace hard-coded `defaultManagedJobTimezone` with configurable `scheduler.defaultTimezone` defaulting to `UTC`.
- Split `config.Service` and `middleware.Service` interfaces by responsibility through embedded category interfaces.
- Delete empty placeholder packages `pkg/auditi18n/` and `pkg/audittype/`.

### Frontend performance and operability

- Add visibility-aware polling for server monitor (30s auto-refresh) and user-message unread counts.
- Replace `loadedPaths` with a bounded LRU (limit 50) to prevent unbounded growth.
- Remove `refreshAccessibleState` from language switching; menu titles update reactively through `meta.i18nKey` and `$t()`.
- Switch frontend batch delete from looping single-delete calls to single batch API requests.

### Documentation and module decoupling

- Standardize OpenSpec main spec structure to require `## Purpose` and `## Requirements` sections.
- Define module enable/disable configuration with graceful backend degradation.
- Ensure exported methods, structs, and key fields carry proper Go doc comments.

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
