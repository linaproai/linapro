## Why

LinaPro needs a formal, stable, and extensible plugin platform to support source-code plugins compiled into the host, dynamically installable WASM runtime plugins, frontend page integration, backend hook and slot extension points, permission governance, multi-node hot upgrade, and a full host-service capability model for dynamic plugins. Without a unified plugin contract, lifecycle management, runtime loading, host-service governance, startup automation, and installation UX, the system cannot sustainably extend business capabilities without invasive modifications to core code.

## What Changes

- Define a unified plugin contract with `plugin.yaml` as the entry manifest, covering source plugins under `apps/lina-plugins/<plugin-id>/` and dynamic plugins discoverable from `plugin.dynamic.storagePath`.
- Establish the plugin lifecycle state machine: discovery, install, enable, disable, uninstall, upgrade, and hot-update, with distinct semantics for source and dynamic plugins.
- Implement dynamic WASM plugin runtime loading, including manifest validation, custom-section artifact parsing, frontend asset extraction and hosting, and `wazero`-based execution.
- Build the dynamic plugin REST runtime with route contracts extracted from `g.Meta`, fixed-prefix dispatch at `/api/v1/extensions/{pluginId}/...`, host-managed authentication and permission checks, protobuf bridge envelopes, and real WASM bridge execution with 501 fallback.
- Unify dynamic plugin resource declaration through `go:embed`, where the builder reads embedded resources and converts them to host-governable snapshot custom sections.
- Extend host-service capabilities from discrete opcodes to a structured host-service model with `runtime`, `storage`, `network`, `data`, `cache`, `lock`, and `notify` services, each with resource authorization, execution context, and audit.
- Add startup auto-enable via `plugin.autoEnable` in the host main config file, with a dedicated bootstrap phase before plugin wiring, fail-fast behavior, and cluster-aware primary-node execution.
- Add an install-and-enable shortcut in the plugin installation dialog, with permission gating, partial-success messaging, and E2E coverage.
- Add mock-data installation support with `installMockData` option, transactional mock SQL execution, structured rollback errors, and startup bootstrap integration.
- Show dynamic route exposure in the authorization review dialog alongside host-service authorization, with backend route projection and collapsible route lists.
- Make plugin list queries read-only, safe metadata lookup for host-service table comments, and session activity write throttling.
- Converge cluster deployment topology: `cluster.enabled` switch, `cluster.Service` as the sole topology facade, leader election as internal implementation detail, and plugin runtime convergence via generation model.
- Unify duration configuration across `jwt.expire`, `session.timeout`, `session.cleanupInterval`, and `monitor.interval` using duration strings parsed to `time.Duration`.
- Rebuild the notification domain with `sys_notify_channel`, `sys_notify_message`, and `sys_notify_delivery` tables, replacing `sys_user_message`.
- Establish declarative permission middleware for static APIs with access context caching and topology-revision-based invalidation.

## Capabilities

### New Capabilities
- `plugin-manifest-lifecycle`: Unified plugin directory structure, manifest schema, resource ownership, install/enable/disable/uninstall/upgrade lifecycle, and manifest-driven menu governance.
- `plugin-runtime-loading`: Dynamic WASM plugin discovery, validation, loading, hot-switch, generation propagation, and multi-node convergence.
- `plugin-hook-slot-extension`: Backend hooks, frontend slots, callback registration extension points, execution order, failure isolation, and observability.
- `plugin-ui-integration`: Plugin page mounting (iframe, new-tab, embedded-mount), frontend resource hosting, slot outlet rendering, and generation-aware refresh prompts.
- `plugin-permission-governance`: Plugin menu and permission reuse of Lina governance modules, role authorization persistence across disable/enable cycles, and runtime permission context.
- `plugin-embed-snapshot-packaging`: Dynamic plugin `go:embed` resource declaration, builder snapshot generation, and directory-scan fallback compatibility.
- `plugin-host-service-extension`: Structured host-service protocol, capability auto-derivation from `hostServices`, resource authorization at install/enable time, and execution context with audit.
- `plugin-storage-service`: Logical storage space isolation, path-prefix authorization, and `put/get/delete/list/stat` methods.
- `plugin-network-service`: Outbound HTTP via authorized URL patterns with scheme/host/port/path matching and default-deny.
- `plugin-data-service`: Table-level data access via structured CRUD/transaction methods, DAO/ORM execution, `DoCommit` interception, and `plugindb` guest SDK.
- `plugin-cache-service`: Distributed KV cache via MySQL `MEMORY` table with namespace isolation, strict length validation, and expiry cleanup.
- `plugin-lock-service`: Named lock resources reusing host distributed lock with ticket-based renew/release.
- `plugin-notify-service`: Unified notification domain with channel-based send, message records, and delivery tracking.
- `plugin-startup-bootstrap`: `plugin.autoEnable` config, startup bootstrap phase, source/dynamic branching, fail-fast, and cluster-aware primary execution.
- `plugin-mock-data-installation`: Optional mock-data loading during install, transactional mock SQL, structured rollback errors, and startup bootstrap integration.
- `plugin-api-query-performance`: Read-only plugin list queries, safe metadata lookup, and session activity write throttling.
- `plugin-install-enable-shortcut`: Install-and-enable shortcut in the installation dialog with permission gating and partial-success messaging.
- `demo-control-guard`: Demo read-only mode controlled by plugin enabled state, with clear write-blocking messages.
- `system-api-docs`: OpenAPI projection of dynamic plugin routes with runtime-aware response semantics.
- `cluster-deployment-mode`: `cluster.enabled` switch, single-node default, and cluster-aware plugin lifecycle.
- `cluster-topology-boundaries`: `cluster.Service` as sole topology facade with election encapsulation.
- `config-duration-unification`: Unified duration-string configuration for `jwt.expire`, `session.timeout`, `session.cleanupInterval`, and `monitor.interval`.

### Modified Capabilities
- `menu-management`: Plugin menu ownership, `menu_key` stability, manifest-driven sync, and visibility linkage with plugin state.
- `role-management`: Plugin menu and permission authorization with persistence across disable/enable cycles.
- `user-auth`: Authentication lifecycle hooks for plugins with failure isolation.
- `module-decoupling`: Plugin dimension extension for graceful degradation when disabled, missing, or upgrading.
- `online-user`: Duration-string session config and throttled `last_active_time` writes.
- `server-monitor`: Duration-string monitor interval and cluster-aware cleanup.
- `cron-jobs`: Primary-node-only vs all-node task classification with cluster mode awareness.
- `leader-election`: Cluster-mode-only election with single-node bypass.

## Impact

- Backend: New plugin registration, lifecycle management, runtime loading, hook bus, resource indexing, host-service dispatch, multi-node convergence, startup bootstrap, declarative permission middleware, notification domain, and cluster topology infrastructure.
- Frontend: Plugin page mounting protocol, resource access mechanism, slot extension registry, generation-aware refresh prompts, install-and-enable shortcut, mock-data checkbox, route exposure review in authorization dialog, and dynamic routing adjustments.
- Data layer: New tables for `sys_plugin`, `sys_plugin_release`, `sys_plugin_migration`, `sys_plugin_resource_ref`, `sys_plugin_node_state`, `sys_plugin_state`, `sys_kv_cache`, `sys_notify_channel`, `sys_notify_message`, `sys_notify_delivery`, and removal of `sys_user_message`.
- Build and delivery: `apps/lina-plugins/` source scanning, `hack/build-wasm` builder for WASM artifacts, `go:embed` resource declaration, and unified output directory.
- Configuration: `plugin.autoEnable`, `cluster.enabled`, `cluster.election.*`, duration-string keys, and host-service authorization snapshots.
