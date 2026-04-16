## ADDED Requirements

### Requirement: 领导选举启动
系统 SHALL 在服务启动时自动参与领导选举，尝试成为主节点。

#### Scenario: 首次启动成为主节点
- **WHEN** 服务启动且无其他主节点
- **THEN** 系统成功获取领导锁，当前节点成为主节点

#### Scenario: 启动时已有主节点
- **WHEN** 服务启动时已有其他节点持有领导锁且未过期
- **THEN** 当前节点成为从节点，定期尝试获取领导权

#### Scenario: 主节点故障后接管
- **WHEN** 原主节点的领导锁过期
- **THEN** 从节点在下次尝试时成功获取领导锁，成为新的主节点

### Requirement: 租约自动续期
系统 SHALL 为主节点提供租约自动续期功能，确保主节点持续持有领导权。

#### Scenario: 定期续期成功
- **WHEN** 主节点定期执行租约续期
- **THEN** 领导锁的过期时间被更新，主节点继续持有领导权

#### Scenario: 续期失败后降级
- **WHEN** 主节点续期失败（如数据库故障）
- **THEN** 当前节点降级为从节点，停止执行 Master-Only Jobs

### Requirement: 主节点状态查询
系统 SHALL 提供主节点状态查询功能，判断当前节点是否为主节点。

#### Scenario: 查询主节点状态
- **WHEN** 调用 IsLeader 方法
- **THEN** 系统返回当前节点是否为主节点的布尔值

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
