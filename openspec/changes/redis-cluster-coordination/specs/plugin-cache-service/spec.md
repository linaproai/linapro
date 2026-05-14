## ADDED Requirements

### Requirement: 集群模式插件缓存必须使用 coordination KV backend
系统 SHALL 在 `cluster.enabled=true` 时使用 coordination KV backend 承载 host/plugin KV cache。coordination KV backend MUST 通过统一 coordination provider 创建，不得由插件缓存服务自行创建 Redis client。

#### Scenario: 集群模式写入插件缓存
- **WHEN** 插件在集群模式下调用 cache set
- **THEN** 系统将值写入 coordination KV backend
- **AND** key 包含租户、owner type、owner key、namespace 和 logical key
- **AND** 不写入 `sys_kv_cache` 作为集群 KV cache 主实现

#### Scenario: 单机模式继续使用 SQL table backend
- **WHEN** `cluster.enabled=false`
- **THEN** 插件缓存可继续使用 SQL table backend
- **AND** 不要求 coordination KV backend 存在

### Requirement: coordination KV 插件缓存必须使用后端原生 TTL
coordination KV backend SHALL 使用后端原生 TTL 处理缓存过期。coordination KV backend MUST 返回 `RequiresExpiredCleanup=false`，并且不得注册 SQL 过期清理任务。当前集群 coordination backend 为 Redis 时，该 TTL 由 Redis 原生过期能力负责。

#### Scenario: coordination KV TTL 到期
- **WHEN** 插件写入 TTL 为 `5s` 的缓存值
- **AND** 5 秒后读取该 key
- **THEN** 返回缓存未命中
- **AND** 不需要后台 SQL cleanup 才能过期

#### Scenario: coordination KV backend 不注册 KV cleanup job
- **WHEN** 宿主以集群模式和 coordination KV backend 启动
- **THEN** 内置 `host:kvcache-cleanup-expired` 不因 coordination KV backend 注册
- **AND** Redis 过期由 Redis 自身负责

### Requirement: coordination KV 插件缓存递增必须原子
coordination KV backend SHALL 使用 coordination KV 原子递增能力实现 `incr`。并发成功递增不得丢失。

#### Scenario: 多节点并发 coordination KV incr
- **WHEN** 多个节点并发对同一插件缓存 key 执行 `incr(delta=1)`
- **THEN** 每次成功调用返回唯一整数
- **AND** 最终值等于成功调用次数

#### Scenario: 递增字符串值
- **WHEN** 插件对已有字符串缓存值执行 `incr`
- **THEN** 系统返回结构化类型错误
- **AND** 原始字符串值不被修改

### Requirement: 插件缓存仍为有损缓存
无论 backend 是 Redis 还是 SQL table，插件缓存 SHALL 仍被视为有损缓存。系统 MUST 不依赖插件缓存作为权限、配置、插件稳定状态或业务权威数据源。

#### Scenario: Redis key 被清理
- **WHEN** Redis 中某插件缓存 key 被 TTL 或运维清理移除
- **THEN** 插件读取该 key 返回缓存未命中
- **AND** 系统不得因此丢失权威业务状态

### Requirement: coordination KV 插件缓存故障不得伪装写入成功
当 coordination KV backend 写入、删除、递增或过期操作失败时，系统 SHALL 返回结构化错误。系统 MUST 不得在 coordination KV 写失败时向插件报告成功。

#### Scenario: coordination KV 写失败
- **WHEN** 插件调用 cache set
- **AND** coordination KV 返回连接错误
- **THEN** host service 返回错误响应
- **AND** 插件可根据错误决定重试或降级
