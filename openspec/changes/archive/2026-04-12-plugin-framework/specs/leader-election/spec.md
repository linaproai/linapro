## MODIFIED Requirements

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
