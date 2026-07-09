## Why

LinaPro 需要支持多节点部署、滚动发布、插件运行时收敛和跨节点缓存一致性。早期单节点实现依赖进程内状态、普通定时任务和本地派生缓存，在 Kubernetes、多实例负载均衡或动态插件同版本刷新场景下会暴露重复任务执行、权限和配置快照长期陈旧、插件运行态不一致、Wasm 编译缓存误复用、在线会话请求路径压垮主库等问题。

分布式基础设施最初通过 PostgreSQL 表实现锁、leader election 和持久化 cache revision，解决了无外部依赖下的基础多节点治理。但随着认证短 TTL 状态、session hot state、插件 cache、权限拓扑、运行时配置、插件 runtime revision 和跨节点事件都进入高频协调路径，PostgreSQL 不应继续承担集群实时协调职责。最终设计将集群模式收敛为必须配置 Redis coordination，单机模式保持轻量本地或 SQL fallback，PostgreSQL 继续作为权威业务数据源。

## What Changes

- 定义单机和集群部署边界：`cluster.enabled=false`不依赖 Redis，当前节点直接按 primary 运行；`cluster.enabled=true`必须声明`cluster.coordination=redis`并在启动业务服务前完成 Redis 探活。
- 建立统一`coordination` provider 抽象，集中承载 Redis lock、KV、revision、event、health 和 key namespace，业务模块不得直接创建 Redis client 或手写 Redis key。
- 将 leader election、分布式锁、插件锁和动态插件 per-plugin 互斥迁移到 coordination lock，使用 TTL、owner token、compare-and-delete 以及可选 fencing token。
- 将`cachecoord`集群模式迁移到 Redis revision + event，保留请求路径或 watcher 的 revision 兜底；权限、运行时配置和插件运行时等关键域按 fail-closed、visible error 或 conservative-hide 处理。
- 将通用`kvcache`集群后端切换为 coordination KV，使用 Redis TTL 和原子递增；SQL table backend 仅用于单机或测试边界，所有 KV cache 仍是有损缓存。
- 将 JWT revoke、`pre_token`、一次性认证状态和在线会话 hot state 迁移到 coordination KV，认证和会话安全路径在 Redis 不可读时 fail-closed。
- 保留`sys_locker`和`sys_cache_revision`的历史语义：可用于单机、测试、诊断或未来兜底，不再作为集群跨节点一致性的主实现。
- 将单机模式`kvcache`后端从 SQL table 替换为进程内`memory`后端，删除`sys_kv_cache`表和相关清理任务；集群模式继续使用 Redis coordination KV。
- 明确用户登录态权威来源为`sys_online_session`，单机 JWT revoke 仅作为进程内快速拒绝缓存。
- 扩展系统信息和健康诊断，暴露 coordination backend、Redis 健康、node ID、primary 状态、revision/event/lock 最近错误和缓存协调状态。
- 将调度、配置、角色权限、认证、在线用户、插件运行时、插件 host cache/lock 和 pluginbridge 等非 owner 能力中的分布式影响迁移为交叉影响摘要，避免在本分组重复保存完整规范。

## Capabilities

### New Capabilities

- `cluster-coordination-config`：集群 coordination 配置、Redis 连接配置、启动期校验、非 PostgreSQL 方言失败和配置模板边界。
- `coordination-provider`：统一 Redis provider 抽象、锁、KV、revision、event、health、key namespace、租户隔离和 fake provider 测试边界。
- `distributed-cache-coordination`：拓扑感知 cachecoord、Redis revision/event、域策略、关键写路径失效、可观测性和缓存敏感服务实例来源审查。
- `distributed-locker`：分布式锁、coordination lock、owner token、租约续期、fencing token 预留和动态插件 per-plugin 互斥。
- `leader-election`：primary 选举、续约、降级、follower 接管、单机 primary 语义和 Master-Only 任务门禁。

### Modified Capabilities

- `cluster-deployment-mode`：集群模式从 PostgreSQL 表协调演进为 Redis coordination；单机模式不强制依赖 Redis。
- `cluster-topology-boundaries`：`cluster.Service`统一暴露 topology、node ID 和 primary 状态，coordination 由启动编排注入。
- `config-management`、`role-management`、`plugin-runtime-loading`、`user-auth`、`session-hot-state`、`online-user`、`cron-job-management`、`plugin-cache-service`、`plugin-lock-service`、`system-info`：不在本分组保留完整规范，分布式影响由主规范、各自 owner 归档和本设计中的交叉影响摘要承载。

## Impact

- 影响宿主核心的集群启动、coordination provider、leader election、cachecoord、kvcache、auth/session 状态、插件 runtime 收敛、系统信息诊断和 CI Redis smoke 验证的历史追溯。
- 不改变当前运行时代码、HTTP API、数据库、缓存实现、前端 UI、插件源码或构建行为；本归档压缩只调整 OpenSpec 历史文档承载方式。
- `i18n`历史影响仅保留为维护摘要：运行时实现阶段同步过错误码翻译和系统信息 apidoc 翻译，本次压缩不修改任何运行时语言包、`manifest/i18n`或`apidoc i18n JSON`。
- 测试历史影响保留为维护摘要：实现阶段覆盖 fake provider、真实 Redis 可选集成测试、cluster smoke、缓存一致性、认证 fail-closed、session hot state、插件锁、系统信息诊断和审查结论；本次压缩用 OpenSpec 校验和静态扫描验证。
