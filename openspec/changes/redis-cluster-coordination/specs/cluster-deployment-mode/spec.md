## ADDED Requirements

### Requirement: 集群模式必须使用 Redis coordination
系统 SHALL 在 PostgreSQL 集群模式下使用 Redis coordination 作为唯一支持的分布式协调实现。`cluster.enabled=true` MUST 与 `cluster.coordination=redis` 同时成立后才允许进入集群启动流程。

#### Scenario: PostgreSQL 集群模式启用 Redis coordination
- **WHEN** 数据库链接为 PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination=redis`
- **AND** Redis 探活成功
- **THEN** 宿主进入集群模式
- **AND** leader election、cache coordination、session hot state 和 kvcache 均使用 coordination provider

#### Scenario: PostgreSQL 集群模式未配置 coordination
- **WHEN** 数据库链接为 PostgreSQL
- **AND** `cluster.enabled=true`
- **AND** `cluster.coordination` 缺失
- **THEN** 宿主启动失败
- **AND** 不得回退到 PostgreSQL 表协调实现

### Requirement: 单机模式不得强制依赖 Redis
系统 SHALL 在 `cluster.enabled=false` 时保持单机实现精简。单机模式 MUST 不启动 Redis coordination、不注册 Redis event subscriber、不使用 Redis lock 选主。

#### Scenario: 单机模式保留进程内协调
- **WHEN** `cluster.enabled=false`
- **THEN** 当前节点直接按主节点语义运行
- **AND** cache revision 使用进程内状态
- **AND** kvcache 可继续使用 SQL table backend
- **AND** auth/session 不要求 Redis

### Requirement: 集群模式不得使用 PostgreSQL 作为跨节点协调主实现
系统 SHALL 禁止集群模式依赖 `sys_locker`、`sys_cache_revision` 或 `sys_kv_cache` 完成跨节点一致性。上述表 MAY 保留用于单机、测试、诊断或未来兜底实现。

#### Scenario: 集群模式 cachecoord 不写 sys_cache_revision
- **WHEN** `cluster.enabled=true` 且 `cluster.coordination=redis`
- **AND** 业务写路径发布缓存 revision
- **THEN** 系统使用 Redis revision store
- **AND** 不依赖 `sys_cache_revision` 递增来通知其他节点

#### Scenario: 集群模式 leader election 不写 sys_locker
- **WHEN** `cluster.enabled=true` 且 `cluster.coordination=redis`
- **AND** 节点参与 primary election
- **THEN** 系统使用 Redis lock store
- **AND** 不依赖 `sys_locker` 判断 primary

