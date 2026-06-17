## Requirements

### Requirement: 宿主共享 kvcache 服务必须显式选择拓扑后端

系统 SHALL 在 HTTP 启动期显式创建宿主共享 `kvcache.Service`。单机模式使用 SQL table provider；集群模式使用 coordination KV provider。该共享服务 MUST 被注入源码插件缓存 facade、动态插件 cache host service 和其他宿主插件缓存调用路径；这些路径不得各自调用 `kvcache.New()` 并依赖进程默认 provider。
