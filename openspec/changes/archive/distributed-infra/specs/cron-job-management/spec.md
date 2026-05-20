## ADDED Requirements

### Requirement: Master-Only 定时任务必须基于 Redis primary 状态
系统 SHALL 在集群模式下通过 Redis leader election 的 primary 状态决定 Master-Only 定时任务是否执行。

#### Scenario: primary 执行 Master-Only 任务
- **WHEN** Master-Only 定时任务触发
- **AND** 当前节点持有 Redis leader lock
- **THEN** 当前节点执行任务

#### Scenario: follower 跳过 Master-Only 任务
- **WHEN** Master-Only 定时任务触发
- **AND** 当前节点未持有 Redis leader lock
- **THEN** 当前节点跳过任务
- **AND** 不产生业务副作用

### Requirement: coordination KV kvcache backend 不得注册 SQL 过期清理任务
系统 SHALL 根据 kvcache backend 判断是否注册 KV 过期清理任务。coordination KV backend MUST 不注册 SQL table cleanup job。

#### Scenario: 集群 coordination KV backend 启动
- **WHEN** 宿主以集群模式启动并使用 coordination KV kvcache backend
- **THEN** 系统不投射 `host:kvcache-cleanup-expired` 任务
- **AND** Redis TTL 负责缓存过期

### Requirement: 集群 watcher 任务必须作为 revision 兜底
系统 SHALL 保留权限拓扑、运行时配置和插件运行时同步任务作为 Redis event 的兜底机制。watcher MUST 读取 Redis revision 而不是 PostgreSQL revision 表。

#### Scenario: watcher 发现 revision 前进
- **WHEN** 节点错过 Redis event
- **AND** watcher 读取到 Redis revision 高于本地 observed revision
- **THEN** 节点刷新本地派生缓存
- **AND** 更新 observed revision
