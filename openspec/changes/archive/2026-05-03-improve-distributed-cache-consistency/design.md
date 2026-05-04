## Context

The host currently has several cache types:

- `runtimeParamSnapshotCache` caches protected runtime parameters and syncs through shared KV revisions plus a 10 second all-node watcher.
- `accessContextCache` caches token-level permission snapshots and syncs through shared KV revisions plus a short local revision TTL.
- `pluginruntimecache.Controller` coordinates plugin enabled snapshots, frontend bundles, and runtime i18n bundles, but does not cover Wasm compilation cache.
- `frontendBundleCache`, `runtimeBundleCache`, and `wasmModuleCache` are in-process derived caches.
- `sys_kv_cache` serves both plugin host-cache and revisions for several host modules, but its table engine is `MEMORY`. It is suitable as lossy cache, not as the source for critical cache revisions. Current `Incr` is read-modify-write and does not guarantee concurrent atomicity.

In distributed deployment, authoritative cache sources should remain the MySQL business tables, plugin release tables, and runtime parameter tables. In-process caches are derived acceleration layers only. The project is new, so there is no need to preserve legacy SQL compatibility; existing SQL can be adjusted and applied by reinitializing tables.

## Goals / Non-Goals

**Goals:**

- Establish a unified `cachecoord` component for free-form cache-domain revision publishing, optional policy configuration, cross-node synchronization, explicit scoped invalidation, and observability.
- Keep low-cost local invalidation when `cluster.enabled=false`; use persistent shared revisions and request-path freshness checks when `cluster.enabled=true` to guarantee bounded staleness.
- Give permission topology, runtime parameters, plugin runtime, Wasm compilation cache, frontend bundles, and runtime i18n bundles explicit consistency models.
- Clarify that plugin host-cache is lossy cache and not part of critical revision coordination, while fixing `incr` concurrent atomicity while the database is alive.
- Fix lost shared revisions, non-atomic plugin host-cache `incr`, and stale Wasm execution after same-version dynamic-plugin refresh.
- Define failure-handling policy for critical cache write paths so permission or configuration invalidation publish failures do not return silent success.

**Non-Goals:**

- Do not introduce Redis, etcd, NATS, or other external coordination dependencies; the default implementation uses existing MySQL.
- Do not refactor every normal query cache or browser cache. Cover only critical derived caches that affect permissions, configuration, and plugin runtime. Plugin host-cache is governed only for lossy-cache boundaries, concurrent increments, and expiration cleanup semantics.
- Do not change business module REST API semantics. Any new diagnostic API is only for governance and observability.

## Decisions

### 1. Add `cachecoord` as the unified coordination entry point

Add host internal component `internal/service/cachecoord` with these capabilities:

- `MarkChanged(ctx, domain, scope, reason)`: publish a scoped revision change for a cache domain.
- `EnsureFresh(ctx, domain, scope, refresher)`: check shared revisions on request paths or watcher paths and run refresh/invalidation when the local process has not consumed the latest version.
- `Snapshot(ctx)`: return local revision, shared revision, last sync time, latest error, and stale seconds for configured or touched cache domains.
- `ConfigureDomain(...)`: optionally configure the authoritative source, maximum tolerated staleness, and failure policy for a cache domain. Unconfigured legal cache domains can still participate in coordination using default policy.

`cachecoord` does not maintain a global cache-domain allowlist and does not make policy configuration a prerequisite for using a domain. Host modules or plugin extensions that add a cache domain should define and use a stable domain string in their own implementation. Only domains requiring non-default staleness windows or failure policies need to call `ConfigureDomain` in code. Plugin manifests do not own these runtime cache coordination details.

The alternative is to keep independent revision controllers in `config`, `role`, and `pluginruntimecache`. That is smaller short-term work, but it keeps duplicating consistency policy and prevents unified observability and review.

### 2. Use an InnoDB persistent revision table instead of reusing `sys_kv_cache`

Add SQL for `sys_cache_revision`, with suggested fields:

- `domain`: cache domain, such as `runtime-config`, `permission-access`, or `plugin-runtime`.
- `scope`: explicit invalidation scope, such as `global`, `plugin:<id>`, `locale:<locale>`, or `user:<id>`.
- `revision`: monotonically increasing version.
- `reason` and `updated_at`: observability and diagnostic data.

In cluster mode, revision increments must use row-level locking, atomic update, or equivalent transactional behavior. Read-modify-write that can lose increments is forbidden. Single-node mode does not access this table and uses in-process revisions directly.

`kvcache` remains the host generic KV cache foundation module and hides its implementation through backend/provider abstraction. The current default backend is the MySQL `MEMORY` table `sys_kv_cache`; future Redis backends can use the same interface. The public `kvcache` package keeps only backend-agnostic facade, service contract, construction options, default provider adapter, and cache key encoding entry points. MySQL `MEMORY` error codes, cache key parsing, field constraints, CRUD/incr/expire/cleanup implementation details are contained under `internal/service/kvcache/internal/mysql-memory`, so default implementation details do not pollute the generic contract and future Redis providers have clear isolation. Losing cache data after a database restart is acceptable for the MySQL `MEMORY` backend, and callers must recover as cache miss. No backend may use cache data as the reliable source for permissions, configuration, plugin stable state, or other critical revisions.

The alternative is to keep reusing `sys_kv_cache`. That mixes plugin business cache with host coordination metadata, and `MEMORY` table restart clears already-published cache versions. It is not suitable as a critical consistency foundation.

### 3. Critical write paths must be bound to revision publishing

Permission topology, runtime parameters, and plugin stable-state changes are critical cache domains. If the business write succeeds but the corresponding revision cannot be published, callers must not receive silent success. The preferred implementation is to bump `sys_cache_revision` in the same database transaction as the business write. Paths that cannot join the same transaction must publish successfully before returning; otherwise they return a structured business error and log observability data.

Plugin frontend bundles and runtime i18n derived caches can tolerate brief staleness, but they must be invalidated through the `plugin-runtime` domain revision on request paths or background sync.

### 4. Plugin host-cache remains lossy cache and fixes concurrent semantics

`kvcache` is not converted into persistent state storage. It continues to hold explicit plugin/module KV cache data and no longer carries host cache coordination revisions. Service interfaces use `time.Duration` for TTL so MySQL seconds fields, Redis expiration commands, and protocol-level `expireSeconds` do not leak into the generic cache interface.

- `set`: last write wins for the same key; return the current cache result after write; data can be lost after database restart.
- `delete` and `expire`: idempotent.
- `incr`: must be linearizable while the same database and cache table are alive, using single-SQL atomic update or equivalent behavior. Read-modify-write races must not lose increments. Cache value retention after database restart is not guaranteed.
- TTL cleanup: read paths only filter expired rows and return cache miss; they must not delete data. The default MySQL backend uses a built-in hourly primary-node job `host:kvcache-cleanup-expired` to call `CleanupExpired` and delete expired rows in batch. A future Redis backend can rely on native TTL and implement `CleanupExpired` as no-op.

### 5. Dynamic plugin caches invalidate by checksum or generation

For same-version dynamic-plugin refresh, Wasm compilation cache reuse must no longer depend only on `pluginID/version` paths. The implementation can choose either:

- archive paths containing checksum, such as `releases/<plugin>/<version>/<checksum>/<artifact>`; or
- Wasm cache keys using `artifactPath@checksum`.

The `plugin-runtime` revision refresher must cover enabled snapshot, frontend bundle, runtime i18n bundle, and Wasm compilation cache. When non-primary nodes observe a plugin runtime revision change, they must discard old derived caches and rebuild from the current release table plus artifact checksum.

### 6. Freshness and failure fallback are configured by cache domain

Suggested initial policies:

- `permission-access`: maximum staleness 3 seconds. If shared revisions cannot be read and local cache exceeds the grace window, protected APIs should fail closed.
- `runtime-config`: maximum staleness 10 seconds. Reads of auth, upload, scheduler, and related runtime parameters return visible errors after the grace window.
- `plugin-runtime`: maximum staleness 5 seconds. Dynamic routes, plugin menus, and resource permissions conservatively hide or reject uncertain plugin capability if refresh fails.
- `plugin-cache`: not part of critical revision coordination. `sys_kv_cache` is lossy shared cache; restarts or cleanup misses are handled as cache misses by plugins.

These values can be constants or configuration in implementation. If the user wants a looser high-availability-first policy, confirm before applying.

## Risks / Trade-offs

- [Risk] Request-path `EnsureFresh` can increase shared revision reads. Mitigation: use short local revision TTL, batched reads, and a background all-node watcher to reduce hot-path cost.
- [Risk] Returning errors when permission invalidation publish fails can make some management operations fail. Mitigation: this avoids cross-node authorization inconsistency; errors must be structured and retryable.
- [Risk] Keeping `sys_kv_cache` as `MEMORY` means plugin cache is lost on database restart. Mitigation: that is the intended cache semantics and it is forbidden for reliable business state; critical revisions move to `sys_cache_revision`.
- [Risk] Same-version dynamic-plugin refresh with checksum paths increases artifact retention. Mitigation: add cleanup based on release state and retention windows.
- [Risk] Migrating several existing cache controllers to one component can regress behavior. Mitigation: connect domains incrementally and add dual-instance tests for each domain.

## Migration Plan

1. Adjust SQL: add `sys_cache_revision`, preserve `sys_kv_cache` as a `MEMORY` cache table, and stop storing critical cache revisions in `sys_kv_cache`.
2. Implement `cachecoord`, first connecting runtime parameters and permission topology.
3. Refactor `pluginruntimecache` or replace its shared revision logic with `cachecoord`, and add Wasm cache invalidation.
4. Fix `kvcache.Incr` concurrent increment semantics while the same database is alive, abstract backend/provider, use `time.Duration` TTL, and make `Get` expiration handling read-only filtering.
5. Add health diagnostic output and tests.
6. Run `make init`, `make dao`, backend unit tests, and required E2E tests.

Rollback strategy: because the project does not need historical data compatibility, development rollback can restore SQL and code and reinitialize the database.

## Confirmed Clarifications

- The default coordination backend uses only MySQL InnoDB `sys_cache_revision`; Redis, etcd, and NATS are not introduced.
- `sys_kv_cache` keeps `MEMORY` cache-table semantics. Clearing on database restart is an acceptable cache miss, and it is forbidden for reliable business state or critical revisions.
- Permission cache freshness that cannot be confirmed after the window uses fail-closed behavior; runtime configuration returns domain-specific visible errors.
- Cache coordination status is exposed first through backend `Snapshot`, health diagnostics, and logs; no management page is added.
- `sys_locker` `MEMORY` table reliability is not included in this cache-consistency change and should be handled by a separate future change.
- Same-version dynamic-plugin refresh may use checksum or generation to create immutable cache/archive identifiers, with cleanup for old artifacts.
