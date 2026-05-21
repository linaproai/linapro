## ADDED Requirements

### Requirement: 缓存敏感服务实例必须由拓扑感知构造边界共享
系统 SHALL 确保依赖缓存协调、共享修订号、事件订阅、分布式 KV、分布式锁、会话热状态、token 状态或本地派生缓存的服务实例由拓扑感知构造边界创建并共享。`cluster.enabled=true` 时，相关组件 MUST 使用同一 coordination-backed 后端或同一运行期服务实例。

#### Scenario: 集群模式认证中间件使用共享 token 和 session 状态
- **WHEN** `cluster.enabled=true` 且认证中间件校验请求
- **THEN** 认证中间件使用启动期注入的 auth service 和 session store
- **AND** revoked token、pre-token 和 session hot state 读取使用同一 coordination-backed 事实源

### Requirement: 缓存一致性审查必须覆盖实例来源
系统 SHALL 在缓存一致性审查中检查缓存敏感服务的实例来源、共享边界和故障策略。
