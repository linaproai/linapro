## Context

LinaPro currently has a set of foundational distributed capabilities, including:

- `cluster.Service`: Exposes `cluster.enabled`, node ID, primary determination, and starts leader election in cluster mode.
- `locker`: Implements distributed lock, lease renewal, and leader election through `sys_locker` table.
- `cachecoord`: Implements cache domain revision increment, request path freshness check, and local derived cache refresh through `sys_cache_revision` table.
- `kvcache`: Already has backend/provider abstraction, but current default implementation only has SQL table backend using `sys_kv_cache` for plugin/host KV cache.
- `auth` token revoke: Currently stores revoked token short TTL state through `kvcache`.
- `session`: Currently uses `sys_online_session` as online session storage, request path validates and throttles updates to `last_active_time`.
- `role`, `config`, `pluginruntimecache`, `i18n`, `plugin frontend bundle` and other derived caches already connected to shared revision through `cachecoord`.

These implementations satisfy standalone deployment and early multi-node fallback, but have obvious problems in true multi-instance deployment:

1. PostgreSQL is used to carry high-frequency coordination writes such as lock renewal, cache revision increment, KV cache TTL read/write, token revoke, and session activity refresh.
2. `sys_cache_revision` uses table rows as event source, cross-node real-time dependency on request path checks or periodic tasks, unable to achieve low-latency active push.
3. `sys_kv_cache` needs cleanup of expired data, and generic KV cache read/write hotspots compete with authoritative business data for database resources.
4. Distributed locks and leader election use SQL tables to simulate leases, semantically achievable but less natural than Redis native atomic operations and TTL.
5. JWT revoke, `pre_token`, one-time token, and other short-lived TTL security states are more suited to Redis TTL keys; using SQL cache increases main database pressure and extends cross-node convergence chain.
6. Online sessions simultaneously handle request hot path validation and management query projection, needing to split "hot state" and "management view" to avoid high-concurrency requests continuously touching `sys_online_session`.

This is a new project with no historical compatibility burden. This change can directly adjust configuration contracts and cluster mode implementation strategy without needing to preserve the old "cluster mode only depends on PostgreSQL" runtime form.

## Goals / Non-Goals

**Goals:**

- When `cluster.enabled=false`, maintain existing lightweight experience: no Redis required, no Redis connection, continue using PostgreSQL authoritative data with process-internal cache/local revision.
- When `cluster.enabled=true`, mandate `cluster.coordination: redis`, currently the only supported coordination backend is Redis.
- Establish internal unified `coordination` abstraction; business modules only depend on narrow interfaces for lock, KV, revision, event, health, not directly depending on Redis client.
- Migrate cluster-mode high-frequency, short-lived, rebuildable coordination state from PostgreSQL to Redis.
- Maintain PostgreSQL as authoritative business data source; Redis only stores derived state, coordination state, short-lived TTL token state, and rebuildable hot state.
- Clarify security-critical path Redis failure strategy: fail-closed or conservative-hide, not silently pass.
- Clarify lossy cache Redis failure strategy: read failure can degrade as cache miss, write failure must not fake success.
- Clarify Redis key, event, revision, lock namespace, tenant, scope, plugin, node dimensions, avoiding key conflicts and cross-tenant pollution.
- Preserve current `sys_cache_revision`, `sys_kv_cache`, `sys_locker` tables as standalone/testing/diagnostics/future fallback implementation boundary, but cluster mode must not depend on them for cross-node consistency.
- Provide detailed testing strategy including fake coordination provider, Redis real connection integration tests, configuration validation, cluster behavior, and security failure tests.

**Non-Goals:**

- Do not introduce Redis Cluster, Sentinel, ACL, TLS, or connection pool advanced tuning complete operational encapsulation; first version only defines single Redis endpoint configuration and timeout.
- Do not introduce etcd, NATS, Consul, or PostgreSQL LISTEN/NOTIFY as peer implementations; interface reserves extension, but currently only implements Redis.
- Do not migrate business authoritative data to Redis; users, roles, menus, tenants, plugin registry, system configuration, audit, notification, tasks, and file metadata remain authoritative in PostgreSQL.
- Do not modify business REST API semantics; only health check, system info, or apidoc diagnostic field extensions are allowed.
- Do not implement cross-datacenter strong consistency, multi-Redis datacenter replication, or disaster recovery switching strategy; these belong to subsequent deployment governance.
- Do not require standalone mode to support Redis cache; standalone mode maintains minimal dependencies, avoiding complicating development experience.
- Do not implement user-configurable per-store backend such as `stores.lock`, `stores.kvCache`; internally use `cluster.coordination` to select provider.

## Decisions

### Configuration form uses cluster.coordination: redis

```yaml
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: ""
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

Rules:
- `cluster.enabled=false`: Ignore `cluster.coordination` and `cluster.redis`, do not probe Redis.
- `cluster.enabled=true`: `cluster.coordination` required, currently only allows `redis`.
- `cluster.enabled=true` with empty, missing, or non-`redis` `coordination`: Startup failure.
- `coordination=redis` must validate Redis configuration and complete Redis probe before starting services.
- SQLite mode continues forcing `cluster.enabled=false`, so SQLite mode does not allow actually enabling Redis coordination.

### New internal coordination component as sole Redis access layer

New internal component `apps/lina-core/internal/service/coordination/`, exposing narrow interfaces:

```text
coordination.Service
├─ LockStore
├─ KVStore
├─ RevisionStore
├─ EventBus
└─ HealthChecker
```

Dependency direction:

```text
config ──┐
         ▼
 coordination(redis provider)
         ▲
         │
 cluster / locker / cachecoord / kvcache / auth / session / role / config / pluginruntimecache
```

Business modules must not directly import Redis client package. Redis client only allowed in coordination's Redis provider or dedicated internal sub-package.

### Redis key namespace uses stable, diagnosable, deletable hierarchy

Unified key prefix:

```text
linapro:{app}:{env}:{component}:{scope...}
```

Rules:
- All keys must explicitly contain tenant dimension; platform level uses `tenant=0`.
- Business input must be encoded/escaped, prohibiting unsanitized user input directly concatenated into Redis key.
- Key builder must be centralized in coordination or corresponding adapter, prohibiting business modules from hand-writing keys.
- Deletion and invalidation must operate by explicit scope, prohibiting ordinary business path full-database `FLUSHDB` or scanning all LinaPro keys.

### Redis event uses "Pub/Sub fast notification + revision fallback"

Cross-node invalidation uses two-layer mechanism:

1. Write path executes authoritative data write, then calls `RevisionStore.Bump` to increment Redis revision.
2. Same transaction success publishes event at business boundary, notifying other nodes to immediately refresh local cache.
3. Read path or periodic watcher still verifies revision through `RevisionStore.Current`, compensating Pub/Sub message loss or node offline window.

Reliability strategy:
- Pub/Sub for low-latency notification, not as sole source of truth.
- Revision is authoritative coordination state for cross-node consistency.
- Event processing must be idempotent; duplicate events must not cause errors.

### Distributed lock uses Redis SET NX PX + owner token + compare-and-delete

Redis lock acquisition: `SET lockKey ownerToken NX PX leaseMillis`

Renewal: Use Lua or equivalent atomic compare-and-pexpire, only renewing when current value equals owner token.

Release: Use Lua or equivalent atomic compare-and-delete, only deleting when current value equals owner token.

Lock handle must contain: lock name, owner token, node ID, lease duration, acquired at / expire at diagnostic data.

Leader election: `cluster` primary election uses fixed lock name `leader-election`. Successful lock acquisition makes node primary. Renewal failure, Redis unavailability, or owner token mismatch immediately demotes.

### kvcache cluster mode uses coordination KV backend

`kvcache.New()` selects backend by runtime mode:
- `cluster.enabled=false`: SQL table backend or existing default.
- `cluster.enabled=true`: coordination KV backend driven by coordination provider's KV capability.

TTL rules: `ttl < 0` returns business parameter error; `ttl = 0` does not set backend TTL but still not authoritative business state; `ttl > 0` uses coordination KV backend native expiration.

### cachecoord cluster mode migrates to Redis revision/event

`cachecoord.Service` maintains stable upper interface, internally branching by topology/coordination provider:
- Standalone mode: Process-internal revision + local refresh.
- Cluster mode: Redis `RevisionStore` + `EventBus`.

### Auth short TTL state unifies into coordination KV

Cluster mode must store in Redis: JWT revoked token ID, `pre_token`, select-tenant single-use marker, login verification code, one-time auth challenge.

Security strategy: Write revoke failure must not report complete success; read revoke failure must fail-closed, prohibiting default pass on Redis failure.

### Online session adopts Redis hot state + PostgreSQL management projection

Split online session responsibility:
- Redis hot state: Request path token/session validate, last active, force-logout marker, user token index.
- PostgreSQL `sys_online_session`: Online user list, data permission filtering, login time, IP/browser/os, management query projection, audit helper.

### Plugin runtime and dynamic plugin reconciler use Redis revision/event wake-up

Plugin-related runtime changes uniformly publish `plugin-runtime` domain events. Nodes receiving events refresh enabled snapshot, frontend bundle cache, runtime i18n bundle cache, and Wasm module cache.

### Cron and built-in sync tasks adjust based on Redis capability

Master-only jobs continue using `cluster.Service.IsPrimary()` determination, just primary source changes to Redis lock. KV cache expired cleanup not registered under coordination KV backend. Access topology sync, runtime param sync, and other cluster watchers remain as Redis event revision fallback.

### Observability and health checks

New or extended system info/health snapshot includes: `cluster.enabled`, `cluster.coordination`, Redis ping status, latency, recent error, recent success time, lock store state, revision store state, event bus state, kvcache backend name, session hot state backend.

### Error model and degradation boundary

| Category | Example | Strategy |
|----------|---------|----------|
| Configuration error | Cluster enabled but coordination missing | Startup failure |
| Redis connection unavailable | Startup ping failure | Cluster mode startup failure |
| Security state read failure | Revoke/session/pre-token Redis read error | fail-closed |
| Permission/config revision unreadable | `permission-access` / `runtime-config` | Fail beyond MaxStale |
| Plugin runtime revision unreadable | `plugin-runtime` | conservative-hide or visible error |
| Lossy KV cache read failure | Plugin cache get | Cache miss or visible error |
| Lossy KV cache write failure | Plugin cache set/incr | Return error, not fake success |

All caller-visible errors must use `bizerr` encapsulation.

### Data permission and tenant boundary

Redis coordination does not change data permission authoritative boundary. Redis keys must carry tenant dimension. Platform-level `tenant=0` invalidation can cascade per existing `cascadeToTenants` semantics, but ordinary business paths must not use full-tenant clear.

### Testing strategy

Unit tests prioritize fake provider. Redis real connection integration tests use `LINA_TEST_REDIS_ADDR` environment variable to explicitly enable. Module regression tests cover cluster, locker, cachecoord, kvcache, auth, session, role, config, pluginruntimecache, cron, sysinfo/health.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Redis becomes cluster mode new required dependency, increasing deployment complexity | Only mandate when `cluster.enabled=true`; standalone mode does not depend on Redis; configuration template and documentation clearly explain |
| Redis single point of failure causes auth or permission path unavailability | Security path fail-closed; health check exposes failure; future extendable Sentinel/Cluster |
| Pub/Sub message loss causes node local cache not timely cleared | Revision as authoritative coordination state, read path/periodic watcher fallback |
| Redis key design inconsistency causes cross-tenant pollution or cleanup difficulty | Centralized key builder, all keys explicitly contain tenant/scope, prohibit business modules hand-writing keys |
| Simultaneously maintaining PostgreSQL tables and Redis implementation causes maintenance confusion | Document and code clearly distinguish standalone/cluster branching; cluster mode must not use SQL tables for cross-node consistency |
| Session hot state and PostgreSQL projection briefly inconsistent | Clarify PG is management projection, Redis is request hot state; projection syncs by throttling and cleanup recovers |
| Redis write success but PostgreSQL authoritative write failure causes derived state premature change | Bump revision/publish event only after authoritative data write succeeds; must avoid pre-transaction publish |
| PostgreSQL write success but Redis publish failure causes cross-node delay | Critical path returns error or relies on revision fallback; implementation phase decides by domain failure strategy |
| Plugin host cache using Redis exposes more capacity risk | Governed through TTL, namespace, value size limit, and subsequent metrics; plugin cache remains lossy |
| New Redis client dependency may affect build and cross-compilation | Choose pure Go, actively maintained, context-supporting client; included in `go test ./...` and image build |

## Migration Plan

### Development Migration

1. Add configuration structure and validation, but default `cluster.enabled=false`, ensuring standalone development unaffected.
2. Add coordination provider interface and fake provider tests.
3. Add Redis provider, implementing health, KV, lock, revision, event basic capabilities.
4. Create coordination service in startup orchestration, injecting into cluster/locker/cachecoord/kvcache/auth/session and other modules.
5. Switch cluster mode implementation module by module, keeping standalone branch unchanged.
6. Supplement module tests and Redis optional integration tests.
7. Update configuration templates, README/Chinese README, or deployment documentation.
8. Run `openspec validate`, related Go unit tests, necessary E2E.

### Deployment Migration

Standalone deployment: No configuration change needed. `cluster.enabled=false` does not require Redis.

Cluster deployment: Deploy Redis. Configure `cluster.enabled=true`, `cluster.coordination: redis`, `cluster.redis` with address, db, password, and timeout settings. Confirm Redis reachable before startup. Rolling start nodes; only nodes with Redis coordination available can enter service state.

### Rollback Strategy

- Standalone mode rollback: Set `cluster.enabled=false`, application recovers to PostgreSQL + process cache mode.
- Cluster mode rollback to old implementation not supported as a target, as this change explicitly changes cluster mode contract; if rollback required, need to revert code version and restore old configuration.
- Redis short-term failure recovery: After fixing Redis, nodes reconverge through revision watcher and request path; session hot state may require user re-login, which is acceptable security degradation.
