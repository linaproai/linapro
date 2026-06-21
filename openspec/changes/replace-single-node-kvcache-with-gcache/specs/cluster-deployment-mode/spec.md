## MODIFIED Requirements

### Requirement: 集群模式必须使用 Redis coordination
系统 SHALL 在 PostgreSQL 集群模式下使用 Redis coordination 作为唯一支持的分布式协调实现。`cluster.enabled=true`MUST 与`cluster.coordination=redis`同时成立后才允许进入集群启动流程。

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
- **AND** `cluster.coordination`缺失
- **THEN** 宿主启动失败
- **AND** 不得回退到 PostgreSQL 表协调实现
- **AND** 不得回退到节点本地`memory`

### Requirement: 单机模式不得强制依赖 Redis
系统 SHALL 在`cluster.enabled=false`时保持单机实现精简。单机模式 MUST 不启动 Redis coordination、不注册 Redis event subscriber、不使用 Redis lock 选主。单机模式 SHALL 使用进程内状态承载本节点协调和缓存热状态，其中 kvcache 使用`memory`进程内 backend，不依赖 Redis 或`sys_kv_cache`。

#### Scenario: 单机模式保留进程内协调
- **WHEN** `cluster.enabled=false`
- **THEN** 当前节点直接按主节点语义运行
- **AND** cache revision 使用进程内状态
- **AND** kvcache 使用`memory`进程内 backend
- **AND** auth/session 不要求 Redis
- **AND** 宿主不要求`sys_kv_cache`数据库表存在

#### Scenario: 单机模式不注册 Redis 订阅
- **WHEN** `cluster.enabled=false`
- **THEN** 宿主不创建 Redis coordination service
- **AND** 宿主不注册 Redis event subscriber
- **AND** 宿主不通过 Redis lock 判断主节点

### Requirement: 集群模式不得使用 PostgreSQL 作为跨节点协调主实现
系统 SHALL 禁止集群模式依赖`sys_locker`、`sys_cache_revision`或`sys_kv_cache`完成跨节点一致性。`sys_locker`和`sys_cache_revision`MAY 保留用于单机、测试、诊断或未来兜底实现；`sys_kv_cache`不再作为宿主交付数据库表存在。

#### Scenario: 集群模式 cachecoord 不写 sys_cache_revision
- **WHEN** `cluster.enabled=true`且`cluster.coordination=redis`
- **AND** 业务写路径发布缓存 revision
- **THEN** 系统使用 Redis revision store
- **AND** 不依赖`sys_cache_revision`递增来通知其他节点

#### Scenario: 集群模式 leader election 不写 sys_locker
- **WHEN** `cluster.enabled=true`且`cluster.coordination=redis`
- **AND** 节点参与 primary election
- **THEN** 系统使用 Redis lock store
- **AND** 不依赖`sys_locker`判断 primary

#### Scenario: 集群模式 kvcache 不写 sys_kv_cache
- **WHEN** `cluster.enabled=true`且`cluster.coordination=redis`
- **AND** 插件、认证短期状态或宿主模块写入 kvcache
- **THEN** 系统使用 Redis coordination KV backend
- **AND** 不写入`sys_kv_cache`
