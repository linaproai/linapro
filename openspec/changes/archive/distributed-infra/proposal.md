## Why

LinaPro targets multi-node deployment (Kubernetes, multi-instance load balancing) but originally lacked distributed infrastructure for leader election, cross-node cache coordination, and plugin bridge component organization. Specific problems include:

- **Duplicate cron execution**: all nodes run Session Cleanup and Server Monitor Cleanup simultaneously, causing redundant operations and race conditions.
- **Stale cache across nodes**: shared revisions for permissions, runtime configuration, and plugin runtime state reuse a lossy `MEMORY` cache table (`sys_kv_cache`), which loses data on database restart and does not guarantee atomic increment. Nodes can continue using stale authorization or configuration snapshots indefinitely.
- **Unstructured pluginbridge package**: `pkg/pluginbridge` accumulates 40+ production files mixing ABI contracts, codecs, WASM artifact helpers, host call protocols, host service protocols, and guest SDK, making it difficult for developers to distinguish stable contracts from internal protocol details.
- **Stale Wasm compilation cache**: same-version dynamic-plugin refresh cannot reliably invalidate Wasm compilation cache on other nodes because cache keys depend only on mutable artifact paths.
- **Missing login homePath fallback**: `/user/info` returns a fixed `/analytics` homePath, causing 404 for users without that route permission.

## What Changes

- Add `locker` distributed lock component with database-backed lock acquisition, release, lease renewal, and leader election using MySQL `MEMORY` engine table `sys_locker`.
- Add `cron` task classification: Master-Only jobs execute only on the leader node; All-Node jobs execute on every node.
- Add unified `cachecoord` cache coordination component for free-form cache-domain revision publishing, single-node local invalidation, cluster-mode shared persistent revisions, cross-node synchronization, explicit scoped invalidation, and observability.
- Add persistent `sys_cache_revision` (InnoDB) for critical cache-domain revisions; keep `sys_kv_cache` as lossy plugin/module KV cache only.
- Fix `kvcache.Incr` to use single-SQL atomic update; refactor `kvcache` into generic KV cache foundation with backend/provider abstraction and `time.Duration` TTL.
- Bind dynamic-plugin Wasm compilation cache to artifact checksum or generation so same-version refresh invalidates stale cache on all nodes.
- Refactor `pkg/pluginbridge` into responsibility-scoped public subcomponent packages (`contract`, `codec`, `artifact`, `hostcall`, `hostservice`, `guest`) with a thin root-package facade preserving backward compatibility.
- Fix login `homePath` to return the user's first accessible menu route instead of a hardcoded path.

## Capabilities

### New Capabilities

- `distributed-locker`: database-backed distributed lock with acquisition, release, lease renewal, and state checking.
- `leader-election`: automatic leader election on service start, lease auto-renewal, failover, and Master-Only job gating.
- `cache-coordination`: unified cache coordination for revision publishing, freshness checks, topology-aware single-node/cluster strategies, protected runtime parameter bounded consistency, and permission topology cross-node invalidation.
- `pluginbridge-subcomponent-architecture`: pluginbridge subcomponent package structure, dependency boundaries, compatibility facade, and verification requirements.

### Modified Capabilities

- `cron-jobs`: cron task management gains Master-Only / All-Node classification and leader-node check logic.
- `plugin-cache-service`: plugin host-cache boundaries, concurrent `incr` atomicity, expiration cleanup semantics, and separation from critical revision coordination.
- `plugin-runtime-loading`: plugin runtime derived cache invalidation across nodes, Wasm compilation cache checksum binding, and WASM custom section parsing centralization through pluginbridge.

## Impact

**Backend services**:
- New: `internal/service/locker/`, `internal/service/cachecoord/`
- Modified: `internal/service/cron/`, `internal/service/config/`, `internal/service/role/`, `internal/service/kvcache/`, `internal/service/pluginruntimecache/`, `internal/service/plugin/internal/runtime/`, `internal/service/plugin/internal/frontend/`, `internal/service/i18n/`, `internal/service/apidoc/`, `internal/service/sysconfig/`, `internal/service/menu/`, `internal/service/plugin/`, `internal/cmd/cmd_http.go`
- Refactored: `apps/lina-core/pkg/pluginbridge/` into subcomponent packages; `apps/lina-core/pkg/plugindb/`; dynamic plugin demo `apps/lina-plugins/plugin-demo-dynamic/`

**Database**:
- New: `sys_locker` table (MEMORY engine) for distributed lock state
- New: `sys_cache_revision` table (InnoDB) for persistent cache-domain revisions
- Preserved: `sys_kv_cache` as MEMORY lossy plugin/module KV cache only

**Configuration**:
- `manifest/config/config.yaml` gains `locker` configuration (lease duration, renewal interval)

**Tests**:
- Locker unit tests: core lock service, lock instance, lease management, leader election (84.1% coverage)
- Cache coordination: concurrent publishing tests, dual-instance service-level tests, plugin host-cache tests, dynamic-plugin same-version refresh tests
- Pluginbridge: round-trip protocol tests, facade consistency tests, guest SDK tests, wasip1/wasm build verification

**i18n**: cache coordination diagnostic fields added to `/system/info` response with `zh-CN` and `zh-TW` apidoc i18n JSON synchronized; runtime error codes added for cache coordination and kvcache failures with `en-US`, `zh-CN`, and `zh-TW` error.json synchronized; no frontend page, button, menu, or runtime UI copy added by the distributed infrastructure changes. The pluginbridge refactor does not add, modify, or delete any i18n resources.
