## Why

LinaPro needs a formal, stable, and extensible plugin platform to support source-code plugins compiled into the host, dynamically installable WASM runtime plugins, frontend page integration, backend hook and slot extension points, permission governance, multi-node hot upgrade, and a full host-service capability model for dynamic plugins. Without a unified plugin contract, lifecycle management, runtime loading, host-service governance, startup automation, and installation UX, the system cannot sustainably extend business capabilities without invasive modifications to core code.

## What Changes

- Define a unified plugin contract with `plugin.yaml` as the entry manifest, covering source plugins under `apps/lina-plugins/<plugin-id>/` and dynamic plugins discoverable from `plugin.dynamic.storagePath`.
- Establish the plugin lifecycle state machine: discovery, install, enable, disable, uninstall, upgrade, and hot-update, with distinct semantics for source and dynamic plugins.
- Implement dynamic WASM plugin runtime loading, including manifest validation, custom-section artifact parsing, frontend asset extraction and hosting, and `wazero`-based execution.
- Build the dynamic plugin REST runtime with route contracts extracted from `g.Meta`, fixed-prefix dispatch at `/api/v1/extensions/{pluginId}/...`, host-managed authentication and permission checks, protobuf bridge envelopes, and real WASM bridge execution with 501 fallback.
- Unify dynamic plugin resource declaration through `go:embed`, where the builder reads embedded resources and converts them to host-governable snapshot custom sections.
- Extend host-service capabilities from discrete opcodes to a structured host-service model with `runtime`, `storage`, `network`, `data`, `cache`, `lock`, `notify`, and `config` services, each with resource authorization, execution context, and audit.
- Add startup auto-enable via `plugin.autoEnable` in the host main config file, with a dedicated bootstrap phase before plugin wiring, fail-fast behavior, and cluster-aware primary-node execution. The bootstrap phase must synchronize the startup snapshot after any source plugin lifecycle write so that subsequent enable checks within the same startup orchestration read the latest installed state.
- Add an install-and-enable shortcut in the plugin installation dialog, with permission gating, partial-success messaging, and E2E coverage.
- Add mock-data installation support with `installMockData` option, transactional mock SQL execution, structured rollback errors, and startup bootstrap integration.
- Show dynamic route exposure in the authorization review dialog alongside host-service authorization, with backend route projection and collapsible route lists.
- Make plugin list queries read-only, safe metadata lookup for host-service table comments, and session activity write throttling.
- Converge cluster deployment topology: `cluster.enabled` switch, `cluster.Service` as the sole topology facade, leader election as internal implementation detail, and plugin runtime convergence via generation model.
- Unify duration configuration across `jwt.expire`, `session.timeout`, `session.cleanupInterval`, and `monitor.interval` using duration strings parsed to `time.Duration`.
- Rebuild the notification domain with `sys_notify_channel`, `sys_notify_message`, and `sys_notify_delivery` tables, replacing `sys_user_message`.
- Establish declarative permission middleware for static APIs with access context caching and topology-revision-based invalidation.
- Generalize the plugin configuration service from plugin-specific `GetMonitor()` to a business-neutral read-only accessor, keeping each plugin's configuration structure, defaults, and validation inside the plugin. Add the `config` host service for dynamic plugins.
- Normalize plugin IDs with basic safety boundary enforcement (non-empty, 64-char max, kebab-case), official plugin ID structured naming convention (`<author>-<domain>-<capability>` as recommendation), and breaking rename of all 10 official plugins to `linapro-*` prefix.
- Add plugin manifest `dependencies` declaration supporting framework version constraints and plugin dependency constraints, with dependency resolution, topological auto-installation, reverse-dependency protection, and dependency check results in API and UI.
- Introduce runtime upgrade state model separating file discovery from runtime state, with `pending_upgrade`/`abnormal`/`upgrade_failed` states, explicit upgrade preview and execution APIs, unified lifecycle callbacks (`Before*`/`After*` replacing old `Can*` guards), and cluster-consistent upgrade coordination.
- Enable official plugin workspace as optional submodule with host-only build/test verification, and provide plugin workspace management commands (`plugins.init`/`plugins.install`/`plugins.update`/`plugins.status`) with `hack/config.yaml` based source declaration.
- Auto-discover dynamic plugin lifecycle handlers at build time from guest controller methods, eliminating manual `backend/lifecycle/*.yaml` declarations while preserving artifact-embedded contracts as runtime authority.
- Expose source-plugin scoped cache facade through `HostServices.Cache()` for plugin-private KV cache with tenant isolation, namespace isolation, and cluster backend selection.
- Extend dynamic plugin lifecycle with `Upgrade` and `Uninstall` execution-phase callbacks, `Before*`/`After*` lifecycle for tenant disable/delete, and typed manifest snapshot bridge contract.

## Capabilities

### New Capabilities
- `plugin-manifest-lifecycle`: Unified plugin directory structure, manifest schema, resource ownership, install/enable/disable/uninstall/upgrade lifecycle, manifest-driven menu governance, plugin ID safety validation, and dependency declaration recognition.
- `plugin-runtime-loading`: Dynamic WASM plugin discovery, validation, loading, hot-switch, generation propagation, multi-node convergence, and build-time lifecycle contract auto-discovery from guest controller methods.
- `plugin-hook-slot-extension`: Backend hooks, frontend slots, callback registration extension points, execution order, failure isolation, and observability.
- `plugin-ui-integration`: Plugin page mounting (iframe, new-tab, embedded-mount), frontend resource hosting, slot outlet rendering, generation-aware refresh prompts, and host-only empty workspace tolerance.
- `plugin-permission-governance`: Plugin menu and permission reuse of Lina governance modules, role authorization persistence across disable/enable cycles, and runtime permission context.
- `plugin-embed-snapshot-packaging`: Dynamic plugin `go:embed` resource declaration, builder snapshot generation, and directory-scan fallback compatibility.
- `plugin-host-service-extension`: Structured host-service protocol, capability auto-derivation from `hostServices`, resource authorization at install/enable time, and execution context with audit.
- `plugin-storage-service`: Logical storage space isolation, path-prefix authorization, and `put/get/delete/list/stat` methods.
- `plugin-network-service`: Outbound HTTP via authorized URL patterns with scheme/host/port/path matching and default-deny.
- `plugin-data-service`: Table-level data access via structured CRUD/transaction methods, DAO/ORM execution, `DoCommit` interception, and `plugindb` guest SDK.
- `plugin-cache-service`: Distributed KV cache via MySQL `MEMORY` table with namespace isolation, strict length validation, and expiry cleanup. Extended with source-plugin scoped facade through `HostServices.Cache()` for plugin-private KV cache with tenant isolation, namespace isolation, and cluster backend selection.
- `plugin-lock-service`: Named lock resources reusing host distributed lock with ticket-based renew/release.
- `plugin-notify-service`: Unified notification domain with channel-based send, message records, and delivery tracking.
- `plugin-config-service`: Business-neutral read-only configuration access for plugins, including arbitrary key reads, section scanning, basic type parsing, `time.Duration` parsing, and a `config` host service for dynamic plugins.
- `plugin-startup-bootstrap`: `plugin.autoEnable` config, startup bootstrap phase, source/dynamic branching, fail-fast, cluster-aware primary execution, startup snapshot synchronization after source plugin lifecycle writes, and dependency-aware auto-enable with deterministic topological ordering.
- `plugin-mock-data-installation`: Optional mock-data loading during install, transactional mock SQL, structured rollback errors, and startup bootstrap integration.
- `plugin-api-query-performance`: Read-only plugin list queries, safe metadata lookup, and session activity write throttling.
- `plugin-install-enable-shortcut`: Install-and-enable shortcut in the installation dialog with permission gating and partial-success messaging.
- `demo-control-guard`: Demo read-only mode controlled by plugin enabled state (`linapro-ops-demo-guard`), with clear write-blocking messages and minimal session whitelist.
- `system-api-docs`: OpenAPI projection of dynamic plugin routes with runtime-aware response semantics.
- `cluster-deployment-mode`: `cluster.enabled` switch, single-node default, and cluster-aware plugin lifecycle.
- `cluster-topology-boundaries`: `cluster.Service` as sole topology facade with election encapsulation.
- `config-duration-unification`: Unified duration-string configuration for `jwt.expire`, `session.timeout`, `session.cleanupInterval`, and `monitor.interval`.
- `plugin-id-governance`: Plugin ID basic safety boundary enforcement (non-empty, 64-char max, kebab-case), official plugin ID normalization mapping, runtime identity consistency, and governance validation for directory names, manifest IDs, source registration IDs, menu keys, i18n namespaces, and apidoc namespaces.
- `plugin-dependency-management`: Plugin manifest `dependencies` declaration with framework version constraints and plugin dependency constraints, dependency resolution with topological sorting, automatic installation of discovered hard dependencies, reverse-dependency protection on uninstall, and dependency check results exposed via API and UI.
- `plugin-runtime-upgrade`: Runtime upgrade state model (`normal`, `pending_upgrade`, `abnormal`, `upgrade_running`, `upgrade_failed`), startup version drift scanning with status marking (not fail-fast), explicit upgrade preview and execution APIs, unified lifecycle callback model (`Before*`/`After*` replacing old `Can*` guards), upgrade failure diagnostics and retry, and cluster-consistent cache invalidation.
- `plugin-workspace-management`: Plugin workspace de-submodulization, `hack/config.yaml` based plugin source declaration, `plugins.init`/`plugins.install`/`plugins.update`/`plugins.status` cross-platform commands, lock file state tracking, and local dirty protection.
- `official-plugin-workspace-decoupling`: Official source plugin workspace as optional submodule, host-only build/test verification without plugin workspace, plugin-full verification with submodule initialized, and CI matrix separation.

### Modified Capabilities
- `menu-management`: Plugin menu ownership, `menu_key` stability, manifest-driven sync, visibility linkage with plugin state, and plugin autonomous parent mount point selection.
- `role-management`: Plugin menu and permission authorization with persistence across disable/enable cycles.
- `user-auth`: Authentication lifecycle hooks for plugins with failure isolation.
- `module-decoupling`: Plugin dimension extension for graceful degradation when disabled, missing, or upgrading.
- `online-user`: Duration-string session config and throttled `last_active_time` writes.
- `server-monitor`: Duration-string monitor interval and cluster-aware cleanup.
- `cron-jobs`: Primary-node-only vs all-node task classification with cluster mode awareness; cron declaration visibility split into executable handler publishing, authorization preview, and installed declaration projection.
- `leader-election`: Cluster-mode-only election with single-node bypass.
- `source-upgrade-governance`: Removed old development-time upgrade entry requirements; retained framework metadata display requirements.
- `plugin-upgrade-governance`: Source plugin upgrade moved from development-time command to runtime explicit upgrade model; unified dynamic plugin upgrade boundary.
- `project-setup`: Host initialization commands must support host-only workspace; plugin source management through `hack/config.yaml` and `linactl` commands.
- `e2e-suite-organization`: E2E test suite must support host-only and plugin-full separation; plugin workspace missing does not block host test discovery.
- `release-image-build`: Standard build must distinguish host-only build from full build with official plugin submodule.

## Impact

- Backend: New plugin registration, lifecycle management, runtime loading, hook bus, resource indexing, host-service dispatch, multi-node convergence, startup bootstrap with snapshot synchronization, declarative permission middleware, notification domain, cluster topology infrastructure, generalized plugin configuration service, plugin ID governance, dependency resolution engine, runtime upgrade orchestration, unified lifecycle callbacks, and source-plugin cache facade.
- Frontend: Plugin page mounting protocol, resource access mechanism, slot extension registry, generation-aware refresh prompts, install-and-enable shortcut, mock-data checkbox, route exposure review in authorization dialog, dynamic routing adjustments, dependency plan display, runtime upgrade state and actions, and host-only empty workspace tolerance.
- Data layer: New tables for `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, `sys_plugin_state`, `sys_kv_cache`, `sys_notify_channel`, `sys_notify_message`, `sys_notify_delivery`, and removal of `sys_user_message`.
- Build and delivery: `apps/lina-plugins/` source scanning, `hack/build-wasm` builder for WASM artifacts, `go:embed` resource declaration, unified output directory, build-time lifecycle auto-discovery, host-only vs plugin-full build modes, and CI matrix separation.
- Configuration: `plugin.autoEnable`, `cluster.enabled`, `cluster.election.*`, duration-string keys, host-service authorization snapshots, `hack/config.yaml` plugin sources, and workspace management lock files.
- Developer tools: `linactl plugins.init`/`plugins.install`/`plugins.update`/`plugins.status` commands, `linactl test.go`/`test.host`/`test.plugins`/`test.scripts` test matrix, and cross-platform `make` wrappers.
