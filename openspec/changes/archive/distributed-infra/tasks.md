## 1. Distributed Locker Database and Configuration

- [x] 1.1 Create `manifest/sql/010-distributed-locker.sql`, define `sys_locker` table (MEMORY engine)
- [x] 1.2 Run `make init` to sync SQL changes to database
- [x] 1.3 Run `make dao` to generate DAO/DO/Entity code
- [x] 1.4 Add `locker` configuration in `manifest/config/config.yaml` (lease duration, renewal interval)
- [x] 1.5 Create `internal/service/config/config_locker.go`, implement configuration reading logic

## 2. Distributed Locker Component

- [x] 2.1 Create `internal/service/locker/locker.go`, implement core lock service (Lock, TryLock, IsLeader)
- [x] 2.2 Create `internal/service/locker/locker_instance.go`, implement lock instance (Unlock, Renew)
- [x] 2.3 Create `internal/service/locker/locker_lease.go`, implement lease renewal management
- [x] 2.4 Create `internal/service/locker/locker_election.go`, implement leader election (Start, Stop)

## 3. Cron Task Classification

- [x] 3.1 Modify `internal/service/cron/cron.go`, add locker service dependency and isLeader state
- [x] 3.2 Modify `internal/service/cron/cron_session.go`, add leader-node check logic (Master-Only)
- [x] 3.3 Modify `internal/service/cron/cron_servermon_cleanup.go`, add leader-node check logic (Master-Only)
- [x] 3.4 Confirm `internal/service/cron/cron_servermon.go` as All-Node task (no modification needed)

## 4. Leader Election Startup Integration

- [x] 4.1 Modify `internal/cmd/cmd_http.go`, initialize locker service and start leader election

## 5. Distributed Locker Unit Tests

- [x] 5.1 Create `locker_test.go` - core lock service tests (Lock, LockFunc, IsLeader)
- [x] 5.2 Create `locker_instance_test.go` - lock instance tests (Unlock, Renew, IsHeld)
- [x] 5.3 Create `locker_lease_test.go` - lease renewal management tests (Start, Stop, StoppedChan)
- [x] 5.4 Create `locker_election_test.go` - leader election tests (Start, Stop, tryAcquire)
- [x] 5.5 Test coverage reaches 84.1% (exceeds 80% target)

## 6. Cache Coordination Data Model

- [x] 6.1 Add persistent `sys_cache_revision` (InnoDB) to host delivery SQL and preserve lossy `MEMORY` cache semantics of `sys_kv_cache`
- [x] 6.2 Add idempotent indexes, unique keys, and concurrent-increment constraints for `sys_cache_revision`; review existing TTL query indexes and unique keys on `sys_kv_cache`
- [x] 6.3 Run `make init` and `make dao`, updating DAO/DO/Entity artifacts only through code generation
- [x] 6.4 Review `cluster.enabled` and `cluster.Service` topology abstractions and determine single-node and cluster-mode branch integration points for `cachecoord`

## 7. Unified Cache Coordination Component

- [x] 7.1 Add `internal/service/cachecoord`, define cache domain, scope, authoritative source, consistency model, staleness window, and failure policy
- [x] 7.2 Implement core interfaces: `ConfigureDomain`, `MarkChanged`, `EnsureFresh`, `Snapshot`
- [x] 7.3 In `cluster.enabled=false`, implement in-process revision, local invalidation, and synchronous refresh without shared coordination table
- [x] 7.4 In `cluster.enabled=true`, implement atomic shared revision increments, idempotent publish, request-path freshness checks, and watcher synchronization
- [x] 7.5 Define `bizerr` codes for cache coordination failures and log observable errors with project `logger` component and propagated `ctx`

## 8. Critical Cache-Domain Integration

- [x] 8.1 Connect protected runtime parameter cache to `cachecoord` and reliably publish `runtime-config` revision after parameter write transactions
- [x] 8.2 Execute freshness checks on runtime parameter reads and return visible errors when failure window is exceeded
- [x] 8.3 Connect role, menu, user-role, and plugin permission topology write paths to `permission-access` revision publishing
- [x] 8.4 Execute permission snapshot freshness checks before protected API permission validation, fail closed when freshness cannot be confirmed after failure window
- [x] 8.5 Connect plugin install, enable, disable, uninstall, upgrade, and same-version refresh to `plugin-runtime` revision publishing

## 9. Plugin Runtime Derived Caches

- [x] 9.1 Refactor shared revision logic in `pluginruntimecache` and reuse `cachecoord` for plugin enabled snapshot refresh
- [x] 9.2 Include plugin frontend bundle cache, runtime i18n bundle cache, and dynamic route derived cache in `plugin-runtime` scoped invalidation
- [x] 9.3 Bind Wasm compilation cache keys to active release checksum or generation so same-version refresh cannot keep hitting old cache
- [x] 9.4 Adjust dynamic plugin artifact archive or cache-key strategy so other nodes can verify active release by checksum or generation
- [x] 9.5 Add old artifact and old Wasm compilation cache cleanup without deleting content referenced by current active release

## 10. Plugin Host-Cache Reliability

- [x] 10.1 Remove `sys_kv_cache` from host critical revision paths so it only stores lossy plugin/module KV cache data
- [x] 10.2 Implement `incr` with one SQL atomic update, ensuring concurrent increments from multiple nodes do not lose deltas while shared database is alive
- [x] 10.3 Return structured errors for non-integer increments and oversized namespaces, keys, or values; do not truncate or partially write data
- [x] 10.4 Change expiration cleanup to single-key lazy miss handling plus background batch cleanup, avoiding full table scans on ordinary read/write paths
- [x] 10.5 In cluster mode, limit expired-row batch cleanup pressure through primary-node coordination or idempotent batching, treat database restart as cache miss

## 11. Cache Coordination Observability and Tests

- [x] 11.1 Expose cache coordination status snapshot with at least domain, scope, local revision, shared revision, last sync time, latest error, and stale seconds
- [x] 11.2 Maintain apidoc i18n resources if new HTTP diagnostic endpoints or API documentation fields are added
- [x] 11.3 Add concurrent publishing tests for `sys_cache_revision`, verify persistent atomic increments and no loss after database restart
- [x] 11.4 Add dual-instance service-level tests covering cross-node invalidation and bounded staleness for runtime parameters, permission topology, and plugin runtime cache
- [x] 11.5 Add plugin host-cache tests for concurrent `incr`, TTL, cache miss after database restart, oversized input, and non-integer increments
- [x] 11.6 Add dynamic-plugin same-version refresh tests verifying old Wasm, frontend bundle, and i18n derived caches are invalidated after checksum or generation changes

## 12. Pluginbridge Subcomponent Skeleton

- [x] 12.1 Create `pkg/pluginbridge/{contract,codec,artifact,hostcall,hostservice,guest}` subcomponent directories with compliant package comments and file-purpose comments
- [x] 12.2 Define subcomponent dependency direction, migrate low-dependency contract/artifact/codec capabilities first, ensure no subcomponent imports root `pluginbridge`
- [x] 12.3 Move protobuf wire, WASM section low-level reading, and other pure implementation details into corresponding subcomponent `internal` packages

## 13. Pluginbridge Contract, Artifact, and Codec Migration

- [x] 13.1 Migrate `BridgeSpec`, `RouteContract`, `BridgeRequestEnvelopeV1`, `BridgeResponseEnvelopeV1`, `IdentitySnapshotV1`, `CronContract`, `ExecutionSource` to `contract` subcomponent
- [x] 13.2 Migrate bridge request/response/route/identity/HTTP snapshot encoding/decoding to `codec` subcomponent, preserve existing round trip tests
- [x] 13.3 Migrate WASM section constants, `RuntimeArtifactMetadata`, `ReadCustomSection`, `ListCustomSections` to `artifact` subcomponent, update i18n, apidoc, and runtime call paths
- [x] 13.4 Add facade and subcomponent consistency tests covering bridge envelope and WASM section representative entries

## 14. Pluginbridge Hostcall and Hostservice Migration

- [x] 14.1 Migrate host call opcodes, status codes, `HostCallResponseEnvelope`, and generic host call codec to `hostcall` subcomponent
- [x] 14.2 Migrate `HostServiceSpec`, capability derivation, host service manifest encoding/decoding, and service/method constants to `hostservice` subcomponent
- [x] 14.3 Migrate runtime, storage, network, data, cache, lock, config, notify, cron host service payload codecs to `hostservice` subcomponent, preserve field numbers and default value semantics
- [x] 14.4 Update Wasm host function, runtime, and plugindb host code to prefer importing `hostcall` / `hostservice` / `codec` and other precise subcomponents

## 15. Pluginbridge Guest SDK and Root Facade

- [x] 15.1 Migrate guest runtime, guest controller dispatcher, context response helper, BindJSON/WriteJSON, ErrorClassifier to `guest` subcomponent
- [x] 15.2 Migrate guest host service client helpers to `guest` subcomponent, maintain `Runtime()`, `Storage()`, `HTTP()`, `Data()`, `Cache()`, `Lock()`, `Config()`, `Notify()`, `Cron()` compatible entries
- [x] 15.3 Converge root `pluginbridge` package into thin facade using type alias, const alias, and wrapper functions; root directory production source limited to 1-3 files
- [x] 15.4 Update dynamic plugin demo or add compatibility tests to ensure both root package old entries and `guest` subcomponent entries compile and work

## 16. Pluginbridge Verification

- [x] 16.1 Run and fix `go test ./pkg/pluginbridge/...`
- [x] 16.2 Run and fix plugin runtime, WASM host function, and plugindb tests: `go test ./internal/service/plugin/internal/runtime/... ./internal/service/plugin/internal/wasm/... ./pkg/plugindb/...`
- [x] 16.3 Run normal Go test and `GOOS=wasip1 GOARCH=wasm go build ./...` on `apps/lina-plugins/plugin-demo-dynamic`
- [x] 16.4 Run `openspec validate`, ensure proposal, design, specs, and tasks are archive-ready

## 17. Ancillary Fixes

- [x] 17.1 Fix `/user/info` homePath to return user's first accessible menu route instead of hardcoded `/analytics`; fall back to registered safe page when no accessible business menus exist
- [x] 17.2 Fix online user list API to support pagination parameters instead of returning all records
- [x] 17.3 Align role assign user page with `ruoyi-plus-vben5` reference: style, interactions, bulk revoke button, email column, column width adjustments
- [x] 17.4 Fix menu management create/edit drawer: permission identifier required for menu and button types, input displays below menu name
- [x] 17.5 Fix `TC0063-auth-menu.ts` timeout in `beforeAll/afterAll` stages

## 18. Cache Coordination Feedback Items

- [x] 18.1 FB-1: Remove remaining `kvcache` fallback from critical cache revision paths so `role`, `config`, and `pluginruntimecache` coordinate only through `cachecoord`
- [x] 18.2 FB-2: Remove `recover` fallback from runtime parameter snapshot reads so runtime-config freshness and load failures propagate as explicit errors
- [x] 18.3 FB-3: Converge `cachecoord` multi-instance construction so in-process critical cache coordination state is managed by one coordinator
- [x] 18.4 FB-4: Remove `cachecoord` dependency on built-in domain allowlists and prior registration, allowing host modules and plugins to directly use new legal cache domains
- [x] 18.5 FB-5: Converge `kvcache` into generic KV cache foundation with backend/provider abstraction, switch TTL to `time.Duration`, add hourly MySQL backend expiration cleanup job, ensure query paths do not delete expired rows
- [x] 18.6 FB-6: Move MySQL `MEMORY` backend implementation from `kvcache` facade package into independent `kvcache/internal` package
- [x] 18.7 FB-7: Further isolate MySQL `MEMORY` backend into `kvcache/internal/mysql-memory` for future Redis provider extension

## 19. Distributed Locker Feedback Items

- [x] 19.1 FB-1: Refactor `cron.Service` to accept specific configuration objects instead of entire `config.Service` (following `election.Service` pattern)
- [x] 19.2 FB-8: Fix `TC0063-auth-menu.ts` timeout in `beforeAll/afterAll` stages

## 20. Pluginbridge Feedback Items

- [x] 20.1 FB-1: Add bugfix feedback test coverage requirement to project standards and `lina-review` skill
