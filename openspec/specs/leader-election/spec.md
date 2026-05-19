# 领导选举规范

## Purpose
待定 - 由归档变更 distributed-locker 创建。归档后更新目的。
## Requirements
### Requirement: 领导选举启动
系统 SHALL 仅在集群部署模式开启时参与领导选举并尝试成为主节点；单节点模式下不得启动领导选举循环。

#### Scenario: 集群模式首次启动成为主节点
- **WHEN** `cluster.enabled=true` 且服务启动时无其他主节点
- **THEN** 系统成功获取领导锁
- **AND** 当前节点成为主节点

#### Scenario: 集群模式启动时已有主节点
- **WHEN** `cluster.enabled=true` 且服务启动时已有其他节点持有领导锁且未过期
- **THEN** 当前节点成为从节点
- **AND** 系统按配置的续试周期继续尝试获取领导权

#### Scenario: 主节点故障后接管
- **WHEN** `cluster.enabled=true` 且原主节点的领导锁过期
- **THEN** 从节点在下一次尝试时成功获取领导锁
- **AND** 当前节点成为新的主节点

#### Scenario: 单节点模式跳过领导选举
- **WHEN** `cluster.enabled=false` 且服务启动
- **THEN** 系统不创建领导选举循环
- **AND** 当前节点直接按主节点语义运行

### Requirement: 租约自动续期
系统 SHALL 仅在集群部署模式开启且当前节点已成为主节点时提供租约自动续期功能。

#### Scenario: 集群模式定期续期成功
- **WHEN** `cluster.enabled=true` 且主节点定期执行租约续期
- **THEN** 领导锁的过期时间被更新
- **AND** 主节点继续持有领导权

#### Scenario: 集群模式续期失败后降级
- **WHEN** `cluster.enabled=true` 且主节点续期失败
- **THEN** 当前节点降级为从节点
- **AND** 停止执行主节点专属任务

#### Scenario: 单节点模式不启动续期
- **WHEN** `cluster.enabled=false`
- **THEN** 系统不启动租约续期逻辑

### Requirement: 主节点状态查询
系统 SHALL 提供统一的主节点状态查询能力。单节点模式下查询结果必须为当前节点是主节点；集群模式下查询结果取决于领导选举状态。

#### Scenario: 单节点模式查询主节点状态
- **WHEN** `cluster.enabled=false` 且调用主节点状态查询
- **THEN** 系统返回 `true`

#### Scenario: 集群模式查询主节点状态
- **WHEN** `cluster.enabled=true` 且调用主节点状态查询
- **THEN** 系统返回当前节点是否持有领导权的布尔值

### Requirement: 定时任务分类执行
系统 SHALL 根据定时任务类型决定是否在当前节点执行。

#### Scenario: Master-Only 任务在主节点执行
- **WHEN** Master-Only 定时任务触发且当前节点是主节点
- **THEN** 任务正常执行

#### Scenario: Master-Only 任务在从节点跳过
- **WHEN** Master-Only 定时任务触发且当前节点是从节点
- **THEN** 任务跳过执行，不产生任何副作用

#### Scenario: All-Node 任务在所有节点执行
- **WHEN** All-Node 定时任务触发
- **THEN** 任务在所有节点正常执行，无论主从状态

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

