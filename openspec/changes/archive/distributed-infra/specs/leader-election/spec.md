## ADDED Requirements

### Requirement: 集群模式领导选举必须使用 Redis 锁
系统 SHALL 在 `cluster.enabled=true` 且 `cluster.coordination=redis` 时使用 Redis lock store 参与领导选举。领导锁 MUST 使用固定 lock name、节点 owner token 和 TTL 租约。

#### Scenario: 首个节点成为 primary
- **WHEN** 集群模式下第一个节点启动
- **AND** Redis 中不存在领导锁
- **THEN** 节点获取领导锁
- **AND** `IsPrimary` 返回 true

#### Scenario: 第二个节点成为 follower
- **WHEN** 集群模式下已有节点持有领导锁
- **AND** 第二个节点启动
- **THEN** 第二个节点无法获取领导锁
- **AND** `IsPrimary` 返回 false

### Requirement: 领导锁续约失败必须立即降级
系统 SHALL 在 primary 节点续约领导锁失败时立即降级为 follower。降级后 MUST 停止执行主节点专属后台任务。

#### Scenario: Redis 续约失败
- **WHEN** primary 节点续约领导锁时 Redis 返回错误
- **THEN** 当前节点将本地 primary 状态置为 false
- **AND** 后续 Master-Only 任务触发时被跳过
- **AND** 系统记录带 ctx 的 warning 日志

#### Scenario: owner token 不匹配
- **WHEN** primary 节点续约时发现领导锁 owner token 不匹配
- **THEN** 当前节点降级为 follower
- **AND** 不再认为自己持有领导权

### Requirement: follower 必须按配置重试竞选
系统 SHALL 让 follower 按配置的续约/重试间隔尝试获取领导锁。原 primary 失效后，其他节点 MUST 能在租约过期后的下一轮重试中接管。

#### Scenario: primary 崩溃后接管
- **WHEN** 当前 primary 节点崩溃且不再续约
- **AND** 领导锁 TTL 到期
- **THEN** follower 在下一轮重试中可获取领导锁
- **AND** 新节点成为 primary

### Requirement: 单机模式不得启动 Redis 领导选举
系统 SHALL 在 `cluster.enabled=false` 时跳过 Redis 领导选举。当前节点 MUST 直接按 primary 语义运行。

#### Scenario: 单机模式 IsPrimary
- **WHEN** `cluster.enabled=false`
- **THEN** 不创建 Redis leader lock
- **AND** `IsPrimary` 返回 true

