## Why

The project already has several in-process caches and shared revision mechanisms, but shared revisions reuse `sys_kv_cache`. That table is a lossy cache table based on the `MEMORY` engine. It is appropriate for plugin host-cache data, but not for persistent revision state for critical cache domains such as permissions, runtime configuration, plugin runtime state, and Wasm modules. Critical derived caches need independent cluster cache coordination that explicitly defines the authoritative source, consistency window, invalidation triggers, and failure fallback.

## What Changes

- Add a unified host distributed cache coordination capability, distinguishing local invalidation when `cluster.enabled=false` from shared revision and cross-instance synchronization when `cluster.enabled=true`.
- Coordinate critical cache domains: permission topology, protected runtime parameters, plugin runtime enabled snapshots, plugin frontend bundles, plugin i18n bundles, and dynamic-plugin Wasm compilation cache.
- Fix shared revision publishing reliability by requiring revision increments to be atomic, persistent, idempotent, and observable.
- Preserve plugin host-cache as lossy cache semantics and keep `sys_kv_cache` from becoming persistent business storage; also fix `incr` atomicity while the same database instance is alive.
- Adjust dynamic-plugin same-version refresh so other nodes cannot continue using stale Wasm compilation cache or stale frontend/i18n derived caches.
- Define failure handling for invalidation publishing on critical write paths: permission and configuration cache invalidation failures must not be silently swallowed.

## Capabilities

### New Capabilities

- `distributed-cache-coordination`: defines unified host cache coordination, revision publishing, free-form cache domains, optional policy configuration, cross-node synchronization, staleness windows, fallback behavior, and observability.

### Modified Capabilities

- `plugin-cache-service`: changes lossy plugin host-cache boundaries, TTL, `incr` atomicity, and expired-data cleanup requirements.
- `plugin-runtime-loading`: changes cross-node invalidation requirements for dynamic-plugin runtime cache, Wasm compilation cache, frontend bundles, and i18n derived caches.
- `config-management`: protected runtime parameter cache must use the unified coordination mechanism for cross-node visibility and bounded staleness.
- `role-management`: role, menu, user-role, and plugin permission topology changes must reliably invalidate token permission snapshots through the unified coordination mechanism.

## Impact

- Backend: `internal/service/kvcache`, `internal/service/config`, `internal/service/role`, `internal/service/pluginruntimecache`, `internal/service/plugin/internal/runtime`, `internal/service/plugin/internal/frontend`, `internal/service/i18n`, `internal/service/sysconfig`, `internal/service/menu`, and `internal/service/plugin`.
- Database: add persistent `sys_cache_revision`; keep `sys_kv_cache` as a `MEMORY` cache table and stop using it for critical cache revisions. The project is new, so existing SQL can be modified directly and databases reinitialized.
- Runtime: add cache-domain status query or health-diagnostic output exposing local revision, shared revision, last sync time, latest error, and stale seconds.
- Tests: add concurrency unit tests, dual-instance service-level tests, and dynamic-plugin same-version refresh tests covering non-lost revisions, cross-node invalidation, bounded staleness, and failure fallback.
- i18n: this change does not add user-visible UI copy. If health diagnostics or API documentation fields are added, maintain apidoc i18n resources accordingly.
