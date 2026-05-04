## 1. Data model and coordination foundation

- [x] 1.1 Adjust host delivery SQL to add persistent `sys_cache_revision` and preserve the lossy `MEMORY` cache semantics of `sys_kv_cache`
- [x] 1.2 Add idempotent indexes, unique keys, and concurrent-increment constraints required by `sys_cache_revision`; review existing TTL query indexes and unique keys on `sys_kv_cache`
- [x] 1.3 Run `make init` and `make dao`, updating DAO/DO/Entity artifacts only through code generation
- [x] 1.4 Review current `cluster.enabled` and `cluster.Service` topology abstractions and determine single-node and cluster-mode branch integration points for `cachecoord`

## 2. Unified cache coordination component

- [x] 2.1 Add `internal/service/cachecoord`, defining cache domain, scope, authoritative source, consistency model, staleness window, and failure policy
- [x] 2.2 Implement core interfaces including `ConfigureDomain`, `MarkChanged`, `EnsureFresh`, and `Snapshot`
- [x] 2.3 In `cluster.enabled=false`, implement in-process revision, local invalidation, and synchronous refresh without relying on a shared coordination table
- [x] 2.4 In `cluster.enabled=true`, implement atomic shared revision increments, idempotent publish, request-path freshness checks, and watcher synchronization
- [x] 2.5 Define `bizerr` codes for cache coordination failures and log observable errors with the project `logger` component and propagated `ctx`

## 3. Critical cache-domain integration

- [x] 3.1 Connect protected runtime parameter cache to `cachecoord` and reliably publish the `runtime-config` revision after parameter write transactions
- [x] 3.2 Execute freshness checks on runtime parameter reads and return visible errors according to the domain policy when the failure window is exceeded
- [x] 3.3 Connect role, menu, user-role, and plugin permission topology write paths to `permission-access` revision publishing
- [x] 3.4 Execute permission snapshot freshness checks before protected API permission validation, and fail closed when freshness cannot be confirmed after the failure window
- [x] 3.5 Connect plugin install, enable, disable, uninstall, upgrade, and same-version refresh to `plugin-runtime` revision publishing

## 4. Plugin runtime derived caches

- [x] 4.1 Refactor shared revision logic in `pluginruntimecache` and reuse `cachecoord` for plugin enabled snapshot refresh
- [x] 4.2 Include plugin frontend bundle cache, runtime i18n bundle cache, and dynamic route derived cache in `plugin-runtime` scoped invalidation
- [x] 4.3 Bind Wasm compilation cache keys to the active release checksum or generation so same-version refresh cannot keep hitting old cache
- [x] 4.4 Adjust dynamic plugin artifact archive or cache-key strategy so other nodes can verify the active release by checksum or generation
- [x] 4.5 Add old artifact and old Wasm compilation cache cleanup without deleting content referenced by the current active release

## 5. Plugin host-cache reliability

- [x] 5.1 Remove `sys_kv_cache` from host critical revision paths so it only stores lossy plugin/module KV cache data
- [x] 5.2 Implement `incr` with one SQL atomic update or equivalent behavior, ensuring concurrent increments from multiple nodes do not lose deltas while the shared database is alive
- [x] 5.3 Return structured errors for non-integer increments and oversized namespaces, keys, or values; do not truncate or partially write data
- [x] 5.4 Change expiration cleanup to single-key lazy miss handling plus background batch cleanup, avoiding full table scans on ordinary read/write paths
- [x] 5.5 In cluster mode, limit expired-row batch cleanup pressure through primary-node coordination or idempotent batching, and treat database restart as cache miss

## 6. Observability, tests, and acceptance

- [x] 6.1 Expose a cache coordination status snapshot with at least domain, scope, local revision, shared revision, last sync time, latest error, and stale seconds
- [x] 6.2 If new HTTP diagnostic endpoints or API documentation fields are added, maintain apidoc i18n resources; if the change only integrates health checks or logs, record no runtime UI i18n impact
- [x] 6.3 Add concurrent publishing tests for `sys_cache_revision`, verifying persistent atomic increments and no loss after database restart
- [x] 6.4 Add dual-instance service-level tests covering cross-node invalidation and bounded staleness for runtime parameters, permission topology, and plugin runtime cache
- [x] 6.5 Add plugin host-cache tests for concurrent `incr`, TTL, cache miss after database restart, oversized input, and non-integer increments
- [x] 6.6 Add dynamic-plugin same-version refresh tests verifying old Wasm, frontend bundle, and i18n derived caches are invalidated after checksum or generation changes
- [x] 6.7 Run backend unit tests, required service-level tests, and `openspec status --change improve-distributed-cache-consistency`, confirming the change is applicable

Main-task i18n impact judgment: added cache coordination diagnostic fields in the `/system/info` response and synchronized `zh-CN` and `zh-TW` apidoc i18n JSON; no frontend page, button, menu, or runtime UI copy was added. Runtime error copy impacts during feedback are recorded separately in FB-2.

## Feedback

- [x] **FB-1**: Remove remaining `kvcache` fallback from critical cache revision paths so `role`, `config`, and `pluginruntimecache` coordinate distributed revisions only through `cachecoord`
- [x] **FB-2**: Remove `recover` fallback from runtime parameter snapshot reads so runtime-config freshness and load failures propagate as explicit errors
- [x] **FB-3**: Converge `cachecoord` multi-instance construction so in-process critical cache coordination state is managed by one coordinator
- [x] **FB-4**: Remove `cachecoord` dependency on built-in domain allowlists and prior registration, allowing host modules and plugins to directly use new legal cache domains
- [x] **FB-5**: Converge `kvcache` into a generic KV cache foundation with backend/provider abstraction, switch TTL to `time.Duration`, add the hourly MySQL backend expiration cleanup job, and ensure query paths do not delete expired rows
- [x] **FB-6**: Move MySQL `MEMORY` backend implementation from the `kvcache` facade package into an independent `kvcache/internal` package so default backend details do not pollute the generic service contract
- [x] **FB-7**: Further isolate the MySQL `MEMORY` backend into `kvcache/internal/mysql-memory`, separating backend implementations for future Redis provider extension

FB-2 i18n impact judgment: no API docs, frontend UI copy, menus, buttons, or plugin manifest resources were added or changed. Because caller-visible `bizerr` runtime error codes were added, `en-US`, `zh-CN`, and `zh-TW` `error.json` plus packed manifest resources were synchronized.

FB-3 i18n impact judgment: this only adjusts backend cache coordination object construction and in-process topology reuse. It does not add or modify API docs, frontend UI copy, menus, buttons, plugin manifests, or runtime translation resources.

FB-3 cache consistency judgment: real cache data continues to live in each business domain's process-level cache, and authoritative data sources are unchanged. `cachecoord.Default` only unifies critical cache-domain revision observation state, topology view, and diagnostics. Single-node mode continues to use local revisions and local invalidation. Cluster mode continues to use shared `sys_cache_revision`, request-path freshness checks, and watcher synchronization for cross-instance invalidation. Maximum staleness windows and failure fallback strategies remain those configured for `runtime-config`, `permission-access`, and `plugin-runtime`. The new default coordinator avoids duplicate freshness state in the same process while preserving `cachecoord.New` for tests or explicit isolation.

FB-4 i18n impact judgment: this only adjusts backend cache coordination domain admission and policy configuration. It does not add or modify API docs, frontend UI copy, menus, buttons, plugin manifests, runtime error codes, or translation resources.

FB-4 cache consistency judgment: `cachecoord` no longer treats a built-in domain list or pre-registration as an admission gate. Any legal domain/scope still uses process-local revision, local invalidation, and synchronous refresh in single-node mode, and persistent `sys_cache_revision`, request-path freshness checks, and watcher synchronization in cluster mode. Unconfigured domains use default authority notes, shared-revision consistency model, 5 second maximum staleness window, and visible-error fallback. `runtime-config`, `permission-access`, and `plugin-runtime` authoritative sources, maximum staleness windows, and fallback policies are configured by their owning business code, avoiding `cachecoord` or manifest changes when plugins or future modules add domains while preserving observability and reviewability for critical domains.

FB-5 i18n impact judgment: this added one host built-in scheduled job plus handler display name and description, and added runtime error code `CRON_KVCACHE_DEPENDENCY_MISSING`; `manifest/i18n/<locale>/job.json`, `manifest/i18n/<locale>/error.json`, and packed manifest mirrors were synchronized. No frontend page, button, menu, or API documentation field was added.

FB-5 cache consistency judgment: `kvcache` authoritative semantics are lossy KV cache, not reliable business state. The current default backend authority is the shared MySQL `MEMORY` table `sys_kv_cache`, and its consistency model is read/write visibility through that shared cache table. Database restart or `MEMORY` table clearing is recovered as cache miss. Writes, deletes, increments, and expiration updates still apply directly to the shared cache table. Query paths only return miss for expired entries and no longer delete expired rows. For the MySQL backend, the maximum expired-row retention window is the one-hour schedule of the built-in primary-node `host:kvcache-cleanup-expired` task. If the task fails, expired rows can keep occupying cache table space, but read paths still never return expired values; the next task run or manual trigger recovers cleanup. Future Redis backends can rely on native TTL and implement `CleanupExpired` as no-op without projecting the MySQL cleanup task. Critical permission, configuration, and plugin-runtime revisions remain forbidden from using `kvcache` and continue to be coordinated by `cachecoord` and `sys_cache_revision`.

FB-6 i18n impact judgment: this only adjusts backend `kvcache` package structure by moving MySQL `MEMORY` backend implementation into `kvcache/internal`. It does not add or modify API docs, frontend UI copy, menus, buttons, plugin manifests, runtime error codes, or translation resources.

FB-6 cache consistency judgment: this does not change `kvcache` authoritative data source, consistency model, TTL semantics, background cleanup strategy, or failure fallback. The public `kvcache` facade still accesses the default MySQL `MEMORY` backend through backend/provider abstraction. Read paths still only filter expiration, and write paths plus the built-in cleanup task keep the FB-5 semantics. Package convergence only reduces coupling between default implementation and generic contract.

FB-7 i18n impact judgment: this only changes the package path of the backend MySQL `MEMORY` implementation from `kvcache/internal` to `kvcache/internal/mysql-memory`. It does not add or modify API docs, frontend UI copy, menus, buttons, plugin manifests, runtime error codes, or translation resources.

FB-7 cache consistency judgment: this does not change `kvcache` authoritative data source, consistency model, TTL semantics, background cleanup strategy, or failure fallback. The public `kvcache` facade still accesses the default MySQL `MEMORY` backend through backend/provider abstraction. The MySQL `MEMORY` implementation directory split only affects code organization and does not change read-path expiration filtering, write-path changes, hourly cleanup, or future Redis backend semantics.
