# 集群部署模式规范

## Purpose
定义宿主在单节点与 PostgreSQL 集群模式下的部署拓扑、主节点语义、集群组件启停边界和插件运行时收敛规则，确保不同部署形态具备一致且可解释的启动行为。
## Requirements
### Requirement: 集群部署模式配置

宿主 SHALL 提供基于配置文件的集群部署模式开关。系统必须支持 `cluster.enabled` 作为总开关，并支持 `cluster.election.lease`、`cluster.election.renewInterval` 作为选举子配置。未显式配置时，`cluster.enabled` 必须默认为 `false`。当前支持的运行时数据库为 PostgreSQL；当数据库链接为 PostgreSQL 方言（`database.default.link` 以 `pgsql:` 开头）时，`cluster.enabled` 按用户配置生效，可启用集群模式。不支持的数据库方言必须在启动前失败，不得通过方言钩子改写集群配置后继续启动。

#### Scenario: 默认按单节点模式启动

- **WHEN** 配置文件未声明 `cluster.enabled`
- **THEN** 宿主按单节点模式启动
- **AND** 当前节点被视为主节点

#### Scenario: 显式开启集群模式

- **WHEN** 配置文件声明 `cluster.enabled=true` 且数据库链接为 PostgreSQL 方言
- **THEN** 宿主按集群模式启动
- **AND** 选主和主节点专属行为由集群模式统一控制

#### Scenario: SQLite 方言启动前失败

- **WHEN** 配置文件 `database.default.link` 以 `sqlite:` 开头
- **AND** 配置文件声明任意 `cluster.enabled` 值
- **THEN** 宿主在方言解析阶段启动失败
- **AND** 不启动选举循环、租约续期或节点投影同步

### Requirement: 单节点模式主节点语义

当 `cluster.enabled=false` 时，宿主 SHALL 将当前节点视为主节点，并跳过仅为多节点部署服务的宿主协调逻辑。

#### Scenario: 单节点模式跳过选主基础设施
- **WHEN** 宿主以单节点模式启动
- **THEN** 系统不启动领导选举循环
- **AND** 系统不启动租约续期逻辑

#### Scenario: 单节点模式直接执行主节点专属任务
- **WHEN** 宿主以单节点模式运行且触发主节点专属调度逻辑
- **THEN** 当前节点直接执行该逻辑
- **AND** 不需要额外的主节点判定等待

### Requirement: 插件运行时拓扑收敛

宿主 SHALL 根据集群部署模式控制动态插件运行时的收敛方式。单节点模式下，动态插件管理操作必须在当前节点同步完成；集群模式下，仍然允许由主节点负责最终切换与收敛。

#### Scenario: 单节点模式同步完成动态插件切换
- **WHEN** 宿主以单节点模式执行动态插件安装、启用、禁用、卸载或升级
- **THEN** 当前节点同步完成目标插件的状态切换
- **AND** 不依赖宿主主节点轮询才能生效

#### Scenario: 集群模式保留主节点收敛
- **WHEN** 宿主以集群模式执行动态插件管理操作
- **THEN** 系统允许先记录目标状态
- **AND** 由主节点执行最终切换与收敛

### Requirement: 节点投影同步仅在集群模式启用

宿主 SHALL 仅在集群模式下维护动态插件的节点投影状态。单节点模式不得要求 `sys_plugin_node_state` 成为插件运行态生效的前置条件。

#### Scenario: 单节点模式不写入节点投影
- **WHEN** 宿主以单节点模式同步插件元数据或运行时状态
- **THEN** 系统不依赖 `sys_plugin_node_state` 记录当前插件状态
- **AND** 插件治理视图仍然能够根据宿主稳定状态推导出当前状态

#### Scenario: 集群模式写入节点投影
- **WHEN** 宿主以集群模式同步动态插件运行时状态
- **THEN** 系统写入或更新当前节点对应的插件投影记录
- **AND** 记录包含节点标识、目标状态、当前状态和 generation

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

