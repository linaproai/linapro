## Summary

- [x] 扩展`form-ui`行数计算逻辑，派生是否存在超出`collapsedRows`的可折叠字段
- [x] 调整表单操作区渲染条件，使展开/折叠按钮仅在存在可折叠字段时显示
- [x] 确认`useVbenVxeGrid`搜索表单集成保留默认折叠能力，搜索/重置/刷新语义不变
- [x] E2E TC006：一行搜索区隐藏按钮、多行搜索区保留按钮、布局无重叠
- [x] 前端类型检查、E2E 验证、格式检查、截图视觉检查均通过
- [x] 影响分析：无 API、后端、数据库、权限、缓存、工具链、插件、i18n 影响
- [x] `lina-review`通过，覆盖前端 UI、E2E、OpenSpec、文档、架构和 i18n 审查

## 执行记录

- 前端类型检查：`pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`，通过。
- E2E 验证：`cd hack/tests && npx playwright test e2e/dashboard/TC006-search-collapse-visibility.ts`，通过。
- 截图：`temp/20260607/164638-search-collapse-single-row.png`、`164642-search-collapse-multi-row-expanded.png`、`164642-search-collapse-multi-row-collapsed.png`、`164643-search-collapse-multi-row-reexpanded.png`；未发现搜索区重叠、截断或原始 i18n key。
