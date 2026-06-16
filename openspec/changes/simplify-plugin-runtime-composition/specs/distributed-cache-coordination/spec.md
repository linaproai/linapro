## ADDED Requirements

### Requirement: kvcache 拓扑选择必须由构造边界显式表达

系统 SHALL 在拓扑感知构造边界显式表达 `kvcache` 后端选择。`cluster.enabled=false` 时构造 SQL table 后端；`cluster.enabled=true` 时构造 coordination KV 后端并复用宿主统一 coordination provider。生产代码不得通过进程级默认 provider 修改或读取来表达当前拓扑。

#### Scenario: 集群模式后端选择可追溯到 coordination provider

- **WHEN** 审查集群模式 HTTP runtime 的 `kvcache` 构造路径
- **THEN** 可以从共享 `coordination.Service` 追溯到 `kvcache` provider 和共享 `kvcache.Service`
- **AND** 该路径不创建独立 Redis client 或仅当前节点可见的本地默认实例

#### Scenario: 单机模式不触碰 coordination 后端

- **WHEN** `cluster.enabled=false` 且宿主创建共享 `kvcache.Service`
- **THEN** 构造路径选择 SQL table provider
- **AND** 不初始化或要求 coordination KV backend
