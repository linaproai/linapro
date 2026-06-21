## 1. Manifest and Persistence Model

- [x] 1.1 Add`PluginDistribution`named type, constants, normalization and validation helpers in plugin internal types.
- [x] 1.2 Add`distribution`to catalog manifest parsing, validation and manifest tests, including default, invalid value,`builtin + dynamic`and`builtin + unregistered source`cases.
- [x] 1.3 Add`sys_plugin.distribution`to the host plugin registry baseline schema and regenerate DAO/DO/Entity artifacts as required.
- [x] 1.4 Persist`distribution`in plugin registry synchronization and release manifest snapshot serialization/deserialization.
- [x] 1.5 Add or update store tests proving manifest sync writes`builtin`to registry and release snapshots.

## 2. API and Management Projection

- [x] 2.1 Add`distribution`to plugin list/detail DTOs and generated controller/API artifacts.
- [x] 2.2 Add a guarded read-only include-builtin query path for diagnostics, while default list queries hide`builtin`.
- [x] 2.3 Update management read model, cache clone and projection tests to include`distribution`without adding list-time manifest parsing.
- [x] 2.4 Add service-side guards that reject ordinary management install, enable, disable, uninstall, upgrade and tenant provisioning policy updates for`builtin`.
- [x] 2.5 Add stable`bizerr`code and host runtime/apidoc i18n resources for builtin management action denial.

## 3. Startup Builtin Reconciliation

- [x] 3.1 Add a dedicated`BootstrapBuiltinPlugins(ctx)`or equivalent startup entry separate from`BootstrapAutoEnable`.
- [x] 3.2 Implement builtin dependency ordering and source-only install/enable reconciliation using existing lifecycle paths without loading mock data.
- [x] 3.3 Reuse the unified source upgrade execution path for startup safe upgrades and preserve failure diagnostics.
- [x] 3.4 Wire startup ordering as source manifest sync, builtin reconciliation,`plugin.autoEnable`, tenant provisioning and runtime/front-end/cron wiring.
- [x] 3.5 Validate cluster behavior so only primary performs shared lifecycle writes and non-primary nodes wait for shared convergence.
- [x] 3.6 Record DI source checks for any new runtime dependencies or explicitly record no new dependency impact.

## 4. Frontend Management UI

- [x] 4.1 Update plugin API types and list query defaults so the ordinary plugin management page does not request builtin diagnostics.
- [x] 4.2 Hide install, enable, disable, uninstall, manual upgrade and tenant provisioning actions when a returned plugin has`distribution=builtin`.
- [x] 4.3 Add or update focused UI/E2E coverage for default hidden builtin plugins and read-only builtin detail behavior, or document an equivalent verification boundary.

## 5. Verification and Governance

- [x] 5.1 Run SQL idempotency/static checks, DAO generation verification and changed backend package tests.
- [x] 5.2 Run API/controller generation or compile checks covering plugin API signature and route binding changes.
- [x] 5.3 Run cache, lifecycle, management and startup package tests covering changed behavior.
- [x] 5.4 Run frontend type checks or targeted tests if frontend files changed.
- [x] 5.5 Run`openspec validate builtin-plugin-distribution --strict`.
- [x] 5.6 Record impact analysis for`i18n`, cache consistency, data permission, SQL, dev tooling cross-platform impact and E2E quality review.

## Verification and Impact Notes

- SQL and DAO: `sys_plugin.distribution` is part of the fresh-project baseline schema in `apps/lina-core/manifest/sql/008-plugin-framework.sql`. `make dao` regenerated only `sys_plugin` DAO/DO/Entity distribution fields. No separate compatibility migration, Seed DML, Mock data, explicit auto-increment IDs, or soft-delete/time-maintenance changes were added by this change.
- API/controller generation: `make ctrl` was run for plugin API DTO changes. The generator also rewrote import aliases in API wrapper files and attempted to append duplicate no-arg controller constructors for dependency-injected controllers; the duplicate stubs were removed, leaving the existing explicit DI constructors intact.
- Backend tests: passed `go test ./internal/service/plugin/internal/store -count=1`, `go test ./internal/service/plugin/internal/lifecycle -count=1`, `go test ./internal/controller/plugin -count=1`, `go test ./internal/service/plugin -count=1`, `go test ./internal/cmd ./internal/cmd/internal/httpstartup -count=1`, `go test ./pkg/plugin/capability/capmodel ./pkg/plugin/pluginbridge/contract ./pkg/plugin/pluginhost -count=1`, and focused distribution/bootstrap test runs. `go test ./internal/service/plugin/internal/runtime -count=1` exposed an order-sensitive existing route-dispatch failure; focused reruns of `TestDispatchDynamicRouteAllowsPluginOwnedPathShapes` and related runtime subsets passed.
- Frontend and E2E: passed `pnpm -F @lina/web-antd typecheck` and `pnpm --dir hack/tests test:validate`. Added `hack/tests/e2e/extension/plugin/TC016-builtin-plugin-management-readonly.ts` for default hidden builtin list behavior and diagnostic read-only action hiding. Direct Playwright execution was attempted but blocked in global login at`http://127.0.0.1:9120/admin/auth/login?redirect=%252Fauth%252Flogin`; screenshots were therefore not produced in this environment.
- i18n: runtime error translations were added in host `en-US` and `zh-CN` error resources, and `zh-CN` apidoc translations were added for `includeBuiltin` and `distribution`. Frontend UI hiding uses existing labels and adds no new user-visible runtime text. JSON parsing validation passed for touched host i18n/apidoc JSON.
- Cache consistency: builtin install/enable/upgrade reuse existing lifecycle/upgrade cache publication, enabled snapshot refresh, runtime revision, and management-list cache key invalidation. Startup builtin reconciliation now uses the projection startup snapshot context for lifecycle convergence, avoiding redundant governance reads while preserving snapshot updates after writes. Cluster behavior remains primary-writes/non-primary-waits through the existing lifecycle topology policy.
- Data permission: plugin management is platform governance data. Default list queries hide `builtin`, diagnostic inclusion is read-only, and service-side write guards reject ordinary lifecycle mutations before side effects. No new business-data query path or tenant/organization data exposure was introduced.
- Dev tooling cross-platform: no long-lived scripts, CI entries, Makefile targets, or linactl commands were added or changed by this OpenSpec change. Existing cross-platform generator entries `make ctrl` and `make dao` were executed.
- DI source check: no new runtime service dependency owner or constructor parameter was added. The new root `BootstrapBuiltinPlugins(ctx)` method reuses existing `lifecycleSvc`, `upgradeSvc`, `integrationSvc`, `configSvc`, and startup `pluginSvc` instances; lifecycle receives the already-injected `store`, `catalog`, `integration`, `runtime`, cache publisher, and topology services. Startup wiring calls the existing shared `pluginSvc` instance.
- E2E quality review: triggered because the plugin management UI has user-observable hiding behavior. Coverage asset `TC016` asserts no `includeBuiltin` query parameter on the ordinary page, builtin rows absent from default list, diagnostic builtin rows hiding enable, tenant provisioning, install, upgrade, and uninstall actions, and detail remaining readable with i18n key checks. Live browser execution is pending a working authenticated E2E environment.
- Plugin docs and rules: `.agents/rules/plugin.md`, `apps/lina-core/pkg/plugin/README.md`, `apps/lina-core/pkg/plugin/README.zh-CN.md`, `apps/lina-plugins/README.md`, `apps/lina-plugins/README.zh-CN.md`, and the `linapro-demo-source` sample manifest/docs now document `distribution`, the `marketplace` default, the `builtin` source-only constraint, and startup-managed lifecycle semantics.

## Feedback

- [x] **FB-1**: 删除`distribution`独立兼容迁移并清理旧记录兜底实现
- [x] **FB-2**: 为所有插件`plugin.yaml`显式补充`distribution`配置项和注释

### FB-1 Verification and Impact

- Root cause: `apps/lina-core/manifest/sql/013-builtin-plugin-distribution.sql` was an `ALTER TABLE ... ADD COLUMN` compatibility migration for an existing`sys_plugin`table. The project has no compatibility burden, so`distribution`belongs in the fresh baseline schema instead of a separate backfill migration.
- SQL: `013-builtin-plugin-distribution.sql` was deleted. `sys_plugin.distribution` remains in`apps/lina-core/manifest/sql/008-plugin-framework.sql`as a baseline`CREATE TABLE IF NOT EXISTS`field. No Seed DML, Mock data, explicit auto-increment ID writes, soft-delete changes, or separate backfill were introduced.
- Backend: release`manifest_snapshot` parsing now requires explicit canonical`distribution`and rejects missing, invalid, or non-canonical persisted values. Runtime registry projection and plugin list filtering no longer default empty registry/snapshot/item distribution values to`marketplace`; only manifest input keeps the intentional omitted-value default.
- Documentation/spec: OpenSpec design/tasks/spec text and`localdocs/builtin-plugin-distribution-design.md`now describe baseline SQL and explicitly reject independent compatibility migration/backfill.
- i18n: no new runtime UI text, API documentation text, plugin manifest text, or language resources were added by this feedback fix.
- Cache consistency: no new cache or invalidation path was added. Existing plugin lifecycle cache refresh behavior is unchanged.
- Data permission: no new read/write data operation or tenant/organization boundary was added.
- Dev tooling cross-platform: no Makefile, script, CI,`linactl`, or long-lived tooling entry changed.
- DI source check: no new runtime dependency owner, constructor parameter, or service graph path was added.
- Verification passed: `go test ./internal/service/plugin/internal/store -count=1`; `go test ./internal/service/plugin/internal/runtime ./internal/service/plugin ./internal/controller/plugin ./pkg/plugin/capability/capmodel ./pkg/plugin/pluginbridge/contract ./pkg/plugin/pluginhost -count=1`; `go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/lifecycle ./internal/service/plugin/internal/management ./internal/cmd ./internal/cmd/internal/httpstartup -count=1`; `go test ./internal/service/plugin/internal/... -count=1`; `git diff --check`; `openspec validate builtin-plugin-distribution --strict`; static scan for`013-builtin`,`ALTER TABLE.*distribution`, legacy snapshot/defaulting patterns, and registry/snapshot/item distribution normalization.

### FB-2 Verification and Impact

- Root cause: `distribution` was supported by the manifest parser and documented in plugin workspace docs, but most checked-in plugin manifests still relied on the implicit `marketplace` default. Developers reading a plugin's `plugin.yaml` could not see the lifecycle governance choice locally.
- Plugin manifests: all 11 files under`apps/lina-plugins/*/plugin.yaml` now declare`distribution: marketplace`immediately after`type`, with comments describing`marketplace`versus`builtin`. `linapro-ops-demo-guard` explicitly remains`marketplace`so demo read-only protection still requires plugin management or`plugin.autoEnable` and is not auto-enabled by builtin startup reconciliation.
- Behavior: no plugin was changed to`distribution: builtin`; runtime startup, plugin list visibility, write guards, SQL, API DTOs, and frontend behavior remain unchanged by this feedback fix.
- i18n: comments were added to YAML only. No runtime user-visible text, API documentation source text, plugin i18n resource, or apidoc translation changed.
- Cache consistency: no new cache, invalidation path, runtime snapshot, or cross-node behavior changed. All plugins remain`marketplace`, so builtin lifecycle cache publication paths are not newly exercised by these manifests.
- Data permission: no new read/write API, business data query, tenant/organization visibility path, or plugin host-service data access path changed.
- Dev tooling cross-platform: no Makefile, script, CI,`linactl`, Node script, or shell entry changed.
- DI source check: no runtime dependency owner, constructor parameter, service graph path, or startup assembly changed.
- Testing/E2E: no user-observable UI behavior or E2E asset changed; this is a governance manifest clarification, so static manifest checks and OpenSpec validation are the matching verification.
- Verification passed: static scan confirmed every`apps/lina-plugins/*/plugin.yaml`has a supported explicit`distribution`; Ruby YAML parsing loaded all plugin manifests; `git diff --check` passed inside`apps/lina-plugins`; `openspec validate builtin-plugin-distribution --strict` passed.
