## Why

管理员在插件管理页启用/禁用插件时，宿主会全量重建菜单与动态路由，并对当前路由执行 `force replace`，导致插件管理页组件被 remount，列表筛选、滚动位置与开关交互状态丢失。正确性上需要同步菜单/路由，但在宿主静态页上不必强制导航当前页。

## What Changes

- 插件注册表变更后仍重建可访问菜单与动态路由，保持侧边栏与路由表即时同步。
- 当当前路由仍可访问且不需要立即应用新的插件页 meta（如 iframe asset / generation）时，改为静默刷新路由，跳过 `router.replace({ force: true })`。
- 当当前路由已不可访问、或当前停留在插件页且 generation 变更需要 rematch 时，保留既有强制导航或提示刷新语义。
- 禁用/卸载路径继续关闭相关插件 Tab；不改变后端 API 与权限模型。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `plugin-ui-integration`: 明确插件启停后的 access 刷新采用“静默刷新路由 + 条件导航”，宿主静态页（含插件管理页）不得因 force rematch 被整页 remount。

## Impact

- 前端：`apps/lina-vben/apps/web-antd/src/layouts/basic.vue`、`src/router/access-refresh.ts` 及相关单元测试。
- 用户体验：插件管理页启停插件时页面不闪烁；菜单/路由仍即时更新。
- 无后端、数据库、i18n 文案、权限契约变更。
