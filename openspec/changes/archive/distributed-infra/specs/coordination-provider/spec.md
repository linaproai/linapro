## ADDED Requirements

### Requirement: 宿主必须通过 coordination provider 访问 Redis
系统 SHALL 提供内部 coordination provider 抽象，用于承载集群模式下的分布式锁、KV、revision、event 和 health 能力。业务模块 MUST 依赖 coordination provider 的窄接口，不得直接导入 Redis 客户端或手写 Redis key。

#### Scenario: 业务模块接入 coordination provider
- **WHEN** `cachecoord`、`locker`、`kvcache`、`auth`、`session` 或插件运行时需要集群协调能力
- **THEN** 它们通过 coordination provider 接口调用
- **AND** 它们不得直接创建 Redis client
- **AND** Redis 客户端依赖仅存在于 provider 实现包内

#### Scenario: 测试替换 provider
- **WHEN** 单元测试需要验证锁、KV、revision 或 event 语义
- **THEN** 测试可注入 fake coordination provider
- **AND** 测试不需要真实 Redis 服务即可覆盖业务分支

### Requirement: coordination provider 必须提供锁能力
coordination provider SHALL 提供分布式锁能力，支持获取、续约、释放、状态检查和 owner token 校验。Redis provider MUST 使用原子 `SET NX PX` 或等价机制获取锁，并使用 compare-and-delete 或等价机制释放锁。

#### Scenario: 获取 Redis 锁
- **WHEN** 节点调用 lock store 获取未被持有的锁
- **THEN** Redis provider 使用带 TTL 的原子写入创建锁
- **AND** 返回包含 lock name、owner token、node ID 和租约信息的 lock handle

#### Scenario: 防止误删他人锁
- **WHEN** 节点 A 的锁已过期且节点 B 已获取同名锁
- **AND** 节点 A 尝试释放旧 handle
- **THEN** Redis provider 不删除节点 B 的锁
- **AND** 释放操作返回未持有或等价错误

### Requirement: coordination provider 必须提供 KV 能力
coordination provider SHALL 提供短 TTL KV 能力，支持 get、set、set-if-absent、delete、incr、expire、ttl 和 compare-and-delete。KV 能力 MUST 支持 context 取消、超时和结构化错误返回。

#### Scenario: 写入短 TTL key
- **WHEN** 调用方写入 TTL 为 `60s` 的 key
- **THEN** Redis provider 使用 Redis 原生过期能力保存该 key
- **AND** key 在 TTL 到期后自然不可读

#### Scenario: 原子递增
- **WHEN** 多个节点并发对同一 coordination KV key 执行 `IncrBy(1)`
- **THEN** 每次成功调用返回唯一且单调递增的整数
- **AND** 最终值等于初始值加成功调用次数

### Requirement: coordination provider 必须提供 revision 能力
coordination provider SHALL 提供按租户、缓存域和 scope 递增 revision 的能力。Redis provider MUST 使用原子递增实现 revision bump，并支持读取当前 revision。

#### Scenario: 发布租户级 revision
- **WHEN** 调用方为 `tenantId=1001`、`domain=permission-access`、`scope=global` 发布变更
- **THEN** Redis provider 递增对应 revision key
- **AND** 返回新的 revision
- **AND** 不影响其他租户或其他 scope 的 revision

#### Scenario: 平台级联 revision
- **WHEN** 调用方发布 `tenantId=0` 且 `cascadeToTenants=true` 的平台默认变更
- **THEN** provider 记录平台级 revision
- **AND** event 载荷明确携带 `cascadeToTenants=true`

### Requirement: coordination provider 必须提供事件能力
coordination provider SHALL 提供跨节点事件发布与订阅能力。Redis provider MUST 至少支持低延迟事件通知；即使事件丢失，系统也 MUST 能通过 revision 读取兜底收敛。

#### Scenario: 发布缓存失效事件
- **WHEN** 一个节点发布 cache invalidation event
- **THEN** 其他订阅节点可收到事件并刷新本地缓存
- **AND** 事件包含 kind、domain、scope、tenantId、revision、reason、sourceNode 和 createdAt

#### Scenario: 事件丢失后 revision 兜底
- **WHEN** 节点离线期间错过 Pub/Sub 事件
- **AND** 节点恢复后执行 revision check
- **THEN** 节点发现 shared revision 高于本地 observed revision
- **AND** 节点刷新对应本地派生缓存

### Requirement: coordination provider 必须集中管理 Redis key namespace
系统 SHALL 通过集中 key builder 生成 Redis key。key MUST 包含 LinaPro namespace、组件名和必要业务维度。租户相关 key MUST 显式包含 tenant 维度。

#### Scenario: 生成租户 KV key
- **WHEN** 插件在租户 1001 下写入 namespace `counter`、key `daily`
- **THEN** Redis key 包含租户维度 `1001`
- **AND** 租户 1002 读取相同 namespace/key 不会命中该 key

#### Scenario: 禁止普通业务清空全部 key
- **WHEN** 普通业务模块需要失效某个缓存
- **THEN** 模块只能通过显式 domain/scope/tenant 操作
- **AND** 不得调用 Redis `FLUSHDB` 或扫描删除所有 LinaPro key

### Requirement: coordination health 必须可诊断
coordination provider SHALL 暴露健康快照，包括 backend、Redis 连通性、最近成功 ping 时间、最近错误、订阅状态和关键 store 状态。

#### Scenario: Redis 健康
- **WHEN** Redis ping 成功
- **THEN** health snapshot 显示 backend=`redis`
- **AND** 显示 Redis 状态为 healthy
- **AND** 记录最近成功时间

#### Scenario: Redis 故障
- **WHEN** Redis ping 失败或订阅循环异常
- **THEN** health snapshot 显示最近错误
- **AND** 系统日志通过项目 logger 记录带 ctx 的错误

