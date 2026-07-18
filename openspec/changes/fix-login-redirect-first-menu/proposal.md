## Why

当管理员在菜单管理中禁用工作台（如 `/dashboard/analytics`）后，用户登录成功仍会落到写死的默认首页路径，而该路由不再出现在左侧可访问菜单与动态路由中，登录后直接进入 404。需要把登录成功后的落地页改为“当前用户左侧可访问菜单中的第一个可导航菜单”，避免写死地址。

## What Changes

- 明确登录成功落地路径解析优先级：显式 `redirect` 查询参数 → 用户 `homePath`（且当前会话可访问）→ 左侧菜单中第一个可导航叶子菜单 → 安全兜底路径（如个人中心）。
- 前端统一封装落地路径解析，替换 `auth`、路由守卫、访问刷新、布局首页等处对 `preferences.app.defaultHomePath`（写死 `/dashboard/analytics`）的无条件回退。
- 登录后生成动态路由时，若候选落地路径不在可访问路由中，自动改跳到第一个可访问菜单，消除 404。
- 复核并必要时收紧后端 `homePath` 解析，确保仅从启用且适合作为落地页的菜单中选取，与侧栏可见可访问顺序一致。
- 补充单元测试（及必要的 E2E）覆盖“工作台菜单禁用后登录不 404、落在首个可访问菜单”场景。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `user-auth`: 补充登录后首页/`homePath` 与前端落地路径的选择规则；当默认工作台菜单不可用时，必须落到用户可访问的第一个菜单，而不是写死工作台地址。

## Impact

- 前端：`apps/lina-vben/apps/web-antd` 的登录跳转（`store/auth.ts`）、权限守卫（`router/guard.ts`）、访问刷新（`router/access-refresh.ts`）、布局首页（`layouts/basic.vue`），以及可能新增的落地路径工具函数与测试。
- 后端：`apps/lina-core/internal/controller/user` 的 `resolveHomePath` 及其单测（若需对齐可见性/可导航规则）。
- 偏好默认值 `defaultHomePath` 可保留为配置占位，但不得在目标路径不可访问时继续作为强制落地页。
- 规范：`openspec/specs/user-auth` 增量要求。
- 影响面：`i18n` 无用户可见文案变更；数据权限/缓存无影响；主要验证为前端路由与登录流程。
