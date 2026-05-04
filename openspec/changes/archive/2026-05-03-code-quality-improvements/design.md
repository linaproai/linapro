## Context

`temp/codebase-review-report.md` produced a verified fix list spanning backend, SQL, frontend, and operations. The items are independent, have limited blast radius, and already have a clear current-state plus expected-behavior conclusion, making them suitable for one focused host hardening iteration.

The following active iterations already own adjacent work:

- `backend-hardcoded-chinese-i18n-governance`: backend Chinese literals and i18n governance.
- `plugin-api-query-performance`: plugin list query side effects and session activity write throttling.
- `regression-feedback-localization-ui`: English frontend copy/layout and server-side default interval regressions.

This change must avoid overlapping those iterations and focus on host baseline hardening: transactional correctness, SQL performance and consistency, batch APIs and frontend integration, and host runtime observability and operability.

Constraints:

1. The project has no legacy burden; SQL is modified in place and verified by `make init`.
2. All SQL must remain idempotent.
3. Backend business errors must be wrapped with `bizerr`.
4. Backend durations must use `time.Duration`, and configuration uses unit-bearing strings such as `"30s"`.
5. All API documentation source text uses English.
6. Active-iteration documents were originally written in Chinese; archive documents are normalized to English.

## Goals / Non-Goals

**Goals:**

- Eliminate identified transactional breakpoints in user deletion, role deletion, menu deletion, and `AssignUsers`, preventing partial success and orphaned data.
- Close the identified security gap: upload file routes must pass unified authentication and authorization.
- Add identified SQL performance and consistency fixes: common query indexes, dictionary soft-delete fields, and scheduled-job foreign-key replacement.
- Replace the frontend anti-pattern of looping over single-delete APIs with real batch APIs.
- Add `/health` and graceful shutdown to meet the minimum operational requirements for containerized deployment.
- Replace the hard-coded Shanghai timezone with configuration and remove misleading empty packages.

**Non-Goals:**

- Full audit-log modeling and persistence.
- API rate limiting, TraceID middleware, request cancellation, Vue global error boundary, and similar cross-cutting infrastructure.
- DI containerization and `cmd_http.go` controller assembly refactoring.
- i18n translation-key coverage governance and backend hard-coded Chinese cleanup.
- Dictionary-management spec changes; the spec is already correct and only SQL implementation needs alignment.

## Decisions

### D1: Deletion transactions converge on `dao.Xxx.Transaction`

User, role, and menu deletion plus associated cleanup are placed in one transaction closure. Any failure inside the transaction returns the error and rolls back. Notifications such as `NotifyAccessTopologyChanged` run after commit.

**Why not only add retries:** retries do not provide atomicity, and association cleanup failures usually indicate external state problems. Rolling back is safer than repeatedly logging warnings.

**Alternative:** compensation transactions or Saga. That is overdesigned for this project scale; GoFrame transaction closures cover the need.

### D2: `AssignUsers` uses one transaction and batch insert

Collect new associations, construct `[]do.SysUserRole`, and execute one `Insert(slice)`. Any failure rolls back the whole operation.

**Why not chunking:** one assignment operation is constrained by UI selection size. At the expected scale under 1000 rows, one insert has comparable cost and keeps the transaction simpler.

### D3: Upload file access belongs to the file module

`GET /api/v1/uploads/*` is maintained by standard DTOs in `api/file/v1` and controller code under `internal/controller/file`, and is mounted with file controller routes under the protected static API group. The permission tag reuses existing file download permission `system:file:download` and is enforced by unified Auth and Permission middleware. The controller must not concatenate local upload directories or call `ServeFile` directly; it passes the relative storage path to the file service, resolves metadata, and reads the stream through the configured storage backend. This keeps local storage, distributed deployment, and object storage behavior consistent.

**Why not signed-on-demand links:** accessing uploaded files is a business action that should go through unified authorization and audit. Signed URL mode is out of scope. Anonymous access remains reserved for explicitly public download scenarios, which are not required now.

### D4: SQL indexes and foreign-key replacement

| Table | Change | Reason |
|---|---|---|
| `sys_user_role` | `KEY idx_role_id (role_id)` | Common paths: query users by role and delete associations by role |
| `sys_role_menu` | `KEY idx_menu_id (menu_id)` | Cascade-delete path for menu-related role associations |
| `sys_user` | `KEY idx_status / idx_phone / idx_created_at` | Status filters, phone search, and created-time range filters |
| `sys_online_session` | `KEY idx_last_active_time (last_active_time)` | Timeout session cleanup |
| `sys_dict_type / sys_dict_data` | `deleted_at DATETIME DEFAULT NULL` | Align soft-delete behavior with other business tables and the existing `dict-management` spec |
| `sys_job` | Remove `fk_sys_job_group_id`, add `KEY idx_group_id` | Align with other association tables and reduce foreign-key lock overhead in high-concurrency scheduler paths |

**Why `KEY` instead of `INDEX`:** existing project SQL style uses `KEY`, so this follows the repository convention.

**Why soft delete for dictionary tables instead of hard delete:**

1. The `dict-management` spec already declares `deleted_at` in the table design.
2. Dictionary types and data are widely referenced. Hard deletion can make historic logs lose label interpretation, while soft delete preserves audit recovery.

### D5: Menu `isDescendant` uses in-memory traversal

Load all menus once with `dao.SysMenu.Ctx(ctx).Scan(&all)`, build `parentChildren := map[int][]int`, and run BFS to determine whether `targetId` is under `parentId`. Complexity changes from per-depth SQL round trips to one `O(N)` load and in-memory traversal.

**Why not add an explicit path column:** that requires maintaining path data in every create and move API, expanding the scope. Current data size does not justify a path-column design over in-memory traversal.

### D6: RESTful batch delete design

```text
DELETE /api/v1/user?ids=1&ids=2&ids=3
DELETE /api/v1/role?ids=1&ids=2&ids=3
```

- Use `DELETE` with repeated query parameters for `ids`; existing scheduler-management APIs already use similar batch operation patterns.
- DTO `BatchDeleteReq` uses `Ids []int json:"ids" v:"required|min-length:1" dc:"..." eg:"1"`.
- Service layer implements `BatchDelete(ctx, ids)`: one transaction reuses all protections from single delete, including built-in admin, current user, and role reference checks. Any `bizerr` rejects the whole batch.
- Frontend `userBatchDelete` and `roleBatchDelete` issue one request, and UI errors use the existing unified bizerr handling.

**Why not forward to the single-delete API:** that would still create N transactions and cannot guarantee whole-batch rollback.

**Why query parameters instead of a body:** this matches the project style for query-string resource selection and keeps the `DELETE` semantics bodyless with good browser and middleware compatibility.

### D7: Server monitor automatic refresh and visibility awareness

Use `useIntervalFn` plus `useDocumentVisibility` from `@vueuse/core`: start 30 second polling while visible, pause while hidden, and explicitly stop on component unmount.

`store/message.ts` polling follows the same visibility-aware model: pause while hidden and trigger one `fetchUnreadCount` immediately when visible again.

**Why not clean up on route switching only:** message polling is global and independent of routes. Page visibility is the appropriate control signal.

### D8: `loadedPaths` becomes a bounded LRU

Replace `Set<string>` with a simple LRU using `Map`, a size limit, and `delete + set` on hits. The threshold is 50 entries; the oldest entry is evicted when the limit is exceeded.

**Why not `WeakMap`:** keys are string paths, not objects.

### D9: Language switching does not reload full permissions

In the `watch(preferences.app.locale, ...)` path in `bootstrap.ts`:

- Keep `syncPublicFrontendSettings(locale)` because public configuration depends on locale.
- Keep `useDictStore().resetCache()` because dictionary cache is keyed by language.
- Remove `refreshAccessibleState(router)`.
- Add `meta.i18nKey` to backend menu route responses, preserve it in the frontend menu model, and redraw menu titles locally through runtime language packs after language switching without requesting `/api/v1/user/info` or `/api/v1/menus/all`.
- Add defensive scanning: menu and route `meta.title` must use `$t(...)` or an i18n key, not a string evaluated once at startup. Hard-coded cases are converted to reactive access.

### D10: Host runtime operations as new `host-runtime-operations` capability

#### `/api/v1/health` health probe

- Public route with no authentication.
- Implementation: run `dao.SysUser.Ctx(ctx).Limit(1).Count()` as a DB probe; exceeding `health.timeout` (default `5s`) marks the service unavailable.
- Response: `200 {status:"ok"}` or `503 {status:"unavailable", reason:"database probe failed"}`. Internal errors are logged but not exposed to anonymous callers.

#### Graceful shutdown

- HTTP entry uses GoFrame `Server.Run()` and reuses built-in `SIGTERM` / `SIGINT` signal handling and HTTP graceful shutdown.
- `cmd_http.go` no longer registers `os/signal` and no longer directly repeats HTTP server `Shutdown()`.
- After `Server.Run()` returns, host-owned runtime resources are cleaned up in order:
  1. call `cronSvc.Stop(ctx)`;
  2. stop the cluster service;
  3. close the database connection pool.
- Runtime cleanup is bounded by `shutdown.timeout` (default `30s`); timeout returns an error and logs a warning.

#### Upload route protection

See D3.

#### Stale empty package cleanup

Delete `apps/lina-core/pkg/auditi18n/` and `apps/lina-core/pkg/audittype/`. Verify with grep that no imports remain, and fix any references during removal.

#### Configurable scheduler default timezone

`defaultManagedJobTimezone` in `cron_managed_jobs.go` is no longer hard-coded. It is read from configuration key `scheduler.defaultTimezone` and defaults to `UTC`. `config.template.yaml` adds `scheduler.defaultTimezone: "UTC"` with English comments and README documentation.

### D11: Split host foundation service interfaces by responsibility

`config.Service` remains the full configuration service composite, but no longer flattens all methods directly. It embeds smaller interfaces for cluster, auth, login, frontend, i18n, cron, host runtime, delivery metadata, plugin, upload, and runtime parameter synchronization. This lets callers depend on narrower reader/syncer interfaces in future refactors and tests.

`middleware.Service` also remains compatible for callers, but is split into:

- `HTTPMiddleware`: methods and factories that can be directly installed into GoFrame route groups.
- `RuntimeSupport`: non-middleware support methods such as `SessionStore()` and `PublishedRouteMiddlewares()`.

These splits do not change runtime behavior and do not add user-visible copy or API contracts, so i18n resources are not required.

## Risks / Trade-offs

- [SQL reinitialization disrupts local development data] Mitigation: tasks explicitly use `modify SQL -> make init`; review checks that no migration-path assumptions remain.
- [Protected upload routes break third-party image references] Mitigation: current pages and plugins access uploads with login state; there is no product requirement for anonymous access. Review samples all `<img src="/api/v1/uploads/...">` uses.
- [Stricter deletion transactions increase visible errors] Mitigation: this is intentional because errors were previously swallowed. E2E covers rollback UI behavior when association deletion fails.
- [Menu `isDescendant` loads all menus once] Mitigation: menu count is far below 1000; memory cost is negligible. A path-column design can be revisited for future large-scale menus.
- [Removing `sys_job` foreign key shifts consistency to application code] Mitigation: job write paths already validate `group_id` in the service layer; review confirms all writes go through that layer.
- [Graceful shutdown timeout handling] Mitigation: HTTP graceful shutdown remains GoFrame-owned; host-owned cleanup is bounded by `shutdown.timeout` (default `30s`) and logs warnings on timeout.
- [`/health` DB probe adds baseline QPS] Mitigation: `Limit(1).Count()` is lightweight, and Kubernetes probe intervals are normally at least 10 seconds.

## Migration Plan

1. Before merge in development:
   - Run `make init` to rebuild the database and verify idempotency plus new indexes and `deleted_at` columns.
   - Run `make dao` to regenerate entities and verify `SysDictType` / `SysDictData` include `DeletedAt`.
   - Run `make ctrl` to regenerate controller skeletons for user and role batch delete APIs.
2. Deployment: because the foreign key is removed, an old database would need `ALTER TABLE sys_job DROP FOREIGN KEY fk_sys_job_group_id;`. Since this project has no legacy burden, use `make init` to reinitialize instead of adding migration scripts.
3. Container deployments should update:
   - `livenessProbe` / `readinessProbe` to `/api/v1/health`;
   - `terminationGracePeriodSeconds >= shutdown.timeout` (default 30 seconds).
4. Monitor cron job timezone: explicitly set `scheduler.defaultTimezone` in config or use the UTC default.

## Open Questions

- Should the final upload route permission tag be `file:read` or a more specific `file:download`? The apply phase must compare existing file module menu/button permissions and avoid permission-code conflicts.
- Should `/health` return primary/secondary role information when `cluster.enabled=true`? For this iteration it returns `{status, mode}` where `mode` is `master` or `slave`, but does not implement a leader-switching probe to avoid overlapping `leader-election`.
