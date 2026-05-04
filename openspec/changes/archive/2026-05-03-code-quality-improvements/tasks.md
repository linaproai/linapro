## 1. SQL structure and index adjustments

- [x] 1.1 Modify `apps/lina-core/manifest/sql/001-project-init.sql`: add `KEY idx_status (status)`, `KEY idx_phone (phone)`, and `KEY idx_created_at (created_at)` to `sys_user`; keep SQL idempotent
- [x] 1.2 Modify `apps/lina-core/manifest/sql/002-dict-dept-post.sql`: add `deleted_at DATETIME DEFAULT NULL` to `sys_dict_type` and `sys_dict_data`, aligning with the existing `dict-management` spec
- [x] 1.3 Modify `apps/lina-core/manifest/sql/008-menu-role-management.sql`: add `KEY idx_role_id (role_id)` to `sys_user_role` and `KEY idx_menu_id (menu_id)` to `sys_role_menu`
- [x] 1.4 Locate the SQL file containing `sys_online_session` and add `KEY idx_last_active_time (last_active_time)`
- [x] 1.5 Modify `apps/lina-core/manifest/sql/014-scheduled-job-management.sql`: remove `CONSTRAINT fk_sys_job_group_id` and add `KEY idx_group_id (group_id)`
- [x] 1.6 Run `make init` in `apps/lina-core` and verify all SQL is idempotent and the new structure is correct
- [x] 1.7 Run `make dao` in `apps/lina-core`, regenerate dao/do/entity, and confirm `SysDictType` / `SysDictData` include `DeletedAt`

## 2. Backend transactional correctness fixes

- [x] 2.1 Modify `Delete` in `apps/lina-core/internal/service/user/user.go`: wrap user soft delete, organization cleanup, and role association cleanup in `dao.SysUser.Transaction(ctx, ...)`; return any transaction error; call `NotifyAccessTopologyChanged` after commit
- [x] 2.2 Modify `Delete` in `apps/lina-core/internal/service/role/role.go`: change transaction-internal cleanup failures for `sys_role_menu` and `sys_user_role` from `Warningf` to `return err`
- [x] 2.3 Refactor `AssignUsers` in `apps/lina-core/internal/service/role/role.go`: build `[]do.SysUserRole` and perform one `Insert(slice)` inside a transaction; remove per-row `Warningf` fallback
- [x] 2.4 Modify `Delete` in `apps/lina-core/internal/service/menu/menu.go`: change transaction-internal `sys_role_menu` cleanup failure from `Warningf` to `return err`
- [x] 2.5 Confirm required `bizerr.Code` values exist in `apps/lina-core/internal/service/role` and `user` `*_code.go` files, such as `CodeUserBuiltinAdminDeleteDenied`, `CodeUserCurrentDeleteDenied`, and `CodeRoleBuiltinDeleteDenied`; add missing values and sync `manifest/i18n/<locale>/error.json` if needed
- [x] 2.6 Add GoFrame unit/integration tests for rollback on `user.Delete` association cleanup failure, `role.Delete` association cleanup failure, and mid-operation `role.AssignUsers` failure

## 3. Backend batch delete APIs

- [x] 3.1 Add an independent file under `apps/lina-core/api/user/v1/` defining `BatchDeleteReq`/`BatchDeleteRes`: `DELETE /api/v1/user`, query `Ids []int json:"ids" v:"required|min-length:1"`, English `dc`/`eg`, and `g.Meta` permission `system:user:remove`
- [x] 3.2 Add an independent file under `apps/lina-core/api/role/v1/` defining `BatchDeleteReq`/`BatchDeleteRes`: `DELETE /api/v1/role`, permission `system:role:remove`
- [x] 3.3 Run `make ctrl` in `apps/lina-core`, regenerate controller skeletons, and implement new `BatchDelete` methods in `controller/user/` and `controller/role/` by delegating to services
- [x] 3.4 Add `BatchDelete(ctx, ids []int) error` to `apps/lina-core/internal/service/user/user.go`: reuse all `Delete` protections in one transaction and notify access topology once after commit
- [x] 3.5 Add `BatchDelete(ctx, ids []int) error` to `apps/lina-core/internal/service/role/role.go`: reuse all `Delete` protections in one transaction and notify after commit
- [x] 3.6 Add service-layer batch delete tests for success, built-in admin rejection, current-user rejection, and empty-list validation failure

## 4. Backend performance and default-value changes

- [x] 4.1 Rewrite `isDescendant` in `apps/lina-core/internal/service/menu/menu.go`: load the parent-child map once with `dao.SysMenu.Ctx(ctx).Scan(&all)` and then run in-memory BFS/DFS; keep the function signature unchanged
- [x] 4.2 Remove the `defaultManagedJobTimezone = "Asia/Shanghai"` constant from `apps/lina-core/internal/service/cron/cron_managed_jobs.go`; read `scheduler.defaultTimezone` and default to `UTC`
- [x] 4.3 Add `scheduler.defaultTimezone: "UTC"` and English comments to `apps/lina-core/manifest/config/config.template.yaml`
- [x] 4.4 Add service-layer unit tests for `isDescendant` correctness boundaries: self is not descendant, cross-depth match, and missing ids

## 5. Upload route authorization and operational endpoints

- [x] 5.1 Move `GET /api/v1/uploads/*` into `api/file/v1` and `internal/controller/file`, and mount it with the file controller under the protected route group with Auth and Permission middleware
- [x] 5.2 Use the existing `system:file:download` permission tag for the upload access route and enforce it through the unified permission middleware
- [x] 5.3 Add a standard API/controller `GET /api/v1/health` route: run `dao.SysUser.Ctx(ctx).Limit(1).Count()` as a probe; return `{status:"ok", mode:"<single|master|slave>"}` when healthy, and `503 {status:"unavailable", reason:"database probe failed"}` on failure
- [x] 5.4 Add `health.timeout: "5s"` and `shutdown.timeout: "30s"` to `apps/lina-core/manifest/config/config.template.yaml`; parse them as `time.Duration`
- [x] 5.5 Use GoFrame `Server.Run()` in `apps/lina-core/internal/cmd/cmd_http.go` for built-in signal handling and HTTP graceful shutdown; after `Server.Run()` returns, clean up host-owned resources in order: Cron scheduler `Stop`, cluster service `Stop`, and database pool `Close`, bounded by `shutdown.timeout`
- [x] 5.6 If `internal/service/cron` has no `Stop(ctx)` method, add it in the component main file so registered jobs stop accepting new triggers and wait for in-flight jobs during shutdown

## 6. Stale empty package cleanup

- [x] 6.1 Grep the repository and confirm `apps/lina-core/pkg/auditi18n` and `apps/lina-core/pkg/audittype` have no imports
- [x] 6.2 Delete empty directories `apps/lina-core/pkg/auditi18n/` and `apps/lina-core/pkg/audittype/`

## 7. Frontend batch operations and adaptation

- [x] 7.1 Add `userBatchDelete(ids: number[])` in `apps/lina-vben/apps/web-antd/src/api/system/user/index.ts`, using repeated `ids` query parameters
- [x] 7.2 Add `roleBatchDelete(ids: number[])` in `apps/lina-vben/apps/web-antd/src/api/system/role/index.ts`, using repeated `ids` query parameters
- [x] 7.3 Modify batch delete in `apps/lina-vben/apps/web-antd/src/views/system/user/index.vue`: replace the loop over `userDelete(id)` with one `await userBatchDelete(ids)`, keeping existing generic bizerr display
- [x] 7.4 Modify batch delete in `apps/lina-vben/apps/web-antd/src/views/system/role/index.vue`: replace with `roleBatchDelete(ids)`

## 8. Frontend polling, cache, and language-switching optimizations

- [x] 8.1 Modify `apps/lina-vben/apps/web-antd/src/views/monitor/server/index.vue`: use `useIntervalFn` plus `useDocumentVisibility` from `@vueuse/core` for 30 second auto-refresh; pause while hidden, refresh once immediately when visible, and explicitly stop on unmount
- [x] 8.2 Modify `apps/lina-vben/apps/web-antd/src/store/message.ts`: replace raw `setInterval` with visibility-aware polling, pause while hidden, refresh once immediately when visible, and stop on logout or store disposal
- [x] 8.3 Modify `loadedPaths` in `apps/lina-vben/apps/web-antd/src/router/guard.ts` to a bounded LRU with default limit 50; move hits to the tail and evict oldest entries on overflow
- [x] 8.4 Modify `watch(preferences.app.locale, ...)` in `apps/lina-vben/apps/web-antd/src/bootstrap.ts`: keep `syncPublicFrontendSettings` and `useDictStore().resetCache()`, remove `refreshAccessibleState(router)`, and update menu titles locally through `meta.i18nKey` and runtime language packs without calling `/menus/all`
- [x] 8.5 Scan all `meta.title` definitions under `apps/lina-vben/apps/web-antd/src/router/routes/modules/` and ensure they use i18n keys or functions returning `$t(...)`; fix static hard-coded strings if found

## 9. Unit tests, E2E, and regression verification

- [x] 9.1 Create `hack/tests/e2e/auth/TC0147-health-endpoint-anonymous-access.ts`: anonymous `GET /api/v1/health` returns 200 and `status="ok"`, following `lina-e2e` conventions
- [x] 9.2 Create `hack/tests/e2e/iam/user/TC0148-user-batch-delete-single-request.ts`: selecting multiple users in the user list sends exactly one `DELETE /api/v1/user?ids=...` request and removes data in batch
- [x] 9.3 Create `hack/tests/e2e/iam/role/TC0149-role-batch-delete-single-request.ts`: role list batch delete uses one endpoint call
- [x] 9.4 Create `hack/tests/e2e/monitor/server/TC0150-server-monitor-visibility-aware-polling.ts`: verify 30 second refresh while visible, stop while hidden, and immediate refresh on restore
- [x] 9.5 Create `hack/tests/e2e/settings/file/TC0151-uploads-route-requires-auth.ts`: unauthenticated upload access is rejected, authenticated without permission is forbidden, and authenticated with permission can read an uploaded file through the storage backend
- [x] 9.6 Add a case under `hack/tests/e2e/i18n/TC0152-language-switch-no-user-info-reload.ts`: after language switching, menu titles update reactively and `/api/v1/user/info` or `/api/v1/menus/all` are not refetched
- [x] 9.7 Run `go test ./...` in `apps/lina-core` and confirm service-layer tests pass
- [x] 9.8 Run `pnpm test` in `hack/tests` and confirm all E2E tests pass

## 10. Review and archive readiness

- [x] 10.1 Run `/lina-review` for a full change review covering code, SQL, E2E, and specification compliance
- [x] 10.2 Append and complete repair tasks in `tasks.md` based on review findings; sync any spec deltas if behavior changed
- [x] 10.3 Before archive, rerun `openspec validate code-quality-improvements --strict` and `make test`, confirming no regressions

## Feedback

- [x] **FB-1**: Normalize SQL line comments so each comment uses English above Chinese on separate lines
- [x] **FB-2**: Move upload file access routing into the file API module and reuse file storage access logic
- [x] **FB-3**: Remove redundant custom HTTP signal handling and rely on GoFrame Server.Run graceful shutdown
- [x] **FB-4**: Split cmd_http.go by responsibility to reduce single-file complexity
- [x] **FB-5**: Change health probe default timeout to 5s
- [x] **FB-6**: Split config Service into categorized embedded interfaces
- [x] **FB-7**: Keep plugin install SQL seed-only and avoid cleanup DELETE statements in install scripts
- [x] **FB-8**: Split middleware Service into HTTP middleware and non-middleware support interfaces
- [x] **FB-9**: Harden dict E2E deletion targeting so tests cannot soft-delete built-in system dictionaries
- [x] **FB-10**: Reduce redundant database reads and writes during lina-core startup reconciliation
- [x] **FB-11**: Make persistent cron registration idempotent when startup handler restoration refreshes the same job
- [x] **FB-12**: Load small plugin and menu governance tables as startup snapshots to avoid N+1 reconciliation queries
- [x] **FB-13**: Load scheduled-job startup governance rows as snapshots to avoid built-in job reconciliation N+1 queries
