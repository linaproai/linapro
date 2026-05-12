## Context

LinaPro 当前已经具备一套基础分布式能力，包括:

- `cluster.Service`: 暴露 `cluster.enabled`、节点 ID、primary 判定，并在集群模式下启动 leader election。
- `locker`: 通过 `sys_locker` 表实现分布式锁、租约续约和 leader election。
- `cachecoord`: 通过 `sys_cache_revision` 表实现缓存域 revision 递增、请求路径 freshness check 和本地派生缓存刷新。
- `kvcache`: 已经具备 backend/provider 抽象，但当前默认实现只有 SQL table backend，使用 `sys_kv_cache` 存储插件/宿主 KV cache。
- `auth` token revoke: 当前通过 `kvcache` 存储 revoked token 的短 TTL 状态。
- `session`: 当前使用 `sys_online_session` 作为在线会话存储，请求路径会校验并节流更新 `last_active_time`。
- `role`、`config`、`pluginruntimecache`、`i18n`、`plugin frontend bundle` 等派生缓存已经通过 `cachecoord` 接入共享 revision。

这些实现满足单机部署和早期多节点兜底，但在真正的多实例部署中有明显问题:

1. PostgreSQL 被用来承载高频协调写入，例如锁续约、cache revision 递增、KV cache TTL 读写、token revoke 和会话活动刷新。
2. `sys_cache_revision` 以表行作为事件源，跨节点实时性依赖请求路径检查或周期任务，无法做到低延迟主动推送。
3. `sys_kv_cache` 需要清理过期数据，且通用 KV cache 的读写热点会与权威业务数据争用数据库资源。
4. 分布式锁和 leader election 使用 SQL 表模拟租约，语义可实现但不如 Redis 原生原子操作和 TTL 自然。
5. JWT revoke、`pre_token`、一次性 token 等短 TTL 安全状态更适合 Redis TTL key；使用 SQL cache 会增加主库压力，也延长跨节点收敛链路。
6. 在线会话同时承担请求热路径校验和管理查询投影，需要拆分“热状态”和“管理视图”，否则高并发请求会不断触碰 `sys_online_session`。

本项目是新项目，无历史兼容负担。本次变更可以直接调整配置契约和集群模式实现策略，不需要保留旧的“集群模式仅依赖 PostgreSQL”的运行形态。

stakeholders:

- 框架核心维护者: 需要稳定、可复用、可测试的分布式协调抽象。
- LinaPro 使用方运维: 需要明确的单机/集群部署差异、启动失败原因、健康诊断和故障边界。
- 插件作者: 需要继续通过 host service 使用 cache/lock，不直接感知 Redis。
- 业务系统开发者: 需要知道 Redis 只承载协调状态，不替代 PostgreSQL 权威业务数据。
- AI 协作研发流程: 需要 OpenSpec 文档足够明确，便于后续 `/opsx:apply` 按任务拆分实现。

## Goals / Non-Goals

**Goals:**

- 在 `cluster.enabled=false` 时保持现有轻量体验: 不要求 Redis，不连接 Redis，继续使用 PostgreSQL 权威数据与进程内缓存/本地 revision。
- 在 `cluster.enabled=true` 时强制要求 `cluster.coordination: redis`，当前唯一支持的 coordination backend 为 Redis。
- 建立内部统一 `coordination` 抽象，业务模块只依赖锁、KV、revision、event、health 等窄接口，不直接依赖 Redis 客户端。
- 将集群模式下的高频、短生命周期、可重建协调状态从 PostgreSQL 迁移到 Redis。
- 保持 PostgreSQL 作为权威业务数据源，Redis 只存派生状态、协调状态、短 TTL token 状态和可重建 hot state。
- 明确安全关键路径的 Redis 故障策略: fail-closed 或 conservative-hide，不允许静默放行。
- 明确 lossy cache 的 Redis 故障策略: 读失败可按 cache miss 降级，写失败不得伪装成功。
- 明确 Redis key、event、revision、lock 的 namespace、tenant、scope、plugin、node 维度，避免 key 冲突和跨租户污染。
- 保留当前 `sys_cache_revision`、`sys_kv_cache`、`sys_locker` 表作为单机/测试/诊断/未来兜底实现边界，但集群模式不依赖它们完成跨节点一致性。
- 提供详细测试策略，包括 fake coordination provider、Redis 真连接集成测试、配置校验、集群行为和安全故障测试。

**Non-Goals:**

- 不引入 Redis Cluster、Sentinel、ACL、TLS、连接池高级调优的完整运维封装；首版只定义单 Redis endpoint 配置与超时。
- 不引入 etcd、NATS、Consul 或 PostgreSQL LISTEN/NOTIFY 作为同级实现；接口预留扩展，但当前只实现 Redis。
- 不把业务权威数据迁移到 Redis；用户、角色、菜单、租户、插件注册表、系统配置、审计、通知、任务、文件元数据仍以 PostgreSQL 为权威源。
- 不修改业务 REST API 语义；只允许健康检查、系统信息或 apidoc 诊断字段扩展。
- 不实现跨机房强一致、多 Redis 数据中心复制或灾备切换策略；这些属于后续部署治理。
- 不要求单机模式支持 Redis cache；单机模式保持最小依赖，避免把开发体验复杂化。
- 不实现用户可配置的 per-store backend，例如 `stores.lock`、`stores.kvCache`；内部统一使用 `cluster.coordination` 选择 provider。

## Decisions

### 1. 配置形态使用 `cluster.coordination: redis`

**选择**:

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

**规则**:

- `cluster.enabled=false`: 忽略 `cluster.coordination` 与 `cluster.redis`，不探活 Redis。
- `cluster.enabled=true`: `cluster.coordination` 必填，当前仅允许 `redis`。
- `cluster.enabled=true` 且 `coordination` 为空、缺失或非 `redis`: 启动失败。
- `coordination=redis` 时必须校验 Redis 配置并在启动服务前完成 Redis 探活。
- SQLite 模式继续强制 `cluster.enabled=false`，因此 SQLite 模式不允许实际启用 Redis coordination。

**理由**:

- 标量配置比 `coordination.backend` 更直观，当前只有一个实现类型，用户不需要理解内部 store 组合。
- 删除 per-store 配置可以避免用户配置出不一致组合，例如 lock 走 Redis、KV 走 PostgreSQL、revision 又走别的实现。
- 未来新增 `etcd` 或 `nats` 时可以继续使用同一字段:

```yaml
cluster:
  enabled: true
  coordination: etcd
  etcd:
    endpoints:
      - "127.0.0.1:2379"
```

**替代方案**:

| 方案 | 优点 | 缺点 | 结论 |
|------|------|------|------|
| `coordination.backend: redis` | 未来可在同对象下追加通用配置 | 当前层级冗余 | 不采用 |
| `stores.lock: redis` 等 per-store 配置 | 高度灵活 | 暴露内部复杂度，容易配置出不一致系统 | 不采用 |
| `redis.enabled: true` | 简单 | 混淆 Redis 是协调后端还是普通缓存 | 不采用 |

### 2. 新增内部 `coordination` 组件作为唯一 Redis 接入层

**选择**:

新增内部组件 `apps/lina-core/internal/service/coordination/`，对上暴露窄接口:

```text
coordination.Service
├─ LockStore
├─ KVStore
├─ RevisionStore
├─ EventBus
└─ HealthChecker
```

建议接口职责:

- `LockStore`: `Acquire(ctx, key, owner, ttl)`, `Renew(ctx, handle, ttl)`, `Release(ctx, handle)`, `IsHeld(ctx, handle)`。
- `KVStore`: `Get`, `Set`, `SetNX`, `Delete`, `IncrBy`, `Expire`, `TTL`, `CompareAndDelete`。
- `RevisionStore`: `Bump(ctx, domain, scope, tenantScope, reason)`, `Current(ctx, domain, scope, tenantScope)`。
- `EventBus`: `Publish(ctx, event)`, `Subscribe(ctx, consumer)`, `PollMissed(ctx, checkpoint)` 或等价可靠补偿接口。
- `HealthChecker`: `Ping(ctx)`, `Snapshot(ctx)`。

**依赖方向**:

```text
config ──┐
         ▼
 coordination(redis provider)
         ▲
         │
 cluster / locker / cachecoord / kvcache / auth / session / role / config / pluginruntimecache
```

业务模块不得直接导入 Redis 客户端包。Redis 客户端只允许存在于 `coordination` 的 Redis provider 或专属内部子包中。

**理由**:

- 统一 Redis 连接池、超时、key namespace、错误分类、健康检查和日志。
- 后续新增 etcd/NATS/PostgreSQL provider 时，不需要修改业务模块。
- 避免各模块各自实现 Redis key 规则，降低跨租户污染和清理困难。

**替代方案**:

- 各模块直接引入 Redis: 实现快，但后续难治理，违反核心宿主边界和可替换设计。
- 只给 `kvcache` 做 Redis backend: 无法覆盖锁、revision、event 和 token 状态，不能解决整体分布式协调问题。

### 3. Redis key namespace 使用稳定、可诊断、可删除的层级

**选择**:

统一 key 前缀:

```text
linapro:{app}:{env}:{component}:{scope...}
```

首版若没有显式 `app/env` 配置，可使用保守默认:

```text
linapro:default:default:{component}:{scope...}
```

建议 key 形态:

```text
linapro:default:default:lock:{name}
linapro:default:default:rev:{tenant}:{domain}:{scope}
linapro:default:default:kv:{tenant}:{ownerType}:{ownerKey}:{namespace}:{key}
linapro:default:default:auth:revoke:{tokenId}
linapro:default:default:auth:pre-token:{preTokenId}
linapro:default:default:session:{tenant}:{tokenId}
linapro:default:default:session:user-index:{tenant}:{userId}
linapro:default:default:event:cache-invalidate
linapro:default:default:event:stream:coordination
```

**规则**:

- 所有 key 的 tenant 维度必须显式出现；平台级使用 `tenant=0`。
- 业务输入必须经过 encode/escape，禁止把未清洗的用户输入直接拼进 Redis key。
- key builder 必须集中在 coordination 或对应 adapter 中，禁止业务模块手写 key。
- 删除和失效必须按显式 scope 操作，禁止普通业务路径全库 `FLUSHDB` 或扫描所有 LinaPro key。

**理由**:

- 便于运维诊断、按环境隔离、按组件清理。
- 便于未来引入多实例共享同一个 Redis 但使用不同 namespace。
- 强制 tenant/scope 维度可以延续当前多租户缓存治理要求。

### 4. Redis event 使用“Pub/Sub 快速通知 + revision 兜底”

**选择**:

跨节点失效采用两层机制:

1. 写路径执行权威数据写入后，调用 `RevisionStore.Bump` 递增 Redis revision。
2. 同一事务成功后的业务边界发布 event，通知其他节点立刻刷新本地缓存。
3. 读路径或周期 watcher 仍通过 `RevisionStore.Current` 验证 revision，补偿 Pub/Sub 丢消息或节点离线窗口。

event 载荷建议:

```json
{
  "id": "uuid-or-revision-derived-id",
  "kind": "cache.invalidate",
  "domain": "permission-access",
  "scope": "global",
  "tenantId": 1001,
  "cascadeToTenants": false,
  "revision": 42,
  "reason": "role_permissions_changed",
  "sourceNode": "node-a",
  "createdAt": "2026-05-12T10:00:00Z"
}
```

**可靠性策略**:

- Pub/Sub 用于低延迟通知，不作为唯一事实源。
- revision 是跨节点一致性的权威协调状态。
- 如果实现 Redis Streams，则 Streams 可作为 missed event 补偿；若首版不实现 Streams，则必须保留 request-path `EnsureFresh` 与 watcher 轮询 revision。
- event 处理必须幂等；重复事件不得导致错误。
- 事件消费者必须忽略本节点已确认消费的 revision 或重复 event id。

**理由**:

- Redis Pub/Sub 简单低延迟，但节点掉线期间不保留消息。
- 单独 revision 可以保证最终收敛，延续现有 `cachecoord` 的最大陈旧窗口模型。
- Pub/Sub + revision 的组合比单纯轮询 PostgreSQL 行更低延迟，也比仅 Pub/Sub 更可靠。

**替代方案**:

- 只用 Pub/Sub: 丢消息风险高。
- 只用 Redis Streams: 实现复杂度较高，首版可作为增强项。
- 继续用 `sys_cache_revision`: 主库压力和通知延迟问题仍存在。

### 5. 分布式锁使用 Redis `SET NX PX` + owner token + compare-and-delete

**选择**:

Redis lock acquisition:

```text
SET lockKey ownerToken NX PX leaseMillis
```

续约:

- 使用 Lua 或等价原子 compare-and-pexpire，仅当当前 value 等于 owner token 才续约。

释放:

- 使用 Lua 或等价原子 compare-and-delete，仅当当前 value 等于 owner token 才删除。

lock handle 必须包含:

- lock name
- owner token
- node ID
- lease duration
- acquired at / expire at diagnostic data

**leader election**:

- `cluster` primary election 使用固定 lock name，例如 `leader-election`。
- 成功获取 lock 的节点成为 primary。
- 续约失败、Redis 不可用或 owner token 不匹配时立即 demote。
- follower 按 `renewInterval` 或配置的 retry interval 尝试抢锁。

**fencing token**:

对于需要严格防止旧 leader 写入的场景，`LockStore.Acquire` 可同时递增 `lock:fence:{name}` revision，返回 fencing token。首版至少在接口中预留 fencing token，后续任务可按模块需要接入。

**理由**:

- Redis TTL 原生表达租约，不需要 SQL 过期扫描。
- owner token 防止误删其他节点新锁。
- compare-and-delete 是 Redis 分布式锁的基本安全要求。

**替代方案**:

- Redlock 多节点算法: 对当前单 Redis 配置过度复杂；首版不做。
- PostgreSQL `sys_locker`: 保留兜底，但集群模式下不再作为主实现。

### 6. `kvcache` 集群模式使用 Redis backend，单机模式继续 SQL table backend

**选择**:

`kvcache.New()` 根据运行模式选择 backend:

- `cluster.enabled=false`: SQL table backend 或现有默认 backend。
- `cluster.enabled=true`: Redis backend，由 coordination KV 能力驱动。

Redis backend 语义:

- `Get`: Redis `GET`，不存在或过期返回 miss。
- `Set`: Redis `SET`，ttl > 0 使用 PX/EX，ttl = 0 可持久到 Redis 但仍属于 lossy cache。
- `Delete`: Redis `DEL`，幂等。
- `Incr`: Redis `INCRBY`，ttl > 0 时按约定设置或刷新 TTL。
- `Expire`: Redis `PEXPIRE` 或 `PERSIST`。
- `CleanupExpired`: no-op，`RequiresExpiredCleanup=false`。

**TTL 规则**:

- `ttl < 0`: 返回业务参数错误。
- `ttl = 0`: 不设置 Redis TTL，但仍不得作为权威业务状态。
- `ttl > 0`: 使用 Redis 原生过期。

**值编码**:

需要保留当前 `Item` 语义，包括 string/int 和 optional expireAt。Redis backend 可选择:

- value 中存 JSON envelope，包含 kind/value/int/expireAt。
- int 值直接用 Redis integer string，metadata 由 key 类型推导。

推荐首版采用 envelope，减少后续扩展歧义；对 `INCRBY` 可使用单独 integer key 或 Lua 维护 envelope，具体实现阶段评估复杂度。若为了性能选择直接 integer key，必须在测试中覆盖 string/int 类型冲突。

**理由**:

- Redis 原生适合 TTL、热点读写和原子计数。
- 继续保留 `kvcache` facade，插件 host service 无需知道 backend。
- 单机模式不引入 Redis 依赖。

### 7. `cachecoord` 集群模式迁移到 Redis revision/event

**选择**:

`cachecoord.Service` 保持对上接口稳定，内部根据 topology/coordination provider 分支:

- 单机模式: 进程内 revision + 本地刷新。
- 集群模式: Redis `RevisionStore` + `EventBus`。

`MarkTenantChanged` 流程:

1. 规范化 `(tenantID, cascadeToTenants, domain, scope, reason)`。
2. 调用 Redis `INCR` 递增 revision key。
3. 写入本节点 observed revision。
4. 发布 cache invalidation event。
5. 返回 revision。

`EnsureFresh` 流程:

1. 先检查本地 observed revision 是否在本地短 TTL 内。
2. 若过期或未初始化，读取 Redis current revision。
3. 若 revision 前进，调用 refresher。
4. refresher 成功后写入 observed revision。
5. Redis 读取失败时按 domain `MaxStale` 和 `FailureStrategy` 处理。

**cache domain 策略**:

| Domain | Authority Source | Redis Coordination | Max Stale | Failure Strategy |
|--------|------------------|--------------------|-----------|------------------|
| `permission-access` | 角色、菜单、用户角色、插件权限表 | revision + event | 3s | fail-closed |
| `runtime-config` | `sys_config` 受保护运行时参数 | revision + event | 10s | visible error |
| `plugin-runtime` | 插件注册表、release、node state、artifact | revision + event | 5s | conservative-hide |
| `i18n-runtime` 如独立拆分 | manifest/i18n、动态插件 i18n | revision + event | 5s | visible error 或 stale-if-error |

**理由**:

- 保留现有 cachecoord 调用点，降低改造面。
- Redis event 降低跨节点延迟，revision 保障最终一致。
- failure strategy 明确审查边界。

### 8. 认证短 TTL 状态统一进入 coordination KV

**选择**:

以下状态在集群模式下必须存 Redis:

- JWT revoked token ID。
- `pre_token`。
- select-tenant single-use marker。
- 登录验证码、一次性认证 challenge 或后续类似短 TTL 认证状态。

**key 示例**:

```text
auth:revoke:{tokenId} -> "1", ttl = jwt remaining lifetime
auth:pre-token:{preTokenId} -> serialized payload, ttl = 60s
auth:pre-token-consumed:{preTokenId} -> "1", ttl = short replay window
```

**安全策略**:

- 写 revoke 失败: 登出、切换租户、强退等操作不得报告完全成功；返回结构化错误。
- 读 revoke 失败: token 校验 fail-closed，返回认证失败或系统繁忙错误，禁止默认放行。
- `pre_token` 读写失败: 登录两阶段流程 fail-closed。
- 本地 memory cache 可作为加速层，但不得替代 Redis 的集群可见状态。

**理由**:

- 认证撤销是安全路径，不允许短暂不一致变成权限绕过。
- Redis TTL 与 single-use token 语义天然匹配。

### 9. 在线会话采用 Redis hot state + PostgreSQL 管理投影

**选择**:

拆分在线会话职责:

- Redis hot state: 请求路径 token/session validate、last active、强退标记、用户 token index。
- PostgreSQL `sys_online_session`: 在线用户列表、数据权限过滤、登录时间、IP/browser/os、管理查询投影、审计辅助。

登录:

1. 签发 JWT。
2. 写 Redis session key，ttl = session timeout。
3. 写 `sys_online_session` 投影行。
4. 预热权限 token access snapshot。

请求校验:

1. 校验 JWT 签名和 claims。
2. 查 revoke 状态。
3. 查 Redis session hot key。
4. Redis session 存在则续 TTL 或按节流策略更新 last active hot state。
5. 按低频节流将 last active 投影回 PostgreSQL，例如每 1 分钟或登出/强退时同步。

强退:

1. 检查当前操作者数据权限和目标 session 可见性，查询 PostgreSQL 投影。
2. 写 revoke key。
3. 删除 Redis session key。
4. 删除或标记 PostgreSQL 投影。
5. 发布 session invalidation event。

过期:

- Redis 原生 TTL 使 session 自动过期。
- PostgreSQL 投影通过 primary job 或所有节点幂等任务清理超过 timeout 的行。
- 在线列表可以根据投影 `last_active_time` 过滤；如需要更实时，可在列表查询时批量校验 Redis hot key，但必须控制批量上限。

**故障策略**:

- Redis 不可用时，请求路径 session validate fail-closed。
- 在线列表 Redis 批量校验失败时，可以返回 PostgreSQL 投影并标记 diagnostic warning，或返回可见错误；实现阶段应按现有 API 语义选择。

**理由**:

- 请求路径读写进入 Redis，避免主库热点。
- PostgreSQL 保留管理查询和数据权限过滤能力。
- 该 hybrid 模型比纯 Redis 或纯 PostgreSQL 更适合 LinaPro 的管理工作台。

### 10. 插件运行时和动态插件 reconciler 使用 Redis revision/event 唤醒

**选择**:

插件相关运行时变化统一发布 `plugin-runtime` domain 事件:

- source plugin 注册/启用/禁用。
- dynamic plugin install/enable/disable/uninstall/upgrade。
- active release 变化。
- dynamic plugin frontend bundle/i18n/wasm artifact 变化。
- plugin node state 需要其他节点重新投影或重新收敛。

节点收到 event 后:

1. `plugin` root facade 刷新 enabled snapshot。
2. frontend bundle cache 按 plugin 或 global scope 失效。
3. runtime i18n bundle cache 按 dynamic-plugin/source-plugin sector 失效。
4. Wasm module cache 按 artifact checksum/generation 失效。
5. dynamic runtime reconciler 如观察到 `ScopeReconciler` revision 前进，则执行收敛。

**理由**:

- 当前 pluginruntimecache 已经通过 cachecoord 聚合多个派生缓存；只需把底层 coordination 从 PostgreSQL revision 切到 Redis。
- Redis event 可以让非 primary 节点更快感知动态插件状态变化。

### 11. Cron 与内置同步任务根据 Redis 能力调整

**选择**:

- Master-only job 继续以 `cluster.Service.IsPrimary()` 判定，只是 primary 来源改为 Redis lock。
- KV cache expired cleanup 在 Redis backend 下不再注册或执行。
- Access topology sync、runtime param sync 等集群 watcher 仍可保留，用于 revision 兜底；但事件驱动成功时不应依赖 10s 轮询才能收敛。
- Session cleanup:
  - Redis hot state 自动 TTL。
  - PostgreSQL 投影 cleanup 仍由 primary job 定期清理。

**理由**:

- 事件驱动降低延迟，周期任务保留兜底。
- 不删除已有治理任务形态，降低实现风险。

### 12. 可观测性与健康检查

**选择**:

新增或扩展系统信息/健康快照，至少包含:

- `cluster.enabled`
- `cluster.coordination`
- Redis ping 状态、延迟、最近错误、最近成功时间。
- lock store 状态: 当前 node ID、primary 状态、leader lock remaining TTL、renew error。
- revision store 状态: 已配置 domains、local/shared revision、last synced at、recent error、stale seconds。
- event bus 状态: subscriber running、last event received at、last event error、dropped/ignored duplicate count 如可得。
- kvcache backend name 与是否需要 expired cleanup。
- session hot state backend。

**i18n 影响**:

- 如果新增 API 字段进入 apidoc，必须同步维护 `manifest/i18n/<locale>/apidoc/**/*.json`。
- 如果前端展示新增健康状态文案，必须同步维护前端运行时语言包和宿主/插件 manifest i18n。
- 本提案文档本身不新增运行时 UI 文案；实现阶段必须逐项评估。

### 13. 错误模型与降级边界

**错误分类**:

| 类别 | 示例 | 策略 |
|------|------|------|
| 配置错误 | 集群启用但 coordination 缺失 | 启动失败 |
| Redis 连接不可用 | 启动 ping 失败 | 集群模式启动失败 |
| 安全状态读取失败 | revoke/session/pre-token Redis read error | fail-closed |
| 权限/配置 revision 不可读 | `permission-access` / `runtime-config` | 超出 MaxStale 后失败 |
| 插件运行时 revision 不可读 | `plugin-runtime` | conservative-hide 或返回可见错误 |
| lossy KV cache 读失败 | 插件 cache get | cache miss 或可见错误，按 host service 既有契约 |
| lossy KV cache 写失败 | 插件 cache set/incr | 返回错误，不伪装成功 |
| event publish 失败但 revision bump 成功 | cache invalidation event | 记录错误，依赖 revision 兜底；关键域可返回错误 |

所有调用端可见错误必须使用 `bizerr` 封装，并在对应模块集中定义错误码。

### 14. 数据权限与租户边界

Redis coordination 不改变数据权限的权威边界:

- 列表查询、详情、强退、在线用户管理仍以 PostgreSQL 投影和现有 `tenantcap` / `datascope` 过滤作为权限边界。
- Redis key 中必须带 tenant 维度，避免租户 A 的 session/cache/revision 污染租户 B。
- 平台级 `tenant=0` 失效可以按已有 `cascadeToTenants` 语义级联，但普通业务路径不得使用全租户清空。
- 插件 host cache 默认使用当前租户维度，平台共享写入继续要求平台权限和审计。

### 15. 测试策略

**单元测试优先使用 fake provider**:

- coordination provider interface tests。
- lock acquire/renew/release/fencing token。
- revision bump/current/event publish。
- KV TTL/incr/delete/expire。
- fail-closed / conservative-hide 错误路径。

**Redis 真连接集成测试**:

- 使用环境变量显式启用，例如 `LINA_TEST_REDIS_ADDR=127.0.0.1:6379`。
- 不依赖开发机 Redis 默认存在。
- 测试前后使用独立 namespace，禁止 `FLUSHDB`。

**模块回归测试**:

- `cluster`: 集群启用 Redis 配置校验、primary 切换、续约失败 demote。
- `locker`: Redis lock 语义、owner token 防误删。
- `cachecoord`: 双实例 revision/event 收敛、Pub/Sub 丢消息后 revision 兜底。
- `kvcache`: Redis backend string/int/TTL/incr/type conflict。
- `auth`: revoke/pre-token/select-tenant/switch-tenant fail-closed。
- `session`: Redis hot state validate、PG 投影、强退、过期清理。
- `role`: 权限拓扑变更跨节点失效。
- `config`: runtime param 变更跨节点刷新。
- `pluginruntimecache`: 动态插件 runtime、frontend、i18n、wasm 缓存失效。
- `cron`: Redis backend 下不注册 KV expired cleanup，primary job 仍只在 primary 执行。
- `sysinfo/health`: coordination 诊断字段。

**E2E 触发条件**:

- 若实现改变用户可观察的在线用户强退、系统信息页面或登录/租户切换行为，需要新增或更新对应 E2E。
- 若只改内部 provider 与单元测试覆盖充分，可不新增 E2E，但需在任务验证结论中说明无前端可观察行为。

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Redis 成为集群模式新必需依赖，增加部署复杂度 | 只在 `cluster.enabled=true` 强制；单机模式不依赖 Redis；配置模板和文档明确说明 |
| Redis 单点故障导致认证或权限路径不可用 | 安全路径 fail-closed；健康检查暴露故障；后续可扩展 Sentinel/Cluster |
| Pub/Sub 丢消息导致节点本地缓存未及时清理 | revision 作为权威协调状态，读路径/周期 watcher 兜底 |
| Redis key 设计不统一导致跨租户污染或清理困难 | 集中 key builder，所有 key 显式包含 tenant/scope，禁止业务模块手写 key |
| 同时保留 PostgreSQL 表与 Redis 实现造成维护混乱 | 文档和代码中明确单机/集群分支；集群模式不得使用 SQL 表完成跨节点一致性 |
| 会话 hot state 与 PostgreSQL 投影短暂不一致 | 明确 PG 是管理投影，Redis 是请求热状态；投影按节流同步和 cleanup 恢复 |
| Redis 写成功但 PostgreSQL 权威写失败导致派生状态提前变更 | 权威数据写成功后再 bump revision/publish event；必须避免事务前发布 |
| PostgreSQL 写成功但 Redis publish 失败导致跨节点延迟 | 关键路径返回错误或依赖 revision 兜底；实现阶段按 domain failure strategy 决定 |
| 插件 host cache 使用 Redis 后暴露更多容量风险 | 通过 TTL、namespace、value size limit 和后续指标治理；插件 cache 仍是 lossy |
| 新增 Redis 客户端依赖可能影响构建和交叉编译 | 选择纯 Go、维护活跃、支持 context 的客户端；纳入 `go test ./...` 和镜像构建 |

## Migration Plan

### 开发迁移

1. 新增配置结构和校验，但默认 `cluster.enabled=false`，确保单机开发不受影响。
2. 新增 coordination provider 接口与 fake provider 测试。
3. 新增 Redis provider，实现 health、KV、lock、revision、event 基础能力。
4. 在启动编排中创建 coordination service，并注入 cluster/locker/cachecoord/kvcache/auth/session 等模块。
5. 逐个模块切换集群模式实现，保持单机分支不变。
6. 补齐模块测试与 Redis 可选集成测试。
7. 更新配置模板、README/中文 README 或部署文档。
8. 运行 `openspec validate redis-cluster-coordination --strict`、相关 Go 单测、必要 E2E。

### 部署迁移

单机部署:

- 无需变更配置。
- `cluster.enabled=false` 时不要求 Redis。

集群部署:

- 部署 Redis。
- 配置:

```yaml
cluster:
  enabled: true
  coordination: redis
  redis:
    address: "<redis-host>:6379"
    db: 0
    password: "<secret>"
    connectTimeout: 3s
    readTimeout: 2s
    writeTimeout: 2s
```

- 启动前确认 Redis 可达。
- 滚动启动节点；只有 Redis coordination 可用的节点才能进入服务状态。

### 回滚策略

- 单机模式回滚: 设置 `cluster.enabled=false`，应用恢复为 PostgreSQL + 进程缓存模式。
- 集群模式回滚到旧实现不作为支持目标，因为本变更明确改变集群模式契约；若必须回滚，需要回退代码版本并恢复旧配置。
- Redis 短期故障恢复: 修复 Redis 后，节点通过 revision watcher 和请求路径重新收敛；会话 hot state 可能需要用户重新登录，这是可接受的安全降级。

## Open Questions

1. Redis event 首版是否实现 Streams 作为可靠补偿，还是仅使用 Pub/Sub + revision watcher?
   - 建议首版使用 Pub/Sub + revision watcher，Streams 作为后续增强，除非实现阶段发现动态插件收敛对离线补偿要求更高。
2. Redis key namespace 是否需要新增显式配置，例如 `cluster.redis.namespace`?
   - 建议首版不暴露，使用内置默认和未来可扩展字段，避免配置膨胀。
3. 在线用户列表是否需要在查询时实时批量校验 Redis session key?
   - 建议首版以 PostgreSQL 投影为主，cleanup job 保证最终清理；批量 Redis 校验作为可选优化。
4. `kvcache` Redis backend 的 int/string value 是否使用 JSON envelope 或分类型 key?
   - 建议实现阶段通过复杂度和性能评估确定；无论选择哪种，都必须保持现有 `Item` 合同和类型冲突测试。
5. Redis 不可用时普通插件 cache 读是否统一返回 miss 还是返回可见错误?
   - 建议按 host service 契约返回错误更诚实；只允许内部非关键读在明确标注时降级 miss。
