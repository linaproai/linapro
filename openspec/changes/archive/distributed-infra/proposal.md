## Why

LinaPro targets multi-node deployment (Kubernetes, multi-instance load balancing) but originally lacked distributed infrastructure for leader election, cross-node cache coordination, and plugin bridge component organization. Specific problems include:

- **Duplicate cron execution**: all nodes run Session Cleanup and Server Monitor Cleanup simultaneously, causing redundant operations and race conditions.
- **Stale cache across nodes**: critical cache domains (permissions, runtime configuration, plugin runtime state) previously reused `sys_kv_cache`, a lossy `MEMORY` engine cache table that loses data on database restart and does not guarantee atomic increment. Nodes could continue using stale authorization or configuration snapshots indefinitely. The solution introduces a persistent `sys_cache_revision` (InnoDB) table for critical cache-domain revisions, keeps `sys_kv_cache` as lossy plugin/module KV cache only, and fixes `kvcache.Incr` atomicity with single-SQL atomic update.
- **Unstructured pluginbridge package**: `pkg/pluginbridge` accumulates 40+ production files mixing ABI contracts, codecs, WASM artifact helpers, host call protocols, host service protocols, and guest SDK, making it difficult for developers to distinguish stable contracts from internal protocol details.
- **Stale Wasm compilation cache**: same-version dynamic-plugin refresh cannot reliably invalidate Wasm compilation cache on other nodes because cache keys depend only on mutable artifact paths. The solution binds Wasm compilation cache to artifact checksum so same-version refresh invalidates stale cache on all nodes.
- **Missing login homePath fallback**: `/user/info` returns a fixed `/analytics` homePath, causing 404 for users without that route permission.

## What Changes

- Add `locker` distributed lock component with database-backed lock acquisition, release, lease renewal, and leader election using MySQL `MEMORY` engine table `sys_locker`.
- Add `cron` task classification: Master-Only jobs execute only on the leader node; All-Node jobs execute on every node.
- Add unified `cachecoord` cache coordination component for free-form cache-domain revision publishing, single-node local invalidation, cluster-mode shared persistent revisions, cross-node synchronization, explicit scoped invalidation, and observability.
- Add persistent `sys_cache_revision` (InnoDB) for critical cache-domain revisions; keep `sys_kv_cache` as lossy plugin/module KV cache only.
- Fix `kvcache.Incr` to use single-SQL atomic update; refactor `kvcache` into generic KV cache foundation with backend/provider abstraction and `time.Duration` TTL.
- Coordinate critical cache domains (permission topology, runtime parameters, plugin runtime) through `cachecoord` with persistent revisions, freshness checks, and failure policies.
- Bind dynamic-plugin Wasm compilation cache to artifact checksum or generation so same-version refresh invalidates stale cache on all nodes.
- Refactor `pkg/pluginbridge` into responsibility-scoped public subcomponent packages (`contract`, `codec`, `artifact`, `hostcall`, `hostservice`, `guest`) with a thin root-package facade preserving backward compatibility.
- Fix login `homePath` to return the user's first accessible menu route instead of a hardcoded path.

## Capabilities

### New Capabilities

- `distributed-locker`: database-backed distributed lock with acquisition, release, lease renewal, and state checking.
- `leader-election`: automatic leader election on service start, lease auto-renewal, failover, and Master-Only job gating.
- `distributed-cache-coordination`: defines unified host cache coordination, revision publishing, free-form cache domains, optional policy configuration, cross-node synchronization, staleness windows, fallback behavior, and observability.
- `pluginbridge-subcomponent-architecture`: pluginbridge subcomponent package structure, dependency boundaries, compatibility facade, and verification requirements.

### Modified Capabilities

- `cron-jobs`: cron task management gains Master-Only / All-Node classification and leader-node check logic.
- `plugin-cache-service`: changes lossy plugin host-cache boundaries, TTL, `incr` atomicity, and expired-data cleanup requirements; plugin cache must not carry critical cache revisions.
- `plugin-runtime-loading`: changes cross-node invalidation requirements for dynamic-plugin runtime cache, Wasm compilation cache checksum binding, frontend bundles, and i18n derived caches.
- `config-management`: protected runtime parameter cache must use the unified coordination mechanism for cross-node visibility and bounded staleness.
- `role-management`: role, menu, user-role, and plugin permission topology changes must reliably invalidate token permission snapshots through the unified coordination mechanism.

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


---

## Redis Cluster Coordination

## Why

LinaPro 现有分布式能力主要依赖 PostgreSQL 表模拟锁、缓存修订号、分布式 KV cache 与短 TTL token 状态；该方案在单机模式下足够简洁，但在多节点部署时会把主业务库推向高频协调路径，导致性能、故障隔离与跨节点实时性都不理想。

本变更引入 Redis 作为集群模式下强制启用的统一分布式协调后端：单机模式继续保持 PostgreSQL + 进程缓存的轻量形态，集群模式则通过解耦的 coordination provider 承载分布式锁、缓存修订、跨节点事件、短 TTL 状态和分布式 KV cache。

## What Changes

- **BREAKING**: `cluster.enabled=true` 时必须显式配置 `cluster.coordination: redis`，当前仅支持 `redis`；缺失、空值或非法值均在启动期失败。
- **BREAKING**: `cluster.enabled=true` 且 Redis 配置不可用或启动期探活失败时，宿主服务拒绝以集群模式启动。
- 新增 `cluster.redis` 配置段，用于声明 Redis 地址、数据库、密码、连接超时、读超时与写超时。
- 新增内部 coordination provider 抽象，Redis 作为首个实现；业务模块不得直接散落依赖 Redis 客户端。
- 将集群模式下的 leader election、分布式锁与插件锁从 PostgreSQL 表协调迁移到 coordination lock 能力。
- 将集群模式下的 `cachecoord` 从 PostgreSQL `sys_cache_revision` 行修订迁移到 Redis revision + event 协调；单机模式继续使用进程内 revision。
- 将集群模式下的 `kvcache` 后端切换为 Redis，利用原生 TTL 与原子 `INCRBY`，避免主库承载插件/宿主短生命周期 KV 热点。
- 将 JWT revoke、`pre_token`、一次性 token 等短 TTL 认证状态统一迁移到 coordination KV 能力。
- 引入在线会话 hot state 的 Redis 存储策略，降低请求路径对 `sys_online_session` 的 touch/validate 压力；PostgreSQL 保留在线用户管理、数据权限过滤和审计/投影边界。
- 将角色权限拓扑、运行时配置、插件运行时、动态插件 reconciler、runtime i18n 与 frontend bundle 等跨节点派生缓存失效统一接入 Redis event/revision。
- 保留 PostgreSQL 作为用户、角色、菜单、租户、插件注册表、系统配置、审计日志、任务日志、通知消息等权威业务数据源。
- 保留 `sys_cache_revision`、`sys_kv_cache`、`sys_locker` 作为单机模式、测试、诊断或未来兜底实现的存储边界，但集群模式不得依赖它们完成跨节点一致性。

## Capabilities

### New Capabilities

- `cluster-coordination-config`: 定义集群协调配置、Redis 配置、启动期校验、单机/集群分支和配置错误处理。
- `coordination-provider`: 定义内部 coordination provider 抽象、Redis provider 能力集合、健康检查、命名空间、故障语义与可观测性。
- `session-hot-state`: 定义在线会话热状态在集群模式下的 Redis 存储、PostgreSQL 管理投影、强退、过期、列表查询与降级策略。

### Modified Capabilities

- `cluster-deployment-mode`: 集群模式从“可选共享 PostgreSQL 协调”收敛为“必须配置 coordination backend，当前为 Redis”。
- `cluster-topology-boundaries`: 集群拓扑、节点身份、primary 判定和拓扑注入必须依赖统一 coordination 抽象，避免业务组件各自连接 Redis。
- `distributed-locker`: 集群模式下分布式锁与 leader election 使用 coordination lock；PostgreSQL locker 仅保留单机/测试/兜底边界。
- `leader-election`: primary 选举租约、续约、释放、失联恢复与 fencing token 语义改为 Redis 原子锁模型。
- `distributed-cache-coordination`: 集群模式下缓存 revision 与跨节点失效事件使用 Redis revision + event；继续保留 tenant scope、显式 scope、幂等和最大陈旧窗口要求。
- `plugin-cache-service`: 集群模式下 host/plugin KV cache 使用 coordination KV backend；当前 coordination backend 为 Redis 时，TTL、`incr`、删除、过期和缓存 miss 由 Redis coordination KV 能力承载。
- `user-auth`: JWT revoke、`pre_token`、租户切换旧 token 撤销、登出和认证短期状态改为 coordination KV，并明确 Redis 故障时的 fail-closed 策略。
- `online-user`: 在线用户列表、强退、会话过期、数据权限过滤与会话热路径需要适配 Redis hot state + PostgreSQL 投影模型。
- `role-management`: 权限拓扑 revision、token access snapshot 失效和跨节点同步使用 coordination revision/event。
- `config-management`: 受保护运行时参数 revision、进程本地快照失效和跨节点同步使用 coordination revision/event。
- `plugin-runtime-loading`: 插件运行时缓存、动态插件 reconciler、frontend bundle、runtime i18n、Wasm 派生缓存失效使用 coordination revision/event。
- `cron-job-management`: 主节点任务、所有节点任务、会话清理、KV 过期清理与内置集群同步任务需要根据 Redis coordination 能力调整执行策略。
- `plugin-lock-service`: 插件通过 host service 使用的 lock 能力在集群模式下走 coordination lock，并继承租约、token 校验释放和故障语义。
- `system-info`: 系统运行信息与健康诊断需要暴露 coordination backend、Redis 连通性、revision/event/lock 健康状态与最近错误。

## Impact

**后端配置与启动链路**:
- 修改 `apps/lina-core/internal/service/config/` 的 cluster 配置结构、解析、默认值和校验。
- 修改 `apps/lina-core/manifest/config/config.template.yaml`，新增 `cluster.coordination` 与 `cluster.redis` 示例配置。
- 修改 `apps/lina-core/internal/cmd/` 启动编排，确保集群模式在启动服务前完成 Redis 配置校验、连接探活和 coordination provider 注入。
- SQLite 模式继续强制 `cluster.enabled=false`，不得要求 Redis。

**后端基础组件**:
- 新增或重构 `apps/lina-core/internal/service/coordination/`，承载 provider 抽象、Redis 实现、健康检查、命名空间与事件协议。
- 修改 `cluster`、`locker`、`cachecoord`、`kvcache`、`session`、`auth`、`role`、`config`、`cron`、`pluginruntimecache`、`plugin`、`i18n`、`sysinfo` 等服务。
- 插件 Wasm host service 的 cache/lock 能力需要通过宿主统一 coordination/kvcache/locker facade 进入，不直接依赖 Redis。

**外部依赖**:
- 新增 Redis 客户端依赖，要求支持 context、连接池、超时配置、`SET NX PX`、Lua 或等价 compare-and-delete、`INCR`、TTL、Pub/Sub 或 Streams。
- 开发、测试、部署文档需要说明集群模式必须准备 Redis；单机模式无需 Redis。

**数据库与 SQL**:
- 不以 Redis 替代 PostgreSQL 权威业务数据。
- 现有 `sys_cache_revision`、`sys_kv_cache`、`sys_locker` 表保留，不在集群模式跨节点一致性路径中作为主实现。
- 不新增本迭代专用 SQL，除非实现阶段发现需要保存 coordination 诊断投影或会话管理投影字段。

**API 与前端**:
- 不新增用户业务 API。
- 健康检查或系统信息响应可能新增 coordination/redis 诊断字段，需要同步 apidoc i18n JSON。
- 前端如展示系统信息或健康状态，需要使用既有 i18n 语言包，不硬编码新增文案。

**缓存一致性与安全**:
- 安全路径包括 token revoke、pre-token、权限拓扑、运行时配置和插件启用状态；Redis 不可用时必须 fail-closed 或 conservative-hide，禁止静默放行。
- 普通 lossy cache 允许在 Redis 故障时返回 cache miss，但不得伪造成功写入。

**测试**:
- 新增配置解析与启动失败单元测试。
- 新增 Redis provider 单元/集成测试，优先使用可替换 fake provider 覆盖语义，Redis 真连接测试通过环境变量显式启用。
- 更新 `cachecoord`、`kvcache`、`locker`、`auth`、`session`、`role`、`config`、`pluginruntimecache` 和插件 host service 相关测试。
- 涉及在线用户强退或系统信息可见行为时，需要补充 E2E 或现有 TC 子断言。
