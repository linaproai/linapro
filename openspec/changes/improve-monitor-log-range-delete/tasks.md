## Tasks

- [x] 更新操作日志 OpenSpec 增量规范，移除页面勾选删除语义并记录范围删除入口
- [x] 更新登录日志 OpenSpec 增量规范，移除页面勾选删除语义并记录范围删除入口
- [x] 修改操作日志页面：移除复选框、选中状态和按 ID 删除入口，新增范围删除弹窗
- [x] 修改登录日志页面：移除复选框、选中状态和按 ID 删除入口，新增范围删除弹窗
- [x] 更新两个插件的 `zh-CN` 与 `en-US` 运行时语言包
- [x] 更新操作日志和登录日志 E2E，覆盖无复选框和范围删除请求
- [x] 运行验证：OpenSpec 严格校验、前端类型检查或定向测试、相关 E2E/静态验证

## Feedback

- [x] **FB-1**: 删除弹窗补充日期区域间距和删除所有日志选择

## 验证记录

- `openspec validate improve-monitor-log-range-delete --strict`：通过。
- `pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`：通过。
- `pnpm -C apps/lina-vben --filter @lina/web-antd i18n:check`：通过。
- `cd hack/tests && npx tsc --noEmit --project tsconfig.json`：通过。
- `cd hack/tests && pnpm test:validate`：通过。
- `cd hack/tests && pnpm exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-monitor-loginlog/hack/tests/e2e/TC006-loginlog-delete.ts -g "TC006c" --workers=1 --timeout=45000`：通过，`1 passed`。
- `cd hack/tests && pnpm exec playwright test /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-monitor-operlog/hack/tests/e2e/TC005-operlog-delete.ts /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-monitor-loginlog/hack/tests/e2e/TC006-loginlog-delete.ts --workers=1`：通过，`6 passed`。
- Playwright 临时截图验证：通过，截图保存至 `temp/20260605/104302-monitor-operlog-page.png`、`temp/20260605/104302-monitor-operlog-delete-dialog.png`、`temp/20260605/104302-monitor-operlog-delete-all-dialog.png`、`temp/20260605/104302-monitor-loginlog-page.png`、`temp/20260605/104302-monitor-loginlog-delete-dialog.png` 与 `temp/20260605/104302-monitor-loginlog-delete-all-dialog.png`，未发现原始 `i18n` key、布局重叠或控件渲染异常；勾选全部日志后日期控件置灰。
- `jq empty apps/lina-plugins/linapro-monitor-operlog/manifest/i18n/zh-CN/plugin.json apps/lina-plugins/linapro-monitor-operlog/manifest/i18n/en-US/plugin.json apps/lina-plugins/linapro-monitor-loginlog/manifest/i18n/zh-CN/plugin.json apps/lina-plugins/linapro-monitor-loginlog/manifest/i18n/en-US/plugin.json`：通过。
- 旧语义静态检索：通过，无 `按范围删除`、`请选择需要删除`、`deleteSelectedConfirm`、`checkboxConfig`、`hasChecked`、`getCheckboxRecords` 残留。
- `git diff --check && git -C apps/lina-plugins diff --check`：通过。

## 影响确认

- i18n：有影响，已同步两个启用 `i18n` 插件的 `zh-CN` 与 `en-US` 运行时语言包，并通过 `i18n:check`。
- 数据权限：有影响，删除动作继续复用已有 `clean` 服务路径，租户过滤和时间范围均在数据库侧生效，不新增绕过边界。
- 缓存一致性：生产无影响，不修改后端缓存、快照、失效或集群同步；E2E 插件投影刷新时清理浏览器端运行时 `i18n` 持久化缓存，避免同版本源码插件资源变更后测试页面复用旧翻译包。
- 开发工具跨平台：无影响，不修改 Makefile、脚本、CI、代码生成或 `linactl`。
- 测试策略：有影响，已更新两个插件既有 E2E 用例并定向执行通过。
- 后端 Go、SQL 和 API 契约：无新增变更，复用已有 `DELETE /clean` 范围清理契约，不修改 Go 生产代码、SQL、DAO 或路由。
