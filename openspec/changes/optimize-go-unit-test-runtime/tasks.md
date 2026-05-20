## 1. 测试执行入口治理

- [x] 1.1 为 `linactl test.go` 增加 Go package 测试发现能力，区分含 `_test.go` 的包和无测试包。
- [x] 1.2 调整 `linactl test.go` 默认执行计划，只对含测试文件的包执行单元测试，并输出模块级测试计划摘要。
- [x] 1.3 为 `linactl test.go` 增加模块耗时统计、race/verbose 状态和无测试包数量摘要。
- [x] 1.4 补充 `linactl` 单元测试，覆盖测试包发现、命令参数、无测试包跳过和执行摘要。

## 2. 重型插件测试治理

- [x] 2.1 审查最慢的真实 dynamic Wasm 单测，标识普通逻辑测试与真实 smoke 测试边界。
- [x] 2.2 将普通插件 runtime/cron 逻辑测试改为使用 synthetic artifact、fake executor 或轻量 fixture，避免重复执行真实 bundled Wasm 样例。
- [x] 2.3 为真实 bundled dynamic Wasm 样例保留最小 smoke 覆盖，并复用一次性 artifact fixture。

## 3. 测试基础设施清理

- [x] 3.1 修正 GitHub Actions PostgreSQL health check，显式使用 `postgres` 用户和 `linapro` 数据库。
- [x] 3.2 评估并启用安全的 Go module/build cache 配置，覆盖 host-only 与 plugin-full 单测需要的 `go.sum` 文件。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate optimize-go-unit-test-runtime --strict`。
- [x] 4.2 运行 `cd hack/tools/linactl && go test ./... -count=1` 验证工具改动。
- [x] 4.3 运行受影响的插件 runtime/integration 包测试，保留 `-race` 覆盖并记录耗时变化。
- [x] 4.4 记录 i18n、缓存一致性、开发工具跨平台影响评估，并执行 `lina-review` 审查。

## 5. 执行记录

- i18n 影响评估：本变更只调整测试工具、测试用例和 CI 配置，不新增、修改或删除用户可见运行时文案，不需要维护前端运行时语言包、插件 manifest/i18n 或 apidoc i18n 资源。
- 缓存一致性影响评估：生产运行时缓存逻辑未变更；`setup-go` cache 只作用于 CI Go module/build cache，不涉及 LinaPro 运行时缓存一致性。dynamic Wasm 测试仍显式清理自身 artifact、治理表和测试状态。
- 数据权限影响评估：未新增或修改后端数据操作接口；仅测试用例继续使用既有测试数据库清理逻辑。
- 开发工具跨平台影响评估：`linactl test.go` 的包发现与执行计划使用 Go 工具链和 Go 标准库实现，未新增 shell、sed、awk、grep 等平台专属默认开发路径依赖。
- `lina-review` 审查结论：未发现阻断问题。默认 Go 单元测试入口仍保持 `race=true`，测试包和无测试包编译 smoke 在 `race=true` 时均携带 `-race`；本变更不涉及生产 API、运行时业务行为、数据库 schema、前端页面或用户可见文案；新增/修改的 Go 测试保持自包含与 `t.Cleanup` 清理；CI cache 只作用于 Go module/build cache，不改变运行时缓存一致性。
- 验证记录：`openspec validate optimize-go-unit-test-runtime --strict` 已通过；`git diff --check` 已通过；`cd hack/tools/linactl && go test ./... -count=1` 已通过；`cd apps/lina-core && go test -race ./internal/service/plugin/internal/integration -run 'TestListCronDeclarationsDiscoversSyntheticDynamicPreview|TestListCronDeclarationsDiscoversDisabledDynamicPlugin|TestListInstalledCronDeclarationsDiscoversInstalledDisabledDynamicPlugin' -count=1` 已通过，耗时约 1.54s；`cd apps/lina-core && go test -race ./internal/service/plugin/internal/runtime -run 'TestExecuteDynamicWasmBridgeReturnsGuestResponse|TestExecuteDeclaredCronJobUsesWasmBridge' -count=1` 已通过，耗时约 25.93s。
