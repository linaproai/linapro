## 1. OpenSpec 边界与影响记录

- [x] 1.1 写入提案、设计和增量规范，明确本轮只处理插件运行时组合 delegate、内部 cache/upgrade adapter 和 `kvcache` 显式后端选择。
  - 已写入 `proposal.md`、`design.md` 和 `plugin-service-layout`、`service-dependency-injection-governance`、`plugin-cache-service`、`distributed-cache-coordination` 增量规范。
- [x] 1.2 记录影响分析：`i18n`、缓存一致性、数据权限、API 契约、数据库、开发工具跨平台、测试策略和 `apps/lina-core/pkg/plugin` README 同步判断。
  - 影响分析：无运行时用户可见文案、API 文档源文本、插件清单或语言包资源变化，`i18n`无影响；无 HTTP API、DTO、路由、权限标签或响应结构变化，API 契约无影响；无 SQL、DAO、DO、Entity、索引、软删除或时间字段变化，数据库无影响；不新增或修改数据读写接口、下载、聚合、授权关系或执行动作，数据权限无影响；未修改 Makefile、脚本、CI、`linactl` 或工具型 Go 代码，开发工具跨平台无影响；本变更仅收紧宿主内部组合和共享缓存服务创建，不改变公开插件 SDK 或 host service 契约，`apps/lina-core/pkg/plugin` README 无需同步。

## 2. 插件运行时组合层简化

- [x] 2.1 为 `RuntimeDelegate` 增加绑定状态检查，并让可返回错误的未绑定运行期调用返回明确诊断错误。
  - 已增加 `RuntimeDelegate.Bound()` 和 `BindService` 错误返回；认证事件、OpenAPI 投影和租户治理 guard 未绑定时返回明确错误。DI 来源检查：未新增运行期依赖；delegate owner 仍为 HTTP 启动组合根，创建位置为 `newHTTPRuntime` 和测试 fixture，绑定路径为插件根服务构造完成后调用 `BindService(pluginSvc)`。
- [x] 2.2 收紧 `runtimeCacheChangeNotifierProvider`、`dependencyValidatorProvider`、`upgradeCachePublisher` 和 `upgradeCacheFreshener` 的 nil service 语义，避免缺失依赖时静默成功。
  - 已让 integration delegate、runtime cache notifier、dependency validator 和 upgrade cache adapter 在未绑定或缺失根 service 时返回明确错误。缓存一致性影响：变更不改变缓存权威源、失效范围、跨实例同步或陈旧窗口，只防止副作用、失效或刷新未执行却返回成功。
- [x] 2.3 更新插件服务测试，覆盖未绑定 delegate、nil adapter、生产构造绑定和既有成功路径。
  - 已新增 `plugin_runtime_delegates_test.go` 覆盖未绑定 delegate、内部 provider 和 upgrade adapter；已运行 `cd apps/lina-core && go test ./internal/service/plugin -count=1` 通过。

## 3. kvcache 启动期显式后端选择

- [x] 3.1 调整 HTTP runtime 启动装配，按 `cluster.enabled` 显式创建 `kvcache` provider 和共享 `kvcache.Service`。
  - 已新增 `newHTTPKVCacheProvider`，单机返回 SQL table provider，集群返回 coordination KV provider；`kvCacheSvc` 通过 `kvcache.New(kvcache.WithProvider(kvCacheProvider))` 显式创建。DI 来源检查：`kvCacheSvc` owner 为 HTTP 启动组合根，创建位置为 `newHTTPRuntime`，传递给 `auth`、插件 host services、cron 和 `httpRuntime` 共享字段；集群后端复用同一 `coordination.Service`，不新增独立 Redis client。
- [x] 3.2 从 HTTP 生产启动路径移除对 `kvcache.SetDefaultProvider` 的拓扑选择依赖，并保留必要测试兼容边界。
  - 已删除 `configureDistributedKVCache` 和 `configureLocalKVCache` 生产启动调用，`SetDefaultProvider` 仅保留为无显式 provider 时的 fallback/test 兼容入口，注释已同步。
- [x] 3.3 更新 `kvcache` 和 HTTP 启动测试，覆盖单机 SQL table provider、集群 coordination KV provider 和无全局默认依赖的构造路径。
  - 已新增 `http_kvcache_test.go` 覆盖单机、集群、集群缺失 coordination 和进程默认值不影响 HTTP 显式 provider；已运行 `cd apps/lina-core && go test ./internal/service/kvcache -count=1`、`go test ./internal/cmd/internal/httpstartup -count=1` 通过。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate simplify-plugin-runtime-composition --strict`。
  - 已验证：`openspec validate simplify-plugin-runtime-composition --strict` 通过。
- [x] 4.2 运行覆盖变更包的 Go 测试：插件服务、`kvcache` 和 HTTP 启动绑定包。
  - 已验证：`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/kvcache ./internal/cmd/internal/httpstartup ./internal/service/user ./internal/controller/i18n -count=1` 通过；`cd apps/lina-core && go test ./internal/cmd -count=1` 通过。因当前工作区存在非本变更范围的 WASM host service 脏改，额外运行 `cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1` 通过。
- [x] 4.3 运行格式或静态检查，例如 `git diff --check`。
  - 已验证：`git diff --check` 通过。
- [x] 4.4 执行 `lina-review`，确认复杂度治理、显式依赖注入、缓存一致性、测试覆盖和无影响判断满足规则。
  - Lina 审查范围：`apps/lina-core/internal/service/plugin/plugin_runtime_delegates.go`、`plugin_upgrade_adapters.go`、`plugin_test.go`、`plugin_runtime_delegates_test.go`，`apps/lina-core/internal/service/kvcache/kvcache_backend.go`，`apps/lina-core/internal/cmd/internal/httpstartup/http_runtime.go`、`http_test.go`、`http_kvcache_test.go`，`apps/lina-core/internal/service/user/user_test_dependencies_test.go`，以及 `openspec/changes/simplify-plugin-runtime-composition`。工作区中 `hack/tools/linactl/*` 既有暂存修改/删除、`apps/lina-core/internal/service/plugin/internal/wasm/*` 既有脏改、`apps/lina-core/internal/service/jobmgmt/jobmgmt_i18n_test.go` 既有脏改与本 OpenSpec 变更无关，未纳入本轮实现范围；其中 WASM 包已额外运行包测试确认当前工作区可编译。
  - 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/dev-tooling.md`；后端审查同步使用 `goframe-v2` 技能。
  - 审查结论：未发现阻塞问题。实现未新增多层抽象，只保留必要启动 delegate 并收紧 `RuntimeDelegate`、integration delegate、cache notifier、dependency validator 和 upgrade adapter 的未绑定错误语义；`kvcache` 生产后端选择从进程默认值改为 HTTP 启动期显式 provider，集群模式复用同一 `coordination.Service`，单机模式使用 SQL table provider；新增测试自包含，涉及全局默认 provider 的测试保存并恢复原值。无 API、SQL、前端、用户可见文案、数据权限或业务插件目录影响；未触发 E2E 质量审查。
