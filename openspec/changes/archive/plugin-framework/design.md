## Context

LinaPro's extension model previously relied on direct source modification of the host codebase. The plugin platform establishes a unified contract, lifecycle, runtime, governance, and host-service capability model covering source plugins compiled into the host, dynamically installable WASM runtime plugins, frontend page integration, backend hook/slot extensions, permission governance, multi-node hot upgrade, startup automation, installation UX, and structured host-service capabilities for dynamic plugins.

## 1. Plugin Contract and Lifecycle

### 1.1 Unified Plugin Contract

All plugins use `plugin.yaml` as the entry manifest. Source plugins reside under `apps/lina-plugins/<plugin-id>/`; dynamic plugins are discovered from `plugin.dynamic.storagePath`. The manifest requires only `id`, `name`, `version`, and `type` (`source` or `dynamic`). Dynamic plugin `wasm` is the runtime artifact semantic, not a first-level type.

SQL, frontend pages, slots, menus, and permissions follow directory and code conventions rather than being redundantly declared in the manifest. Menu registration uses manifest `menus` metadata with `menu_key` as the stable business identifier and `parent_key` for parent resolution.

### 1.2 Plugin Lifecycle State Machine

Source plugins are discovered via directory scanning and registered in `sys_plugin`. On first sync they enter a discovered-only state; administrators or `plugin.autoEnable` advance them to installed and enabled. The management page does not expose install/uninstall for source plugins.

Dynamic plugins follow the full lifecycle: upload to staging, install with migration execution and resource registration, enable with authorization confirmation, disable with hook/slot/page/menu suspension, uninstall with governance resource cleanup, and upgrade with generation-based hot-switch.

Upgrade uses `desired_state/current_state/generation/release_id` state machine. The primary node Reconciler drives shared migrations and release switches; follower nodes converge local projections. Failed releases are marked `failed` and rolled back to the stable release.

### 1.3 Plugin Governance Resources

Five metadata tables track plugin state:
- `sys_plugin`: Current install/enable state, type, error status
- `sys_plugin_release`: Version, artifact info, resource paths, manifest snapshot
- `sys_plugin_migration`: SQL migration execution records with `install`, `uninstall`, and `mock` directions
- `sys_plugin_resource_ref`: Ownership references for menus, configs, dicts, files, host-service resources
- `sys_plugin_node_state`: Multi-node convergence state, heartbeat, and error info

## 2. Dynamic Plugin Runtime

### 2.1 WASM Artifact and Loading

Dynamic plugins compile to `.wasm` artifacts containing custom sections for manifest, frontend assets, install/uninstall SQL, route contracts, bridge ABI, host-service governance snapshot, and capability declarations. The host validates file headers, ABI version, custom sections, and embedded manifest during upload.

The `wazero` runtime loads artifacts, calls `_initialize` if present, and provides a restricted execution environment. Frontend assets are extracted from custom sections and cached in memory, with startup warmup and request-time lazy loading fallback.

### 2.2 Dynamic Route Runtime

Route contracts are extracted from `backend/api/**/*.go` `g.Meta` during build and embedded in the `lina.plugin.backend.routes` custom section. The host restores `manifest.Routes` on artifact load.

Dynamic routes are fixed under `/api/v1/extensions/{pluginId}/...`. The host dispatches through standard `RouterGroup + Middleware` registration, performs route matching with path parameter support, applies authentication and permission checks based on `access` (`login`/`public`) and `permission` declarations, then executes through the WASM bridge.

The bridge uses protobuf-encoded `DynamicRouteBridgeRequestEnvelopeV1`/`DynamicRouteBridgeResponseEnvelopeV1` with versioned binary protocol. Text codecs are rejected. The guest exports `lina_dynamic_route_alloc` and `lina_dynamic_route_execute`; the host serializes the request snapshot, writes to guest memory, invokes execution, and deserializes the response.

Dynamic route permissions are materialized as hidden menu items under `sys_menu.perms`, synchronized on plugin enable/disable/uninstall/version change.

### 2.3 Host Functions and Host Services

Host services evolved from discrete opcodes (`host:log`, `host:state`, `host:db:*`) to a structured model. The `lina_env.host_call` entry is preserved but converged to a single `service invoke` channel. All capabilities are published through the host-service registry.

The plugin declares `hostServices` in `plugin.yaml`; the builder validates and embeds them in a custom section. The host derives coarse-grained `capabilities` automatically from `hostServices.methods`. Runtime calls pass through capability check, service/method dispatch, resource authorization, execution context, and audit.

**Runtime service**: `log.write`, `state.get/set/delete`, `info.now/uuid/node`
**Storage service**: `put/get/delete/list/stat` with logical path authorization via `resources.paths`, path normalization, prefix matching, and default-deny
**Network service**: `request` with URL pattern authorization, scheme/host/port/path matching, glob wildcards, and platform-level header protection
**Data service**: `list/get/create/update/delete/transaction` with table-level authorization via `resources.tables`, DAO/ORM execution through `gdb` interceptors, `DoCommit` governance, and `plugindb` guest SDK
**Cache service**: `get/set/delete/incr/expire` via MySQL `MEMORY` table with namespace/key/value length validation
**Lock service**: `acquire/renew/release` reusing host distributed lock with ticket-based isolation
**Notify service**: `send` through authorized notification channels with unified notification domain tables

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
