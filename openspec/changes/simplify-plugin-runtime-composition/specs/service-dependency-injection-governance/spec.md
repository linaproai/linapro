## ADDED Requirements

### Requirement: 缓存敏感服务后端选择必须来自启动期显式装配

系统 SHALL 要求宿主启动期根据拓扑显式创建缓存敏感服务的共享实例或共享后端。生产路径 MUST NOT 依赖包级默认 provider、进程级可变默认值或构造函数隐式 fallback 来决定 `kvcache`、插件缓存、WASM cache host service 或源码插件缓存 facade 的后端类型。

#### Scenario: HTTP 启动期创建共享 kvcache 服务

- **WHEN** 宿主 HTTP runtime 初始化共享 `kvcache.Service`
- **THEN** 启动装配根据 `cluster.enabled` 和 coordination 初始化结果显式选择 provider
- **AND** 使用该 provider 创建一个共享 `kvcache.Service`
- **AND** 后续插件 host service、源码插件缓存 facade 和 WASM cache dispatcher 复用该共享实例

#### Scenario: 生产路径不依赖默认 provider 选择后端

- **WHEN** 审查 HTTP 启动装配、插件 host service 配置或 WASM host service 配置
- **THEN** 不得发现通过进程级默认 provider 隐式选择 `kvcache` 后端的生产接线
- **AND** 测试 helper 若使用默认 provider 必须保存并恢复全局状态
