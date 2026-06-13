## 1. 基线与治理测试

- [x] 1.1 盘点当前`plugin.New()`、`runtime_wiring.go`、`integration_wiring.go`、`lifecycle_wiring.go`、`sourceupgrade.New`、`wasm.Configure*`、`runtimecache`和`capabilityhost/internal/*cap`调用点，记录依赖 owner、共享实例策略和迁移顺序。
- [x] 1.2 新增或扩展静态边界测试，先覆盖 B 的最终验收目标：生产 wiring setter 清零、`ValidateRequiredDependencies`清零、包级 mutable runtime state 清零、`internal/service/plugin/runtimecache`导入清零、`wasm.Configure*`生产调用清零、`testutil_services.go`不复刻 setter 接线。
- [x] 1.3 记录本变更无 HTTP API、DTO、SQL schema、前端页面、插件 manifest wire 和`apps/lina-plugins/*`文件影响；记录`i18n`、数据权限、缓存一致性和测试策略影响判断。

  记录：`plugin.New()`已按 catalog/store/lifecycle/runtime/integration/sourceupgrade 分阶段构造内部 service；`integration_wiring.go`和`lifecycle_wiring.go`已删除；`runtime_wiring.go`只保留 nil-safe helper；`sourceupgrade.New`接收窄服务契约，不接收插件根门面；`internal/wasm`不再提供生产`Configure*`包级入口；`runtimecache`迁移到`cachecoord/revisionctrl`；`capabilityhost/internal/*cap`已合并。DI owner：`cachecoord/revisionctrl`由 cachecoord 服务与拓扑 owner 提供共享 revision 后端；WASM runtime 由插件服务组合根持有并复用启动期 capability directory、config factory、host config service 和 manifest factory；integration shared state 由插件服务组合根显式创建；lifecycle observer 与 runtime reconciler 状态归属各自 service 实例。影响分析：无 HTTP API、DTO、SQL schema、前端页面、插件 manifest wire 和`apps/lina-plugins/*`文件变更；无运行时用户可见文案或翻译资源变更，`i18n`仅消费迁移后的 revision controller；数据权限语义不变，能力适配器继续使用既有 tenant/data-scope 过滤；缓存一致性见 5.3；测试策略为后端单元测试、启动绑定测试、OpenSpec 严格校验和静态治理测试，无 UI 变化，未触发 E2E。

## 2. 构造函数直化与循环切断

- [x] 2.1 将`integration`的 topology、bizctx、capability、organization capability、dynamic job executor 和共享 state 改为构造函数显式参数，删除`integration_wiring.go`生产 setter。
- [x] 2.2 将`lifecycle`的 reconciler 和 topology 改为构造函数显式参数，删除`lifecycle_wiring.go`生产 setter。
- [x] 2.3 将`runtime`的 topology、menu/resource/hook/permission 窄接口、JWT/upload/user context/session/cache notifier/dependency validator/storage cleanup 等依赖改为构造函数显式参数，删除`runtime_wiring.go`生产 setter 和`ValidateRequiredDependencies`。
- [x] 2.4 将`sourceupgrade`依赖插件门面的路径改为窄接口注入，删除插件根包对`sourceupgradeinternal.Service`的实现断言，确保`executeRuntimeUpgradeByType`的 source 路径仍保留既有治理语义。
- [x] 2.5 重写`plugin.New()`为分阶段构造表达式：叶子和存储、运行层、编排层、门面组装；构造后不再有内部 service setter 调用。

  记录：`integration_wiring.go`和`lifecycle_wiring.go`已删除；`runtime.New`逐项接收 topology、菜单/资源/hook/权限窄接口、JWT/upload/user context/session/cache notifier/dependency validator/storage cleanup/WASM runtime 等依赖，`runtime_wiring.go`仅保留 nil-safe helper；`sourceupgrade.New`接收 catalog/store/lifecycle/runtime/integration/i18n/dependency validator 窄能力，不接收插件根门面；`plugin.New()`按 catalog/store/lifecycle/runtime/integration/sourceupgrade 顺序分阶段构造内部 service，并逐项接收 capability directory、tenant/org governance capability、WASM config factory、host config service 和 manifest factory。HTTP 组合根通过`plugin.RuntimeDelegate`打破 host services directory 与插件状态/生命周期契约之间的启动环，所有消费者只依赖 auth/role/menu/apidoc/provider/storage 所需窄契约；`httpstartup`生产路径不再调用`pluginSvc.Set*`或`pluginSvc.ConfigureWasmHostServices`。DI 来源检查：`RuntimeDelegate`由 HTTP 组合根创建，先传入 auth、role、menu、apidoc、tenant/org/AI provider runtime、storage runtime 和 capabilityhost，真实`pluginSvc`构造完成后在 HTTP 服务对外前绑定；capability directory 由`pluginsvc.NewHostServices`使用启动期共享 apiDoc/auth/bizctx/i18n/plugin delegate/session/AI/org/tenant/notify/cache/lock/storage 实例创建；WASM runtime 由`plugin.New`内部调用`wasm.NewRuntime`创建并持有，无包级配置快照。影响分析：无 HTTP API、DTO、SQL schema、前端页面、用户可见文案、插件 manifest wire 或`apps/lina-plugins/*`文件影响；`i18n`资源无变更；数据权限语义不变；缓存一致性不改变权威源和 revision 协议；测试策略新增 root setter 静态治理、插件全包 Go 测试、用户测试、i18n controller 测试和 HTTP cmd 启动绑定测试。
- [x] 2.6 收窄 HTTP middleware 启动装配依赖，删除未使用的完整`pluginsvc.Service`字段和构造参数，降低根组合图中不必要的插件门面宽依赖。

  记录：`middleware.New`不再接收`pluginSvc`，`serviceImpl`不再保存完整插件服务；HTTP 启动组合根和测试调用点已同步更新，`TestPluginWiringStateStaticBoundaries`阻断 middleware 重新导入完整插件门面。DI 来源检查：本项未新增运行期依赖，只删除未使用的宽依赖；middleware 继续复用启动期共享的 auth、bizctx、config、i18n、role 和 tenant 实例。影响分析：无 HTTP API、DTO、SQL schema、前端页面、用户可见文案、插件 manifest wire 或`apps/lina-plugins/*`文件影响；数据权限语义不变；缓存一致性无新增影响；测试策略为`go test ./internal/service/plugin -run TestPluginWiringStateStaticBoundaries -count=1`、`go test ./internal/service/middleware -count=1`和`go test ./internal/cmd -count=1`。

## 3. 包级状态实例化

- [x] 3.1 将`integration.defaultSharedState`改为组合根显式创建并注入，更新 route guard 和`integration_shared_state_test.go`覆盖同进程多实例共享状态。
- [x] 3.2 将生命周期 observer 表迁入`serviceImpl`实例或显式 observer registry，并更新相关测试，消除根包包级 observer map。
- [x] 3.3 将`pluginRuntimeCacheObservedRevision`迁入插件 service 实例或缓存协调 controller 实例，确保 runtime cache refresh 语义不变。
- [x] 3.4 将`runtime`的`reconcilerOnce`和`reconcileMu`迁入 runtime service 实例或显式共享对象，更新 reconciler safety 测试。

  记录：`integration`共享状态由`integration.NewSharedState()`创建并注入，`integration_shared_state_test.go`覆盖同进程多实例共享；生命周期 observer 由`serviceImpl.lifecycleObservers`持有；runtime cache observed revision 由`revisionctrl.Controller`/`ObservedRevision`实例持有；runtime reconciler `sync.Once`和`sync.Mutex`为`runtime.serviceImpl`字段。

## 4. WASM host service runtime 实例化

- [x] 4.1 新增 WASM host service runtime 实例构造入口，显式接收 capability directory、plugin config factory、host config service 和 manifest factory，缺失依赖时返回 error。
- [x] 4.2 改造 dynamic runtime 和 host call 分发路径，从注入的 WASM runtime 实例读取依赖，不再读取包级`atomic.Pointer`快照。
- [x] 4.3 删除生产`wasm.Configure*`入口和包级 snapshot；测试 fixture 改为创建独立 WASM runtime 实例。
- [x] 4.4 审查`apps/lina-core/pkg/plugin`下 README/中文镜像是否需要因 WASM 配置方式变化同步更新，并记录结论。

  记录：WASM runtime 构造入口为`wasm.NewRuntime(domainServices, pluginConfigFactory, hostConfigService, manifestFactory)`，生产`internal/wasm`包内无`func Configure*`和`atomic.Pointer`配置快照；缺失依赖由构造期错误暴露。`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`描述的是稳定插件公共契约，本次仅调整宿主内部 WASM host service 装配方式，无需同步更新。

## 5. runtime revision controller 迁移

- [x] 5.1 将`apps/lina-core/internal/service/plugin/runtimecache`迁移到缓存协调职责包，例如`internal/service/cachecoord/revisionctrl`，并更新 package 注释和导入。
- [x] 5.2 更新`plugin`、`runtime`、`i18n`和`internal/cmd`测试引用新路径；`WithTenantScope`改为返回副本或更名为语义明确的 setter，并补充测试。
- [x] 5.3 记录缓存一致性：权威源、一致性模型、失效触发点、跨实例同步、最大陈旧窗口、故障降级和可观测性未因迁包改变。

  记录：本次迁包不改变缓存协议，仅将原插件树下 revision controller 移至`internal/service/cachecoord/revisionctrl`。权威源仍为插件注册表、active release、插件节点状态和 artifact；一致性模型仍为`cachecoord.ConsistencySharedRevision`；失效触发点仍为插件 runtime/i18n/frontend/WASM 派生缓存变更后调用`MarkChanged`/`PublishChanged`；集群模式继续通过`sys_cache_revision`共享修订号跨实例同步；最大陈旧窗口仍为`5s`；故障降级仍为 conservative hide；可观测性继续由`cachecoord.DomainSpec`记录 domain、scope、reason 和 revision。

## 6. capabilityhost 目录收敛

- [x] 6.1 将`capabilityhost/internal/*cap`单文件领域适配器合并到`capabilityhost`包内职责文件，保持最小导出面。
- [x] 6.2 更新 imports、测试和构造路径；`capabilityhost.New`继续逐项接收接口型依赖，不引入 options/deps 聚合结构体。
- [x] 6.3 静态确认`capabilityhost/internal/`目录不再包含领域微包；记录无法合并的例外（若有）。

  记录：`capabilityhost/internal/*cap`已全部合并为`capabilityhost_<domain>.go`包内职责文件，未保留无法合并的例外；静态测试`TestCapabilityHostInternalMicroPackagesRemoved`覆盖旧目录不得重新出现。

## 7. testutil 与验证

- [x] 7.1 重构`internal/testutil`，删除与`plugin.New()`同构的 setter 接线复刻；测试 fixture 复用生产构造函数或测试专用组合根。

  记录：`internal/testutil/testutil_services.go`直接调用`lifecycle.New`、`wasm.NewRuntime`、`runtime.New`、`integration.New`等构造函数创建测试组合根，不再维护旧`SetDependencyValidator`、`SetMenuManager`、`SetHookDispatcher`或`SetReconciler`接线复刻；`TestPluginWiringStateStaticBoundaries`覆盖回归。
- [x] 7.2 运行`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`。
- [x] 7.3 运行`cd apps/lina-core && go test ./internal/service/i18n/... -count=1`。
- [x] 7.4 运行`cd apps/lina-core && go test ./internal/cmd -count=1`。
- [x] 7.5 运行`openspec validate straighten-plugin-wiring-state --strict`和`git diff --check`。
- [x] 7.6 完成实现、验证和影响记录后调用`lina-review`审查；审查通过后再进入后续变更 C。

  记录：`lina-review`已按当前工作区状态审查`straighten-plugin-wiring-state`，范围来源为`git status --short`、`git ls-files --others --exclude-standard`、OpenSpec 状态和本变更任务上下文；已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`和`.agents/rules/data-permission.md`。审查未发现阻塞问题；验证证据包括插件全包、`i18n`、`internal/cmd`、`middleware`和`cachecoord` Go 测试，`openspec validate straighten-plugin-wiring-state --strict`，`git diff --check`，以及旧 root setter、旧 runtimecache import、WASM `Configure*`和`capabilityhost/internal`微包清零的静态检索。审查结论仅覆盖本变更 B；`localdocs/plugin-service-complexity-refactor-plan.md`完整方案仍剩 C/D/E 后续变更未落地。
