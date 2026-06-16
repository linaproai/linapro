## ADDED Requirements

### Requirement: 宿主共享 kvcache 服务必须显式选择拓扑后端

系统 SHALL 在 HTTP 启动期显式创建宿主共享 `kvcache.Service`。单机模式使用 SQL table provider；集群模式使用 coordination KV provider。该共享服务 MUST 被注入源码插件缓存 facade、动态插件 cache host service 和其他宿主插件缓存调用路径；这些路径不得各自调用 `kvcache.New()` 并依赖进程默认 provider。

#### Scenario: 单机模式显式使用 SQL table provider

- **WHEN** `cluster.enabled=false` 且宿主启动创建共享 `kvcache.Service`
- **THEN** 启动装配使用 `kvcache.NewSQLTableProvider()` 或等价 SQL table provider
- **AND** 不要求 coordination KV backend 存在

#### Scenario: 集群模式显式使用 coordination KV provider

- **WHEN** `cluster.enabled=true` 且 coordination 服务已初始化
- **THEN** 启动装配使用 `kvcache.NewCoordinationKVProvider(coordinationSvc)` 或等价 coordination KV provider
- **AND** cache host service 写入、删除、递增和过期操作使用 coordination KV backend

#### Scenario: 缺失集群 coordination 依赖时不退回默认后端

- **WHEN** `cluster.enabled=true` 但共享 coordination KV provider 无法创建
- **THEN** 启动或配置入口返回明确错误
- **AND** 系统不得静默退回 SQL table provider 或包级默认 provider
