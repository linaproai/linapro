## 1. 落地路径解析工具

- [x] 1.1 新增前端纯函数：从可访问菜单/路由解析首个可导航路径，并实现落地路径优先级（redirect → homePath → 首个菜单 → 兜底）
- [x] 1.2 为解析函数补充单元测试：工作台缺失、redirect 优先、不可访问 redirect 降级、空菜单兜底

## 2. 登录与路由接入

- [x] 2.1 改造 `store/auth.ts` 登录成功/租户选择/外部登录 handoff 跳转，改为使用统一落地解析（不再无条件回退写死 `defaultHomePath`）
- [x] 2.2 改造 `router/guard.ts`：动态路由装配完成后校验目标路径可访问性，不可访问则改跳首个可访问菜单
- [x] 2.3 改造 `router/access-refresh.ts`、`layouts/basic.vue` 等首页/回退入口，统一走可访问落地解析

## 3. 后端 homePath 对齐（按需）

- [x] 3.1 复核 `resolveHomePath` 与侧栏可导航规则是否一致；若隐藏/停用菜单仍可能被选中则收紧候选条件
- [x] 3.2 补充或更新后端单测：工作台菜单缺失时 homePath 落到下一可导航菜单

## 4. 验证

- [x] 4.1 运行新增/更新的前端单元测试与相关后端单测
- [x] 4.2 执行 `openspec validate fix-login-redirect-first-menu --strict`
- [x] 4.3 记录 i18n/数据权限/缓存无影响判断；确认无新增运行期 DI
- [x] 4.4 新增 E2E：`hack/tests/e2e/auth/TC011-login-landing-first-menu.ts`（工作台 homePath 不可用时落到首个可访问菜单）

### 影响分析记录

- **i18n**：无用户可见文案变更。
- **数据权限**：无影响。
- **缓存**：无影响。
- **DI**：无新增运行期依赖。
- **跨平台**：无开发工具脚本变更。
