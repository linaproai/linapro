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

## 18. Unit Tests, E2E, and Regression Verification

- [x] 18.1 Create E2E test for anonymous health probe access (`TC0147`)
- [x] 18.2 Create E2E test for user batch delete (`TC0148`)
- [x] 18.3 Create E2E test for role batch delete (`TC0149`)
- [x] 18.4 Create E2E test for server monitor visibility-aware polling (`TC0150`)
- [x] 18.5 Create E2E test for upload route requires auth (`TC0151`)
- [x] 18.6 Create E2E test for language switch no user-info reload (`TC0152`)
- [x] 18.7 Add Go unit tests for Excel helpers, invalid hostServices input, invalid runtime config values, and panic allowlist
- [x] 18.8 Run `go test ./...` and confirm all service-layer tests pass
- [x] 18.9 Run `pnpm test` and confirm all E2E tests pass

## 19. Review and Archive Readiness

- [x] 19.1 Run `/lina-review` for full change review covering code, SQL, E2E, and specification compliance
- [x] 19.2 Append and complete repair tasks based on review findings; sync spec deltas if behavior changed
- [x] 19.3 Rerun `openspec validate` and `make test`, confirming no regressions

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
