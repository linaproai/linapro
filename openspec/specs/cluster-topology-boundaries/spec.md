# 集群拓扑边界规范

## Purpose
待定 - 由归档变更 plugin-framework 创建。归档后更新目的。
## Requirements
### Requirement: Cluster Topology Ownership

宿主 SHALL 通过 `cluster.Service` 统一暴露集群模式、主节点判定和当前节点标识；所有需要这些语义的内部消费者 SHALL 从该拓扑抽象读取，而不是自行维护独立的包级集群状态。

#### Scenario: Plugin runtime reads cluster topology

- **WHEN** 插件运行时或插件宿主集成需要判断宿主是否处于集群模式、当前节点是否为主节点，或获取当前节点标识
- **THEN** 它们 SHALL 通过注入的拓扑抽象读取这些值
- **AND** `plugin` 组件 SHALL NOT 维护独立的包级集群模式或主节点全局状态

### Requirement: Election Encapsulation

宿主 SHALL 将领导选举视为 `cluster` 组件的内部实现细节，对外只暴露 `cluster.Service` 作为拓扑门面。

#### Scenario: Cluster mode starts topology services

- **WHEN** 宿主以集群模式启动
- **THEN** `cluster.Service` SHALL 在内部负责选主服务的创建、启动和停止
- **AND** 调用方 SHALL NOT 直接依赖独立的顶层 `election` service 组件

### Requirement: 集群拓扑必须由统一 coordination 注入
系统 SHALL 通过统一启动编排创建 coordination provider，并将其注入 cluster、locker、cachecoord、kvcache、auth、session、cron 和插件运行时等需要集群协调的组件。业务组件 MUST 不自行解析 Redis 配置。

#### Scenario: 启动编排注入 coordination
- **WHEN** 宿主以集群模式启动
- **THEN** 启动编排先创建 Redis coordination provider
- **AND** cluster service 使用该 provider 进行 primary election
- **AND** 其他组件通过构造参数或明确 setter 接收 provider 或 provider-backed service

#### Scenario: 禁止组件自行读取 Redis 配置
- **WHEN** `role` 或 `pluginruntimecache` 需要发布跨节点 revision
- **THEN** 它们通过 `cachecoord` 或 coordination-backed controller 完成
- **AND** 不读取 `cluster.redis.address`
- **AND** 不创建 Redis client

### Requirement: 节点身份必须贯穿 coordination 事件
系统 SHALL 在 coordination lock、revision event、插件运行时事件和健康诊断中携带稳定 node ID。node ID MUST 由 cluster/topology 层统一提供。

#### Scenario: 发布事件包含 sourceNode
- **WHEN** 节点发布 cache invalidation event
- **THEN** event payload 包含当前 node ID
- **AND** 接收节点可忽略或诊断来自自身的重复事件

#### Scenario: 健康诊断包含 node ID
- **WHEN** 查询系统信息或健康状态
- **THEN** 响应包含当前 node ID
- **AND** 响应包含当前节点是否为 primary

### Requirement: Primary 判定必须与 Redis lock 状态一致
系统 SHALL 在集群模式下以 Redis leader lock 的持有状态作为 `IsPrimary` 的权威来源。续约失败或锁丢失后，`IsPrimary` MUST 立即返回 false。

#### Scenario: 续约失败后 primary 状态变更
- **WHEN** 当前 primary 节点无法续约 leader lock
- **THEN** cluster service 将本节点降级为 follower
- **AND** `IsPrimary` 返回 false
- **AND** 主节点专属任务停止执行

