## Context

登录成功后的落地路径当前由多层逻辑共同决定：

1. 后端 `GET /user/info` 已通过 `resolveHomePath` 从**启用**菜单树中选取第一个可访问内部路由，作为 `homePath`。
2. 前端多处使用 `userInfo.homePath || preferences.app.defaultHomePath`，而 `defaultHomePath` 写死为 `/dashboard/analytics`。
3. 动态路由在 backend access 模式下由后端菜单生成；禁用工作台后 `/dashboard/*` 不会注册，但登录仍可能落到该写死路径，导致 404。
4. 首次进入权限守卫时才会 `generateAccess`；若候选路径不在可访问路由中，没有“改跳首个菜单”的二次校验。

根因不是“后端完全没有 homePath”，而是**前端对写死默认首页的无条件回退**，以及**落地前未校验路径是否仍可访问**。

## Goals / Non-Goals

**Goals:**

- 登录成功、已登录访问登录页、动态路由首次装配、访问刷新、布局“首页”等路径，统一采用可访问的落地页。
- 当工作台菜单被禁用（或任意默认路径从侧栏消失）时，用户进入左侧菜单中**第一个可导航叶子菜单**，不再出现 404。
- 保留显式 `redirect` 查询参数的优先权（用户原本想去的页面）。
- 以可单测的纯函数封装落地解析，降低多处散落 `|| defaultHomePath` 的回归风险。

**Non-Goals:**

- 不改菜单管理 CRUD、不改权限模型、不引入新的后端首页配置项。
- 不删除 `defaultHomePath` 偏好字段本身（仍可作为配置占位/兼容），只禁止其在目标不可访问时强制落地。
- 不调整多租户平台端 `/platform/*` 的特殊落地规则（仍由 `tenantStore.resolveFallbackPath` 处理租户维度约束）。
- 不在本变更中重构全部静态 `dashboard` 路由模块。

## Decisions

### D1: 前端统一 `resolvePostLoginLandingPath`

新增纯函数（建议放在 `apps/web-antd/src/router/` 或 `utils/`），输入：

- `explicitRedirect?: string`（来自 query）
- `homePath?: string`（用户信息）
- `accessibleMenus` 或可导航路径列表
- `isPathAccessible(path)`（基于当前 router / accessibleRoutes）
- `tenantFallback(path)`（可选，复用现有 `resolveFallbackPath`）

优先级：

1. 非空且可访问的 `explicitRedirect`
2. 非空且可访问的 `homePath`（经租户 fallback 后）
3. 侧栏菜单深度优先遍历得到的**第一个可导航叶子路径**
4. 最终兜底：`/profile`（与后端 `resolveHomePath` 空树兜底一致）或租户平台默认页

**不**再把 `/dashboard/analytics` 当作“菜单不可用时”的强制回退。

### D2: 在权限守卫首次装配后做落地纠偏

`authStore` 登录成功后仍可先 `router.push(候选路径)`；真正可靠的校验放在 `setupAccessGuard` 在 `generateAccess` 完成之后：

- 解析 `redirect` / `to` / `homePath`
- 若解析结果对应路由不存在或不在可访问集合中，替换为第一个可访问菜单路径

这样覆盖：密码登录、外部登录 handoff、已登录再进登录页、刷新后会话恢复等路径。

### D3: “第一个菜单”的定义

与侧栏展示一致：

- 仅目录/菜单类型，跳过按钮、外链 iframe、`hideInMenu` 项
- 深度优先、保持后端/菜单生成顺序
- 取第一个带可导航 `path` 的叶子（或有 component 的菜单项）

后端 `resolveHomePath` 继续作为 `homePath` 主来源；前端纠偏是防御层，防止 `homePath` 与当前会话动态路由短暂不一致或写死默认值污染。

### D4: 后端 `homePath` 是否收紧

复核 `isHomePathCandidate`：

- 已排除外链、iframe、非 Menu 类型；优先稳定宿主路由
- 若菜单树中仍包含 `visible=0` 的隐藏项却被选为 homePath，则收紧为要求可见（与侧栏一致）

本变更以**前端纠偏**为主；后端仅在发现与侧栏不一致时做小范围修正，避免扩大 GetInfo 语义。

### D5: 测试策略

- 单元测试：落地路径优先级、禁用工作台后取首个菜单、空菜单兜底、显式 redirect 优先
- 后端既有 `resolveHomePath` 测试补充“工作台缺失时落到下一菜单”
- E2E：若成本可控，增加“禁用工作台菜单 → 登录 → 非 404 且 URL 为侧栏首项”；否则以单元 + 手工验证记录于 tasks

## Risks / Trade-offs

- **[Risk] 登录瞬间菜单尚未就绪，首跳仍可能短暂错误** → 在守卫 `generateAccess` 后二次 resolve 并 `replace`，以装配后结果为准。
- **[Risk] 显式 redirect 指向已禁用页面** → 视为不可访问，降级到首个菜单，避免 404；不静默打开无权限页。
- **[Risk] 多租户平台用户被错误带到租户菜单** → 继续经 `resolveFallbackPath` 约束平台/租户前缀。
- **[Trade-off] 不删除 defaultHomePath** → 兼容旧偏好与测试夹具；文档与实现明确其不再是强制落地页。

## Migration Plan

- 纯行为修复，无数据迁移、无 API 破坏性变更。
- 部署后：禁用工作台菜单的环境登录应直接进入首个可访问菜单。
- 回滚：恢复前端回退 `defaultHomePath` 的旧逻辑即可。

## Open Questions

- 无阻塞问题。若产品希望空菜单时展示专用空状态页而非 `/profile`，可后续单独变更。
