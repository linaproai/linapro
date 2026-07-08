# 任务

## 1. 实现

- [x] 1.1 修复角色权限树在“独立选择”和“父子联动”切换时重算并回写非原始授权项的问题。
- [x] 1.2 保持角色抽屉关闭前差异检测、权限选中数量和提交数据与真实授权选择一致。

## 2. 测试

- [x] 2.1 补充组件级测试，覆盖按钮权限在模式切换后不扩展为额外菜单 ID。
- [x] 2.2 新增`hack/tests/e2e/iam/role/TC005-role-menu-selection-mode.ts`，覆盖`Issue #82`反馈的角色编辑模式切换场景。

## 3. 验证

- [x] 3.1 运行新增或更新的前端单元测试并通过。
- [x] 3.2 运行新增`E2E`用例并通过。
- [x] 3.3 运行`openspec validate fix-role-menu-selection-mode --strict`并通过。
- [x] 3.4 完成`lina-review`审查闭环。

## 影响分析

- 修改文件：`apps/lina-vben/apps/web-antd/src/components/tree/src/helper.tsx`、`apps/lina-vben/apps/web-antd/src/components/tree/src/menu-select-table.vue`、`apps/lina-vben/apps/web-antd/src/components/tree/index.ts`、`apps/lina-vben/apps/web-antd/src/views/system/role/role-drawer.vue`、`hack/tests/pages/RolePage.ts`、`hack/tests/e2e/iam/role/TC005-role-menu-selection-mode.ts`。
- 受影响模块：宿主默认用户权限模块中的角色管理页面和菜单权限树组件。
- 根因：角色编辑抽屉使用后端返回的`checkedKeys`初始化`menuIds`，但权限树模式切换会从当前展示勾选状态重新计算并回写`menuIds`；仅按钮授权在父子联动展示中可能被重算为包含父级菜单的提交数据。
- 修复策略：权限树新增独立的真实待提交选择状态；模式切换只按当前真实选择重绘表格，不触发`menuIds`回写；角色编辑打开时根据已保存授权是否缺少祖先节点推断初始选择模式；提交时移除仅用于 UI 展示的`menuCheckStrictly`字段。
- `i18n`影响：不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、提示信息、API 文档源文本或语言包资源；已运行前端`i18n` key 检查。
- 缓存一致性影响：不新增或修改缓存、缓存失效、权限快照缓存或分布式协调路径，确认无缓存一致性影响。
- 数据权限影响：不修改后端数据读取、写操作、数据权限过滤或租户/组织边界；前端只避免提交非用户真实选择的授权关系，确认无新增数据权限接入点。
- 开发工具跨平台影响：不修改脚本、构建工具、`Makefile`、`linactl`或跨平台执行入口，确认无开发工具跨平台影响。
- 后端/API/数据库影响：不修改`apps/lina-core`、HTTP API 契约、DTO、SQL、DAO、服务层或控制器，确认无后端编译门禁、接口契约和数据库迁移影响。
- 插件影响：不修改`apps/lina-plugins`或插件宿主能力，确认无插件目录结构和插件能力 README 影响。
- 已读取规则：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`。

## 验证记录

- `openspec validate fix-role-menu-selection-mode --strict`
- `pnpm -C apps/lina-vben exec vitest run --dom apps/web-antd/src/components/tree/src/helper.test.ts`
- `pnpm -C apps/lina-vben -F @lina/web-antd run typecheck`
- `pnpm -C apps/lina-vben -F @lina/web-antd run i18n:check`
- `pnpm -C hack/tests test:validate`
- `pnpm -C hack/tests exec playwright test hack/tests/e2e/iam/role/TC005-role-menu-selection-mode.ts`
- `pnpm -C hack/tests exec playwright test hack/tests/e2e/iam/role/TC001-role-crud.ts hack/tests/e2e/iam/role/TC004-role-permission-drawer-close.ts`
