## Why

LinaPro's existing distributed capabilities mainly rely on PostgreSQL tables to simulate locks, cache revision numbers, distributed KV cache, and short-lived TTL token states; this approach is sufficient for standalone mode and early multi-node fallback, but in true multi-node deployment it pushes the main business database toward high-frequency coordination paths, resulting in suboptimal performance, fault isolation, and cross-node real-time behavior.

This change introduces Redis as the mandatory unified distributed coordination backend for cluster mode: standalone mode continues to maintain the lightweight PostgreSQL + process cache form, while cluster mode decouples coordination provider to handle distributed locks, cache revisions, cross-node events, short-lived TTL states, and distributed KV cache.

## What Changes

- **BREAKING**: When `cluster.enabled=true`, `cluster.coordination: redis` must be explicitly configured; currently only `redis` is supported; missing, empty, or invalid values cause startup failure.
- **BREAKING**: When `cluster.enabled=true` and Redis configuration is unavailable or startup probe fails, the host service refuses to start in cluster mode.
- Add `cluster.redis` configuration section for declaring Redis address, database, password, connect timeout, read timeout, and write timeout.
- Add internal coordination provider abstraction with Redis as the first implementation; business modules must not directly depend on Redis client.
- Migrate cluster-mode leader election, distributed locks, and plugin locks from PostgreSQL table coordination to coordination lock capability.
- Migrate cluster-mode `cachecoord` from PostgreSQL `sys_cache_revision` row revision to Redis revision + event coordination; standalone mode continues using process-internal revision.
- Migrate cluster-mode `kvcache` backend to Redis, leveraging native TTL and atomic `INCRBY`, avoiding main database bearing plugin/host short-lived KV hotspots.
- Unify JWT revoke, `pre_token`, one-time token, and other short-lived TTL authentication states to coordination KV capability.
- Introduce Redis storage strategy for online session hot state, reducing request-path pressure on `sys_online_session`; PostgreSQL retains online user management, data permission filtering, and audit/projection boundaries.
- Unify cross-node derived cache invalidation for role permission topology, runtime configuration, plugin runtime, dynamic plugin reconciler, runtime i18n, and frontend bundle to Redis event/revision.
- Preserve PostgreSQL as authoritative data source for users, roles, menus, tenants, plugin registry, system configuration, audit logs, task logs, notification messages, and other business data.
- Preserve `sys_cache_revision`, `sys_kv_cache`, `sys_locker` as standalone mode, testing, diagnostics, or future fallback implementation boundaries, but cluster mode must not depend on them for cross-node consistency.

## Capabilities

### New Capabilities

- `cluster-coordination-config`: Define cluster coordination configuration, Redis configuration, startup validation, standalone/cluster branching, and configuration error handling.
- `coordination-provider`: Define internal coordination provider abstraction, Redis provider capability set, health checks, namespace, fault semantics, and observability.
- `session-hot-state`: Define online session hot state Redis storage in cluster mode, PostgreSQL management projection, forced logout, expiration, list query, and degradation strategy.

### Modified Capabilities

- `cluster-deployment-mode`: Cluster mode converges from "optional shared PostgreSQL coordination" to "must configure coordination backend, currently Redis".
- `cluster-topology-boundaries`: Cluster topology, node identity, primary determination, and topology injection must depend on unified coordination abstraction, avoiding business components individually connecting to Redis.
- `distributed-locker`: Cluster-mode distributed locks and leader election use coordination lock; PostgreSQL locker only retains standalone/testing/fallback boundary.
- `leader-election`: Primary election lease, renewal, release, disconnection recovery, and fencing token semantics change to Redis atomic lock model.
- `distributed-cache-coordination`: Cluster-mode cache revision and cross-node invalidation events use Redis revision + event; continue retaining tenant scope, explicit scope, idempotency, and maximum stale window requirements.
- `plugin-cache-service`: Cluster-mode host/plugin KV cache uses coordination KV backend; current coordination backend is Redis, TTL, `incr`, delete, expiration, and cache miss are handled by Redis coordination KV capability.
- `user-auth`: JWT revoke, `pre_token`, tenant switch old token revocation, logout, and authentication short-term states change to coordination KV, with explicit Redis fault fail-closed strategy.
- `online-user`: Online user list, forced logout, session expiration, data permission filtering, and session hot path need to adapt to Redis hot state + PostgreSQL projection model.
- `role-management`: Permission topology revision, token access snapshot invalidation, and cross-node synchronization use coordination revision/event.
- `config-management`: Protected runtime parameter revision, process-local snapshot invalidation, and cross-node synchronization use coordination revision/event.
- `plugin-runtime-loading`: Plugin runtime cache, dynamic plugin reconciler, frontend bundle, runtime i18n, and Wasm derived cache invalidation use coordination revision/event.
- `cron-job-management`: Master node tasks, all node tasks, session cleanup, KV expiration cleanup, and built-in cluster synchronization tasks need to adjust execution strategy based on Redis coordination capability.
- `plugin-lock-service`: Plugin lock capability used through host service in cluster mode goes through coordination lock, inheriting lease, token verification release, and fault semantics.
- `system-info`: System runtime information and health diagnostics need to expose coordination backend, Redis connectivity, revision/event/lock health status, and recent errors.

## Impact

**Backend Configuration and Startup Chain:**
- Modify `apps/lina-core/internal/service/config/` cluster configuration structure, parsing, defaults, and validation.
- Modify `apps/lina-core/manifest/config/config.template.yaml`, adding `cluster.coordination` and `cluster.redis` example configuration.
- Modify `apps/lina-core/internal/cmd/` startup orchestration, ensuring cluster mode completes Redis configuration validation, connection probe, and coordination provider injection before starting services.
- SQLite mode continues forcing `cluster.enabled=false`, must not require Redis.

**Backend Foundation Components:**
- Add or refactor `apps/lina-core/internal/service/coordination/`, carrying provider abstraction, Redis implementation, health checks, namespace, and event protocol.
- Modify `cluster`, `locker`, `cachecoord`, `kvcache`, `session`, `auth`, `role`, `config`, `cron`, `pluginruntimecache`, `plugin`, `i18n`, `sysinfo` and other services.
- Plugin Wasm host service cache/lock capability needs to enter through host unified coordination/kvcache/locker facade, not directly depending on Redis.

**External Dependencies:**
- Add Redis client dependency, requiring support for context, connection pool, timeout configuration, `SET NX PX`, Lua or equivalent atomic compare-and-delete, `INCR`, TTL, Pub/Sub or Streams.
- Development, testing, deployment documentation needs to explain cluster mode must prepare Redis; standalone mode does not need Redis.

**Database and SQL:**
- Do not use Redis to replace PostgreSQL authoritative business data.
- Existing `sys_cache_revision`, `sys_kv_cache`, `sys_locker` tables are preserved, not serving as primary implementation in cluster-mode cross-node consistency paths.
- No new iteration-specific SQL unless implementation phase discovers need to save coordination diagnostic projection or session management projection fields.

**API and Frontend:**
- No new user business APIs.
- Health check or system info responses may add coordination/redis diagnostic fields, needing synchronous apidoc i18n JSON.
- Frontend displaying system info or health status needs to use existing i18n language packs, not hardcoded new copy.

**Cache Consistency and Security:**
- Security paths include token revoke, pre-token, permission topology, runtime configuration, and plugin enable state; Redis unavailability must fail-closed or conservative-hide, not silently pass.
- Ordinary lossy cache allows returning cache miss on Redis failure, but must not fake successful writes.

**Testing:**
- Add configuration parsing and startup failure unit tests.
- Add Redis provider unit/integration tests, prioritizing replaceable fake provider coverage for semantics, Redis real connection tests explicitly enabled via environment variables.
- Update `cachecoord`, `kvcache`, `locker`, `auth`, `session`, `role`, `config`, `pluginruntimecache` and plugin host service related tests.
- When affecting online user forced logout or system info visible behavior, supplement E2E or existing TC sub-assertions.
