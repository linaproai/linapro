## 1. Access 刷新决策

- [x] 1.1 在 `access-refresh` 中实现“重建后按当前路由判定是否 force 导航”的决策：仍可访问且无需路径纠正时静默；不可访问时 fallback；存在 replacementPath 时导航纠正
- [x] 1.2 确保显式 `skipRouteNavigation` / `forceDefaultRoute` / `pendingPluginPageRefresh` 语义与自动静默兼容，且自动静默在 `generateAccess` 之后基于当次结果判定（不经队列 OR 误吞必要导航）

## 2. 测试与验证

- [x] 2.1 为静默/强制导航决策补充或更新单元测试
- [x] 2.2 运行相关前端单测并通过 `openspec validate plugin-registry-silent-route-refresh --strict`
