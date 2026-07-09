## Requirements

### Requirement: 动态路由认证必须复用宿主权限和会话 owner

系统 SHALL 在动态插件登录路由进入 guest 前复用宿主 `role` 和 `session` owner 完成认证、权限和数据范围快照构建。插件运行时 MUST 通过构造函数显式接收 `role` 访问投影契约和共享 `session.Store`，不得在请求路径中直接查询角色治理表、创建独立 session store 或缓存 session 有效结论。

### Requirement: 动态请求热路径必须复用已解析 manifest

系统 SHALL 让动态插件请求在进入 guest 前复用已解析的 manifest，避免 `matchDynamicRoute` 和 `prepareDynamicRouteRuntime` 重复解析同一产物。

### Requirement: 动态插件请求热路径 manifest 复用

系统 SHALL 在动态路由请求上下文中复用已解析 manifest，消除同请求重复 artifact 解析。

### Requirement: WASM 编译缓存必须按插件精细失效

系统 SHALL 将已知单插件变更的 WASM 编译缓存失效收敛为按插件或 artifact 路径失效，并把编译过程移出全局缓存写锁。

### Requirement: 集群 peer 变化后必须执行有界差异对账

系统 SHALL 在集群 peer 观察到 `plugin-runtime` 修订号变化时执行有界差异对账，而不是无条件清空全部派生缓存。
