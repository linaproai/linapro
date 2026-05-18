## ADDED Requirements

### Requirement: 缓存敏感服务实例必须由拓扑感知构造边界共享
系统 SHALL 确保依赖缓存协调、共享修订号、事件订阅、分布式 KV、分布式锁、会话热状态、token 状态或本地派生缓存的服务实例由拓扑感知构造边界创建并共享。`cluster.enabled=true` 时，相关组件 MUST 使用同一 coordination-backed 后端或同一运行期服务实例，MUST 不创建仅当前节点可见的孤立默认实例。

#### Scenario: 集群模式认证中间件使用共享 token 和 session 状态
- **WHEN** `cluster.enabled=true` 且认证中间件校验请求
- **THEN** 认证中间件使用启动期注入的 auth service 和 session store
- **AND** revoked token、pre-token 和 session hot state 读取使用同一 coordination-backed 事实源
- **AND** 认证中间件不得自行构造仅当前节点可见的 auth/session 服务图

#### Scenario: 插件运行时缓存使用共享插件治理实例
- **WHEN** 插件管理、动态路由、源码插件 route registrar 或插件运行时缓存读取插件启用状态
- **THEN** 这些路径复用同一插件治理实例或同一共享 revision/event 后端
- **AND** 不得因为多个插件服务实例持有不同 enabled snapshot 而暴露已禁用或已卸载插件

#### Scenario: 运行时配置和 i18n 缓存使用共享失效路径
- **WHEN** 运行时配置或 i18n 资源在集群模式下变更
- **THEN** 使用注入的共享 cachecoord/coordination 依赖发布范围化失效
- **AND** 消费方复用同一运行期配置或 i18n 服务实例观察刷新
- **AND** 不得通过局部新实例绕过已配置的失效策略

### Requirement: 缓存一致性审查必须覆盖实例来源
系统 SHALL 在缓存一致性审查中检查缓存敏感服务的实例来源、共享边界和故障策略。审查 MUST 识别新增或修改的 `New()` 调用是否会创建独立缓存状态或独立订阅状态。

#### Scenario: 审查新增缓存服务构造
- **WHEN** 变更新增或修改认证、权限、插件、配置、i18n、session、cachecoord、kvcache、lock 或 notify 相关服务构造
- **THEN** 审查确认该构造来自启动期或 registrar 显式传入的共享依赖
- **AND** 审查标记会导致本地状态分裂的隐式构造

#### Scenario: 审查无缓存影响变更
- **WHEN** 变更确认不涉及缓存、派生状态、失效或跨实例同步
- **THEN** 审查结论明确记录无缓存一致性影响
- **AND** 不得省略该判断
