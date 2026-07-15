## Context

插件启用/禁用/安装/卸载/升级后，前端通过 `notifyPluginRegistryChanged` 通知监听器。`basic.vue` 中监听器会：

1. 重载运行时语言包
2. 刷新插件状态与 capability
3. 检测当前插件页 generation 是否变化
4. 调用 `refreshPluginAwareAccess()` → `refreshAccessibleState()` 全量重建菜单与动态路由

`performAccessibleStateRefresh` 在当前路由仍可访问时默认执行：

```ts
router.replace({ force: true, path, query, hash })
```

目的是让插件 iframe 的 `meta.iframeSrc`（含 version/generation）立即 rematch。副作用是宿主静态页（尤其是 `/system/plugin` 插件管理页）也会 remount。

现有能力已具备：

- `skipRouteNavigation`：只重建 access，不导航
- `resolveAccessibleRouteRefreshTarget`：判断当前路由是否仍可访问
- `getPendingPluginPageRefresh` / `detectPendingPluginPageRefresh`：generation 变更提示
- `closePluginTabs`：禁用/卸载时关闭插件相关 Tab

## Goals / Non-Goals

**Goals:**

- 插件注册表变更后菜单与动态路由仍即时同步
- 插件管理页等宿主静态页启停插件时不因 force rematch 闪烁/丢状态
- 当前路由失效时仍离开失效页
- 插件页 generation 变更仍保留提示与按需 remount 语义

**Non-Goals:**

- 不做按插件的增量路由 diff / 局部 patch
- 不改后端启停 API、权限标签或菜单投影规则
- 不改 tab 清理启发式算法本身（继续用 `closePluginTabs`）
- 不优化 `fetchUserInfo` / 菜单重拉的性能路径（可后续独立变更）

## Decisions

### D1：在 access 刷新决策层做条件静默，而不是删除 force 导航

- **选择**：`performAccessibleStateRefresh`（或紧邻的决策函数）在“当前路由仍可访问且不需要立即 rematch 插件页 meta”时跳过 force 导航。
- **替代**：仅在 `basic.vue` 插件监听器硬编码 `skipRouteNavigation: true`。
- **理由**：菜单 CRUD 等其他入口也走 `refreshAccessibleState`；把“是否需要导航”收敛到 access-refresh，避免各调用方各自猜测。插件监听器可继续调用 `refreshPluginAwareAccess()`，由底层按当前路由判定。

### D2：静默条件

同时满足则静默（`skipRouteNavigation` 等效行为）：

1. 未要求 `forceDefaultRoute`
2. 当前路由在新生成的 accessible routes 中仍可访问（`resolveAccessibleRouteRefreshTarget.accessible === true` 且无强制 `replacementPath` 需要路径切换，或 replacement 与当前 path 等价）
3. 不存在针对当前路由的 `pendingPluginPageRefresh`（generation 变更需要用户确认 remount 时，已有逻辑会提前 return；此处保持不强制导航）

不满足时：

- 不可访问 → replace 到 fallback
- 需要 `replacementPath` 纠正 hosted 路径 → 按现有逻辑导航
- 显式 `forceDefaultRoute` → 强制到默认页

说明：`pendingPluginPageRefresh` 存在时，现有代码已 `return` 不做 force replace，静默策略与之兼容。

### D3：插件管理页是首要受益场景，但不做白名单特例

- **选择**：按“当前路由是否仍可访问 + 是否需要 rematch”通用判定，不硬编码 `PluginManagement` 路由名。
- **理由**：角色管理、用户管理等宿主页在插件状态变化时同样不应闪烁；硬编码名单会漏场景。

### D4：禁用时的 Tab 清理仍在业务动作处执行

- 插件管理页 `handleStatusChange` / uninstall reload 继续调用 `closePluginTabs`。
- 静默导航不替代 Tab 清理；失效当前页由 access-refresh 的 accessible 判定兜底。

### D5：测试策略

- 以单元测试锁定决策函数：可访问宿主页 → 不 force navigate；不可访问 → fallback；replacementPath → 导航到纠正路径。
- 不新增 E2E 作为门禁（行为已可由现有插件生命周期 E2E 间接覆盖）；若后续回归再补。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 静默后 current matched 与新路由表短暂不一致 | 仅在仍可访问时静默；下次自然导航会 rematch |
| 已开 iframe Tab 的旧 `iframeSrc` 不更新 | generation 变更走 pending refresh notice；禁用走 closePluginTabs |
| 调用方传入 `skipRouteNavigation: true` 与自动静默叠加 | 保持参数语义：显式 skip 仍优先跳过导航；自动静默是默认路径上的“可访问则 skip” |
| 队列合并时 `skipRouteNavigation` 用 OR 语义可能吞掉一次必要导航 | 审查现有队列：`accessRefreshSkipRouteNavigation \|\|= skip`；自动静默应在 `performAccessibleStateRefresh` 内按**当次重建结果**判定，而不是提前把 skip 置 true 进队列 |

**关键实现约束**：自动静默必须在 `generateAccess` 之后、基于最新 `accessibleRoutes` 与 `currentRoute` 判定，不能在入队时无脑 `skipRouteNavigation=true`。

## Migration Plan

- 纯前端行为变更，无数据迁移。
- 回滚：恢复“可访问即 force replace”即可。

## Open Questions

- 无。若后续发现某些宿主页必须 force rematch 才能刷新页面内订阅的路由 meta，再按页面补充白名单。
