# Tasks

## Feedback

- [x] **FB-1**: 修复 Go 测试替身未实现 `bizctx.Service.Current()` 和 `tenant-filter` 测试依赖缺失导致的单元测试失败
- [x] **FB-2**: 修复 `internal/cmd` 前端静态 fallback 测试中 `/admin` 返回 `404` 的回归
- [x] **FB-3**: 运行并确认完整前端单元测试、Go 单元测试和 Playwright E2E 测试全部通过
- [x] **FB-4**: 修复旧动态插件归档产物声明 `hostruntime` 导致 `make dev` 后端启动解析失败
- [x] **FB-5**: 修复动态插件生命周期 E2E 未安装 `linapro-demo-source` 依赖导致授权弹窗确认按钮禁用并超时
- [x] **FB-6**: 修复动态插件路由权限菜单挂载父级解析过严导致 `linapro-demo-dynamic` 安装失败
- [x] **FB-7**: 修复 host service E2E 临时动态插件仍引用旧 guest API 包路径和手写路由契约缺少 `requestType` 导致的失败
- [x] **FB-8**: 修复 host service E2E 临时动态插件声明动态路由权限但缺少插件父菜单导致安装失败
- [x] **FB-9**: 修复 raw SQL 旧能力负例 artifact 未命中当前 hostServices 校验链路导致上传意外成功
- [x] **FB-10**: 修复 root plugin 测试服务未发布 source-plugin capability services 导致启动一致性单测被 tenant provider 构造错误抢先失败

## FB-10 执行记录

- 根因分析：`TestValidateStartupConsistencyRejectsInvalidPluginGovernance` 与 `TestValidateStartupConsistencyRejectsPlatformOnlyTenantScoped` 通过 `newTestService()` 使用真实 `tenantcap.New(serviceImpl, bizCtxProvider)` 作为启动一致性租户能力，但 root plugin 测试服务没有像生产 HTTP 启动路径一样调用 `serviceImpl.SetCapabilities(...)` 发布 source-plugin 能力目录。`ValidateStartupConsistency` 在收集到无效 `scope_nature/install_mode` 治理问题后继续执行租户 membership 检查，触发 `linapro-tenant-core` provider 懒构造；由于 `TenantProviderEnv` 读取不到 `BizCtx`，provider 返回 `linapro-tenant-core provider requires host bizctx service`，覆盖了测试期望的稳定启动一致性 `bizerr`。
- 修复记录：已在 `apps/lina-core/internal/service/plugin/plugin_test.go` 的 root plugin 测试服务中发布最小 `capability.Services`/`pluginhost.Services` 测试目录，使 `ServicesForPlugin(...).BizCtx()` 返回测试服务图内的 `bizCtxProvider`，并通过 `pluginlifecycle.New(serviceImpl)` 提供 nil-tolerant 生命周期能力。该修复只补齐测试启动图，保持生产启动路径不变。
- 影响分析：问题位于 root plugin 单元测试夹具服务图，不是生产启动装配缺口；真实 HTTP wiring 已通过 `pluginhostservices` 发布非空 `BizCtx` adapter。本次反馈只补齐测试服务能力装配，不修改 HTTP API、前端 UI、SQL schema、数据库访问路径、数据权限策略或缓存失效机制。`i18n` 无运行时文案、菜单、语言包、插件清单或 API 文档源文本影响。缓存一致性无新增缓存或失效逻辑。数据权限无新增数据操作边界。开发工具跨平台无脚本或工具入口变更。已读取规则文件：`openspec`、`documentation`、`architecture`、`data-permission`、`plugin`、`api-contract`、`backend-go`、`database`、`cache-consistency`、`dev-tooling`、`frontend-ui`、`testing`、`i18n`。
- 验证记录：`cd apps/lina-core && go test ./internal/service/plugin -run 'TestValidateStartupConsistencyRejects(InvalidPluginGovernance|PlatformOnlyTenantScoped)$' -count=1` 通过；`cd apps/lina-core && go test ./internal/service/plugin -run TestValidateStartupConsistency -count=1` 通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1` 通过；`make test.go` 通过，覆盖宿主、`linactl` 与官方插件 Go 单元测试和编译烟测。

## FB-8/FB-9 执行记录

- 根因分析：`TC004`、`TC005` 的临时动态插件通过 DTO 元数据生成了带 `permission` 的动态 routes，但生成的 `plugin.yaml` 没有声明任何插件菜单。当前宿主安装治理明确要求动态 route permission 必须挂到插件自身菜单下，因此安装被拒绝。`TC006` 手写 WASM artifact 同样只有 routes/permissions 和 hostServices，没有插件菜单，导致授权确认后的安装业务失败。`TC004` raw SQL 负例仍在构造旧 `lina.plugin.backend.capabilities` section；当前产品约束已要求 capability 集合由 `hostServices` 自动推导，该旧 section 不再是有效 raw SQL 申请，因此上传返回成功。
- 修复记录：已在 `TC004`、`TC005` 的临时动态插件 `plugin.yaml` fixture 中声明插件页面菜单，并将 route permission 挂到该插件菜单下；已在 `TC006` 手写 manifest section 中补充插件菜单，并扩展清理逻辑删除插件菜单与角色菜单关联；已将 raw SQL 负例改为当前 host service schema 下的非法 `data.rawSql` 方法声明，使上传链路在 host service 校验阶段拒绝。后续聚焦验证发现 `TC004`、`TC005` 的临时插件把反射型 `MustNewGuestControllerRouteDispatcher` 编进 wasip1 运行时会在 WASM init 阶段 panic，已补充 `wasip1` 专用显式 requestType dispatcher，保留 host-build 反射 dispatcher 为 `!wasip1`。
- 影响分析：本次只修改宿主级插件 E2E fixture 与当前 OpenSpec 任务记录，不修改宿主产品逻辑、HTTP API、后端生产代码、SQL schema、缓存实现或数据权限实现。`i18n`影响为测试 fixture 内新增临时插件菜单中文/英文文案，不涉及宿主运行时语言包、插件 i18n 资源或 apidoc 资源。缓存一致性无影响。数据权限无新增业务数据访问边界，既有 host service/data service 授权语义保持不变。开发工具跨平台无新增脚本或工具入口，`linactl wasm` 仍按既有入口被 E2E fixture 调用。已读取规则文件：`testing`、`openspec`、`documentation`、`plugin`、`backend-go`、`api-contract`、`database`、`data-permission`、`cache-consistency`、`i18n`。
- 验证记录：`pnpm -C hack/tests test:validate` 通过；`openspec validate stabilize-full-test-suite --strict` 通过；`git diff --check -- hack/tests/e2e/extension/plugin/TC004-runtime-wasm-host-services.ts hack/tests/e2e/extension/plugin/TC005-runtime-wasm-host-services-low-priority.ts hack/tests/e2e/extension/plugin/TC006-plugin-host-service-authorization-review.ts openspec/changes/stabilize-full-test-suite/tasks.md` 通过；聚焦验证 `pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC004-runtime-wasm-host-services.ts /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC005-runtime-wasm-host-services-low-priority.ts /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC006-plugin-host-service-authorization-review.ts --project=chromium` 通过，结果 `6 passed (34.5s)`。

## 综合验证记录

- Go 单元测试：`make test.go` 通过，结果覆盖 `13` 个模块、`129` 个测试包与 `170` 个编译烟测包，包含宿主、`linactl` 和官方插件。
- 前端单元测试：`pnpm -C apps/lina-vben test:unit` 通过，结果 `45` 个文件、`360` 个测试通过。
- Playwright E2E：通过 subagent 执行 `env -u FORCE_COLOR NO_COLOR=1 PLAYWRIGHT_HTML_OPEN=never make test`，日志 `temp/final-e2e-20260527-0240.log`，结果 parallel `35 passed`，serial `522 passed, 8 skipped`，总计 `557 passed, 8 skipped, 0 failed`。
- E2E 治理验证：`pnpm -C hack/tests test:validate` 通过，校验 `238` 个 E2E 文件、`17` 个 scope、`6` 个 smoke 文件、`220` 个 serial 文件；`pnpm -C hack/tests exec tsc --noEmit` 通过。
- OpenSpec 与格式治理：`openspec validate stabilize-full-test-suite --strict` 通过；`git diff --check` 通过。
- 运行状态：最终验证后 `make status` 显示后端 `http://127.0.0.1:9120/` 与前端 `http://127.0.0.1:5666/` 均为 `running`。
- 网络重试：本轮最终验证日志未出现依赖下载失败、连接失败或需要重试的网络错误。
