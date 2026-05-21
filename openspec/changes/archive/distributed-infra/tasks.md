## 1. Configuration Model and Startup Validation

- [x] 1.1 Extend `internal/service/config` cluster configuration structure, adding `coordination` field and Redis configuration structure, all timeout fields using `time.Duration`
- [x] 1.2 Update `manifest/config/config.template.yaml`, adding `cluster.coordination: redis` and `cluster.redis` example, with comments clarifying standalone mode does not need Redis, cluster mode must configure Redis
- [x] 1.3 Implement configuration validation when `cluster.enabled=true`: `coordination` required, currently only allows `redis`, Redis address required, timeout must be duration string with unit
- [x] 1.4 Maintain SQLite dialect forcing `cluster.enabled=false`, ensuring SQLite mode does not connect Redis even when Redis coordination is configured
- [x] 1.5 Add coordination initialization phase in HTTP runtime startup orchestration, ensuring Redis probe succeeds before starting cluster, cron, plugin runtime, and HTTP services
- [x] 1.6 Add unit tests for configuration parsing, illegal coordination, missing Redis address, illegal timeout, SQLite ignoring Redis configuration
- [x] 1.7 Define startup configuration error codes or startup diagnostic errors, ensuring users can clearly see failure fields and fix suggestions

## 2. Coordination Provider Abstraction

- [x] 2.1 Add `internal/service/coordination` package, defining `Service`, `Provider`, `LockStore`, `KVStore`, `RevisionStore`, `EventBus`, `HealthChecker` and other narrow interfaces
- [x] 2.2 Define coordination common types: backend name, lock handle, fencing token, revision key, event payload, tenant invalidation scope, health snapshot
- [x] 2.3 Implement centralized Redis key builder, unifying namespace, tenant, domain, scope, owner, plugin, node dimension encoding, prohibiting business modules from hand-writing Redis keys
- [x] 2.4 Implement fake/in-memory provider for unit test coverage of lock, KV, revision, event, health semantics
- [x] 2.5 Define coordination error classification and `bizerr` error codes: configuration error, connection error, lock not held, revision unavailable, event publish failure, KV operation failure
- [x] 2.6 Add interface-level unit tests for provider abstraction, covering key generation, tenant isolation, context cancel, error classification, and health snapshot

## 3. Redis Provider Implementation

- [x] 3.1 Select and integrate Redis Go client, requiring support for context, connection pool, timeout, `SET NX PX`, Lua or equivalent atomic compare operations, Pub/Sub
- [x] 3.2 Implement Redis connection initialization, authentication, DB selection, timeout configuration, Ping probe, and close flow
- [x] 3.3 Implement Redis `LockStore`: Acquire, Renew, Release, IsHeld, owner token validation, TTL, optional fencing token
- [x] 3.4 Implement Redis `KVStore`: Get, Set, SetNX, Delete, IncrBy, Expire, TTL, CompareAndDelete
- [x] 3.5 Implement Redis `RevisionStore`: Atomic Bump and Current by tenant/domain/scope, supporting cascade metadata
- [x] 3.6 Implement Redis `EventBus`: Publish cache invalidation event, subscribe loop, duplicate event idempotent processing, source node identification
- [x] 3.7 Implement Redis health snapshot: ping status, recent success time, recent error, subscriber status, backend name
- [x] 3.8 Add Redis real connection integration tests, explicitly enabled via `LINA_TEST_REDIS_ADDR`, using independent namespace, prohibiting `FLUSHDB`
- [x] 3.9 Add Redis provider failure tests, covering connection failure, timeout, owner token mismatch, event publish failure, and context cancel

## 4. Cluster and Leader Election Migration

- [x] 4.1 Modify `internal/service/cluster` constructor, cluster mode receiving coordination lock store, standalone mode maintaining no Redis dependency
- [x] 4.2 Switch leader election from `locker.New()` SQL locker to coordination lock, using fixed leader lock name, node ID, owner token, and lease TTL
- [x] 4.3 Implement leader lock renewal failure, owner token mismatch, Redis error immediately demoting and stopping primary state
- [x] 4.4 Preserve standalone mode `IsPrimary=true` semantics, not starting Redis election cycle
- [x] 4.5 Preserve or adjust SQL locker only for standalone/testing/fallback boundary, ensuring cluster mode does not depend on `sys_locker`
- [x] 4.6 Update cluster/leader election unit tests, covering Redis primary, follower, renewal failure, primary takeover, and standalone skipping Redis

## 5. Distributed Locker and Plugin Lock Migration

- [x] 5.1 Modify `internal/service/locker`, abstracting coordination lock backed implementation and existing SQL implementation deployment branching
- [x] 5.2 Implement Redis lock instance Unlock, Renew, IsHeld, ensuring release/renew both validate owner token
- [x] 5.3 Modify plugin Wasm host lock service, making cluster-mode plugin locks go through coordination lock
- [x] 5.4 Plugin lock key must include plugin ID, tenant dimension, and logical lock name; platform shared locks must require explicit capability and audit
- [x] 5.5 Add unit tests for plugin locks, covering different plugins same lock isolation, different tenants same lock isolation, non-holder release failure, Redis failure returning error
- [x] 5.6 Update `plugin-lock-service` related apidoc or host service documentation, synchronizing i18n if response error semantics change

## 6. Cachecoord Redis Revision/Event Migration

- [x] 6.1 Modify `internal/service/cachecoord`, cluster mode using coordination `RevisionStore` and `EventBus`
- [x] 6.2 Preserve standalone mode process-internal revision branch, not connecting Redis, not accessing shared revision table
- [x] 6.3 Implement `MarkTenantChanged` Redis revision bump + event publish + local observed revision update flow
- [x] 6.4 Implement `EnsureFresh` local revision TTL, Redis current revision read, refresher call, observed revision update, and domain failure strategy
- [x] 6.5 Ensure tenant scope, cascadeToTenants, tenant=-1 operation full-clear semantics explicitly expressed in Redis key/event
- [x] 6.6 Maintain `cachecoord.Snapshot` observable fields, adding Redis backend, event status, and recent error
- [x] 6.7 Add dual-instance fake provider tests for cachecoord, covering event convergence, event loss revision fallback, duplicate event idempotency, permission fail-closed
- [x] 6.8 Add optional integration tests for Redis real connection scenario, covering concurrent revision bump and cross-instance event notification

## 7. Kvcache Coordination KV Backend

- [x] 7.1 Add `kvcache` coordination KV backend provider, implementing Get, Set, Delete, Incr, Expire, CleanupExpired through coordination KVStore
- [x] 7.2 Modify kvcache default constructor or startup injection logic: standalone mode uses SQL table backend, cluster mode uses coordination KV backend
- [x] 7.3 Design and implement Redis value encoding, preserving `Item` string/int/expireAt semantics, clarifying int/string type conflict handling
- [x] 7.4 Use coordination KV backend native TTL, coordination KV backend `RequiresExpiredCleanup=false`
- [x] 7.5 Coordination KV backend write failure, delete failure, increment failure must return structured error, not fake success
- [x] 7.6 Update cron built-in task projection logic, coordination KV backend not registering `host:kvcache-cleanup-expired`
- [x] 7.7 Add kvcache coordination KV backend unit tests, covering string, int, TTL, incr concurrency, Expire, Delete, type conflict, Redis failure
- [x] 7.8 Update plugin Wasm host cache service tests, confirming cluster mode goes through coordination KV backend with tenant key isolation

## 8. Auth Token State Migration

- [x] 8.1 Modify JWT revoke store, cluster mode using coordination KV to write revoked token, TTL equal to JWT remaining lifetime
- [x] 8.2 Preserve local memory revoke cache as current node acceleration layer, but cluster mode must use Redis revoke state as cross-node source of truth
- [x] 8.3 Implement revoke read failure fail-closed, prohibiting Redis failure from passing based on JWT signature alone
- [x] 8.4 Migrate `pre_token`, select-tenant single-use state, and replay marker to coordination KV
- [x] 8.5 Confirm logout, switch-tenant, force logout all write Redis revoke, returning structured error or explicit partial failure on write failure
- [x] 8.6 Add auth unit tests, covering logout revoke, switch-tenant old token invalidation, pre-token single use, Redis read failure fail-closed
- [x] 8.7 If login/tenant selection frontend behavior is affected, update corresponding E2E sub-assertions; otherwise explain no frontend-visible changes in verification conclusion

## 9. Session Hot State Migration

- [x] 9.1 Design session hot state payload, containing token ID, tenant ID, user ID, username, login time, last active, IP/browser/os and other necessary fields
- [x] 9.2 Modify session store, cluster mode login simultaneously writing Redis hot state and `sys_online_session` PostgreSQL projection
- [x] 9.3 Modify authentication middleware, request path first verifying JWT/revoke, then reading Redis session hot state, fail-closed when Redis unreadable
- [x] 9.4 Implement Redis session TTL refresh and last active hot state update
- [x] 9.5 Implement PostgreSQL `last_active_time` throttled write-back, avoiding writing main database on every request
- [x] 9.6 Modify forced logout flow, first validating PostgreSQL projection visibility, then deleting Redis session, writing revoke, deleting or marking projection
- [x] 9.7 Preserve PostgreSQL projection cleanup task, cleaning up Redis-expired or long-inactive projection rows
- [x] 9.8 Add session unit tests, covering login dual-write, request validation, tenant mismatch, throttled write-back, forced logout, Redis failure fail-closed, projection cleanup
- [x] 9.9 Update `monitor-online` plugin tests or E2E, confirming online list and forced logout under Redis hot state model still comply with data permissions

## 10. Role Permission Cache Integration

- [x] 10.1 Modify role access revision controller, cluster mode using Redis-backed cachecoord revision/event
- [x] 10.2 Confirm role, menu, user role, role menu, plugin permission governance write paths all publish `permission-access` revision
- [x] 10.3 Confirm token access cache key and reverse index include tenant dimension
- [x] 10.4 Implement permission revision read failure beyond stale window fail-closed
- [x] 10.5 Add role unit tests, covering cross-node revision/event invalidation, same user multi-tenant permission isolation, Redis failure fail-closed

## 11. Runtime Config Cache Integration

- [x] 11.1 Modify runtime param revision controller, cluster mode using Redis-backed cachecoord revision/event
- [x] 11.2 Confirm `sys.jwt.expire`, `sys.session.timeout`, login blacklist, cron configuration and other protected parameter write paths all publish `runtime-config` revision
- [x] 11.3 Implement runtime-config Redis revision unreadable beyond stale window returning structured visible error
- [x] 11.4 Maintain standalone mode process-internal revision and local gcache snapshot behavior
- [x] 11.5 Add config unit tests, covering cross-node snapshot refresh, Redis event loss revision fallback, Redis failure error propagation, standalone mode no Redis

## 12. Plugin Runtime Cache Integration

- [x] 12.1 Modify `pluginruntimecache` controller, cluster mode underlying using Redis-backed cachecoord
- [x] 12.2 Confirm plugin install, enable, disable, uninstall, upgrade, active release switch, dynamic artifact update all publish `plugin-runtime` revision/event
- [x] 12.3 Modify dynamic plugin reconciler wake-up, using Redis revision/event to trigger and retaining safety sweep fallback
- [x] 12.4 Ensure receiving plugin-runtime event refreshes enabled snapshot, frontend bundle, runtime i18n, and Wasm derived cache
- [x] 12.5 Implement plugin-runtime freshness unconfirmable conservative-hide or structured error, not exposing possibly disabled/uninstalled plugins
- [x] 12.6 Add plugin runtime unit tests, covering cross-node enable/disable, event loss fallback, reconciler revision, frontend/i18n/wasm cache invalidation

## 13. Cron and Built-in Task Adjustment

- [x] 13.1 Modify Master-Only task determination, ensuring based on Redis leader lock primary state
- [x] 13.2 Redis kvcache backend not projecting KV SQL expired cleanup job
- [x] 13.3 Preserve access topology sync, runtime param sync, plugin runtime sync watcher as Redis event revision fallback
- [x] 13.4 Ensure session cleanup continues cleaning PostgreSQL projection, while Redis session hot state expires by TTL
- [x] 13.5 Add cron unit tests, covering primary execution, follower skip, coordination KV backend not registering KV cleanup, watcher using Redis revision

## 14. System Info, Health Check, and Observability

- [x] 14.1 Extend system info or health response, exposing coordination backend, Redis ping status, node ID, primary status, recent error
- [x] 14.2 Extend cachecoord snapshot, exposing Redis shared revision, event subscriber status, last sync time, stale seconds
- [x] 14.3 Ensure diagnostic response does not expose Redis password, complete sensitive connection string, or token key
- [x] 14.4 Synchronize apidoc i18n JSON, covering new coordination/redis diagnostic fields
- [x] 14.5 If frontend displays new diagnostic fields, synchronize frontend runtime language pack and host manifest i18n
- [x] 14.6 Add sysinfo/health unit tests or interface tests, covering Redis healthy/unhealthy, sensitive information desensitization, and apidoc i18n existence

## 15. Documentation and Deployment Instructions

- [x] 15.1 Update related README/README.zh-CN or deployment documentation, explaining standalone mode does not need Redis, cluster mode must configure Redis
- [x] 15.2 Update development environment documentation, explaining Redis integration tests explicitly enabled via `LINA_TEST_REDIS_ADDR`
- [x] 15.3 Update configuration examples and comments, explaining current coordination only supports `redis`, future extendable to other backends
- [x] 15.4 Check whether new or modified documentation needs Chinese/English README synchronization, following markdown format specifications

## 16. Regression Testing and Verification

- [x] 16.1 Run `cd apps/lina-core && go test ./internal/service/config ./internal/service/coordination ./internal/service/cluster ./internal/service/locker ./internal/service/cachecoord ./internal/service/kvcache -count=1`
- [x] 16.2 Run `cd apps/lina-core && go test ./internal/service/auth ./internal/service/session ./internal/service/role ./internal/service/config ./internal/service/cron -count=1`
- [x] 16.3 Run `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/pluginruntimecache ./internal/service/i18n ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`
- [x] 16.4 Run explicit integration tests with Redis available: `cd apps/lina-core && LINA_TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service/coordination ./internal/service/cachecoord ./internal/service/kvcache ./internal/service/session -run Redis -count=1`
- [x] 16.5 If implementation affects online users, login, system info pages, add or update corresponding E2E per `lina-e2e` specification
- [x] 16.6 Run `openspec validate --strict`
- [x] 16.7 Run `git diff --check`
- [x] 16.8 Complete implementation and call `lina-review` for code, specification, i18n, cache consistency, and data permission review

## Feedback

- [x] **FB-1**: Redis provider package boundary needs to converge to `coordination/internal/redis`, `kvcache` only retaining coordination KV adapter layer, avoiding business cache layer directly expressing Redis backend
- [x] **FB-2**: Project introduction documentation still describes `OpenSpec` as built-in required workflow, should adjust to optional but recommended dependency component with good framework support
- [x] **FB-3**: `lina-archive-consolidate` should only read date-prefixed archived changes when no change list is specified, avoiding re-aggregating already-generated aggregate archive directories
- [x] **FB-4**: Main CI lacks active OpenSpec change completion status check, allowing incomplete changes to enter main CI pass path
- [x] **FB-5**: `cmd_http_routes.go` new route registration controllers should define variables first, avoiding directly initializing objects in route binding parameters
- [x] **FB-6**: sysinfo controller should not simultaneously retain `NewV1` and `NewV1WithDiagnostics` two initialization entries, should unify to injectable diagnostic dependency `NewV1`
- [x] **FB-7**: `runtime.Service` interface has too many methods, should split into narrow interfaces by runtime responsibility and compose through embedding
- [x] **FB-8**: Main CI and Nightly Build lack real Redis service and cluster startup smoke, Redis coordination regression only depends on local manual verification
- [x] **FB-9**: Redis CI and cluster smoke workflow/script lack necessary maintenance comments, making it difficult for subsequent maintainers to understand service dependencies, environment variables, and assertion boundaries
- [x] **FB-10**: Main CI regular Go unit tests and SQLite smoke do not correctly isolate Redis coordination test boundary
