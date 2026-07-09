# Design

## Deployment Model

`cluster.enabled`是宿主分布式形态的总开关。单机模式不要求 Redis，不连接 Redis，不启动 Redis event subscriber，不执行 Redis leader election；当前节点直接视为 primary，cache revision 使用进程内状态，通用 KV cache 可继续使用 SQL table backend。

集群模式必须显式配置`cluster.coordination=redis`，当前不支持 PostgreSQL 表协调作为集群主实现。HTTP 服务、业务路由、cron、插件 runtime reconciler 和缓存 watcher 启动前必须先完成 Redis 配置校验和探活；Redis 不可达时拒绝以集群模式启动。非 PostgreSQL 数据库链接在方言解析阶段失败，不进入 Redis coordination 探活或业务启动流程。

早期设计使用`sys_locker`、`sys_cache_revision`和`sys_kv_cache`降低外部依赖，这一阶段保留了单机和测试价值，但不适合作为真实集群的高频协调路径。最终设计将 PostgreSQL 保持为用户、角色、菜单、租户、插件注册表、系统配置、审计日志、任务日志和通知消息的权威业务数据源；Redis 只承载可重建协调状态、短 TTL 安全状态、revision/event、lock、session hot state 和有损 KV cache。

## Coordination Provider

宿主通过内部`coordination` provider 统一访问 Redis。provider 暴露窄接口：`LockStore`、`KVStore`、`RevisionStore`、`EventBus`和`HealthChecker`。Redis client、key builder、错误分类、连接池、超时、健康检查和日志归 provider 实现所有；`cachecoord`、`locker`、`kvcache`、`auth`、`session`、`cluster`、`cron`、插件 runtime 和插件 host service 只能通过注入的 provider 或 provider-backed 服务访问协调能力。

Redis key 使用集中 namespace，包含 LinaPro namespace、组件、租户、domain、scope、plugin、owner 和 node 等必要维度。租户相关 key 必须显式包含 tenant ID，平台级使用保留 tenant 维度；普通业务路径只能按显式 domain、scope、tenant 或插件资源失效，禁止`FLUSHDB`或无范围扫描删除。

测试使用 fake/in-memory provider 覆盖 lock、KV、revision、event 和 health 语义；真实 Redis 集成测试必须通过显式环境变量启用，使用独立 namespace，只删除精确测试 key。

## Locking And Leader Election

分布式锁在集群模式下使用 coordination lock。Redis provider 通过`SET NX PX`或等价原子写入获取带 TTL 的锁，续约和释放必须校验 owner token，释放必须使用 compare-and-delete 或等价原子操作，避免旧 handle 删除新持有者的锁。接口预留 fencing token，用于后续需要防止旧 primary 写入的场景。

leader election 属于`cluster.Service`内部实现。集群模式使用固定 leader lock name、node ID、owner token 和租约 TTL；获取锁的节点成为 primary，续约失败、owner token 不匹配或 Redis 错误时立即降级为 follower，后续 Master-Only 任务必须跳过。follower 按配置重试，原 primary 崩溃且 TTL 到期后，其他节点可在下一轮获取领导权。单机模式不启动选举循环，`IsPrimary`恒为 true。

动态插件协调器在集群模式下使用 per-plugin 分布式锁保护生命周期 SQL、迁移账本、菜单和权限治理资源同步、active release 切换、frontend bundle 切换以及 runtime revision 发布。锁名称包含插件 ID；插件 host lock service 还必须包含租户维度和逻辑锁名，除非显式使用平台共享锁能力并通过权限审计。

## Cache Coordination

`cachecoord`是宿主派生缓存一致性的统一入口，负责发布显式范围 revision、执行 freshness check、刷新本地派生缓存并暴露诊断状态。单机模式只更新进程内 revision 并立即失效或刷新本进程缓存；集群模式使用 Redis revision store 和 event bus。

集群写路径先递增 Redis revision，再发布`cache.invalidate` event；Pub/Sub 用于低延迟通知，revision 是跨节点一致性的权威协调状态。请求路径或 watcher 必须保留 revision check，以补偿 Pub/Sub 丢消息、节点离线或事件乱序。event 必须携带 kind、domain、scope、tenant ID、cascade 标记、revision、reason、source node 和创建时间，处理必须幂等。

缓存域策略不得阻塞使用：合法 cache domain 可以直接发布和读取 revision；关键域在 owner 实现中声明权威数据源、最大陈旧窗口和故障策略。权限域超出陈旧窗口时 fail-closed，运行时配置返回结构化可见错误，插件运行时采取 conservative-hide，普通有损 cache 可按 cache miss 或 host service 契约返回错误。

关键写路径不得静默成功。角色、菜单、用户角色、插件权限、运行时配置、插件 stable state 等变更在业务数据提交后必须可靠发布对应 revision；发布失败时返回结构化错误或回滚，不能让调用方误以为变更已在所有节点生效。缓存敏感服务实例必须由启动期或 registrar 显式共享，审查新增`New()`调用时必须确认不会创建本地孤立 auth、session、permission、plugin、i18n、cachecoord、kvcache、lock 或 notify 状态。

## KV Cache And Short-Lived State

`kvcache`保持后端无关 facade。单机模式使用进程内`memory`后端（基于 GoFrame `gcache`），集群模式使用 coordination KV backend。coordination KV 通过 provider 的 KV 能力实现`get`、`set`、`set-if-absent`、`delete`、`incr`、`expire`、`ttl`和 compare-and-delete；Redis backend 使用原生 TTL 和原子递增，`RequiresExpiredCleanup=false`，不注册 SQL 过期清理任务。

**单机 memory 后端决策**：单机部署只有一个宿主进程，进程内缓存足以满足同进程插件、认证和宿主模块的缓存共享。相比数据库表，进程内缓存降低数据库访问频次，符合"缓存不应成为数据库热路径"的性能目标。`memory`后端要求写入、递增和过期更新使用正 TTL，不创建永不过期缓存条目。

**sys_kv_cache 删除决策**：项目无历史兼容负担，直接从宿主 SQL 基线删除`sys_kv_cache`表、索引和注释，不新增过渡清理 SQL。删除 SQL table backend、`BackendSQLTable`常量和 KV 过期清理定时任务。

**认证权威源决策**：用户登录态权威来源为`sys_online_session`。完整认证链包含 JWT 签名/类型/revoke 快速检查和基于`sys_online_session`的会话存在性/租户归属/超时校验两个阶段。单机 JWT revoke 仅作为进程内快速拒绝缓存；服务重启后依赖 session 表拒绝已退出或强制下线 token。

所有 KV cache 都是有损缓存，不能作为权限、配置、插件 stable state、租户隔离、业务权威数据或关键 revision 的事实源。写入、删除、递增或过期操作失败不得伪装成功；插件 cache 和源码插件 cache 均通过宿主授权 namespace、插件 ID、租户维度和逻辑 key 生成内部 key，不暴露 Redis client、SQL backend、owner type 或底层连接。

认证短 TTL 状态在集群模式下统一进入 coordination KV，包括 JWT revoke、`pre_token`、租户选择 single-use marker、登录验证码、一次性 challenge 和后续同类安全状态。revoke 读取失败、`pre_token`读取或消费失败、session hot state 读取失败都必须 fail-closed，不能仅凭 JWT 签名放行。

## Session Hot State

集群模式将请求路径会话热状态放入 Redis，PostgreSQL `sys_online_session`保留为在线用户管理、数据权限过滤、登录信息展示和清理投影。登录时同时写 Redis session hot key 和 PostgreSQL 投影；Redis session key 包含 tenant ID 和 token ID，payload 包含 user ID、username、tenant ID、clientType、login time、last active time 和客户端上下文。

认证中间件先校验 JWT 和 revoke 状态，再读取 Redis session hot state；Redis 不可读时 fail-closed。请求路径在 Redis 中刷新 TTL 和 last active，并按节流窗口写回 PostgreSQL 投影，避免每个请求写主库。强制下线先校验调用方对 PostgreSQL 投影的可见性，再写 revoke、删除 Redis hot key，并删除或标记投影；仅删除投影不视为完成强退。投影清理任务仍保留，用于删除 Redis 已过期或长时间不活跃的在线会话投影。

## Runtime And Plugin Convergence

权限拓扑、运行时配置和插件运行时是分布式缓存协调的关键消费者。角色、菜单、用户角色和插件权限治理写路径发布`permission-access` revision；受保护 API 在权限 freshness 不可确认并超过窗口时 fail-closed。受保护运行时参数发布`runtime-config` revision；读取认证、会话、上传、调度等参数前确认 freshness，超出窗口时返回可见错误。

插件 install、enable、disable、uninstall、upgrade、active release 切换、dynamic artifact 更新、frontend bundle、runtime i18n 和 Wasm 派生缓存变化统一发布`plugin-runtime` revision/event。收到事件或 watcher 发现 revision 前进后，节点刷新 enabled snapshot、frontend bundle、runtime i18n、Wasm module cache、manifest 资源视图和 artifact 默认配置视图。Wasm 编译缓存必须绑定 active release 的 checksum 或 generation，同版本刷新不得继续命中旧缓存。

插件运行时 freshness 不可确认时采取 conservative-hide，避免暴露可能已禁用、卸载或权限变化的插件能力。动态插件 reconciler 使用 Redis revision/event 唤醒，并保留 safety sweep 兜底；per-plugin 锁串行化同一插件共享副作用，不同插件可独立收敛。

## Observability And Governance

系统信息或健康诊断暴露 cluster enabled、coordination backend、Redis ping 状态、最近成功时间、最近错误、node ID、primary 状态、lock 状态、revision/event subscriber 状态、kvcache backend、session hot state backend 和 cachecoord snapshot。诊断响应不得暴露 Redis 密码、完整敏感连接串或 token key；新增 API 字段时同步 apidoc i18n。

实现阶段补充了 fake provider 单元测试、真实 Redis 可选集成测试、Redis cluster smoke、cachecoord 双实例事件和 revision 测试、coordination KV、auth revoke、session hot state、role/config/plugin runtime freshness、plugin lock、system info 脱敏与 apidoc i18n 检查。CI 中 Redis 集成测试独立 job 显式提供 PostgreSQL/Redis service，nightly cluster smoke 验证真实 Redis coordination 启动、primary 收敛、登录、session hot state 和 system info 诊断。

## Cross-Domain Impacts

- `config-management`影响运行时配置缓存 freshness；当前契约由`openspec/specs/config-management/spec.md`承载，历史 owner 为`archive/system-config`，本分组只保留 Redis revision/event 和 failure strategy 的分布式设计。
- `role-management`影响`permission-access`revision、权限快照失效和租户维度缓存 key；当前契约由`openspec/specs/role-management/spec.md`承载，历史 owner 为`archive/user-management`。
- `cron-job-management`和`cron-jobs`影响 Master-Only 任务、KV cleanup 投射和 watcher 兜底；当前契约由`openspec/specs/cron-job-management/spec.md`、`openspec/specs/cron-jobs/spec.md`承载，历史 owner 为`archive/scheduled-jobs`。
- `user-auth`和`session-hot-state`影响 JWT revoke、`pre_token`、认证 fail-closed、session hot state 和租户会话隔离；当前契约由`openspec/specs/user-auth/spec.md`、`openspec/specs/session-hot-state/spec.md`承载，历史 owner 为`archive/user-auth`。
- `online-user`影响 PostgreSQL 投影、强退和数据权限过滤；当前契约由`openspec/specs/online-user/spec.md`承载，历史 owner 为`archive/system-governance`，请求热状态设计由本分组摘要说明。
- `plugin-cache-service`、`plugin-lock-service`、`plugin-runtime-loading`和`pluginbridge-subcomponent-architecture`影响插件 host cache、host lock、runtime revision、Wasm cache 和 bridge 包结构；当前契约由对应`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/plugin-framework`。
- `system-info`影响 coordination 和 cachecoord 诊断字段、脱敏和 apidoc i18n；当前契约由`openspec/specs/system-info/spec.md`承载，历史 owner 为`archive/system-governance`。
- `cache-coordination`是早期英文能力名，语义已并入`distributed-cache-coordination`；当前契约由`openspec/specs/distributed-cache-coordination/spec.md`承载，本分组不再保留旧能力名的完整规范副本。
