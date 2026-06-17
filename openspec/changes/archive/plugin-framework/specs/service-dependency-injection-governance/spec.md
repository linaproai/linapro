## Requirements

### Requirement: Capability Provider Manager 必须由宿主显式持有

系统 SHALL 要求 Capability Provider Manager 由宿主显式持有，不得通过包级单例或全局变量管理。

### Requirement: 缓存敏感服务后端选择必须来自启动期显式装配

系统 SHALL 要求宿主启动期根据拓扑显式创建缓存敏感服务的共享实例或共享后端。生产路径 MUST NOT 依赖包级默认 provider、进程级可变默认值或构造函数隐式 fallback 来决定 `kvcache`、插件缓存、WASM cache host service 或源码插件缓存 facade 的后端类型。
