## Context

当前宿主存在多类缓存：

- `runtimeParamSnapshotCache` 缓存受保护运行时参数，并通过共享 KV 修订号与 10 秒 all-node watcher 同步。
- `accessContextCache` 缓存 token 级权限快照，并通过共享 KV 修订号与短本地 revision TTL 同步。
- `pluginruntimecache.Controller` 协调插件启用快照、前端包和运行时 i18n 包，但没有覆盖 Wasm 编译缓存。
- `frontendBundleCache`、`runtimeBundleCache`、`wasmModuleCache` 都是进程内派生缓存。
- `sys_kv_cache` 同时服务插件 host-cache 与多个宿主模块修订号，但表引擎为 `MEMORY`，适合作为可丢失缓存，不适合作为关键缓存修订号来源；`Incr` 当前为读改写，无法保证并发原子性。

分布式部署下，缓存权威数据源仍应是 MySQL 中的业务表、插件发布表和运行时参数表；进程内缓存只能作为派生加速层。项目是全新项目，不需要保留旧 SQL 兼容性，可以直接调整原有 SQL 并通过重新初始化数据表落地。

## Goals / Non-Goals

**Goals:**

- 建立统一 `cachecoord` 缓存协调组件，承载自由缓存域修订号发布、可选策略配置、跨节点同步、显式作用域失效和观测。
- 在 `cluster.enabled=false` 时保持低成本本地失效；在 `cluster.enabled=true` 时使用持久共享修订号与请求路径 freshness check 保证有界陈旧。
- 让权限拓扑、运行时参数、插件运行时、Wasm 编译缓存、前端包和运行时 i18n 包具备明确一致性模型。
- 明确插件 host-cache 是可丢失缓存，不参与关键修订号协调；同时修复其 `incr` 在数据库存活期间的并发原子性问题。
- 修复共享修订号丢失、插件 host-cache `incr` 非原子、同版本动态插件刷新后旧 Wasm 继续执行等问题。
- 为关键缓存写路径定义失败处理策略，避免权限或配置失效发布失败后静默返回成功。

**Non-Goals:**

- 本次不引入 Redis、etcd、NATS 等外部协调依赖；默认实现基于现有 MySQL。
- 本次不重构所有普通查询缓存或前端浏览器缓存，只覆盖会影响权限、配置和插件运行时的关键派生缓存；插件 host-cache 只治理其可丢失缓存边界、并发递增和过期清理语义。
- 本次不改变业务模块 REST API 语义；如新增诊断 API，仅作为治理与观测能力。

## Decisions

### 1. 新增 `cachecoord` 作为统一协调入口

新增宿主内部组件 `internal/service/cachecoord`，提供以下能力：

- `MarkChanged(ctx, domain, scope, reason)`：发布某个缓存域的作用域修订号变更。
- `EnsureFresh(ctx, domain, scope, refresher)`：请求路径或 watcher 路径检查共享修订号，并在本地未消费最新版本时执行刷新/失效。
- `Snapshot(ctx)`：返回已配置策略或已触达缓存域的本地修订号、共享修订号、最后同步时间、最近错误和陈旧秒数。
- `ConfigureDomain(...)`：可选配置缓存域的权威数据源、最大可接受陈旧时间和故障策略；未配置的合法缓存域仍可直接使用默认策略参与协调。

`cachecoord` 不维护全局缓存域清单，也不把策略配置作为 domain 使用准入门槛。宿主模块或插件扩展新增缓存域时，应在自身实现代码中定义并使用稳定 domain 字符串；只有需要区别于默认陈旧窗口或故障策略时，才在代码路径中调用 `ConfigureDomain` 配置该 domain 的一致性契约。插件 manifest 不承载这类运行时缓存协调细节。

备选方案是继续在 `config`、`role`、`pluginruntimecache` 中分别维护修订号控制器。该方案短期改动少，但会继续复制一致性策略，且无法统一观测和审查。

### 2. 使用 InnoDB 持久修订号表替代 `sys_kv_cache` 修订号复用

新增 SQL，提供 `sys_cache_revision` 表，建议字段：

- `domain`：缓存域，例如 `runtime-config`、`permission-access`、`plugin-runtime`。
- `scope`：显式失效作用域，例如 `global`、`plugin:<id>`、`locale:<locale>`、`user:<id>`。
- `revision`：单调递增版本。
- `reason`、`updated_at`：观测与诊断信息。

集群模式下修订号递增必须在事务中使用行级锁或原子更新，禁止读改写导致丢失递增。单节点模式不访问该表，直接使用进程内 revision。

`sys_kv_cache` 保持 `MEMORY` 缓存表语义，用于插件和宿主模块显式 KV 缓存。数据库重启后缓存丢失是可接受行为，调用方必须按缓存未命中恢复，禁止把它用作权限、配置、插件稳定状态或其他关键修订号的可靠来源。

备选方案是继续复用 `sys_kv_cache`。该方案会把插件业务缓存和宿主协调元数据混在一起，并且 `MEMORY` 表重启清空后会丢失已发布的缓存版本，不适合作为关键一致性基础。

### 3. 关键写路径必须与修订号发布绑定

权限拓扑、运行时参数和插件稳定状态变更属于关键缓存域。写入成功后如果无法发布对应修订号，调用端不得静默成功。优先做法是让业务写入与 `sys_cache_revision` bump 处在同一个数据库事务中；无法同事务接入的路径必须在返回前发布成功，否则返回结构化业务错误并记录日志。

插件前端包和运行时 i18n 等派生缓存可以允许短暂陈旧，但必须通过 `plugin-runtime` 域修订号在请求路径或后台同步中失效。

### 4. 插件 host-cache 保留可丢失缓存语义并修复并发语义

`sys_kv_cache` 不改为 InnoDB，继续承载插件/模块显式 KV 缓存数据，且不再承担宿主缓存协调修订号职责。

- `set`：同 key last-write-wins，写入后返回当前缓存结果；数据库重启后可丢失。
- `delete`、`expire`：幂等。
- `incr`：同一数据库存活期间必须线性递增，使用单 SQL 原子更新或等价机制，不能因读改写竞态丢增量；数据库重启后的缓存值不承诺保留。
- TTL 清理：读路径可懒清理当前 key；全表清理交给主节点或可重入后台任务，避免每次读写扫描全表。

### 5. 动态插件缓存按 checksum 或 generation 失效

动态插件同版本刷新时，不再仅依赖 `pluginID/version` 路径判断 Wasm 编译缓存是否可复用。实现可选择：

- 归档路径包含 checksum，例如 `releases/<plugin>/<version>/<checksum>/<artifact>`；或
- Wasm 缓存 key 使用 `artifactPath@checksum`。

`plugin-runtime` 修订号刷新器必须覆盖集成启用快照、前端包、运行时 i18n 包和 Wasm 编译缓存。非主节点在观察到插件运行时 revision 变化后，必须丢弃旧派生缓存，再按当前发布表与 artifact checksum 重建。

### 6. Freshness 与故障降级策略按缓存域配置

建议初始策略：

- `permission-access`：最大陈旧窗口 3 秒；共享修订号不可读且本地缓存超过宽限窗口后，受保护 API 应失败关闭。
- `runtime-config`：最大陈旧窗口 10 秒；认证、上传、调度等运行时参数读取在超过宽限窗口后返回可见错误。
- `plugin-runtime`：最大陈旧窗口 5 秒；动态路由、插件菜单和资源权限读取在刷新失败时保守隐藏或拒绝不确定插件能力。
- `plugin-cache`：不进入关键修订号协调；`sys_kv_cache` 是可丢失共享缓存，重启或清理造成的未命中由插件按缓存未命中处理。

这些数值可在实现阶段做成常量或配置项；如用户希望更宽松的高可用优先策略，需要在应用前确认。

## Risks / Trade-offs

- [Risk] 请求路径 `EnsureFresh` 会增加共享修订号读取压力。→ 使用短本地 revision TTL、批量读取和后台 all-node watcher 降低热路径成本。
- [Risk] 权限失效发布失败后返回错误可能让部分管理操作失败。→ 这是为了避免跨节点授权不一致；错误必须结构化并可重试。
- [Risk] `sys_kv_cache` 保持 `MEMORY` 后数据库重启会丢失插件缓存。→ 这是缓存语义本身，禁止用于可靠业务状态；关键修订号改由 `sys_cache_revision` 持久化。
- [Risk] 动态插件同版本刷新使用 checksum 路径会增加 artifact 保留量。→ 增加基于发布状态和保留窗口的清理任务。
- [Risk] 多个现有缓存控制器迁移到统一组件存在回归风险。→ 按缓存域分阶段接入，并为每个域增加双实例测试。

## Migration Plan

1. 调整 SQL：新增 `sys_cache_revision`，保留 `sys_kv_cache` 的 `MEMORY` 缓存语义，并停止通过 `sys_kv_cache` 存储关键缓存修订号。
2. 实现 `cachecoord`，先接入运行时参数和权限拓扑两个关键域。
3. 重构 `pluginruntimecache` 或用 `cachecoord` 替代其共享 revision 逻辑，并补齐 Wasm 缓存失效。
4. 修复 `kvcache.Incr` 同一数据库存活期间的并发递增语义，以及 `Set`、`Get`、`Expire` 的 TTL 清理策略。
5. 增加健康诊断输出与测试覆盖。
6. 执行 `make init`、`make dao`，再运行后端单元测试和必要 E2E。

回滚策略：由于项目无需兼容历史数据，开发阶段回滚可直接还原 SQL 与代码并重新初始化数据库。

## Confirmed Clarifications

- 协调后端默认只使用 MySQL InnoDB 的 `sys_cache_revision`，暂不引入 Redis、etcd 或 NATS。
- `sys_kv_cache` 保持 `MEMORY` 缓存表语义，数据库重启后清空是可接受的缓存未命中，禁止用于可靠业务状态或关键修订号。
- 权限类缓存 freshness 无法确认且超过窗口时采用失败关闭策略；运行时配置按域返回明确错误。
- 缓存协调状态先通过后端 `Snapshot`、健康诊断和日志暴露，不新增管理端页面。
- `sys_locker` 的 `MEMORY` 表可靠性不并入本次缓存一致性治理，后续单独开变更处理。
- 动态插件同版本刷新允许按 checksum 或 generation 生成不可变缓存/归档标识，并配套旧 artifact 清理任务。
