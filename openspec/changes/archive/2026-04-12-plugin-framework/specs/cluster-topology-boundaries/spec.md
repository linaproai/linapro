## ADDED Requirements

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
