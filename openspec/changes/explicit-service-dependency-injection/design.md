## Context

LinaPro 当前已经有一条清晰的 HTTP 启动编排路径：`cmd.Http` 创建 `configSvc`，`newHTTPRuntime` 构造 `cluster/plugin/job/middleware/cron` 等长生命周期服务，`bindHostAPIRoutes` 绑定宿主控制器，源码插件通过 `pluginhost.HTTPRegistrar`、`CronRegistrar`、Hook 和 host service 与宿主协作。

问题在于这条启动编排还没有成为完整的运行期共享服务图。大量 Controller、Middleware、Service、`pkg/pluginservice/*` 适配器和源码插件控制器仍在自身 `New()` 或 `NewV1()` 中再次调用其他 `service.New()`。典型风险包括：

- `auth.New()` 内部重新构造 `plugin/role/tenant/session`，可能与启动期 `runtime.pluginSvc`、`runtime.middlewareSvc` 使用的实例不一致。
- `middleware.New()` 内部重新构造 `auth/config/i18n/plugin/role/tenant`，而中间件是认证、权限、租户、session hot state 和运行时配置的高频路径。
- `pkg/pluginservice/auth/session/notify/i18n/config/pluginstate` 面向插件发布宿主能力，但当前可由插件或适配器自行构造孤立宿主服务图。
- 源码插件 `backend/plugin.go` 在路由注册回调中直接调用插件 Controller 的 `NewV1()`，插件 Controller 又自行构造插件 Service 或宿主 service adapter。
- WASM host service 使用包级变量和 `ConfigureXxxHostService`，需要确保启动期注入的是同一套宿主运行期依赖，而不是默认孤立实例。

本次需求明确不接受新增宿主私有组装层、通用 DI 容器或聚合接口依赖结构。因此设计必须复用现有启动编排和插件注册接缝，在已有边界上通过构造函数逐项参数完成依赖收敛。

相关利益方：

- 核心宿主维护者：需要依赖关系透明、可审查、可测试，并避免新的隐式服务图。
- 插件作者：需要稳定、清晰的宿主依赖入口，不直接触碰宿主内部构造细节。
- 运维和集群部署使用方：需要权限、配置、插件状态、session、i18n 和缓存协调在多实例下遵循同一事实源和失效策略。
- OpenSpec / lina-review 工作流：需要把依赖管理规则转化为可验证、可阻断的审查标准。

## Goals / Non-Goals

**Goals:**

- 后端生产代码中 Controller、Middleware、Service、插件适配器和 WASM host service 的接口型依赖通过构造函数参数逐项显式传入，禁止使用聚合依赖结构体整体传递多个接口对象。
- 复用现有 `cmd_http_runtime.go`、`cmd_http_routes.go`、插件 registrar 回调和 host service 注册位置作为构造边界，不新增 `internal/appgraph`、通用 service registry 包或第三方 DI 容器。
- 消除缓存敏感路径中的孤立实例，确保认证、权限、租户、session、插件运行时、运行时配置、i18n、notify、cachecoord 和插件管理复用启动期同一套实例或同一共享后端。
- 将显式依赖注入、共享实例和隐式构造禁止写入项目规范、OpenSpec 规范和 `lina-review` 审查标准。
- 提供静态扫描和测试证据，识别业务代码中新增的关键服务 `New()` 隐式构造。
- 保持测试可替换性：测试可以直接传入 fake service 或使用测试 helper 组装调用参数，不强迫启动完整 HTTP runtime。

**Non-Goals:**

- 不引入 `wire`、`dig`、`fx` 或其他通用 DI 框架。
- 不新增单独的宿主私有组装层、应用对象图包或全局 service locator。
- 不要求所有无状态小工具变成单例；无状态服务可以共享实例，也可以在明确无缓存和无运行期状态时局部构造，但必须在规范和审查中可解释。
- 不修改 REST API 语义、数据库 schema、前端 UI 或 SQL seed。
- 不把依赖注入做成动态配置能力；依赖关系仍由 Go 编译期类型约束。
- 不在本迭代重写整个插件体系；插件改造以源码插件和宿主发布服务的构造入口为主，动态插件 guest 协议保持不变。

## Decisions

### 1. 显式依赖注入是唯一生产构造规则

生产后端组件的构造函数必须逐项接收接口型依赖：

```go
func New(configSvc config.Service, pluginSvc plugin.Service, roleSvc role.Service) Service
```

Controller 同样使用直接参数：

```go
func NewV1(authSvc auth.Service, bizCtxSvc bizctx.Service) authapi.IAuthV1
```

选择规则：

- 接口型依赖一律拆分为独立构造函数参数，即使依赖数量超过 2 个也不得改用 `Dependencies`、`Deps`、`Options` 等聚合结构体承载。
- 纯值配置（如字符串、布尔、`time.Duration`、容量阈值等）可以使用专门配置结构体，但不得混入接口型运行期依赖。
- 构造函数必须校验关键依赖，缺失时返回错误或在启动构造边界快速失败；不得在构造函数内部静默 `New()` 补齐关键依赖。
- 测试 helper 可以集中准备 fake 依赖，但调用生产构造函数时仍应逐项传入，生产路径不得依赖测试默认值。

替代方案：

| 方案 | 优点 | 缺点 | 结论 |
|------|------|------|------|
| 通用 DI 容器 | 可自动解析依赖 | 增加框架复杂度，运行期错误更隐蔽 | 不采用 |
| 每个包 `Instance()` | 改动少 | 隐藏依赖、测试替换困难、插件作用域不清晰 | 不采用 |
| 新增 `internal/appgraph` | 对象图集中 | 用户明确不接受新增组装层，且现有启动编排已能承载 | 不采用 |
| 现有启动编排显式传参 | 改动可控、依赖透明、无新框架 | 需要较多构造函数签名迁移 | 采用 |

### 2. 复用现有启动编排作为构造边界

宿主构造边界保持在现有文件中：

- `cmd_http_runtime.go` 构造长生命周期服务。
- `cmd_http_routes.go` 构造并绑定宿主 Controller。
- `cmd_http_apidoc.go`、`cmd_http_hooks.go` 等启动 helper 接收 runtime 中已有实例。
- `pluginhost` registrar 在源码插件注册回调中暴露宿主发布依赖。
- WASM host service 在插件 runtime 构造或启动配置阶段接收共享服务。

允许新增轻量结构用于分组：

```go
type httpRuntime struct {
    configSvc     config.Service
    pluginSvc     plugin.Service
    authSvc       auth.Service
    roleSvc       role.Service
    i18nSvc       i18n.Service
    middlewareSvc middleware.Service
}
```

但不新增独立应用组装包，也不引入任何全局 registry 来让业务代码随处查询服务。

### 3. 高风险缓存一致性路径必须共享实例

以下组件必须复用启动期共享实例或同一共享后端，不允许在业务构造函数里创建孤立服务图：

- `auth.Service`、`session.Store`、JWT revoke / pre-token state。
- `middleware.Service` 及其依赖的 `auth/config/i18n/bizctx/role/tenant/plugin`。
- `role.Service`、权限 token access cache 和数据权限服务。
- `plugin.Service`、plugin runtime cache、source route bindings、enabled snapshot、frontend bundle 和 runtime i18n。
- `config.Service` 与 runtime parameter cache。
- `i18n.Service` 与 runtime bundle/content cache。
- `cachecoord.Service`、`coordination.Service`、`kvcache.Service`、`hostlock.Service`。
- `pkg/pluginservice/*` 面向源码插件发布的宿主服务适配器。
- WASM host service 中 cache、lock、notify、storage、config、runtime 等宿主能力。

判断原则：

- 如果组件持有缓存、失效观察状态、订阅状态、session/token 状态、插件 enabled snapshot、运行时配置快照、权限快照或跨实例协调依赖，就必须共享。
- 如果组件完全无状态、只封装纯函数或 DAO 查询，允许局部构造，但应优先接收其依赖并在审查结论中说明无缓存影响。

### 4. 源码插件通过 registrar 获取宿主发布依赖

`pluginhost.HTTPRegistrar`、`CronRegistrar` 或相邻接口需要暴露稳定的宿主依赖目录，例如：

```go
type HostServices interface {
    BizCtx() pluginservicebizctx.Service
    Config() pluginserviceconfig.Service
    I18n() pluginservicei18n.Service
    Notify() pluginservicenotify.Service
    Auth() pluginserviceauth.Service
    Session() pluginservicesession.Service
    PluginState() pluginservicepluginstate.Service
}
```

源码插件路由注册应改为：

```go
services := registrar.HostServices()
noticeSvc := noticesvc.New(services.BizCtx(), services.Notify())
group.Bind(noticecontroller.NewV1(noticeSvc))
```

插件可以继续构造自身无状态业务服务，但其宿主能力依赖必须来自 registrar 发布的共享适配器。这样插件业务仍保持独立，宿主服务实例由宿主运行期治理。

### 5. `pkg/pluginservice/*` 变成适配器契约和构造边界

`pkg/pluginservice/*` 包继续承载源码插件可见的宿主能力接口，但普通插件运行路径不再调用这些包的无参 `New()` 创建孤立适配器。

迁移方式：

- 为每个包增加显式依赖构造函数，或将 `New` 改为逐项接收接口型依赖。
- 在宿主启动期统一创建 adapter set，并通过 registrar 暴露。
- 无参 `New()` 如需保留给测试或过渡，应标记为测试/临时兼容路径，并被静态扫描限制不得出现在生产插件注册路径。
- 插件侧不直接导入宿主 `internal/service/*`，仍通过 `pkg/pluginservice/*` 接口解耦。

### 6. WASM host service 由插件 runtime 注入共享实例

WASM host service 当前存在 `ConfigureCacheHostService`、`ConfigureLockHostService` 等包级配置入口。该模式可以保留，但必须由宿主启动/插件 runtime 构造路径显式调用，并注入来自同一 `httpRuntime` 的服务或共享后端。

规则：

- 默认孤立实例只能用于测试或单独包初始化兜底，不得成为生产启动后的实际服务图。
- 每个 `ConfigureXxxHostService(nil)` 的行为必须明确：测试恢复默认、启动关闭或单机兜底，不能在业务路径中清空共享实例。
- 需要请求上下文、插件 ID、租户或权限状态的 host service，必须从 runtime dispatch context 获取，不得从全局服务再推断。

### 7. 审查标准和静态验证同步落地

本变更必须同步更新：

- `AGENTS.md`：新增显式依赖注入规则、禁止隐式服务构造、缓存敏感服务共享实例要求。
- `.agents/skills/lina-review/SKILL.md`：新增后端依赖注入审查项，检查 Controller、Service、Middleware、插件注册回调、插件 host service 和缓存敏感组件。
- OpenSpec baseline：归档后同步到 `backend-conformance`、缓存协调和插件扩展相关规范。

静态验证建议：

- 扫描生产 Go 文件中关键包的 `.New(` 调用，允许列表只包含启动构造边界、测试文件、无状态组件和明确豁免项。
- 扫描 Controller `_new.go` 中是否调用关键 service `New()`。
- 扫描 `pkg/pluginservice/*` 无参 `New()` 是否出现在源码插件生产路径。
- 对高风险组件增加单元测试，断言 Controller、Middleware、plugin registrar 和 WASM host service 共享同一 fake/revision/session/cache 实例。

### 8. 分阶段迁移，避免一次性不可审查重写

建议按风险优先级推进：

1. 规范和审查标准先落地，建立静态扫描基线。
2. 迁移 `middleware`、`auth`、`role`、`session`、`config`、`i18n`、`plugin` 等高风险宿主服务。
3. 迁移宿主 Controller 构造入口和 `bindHostAPIRoutes`。
4. 迁移 `pkg/pluginservice/*` 和 source plugin registrar 的宿主发布服务目录。
5. 迁移源码插件 Controller/Service。
6. 迁移 WASM host service 配置入口，并补齐集群/单机实例共享验证。
7. 收紧静态扫描允许列表，阻止新隐式构造。

## Risks / Trade-offs

- 构造函数签名大面积变化 → 分阶段迁移，每阶段只覆盖一组依赖链，并通过包级测试验证。
- 循环依赖暴露出来 → 使用窄接口和职责拆分降低耦合，不通过全局 registry 或聚合依赖结构体绕过循环。
- 测试构造成本上升 → 提供测试 helper 和 fake dependencies，但保持生产路径显式。
- 过渡期同时存在旧 `New()` 和新 `New(deps)` → 静态扫描标记旧入口使用位置，任务末期移除或限制旧入口。
- 插件注册回调接口扩展影响所有源码插件 → 先在 registrar 中新增能力，再逐个插件迁移，保留编译期接口检查。
- 缓存敏感实例共享验证不足 → 对中间件、plugin runtime、pluginservice adapter 和 WASM host service 使用同一 fake backend 的测试证明共享。
- 无状态服务被过度工程化 → 审查允许明确无状态且无缓存影响的轻量局部构造，但必须不依赖关键运行时状态。
- `lina-review` 规则过于宽泛导致误报 → 维护允许列表和示例，区分生产路径、测试路径、启动构造边界和无状态 helper。

## Migration Plan

1. 更新 OpenSpec 规范、`AGENTS.md` 和 `lina-review` 审查标准。
2. 建立当前隐式构造扫描清单，按宿主、插件、pluginservice、WASM host service 分类。
3. 优先改造中间件和宿主核心服务依赖链，确保认证、权限、session、runtime config、plugin runtime 和 i18n 共享实例。
4. 改造宿主 Controller 构造函数和路由绑定，所有 Controller 从 `httpRuntime` 或其局部变量接收依赖。
5. 改造源码插件 registrar 和 `pkg/pluginservice/*`，使插件通过 host-published services 获取宿主能力。
6. 改造源码插件 Controller/Service 构造函数，移除插件生产路径中的宿主 service adapter 无参构造。
7. 改造 WASM host service 配置入口，确保启动期注入共享实例。
8. 收紧静态扫描，剩余豁免必须在文档中列明原因。
9. 运行后端单元测试、源码插件测试、OpenSpec 校验、静态扫描和 `lina-review`。

## Open Questions

- 是否允许在过渡期短暂保留旧无参 `New()`，但通过注释和静态扫描禁止生产路径调用？建议允许一个迭代内过渡，最终收敛。
- `pluginhost` 的宿主服务目录放在 `HTTPRegistrar`、`RouteRegistrar` 还是新增独立 `HostContext` 更清晰？建议优先新增 `registrar.HostServices()`，避免改变现有路由 API 语义。
- 哪些无状态服务可以纳入允许列表？建议实现时按扫描结果逐项记录，默认不预先扩大豁免。
