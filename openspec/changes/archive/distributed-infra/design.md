## Context

LinaPro uses GoFrame with `gcron` for scheduled tasks, in-process derived caches for permissions and configuration, and `pkg/pluginbridge` for dynamic plugin Wasm bridge infrastructure. In single-node deployment these work correctly, but multi-node deployment (Kubernetes, multiple load-balanced instances) exposes several infrastructure gaps:

- All nodes execute every cron job simultaneously, causing duplicate Session Cleanup and race conditions in Server Monitor Cleanup.
- Critical cache domains (permissions, runtime configuration, plugin runtime state) share `sys_kv_cache`, a `MEMORY` engine table that loses data on restart and whose `Incr` uses non-atomic read-modify-write. Nodes can use stale authorization or configuration snapshots indefinitely.
- Dynamic-plugin same-version refresh cannot invalidate Wasm compilation cache on other nodes because cache keys depend only on mutable artifact paths.
- `pkg/pluginbridge` mixes ABI contracts, codecs, WASM artifact helpers, host call protocols, host service protocols, and guest SDK in one flat package of 40+ files, making it hard to distinguish stable contracts from internal details.
- `/user/info` returns a hardcoded `/analytics` homePath, causing 404 for users without that route permission.

## Goals / Non-Goals

**Goals:**

- Implement database-backed distributed lock supporting lock acquisition, release, lease renewal, and leader election.
- Distinguish Master-Only and All-Node cron jobs; Master-Only jobs execute only on the leader node with automatic failover.
- Establish unified `cachecoord` component for free-form cache-domain revision publishing, single-node local invalidation, cluster-mode shared persistent revisions, cross-node synchronization, and observability.
- Move critical cache revisions to persistent InnoDB `sys_cache_revision`; keep `sys_kv_cache` as lossy plugin/module KV cache only.
- Fix `kvcache.Incr` atomicity, abstract backend/provider, use `time.Duration` TTL, and add background expiration cleanup.
- Bind dynamic-plugin Wasm compilation cache to artifact checksum or generation for same-version refresh consistency.
- Refactor `pkg/pluginbridge` into responsibility-scoped subcomponent packages with backward-compatible root facade.
- Fix login homePath to return user's first accessible menu route.

**Non-Goals:**

- Do not introduce Redis, etcd, NATS, or other external coordination dependencies; default implementations use existing MySQL.
- Do not refactor every normal query cache or browser cache; cover only critical derived caches affecting permissions, configuration, and plugin runtime.
- Do not change business module REST API semantics. Diagnostic APIs are only for governance and observability.
- Do not change dynamic plugin Wasm bridge protocol, host call entry, host service method names, or payload field numbers.
- Do not modify database schema, REST API, or frontend pages beyond what is required for the listed infrastructure improvements.

## Distributed Locking and Leader Election

### Storage: MySQL MEMORY Engine

| Approach | Pros | Cons |
|----------|------|------|
| MySQL InnoDB | Persistent, supports transactions | Lower performance, needs periodic cleanup |
| MySQL MEMORY | Very fast read/write, auto-cleanup on restart | Not persistent, no transaction support |
| Redis | High performance, native TTL | External dependency |

Distributed lock state is temporary; losing it on service restart is acceptable. MEMORY engine's low latency suits frequent lock operations. No external dependency is introduced.

### Leader Election Strategy: Optimistic Lock + Lease Renewal

```
Election flow:
1. Try to acquire lock (INSERT or UPDATE expire_time)
2. Success -> become leader, start lease renewal goroutine
3. Failure -> become follower, wait for next election opportunity
4. Leader periodically renews lease (UPDATE expire_time)
5. Renewal failure -> demote to follower
```

Lease duration is 30 seconds, renewal interval is 10 seconds. Even during network partitions, the old leader's lock expires within 30 seconds and a new leader can acquire it.

### Cron Task Classification

| Task | Type | Reason |
|------|------|--------|
| Session Cleanup | Master-Only | Cleanup needs only one node |
| Server Monitor Collector | All-Node | Each node collects its own system resource data |
| Server Monitor Cleanup | Master-Only | Cleanup needs only one node |

### Component Architecture

```
service/locker/
  locker.go           # Core lock service: Lock, TryLock, IsLeader
  locker_instance.go  # Lock instance: Unlock, Renew
  locker_lease.go     # Lease renewal management: StartRenewal
  locker_election.go  # Leader election: Start, Stop
```

### Risks

- **Database failure prevents election**: all nodes demote to followers, Master-Only Jobs do not execute. This is safe degradation.
- **Network partition causes brief dual-leader**: 30-second lease ensures old leader's lock expires; new leader acquires within one renewal cycle.
- **MEMORY table size limit**: `sys_locker` stores only one lock record (leader election lock), well within the default 16MB limit.

## Unified Cache Coordination

### Architecture: `cachecoord` Component

Add `internal/service/cachecoord` with these capabilities:

- `MarkChanged(ctx, domain, scope, reason)`: publish a scoped revision change for a cache domain.
- `EnsureFresh(ctx, domain, scope, refresher)`: check shared revisions on request or watcher paths and refresh when local process has not consumed the latest version.
- `Snapshot(ctx)`: return local revision, shared revision, last sync time, latest error, and stale seconds for configured or touched cache domains.
- `ConfigureDomain(...)`: optionally configure authoritative source, maximum tolerated staleness, and failure policy for a cache domain. Unconfigured legal domains use default policy.

`cachecoord` does not maintain a global cache-domain allowlist. Policy configuration is not a prerequisite for using a domain. Only domains requiring non-default staleness windows or failure policies need to call `ConfigureDomain`.

### Persistent Revision Table

`sys_cache_revision` (InnoDB) with fields:

- `domain`: cache domain string (e.g., `runtime-config`, `permission-access`, `plugin-runtime`)
- `scope`: explicit invalidation scope (e.g., `global`, `plugin:<id>`, `locale:<locale>`)
- `revision`: monotonically increasing version
- `reason`, `updated_at`: observability data

In cluster mode, revision increments use row-level locking or atomic update. Read-modify-write that can lose increments is forbidden. Single-node mode does not access this table and uses in-process revisions directly.

### Critical Write Path Binding

Permission topology, runtime parameters, and plugin stable-state changes are critical cache domains. If the business write succeeds but revision publishing fails, callers must not receive silent success. Preferred: bump `sys_cache_revision` in the same database transaction. Paths that cannot join the same transaction must publish successfully before returning; otherwise they return a structured business error.

### kvcache Refactoring

`kvcache` becomes a generic KV cache foundation with backend/provider abstraction:

- `set`: last write wins; data can be lost after database restart.
- `delete` and `expire`: idempotent.
- `incr`: linearizable while the same database is alive using single-SQL atomic update. Read-modify-write races must not lose increments.
- TTL: uses `time.Duration` throughout; MySQL seconds fields and Redis expiration commands do not leak into the generic interface.
- Expiration cleanup: read paths only filter expired rows and return cache miss (no deletion); background hourly primary-node job `host:kvcache-cleanup-expired` calls `CleanupExpired` to delete expired rows in batch. Future Redis backends can rely on native TTL and implement `CleanupExpired` as no-op.
- Backend isolation: MySQL `MEMORY` implementation lives in `kvcache/internal/mysql-memory`; public facade keeps only backend-agnostic contract.

### Cache-Domain Policies

| Domain | Max Staleness | Failure Fallback |
|--------|--------------|-----------------|
| `permission-access` | 3 seconds | Fail closed: reject requests |
| `runtime-config` | 10 seconds | Visible error after grace window |
| `plugin-runtime` | 5 seconds | Conservatively hide/reject uncertain capability |
| `plugin-cache` | N/A (lossy) | Cache miss on restart or cleanup miss |

### Single-Node vs Cluster Mode

- `cluster.enabled=false`: in-process revision, local invalidation, synchronous refresh. No shared coordination table.
- `cluster.enabled=true`: persistent shared revisions in `sys_cache_revision`, request-path freshness checks, watcher synchronization, cross-instance invalidation.

### Risks

- Request-path `EnsureFresh` increases shared revision reads: mitigated by short local revision TTL, batched reads, and background watcher.
- Returning errors on permission invalidation publish failure can make management operations fail: this avoids cross-node authorization inconsistency; errors are structured and retryable.
- `sys_kv_cache` as MEMORY means plugin cache lost on database restart: this is intended cache semantics, forbidden for reliable business state.
- Same-version refresh with checksum paths increases artifact retention: cleanup based on release state and retention windows.

## Plugin Runtime Cache Coordination

### Dynamic Plugin Cache Invalidation

After plugin install, enable, disable, uninstall, upgrade, or same-version refresh, the system uses `cachecoord` to invalidate or refresh plugin runtime derived caches on all nodes. Non-primary nodes observing a plugin runtime revision change must:

- Refresh the plugin enabled snapshot
- Invalidate plugin frontend bundle cache
- Invalidate runtime i18n bundle cache
- Invalidate Wasm compilation cache

### Wasm Compilation Cache Checksum Binding

Cache keys must not depend only on `pluginID/version`. Implementation uses either:
- Archive paths containing checksum: `releases/<plugin>/<version>/<checksum>/<artifact>`
- Wasm cache keys using `artifactPath@checksum`

When non-primary nodes observe a plugin runtime revision change, they discard old derived caches and rebuild from the current release table plus artifact checksum.

### Old Artifact Cleanup

Old dynamic plugin artifacts can be cleaned according to retention policy, but the artifact referenced by the current active release must not be deleted.

## Pluginbridge Subcomponent Architecture

### Motivation

`pkg/pluginbridge` is shared infrastructure for dynamic plugin Wasm bridge, used by host runtime, WASM host functions, `plugindb`, dynamic plugin examples, and guest code. The flat package structure mixes:

- Stable ABI and manifest contracts: `BridgeSpec`, `RouteContract`, `CronContract`, `HostServiceSpec`
- Bridge envelope codecs: request, response, route, identity, HTTP snapshot, protobuf wire tools
- WASM artifact helpers: custom section constants and reading
- Host call protocol: opcodes, host_call envelope, status codes
- Host service protocol: runtime, storage, network, data, cache, lock, config, notify, cron payloads and codecs
- Guest SDK: guest runtime, controller dispatcher, context, BindJSON, WriteJSON, host service client helpers

### Subcomponent Structure

```
pkg/pluginbridge/
  pluginbridge.go      # Root facade: aliases + wrappers
  contract/            # ABI, route, cron, execution source contracts
  codec/               # Bridge request/response envelope encoding/decoding
  artifact/            # Wasm section constants, custom section reading, runtime metadata
  hostcall/            # Host call opcodes, generic host call envelope, status codes
  hostservice/         # Host service spec, capability derivation, payload codecs
  guest/               # Guest runtime, controller dispatcher, BindJSON, host service clients
```

### Dependency Direction

```
contract
  ^
codec -> internal/wire
  ^
artifact -> internal/wasmsection
  ^
hostservice -> contract, codec/internal wire
  ^
hostcall -> hostservice
  ^
guest -> contract, codec, hostcall, hostservice
  ^
pluginbridge facade -> all subcomponents
```

Bottom-level packages must not import root facade or guest SDK. Any subcomponent's `internal` package serves only that subcomponent or its parent path.

### Root Facade Compatibility

Root package `pluginbridge` continues to expose existing constants, types, and functions via:
- `type X = contract.X`
- `const X = contract.X`
- `func EncodeRequestEnvelope(...) { return codec.EncodeRequestEnvelope(...) }`
- `func Runtime() guest.RuntimeHostService { return guest.Runtime() }`

Host internal code migrates to precise subcomponent imports. Plugin guest code can continue using root facade.

### Verification Requirements

- `EncodeRequestEnvelope` / `DecodeRequestEnvelope` byte-level round trip unchanged
- All host service payload `Marshal` / `Unmarshal` round trips unchanged
- WASM custom section reading error boundaries unchanged
- `HostCallResponseEnvelope` and structured host service envelope unchanged
- Guest runtime, typed controller dispatcher, BindJSON/WriteJSON behavior unchanged
- Root facade and subcomponent direct calls produce identical results

## Ancillary Fixes

### Login homePath Fallback

`/user/info` previously returned a hardcoded `/analytics` homePath. Now it returns the user's first accessible menu route. If the user has no accessible business menus, it falls back to a registered safe page instead of redirecting to 404.

### Menu Management Permission Identifier

Menu and button type menus require permission identifiers. The permission identifier input displays below the menu name in the create/edit drawer. Directory type menus do not require permission identifiers.

### Online User List Pagination

The online user list API now supports pagination parameters instead of returning all records for client-side pagination.

### Role Assign User Page

Role assign user page aligns with `ruoyi-plus-vben5` reference implementation: adds bulk revoke button in toolbar, adds email column, adjusts column widths for complete timestamp display.
