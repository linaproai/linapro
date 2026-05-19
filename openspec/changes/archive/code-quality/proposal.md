## Why

The LinaPro backend accumulated systematic deviations across GoFrame v2 ORM conventions, REST API contract consistency, production `panic` discipline, transactional correctness, SQL performance, API documentation quality, module decoupling, host runtime operability, backend source readability, service dependency management, and API response safety. These deviations degraded development velocity, introduced runtime risks (unnecessary panics, swallowed transaction errors, full table scans, implicit service graph construction, entity field leakage), and created friction during OpenSpec validation and archival. A consolidated code-quality hardening iteration was needed to establish a stable backend baseline before further feature work.

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

### Startup SQL efficiency

- Default SQL debug configuration set to `false`; startup logs no longer output ORM SQL detail by default.
- Shared `StartupContext` introduced for one HTTP startup orchestration, carrying catalog, integration, and job startup snapshots; `BootstrapAutoEnable`, plugin route registration, runtime frontend prewarm, and cron builtin sync all reuse the same context.
- Plugin manifest synchronization becomes difference-driven: no-op synchronization produces no transactions, no writes, no post-write reads.
- Built-in cron job registration uses declaration-derived projection snapshots; persistent scheduler startup scan excludes `is_builtin=1` jobs.
- Structured startup summary log replaces SQL-detail-based observability.

### API response hardening

- Replace all `entity.*` embeds in host and source-plugin API response DTOs with explicit response structs that only expose permitted fields. Passwords, soft-delete timestamps, storage paths, hashes, and internal tenant governance fields are excluded from API responses.
- Source-plugin API DTOs migrated from `*Entity` naming and `*_entity.go` placement to independent response DTOs in plugin API main source files.
- Plugin API contract unit tests migrated from aggregate `apps/lina-plugins` root to each plugin's own test directory.

### API timestamp contract

- Public HTTP JSON response DTO fields representing exact instants must return Unix timestamps in milliseconds (`int64`/`*int64`). `time.Time`, `*time.Time`, `gtime.Time`, `*gtime.Time`, and formatted strings are prohibited for instant fields at the response DTO boundary.
- Calendar-date fields (`birthday`, `businessDate`, `periodDate`) may use `YYYY-MM-DD` strings but must document `date-only` semantics.
- Existing host and source-plugin instant fields migrated through `apps/lina-core/pkg/apitime` projection helper; GoFrame DAO generation configured with `stdTime: true` to use `*time.Time` internally.
- Rule added to `lina-review` manual review checklist.

### API enum contracts

- Cross-module stable enum values abstracted into small public contract components: `pkg/listorder` (sort direction), `pkg/tenantoverride` (tenant override mode), `pkg/statusflag` (common 0/1 flags).
- Existing stable contracts (`pkg/menutype`, `pkg/pluginbridge`) reused directly in API DTOs without secondary constant forwarding.
- Domain-private enums remain in their respective API or service packages.
- External JSON field names, values, and defaults remain unchanged.

### Backend source readability governance

- Host and source-plugin `internal/service` component main files serve as contract entry points: only package comments, core types, interface definitions, implementation structs, compile-time assertions, and constructors remain; business logic migrates to responsibility-named files.
- `lina-core/pkg` public components follow the same main-file responsibility governance.
- Interface method comments must describe function, key inputs, outputs, errors, and applicable constraints (permissions, data permissions, caching, i18n, transactions, idempotency, concurrency).
- File-level top comments must explain file purpose, main logic, and caveats; main files use component-level package comments, non-main files use file-level comments.
- `linactl` command files named `command_<command>.go`; complex shared implementations migrated to `internal/<component>/` sub-packages.
- All rules added to `AGENTS.md` and `lina-review` review checklist.

### Explicit service dependency injection

- Host and source-plugin Controllers, Middleware, Service, plugin host service adapters, and WASM host services must receive runtime dependencies through explicit constructor parameters. No implicit `service.New()` calls in business constructors or request paths.
- Aggregate dependency structs (`Dependencies`, `Deps`, `Options`) prohibited for interface-typed runtime dependencies; each interface dependency must be a separate constructor parameter.
- Cache-sensitive components (auth, session, role, plugin, config, i18n, cachecoord, kvcache, locker, notify, host service adapters) must share startup-period instances or shared backends.
- Source plugins receive host-published dependencies through registrar `HostServices()` directory; plugin controllers and services construct from published adapters.
- Initialization and registration APIs must return `error` to callers; only top-level static registration entry points may `panic` after receiving errors.
- Static scanning and tests enforce that no new implicit `New()` calls appear in non-test, non-startup-boundary production paths.

### Documentation and module decoupling

- Standardize OpenSpec main spec structure to require `## Purpose` and `## Requirements` sections.
- Define module enable/disable configuration with graceful backend degradation.
- Ensure exported methods, structs, and key fields carry proper Go doc comments.

### Out of scope

- Real audit-log modeling and persistence.
- API rate limiting, TraceID middleware, request cancellation, Vue global error boundary, and similar cross-cutting infrastructure.
- DI containerization and `cmd_http.go` controller assembly refactoring (beyond explicit parameter passing).
- Dictionary-management spec changes; the spec is already correct and only SQL implementation needs alignment.

## Capabilities

### New Capabilities

- `host-runtime-operations`: Health probes, graceful shutdown, protected static-resource routing, configurable scheduler timezone, service interface decomposition, and stale package cleanup.
- `cron-job-management`: Configurable default timezone, removal of foreign key constraints from the scheduled job table, and startup registration deduplication.
- `framework-i18n-runtime-performance`: Language switching optimization that avoids reloading full permissions, menus, and routes.
- `user-management`: Transactional deletion, batch delete endpoint, and query indexes.
- `role-management`: Transactional deletion, transactional `AssignUsers`, and batch delete endpoint.
- `user-role-association`: Reverse index on `sys_user_role.role_id` and transactional association cleanup.
- `menu-management`: Transactional deletion, in-memory `isDescendant`, and reverse index.
- `server-monitor`: Visibility-aware automatic frontend polling.
- `user-message`: Visibility-aware unread-message polling.
- `online-user`: Session activity index for timeout cleanup.
- `spec-governance`: OpenSpec main spec structure standardization and archive residual cleanup.
- `startup-sql-efficiency`: Startup SQL count budget, log noise reduction, plugin startup snapshot reuse, no-op sync paths, and startup efficiency regression tests.
- `service-dependency-injection-governance`: Explicit dependency injection, shared instances, implicit construction prohibition, initialization error return, and review enforcement for host, source plugins, dynamic plugin host services, and WASM host services.

### Modified Capabilities

- `backend-conformance`: GoFrame v2 ORM/soft-delete conformance, controller/service layer constraints, documentation completeness, production panic governance, main-file responsibility governance, interface method comment requirements, file-level comment standards, and explicit dependency injection rules.
- `api-contract-consistency`: REST semantics, path parameter binding, API documentation tags, batch delete endpoints, response instant-field timestamp contract, API response DTO hardening, and cross-module enum contract abstraction.
- `module-decoupling`: Module enable/disable configuration with graceful backend degradation.
- `plugin-startup-bootstrap`: Startup bootstrap must reuse the same plugin governance startup snapshot across all startup phases.
- `plugin-manifest-lifecycle`: Plugin manifest synchronization must be difference-driven with no database side effects when no differences exist.
- `distributed-cache-coordination`: Cache-sensitive services must share runtime instances or shared backends; cluster mode must use coordination-backed services.
- `plugin-http-slot-extension`: Source plugin HTTP, middleware, and Cron registration callbacks receive host-published dependency directory.
- `plugin-host-service-extension`: Plugin host service adapters must be constructed by host runtime and shared, not instantiated per-call.

## Impact

- **Backend code**: `internal/service/{user,role,menu,cron,config,middleware,file,auth,session,bizctx,plugin,i18n,notify,usermsg,sysconfig,sysinfo,dict,datascope,tenantcap,orgcap,cluster,coordination,cachecoord,kvcache,locker,hostlock,jobmgmt,jobhandler,jobmeta,startupstats,pluginruntimecache,pluginhostservices,apidoc}/`, `internal/cmd/cmd_http*.go`, `internal/controller/{user,role,file,health,auth,menu,plugin,i18n,config,dict,publicconfig,usermsg,sysinfo,job,joblog,jobgroup,jobhandler}/`, `api/{user,role,file,health,auth,menu,plugin,i18n,config,dict,publicconfig,usermsg,sysinfo,job,joblog,jobgroup,jobhandler}/v1/`, `pkg/{excelutil,closeutil,pluginbridge,pluginhost,pluginservice,plugindb,sourceupgrade,apitime,listorder,tenantoverride,statusflag,menutype,authtoken,bizerr,dbdriver,gdbutil,logger,orgcap,pluginfs,tenantcap,testsupport,dialect,i18nresource}`, and all source-plugin backend service and controller code.
- **SQL**: `001-project-init.sql`, `002-dict-dept-post.sql`, `008-menu-role-management.sql`, `014-scheduled-job-management.sql`, and the SQL file containing `sys_online_session`.
- **Configuration**: `config.template.yaml` adds `scheduler.defaultTimezone`, `health.timeout`, and `shutdown.timeout`; `config.yaml` default `database.default.debug` set to `false`.
- **Frontend**: `views/system/{user,role}/index.vue`, `views/monitor/server/index.vue`, `store/message.ts`, `router/guard.ts`, `bootstrap.ts`, `api/system/{user,role}/index.ts`, and all pages displaying migrated timestamp fields.
- **Tests**: New Go unit tests for transactional rollback, batch delete, panic allowlist, Excel helpers, `isDescendant` boundaries, startup SQL statistics, plugin no-op synchronization, explicit dependency injection, and API DTO hardening; new E2E tests for health endpoint, batch delete, upload route authorization, server monitor polling, and language switching.
- **i18n impact**: Runtime UI language packs and manifest runtime i18n resources were not changed. Host and source-plugin `zh-CN` apidoc i18n JSON resources were updated for migrated timestamp fields and cleaned for removed entity-exposed response fields.
- **No database schema migration**: All SQL changes are applied via `make init` with idempotent scripts; the project has no legacy burden.
- **Operational impact**: `/health` and graceful shutdown let Kubernetes and container orchestrators use standard probes and termination flows. Removing the foreign key reduces extra locking in high-concurrency scheduler paths.
- **API compatibility**: Batch delete endpoints are additive. Existing single-record `DELETE /api/v1/user/{id}` and `DELETE /api/v1/role/{id}` remain unchanged. API response field types for instant fields changed from formatted/time-object values to numeric millisecond timestamps. API responses no longer expose database entity internal fields.
