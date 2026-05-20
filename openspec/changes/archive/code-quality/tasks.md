## 1. OpenSpec Specification Governance

- [x] 1.1 Audit and fix `openspec/specs/` main spec structures that do not conform to current schema requirements
- [x] 1.2 Clean archive residual files and verify related capabilities pass `openspec validate` and archival
- [x] 1.3 Add `spec-governance` specification and documentation for this change

## 2. GoFrame ORM and Soft-Delete Conformance

- [x] 2.1 Audit `apps/lina-core/internal/controller/` and `apps/lina-core/internal/service/` for GoFrame v2 compliance violations
- [x] 2.2 Fix hand-written soft-delete filters (`WhereNull(deleted_at)`), non-recommended ORM usage, and dependency injection issues in production code
- [x] 2.3 Add `deleted_at DATETIME DEFAULT NULL` to `sys_dict_type` and `sys_dict_data` in `002-dict-dept-post.sql`
- [x] 2.4 Run `make init` and `make dao` to regenerate entities and confirm `SysDictType`/`SysDictData` include `DeletedAt`
- [x] 2.5 Ensure all database write operations use DO objects instead of `g.Map`

## 3. Production Panic Governance

- [x] 3.1 Establish a production backend `panic` allowlist documenting retained categories: startup, registration, `Must*`, and unknown panic rethrow
- [x] 3.2 Convert unnecessary `panic` calls in Excel cell coordinate and file-closing helpers into explicit error returns
- [x] 3.3 Split dynamic plugin hostServices normalization into `NormalizeHostServiceSpecsE` (error-returning) and `MustNormalizeHostServiceSpecs` (Must path)
- [x] 3.4 Return explicit `error` values for runtime configuration parsing failures instead of silent degradation
- [x] 3.5 Add a static check script or test that scans production Go files and blocks `panic` calls outside the allowlist
- [x] 3.6 Convert initialization and registration APIs (plugin host registration, registrar, callback registration) to return `error` instead of panicking on expected failures

## 4. REST API Contract Consistency

- [x] 4.1 Unify `apps/lina-core/api/` path parameter binding: `g.Meta` uses `{param}`, input DTO fields use `json:"param"`
- [x] 4.2 Ensure all read operations use `GET`, write operations use `POST`/`PUT`, deletions use `DELETE` with resource-based URLs
- [x] 4.3 Add comprehensive `dc` and `eg` documentation tags to all API DTO input/output fields
- [x] 4.4 If API contracts change, update frontend calls and related E2E tests

## 5. Transactional Correctness Fixes

- [x] 5.1 Modify `Delete` in `internal/service/user/user.go`: wrap user soft delete, organization cleanup, and role association cleanup in one transaction; return errors; notify after commit
- [x] 5.2 Modify `Delete` in `internal/service/role/role.go`: change transaction-internal cleanup failures from `Warningf` to `return err`
- [x] 5.3 Refactor `AssignUsers` in `internal/service/role/role.go`: build `[]do.SysUserRole` and perform one `Insert(slice)` inside a transaction
- [x] 5.4 Modify `Delete` in `internal/service/menu/menu.go`: change `sys_role_menu` cleanup failure from `Warningf` to `return err`
- [x] 5.5 Confirm required `bizerr.Code` values exist; add missing values and sync `manifest/i18n/<locale>/error.json` if needed
- [x] 5.6 Add unit tests for rollback on user/role/menu deletion association cleanup failure and mid-operation `AssignUsers` failure

## 6. SQL Index and Structure Adjustments

- [x] 6.1 Add `KEY idx_status`, `KEY idx_phone`, `KEY idx_created_at` to `sys_user` in `001-project-init.sql`
- [x] 6.2 Add `KEY idx_role_id (role_id)` to `sys_user_role` in `008-menu-role-management.sql`
- [x] 6.3 Add `KEY idx_menu_id (menu_id)` to `sys_role_menu` in `008-menu-role-management.sql`
- [x] 6.4 Add `KEY idx_last_active_time (last_active_time)` to `sys_online_session`
- [x] 6.5 Remove `CONSTRAINT fk_sys_job_group_id` and add `KEY idx_group_id (group_id)` to `sys_job` in `014-scheduled-job-management.sql`
- [x] 6.6 Run `make init` and verify all SQL is idempotent and the new structure is correct

## 7. Backend Batch Delete APIs

- [x] 7.1 Add `BatchDeleteReq`/`BatchDeleteRes` in `api/user/v1/`: `DELETE /api/v1/user`, permission `system:user:remove`
- [x] 7.2 Add `BatchDeleteReq`/`BatchDeleteRes` in `api/role/v1/`: `DELETE /api/v1/role`, permission `system:role:remove`
- [x] 7.3 Run `make ctrl`, implement `BatchDelete` methods in `controller/user/` and `controller/role/`
- [x] 7.4 Add `BatchDelete(ctx, ids) error` to `service/user/` and `service/role/`: reuse all `Delete` protections in one transaction
- [x] 7.5 Add service-layer batch delete tests for success, built-in admin rejection, current-user rejection, and empty-list validation

## 8. Menu Performance: In-Memory isDescendant

- [x] 8.1 Rewrite `isDescendant` in `internal/service/menu/menu.go`: load parent-child map once with `dao.SysMenu.Ctx(ctx).Scan(&all)` and run in-memory BFS
- [x] 8.2 Add unit tests for `isDescendant` correctness: self is not descendant, cross-depth match, missing ids

## 9. Configurable Scheduler Timezone

- [x] 9.1 Remove hard-coded `defaultManagedJobTimezone = "Asia/Shanghai"` from `cron_managed_jobs.go`; read `scheduler.defaultTimezone` and default to `UTC`
- [x] 9.2 Add `scheduler.defaultTimezone: "UTC"` to `config.template.yaml` with English comments

## 10. Upload Route Authorization

- [x] 10.1 Move `GET /api/v1/uploads/*` into `api/file/v1` and `internal/controller/file`; mount under protected route group with Auth and Permission middleware
- [x] 10.2 Use `system:file:download` permission tag; enforce through unified permission middleware
- [x] 10.3 Read files through file service storage backend, not by concatenating local paths in `cmd_http.go`

## 11. Health Probe and Graceful Shutdown

- [x] 11.1 Add `GET /api/v1/health` through standard API/controller flow; run `dao.SysUser.Ctx(ctx).Limit(1).Count()` as DB probe
- [x] 11.2 Return `{status:"ok", mode:"<single|master|slave>"}` on 200, `{status:"unavailable", reason:"..."}` on 503
- [x] 11.3 Add `health.timeout: "5s"` and `shutdown.timeout: "30s"` to `config.template.yaml`; parse as `time.Duration`
- [x] 11.4 Use GoFrame `Server.Run()` for built-in signal handling; after return, clean up in order: cron stop, cluster stop, DB close, bounded by `shutdown.timeout`
- [x] 11.5 Add `Stop(ctx)` to cron component if missing, for graceful shutdown support

## 12. Service Interface Decomposition

- [x] 12.1 Split `config.Service` into embedded category interfaces (cluster, auth, login, frontend, i18n, cron, host runtime, delivery metadata, plugin, upload, runtime parameter sync)
- [x] 12.2 Split `middleware.Service` into `HTTPMiddleware` and `RuntimeSupport` interfaces

## 13. Stale Package Cleanup

- [x] 13.1 Grep repository and confirm `pkg/auditi18n` and `pkg/audittype` have no imports
- [x] 13.2 Delete empty directories `pkg/auditi18n/` and `pkg/audittype/`

## 14. Documentation Completeness

- [x] 14.1 Add proper Go doc comments to exported methods, structs, and key fields across `internal/controller/` and `internal/service/`

## 15. Module Decoupling Specification

- [x] 15.1 Define module enable/disable configuration and graceful backend degradation requirements
- [x] 15.2 Document that module disable only affects feature exposure, not data integrity

## 16. Frontend Batch Operations

- [x] 16.1 Add `userBatchDelete(ids)` in `api/system/user/index.ts` using repeated `ids` query parameters
- [x] 16.2 Add `roleBatchDelete(ids)` in `api/system/role/index.ts` using repeated `ids` query parameters
- [x] 16.3 Replace loop-over-single-delete in `views/system/user/index.vue` with one batch API call
- [x] 16.4 Replace loop-over-single-delete in `views/system/role/index.vue` with one batch API call

## 17. Frontend Polling, Cache, and Language-Switching Optimizations

- [x] 17.1 Add visibility-aware 30s auto-refresh to `views/monitor/server/index.vue` using `useIntervalFn` + `useDocumentVisibility`
- [x] 17.2 Replace raw `setInterval` in `store/message.ts` with visibility-aware polling; pause while hidden, refresh on visibility restore
- [x] 17.3 Replace `loadedPaths` in `router/guard.ts` with bounded LRU (limit 50); move hits to tail, evict oldest on overflow
- [x] 17.4 In `bootstrap.ts` language switching: keep `syncPublicFrontendSettings` and `useDictStore().resetCache()`, remove `refreshAccessibleState(router)`, update menu titles via `meta.i18nKey` and `$t()`
- [x] 17.5 Scan `meta.title` definitions in route modules; ensure i18n keys or `() => $t(...)` are used; fix static hard-coded strings

## 18. Startup SQL Efficiency

- [x] 18.1 Record current startup SQL baseline: SQL count, startup duration, plugin count, builtin job count from backend process start to route binding completion
- [x] 18.2 Set `database.default.debug` default to `false` in deliverable configuration with diagnostic-mode comments
- [x] 18.3 Add default startup log smoke test asserting no ORM SQL detail in startup logs
- [x] 18.4 Design and implement shared `StartupContext` carrying catalog, integration, and job startup snapshots plus statistics collector
- [x] 18.5 Adjust startup orchestration function signatures so `BootstrapAutoEnable`, plugin route registration, runtime frontend prewarm, and cron startup reuse the same context
- [x] 18.6 Implement plugin registry and release metadata post-write snapshot update using `InsertAndGetId` and `existing + data`
- [x] 18.7 Add menu and resource ref difference comparison functions; no-op path skips transactions and writes
- [x] 18.8 Ensure `SyncManifest` orchestration: when registry, release, menu, and resource ref all match, no database writes or post-write reads occur
- [x] 18.9 Add source-plugin no-op sync test covering zero `INSERT`/`UPDATE`/`DELETE`/empty transactions
- [x] 18.10 Add difference sync test covering manifest, release snapshot, menu, route permission, and resource ref changes
- [x] 18.11 Confirm builtin job registration uses declaration-derived projection snapshots; persistent scan excludes `is_builtin=1`
- [x] 18.12 Add scheduler test asserting builtin jobs registered via `RegisterJobSnapshot` and persistent scan only loads non-builtin enabled jobs
- [x] 18.13 Add startup statistics collector recording plugin scan count, sync change count, no-op count, snapshot construction count, builtin projection count, and phase durations
- [x] 18.14 Output structured startup summary log after startup completes; no full SQL text included
- [x] 18.15 Run backend startup smoke tests for MySQL and SQLite configurations

## 19. API Response DTO Hardening

- [x] 19.1 Audit host API definitions that embed or return `entity.*` in response types; confirm affected modules and risk fields
- [x] 19.2 Define independent response DTOs for user, file, system config, dict, scheduled job, job log, and job group responses; only expose necessary fields
- [x] 19.3 Adjust controller response assembly to map allowed fields explicitly; prohibit direct entity pointer assignment to response
- [x] 19.4 Add automated tests or static verification ensuring API layer no longer depends on `internal/model/entity` and user responses do not expose `password`
- [x] 19.5 Migrate source-plugin API DTOs from `*Entity` naming to independent response DTOs in plugin API main source files
- [x] 19.6 Migrate plugin API contract tests from aggregate `apps/lina-plugins` root to each plugin's own test directory

## 20. API Timestamp Response Contract

- [x] 20.1 Add API response instant-field contract to project API design rules: Unix timestamp milliseconds, DTO types, prohibited types, calendar-date exceptions
- [x] 20.2 Add time-field manual review requirement to `lina-review` RESTful/API DTO checklist
- [x] 20.3 Migrate existing host and source-plugin public response DTO instant fields to Unix millisecond timestamps via `pkg/apitime`
- [x] 20.4 Update frontend API types and page display formatting for migrated timestamp fields
- [x] 20.5 Update host and source-plugin `zh-CN` apidoc i18n JSON resources for migrated instant fields
- [x] 20.6 Configure GoFrame DAO generation with `stdTime: true` and timestamp `typeMapping` to use `*time.Time` internally
- [x] 20.7 Regenerate host and source-plugin DAO/DO/entity artifacts; adjust internal services and tests for `*time.Time`

## 21. API Enum Contract Abstraction

- [x] 21.1 Create `pkg/listorder`, `pkg/tenantoverride`, `pkg/statusflag` small public contract components with comments and unit tests
- [x] 21.2 Adjust `apps/lina-core/api` references for sort direction, tenant override mode, menu type, plugin bridge type, and common status flags to use public components
- [x] 21.3 Update affected controller type conversions and constant references; ensure service layer still receives original domain semantics
- [x] 21.4 Run public component tests, API/controller compilation smoke test, and OpenSpec validation

## 22. Backend Source Readability Governance

- [x] 22.1 Update `AGENTS.md` with main-file contract entry point, public component main-file responsibility, interface method detailed comments, and file-level detailed comments
- [x] 22.2 Update `lina-review` checklist with main-file responsibility, interface method comment completeness, file-level comment quality, and batch verification record review items
- [x] 22.3 Scan host `internal/service/**`, `lina-core/pkg/**`, and source-plugin `backend/internal/service/**` main files; establish baseline of components needing migration
- [x] 22.4 Migrate host auth, session, middleware, bizctx main files to contract entry points; move implementations to responsibility files
- [x] 22.5 Migrate host user, role, datascope, tenantcap, orgcap main files to contract entry points
- [x] 22.6 Migrate host config, sysconfig, sysinfo, dict, menu, file main files to contract entry points
- [x] 22.7 Migrate host cron, jobhandler, jobmeta, jobmgmt, startupstats main files to contract entry points
- [x] 22.8 Migrate host cluster, coordination, cachecoord, kvcache, locker, hostlock main files to contract entry points
- [x] 22.9 Migrate host i18n, notify, usermsg, apidoc main files to contract entry points
- [x] 22.10 Migrate host plugin outer service, pluginruntimecache, pluginhostservices main files to contract entry points
- [x] 22.11 Migrate plugin/internal/catalog, runtime, integration, frontend, openapi, lifecycle, wasm main files to contract entry points
- [x] 22.12 Migrate `lina-core/pkg` small public components (authtoken, bizerr, closeutil, dbdriver, excelutil, gdbutil, logger, menutype, orgcap, pluginfs, tenantcap, testsupport) main files
- [x] 22.13 Migrate `lina-core/pkg` plugin and bridge public components (pluginhost, pluginbridge, pluginservice, plugindb, sourceupgrade) main files
- [x] 22.14 Migrate `lina-core/pkg` database, dialect, and resource public components (dialect, i18nresource) main files
- [x] 22.15 Migrate source-plugin org-center and multi-tenant plugin service main files
- [x] 22.16 Migrate source-plugin monitor and content plugin service main files
- [x] 22.17 Migrate source-plugin demo plugin service main files
- [x] 22.18 Add `linactl` command file naming governance: `command_<command>.go` convention with dot-segment semantics
- [x] 22.19 Migrate `linactl` shared implementations to `internal/<component>/` sub-packages; delete old `command_ops.go`
- [x] 22.20 Full-scan host, source-plugin, and `lina-core/pkg` main files to confirm complex implementations migrated or documented exceptions

## 23. Explicit Service Dependency Injection

- [x] 23.1 Scan host and source-plugin production Go files for implicit key service construction calls; classify by Controller, Service, Middleware, pluginservice, source plugin, WASM host service
- [x] 23.2 Mark high-risk cache-consistency paths; record authoritative data sources, shared instance requirements, and migration priority
- [x] 23.3 Update `AGENTS.md` with explicit dependency injection rules, implicit construction prohibition, cache-sensitive instance sharing, and test exemption boundaries
- [x] 23.4 Update `lina-review` checklist with backend dependency injection review items covering Controller, Service, Middleware, source plugin, pluginservice, WASM host service, and cache-sensitive instance sources
- [x] 23.5 Add static scanning script or governance verification to identify new implicit `New()` calls in production paths; maintain allowlist for startup boundaries, test files, and stateless exemptions
- [x] 23.6 Refactor `auth.Service` constructor to explicitly receive config, plugin, orgcap, role, tenant, session store, token state dependencies
- [x] 23.7 Refactor `middleware.Service` constructor to explicitly receive auth, bizctx, config, i18n, plugin, role, tenant dependencies
- [x] 23.8 Refactor role, menu, user, dict, file, usermsg, notify, sysconfig, i18n service constructors to remove implicit cache-sensitive dependency creation
- [x] 23.9 Refactor datascope, tenantcap, orgcap service constructors for explicit dependency passing
- [x] 23.10 Update `cmd_http_runtime.go` runtime structure to hold shared host service instances
- [x] 23.11 Refactor all host Controller `NewV1` constructors to receive service dependencies through explicit parameters
- [x] 23.12 Update `cmd_http_routes.go` to construct Controllers from shared runtime instances before route binding
- [x] 23.13 Add host-published service directory to `pluginhost` HTTP/Cron registrar; expose stable `pkg/pluginservice/*` adapters
- [x] 23.14 Refactor `pkg/pluginservice/*` adapters to receive internal dependencies from host runtime
- [x] 23.15 Migrate source-plugin `backend/plugin.go` route, middleware, and Cron registration callbacks to obtain host-published dependencies from registrar
- [x] 23.16 Refactor source-plugin Controller and Service constructors to remove implicit host adapter creation
- [x] 23.17 Refactor WASM host service `ConfigureXxxHostService` entry points to receive shared host services from startup path
- [x] 23.18 Ensure dynamic plugin host service handlers do not create independent service instances per call
- [x] 23.19 Converge `pkg/pluginservice/*` pure contract source files into `contract` sub-package; remove empty shell contract packages
- [x] 23.20 Expand `ConfigureWasmHostServices` and `pluginhostservices.New` to accept direct parameters instead of aggregate dependency structs
- [x] 23.21 Convert `plugin.New`, `sysinfo.New`, `hostlock.New`, `jobmgmt.NewScheduler`, WASM `ConfigureXxxHostService`, `sourceupgrade.New`, `tenantfilter.New`, `pluginhostservices.New` to return `error` instead of panicking on expected failures
- [x] 23.22 Split `pluginhost_source_plugin.go` (1324 lines) into responsibility files for registration, callbacks, inputs, descriptors, manifest, and values
- [x] 23.23 Convert `pluginhost` registration APIs to return `error`; source-plugin `init` top-level entry points handle errors and choose to panic
- [x] 23.24 Run host core service unit tests, host Controller and cmd route tests, all source-plugin backend tests, static scan, OpenSpec validation, and `lina-review`

## 24. Unit Tests, E2E, and Regression Verification

- [x] 24.1 Create E2E test for anonymous health probe access
- [x] 24.2 Create E2E test for user batch delete
- [x] 24.3 Create E2E test for role batch delete
- [x] 24.4 Create E2E test for server monitor visibility-aware polling
- [x] 24.5 Create E2E test for upload route requires auth
- [x] 24.6 Create E2E test for language switch no user-info reload
- [x] 24.7 Add Go unit tests for Excel helpers, invalid hostServices input, invalid runtime config values, and panic allowlist
- [x] 24.8 Run `go test ./...` and confirm all service-layer tests pass
- [x] 24.9 Run `pnpm test` and confirm all E2E tests pass

## 25. Review and Archive Readiness

- [x] 25.1 Run `/lina-review` for full change review covering code, SQL, E2E, and specification compliance
- [x] 25.2 Append and complete repair tasks based on review findings; sync spec deltas if behavior changed
- [x] 25.3 Rerun `openspec validate` and `make test`, confirming no regressions

## Feedback

- [x] **FB-1**: Unify backend API input DTO parameter tags to `json`, prohibit mixed `p` and `json` usage
- [x] **FB-2**: Remove out-of-scope `dept/post` module switch implementation, restore pure spec-conformance scope
- [x] **FB-3**: Keep existing API route addresses unchanged; only fix parameter tags, documentation tags, and comment consistency
- [x] **FB-4**: Cron runtime configuration reads should return explicit errors instead of degrading through logs
- [x] **FB-5**: `closeutil` and `excelutil` close-error logs should explain nil error pointer misuse and receive caller context
- [x] **FB-6**: Logging calls must propagate `ctx` through the call chain to preserve tracing
- [x] **FB-7**: Panic allowlist check should move into `internal/cmd` test directory and not treat test helpers as production panic boundaries
- [x] **FB-8**: Panic allowlist test should reduce coupling to custom string concatenation and scanning logic
- [x] **FB-9**: Normalize SQL line comments so each comment uses English above Chinese on separate lines
- [x] **FB-10**: Move upload file access routing into the file API module and reuse file storage access logic
- [x] **FB-11**: Remove redundant custom HTTP signal handling and rely on GoFrame Server.Run graceful shutdown
- [x] **FB-12**: Split `cmd_http.go` by responsibility to reduce single-file complexity
- [x] **FB-13**: Change health probe default timeout to 5s
- [x] **FB-14**: Split config Service into categorized embedded interfaces
- [x] **FB-15**: Keep plugin install SQL seed-only and avoid cleanup DELETE statements in install scripts
- [x] **FB-16**: Split middleware Service into HTTP middleware and non-middleware support interfaces
- [x] **FB-17**: Harden dict E2E deletion targeting so tests cannot soft-delete built-in system dictionaries
- [x] **FB-18**: Reduce redundant database reads and writes during lina-core startup reconciliation
- [x] **FB-19**: Make persistent cron registration idempotent when startup handler restoration refreshes the same job
- [x] **FB-20**: Load small plugin and menu governance tables as startup snapshots to avoid N+1 reconciliation queries
- [x] **FB-21**: Load scheduled-job startup governance rows as snapshots to avoid built-in job reconciliation N+1 queries
- [x] **FB-22**: Abstract startup phase observation keys from string literals to named types and constants
- [x] **FB-23**: Implement timestamp response contract in existing public API response DTOs instead of only documenting the rule
- [x] **FB-24**: Replace generated and internal `*gtime.Time` time-field types with `*time.Time`; configure GoFrame DAO generation to use standard library time fields
- [x] **FB-25**: Multiple host and source-plugin new implementation file top comments too generic; must explain implementation responsibility, main flow, and key constraints
- [x] **FB-26**: `linactl` command implementation files missing per-command naming governance and review requirements
- [x] **FB-27**: `linactl` shared implementations still piled in root directory; must migrate to `internal/<component>/` sub-packages
- [x] **FB-28**: `file.Service` and `jobmgmt.Service` create `datascope.Service` inside constructor; data-scope dependency not explicitly injected
- [x] **FB-29**: WASM host service configuration entry silently creates default instance when dependency is nil
- [x] **FB-30**: `cmd_http_plugin_services.go` implements source-plugin host service directory in HTTP init package; boundary not decoupled from startup orchestration
- [x] **FB-31**: `content-notice` controller exposes both `NewV1` and `NewControllerV1`; plugin route binding bypasses unified interface constructor entry
- [x] **FB-32**: `plugin.New` creates runtime key dependencies inside constructor; explicit injection parameters must not use additional struct wrapping
- [x] **FB-33**: `plugin_startup_consistency.go` temporarily creates `tenantcap.Service` and `bizctx.Service` during startup consistency check
- [x] **FB-34**: `monitor-server` config loading exposes both `Load` and `LoadWithReader`; `Load` internally creates host config service
- [x] **FB-35**: `multi-tenant` auth controller exposes both `NewV1` and `NewControllerV1`; plugin route binding bypasses unified interface constructor entry
- [x] **FB-36**: `multi-tenant` platform and tenant controllers expose both `NewV1` and `NewControllerV1`
- [x] **FB-37**: `role.New` creates `datascope.Service` inside constructor; data-scope dependency not explicitly injected
- [x] **FB-38**: `plugin-demo-source` uninstall callback constructs service via `demosvc.New(nil)`; dependency shape inconsistent with cleanup responsibility
- [x] **FB-39**: `pkg/sourceupgrade.New` implicitly creates config, context, cache coordination, i18n, session, and plugin service graph during facade initialization
- [x] **FB-40**: `plugin-demo-source` demo controller exposes both `NewV1` and `NewControllerV1`
- [x] **FB-41**: `user.New` creates `datascope.Service` inside constructor; user management data-scope dependency does not reuse startup-period shared instance
- [x] **FB-42**: `pluginservice/tenantfilter` reads host `bizctx` through package-level global config; not injected through `HostServices`
- [x] **FB-43**: `pkg/pluginservice/*` pure contract source files scattered and thin; plugin host service contracts not unified into `contract` component
- [x] **FB-44**: `ConfigureWasmHostServices` and `pluginhostservices.New` use aggregate dependency struct; explicit injection parameters not directly expanded
- [x] **FB-45**: Project specs, OpenSpec specs, and `lina-review` do not explicitly prohibit passing interface-typed runtime dependencies through aggregate structs
- [x] **FB-46**: `role_new.go` mistakenly appended with empty `ControllerV1` and parameterless `NewV1`; role controller duplicate declaration bypasses explicit dependency injection
- [x] **FB-47**: `internal/service` runtime initialization entries `panic` directly on missing dependencies instead of returning explicit `error`
- [x] **FB-48**: Source-plugin registration and registrar APIs still `panic` internally instead of returning errors to top-level entry decisions
- [x] **FB-49**: `pkg/pluginhost/pluginhost_source_plugin.go` too long; source-plugin registration, input wrapping, descriptor, and snapshot responsibilities mixed
- [x] **FB-50**: `multi-tenant` lifecycle precondition constructs half-initialized service via `newTenantService(nil)`; tenant count dependency not narrowed to actual responsibility
- [x] **FB-51**: Panic allowlist references pre-split file paths; `TestProductionPanicsMatchAllowlist` reports false new/expired panic exceptions
- [x] **FB-52**: Delete API package compatibility aliases and constant forwarding for public contract types; DTO fields should directly reference `pkg` public types
- [x] **FB-53**: Some API responses directly embed database entities; response contract driven by database structure; may expose passwords, soft-delete fields, storage paths, hashes, internal tenant fields
- [x] **FB-54**: Source-plugin API DTOs still use `*Entity` naming, `*_entity.go` placement, soft-delete field exposure, and operlog list returns full request/response payloads
- [x] **FB-55**: `apps/lina-plugins` aggregate root directory contains source-plugin API contract tests; test ownership not closed-loop to each plugin
