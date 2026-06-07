## 1. 表单搜索区行为实现

- [x] 1.1 扩展`form-ui`现有行数计算逻辑，派生当前布局下是否存在超出`collapsedRows`的可折叠字段
- [x] 1.2 调整表单操作区渲染条件，使展开/折叠按钮仅在`showCollapseButton=true`且存在可折叠字段时显示
- [x] 1.3 确认`useVbenVxeGrid`搜索表单集成仍保留默认折叠能力，搜索、重置、搜索面板隐藏和表格刷新语义不变

## 2. E2E — TC006 搜索区展开折叠按钮可见性

- [x] 2.1 创建或更新宿主布局页面对象，封装搜索区按钮可见性和布局断言
- [x] 2.2 创建`hack/tests/e2e/dashboard/TC006-search-collapse-visibility.ts`
- [x] 2.3 实现`TC-6a`：桌面视口下一行搜索区不显示展开/折叠按钮，搜索和重置按钮仍在同一行且无重叠
- [x] 2.4 实现`TC-6b`：多行搜索区继续显示展开/折叠按钮，并可展开额外搜索条件
- [x] 2.5 在关键页面状态捕获截图到`temp/<YYYYMMDD>/`并完成视觉检查记录

## 3. 验证与任务记录

- [x] 3.1 运行前端类型检查，确认表单和表格封装类型通过
- [x] 3.2 运行`cd hack/tests && npx playwright test e2e/dashboard/TC006-search-collapse-visibility.ts`
- [x] 3.3 运行`openspec validate hide-single-row-search-collapse --strict`
- [x] 3.4 在任务记录中明确记录无 API、后端 Go、数据库、数据权限、缓存一致性、开发工具跨平台、插件目录和`i18n`资源影响
- [x] 3.5 完成实现后调用`lina-review`，覆盖前端 UI、E2E 质量和 OpenSpec 规范合规审查

## 执行记录

- 前端类型检查：`pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`，通过。
- E2E 类型检查：`cd hack/tests && npx tsc --noEmit --project tsconfig.json`，通过。
- E2E 验证：`cd hack/tests && npx playwright test e2e/dashboard/TC006-search-collapse-visibility.ts`，通过。
- 格式检查：`git diff --check`，通过。
- 截图检查：已查看`temp/20260607/164638-search-collapse-single-row.png`、`temp/20260607/164642-search-collapse-multi-row-expanded.png`、`temp/20260607/164642-search-collapse-multi-row-collapsed.png`、`temp/20260607/164643-search-collapse-multi-row-reexpanded.png`；未发现搜索区重叠、截断或原始`i18n` key。
- 影响分析：无 API、后端 Go、数据库、数据权限、缓存一致性、开发工具跨平台、插件目录和`i18n`资源影响；本次变更仅调整宿主前端表单展示适配和宿主 E2E。
- `lina-review`：已覆盖前端 UI、E2E 质量、OpenSpec、文档、架构和`i18n`影响审查；未发现阻塞问题。
