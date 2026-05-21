## Feedback

- [x] **FB-1**: 演示控制源码插件允许通过插件管理页面安装，容易误启用全局只读保护
- [x] **FB-2**: 演示控制插件页面安装被拒绝时错误提示夹带生命周期内部信息和英文 fallback，不够友好
- [x] **FB-3**: Nightly E2E 的 `host-only`、`extension-plugin`、`plugins-1-of-5` 与 `plugins-2-of-5` 分片失败

## Implementation

- [x] 1. 为源码插件 `BeforeInstall` 生命周期传递启动自动启用安装上下文。
- [x] 2. 为 `linapro-ops-demo-guard` 注册 `BeforeInstall` 回调并拒绝非 `plugin.autoEnable` 安装。
- [x] 3. 补充 Go 单元测试覆盖普通安装拒绝、启动自动启用允许和插件回调判断。
- [x] 4. 运行 OpenSpec、Go 编译门禁和审查。
- [x] 5. 本地化生命周期 veto 原因并优化演示控制插件页面安装拒绝文案。

## Verification

- [x] `openspec validate restrict-demo-guard-manual-install --strict`
- [x] `cd apps/lina-core && go test ./pkg/pluginhost -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecycleBeforeInstall(RejectsManualWhenStartupAutoEnableRequired|ReceivesStartupAutoEnableFlag)' -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecyclePreconditionLocalizesReasonParams|TestSourceLifecycleBeforeInstallRejectsManualWhenStartupAutoEnableRequired|TestSourceLifecycleBeforeInstallReceivesStartupAutoEnableFlag' -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/i18n -count=1`
- [x] `cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] `tmpdir=$(mktemp -d) && (cd "$tmpdir" && go work init /Users/john/Workspace/github/linaproai/linapro/apps/lina-core /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-ops-demo-guard) && GOWORK="$tmpdir/go.work" go test ./backend -count=1` from `apps/lina-plugins/linapro-ops-demo-guard`
- [x] `go run ./hack/tools/linactl i18n.check`
- [x] `go run ./hack/tools/linactl test.scripts`
- [x] `go test ./apps/lina-core/internal/service/plugin/internal/runtime ./apps/lina-core/internal/service/plugin/internal/wasm -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -run 'TestUpdateStatusPreservesDynamicPluginStorage|TestDynamicLifecycle|TestUninstallForce|TestRuntimeWasm|TestExecuteDynamicWasmBridgeCreatesDemoRecord' -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./pkg/plugindb ./pkg/plugindb/host ./pkg/plugindb/shared ./pkg/pluginbridge/guest ./internal/service/plugin/internal/datahost -count=1`
- [x] `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins go test ./backend/internal/service/dynamic ./backend/api -count=1`
- [x] `pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`
- [x] `pnpm -C hack/tests exec tsc --noEmit --pretty false`
- [x] `E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep 'TC-1k'`
- [x] `E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep 'TC-1b|TC-1k'`
- [x] `E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 E2E_PARALLEL_WORKERS=1 pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC001-runtime-wasm-failure-isolation.ts e2e/extension/plugin/TC008-runtime-wasm-lifecycle-boundaries.ts e2e/extension/plugin/TC011-plugin-runtime-upgrade.ts e2e/extension/plugin/TC012-plugin-status-switch-feedback.ts --config playwright.config.ts --project=chromium --workers=1`

## Review Notes

- [x] i18n 影响：新增并优化 `linapro-ops-demo-guard` 插件运行时错误文案，同步宿主生命周期错误模板、插件 `manifest/i18n/en-US/error.json` 与 `manifest/i18n/zh-CN/error.json`，并通过 `linactl i18n.check`。
- [x] 缓存一致性：本次不新增缓存；启动自动启用仍复用既有 registry 收敛、启用快照刷新、runtime cache 修订通知和集群主节点共享副作用控制。
- [x] 数据权限影响：本次不新增或扩大数据操作接口；插件安装仍为平台插件治理写操作，由既有平台治理权限和生命周期前置条件控制。
- [x] 开发工具脚本影响：未新增或修改开发脚本；`linactl test.scripts` 已通过。
- [x] `/lina-review`：审查范围包含源码插件生命周期输入、启动自动启用安装路径、生命周期 veto 原因本地化、演示控制插件 `BeforeInstall`、panic allowlist、插件 i18n JSON、单元测试和 OpenSpec 记录；未发现阻塞问题。
- [x] FB-3 i18n 影响：未新增或修改用户可见文案和翻译键；前端改动仅调整插件生命周期请求超时、状态开关等待和既有提示复用。
- [x] FB-3 缓存一致性：修复 WASM 编译模块缓存租约与精确路径失效，避免运行中 runtime 被关闭；插件安装、升级、卸载仅按插件 runtime artifact、前端 bundle 和 i18n 插件作用域失效，未新增无作用域全量失效。
- [x] FB-3 数据权限影响：未新增 HTTP 数据操作接口，也未扩大宿主或插件数据访问范围；动态插件示例记录仍走既有插件路由和授权边界。
- [x] FB-3 开发工具脚本影响：未新增或修改默认开发工具入口；验证使用现有 Go、pnpm 与 Playwright 命令。
