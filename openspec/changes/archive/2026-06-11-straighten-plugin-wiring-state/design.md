## Context

`split-plugin-catalog-store-types`完成后，`catalog`已经退回清单事实源，`store`接管治理持久化，`plugintypes`成为叶子类型包。当前剩余复杂度集中在装配层和运行期状态归属：

- `plugin.New()`仍先构造多个半初始化 service，再通过`Set*`回注依赖。
- `runtime`、`integration`和`lifecycle`仍暴露 wiring setter；`runtime.ValidateRequiredDependencies`在运行期补救漏接线。
- `runtime`与`integration`互相保存完整 service 实例，`sourceupgrade`仍反向持有插件门面。
- `integration`、`wasm`、生命周期 observer、runtime cache revision 和 reconciler 仍存在包级可变状态。
- `runtimecache`是通用 revision controller，却位于插件包树下并被`i18n`跨域导入。
- `capabilityhost/internal/*cap`由多个单文件微包组成，给构造和 import 别名制造噪音。

本变更是方案中的变更 B。它必须在生命周期编排下沉和升级体系统一之前完成，因为后续 C/D 需要依赖清晰的构造拓扑和实例状态边界。

## Goals / Non-Goals

**Goals:**

- 让插件服务内部 service 通过构造函数逐项显式接收运行期依赖，删除生产路径 wiring setter 和`ValidateRequiredDependencies`。
- 切断`runtime`、`integration`、`lifecycle`、`sourceupgrade`与插件门面的反向 service 持有，改为窄契约注入。
- 消除插件服务内包级可变运行期状态，把必须跨 service 实例共享的状态由组合根创建并显式注入。
- 将 WASM host service runtime 从包级配置快照迁移为显式实例，并由动态插件 runtime 持有或注入。
- 将 runtime revision controller 迁到缓存协调职责边界，供`plugin`和`i18n`分别实例化。
- 合并`capabilityhost/internal/*cap`微包到`capabilityhost`内文件，降低目录层级和别名噪音。
- 通过静态边界测试和 Go 测试固化 setter 清零、包级 mutable state 清零、旧 runtimecache import 清零和 testutil 接线复刻删除。

**Non-Goals:**

- 不迁移`plugin_lifecycle.go`、`plugin_lifecycle_source.go`和`plugin_auto_enable.go`到`internal/lifecycle`；该内容属于变更 C。
- 不统一 source/dynamic upgrade 体系，也不删除`sourceupgrade`和`runtimeupgrade`包；该内容属于变更 D。
- 不拆分`runtime/route.go`鉴权和分发管线；该内容属于可选变更 E。
- 不修改 HTTP API、DTO、SQL schema、插件 manifest wire、前端页面或动态插件 guest 协议。
- 不用`Options`、`Deps`或聚合结构体包装接口型运行期依赖。

## Decisions

### D1：构造函数一次性接收依赖，删除 setter

`runtime.New`、`integration.New`、`lifecycle.New`和根`plugin.New`按依赖拓扑扩展构造函数参数。所有接口型运行期依赖逐项列在签名中，例如 topology、menu manager、hook dispatcher、JWT config provider、session store、permission menu filter、cache notifier、dependency validator、dynamic job executor、共享 state、capability directory、tenant/org governance capability、WASM config factory、host config service 和 manifest factory。生产 service 构造完成后即为可用状态，不需要`ValidateRequiredDependencies`，也不需要根服务启动 setter 或`ConfigureWasmHostServices`二次配置。

Rationale：这直接满足后端显式依赖注入规则，新增/删除依赖由编译器暴露所有调用点。

Alternatives considered：保留 setter 但增加更严格校验。该方案仍保留半初始化窗口期和运行期兜底，不解决复杂度根因。

Implementation note：HTTP 启动组合根仍存在真实构造环：host services directory 需要插件状态/生命周期契约，插件服务又需要 host services directory 提供 provider env 和 WASM host call 依赖。本变更用`plugin.RuntimeDelegate`作为组合根专用窄代理，先传给 auth、role、menu、apidoc、tenant/org/AI provider runtime、storage runtime 和 capabilityhost；`plugin.New`接收最终 capability directory、tenant/org capability 和 WASM runtime 依赖后返回真实服务，再由组合根在对外服务前绑定代理。该代理只承载稳定窄契约，不作为业务 service locator，也不替代构造函数逐项注入。

### D2：使用窄契约切断 service 互持

`runtime`需要的菜单同步、资源引用同步、hook 分发和权限菜单过滤不会再通过持有`integration.Service`表达，而是由`integration`实现并在组合根以窄接口传入。`integration`需要动态 job 执行时，只持有`runtime`导出的`DynamicJobExecutor`窄契约。`runtime`依赖校验不再回注插件门面，而是注入一个只暴露依赖校验方法的小对象。`sourceupgrade`不再接收`plugin.Service`或门面实现，而是接收实际需要的窄接口。

Rationale：这样保留当前调用语义，同时让依赖图从组合根向内部单向流动。

Alternatives considered：把所有调用点立刻上提到 lifecycle 编排。部分 runtime reconciler 和 route 过滤不在 lifecycle 调用链上，强行上提会制造新的事件总线或转发层，超出当前变更。

Implementation note：审查根组合图时发现`middleware.New`接收完整`pluginsvc.Service`但生产代码未使用该字段。本变更删除该宽依赖，使 middleware 只保留 auth、bizctx、config、i18n、role 和 tenant 等实际运行期依赖，减少后续清理根插件启动 setter 时需要处理的组合根环。

Implementation note：根组合图中 auth、role、menu、apidoc、tenant/org/AI provider runtime 不再接收完整`pluginsvc.Service`，改为接收`RuntimeDelegate`实现的窄契约；`httpstartup`不再调用`pluginSvc.Set*`或`pluginSvc.ConfigureWasmHostServices`。

### D3：共享状态由组合根显式创建

`integration`的 route binding/enablement snapshot 仍需在同进程多 service 实例间共享。组合根创建`integration.SharedState`并传给`integration.New`，HTTP route guard 闭包持有该实例。测试 fixture 也显式创建自己的 shared state，避免测试之间共享包级状态。

生命周期 observer、plugin runtime observed revision、runtime reconciler once/mutex 改为对应 service 实例字段。需要进程级共享的 revision backend 通过`cachecoord`/coordination 服务实现，不通过插件包级变量实现。

Rationale：共享仍然存在，但 owner 明确、可测试、可替换，且不再隐藏在包加载期间。

Alternatives considered：完全取消同进程共享。当前 ghttp route 生命周期可能晚于某个 plugin service 实例，直接取消会破坏 route guard 可见性。

### D4：WASM host service 使用显式 runtime 实例

新增或收敛`wasm.Runtime`等实例对象，构造函数接收 capability services、plugin config factory、host config service 和 manifest factory。host call 分发从 runtime 实例读取依赖。生产启动路径把该实例传给动态插件 runtime；测试通过 fixture 创建独立实例。

Rationale：WASM host service 是缓存敏感和能力敏感路径，不能依赖包级`atomic.Pointer`配置快照。

Alternatives considered：继续保留包级 atomic snapshot 并只替换`Configure*`命名。该方案无法满足“无包级可变状态”验收。

### D5：runtime revision controller 迁入缓存协调边界

将`apps/lina-core/internal/service/plugin/runtimecache`迁移为`apps/lina-core/internal/service/cachecoord/revisionctrl`或等价非插件职责包。`plugin`、`runtime`和`i18n`按各自缓存域实例化 controller/observed revision。`WithTenantScope`必须改为返回副本或重命名为可变 setter；若继续可变，应只在构造阶段使用且测试固化语义。

Rationale：revision controller 服务于插件 runtime、i18n 等多个缓存域，不属于插件领域。

Alternatives considered：吸收到`cachecoord.Service`接口中。当前 controller 还承载本地 observed revision 和刷新回调，直接塞入 service 契约会扩大 cachecoord 公开面；本变更先迁到清晰的 cachecoord 子包，并保留后续继续收敛空间。

### D6：合并 capabilityhost 微包但保持显式依赖

把`capabilityhost/internal/*cap`单文件包迁为`capabilityhost_<domain>.go`等包内文件。`capabilityhost.New`继续逐项接收接口型依赖；只有同一领域内多个相邻单方法接口确认属于同一 owner 时才合并接口。

Rationale：减少目录和 import 别名噪音，同时不违反显式依赖注入规则。

Alternatives considered：用`Deps`结构体缩短`New`签名。该方案违反项目后端规则，拒绝采用。

## Risks / Trade-offs

- [Risk] 构造函数签名短期变长，触发大量测试 fixture 修改。→ Mitigation：先按编译错误逐包迁移，保留接口型依赖逐项参数，新增 testutil helper 只作为测试组合根。
- [Risk] `integration`共享状态从包级变量改为注入后，route guard 可能看不到最新状态。→ Mitigation：保留同一 shared state 实例，并用`integration_shared_state_test.go`覆盖同进程多实例场景。
- [Risk] WASM host service 从包级配置改实例后，现有 host call 入口需要携带 runtime。→ Mitigation：动态 runtime 构造时保存 wasm runtime，测试 fixture 提供独立 runtime；静态检索阻断生产代码继续调用`wasm.Configure*`。
- [Risk] runtime revision controller 迁包影响`i18n`与`cmd`测试 import。→ Mitigation：同批修改`internal/service/i18n/i18n.go`、`internal/cmd/cmd_test.go`和插件 runtime 测试，运行`go test ./internal/cmd -count=1`。
- [Risk] 合并 capabilityhost 子包可能造成文件较大。→ Mitigation：按领域拆分多个`capabilityhost_<domain>.go`文件，避免创建新的微包。

## Migration Plan

1. 先新增静态边界测试，记录当前失败目标：生产 setter、包级可变状态、旧 runtimecache import、WASM Configure 入口和 testutil 接线复刻。
2. 逐包直化构造：`integration` shared state、`lifecycle`、`runtime`、`sourceupgrade`、根`plugin.New()`和`testutil`。
3. 迁移 WASM runtime 实例和启动配置路径。
4. 迁移 runtime revision controller 到 cachecoord 子包，并更新 plugin/i18n/runtime/cmd 测试。
5. 合并 capabilityhost 微包，运行 gofmt 和全量插件测试。
6. 通过`go test ./internal/service/plugin/... -count=1`、`go test ./internal/service/i18n/... -count=1`、`go test ./internal/cmd -count=1`、`openspec validate straighten-plugin-wiring-state --strict`和静态检索验收。

Rollback 使用普通 Git 回退。本变更不引入数据库迁移、API 契约或外部 wire 变化。

## Open Questions

- 无。
