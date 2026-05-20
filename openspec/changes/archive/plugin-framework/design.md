## Context

LinaPro's extension model previously relied on direct source modification of the host codebase. The plugin platform establishes a unified contract, lifecycle, runtime, governance, and host-service capability model covering source plugins compiled into the host, dynamically installable WASM runtime plugins, frontend page integration, backend hook/slot extensions, permission governance, multi-node hot upgrade, startup automation, installation UX, structured host-service capabilities for dynamic plugins, plugin ID governance, dependency management, runtime upgrade, official plugin workspace decoupling, and plugin workspace management.

## 1. Plugin Contract and Lifecycle

### 1.1 Unified Plugin Contract

All plugins use `plugin.yaml` as the entry manifest. Source plugins reside under `apps/lina-plugins/<plugin-id>/`; dynamic plugins are discovered from `plugin.dynamic.storagePath`. The manifest requires only `id`, `name`, `version`, and `type` (`source` or `dynamic`). Dynamic plugin `wasm` is the runtime artifact semantic, not a first-level type.

SQL, frontend pages, slots, menus, and permissions follow directory and code conventions rather than being redundantly declared in the manifest. Menu registration uses manifest `menus` metadata with `menu_key` as the stable business identifier and `parent_key` for parent resolution. Plugins may declare `dependencies` for framework version constraints and plugin dependency constraints.

### 1.2 Plugin Lifecycle State Machine

Source plugins are discovered via directory scanning and registered in `sys_plugin`. On first sync they enter a discovered-only state; administrators or `plugin.autoEnable` advance them to installed and enabled. The management page does not expose install/uninstall for source plugins.

Dynamic plugins follow the full lifecycle: upload to staging, install with migration execution and resource registration, enable with authorization confirmation, disable with hook/slot/page/menu suspension, uninstall with governance resource cleanup, and upgrade with generation-based hot-switch.

Upgrade uses `desired_state/current_state/generation/release_id` state machine. The primary node Reconciler drives shared migrations and release switches; follower nodes converge local projections. Failed releases are marked `failed` and rolled back to the stable release.

### 1.3 Plugin Governance Resources

Five metadata tables track plugin state:
- `sys_plugin`: Current install/enable state, type, error status
- `sys_plugin_release`: Version, artifact info, resource paths, manifest snapshot
- `sys_plugin_migration`: SQL migration execution records with `install`, `uninstall`, `mock`, and `upgrade` directions
- `sys_plugin_resource_ref`: Ownership references for menus, configs, dicts, files, host-service resources
- `sys_plugin_node_state`: Multi-node convergence state, heartbeat, and error info

### 1.4 Unified Lifecycle Callback Model

The project uses a unified lifecycle callback model replacing the old `Can*` guard pattern. Source plugins and dynamic plugins share the same `Before*`/`After*` operation names:

- `BeforeInstall` / `AfterInstall`: Pre-install veto and post-install notification
- `BeforeUpgrade` / `Upgrade` / `AfterUpgrade`: Pre-upgrade veto, custom upgrade execution, and post-upgrade notification
- `BeforeDisable` / `AfterDisable`: Pre-disable veto and post-disable notification
- `BeforeUninstall` / `Uninstall` / `AfterUninstall`: Pre-uninstall veto, custom cleanup execution, and post-uninstall notification
- `BeforeTenantDisable` / `AfterTenantDisable`: Tenant-level disable veto and notification
- `BeforeTenantDelete` / `AfterTenantDelete`: Tenant deletion veto and notification
- `BeforeInstallModeChange` / `AfterInstallModeChange`: Install mode switch veto and notification

`Before*` callbacks return allow/deny decisions with stable reason keys; `After*` callbacks are best-effort event notifications. The `BeforeUninstall` and `AfterUninstall` callbacks receive `purgeStorageData` strategy to distinguish data-preserving from data-cleaning uninstall. The old `RegisterLifecycleGuard`, `CanUninstall`, `CanDisable`, `CanTenantDelete` interfaces are removed.

## 2. Dynamic Plugin Runtime

### 2.1 WASM Artifact and Loading

Dynamic plugins compile to `.wasm` artifacts containing custom sections for manifest, frontend assets, install/uninstall SQL, route contracts, bridge ABI, host-service governance snapshot, lifecycle contracts, and capability declarations. The host validates file headers, ABI version, custom sections, and embedded manifest during upload.

The `wazero` runtime loads artifacts, calls `_initialize` if present, and provides a restricted execution environment. Frontend assets are extracted from custom sections and cached in memory, with startup warmup and request-time lazy loading fallback.

### 2.2 Build-Time Lifecycle Auto-Discovery

Dynamic plugin lifecycle contracts are auto-discovered at build time from guest controller methods. The `build-wasm` tool identifies methods matching source-plugin lifecycle naming conventions (`BeforeInstall`, `AfterInstall`, `BeforeUpgrade`, etc.) that satisfy guest dispatcher bridge handler signatures, and generates `LifecycleContract` entries written to the WASM `lina.plugin.backend.lifecycle` custom section.

`backend/lifecycle/*.yaml` files are demoted from required declarations to optional overrides, used only to customize `requestType`, `internalPath`, or `timeoutMs` for discovered operations. YAML declarations that reference non-existent handlers, duplicate operations, use invalid operations or old `Can*`/guard naming cause build failures. The host runtime continues to read only artifact-embedded lifecycle contracts; it never probes guest methods at runtime.

The lifecycle request manifest snapshot uses a shared typed bridge contract (`ManifestSnapshotV1`) from `pluginbridge/contract`, eliminating hand-written `map[string]interface{}` field names.

### 2.3 Dynamic Route Runtime

Route contracts are extracted from `backend/api/**/*.go` `g.Meta` during build and embedded in the `lina.plugin.backend.routes` custom section. The host restores `manifest.Routes` on artifact load.

Dynamic routes are fixed under `/api/v1/extensions/{pluginId}/...`. The host dispatches through standard `RouterGroup + Middleware` registration, performs route matching with path parameter support, applies authentication and permission checks based on `access` (`login`/`public`) and `permission` declarations, then executes through the WASM bridge.

The bridge uses protobuf-encoded `DynamicRouteBridgeRequestEnvelopeV1`/`DynamicRouteBridgeResponseEnvelopeV1` with versioned binary protocol. Text codecs are rejected. The guest exports `lina_dynamic_route_alloc` and `lina_dynamic_route_execute`; the host serializes the request snapshot, writes to guest memory, invokes execution, and deserializes the response.

Dynamic route permissions are materialized as hidden menu items under `sys_menu.perms`, synchronized on plugin enable/disable/uninstall/version change.

### 2.4 Host Functions and Host Services

Host services evolved from discrete opcodes (`host:log`, `host:state`, `host:db:*`) to a structured model. The `lina_env.host_call` entry is preserved but converged to a single `service invoke` channel. All capabilities are published through the host-service registry.

The plugin declares `hostServices` in `plugin.yaml`; the builder validates and embeds them in a custom section. The host derives coarse-grained `capabilities` automatically from `hostServices.methods`. Runtime calls pass through capability check, service/method dispatch, resource authorization, execution context, and audit.

**Runtime service**: `log.write`, `state.get/set/delete`, `info.now/uuid/node`
**Storage service**: `put/get/delete/list/stat` with logical path authorization via `resources.paths`, path normalization, prefix matching, and default-deny
**Network service**: `request` with URL pattern authorization, scheme/host/port/path matching, glob wildcards, and platform-level header protection
**Data service**: `list/get/create/update/delete/transaction` with table-level authorization via `resources.tables`, DAO/ORM execution through `gdb` interceptors, `DoCommit` governance, and `plugindb` guest SDK
**Cache service**: `get/set/delete/incr/expire` via MySQL `MEMORY` table with namespace/key/value length validation; source plugins use scoped facade through `HostServices.Cache()` with plugin ID binding, tenant isolation, and shared `kvCacheSvc` backend
**Lock service**: `acquire/renew/release` reusing host distributed lock with ticket-based isolation
**Notify service**: `send` through authorized notification channels with unified notification domain tables
**Config service**: `get/exists/string/bool/int/duration` for reading host GoFrame static configuration, with arbitrary key access and no key-pattern restrictions

## 3. Plugin UI Integration

### 3.1 Page Mounting Modes

Three frontend integration modes: `iframe` (host provides menu, permission, context token), `new-tab` (host generates SSO-link jump), `embedded-mount` (plugin provides standard ESM `mount/unmount/update`). Dynamic plugin frontend resources are hosted at `/plugin-assets/<plugin-id>/<version>/...`. Source plugins participate in host frontend build.

### 3.2 Hook and Slot Extension Points

Backend hooks: `auth.login.succeeded`, `auth.logout.succeeded`, `system.started`, `plugin.installed/enabled/disabled/uninstalled`. Callback registration extensions: `http.route.register`, `http.request.after-auth`, `cron.register`, `menu.filter`, `permission.filter`. Execution modes: `blocking` and `async`.

Frontend slots: `layout.user-dropdown.after`, `dashboard.workspace.after`, `layout.header.actions.before/after`, `auth.login.after`, `crud.toolbar.after`, `crud.table.after`. All use typed constants in Go and TypeScript.

### 3.3 Generation-Aware Refresh

When a dynamic plugin hot-upgrades, users on that plugin page see a refresh prompt. Clicking refresh rebuilds menus and dynamic routes without forced navigation. Non-plugin-page users remain unaffected.

## 4. Cluster Deployment and Topology

### 4.1 Cluster Mode

`cluster.enabled` defaults to `false` (single-node). `cluster.Service` exposes `IsEnabled()`, `IsPrimary()`, `NodeID()`. Leader election is an internal implementation detail. Single-node mode skips election, treats the current node as primary, and executes all tasks synchronously.

### 4.2 Plugin Convergence

Single-node mode: plugin operations complete synchronously. Cluster mode: primary node executes shared migrations and release switches; followers converge via `sys_plugin_node_state`. Node identity generation is unified in `cluster.Service`.

## 5. Installation and Bootstrap

### 5.1 Startup Auto-Enable

`plugin.autoEnable` in the host main config file lists plugin IDs for startup auto-enable. Semantics: "install first if needed, then enable." Bootstrap runs before plugin route registration, cron wiring, and bundle warmup. Fail-fast on missing or failed plugins.

Source plugins: synchronous install/enable on primary; followers refresh after convergence. Dynamic plugins: reuse existing authorization snapshots; missing snapshots block startup.

**Startup snapshot synchronization**: The HTTP startup phase creates a startup data snapshot via `WithStartupDataSnapshot` covering plugin governance tables and reuses it across bootstrap, route wiring, and warmup. When a source plugin auto-install writes the installed state to the database through `applySourcePluginStableState`, the helper must also refresh the in-memory startup snapshot so that the subsequent enable check within the same startup orchestration reads the latest `installed`, `status`, `desiredState`, and `currentState` projections. Without this synchronization, the enable phase reads stale `installed=0` from the snapshot and fails with `Plugin is not installed`.

### 5.2 Install-and-Enable Shortcut

The installation dialog offers "Install Only" and "Install and Enable." The frontend calls install then enable sequentially, reusing existing APIs. Requires both `plugin:install` and `plugin:enable` permissions. Partial success (install succeeds, enable fails) shows real `installed but disabled` state.

### 5.3 Mock Data Installation

`installMockData` option in install request. Mock SQL from `manifest/sql/mock-data/` executes in one transaction after install SQL succeeds. Any mock failure rolls back mock data and ledger rows while preserving installed state. Ledger rows use `direction='mock'`. Startup bootstrap supports `withMockData` in structured `plugin.autoEnable` entries.

## 6. Authorization and Route Visibility

Dynamic-plugin authorization review dialogs show route exposure alongside host-service authorization. Backend projects method, real public path, access level, permission key, and summary from the release snapshot. First two routes shown by default with expand action. Route section is read-only review, not authorization items.

## 7. Query Performance and Configuration

### 7.1 Plugin List Read Path

Plugin list queries are read-only; synchronization is explicit via `POST /plugins/sync`. Host-service table comment lookup uses safe metadata APIs with fallback to raw names. Session `last_active_time` writes are throttled over a short window.

### 7.2 Duration Configuration

`jwt.expire`, `session.timeout`, `session.cleanupInterval`, `monitor.interval` use duration strings parsed to `time.Duration`. No legacy integer key compatibility.

### 7.3 Notification Domain

`sys_user_message` is replaced by `sys_notify_channel`, `sys_notify_message`, and `sys_notify_delivery`. `sys_notice` retains content management. `/user/message` facade continues to work via the new tables.

### 7.4 Declarative Permission Middleware

Static APIs declare `permission` in `g.Meta`. Middleware executes permission check. Access context is cached per login token with topology-revision-based invalidation. Cluster mode shares revision via `kvcache`.

## 8. Plugin Configuration Service

### 8.1 Problem

`apps/lina-core/pkg/pluginservice/config` had started exposing plugin-specific strongly typed configuration through `GetMonitor()`. Each new plugin or plugin configuration shape would require another change to a host public component. The configuration service itself should remain business-neutral and provide only stable, general, read-only access.

### 8.2 Generic Key Access Instead of Business Methods

`pluginservice/config.Service` exposes generic methods: `Get(ctx, key)` for raw GoFrame configuration values, `Exists(ctx, key)` for key existence checks, `Scan(ctx, key, target)` for scanning a section into a caller-provided struct, and `String/Bool/Int/Duration(ctx, key, defaultValue)` for basic type reads with default-value support. Each plugin maintains its own `Config` structure and `Load(ctx)` method. For example, `monitor-server` scans the `monitor` section inside the plugin, reads `monitor.interval` as a `time.Duration`, and applies whole-second alignment validation.

The `MonitorConfig` type alias and `GetMonitor()` plugin-specific business method are removed from the public component.

### 8.3 Arbitrary Key Reads with Read-Only Boundary

Source plugins are trusted extensions built in the same process and repository as the host. The configuration service does not add prefix restrictions to keys, so a plugin can read the full configuration file. The service is strictly read-only: no write, save, hot reload, or runtime mutation methods are exposed.

### 8.4 Duration Parsing and Business Validation Separation

The public service parses configuration strings into `time.Duration` and keeps default-value semantics stable. Business constraints such as "must be greater than 0", "must be at least 1 second", and "must align to whole seconds" are validated by the plugin in its own configuration loading method.

### 8.5 Error Returns

Generic read methods return `error` and do not directly `panic`. Plugin startup or cron registration paths can choose fail-fast behavior, while normal business paths can wrap errors as caller-visible business errors.

### 8.6 Config Host Service for Dynamic Plugins

Dynamic plugins cannot import `pkg/pluginservice/config` directly, so the `config` host service is provided through `lina_env.host_call`. A dynamic plugin declares `service: config` in `plugin.yaml` `hostServices`. `methods` may be omitted; omission grants the complete read-only method set: `get`, `exists`, `string`, `bool`, `int`, and `duration`. The request payload carries the key. `get` returns the configuration value as JSON; an empty key or `.` returns the complete static configuration snapshot. `exists` returns a found flag. `string`, `bool`, `int`, and `duration` return string representations of their respective types. The wasip1 guest SDK helpers call the corresponding host service methods directly.

### 8.7 Trust Boundary

Source plugins can read the full host configuration. Dynamic or third-party plugins must use host service authorization and auditing before reusing this capability. The service does not perform runtime cache invalidation; it reads only static configuration files.

## 9. Plugin ID Normalization

### 9.1 Runtime Safety Boundary

Plugin ID runtime validation enforces only basic safety: non-empty, 64-character maximum, and lowercase kebab-case (letters, digits, hyphens). The host does not enforce `<author>-<domain>-<capability>` structure, domain whitelists, or official capability reservations at runtime.

### 9.2 Official Plugin ID Convention

Official plugins use the `<author>-<domain>-<capability>` naming convention with `linapro` as the author segment. `core` is reserved for official foundational capability implementations (e.g., `linapro-tenant-core`, `linapro-org-core`). This is a repository governance convention, not a runtime enforcement rule.

### 9.3 Breaking Official Plugin ID Mapping

All 10 official plugins are renamed to `linapro-*` prefix:

| Old ID | New ID |
|--------|--------|
| `content-notice` | `linapro-content-notice` |
| `monitor-loginlog` | `linapro-monitor-loginlog` |
| `monitor-operlog` | `linapro-monitor-operlog` |
| `monitor-online` | `linapro-monitor-online` |
| `monitor-server` | `linapro-monitor-server` |
| `multi-tenant` | `linapro-tenant-core` |
| `org-center` | `linapro-org-core` |
| `plugin-demo-dynamic` | `linapro-demo-dynamic` |
| `plugin-demo-source` | `linapro-demo-source` |
| `demo-control` | `linapro-ops-demo-guard` |

No backward compatibility aliases are provided. Plugin-owned storage tables are renamed to match the new snake_case plugin ID scope.

### 9.4 Governance Validation

Automated governance scans verify that official plugin directory names, manifest IDs, source registration IDs, menu keys, i18n namespaces, apidoc namespaces, dynamic artifact filenames, configuration entries, and test fixtures all use the current plugin ID. Old official IDs must not appear in runtime code, configuration, tests, or active OpenSpec documents.

## 10. Plugin Dependency Management

### 10.1 Manifest Dependencies Declaration

`plugin.yaml` supports a `dependencies` object with `framework.version` (LinaPro framework version constraint) and `plugins[]` (plugin dependency list). Each plugin dependency specifies `id`, `version` (semver range), `required` (default `true`), and `install` (`auto` or `manual`, default `manual`). Undeclared `dependencies` keeps the plugin valid as dependency-free.

### 10.2 Dependency Resolution

An internal dependency resolver component takes the target plugin ID, discovered manifest collection, registry/release state, and framework version, and produces: dependency check conclusions, auto-install plan, blocker list, dependency chain, topological order, and reverse dependencies. The resolver performs graph construction, deterministic topological sorting (stable by plugin ID within each layer), self-dependency detection, and cycle detection.

### 10.3 Auto-Installation

When a hard dependency declares `install: auto` and the dependency plugin is discovered with a satisfied version but not yet installed, the system automatically installs it before the target plugin. Auto-installation proceeds in topological order, reusing existing source/dynamic install lifecycle for each dependency. Mid-failure stops subsequent installations and returns the installed-so-far list, failed plugin ID, and failure reason. Auto-installed dependencies are not automatically enabled.

### 10.4 Reverse Dependency Protection

Before uninstalling a plugin, the system scans installed plugins' hard dependency declarations. If downstream plugins depend on the target, uninstall is blocked with the downstream plugin list. Uninstall protection reads from installed release snapshots when current workspace manifests are unavailable.

### 10.5 Upgrade Dependency Validation

Source plugin upgrade and dynamic plugin install/upgrade paths validate new version dependency constraints before switching the effective release. New framework version constraints, hard dependency existence, and hard dependency version ranges must be satisfied. Upgrade does not auto-upgrade dependency plugins; it blocks and returns the dependency list requiring manual upgrade first.

### 10.6 Cache Consistency

Read-only dependency checks do not trigger cache invalidation. Multi-plugin state changes from dependency auto-installation reuse existing plugin runtime revision/event, enabled snapshot, frontend bundle, and runtime i18n bundle per-plugin-scope invalidation mechanisms. Cluster mode uses shared revision and event broadcasting.

## 11. Plugin Runtime Upgrade

### 11.1 Runtime State Model

Plugin state is decomposed into three dimensions: `installed` (whether installed), `enabled` (whether enabled), and `runtimeState` (file vs. database version consistency):

- `normal`: Database effective version matches discovered version
- `pending_upgrade`: Database effective version is lower than discovered version
- `abnormal`: Database effective version is higher than discovered version, or manifest/release cannot be safely matched
- `upgrade_running`: An upgrade task is currently executing on this node or cluster
- `upgrade_failed`: The most recent upgrade failed and requires diagnosis or retry

### 11.2 Startup Scanning

The startup phase scans source plugin directories and dynamic plugin artifact/release metadata, comparing database effective version with file discovered version. Pending upgrades are marked but do not block startup. Abnormal states are marked with stable reason codes. Plugin management page and upgrade APIs remain accessible in all non-normal states.

Business entry protection: when a plugin is in `pending_upgrade`, `abnormal`, or `upgrade_failed`, its business routes return `PLUGIN_RUNTIME_UPGRADE_REQUIRED`, its menus are hidden/disabled from navigation, its cron tasks are not scheduled, and its hooks are not dispatched. Plugin management and upgrade APIs remain fully accessible.

### 11.3 Runtime Upgrade API

- `GET /plugins/{id}/upgrade/preview`: Read-only preview returning before/after versions, manifest diff, dependency check, SQL summary, hostServices changes, and risk warnings. Only available for `pending_upgrade` plugins.
- `POST /plugins/{id}/upgrade`: Executes the upgrade with permission check, confirmation validation, and server-side state re-read. Only `pending_upgrade` plugins can enter the upgrade flow.

### 11.4 Upgrade Orchestration

The upgrade flow follows a fixed sequence:
1. Acquire lock and set `upgrade_running` state
2. Re-read effective and target manifest snapshots
3. Validate dependencies, reverse dependencies, framework version, and hostServices authorization changes
4. Execute `BeforeUpgrade` pre-callback (may veto)
5. Execute plugin custom `Upgrade` callback
6. Execute upgrade SQL and record `phase=upgrade`
7. Synchronize governance resources (menus, permissions, resource refs, i18n, apidoc, routes, cron)
8. Switch `sys_plugin.version`, `release_id`, and release state
9. Precisely invalidate plugin-scoped caches and broadcast cluster events
10. Execute `AfterUpgrade` event callback
11. Set `normal` state

Failure at any step sets `upgrade_failed` with failure phase, error code, error message key, from/to manifest snapshots, and retry information. No automatic rollback is performed.

### 11.5 Cluster Upgrade Coordination

`cluster.enabled=false`: local lock and local cache invalidation. `cluster.enabled=true`: distributed lock via `coordination.LockStore`, shared revision/event broadcasting, no concurrent upgrades of the same plugin across nodes. Upgrade success/failure invalidates plugin-scoped caches (frontend bundle, WASM module, runtime i18n) locally and broadcasts via shared plugin-runtime revision.

## 12. Official Plugins Submodule Decoupling

### 12.1 Host-Only Mode

The host can build, test, and run without `apps/lina-plugins` existing. Go workspace, compile-time imports, runtime scanning, frontend scanning, test discovery, and tool entry points are all tolerant of missing or empty plugin workspace. Source plugin discovery returns an empty set; dynamic plugin discovery continues normally.

### 12.2 Plugin-Full Mode

When `apps/lina-plugins` is initialized as a submodule (or contains plugin code), all official plugin Go unit tests, plugin-owned E2E tests, and dynamic plugin wasm builds execute normally. The build system distinguishes host-only (`plugins=0`) from plugin-full (`plugins=1`) modes, with explicit CI matrix separation.

### 12.3 Menu Parent Mount Decoupling

The host menu service no longer hardcodes official plugin IDs to fixed parent directories. Plugin manifests autonomously declare their `parent_key` for menu mounting. Menu sync validates that the referenced parent menu record exists but does not restrict mounting to stable host directories only.

## 13. Plugin Workspace Management

### 13.1 De-submodulization

`make plugins.init` / `linactl plugins.init` converts `apps/lina-plugins` from a Git submodule to a regular directory, preserving all plugin code. The command removes gitlink tracking, `.gitmodules` section, `.git/config` submodule config, and `.git/modules/apps/lina-plugins` metadata. If `.gitmodules` contains other submodules, only the `apps/lina-plugins` section is removed.

### 13.2 Configuration-Based Source Declaration

`hack/config.yaml` declares plugin sources under `plugins.sources`:

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - multi-tenant
        - org-center
```

Each source specifies `repo`, `root` (relative path within repo, `.` for root), `ref` (shared branch/tag/commit), and `items` (string array of plugin IDs). Wildcard `"*"` expands to all plugin directories containing `plugin.yaml` under the source root. Wildcard and explicit IDs cannot be mixed in the same source. Plugin IDs must be globally unique across all sources.

### 13.3 Install, Update, and Status Commands

- `plugins.install`: Temporary checkout of source repo to `temp/`, copy `<root>/<plugin-id>` to `apps/lina-plugins/<plugin-id>`. Blocks if target directory exists (use `update` or `force=1`).
- `plugins.update`: Re-fetches from source and overwrites local directory. Blocks on local dirty state unless `force=1`.
- `plugins.status`: Read-only diagnosis of workspace type, configured plugins, local existence, version, dirty state, lock state, and remote update status.
- Lock file at `apps/lina-plugins/.linapro-plugins.lock.yaml` records source, repo, root, ref, resolved commit, manifest version, and content digest per plugin.
