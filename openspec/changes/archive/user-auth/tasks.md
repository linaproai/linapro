## 1. Governance Implementation

- [x] 1.1 Add the API response instant-field contract to the API design rules in `AGENTS.md`, including Unix timestamp milliseconds, DTO types, prohibited types, and calendar-date exceptions.
- [x] 1.2 Add the time-field manual review requirement to the RESTful/API DTO checklist in the `lina-review` skill.

## 2. Response DTO Migration

- [x] 2.1 Implement existing host and source-plugin public response DTO instant fields using Unix millisecond timestamp values (`int64`/`*int64`) at the HTTP JSON boundary, with internal `time.Time`, `*gtime.Time`, and legacy stored strings projected through `apps/lina-core/pkg/apitime`.
- [x] 2.2 Update affected frontend API types and all reviewed page display points to use `formatTimestamp`, while keeping date-range query parameters as strings.
- [x] 2.3 Update host and plugin `zh-CN` `apidoc i18n JSON` resources, including packed host copies, so migrated instant fields state the Unix millisecond timestamp unit.

## 3. GoFrame DAO Generation Migration

- [x] 3.1 Configure GoFrame DAO generation with `stdTime: true` plus timestamp `typeMapping` entries so regenerated PostgreSQL timestamp fields use `*time.Time` instead of `*gtime.Time`.
- [x] 3.2 Regenerate host and source-plugin DAO/DO/entity artifacts and adjust internal services, plugin host contracts, session/cache/lock/notification/user-message/job/plugin lifecycle code, and affected tests to consume `*time.Time`.
- [x] 3.3 Verify no `*gtime.Time`, `gtime.Time`, or `gtime.NewFromTime` matches remain under `apps/lina-core` and `apps/lina-plugins`.

## 4. Verification and Records

- [x] 4.1 Run `openspec validate` to validate the OpenSpec change.
- [x] 4.2 Run `cd apps/lina-core && go test ./pkg/apitime ./internal/controller/... ./internal/service/menu ./internal/service/role ./internal/service/sysinfo ./internal/service/plugin/internal/runtime -count=1`.
- [x] 4.3 Run `cd apps/lina-core && go test ./internal/cmd -count=1`.
- [x] 4.4 Run source-plugin tests: `GOWORK=off go test lina-plugin-content-notice/backend/... lina-plugin-monitor-loginlog/backend/... lina-plugin-monitor-operlog/backend/... lina-plugin-monitor-online/backend/... lina-plugin-monitor-server/backend/... lina-plugin-multi-tenant/backend/... lina-plugin-org-center/backend/... lina-plugin-demo-source/backend/... -count=1`.
- [x] 4.5 Run `pnpm -F @lina/web-antd run typecheck`.
- [x] 4.6 Record the impact assessment for `i18n`, cache consistency, data permissions, runtime APIs, frontend pages, and development scripts.
- [x] 4.7 Complete implementation and call `lina-review` for code, specification, i18n, cache consistency, and data permission review.

## Records

- i18n impact: runtime UI language packs and manifest runtime language resources were not changed; API documentation i18n resources were changed intentionally for migrated time fields.
- Cache consistency impact: no cache key, invalidation, distributed coordination, or freshness model changed; only expiration timestamp representation changed from `*gtime.Time` to `*time.Time`.
- Data-permission impact: no data access scope was widened or changed; existing list/detail/query paths only project response field types.
- Runtime API impact: existing public response field types for instant fields changed from formatted/time-object-like values to numeric millisecond timestamps.
- Frontend page impact: system config/file/job/job-group/job-log/menu/message/plugin/role/role-auth/user/profile/notification pages and affected source-plugin pages were reviewed and updated where they display migrated fields.
- Development script impact: no static scanning tool, build script, test script, or default developer command was added or modified.
