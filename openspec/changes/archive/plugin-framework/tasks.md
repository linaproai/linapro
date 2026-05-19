## 1. Plugin Contract and Metadata Foundation

- [x] 1.1 Define `plugin.yaml` manifest schema, version strategy, and host validation flow
- [x] 1.2 Establish `apps/lina-plugins/<plugin-id>/` directory convention with `plugin-demo-source` and `plugin-demo-dynamic` as reference samples
- [x] 1.3 Create plugin metadata SQL: `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`
- [x] 1.4 Generate DAO/DO/Entity and build plugin registration, lifecycle, resource reference, and migration service skeletons
- [x] 1.5 Define plugin management API, DTOs, management page structure, and state machine enumerations

## 2. Source Plugin Integration

- [x] 2.1 Implement source plugin scanning, registry sync, frontend resource discovery, and backend registration via `lina-plugins.go`
- [x] 2.2 Implement frontend source plugin manifest generation, page discovery, slot registration, and host build integration
- [x] 2.3 Implement source plugin sync, enable, disable management flow and admin UI
- [x] 2.4 Implement `plugin-demo-source` backend: plugin-directory Go source compilation, public/protected routes, governance integration
- [x] 2.5 Implement `plugin-demo-source` frontend: menu page display, host page integration, and management interaction

## 3. Governance Integration and Extension Points

- [x] 3.1 Extend menu, role, and permission pathways for plugin menu and permission reuse
- [x] 3.2 Build host backend hook bus with first-batch hooks, failure isolation, and execution observability
- [x] 3.3 Build host frontend slot registry with first-batch slots and load-failure degradation
- [x] 3.4 Implement plugin disable/re-enable/uninstall menu hiding, permission invalidation, role relationship preservation, and resource cleanup

## 4. Dynamic WASM Plugin Runtime

- [x] 4.1 Define runtime WASM artifact format, resource embedding convention, and ABI version strategy
- [x] 4.2 Implement WASM plugin installer, validator, resource extractor, and migration executor
- [x] 4.3 Implement WASM plugin loading, hook invocation, timeout control, error isolation, and unload recovery
- [x] 4.4 Implement dynamic plugin static resource hosting and three frontend integration modes
- [x] 4.5 Provide `plugin-demo-dynamic` sample and verify runtime contract and page behavior

## 5. Multi-Node Hot Upgrade and Rollback

- [x] 5.1 Build `desired_state/current_state/generation/release_id` generation model and primary-node switch flow
- [x] 5.2 Integrate leader election and node Reconciler into plugin install/enable/disable/upgrade and convergence
- [x] 5.3 Implement hot upgrade generation switch, old-request natural completion, and node state reporting
- [x] 5.4 Implement upgrade failure rollback, migration exception recovery, and frontend resource switch failure protection
- [x] 5.5 Implement frontend plugin generation awareness and current-page refresh prompt

## 6. Dynamic Plugin REST Runtime

- [x] 6.1 Define dynamic route contract structure, fixed prefix, governance fields, and validation rules
- [x] 6.2 Extract `g.Meta` route metadata from `backend/api/**/*.go` during build and embed in artifact
- [x] 6.3 Implement `/api/v1/extensions/{pluginId}/...` dispatch, route matching, auth, permission check, and context injection
- [x] 6.4 Abstract dynamic route executor interface, request/response snapshot, and v1 bridge envelope
- [x] 6.5 Embed bridge ABI contract in artifact, implement real WASM bridge execution, and update OpenAPI projection
- [x] 6.6 Add Host Functions: protocol layer, codec, capabilities, guest SDK, host dispatcher, log/state/db handlers, `sys_plugin_state` table
- [x] 6.7 Update `plugin-demo-dynamic` with host-call demo route and verify end-to-end

## 7. Dynamic Plugin Embed Snapshot Packaging

- [x] 7.1 Define unified `go:embed` resource declaration convention for dynamic plugins
- [x] 7.2 Adjust `hack/build-wasm` to prioritize embedded resource reading with directory-scan fallback
- [x] 7.3 Update `plugin-demo-dynamic` and documentation for unified resource declaration

## 8. Host Service Extension

- [x] 8.1 Define structured host-service invoke envelope, protocol version, and unified error model
- [x] 8.2 Refactor `lina_env.host_call` to unified host-service channel and implement registry/dispatcher
- [x] 8.3 Extend `plugin.yaml` and builder for `hostServices` declaration with static validation
- [x] 8.4 Write host-service governance snapshot to WASM custom section and restore at load time
- [x] 8.5 Implement `runtime` service (log, state, info)
- [x] 8.6 Implement `storage` service with `resources.paths` authorization
- [x] 8.7 Implement `network` service with URL pattern authorization
- [x] 8.8 Implement `data` service with table-level authorization, DAO/ORM execution, and `plugindb` guest SDK
- [x] 8.9 Implement `cache` service with MySQL `MEMORY` table backend
- [x] 8.10 Implement `lock` service reusing host distributed lock
- [x] 8.11 Implement `notify` service with unified notification domain tables
- [x] 8.12 Update demo, docs, and tests for all host services

## 9. Startup Auto-Enable Bootstrap

- [x] 9.1 Extend host config model and template for `plugin.autoEnable`
- [x] 9.2 Add startup bootstrap phase before plugin route/cron/bundle wiring
- [x] 9.3 Implement source-plugin auto-install and auto-enable with cluster primary protection
- [x] 9.4 Implement dynamic-plugin auto-install and auto-enable with authorization snapshot reuse
- [x] 9.5 Implement fail-fast, convergence waiting, and enabled-snapshot refresh
- [x] 9.6 Fix source plugin startup snapshot synchronization after install so that the enable phase within the same startup orchestration reads the latest installed state

## 10. Install-and-Enable Shortcut

- [x] 10.1 Add `Install and Enable` action in installation dialog with permission gating
- [x] 10.2 Wire composite install-then-enable flow reusing existing APIs
- [x] 10.3 Add E2E coverage for shortcut flow, permission boundaries, and failure messaging

## 11. Mock Data Installation

- [x] 11.1 Add `installMockData` API field and mock SQL asset discovery
- [x] 11.2 Implement transactional mock SQL execution with structured rollback errors
- [x] 11.3 Extend startup auto-enable with `withMockData` structured entries
- [x] 11.4 Add frontend checkbox, list column, help tooltip, and uninstall warning

## 12. Authorization Route Visibility

- [x] 12.1 Add route review fields to plugin-management DTOs with backend projection
- [x] 12.2 Update authorization and detail dialogs with route information section
- [x] 12.3 Add E2E coverage for public/login/permission-bound routes and empty route lists

## 13. Query Performance

- [x] 13.1 Make plugin list queries read-only with explicit sync action
- [x] 13.2 Use safe metadata lookup for host-service table comments with fallback
- [x] 13.3 Add `last_active_time` write throttling for online sessions

## 14. Cluster Deployment and Topology

- [x] 14.1 Add `cluster.enabled` and `cluster.election.*` config semantics
- [x] 14.2 Integrate cluster topology into HTTP startup, leader election, and cron scheduling
- [x] 14.3 Converge dynamic plugin state switch, Reconciler, and node projection for single/cluster modes
- [x] 14.4 Converge `plugin`/`cluster`/`election` component boundaries

## 15. Configuration Duration Unification

- [x] 15.1 Update default and template configs with duration string keys
- [x] 15.2 Refactor config service to parse duration strings to `time.Duration`
- [x] 15.3 Adjust auth, session, and monitor modules to consume `time.Duration`

## 16. Notification Domain

- [x] 16.1 Create `sys_notify_channel`, `sys_notify_message`, `sys_notify_delivery` tables
- [x] 16.2 Implement `notify.Service` and redirect `notice` publish through unified domain
- [x] 16.3 Delete `sys_user_message` and migrate `/user/message` to facade over new tables

## 17. Declarative Permission Middleware

- [x] 17.1 Design and implement declarative permission middleware for static APIs
- [x] 17.2 Migrate host and plugin controllers to declarative permission model
- [x] 17.3 Add access context caching with topology-revision-based invalidation
- [x] 17.4 Add permission coverage audit tests

## 18. Plugin Configuration Service

- [x] 18.1 Update the public interface of `apps/lina-core/pkg/pluginservice/config` to provide `Get`, `Exists`, `Scan`, basic type reads, and `Duration` reads
- [x] 18.2 Remove the `MonitorConfig` type alias and the `GetMonitor()` plugin-specific business method
- [x] 18.3 Add Go comments, error handling, and default-value semantics for generic configuration read methods
- [x] 18.4 Add private configuration loading logic inside the `monitor-server` plugin to maintain the monitor configuration structure, defaults, duration parsing, and business validation
- [x] 18.5 Migrate `monitor-server` scheduled collection registration and cleanup logic to the new generic configuration service read path
- [x] 18.6 Add the dynamic plugin `config` host service constants, capability derivation, codec, guest helpers, and host dispatcher
- [x] 18.7 Add unit tests for `pluginservice/config` covering arbitrary key reads, missing key defaults, struct scanning, basic type reads, and duration parsing
- [x] 18.8 Add unit tests for `monitor-server` plugin configuration loading covering defaults, overrides, invalid duration values, and business validation

## 19. Documentation and Developer Tools

- [x] 19.1 Write plugin development guide covering source and dynamic WASM modes
- [x] 19.2 Write plugin operations guide covering install/stop/uninstall/upgrade/rollback/multi-node
- [x] 19.3 Use `plugin-demo-source` and `plugin-demo-dynamic` as reference samples (no separate template directory)
- [x] 19.4 Provide `hack/build-wasm` builder tool with unified output to `temp/output/`

## 20. E2E and Acceptance Verification

- [x] 20.1 `TC0066-source-plugin-lifecycle`: sync, enable/disable, compilation, slot rendering
- [x] 20.2 `TC0067-runtime-wasm-lifecycle`: upload, install, enable, disable, uninstall, resource hosting, dynamic routes
- [x] 20.3 `TC0068-runtime-wasm-failure-isolation`: hook timeout/error isolation, disable/enable recovery
- [x] 20.4 `TC0069-plugin-permission-governance`: role authorization, menu visibility, permission recovery, data permission
- [x] 20.5 `TC0070-plugin-hot-upgrade`: generation switch, page refresh prompt, non-plugin-page user unaffected, rollback
- [x] 20.6 `TC0071-runtime-wasm-host-services`: core host services success and unauthorized rejection
- [x] 20.7 `TC0072-runtime-wasm-host-services-low-priority`: cache, lock, notify services
- [x] 20.8 `TC0073-plugin-host-service-authorization-review`: install/enable authorization dialog with route review
- [x] 20.9 `TC0074-plugin-management-action-permissions`: upload/install/enable/disable/uninstall permission checks
- [x] 20.10 `TC0075-runtime-wasm-lifecycle-boundaries`: uninstall cleanup and version compatibility
- [x] 20.11 `TC0103-plugin-install-enable-shortcut`: shortcut flow, permission visibility, dynamic-plugin authorization reuse

## 21. Plugin ID Normalization

- [x] 21.1 Add plugin ID parsing/validation component with runtime safety boundary (non-empty, 64-char max, kebab-case) and preserve `<author>-<domain>-<capability>` as official naming recommendation
- [x] 21.2 Rename all 10 official plugin directories, `plugin.yaml` IDs, Go modules, import paths, source registration constants, and GoFrame generation configs to `linapro-*` prefix
- [x] 21.3 Update host official plugin constants, stable menu parent mappings, `orgcap.ProviderPluginID`, `tenantcap.ProviderPluginID`, startup consistency checks, and provider detection logic
- [x] 21.4 Update plugin-owned SQL tables, indexes, constraints, mock data, uninstall SQL, DAO/DO/Entity, and service access code to match new plugin ID snake_case scope
- [x] 21.5 Update manifest menu keys, permissions, cron handlerRefs, dynamic asset paths, extension API paths, i18n/apidoc namespaces, and documentation
- [x] 21.6 Update frontend plugin management, dynamic pages, menu routes, test page objects, and fixtures
- [x] 21.7 Add governance scans confirming no old official ID residual in runtime code, configuration, tests, or active OpenSpec documents
- [x] 21.8 Run full backend Go tests, frontend typecheck, E2E, i18n scan, old ID residual scan, and OpenSpec strict validation

## 22. Plugin Dependency Management

- [x] 22.1 Extend manifest type with `dependencies.framework.version` and `dependencies.plugins[]`; sync dynamic plugin artifact manifest serialization
- [x] 22.2 Implement dependency field normalization and validation (defaults, plugin ID, version range, self-dependency, duplicates, unknown install strategy)
- [x] 22.3 Determine LinaPro framework version authoritative source and implement semver range matching
- [x] 22.4 Implement internal dependency resolver: graph construction, deterministic topological sort, self/cycle detection, hard/soft dependency classification, reverse dependency query
- [x] 22.5 Integrate dependency check into explicit install path with auto-install plan execution before target plugin install
- [x] 22.6 Integrate reverse dependency protection into uninstall path; block uninstall when downstream hard dependencies exist
- [x] 22.7 Integrate dependency resolution into `BootstrapAutoEnable` for topological auto-dependency installation
- [x] 22.8 Integrate dependency validation into source plugin upgrade and dynamic plugin install/refresh paths
- [x] 22.9 Design and implement dependency check API or extend install/detail APIs with framework version check, dependency status, auto-install plan, blockers, and reverse blockers
- [x] 22.10 Update plugin management page with dependency summary, install confirmation auto-install plan, blocker display, and uninstall downstream dependency prompt
- [x] 22.11 Add E2E `TC0235-plugin-dependency-management` covering install dependency plan, dependency blocker, and uninstall reverse dependency blocker
- [x] 22.12 Run plugin catalog, dependency resolver, lifecycle, startup auto-enable, upgrade, and runtime Go unit tests

## 23. Plugin Runtime Upgrade

- [x] 23.1 Design and implement runtime state field covering `normal`, `pending_upgrade`, `abnormal`, `upgrade_running`, `upgrade_failed`
- [x] 23.2 Extend plugin list and detail DTOs with `runtimeState`, `effectiveVersion`, `discoveredVersion`, `upgradeAvailable`, `abnormalReason`, and last failure info
- [x] 23.3 Adjust source and dynamic plugin discovery logic to mark version drift without overwriting effective release
- [x] 23.4 Implement `pending_upgrade`/`abnormal`/`upgrade_failed` business entry protection: routes return upgrade-required, menus hidden, cron paused, hooks not dispatched
- [x] 23.5 Add `GET /plugins/{id}/upgrade/preview` returning before/after manifest, dependency check, SQL summary, hostServices diff, and risk warnings
- [x] 23.6 Add `POST /plugins/{id}/upgrade` with permission check, confirmation validation, state re-read, and runtime upgrade orchestration
- [x] 23.7 Implement upgrade orchestration: lock, pre-check, `BeforeUpgrade` callback, custom `Upgrade` callback, upgrade SQL, governance sync, release switch, cache invalidation, `AfterUpgrade` callback
- [x] 23.8 Implement upgrade failure recording and retry semantics with failure phase, error code, manifest snapshot, and current effective version
- [x] 23.9 Implement unified lifecycle callback interface (`BeforeInstall`, `BeforeUpgrade`, `Upgrade`, `AfterUpgrade`, `BeforeDisable`, `BeforeUninstall`, etc.) and delete old `Can*`/guard registration
- [x] 23.10 Extend dynamic plugin lifecycle with `Upgrade` and `Uninstall` execution-phase callbacks, `purgeStorageData` strategy, and tenant disable/delete lifecycle
- [x] 23.11 Implement upgrade success/failure cache invalidation by plugin ID scope; cluster mode uses distributed lock and shared revision broadcasting
- [x] 23.12 Add frontend plugin management page runtime upgrade UI: status labels, upgrade button, upgrade confirmation dialog, success/failure refresh, and abnormal repair prompt
- [x] 23.13 Add E2E `TC0236-plugin-runtime-upgrade` covering pending upgrade state, upgrade button, upgrade confirmation, and success state refresh
- [x] 23.14 Run comprehensive backend unit tests, frontend typecheck, plugin management E2E, and startup binding package tests

## 24. Official Plugins Submodule Decoupling

- [x] 24.1 Adjust default Go workspace and build entry so host-only state does not fail on missing `apps/lina-plugins` modules
- [x] 24.2 Remove host default entry unconditional compile-time dependency on official plugin aggregate module; implement explicit plugin-full build path
- [x] 24.3 Adjust source plugin manifest scanning to return empty set when workspace missing or empty; preserve dynamic plugin discovery
- [x] 24.4 Adjust frontend Vite plugin page scanning to return empty module set when workspace missing or empty
- [x] 24.5 Adjust Playwright config and test governance scripts for host-only E2E discovery without plugin workspace
- [x] 24.6 Mount official plugin repository as single submodule to `apps/lina-plugins` with `.gitmodules` and initialization documentation
- [x] 24.7 Remove official plugin ID to fixed parent directory host mapping; allow plugins to autonomously declare `parent_key`
- [x] 24.8 Update README/README.zh-CN, CONTRIBUTING, and AGENTS with submodule initialization, host-only verification, and plugin-full verification workflows
- [x] 24.9 Add nightly CI host-only-build-smoke job and separate host-only vs plugin-full CI test matrices

## 25. Plugin Workspace Management

- [x] 25.1 Extend `hack/tools/linactl` with `plugins.sources` config structure supporting `repo`, `root`, `ref`, and string array `items`
- [x] 25.2 Implement `linactl plugins.init` to convert `apps/lina-plugins` from submodule to regular directory preserving content
- [x] 25.3 Implement `linactl plugins.install` / `plugins.update` with temporary Git checkout, directory copy, target protection, dirty blocking, and `force=1` override
- [x] 25.4 Implement `apps/lina-plugins/.linapro-plugins.lock.yaml` lock file recording source, repo, root, ref, resolved commit, manifest version, and content digest
- [x] 25.5 Implement `linactl plugins.status` as read-only diagnosis of workspace type, configured plugins, local state, dirty detection, and remote update status
- [x] 25.6 Support wildcard `"*"` in `plugins.sources.items` to expand all plugins under source root; block mixing wildcard with explicit IDs
- [x] 25.7 Update README/README.zh-CN and `hack/tools/linactl` README with plugin workspace management command usage

## 26. Dynamic Lifecycle Auto-Discovery and Source Plugin Cache Service

- [x] 26.1 Extract controller handler metadata discovery entry in `pluginbridge/guest` reusing dispatcher signature, request type, and internal path derivation rules
- [x] 26.2 Modify `build-wasm` to auto-discover `Before*`/`After*` lifecycle operations from backend controller metadata and generate default `LifecycleContract`
- [x] 26.3 Demote `backend/lifecycle/*.yaml` to optional override merged by operation; fail build on non-existent handlers, duplicate operations, invalid operations, invalid timeouts, and old `Can*`/guard naming
- [x] 26.4 Delete `plugin-demo-dynamic/backend/lifecycle/*.yaml` and verify artifact still contains 14 lifecycle contracts via auto-discovery
- [x] 26.5 Converge lifecycle request manifest snapshot to shared typed `ManifestSnapshotV1` bridge contract; remove hand-written map field names
- [x] 26.6 Add source-plugin scoped `CacheService` contract in `pluginservice/contract` with `Get`, `Set`, `Delete`, `Incr`, `Expire` and `time.Duration` TTL
- [x] 26.7 Add `Cache() contract.CacheService` to `pluginhost.HostServices`; implement plugin-scoped cache adapter binding plugin ID, tenant context, namespace, and logical key to shared `kvCacheSvc`
- [x] 26.8 Modify HTTP route, Cron, managed cron, and hook registration paths to pass plugin-scoped host services with bound cache adapter
- [x] 26.9 Inject shared `kvCacheSvc` from HTTP startup into `pluginhostservices.New`; no `kvcache.New()` in plugin call paths

## 27. Feedback and Bugfixes

- [x] All feedback items from individual change archives have been addressed and merged into the corresponding functional areas above.
