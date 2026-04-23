## Why

当前管理工作台已经完成菜单骨架重构,前端稳定能力边界已经收敛到 `dashboard / iam / setting / scheduler / extension / about` 以及对应的插件能力目录,但 `hack/tests/e2e/` 仍然保留大量以旧工作台视角组织的目录,尤其是 `system/` 大目录承载了多数用例,与现有菜单与插件边界不再对齐。同时,现有 Playwright 套件仍以单 worker、每个测试重复 UI 登录、广泛依赖 `waitForTimeout` 的方式运行,导致完整回归耗时持续升高,开发反馈周期过长。

本次变更需要把 E2E 套件从“能跑”提升到“结构清晰、模块可定位、分层可执行、运行可提速”的状态,为后续菜单继续演进、插件扩展和日常迭代回归提供稳定基础设施。

## What Changes

- 重组 `hack/tests/e2e/` 目录,按当前工作台稳定能力边界与插件归属重新划分模块,拆解过载的 `system/` 大目录,并允许如 `scheduler/job/` 这样的二级模块目录。
- 将不符合测试用例约定的共享 helper、调试脚本从 `e2e/` 目录迁出到专用支持目录,同时修复重复 TC ID 与非 `TC*.ts` 文件混入测试树的问题。
- 为 Playwright 套件新增分层执行策略,至少提供 `smoke`、模块定向运行和 `full` 全量回归三类入口,避免日常开发默认依赖整轮全量 E2E 才能获得反馈。
- 引入认证态复用机制,将 `adminPage` 等高频夹具从“每个测试重复走 UI 登录”改为“每轮执行预生成并复用登录态”。
- 在高频页面对象与公共夹具中收敛固定时长等待,改为基于页面加载、表格渲染、弹窗状态、接口返回等确定性信号的等待策略,并为可并行的模块建立安全运行边界。
- 补充 E2E 套件治理文档或校验脚本,确保后续新增测试持续满足目录归属、TC 唯一性、辅助文件放置和执行分层约定。

## Capabilities

### New Capabilities
- `e2e-suite-organization`: 规范 E2E 测试目录、模块归属、辅助文件位置和 TC 编号治理,使测试结构与当前工作台能力边界一致。
- `e2e-suite-execution-efficiency`: 规范 E2E 套件的分层执行、认证态复用、状态型等待与并行安全边界,降低完整回归和日常开发反馈成本。

### Modified Capabilities
- 无

## Impact

- **测试目录**: `hack/tests/e2e/` 目录结构、内部导入路径、共享 helper/debug 文件位置将发生调整。
- **测试运行器**: `hack/tests/playwright.config.ts`、`hack/tests/package.json` 需要新增分层执行入口、登录态预生成与 worker 策略。
- **夹具与页面对象**: `hack/tests/fixtures/`、`hack/tests/pages/` 中高频登录与等待逻辑需要重构,并补充可复用等待工具。
- **文档与治理**: 需要更新 `hack/tests` 相关说明,并补充自动校验机制,防止重复 TC ID、错误目录归属和辅助文件误放再次出现。
- **回归验证**: 需要对目录迁移后的关键模块运行定向回归,并对全量 E2E 执行时间与稳定性做基线对比验证。
