## 1. 宿主领域能力实现归属

- [x] 1.1 将`apps/lina-core/internal/service/plugin/internal/hostservices`迁移为`internal/service/plugin/internal/capabilityhost`，更新包名、文件注释和测试包名
- [x] 1.2 更新`internal/service/plugin.NewHostServices`及相关导入，保持启动层只依赖`internal/service/plugin`根 facade
- [x] 1.3 静态检索确认生产代码不再导入旧`plugin/internal/hostservices`路径，且`internal/cmd`不直接导入`plugin/internal/capabilityhost`

## 2. 动态 WASM 领域能力目录收敛

- [x] 2.1 删除`ConfigureAITextHostService`、`ConfigureUserHostService`、`ConfigureOrgHostService`、`ConfigureTenantHostService`及对应领域专用全局变量
- [x] 2.2 调整`ConfigureWasmHostServices`和测试工具配置，使普通领域能力只通过`ConfigureDomainHostServices`注入一次
- [x] 2.3 调整`AI`、`User`、`Org`、`Tenant`和`data`host service 分发逻辑，统一通过共享领域能力目录按插件`ID`绑定后获取对应`*cap.Service`
- [x] 2.4 更新`internal/service/plugin/internal/wasm`相关单元测试，确保测试自包含、保存并恢复包级状态

## 3. 动态 guest 领域代理收敛

- [x] 3.1 将`AI`guest 侧普通领域 hostcall 实现迁入`pkg/plugin/pluginbridge/guest/internal/domainhostcall`
- [x] 3.2 将`guest.Services.AI()`返回值收敛为复用或实现`aicap.Service`，移除与`aicap`平行的长期公共`guest.AI*Service`接口
- [x] 3.3 更新`pluginbridge/internal/hostservice`descriptor 覆盖测试，确认 public protocol alias、guest client、非 WASI stub 和 host dispatcher 仍同步

## 4. 治理验证和文档

- [x] 4.1 增加或更新治理测试，阻断普通领域能力专用`Configure*HostService`入口、领域专用全局目录和 guest 公共平行领域接口回归
- [x] 4.2 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，说明`capability`、`pluginhost`、`pluginbridge/protocol`、`pluginbridge/guest`和`capabilityhost`职责边界
- [x] 4.3 在任务记录中明确`i18n`、缓存一致性、数据权限、开发工具跨平台、E2E 和 DI 来源影响判断

执行记录：

- `i18n`：仅更新技术 README 和内部 Go 组织，不新增运行时用户可见文案、菜单、路由、API 文档源文本或语言包，确认无运行时`i18n`资源影响。
- 缓存一致性：不新增缓存、快照、失效或刷新机制；迁移后仍复用启动期共享`capability.Services`及其下游共享服务实例，确认无新的缓存一致性机制影响。
- 数据权限：不新增数据访问路径；动态插件领域调用仍进入既有`*cap.Service`并沿用`CapabilityContext`、租户和数据权限语义，确认无新增数据权限绕行。
- 开发工具跨平台：不修改`Makefile`、脚本、`linactl`或 CI；治理验证通过 Go 测试实现，确认无开发工具跨平台影响。
- E2E：本次为后端内部架构和插件桥接边界收敛，无前端 UI、路由、表单或用户可观察工作流变化，确认不触发 E2E。
- DI 来源：未新增运行期依赖 owner；`internal/service/plugin.NewHostServices`继续由启动层逐项传入共享服务并委托`capabilityhost.New`构造，`ConfigureWasmHostServices`只接收启动期共享的`capability.Services`并调用一次`ConfigureDomainHostServices`，未引入领域专用全局目录、服务定位器或临时`New()`服务图。

## 5. 验证和审查

- [x] 5.1 运行`openspec validate consolidate-plugin-domain-capability-boundaries --strict`
- [x] 5.2 运行`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`
- [x] 5.3 运行`cd apps/lina-core && go test ./internal/cmd -count=1`覆盖启动装配门禁
- [x] 5.4 执行`lina-review`，确认 OpenSpec、架构、插件、后端 Go、文档、测试、缓存、数据权限、`i18n`和开发工具影响均已闭环

验证记录：

- `openspec validate consolidate-plugin-domain-capability-boundaries --strict`：通过。
- `cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`：通过。
- `cd apps/lina-core && go test ./internal/service/file -count=1`：通过；用于验证当前工作区`file.Service`合约与文件服务实现可编译。
- `cd apps/lina-core && go test ./internal/cmd -count=1`：通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic -count=1`：通过；插件子仓库模块未加入父仓库`go.work`，因此使用`GOWORK=off`按插件模块独立验证。

`lina-review`记录：

- 审查范围：当前`consolidate-plugin-domain-capability-boundaries`变更涉及的 OpenSpec 文档、宿主插件能力边界 Go 代码、动态插件 guest/dispatcher/descriptor 同步点、`apps/lina-core/pkg/plugin`中英文 README、`linapro-demo-dynamic`示例适配，以及`internal/service/file`合约编译状态。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 结论：未发现阻塞问题；静态检索确认生产代码无旧`internal/service/plugin/internal/hostservices`导入、无普通领域专用`Configure*HostService`入口、`internal/cmd`不直接导入`capabilityhost`、公共`guest`包不再定义平行`AI*Service`接口。
- 无影响判断：无 HTTP API、SQL、前端 UI、菜单、运行时用户可见文案、缓存机制或开发工具脚本变更；数据权限路径继续通过既有`*cap.Service`和`CapabilityContext`执行。

## Feedback

- [x] **FB-1**: 将`wasm`领域 host service 分发文件按领域能力拆分，并评估是否应作为独立组件维护
- [x] **FB-2**: 统一`wasm`领域 host service 源文件命名，移除额外`domain`前缀
- [x] **FB-3**: 将动态插件集合型领域协议字符串、能力字符串和源码命名统一为领域名称
- [x] **FB-4**: 将`Plugins().Lifecycle()`作为受治理插件领域能力暴露给动态插件，并收敛`guest.PluginService`特例
- [x] **FB-5**: 删除无业务入口的`file.Service` Markdown 内容读写方法
- [x] **FB-6**: 将动态插件配置读取从独立`config`host service 收敛到`plugins.config.get`
- [x] **FB-7**: 将动态插件通知发送从独立`notify`host service 收敛到`notifications.messages.send`
- [x] **FB-8**: 将动态插件 cron 注册从运行时`guest.Services`目录移出
- [x] **FB-9**: 删除动态插件 cron discovery host-call，定时任务统一归属`jobs`领域
- [x] **FB-10**: 将源码插件定时任务注册入口从`Cron`公开契约迁移到`Jobs`领域
- [x] **FB-11**: 将旧动态插件 cron 声明能力迁移为`Jobs`领域的动态任务声明能力
- [x] **FB-12**: 将受治理运行时配置管理契约收敛到`hostconfigcap.AdminService`
- [x] **FB-13**: 新增`DynamicPlugin`动态插件声明期契约并与运行时领域能力分离
- [x] **FB-14**: 将动态插件公开开发入口从`pluginbridge/guest`收敛到`pluginbridge`
- [x] **FB-15**: 将动态`host service`payload codec 所有权从内部`hostservice`迁移到公共`protocol`目录
- [x] **FB-16**: 将动态插件声明期入口从`DynamicPlugin`重命名为`Declarations`
- [x] **FB-17**: 将源码插件声明期入口从`SourcePlugin`重命名为`Declarations`
- [x] **FB-18**: 将源码插件声明期子接口命名收敛为`*Declarations`
- [x] **FB-19**: 为 WASM `org`和`tenant`用户作用域 host service 补齐目标用户可见性校验
- [x] **FB-20**: 将 WASM host-call 错误响应载荷结构化，避免向插件暴露裸错误字符串
- [x] **FB-21**: 将 WASM host service 运行期依赖收敛为并发安全的显式快照读取
- [x] **FB-22**: 修复 WASM host service race 测试替身的共享状态竞争

FB-1执行记录：

- 根因：`hostfn_service_domain_auth.go`同时承载 token auth 与 authz，`hostfn_service_domain_metadata.go`聚合`apidoc`、`bizctx`、`dict`、`i18n`、`route`，`hostfn_service_domain_resources.go`聚合`file`、`infra`、`job`、`notification`、`plugin`、`session`，文件职责与实际领域能力边界不一致。
- 实现：保留`hostfn_service_capability.go`作为共享`capability.Services`目录和 JSON transport helper；新增`hostfn_service_<service>.go`文件分别承载`apidoc`、`auth`、`authz`、`bizctx`、`dict`、`file`、`i18n`、`infra`、`job`、`notification`、`plugin`、`route`、`session`领域分发逻辑；删除原`metadata`和`resources`聚合文件。
- 独立组件评估：当前不建议把这些领域 dispatcher 提取为`wasm`下的新子包或独立组件。原因是 dispatcher 仍依赖`wasm`包内可信`hostCallContext`、`capabilityContextForHostCall`、`capabilityJSONResponse`和共享`domainHostServices`状态；提取子包会迫使这些内部状态或 adapter 契约导出，增加转发型抽象并扩大 WASM host service 的内部导出面。当前更适合以同包领域源文件维护；若未来出现可独立测试的稳定 dispatcher 契约，再评估建立`internal/<subcomponent>`。
- `i18n`：未新增或修改运行时用户可见文案、菜单、路由、API 文档源文本、语言包或翻译缓存；仅移动 Go 注释和既有错误路径，确认无运行时`i18n`影响。
- 缓存一致性：未新增缓存、快照、失效、刷新或分布式一致性机制；`plugin`状态仍由既有`plugincap.Service.State()`提供，确认无缓存一致性影响。
- 数据权限：未新增数据访问路径或授权绕行；`file`、`dict`、`job`、`notification`、`session`等仍通过既有`capabilityContextForHostCall`和各`*cap.Service`执行，确认无新增数据权限影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：本次为后端内部文件组织重构，无前端 UI、路由、表单或端到端工作流变化，不触发 E2E；通过 Go 包测试和 OpenSpec 严格校验覆盖回归。
- DI 来源：未新增运行期依赖 owner、构造参数或启动装配路径；`ConfigureDomainHostServices`和共享`capability.Services`注入路径保持不变。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- `lina-review`：反馈级范围审查未发现阻塞问题；确认本次实际变更只涉及`wasm`领域 host service 同包文件拆分和本反馈任务记录，未引入新依赖、子包抽象、HTTP API、SQL、前端 UI、缓存机制、运行时文案或数据权限绕行。工作区存在其他既有未提交/未跟踪变更和`apps/lina-plugins`子仓库改动，已按审查流程展开状态，但不属于 FB-1 范围。

FB-2执行记录：

- 根因：FB-1拆分后新增文件采用`hostfn_service_domain_<capability>.go`命名，但`wasm`包既有 host service dispatcher 文件统一采用`hostfn_service_<service>.go`，额外`domain`前缀造成同类文件命名不一致。
- 实现：将共享能力目录文件重命名为`hostfn_service_capability.go`；将具体领域 dispatcher 文件统一重命名为`hostfn_service_<service>.go`，例如`hostfn_service_apidoc.go`、`hostfn_service_authz.go`、`hostfn_service_plugin.go`；文件头职责注释同步调整，函数实现和运行时分发逻辑不变。
- 命名约定：`hostfn_service.go`为总 dispatcher；`hostfn_service_<service>.go`为具体 host service dispatcher；`hostfn_service_capability.go`为共享`capability.Services`目录与通用 JSON transport helper。
- `i18n`：未新增或修改运行时用户可见文案、菜单、路由、API 文档源文本、语言包或翻译缓存；仅调整 Go 文件名和源码注释，确认无运行时`i18n`影响。
- 缓存一致性：未新增缓存、快照、失效、刷新或分布式一致性机制，确认无缓存一致性影响。
- 数据权限：未新增数据访问路径、授权判断或查询逻辑；所有领域调用仍通过原`capabilityContextForHostCall`和既有`*cap.Service`执行，确认无新增数据权限影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：本次为后端内部文件命名治理，无前端 UI、路由、表单或端到端工作流变化，不触发 E2E；通过 Go 包测试、静态文件检查和 OpenSpec 严格校验覆盖回归。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录或启动装配路径。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`。
- 验证：`find apps/lina-core/internal/service/plugin/internal/wasm -maxdepth 1 -type f -name 'hostfn_service_domain*.go'`无输出；`rg -n "hostfn_service_domain" apps/lina-core/internal/service/plugin/internal/wasm`无匹配；`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- `lina-review`：反馈级范围审查未发现阻塞问题；确认`wasm`目录已统一为`hostfn_service_<service>.go`命名，旧`hostfn_service_domain*.go`文件不存在，未引入新的包边界、运行期依赖、HTTP API、SQL、前端 UI、运行时文案、缓存机制或数据权限变更。工作区存在其他既有未提交/未跟踪变更和`apps/lina-plugins`子仓库改动，已按审查流程展开状态，但不属于 FB-2 范围。

FB-3执行记录：

- 根因：`capability.Services`领域目录中集合型能力使用`Users()`、`Files()`、`Jobs()`、`Notifications()`、`Plugins()`和`Sessions()`，但动态`hostServices.service`协议字符串、能力字符串、Go 常量和部分源文件仍使用`user`、`file`、`job`、`notification`、`plugin`和`session`单数形式，导致同一领域在契约、授权、guest 调用和`WASM`分发中存在命名切换。
- 实现：将集合型动态协议 service 统一为`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`；能力字符串同步为`host:users`、`host:files`、`host:jobs`、`host:notifications`、`host:plugins`和`host:sessions`；Go 常量、descriptor、public protocol alias、guest hostcall、WASM dispatcher、测试和错误诊断字符串同步改为复数领域名。
- 文件命名：将`wasm`集合型 dispatcher 文件统一为`hostfn_service_users.go`、`hostfn_service_files.go`、`hostfn_service_jobs.go`、`hostfn_service_notifications.go`、`hostfn_service_plugins.go`和`hostfn_service_sessions.go`；将`guest/internal/domainhostcall`对应文件统一为`domainhostcall_users.go`、`domainhostcall_files.go`、`domainhostcall_jobs.go`、`domainhostcall_notifications.go`、`domainhostcall_plugins.go`和`domainhostcall_sessions.go`。
- 规范与文档：更新本变更`design.md`和增量规范，明确集合型领域 service 名必须与`capability.Services`领域目录一致，且不保留旧单数别名；更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`和`apps/lina-plugins`顶层中英文 README 的公开服务目录和示例。
- `i18n`：本次修改技术文档、协议标识符和 Go 注释，不新增或修改运行时用户可见菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本；确认无运行时`i18n`资源影响。`apps/lina-plugins`仅修改顶层 README，未修改启用`i18n`的插件目录资源。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；`plugins`状态仍由既有`plugincap.Service.State()`和启动期共享实例提供，确认无缓存一致性影响。
- 数据权限：仅修改动态插件授权 service 键名，不新增数据访问路径、查询逻辑、授权判断或资源可见性语义；`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`仍通过既有`CapabilityContext`和对应`*cap.Service`执行，确认无新增数据权限绕行。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或跨平台开发工具，确认无开发工具跨平台影响。
- 测试策略：本次为后端插件协议契约和源码命名治理，无前端 UI、路由、表单或用户可观察工作流变化，不触发 E2E；通过 Go 包测试、启动装配测试、OpenSpec 校验、静态文件检查和公开文档检索覆盖回归。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图；仅更新协议常量和同一共享`capability.Services`目录上的分发键名。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/api-contract.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`；后端 Go 实现按`goframe-v2`技能检查，无 DAO/DO/Entity 或数据库操作变更。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- 静态检查：旧单数`wasm`和`domainhostcall`集合型文件名检查均无输出；`rg`确认`apps/lina-core`Go 源码中无旧精确`HostServiceUser/File/Job/Notification/Plugin/Session`、`HostServiceMethodUser/File/Job/Notification/Plugin/Session`或`CapabilityUser/File/Job/Notification/Plugin/Session`标识符；`rg`确认`pluginbridge`中无旧单数 service/capability 字符串；`rg`确认公开 README 中无旧单数`hostServices.service`目录项。
- `lina-review`：反馈级范围审查未发现阻塞问题。审查范围覆盖 FB-3 修改的 OpenSpec 文档、`pluginbridge/internal/hostservice`协议常量与 descriptor、`pluginbridge/protocol`公开别名、`pluginbridge/guest`和`guest/internal/domainhostcall`集合型代理、`wasm`集合型 dispatcher、`apps/lina-core/pkg/plugin`中英文 README、`apps/lina-plugins`顶层中英文 README。`apps/lina-plugins`子仓库已展开状态；其中`linapro-demo-dynamic`既有改动不属于 FB-3，FB-3仅涉及子仓库顶层 README 且未命中插件根目录本地`AGENTS.md`优先级。确认无 HTTP API、SQL、数据库、前端 UI、运行时文案、缓存机制、开发工具脚本或新增 DI 依赖变更；数据权限路径仍通过既有`*cap.Service`和`CapabilityContext`执行。

FB-4执行记录：

- 根因：当前`plugins`动态 host service 只发布 registry/state 方法，`hostservice`测试还显式拒绝`lifecycle.tenant_delete.ensure`；同时`guest.Services.Plugins()`返回公共`guest.PluginService`特例，未直接返回统一`plugincap.Service`，导致源码插件和动态插件在插件领域对象上不一致。
- 实现：在`plugins` host service descriptor、public protocol alias、guest hostcall 和`WASM`dispatcher 中新增`lifecycle.tenant_plugin_disable.ensure`、`lifecycle.tenant_plugin_disabled.notify`、`lifecycle.tenant_delete.ensure`和`lifecycle.tenant_deleted.notify`四个方法；运行时仍先校验`host:plugins`能力和授权快照中的精确 method，再进入`plugincap.LifecycleService`。
- Guest 收敛：将`guest.Services.Plugins()`返回值改为`plugincap.Service`；删除公共`guest.PluginService`接口；新增插件配置 adapter，让动态`Plugins().Config()`满足`plugincap.ConfigService`；`Plugins().Lifecycle()`复用`guest/internal/domainhostcall`中的 lifecycle hostcall client。
- 示例适配：`linapro-demo-dynamic`的配置读取改为依赖`Exists(ctx, key)`和`String/Bool(ctx, key, defaultValue)`，保留示例 payload 中的`Found`语义，并沿用上层`context.Context`传递到配置能力调用。插件根目录`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层规则执行。
- 文档与规范：更新本变更`design.md`和增量规范，明确插件生命周期编排归属`plugins`领域并受方法级授权治理；同步`apps/lina-core/pkg/plugin`和`apps/lina-plugins`中英文 README。
- `i18n`：本次仅修改技术文档、Go 注释、协议标识符和内部错误路径，不新增或修改运行时用户可见菜单、路由、按钮、表单、表格、提示信息、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；`plugins`状态和生命周期仍复用启动期共享`capability.Services`及既有`plugincap.LifecycleService`，确认无新增缓存一致性影响。
- 数据权限：新增方法属于执行类插件治理动作；动态调用先经`host:plugins`和精确 method 授权，再进入既有`plugincap.LifecycleService`，租户删除、租户插件禁用和跨插件生命周期前置检查继续由领域服务负责，确认无数据权限绕行。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或跨平台开发工具，确认无开发工具跨平台影响。
- 测试策略：本次为后端插件桥接契约和内部示例服务签名变更，无前端 UI、路由、表单或用户可观察工作流变化，不触发 E2E；通过 hostservice validation、descriptor 覆盖、guest stub、WASM dispatcher、动态插件 service 单元测试、插件服务大范围测试和启动装配测试验证。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图；新增 lifecycle hostcall 继续通过启动期共享`capability.Services`按插件`ID`绑定后获取`Plugins().Lifecycle()`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`；后端 Go 实现按`goframe-v2`技能检查，无 DAO/DO/Entity、SQL、HTTP API、前端 UI 或脚本变更。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/guest ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- 静态检查：`rg`确认生产 Go 代码中无公共`type PluginService interface`、无`Plugins() PluginService`、无`Config() ConfigHostService`形式的插件领域特例；`HostServiceMethodPluginsLifecycle*`在 descriptor、protocol alias、guest hostcall 和`WASM`dispatcher 中均有覆盖；动态插件示例生产路径无新增`context.Background()`。
- `lina-review`：反馈级范围审查未发现阻塞问题。审查范围覆盖 FB-4 修改的 OpenSpec 文档、`plugins`host service descriptor 和 public protocol alias、`guest.Services.Plugins()`公共契约、`guest/internal/domainhostcall`插件领域代理、`WASM`插件领域 dispatcher、动态插件示例、`apps/lina-core/pkg/plugin`与`apps/lina-plugins`中英文 README。确认无 HTTP API、SQL、数据库、前端 UI、运行时用户可见文案、缓存机制、开发工具脚本或新增 DI 依赖变更；数据权限仍由`host:plugins`、精确 method 授权和既有`plugincap.LifecycleService`执行类治理边界承担。

FB-5执行记录：

- 根因：`file.Service`中新增的`MarkdownContent`和`UpdateMarkdownContent`没有对应 OpenSpec 能力目标、HTTP API、前端入口、插件能力发布或业务消费者，只是为接口变更后的编译阻断补齐实现，扩大了核心文件服务合约。
- 实现：删除`apps/lina-core/internal/service/file/file_markdown.go`；从`file.Service`移除 Markdown 内容读写方法、输入输出结构和`MaxMarkdownContentBytes`常量；移除仅被该实现使用的`CodeFileMarkdownRequired`错误码；同步修正本任务记录中旧的 Markdown 方法补齐描述。
- 架构影响：收缩`apps/lina-core`核心文件服务合约，避免无当前需求支撑的领域方法进入通用`service`语义；不新增模块、接口、抽象层或跨模块调用。
- `i18n`：移除未发布入口使用的错误码 fallback，不新增或修改运行时用户可见文案、菜单、路由、API 文档源文本、语言包或翻译缓存，确认无运行时`i18n`资源影响。
- 缓存一致性：不新增缓存、快照、失效、刷新或分布式一致性机制，确认无缓存一致性影响。
- 数据权限：删除未发布的数据读取和更新路径，不新增列表、详情、下载、批量、写操作或插件数据访问入口，确认无新增数据权限影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：本次为后端 service 合约清理，无前端 UI、路由、表单或用户可观察工作流变化，不触发 E2E；通过 Go 包测试、启动装配测试、静态检索和 OpenSpec 严格校验覆盖回归。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`；后端 Go 实现按`goframe-v2`技能检查，无 DAO/DO/Entity、SQL、HTTP API、前端 UI、插件目录或脚本变更。
- 验证：`test ! -e apps/lina-core/internal/service/file/file_markdown.go`通过；`rg -n "MarkdownContent|UpdateMarkdownContent|MaxMarkdownContentBytes|CodeFileMarkdownRequired|FILE_MARKDOWN_REQUIRED|file_markdown" apps/lina-core -S`无输出；`cd apps/lina-core && go test ./internal/service/file -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- `lina-review`：反馈级范围审查未发现阻塞问题。审查范围覆盖`apps/lina-core/internal/service/file`当前合约、Markdown 实现文件不存在性和本反馈任务记录；确认无无入口方法残留、无 HTTP API/SQL/前端 UI/插件目录/运行时文案/缓存机制/开发工具脚本变更，数据权限路径因删除未发布读写入口而收缩，Go 编译门禁和 OpenSpec 严格校验均已通过。工作区存在其他既有未提交、未跟踪和`apps/lina-plugins`子仓库变更，不属于 FB-5 范围。

FB-6执行记录：

- 根因：插件自身配置已经由插件领域能力`Plugins().Config()`表达，但动态协议仍公开独立`service: config`、`host:config`和独立 config dispatcher，导致开发者无法从能力目录判断配置到底属于插件治理领域还是独立宿主资源。
- 实现：删除公开`HostServiceConfig`和`CapabilityConfig`，将授权方法收敛为`service: plugins`、`method: config.get`和`host:plugins`；`WASM`实现并入`hostfn_service_plugins.go`，通过`ConfigurePluginConfigServiceFactory`复用启动期注入的`plugincap.ConfigServiceFactory`，并继续保留`WithArtifactConfig(pluginID, artifactDefaultConfig)`覆盖语义；guest 公共目录移除独立`ConfigHostService`，由`guest.Services.Plugins().Config()`实现`plugincap.ConfigService`；动态示例插件和文档同步改为`Plugins().Config()`。
- API/i18n：`linapro-demo-dynamic`启用`i18n`，本次修改其 API DTO 文档源文本后同步更新`manifest/i18n/zh-CN/apidoc/plugin-api-main.json`；宿主插件授权 DTO 未新增路由，只移除公开授权投影中的旧 cron 明细并同步宿主 zh-CN apidoc JSON。
- 数据权限：插件配置读取只读取当前插件作用域配置，不新增跨插件、租户业务数据或宿主业务数据访问路径，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新或分布式一致性机制；配置读取仍复用既有配置服务和 artifact 默认配置覆盖路径。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：后端插件桥接行为变化，无前端 UI、路由、表单或端到端页面工作流变化，不触发 E2E；通过 hostservice descriptor、guest、WASM 和动态示例插件 Go 测试覆盖。
- DI 来源：未新增运行期依赖 owner；`ConfigureWasmHostServices`继续逐项接收启动期共享依赖，插件配置工厂由启动装配传入，host call 路径不创建独立服务图。

FB-7执行记录：

- 根因：通知读取已属于`Notifications()`领域，但发送能力仍通过独立`notify.send`暴露，形成`notifications`和`notify`两个相近领域入口，且旧入口无法表达同一领域内“读取无资源、发送按渠道资源授权”的方法级资源边界。
- 实现：删除公开`service: notify`、`host:notify`和`guest.Notify()`；在`notifications`领域新增`messages.send`，descriptor 改为方法级资源类型，`messages.batch_get`保持无资源读取，`messages.send`必须校验渠道`resourceRef`；`notifycap.Service`补齐`Send(ctx, capCtx, input)`，宿主 adapter 将渠道、来源、收件人和 payload 传入既有通知服务；guest 侧通过`guest.Services.Notifications().Send(...)`调用；旧`notify`运行时调用返回`NotFound`。
- API/i18n：未新增 HTTP 路由；通知发送 payload 属于 pluginbridge 协议 DTO，不进入 OpenAPI 路由。未新增运行时用户可见文案或语言包；仅清理技术注释和协议命名。
- 数据权限：发送通知先校验`host:notifications`、精确 method 和授权渠道资源引用，再进入通知领域服务；默认收件人仅为当前调用者身份，不扩大可见性或租户边界。
- 缓存一致性：未新增缓存或失效机制；通知发送仍进入既有通知服务写入路径。
- 开发工具跨平台：未修改开发工具、脚本或 CI。
- 测试策略：新增/更新`WASM`通知发送测试，覆盖默认当前用户收件人、非法 payload JSON、未授权渠道拒绝和旧`notify`服务拒绝；通过 pluginbridge 和插件服务测试复验。
- DI 来源：未新增独立通知服务实例；通知能力由启动期共享`capability.Services.Notifications()`进入 dynamic host dispatcher，`ConfigureWasmHostServices`不再单独接收 notify 运行期依赖。

FB-8执行记录：

- 根因：动态 cron 注册不是业务运行时领域能力；继续暴露在`guest.Services.Cron()`和`plugin.yaml hostServices`中会和运行时`Jobs()`任务领域混淆，并暗示插件可在业务执行期动态注册 cron。
- 实现：从公开 hostservice descriptor 和能力字符串中移除`CapabilityCron`/`service: cron`，`guest.Services`不再暴露`Cron()`；动态插件清单和 README 注释移除普通`service: cron`声明；控制器授权投影移除`cronItems`。本任务的中间方案仍保留发现期`cron.register`入口，后续 FB-9 已按用户确认删除发现期协议、包级`guest.Cron()`和动态插件示例内置 cron。
- API/i18n：移除 HTTP API 响应 DTO 中的`HostServicePermissionCronItem`/`cronItems`投影并同步宿主 zh-CN apidoc JSON；不新增前端 UI 文案或语言包。`linapro-demo-dynamic/plugin.yaml`注释同步移除普通 cron service 可选值。
- 数据权限：运行时拒绝动态 cron 注册，运行时任务读取仍归属`Jobs()`领域能力；后续 FB-9 删除动态发现期注册后，不再存在动态插件 cron 注册数据路径。
- 缓存一致性：未新增缓存、快照、失效或跨实例一致性机制。
- 开发工具跨平台：未修改开发工具、脚本或 CI。
- 测试策略：本次为后端协议/发现期行为变化，无 UI 工作流变化；后续 FB-9 更新动态插件 E2E 反向断言，确认安装授权中不再展示 Cron 授权组。
- DI 来源：未新增运行期依赖；后续 FB-9 删除 runtime 到 integration 的动态 cron executor 依赖路径。

FB-9执行记录：

- 根因：用户进一步明确“不考虑兼容性，定时任务统一由`Jobs`领域能力管理维护”。FB-8 的发现期`cron.register`保留方案仍会让动态插件拥有一条独立于`jobs`的定时任务声明路径，与“动态插件和源码插件领域能力对象统一”的目标冲突。
- 实现：删除`apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cron.go`和对应测试；删除`HostServiceCron`、`HostServiceMethodCronRegister`、cron codec、public protocol alias、`CronContract`、`ExecutionSourceCronDiscovery`、`guest.Cron()`、`CronHostService`和 WASI guest cron helper；删除 runtime 动态 cron discovery/executor、integration wiring 中的动态 cron executor 依赖和动态 cron manifest projection；动态插件示例删除`RegisterCrons`、`CronHeartbeat`、job i18n 资源和后端控制器/服务；E2E 删除动态插件内置 cron 执行断言并改为安装授权中不得出现 Cron 授权组；`jobmgmt`删除动态插件 artifact 内置 job 翻译 fallback，动态 artifact 文案测试改为插件预览文案。
- 源码插件边界：FB-9 执行时源码插件`pluginhost.CronRegistrar`仍作为源码插件生命周期资源注册入口保留；当前 FB-10 已继续将该公开入口迁移为`pluginhost.Jobs()`和`JobsRegistrar`。
- API/i18n：删除`linapro-demo-dynamic`动态内置任务运行时 i18n 资源；未新增 HTTP 路由或 API DTO；宿主 job/i18n 测试去除动态 cron 文案样例。动态插件启用`i18n`，已删除不再使用的插件`job.json`资源，保留插件预览和 API 文档翻译资源。
- 数据权限：删除动态插件主动注册和执行内置 cron 的路径，减少插件通过宿主发布服务影响任务管理的入口；`Jobs().BatchGetJobs`仍按既有 jobcap adapter 和租户过滤读取可见任务，源码插件任务投影仍由宿主 Jobs 管理链控制。
- 缓存一致性：未新增缓存、快照、失效或跨实例一致性机制；删除动态 cron discovery 执行链后无新增缓存一致性风险。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：后端协议、WASM dispatcher、guest SDK、integration、controller/cmd、jobmgmt、i18n 和动态插件后端均通过 Go 测试验证；E2E 文件已更新为反向断言，但本轮未启动浏览器执行 Playwright。
- DI 来源：删除`runtime.DynamicCronService`、integration `DynamicCronExecutor`和启动装配`SetDynamicCronExecutor`路径；未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图。

FB-6/FB-7/FB-8/FB-9共同验证记录：

- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 静态检索：`rg -n 'HostServiceCron|HostServiceMethodCron|CronHostService|guest\.Cron\(|RegisterCronsReq|DiscoverCronContracts|ExecuteDeclaredCronJob|CronCollector|CronRegistrationCollector|DeclaredCronRegistration|ExecutionSourceCronDiscovery|CronContract|hostservice_cron|hostfn_service_cron' apps/lina-core apps/lina-plugins -S`仅剩源码插件`managedCronCollector`命中；`rg -n 'service:\s*cron\b|guest\.Cron|CronHostService|hostfn_service_cron\.go|hostservice_cron|cron\.register' apps/lina-core apps/lina-plugins openspec/changes/consolidate-plugin-domain-capability-boundaries -S`仅剩源码插件`pluginhost`扩展点和 OpenSpec 禁用描述；`rg -n 'service:\s*(config|notify|cron)\b' apps/lina-core apps/lina-plugins -S`无输出；旧`HostServiceNotify`、`HostServiceMethodNotify`、`ConfigurePluginConfigHostService`、`guestServices.Cron()`和`Services.Cron()`扫描无输出。
- JSON 校验：`jq empty apps/lina-plugins/linapro-demo-dynamic/manifest/i18n/zh-CN/apidoc/plugin-api-main.json apps/lina-plugins/linapro-demo-dynamic/manifest/i18n/en-US/apidoc/plugin-api-main.json apps/lina-core/manifest/i18n/zh-CN/apidoc/core-api-plugin.json apps/lina-core/internal/packed/manifest/i18n/zh-CN/apidoc/core-api-plugin.json`通过。
- `openspec validate consolidate-plugin-domain-capability-boundaries --strict`：通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`：通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm ./internal/service/jobmgmt ./internal/service/i18n -count=1`：通过。
- `cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`：通过。
- `cd apps/lina-core && go test ./internal/controller/plugin ./internal/cmd -count=1`：通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`：通过。
- E2E：已更新`linapro-demo-dynamic`运行时 E2E，删除动态内置 cron 执行断言并增加安装授权中不出现 Cron 授权组的反向断言；本轮未启动浏览器执行 Playwright，剩余风险是该 UI 反向断言未在真实浏览器中复验。
- `lina-review`：反馈级审查未发现阻塞问题。审查范围覆盖 OpenSpec 文档、pluginbridge descriptor/protocol/codec、guest 公共目录、WASM dispatcher、启动装配、插件控制器授权投影、宿主和插件 apidoc i18n JSON、`jobmgmt`/`i18n`任务文案本地化、`apps/lina-core/pkg/plugin`与`apps/lina-plugins`中英文 README、`linapro-demo-dynamic`示例清单、后端服务和 E2E 资产。确认旧`config`/`notify`/动态`cron`公开能力不再暴露；数据权限、缓存一致性和开发工具跨平台均无新增影响；无 SQL/DAO/前端页面实现变更。

FB-10执行记录：

- 根因：FB-9 删除动态插件独立 cron host-call 后，源码插件仍保留`pluginhost.Cron()`、`RegisterCron`、`CronRegistrar`、`ExtensionPointCronRegister`和`plugin:<id>/cron:<name>`handler 引用，导致同一类定时任务能力在源码插件与动态插件之间继续暴露为`Cron`和`Jobs`两套领域对象。
- 实现：源码插件公开契约迁移为`pluginhost.Jobs().RegisterJobs(...)`、`ExtensionPointJobsRegister`、`JobsRegistrar`和`JobHandlerRegistration`；`SourcePluginDefinition`改为`GetJobRegistrars()`；宿主 integration/root plugin facade 迁移为`RegisterJobs`、`ManagedJob`、`ListExecutableJobs*`、`ListJobDeclarations*`和`ListInstalledJobDeclarations`；插件任务 handler 引用统一为`plugin:<pluginID>/jobs:<name>`并通过`BuildPluginJobHandlerRef`构造；执行来源和 capability source 统一为`jobs`。
- 源码插件适配：`linapro-demo-source`、`linapro-monitor-server`、`linapro-monitor-loginlog`和`linapro-ai-core`均改用`plugin.Jobs().RegisterJobs(...)`；`linapro-demo-source`补齐与其他源码插件一致的`lina-core v0.0.0`本地`replace`，使插件可用`GOWORK=off`独立编译；源码插件 E2E handler ref 从`/cron:`更新为`/jobs:`。
- 动态插件边界：`apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cron.go`已经删除，动态插件不再存在独立`cron`host service、发现期 host-call 或 guest SDK 入口；后续如需动态内置任务，必须通过正式`jobs`领域声明或安装期资源同步契约补齐。
- API/i18n：未新增 HTTP 路由或 API DTO；本轮仅同步技术文档、OpenSpec、Go 标识符、测试和源码插件注册入口。源码插件 job display name/description 源文本保持既有语义，不新增运行时语言包或 API 文档源文本。确认无新增`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出或批量接口；源码插件任务声明仍由宿主 Jobs 管理链投影和执行，插件启停状态、任务状态和 handler 发布仍走既有 job/jobhandler/plugin lifecycle 边界，确认无新增数据权限绕行。
- 缓存一致性：不新增缓存、快照、失效、刷新或分布式一致性机制；插件启停快照和任务 handler registry 仍复用既有宿主生命周期同步路径，确认无缓存一致性影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或代码生成入口；`linapro-demo-source/go.mod`仅补齐源码插件独立模块的本地依赖映射，无平台专属脚本影响。
- 测试策略：本次为后端插件公开契约和任务 handler ref 重构，无新增前端 UI、路由、表单或用户可观察页面工作流；更新既有源码插件 E2E 资产中的 handler ref，但本轮未启动 Playwright 浏览器执行，Go 编译和服务测试覆盖运行期契约。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图；源码插件 registrar 仍由`pluginhost`收集，宿主调度实现继续由`service/cron`和 jobhandler registry 统一管理；`linapro-monitor-loginlog`的 job callback 同步复用`loginLogServiceForHostServices`返回的共享 service，避免在 Jobs 注册路径构造并行 service 图。
- 插件本地规范：`apps/lina-plugins/linapro-demo-source/AGENTS.md`、`linapro-monitor-server/AGENTS.md`、`linapro-monitor-loginlog/AGENTS.md`和`linapro-ai-core/AGENTS.md`均不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/api-contract.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`；后端 Go 实现按`goframe-v2`技能检查。
- 静态检索：`rg -n 'SourcePluginCron|sourcePluginCron|CronRegistrar|CronRegisterHandler|CronHandlerRegistration|ExtensionPointCronRegister|RegisterCron|\.Cron\(|GetCronRegistrars|NewCronRegistrar|BuildPluginCronHandlerRef|ManagedCronJob|ListExecutableCronJobs|ListCronDeclarations|ListInstalledCronDeclarations|RegisterCrons|ExecutionSourceCron|CapabilitySourceCron|/cron:|cron\.register|RegisterJobss|Jobss|managedCron|collectManagedCron|collectDeclaredCron|buildManagedCron|HostServiceCron|HostServiceMethodCron|CronHostService|guest\.Cron\(' apps/lina-core apps/lina-plugins -S`无输出；`rg -n 'Cron\(\)|RegisterCron|CronRegistrar|ExtensionPointCronRegister|cron\.register|cron contracts|Cron registration' apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md apps/lina-plugins/README.md apps/lina-plugins/README.zh-CN.md -S`无输出；`rg -n 'cronHandlerName|cron handler|cron task|/cron:|RegisterCron|CronRegistrar|ExtensionPointCronRegister' apps/lina-plugins/linapro-demo-source apps/lina-plugins/linapro-monitor-server apps/lina-plugins/linapro-monitor-loginlog apps/lina-plugins/linapro-ai-core -S`无输出。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge/contract ./pkg/plugin/pluginbridge/protocol ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin ./internal/service/jobhandler ./internal/service/cron ./internal/service/jobmgmt -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge/... ./internal/service/plugin/... ./internal/service/jobhandler ./internal/service/cron ./internal/service/jobmgmt ./internal/cmd ./internal/controller/plugin -count=1`通过；`cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过；`cd apps/lina-plugins/linapro-monitor-server && GOWORK=off go test ./backend/... -count=1`通过；`cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过；`cd apps/lina-plugins/linapro-ai-core && GOWORK=off go test ./backend/... -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- `lina-review`：反馈级审查未发现阻塞问题。确认源码插件和动态插件定时任务公开对象已统一到`Jobs`领域，旧`Cron`公开契约、动态`cron`host service、`/cron:`handler ref 和 public guest/protocol 入口均已删除；`service/cron`仅保留调度实现职责，`cron expression`术语仍用于调度表达式，不构成插件领域能力暴露。

FB-11执行记录：

- 根因：FB-9/FB-10 已删除旧动态`cron`host service 和源码插件`Cron`公开契约，但旧动态插件“声明内置定时任务”的能力不能直接丢失；该能力需要迁移到`Jobs`领域，否则动态插件无法像源码插件一样把内置任务投影到宿主任务管理、状态控制、handler 发布和调度执行链。同时前端授权提交逻辑只提交带资源/路径/表的 service，会遗漏`jobs.register`这类仅方法授权的声明。
- 实现：新增`JobContract`、`ExecutionSourceJobsDiscovery`、`jobs.register`协议方法、guest `Jobs().Register(...)`适配器、WASM `jobs.register`发现期 dispatcher、runtime `DiscoverJobContracts`/`ExecuteDeclaredJob`和 integration 动态 Jobs executor；动态插件声明必须在`plugin.yaml`中使用`service: jobs`和`method: jobs.register`，运行期路由、hook、生命周期和任务执行中调用`jobs.register`会被拒绝；动态声明进入`ManagedJob`投影并通过`plugin:<pluginID>/jobs:<name>`handler 执行。
- 动态示例迁移：`linapro-demo-dynamic`删除旧`RegisterCrons`/`CronHeartbeat`API、controller 和 service，新增`RegisterJobs`/`JobHeartbeat`分层文件；示例`plugin.yaml`声明`jobs.register`；README、manifest README 和 E2E 资产同步改为 Jobs 语义，并继续断言安装授权中不出现 Cron 授权组。
- 前端/API 同步：前端插件 host service 授权视图移除`cronItems`和 Cron 专用展示，新增`jobs`、`plugins`、`notifications`服务标签；安装和升级授权提交逻辑改为包含“只有 methods、无资源目标”的 service，确保`jobs.register`、`plugins.config.get`等声明不会被过滤掉；同步 TypeScript DTO、运行时语言包、宿主 zh-CN apidoc JSON 和 packed public 静态资源。
- `i18n`：宿主前端运行时语言包新增`jobs`、`plugins`、`notifications`服务标签并删除 Cron 授权视图文案；宿主 apidoc zh-CN JSON 同步删除`cronItems`并更新 host service 示例；`linapro-demo-dynamic`启用`i18n`，其 API 文档和 job 资源随动态 Jobs 迁移同步。`go run . i18n.check frontend-keys`通过，仍报告 48 个既有模块级`$t()`警告，非本轮新增阻断项。
- 数据权限：`jobs.register`只在宿主 Jobs 发现执行源中收集当前插件自身声明，不读取或写入租户业务数据；普通运行期调用被拒绝。任务读取仍通过既有`Jobs().BatchGetJobs`和 jobcap adapter，动态任务执行进入插件自己的声明 route，不新增跨插件或宿主业务数据访问路径。
- 缓存一致性：未新增缓存、快照、失效或集群一致性机制；动态 Jobs 发现和执行复用现有插件 manifest、授权快照、启停状态、job handler registry 和 scheduler 链路。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI 或长期维护开发工具。验证中使用既有`pnpm`、`go test`、`openspec`和`linactl i18n.check`入口；曾由并发`rsync`同步 packed public 与 Go 编译造成一次临时`go:embed`点文件失败，待同步完成后同包测试顺序重跑已通过。
- 测试策略：后端插件桥接和调度行为通过单元/集成 Go 测试覆盖，前端授权提交通过 TypeScript 类型检查、生产构建、packed public 静态扫描覆盖；动态插件 E2E 文件已更新为 Jobs 正向断言和 Cron 反向断言。尝试执行目标 E2E 子场景时，Playwright 全局登录前置访问`http://127.0.0.1:9120/admin/auth/login`返回`ERR_CONNECTION_REFUSED`，本轮未完成真实浏览器复验，剩余风险是授权弹窗 UI 断言未在完整服务栈中跑通。
- DI 来源：新增的动态 Jobs 执行能力 owner 为`runtime.Service`；`runtime.Service`实现`DynamicJobService`接口，启动和测试装配通过`integrationSvc.SetDynamicJobExecutor(runtimeSvc)`复用同一个 runtime 服务实例；`ExecutionInput.JobCollector`只在单次 Jobs discovery 调用中传递内存 collector，不创建新的运行期服务图。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`、`linapro-demo-source/AGENTS.md`、`linapro-monitor-server/AGENTS.md`、`linapro-monitor-loginlog/AGENTS.md`和`linapro-ai-core/AGENTS.md`均不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/api-contract.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`；后端 Go 实现按`goframe-v2`技能检查，前端实现按`vben`技能检查，E2E 资产按`lina-e2e`技能检查。
- 静态检索：`rg -n 'HostServiceCron|HostServiceMethodCron|CronHostService|guest\.Cron\(|guestServices\.Cron\(|Services\.Cron\(|RegisterCrons|DiscoverCronContracts|ExecuteDeclaredCronJob|ExecutionSourceCronDiscovery|CronContract|cron\.register|service:\s*cron\b|CapabilityCron\b|BuildPluginCronHandlerRef|/cron:' apps/lina-core apps/lina-plugins -S`仅剩`hostfn_service_jobs_test.go`中旧`cron.register`负向测试；`rg -n 'cronItems|HostServicePermissionCronItem|plugin-host-service-summary-label-cron|hostServices\.service\.cron|hostServices\.cron|cronEmpty|service==="cron"|case"cron"|service:"cron"' apps/lina-vben/apps/web-antd/src apps/lina-core/internal/packed/public apps/lina-core/api apps/lina-core/internal/controller apps/lina-core/manifest apps/lina-core/internal/packed/manifest -S`无输出；`rg -n 'SourcePluginCron|sourcePluginCron|CronRegistrar|CronRegisterHandler|CronHandlerRegistration|ExtensionPointCronRegister|RegisterCron|\.Cron\(|GetCronRegistrars|NewCronRegistrar|BuildPluginCronHandlerRef|ManagedCronJob|ListExecutableCronJobs|ListCronDeclarations|ListInstalledCronDeclarations|RegisterJobss|Jobss|managedCron|collectManagedCron|collectDeclaredCron|buildManagedCron|/cron:' apps/lina-core apps/lina-plugins -S`无输出。
- 验证：`pnpm -F @lina/web-antd run typecheck`通过；`pnpm run build`通过并生成新 Web dist；packed public 已从最新 dist 同步并通过旧 Cron chunk 静态扫描；`go run . i18n.check frontend-keys`通过；`jq empty`校验宿主、前端和动态插件 JSON 通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/guest ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/integration -run 'Jobs|Cron|HostService|DefaultDirectory|Stubs|Config|Notifications' -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/controller/plugin ./internal/cmd -count=1`通过；`cd apps/lina-core && go test ./internal/service/cron ./internal/service/jobmgmt ./internal/service/jobhandler ./pkg/plugin/pluginhost -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；源码插件`linapro-demo-source`、`linapro-monitor-server`、`linapro-monitor-loginlog`和`linapro-ai-core`的`GOWORK=off go test ./backend/... -count=1`均通过；`pnpm exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --grep 'TC-1h' --workers=1`未通过，原因为本地 E2E 服务入口`127.0.0.1:9120`连接被拒绝，未进入业务断言；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- `lina-review`：反馈级审查未发现阻塞问题。审查范围覆盖 OpenSpec 文档、Jobs 协议/guest/WASM/runtime/integration、源码插件 Jobs 入口、动态示例插件、宿主插件授权 API 投影、Vben 授权弹窗与升级弹窗、宿主和插件 i18n/apidoc 资源、packed public 资源和 E2E 资产。确认旧动态`cron`公开能力没有恢复，旧`config`/`notify`公开能力未暴露，`jobs.register`按 discovery-only 受控执行，方法级授权服务能被前端提交；无 SQL/DAO 变更，无新增缓存机制，无新增数据权限绕行，无开发工具持久入口变更。

FB-12执行记录：

- 根因：`HostConfig`已经是宿主配置读取领域能力，但受治理运行时配置管理契约仍独立位于`capability/configcap`并通过`AdminServices.Config()`暴露，导致插件开发者需要区分`HostConfig`、`Plugins().Config()`和根管理`Config()`三套配置入口。该边界与其他`*cap`组件普通/管理双面的组织方式不一致，且项目无兼容性负担，应直接收敛。
- 实现：删除公开`apps/lina-core/pkg/plugin/capability/configcap`组件包；将`ConfigKey`、`Projection`、`AdminService`和内部`ScopeService`语义迁移为`hostconfigcap.RuntimeConfigKey`、`RuntimeConfigProjection`、`hostconfigcap.AdminService`和`EnsureRuntimeConfigKeysVisible`；管理方法改名为`BatchGetRuntimeConfig`和`SetRuntimeConfigJSON`，避免与普通`HostConfig().Get`混淆。`capability.AdminServices`删除`Config()`并新增`HostConfig()`，宿主`capabilityhost`管理目录和测试替身同步迁移。
- 宿主实现：原`internal/service/plugin/internal/capabilityhost/internal/configcap`迁移为`internal/.../hostconfigcap`，仍复用既有`sys_config`集合化查询、租户平台 fallback、事务内写入和`RuntimeConfig`共享 revision bump；未改变 SQL、DAO、数据模型、查询条件、事务边界或缓存失效语义。
- 文档与规范：新增增量规范要求宿主配置管理能力必须归属`HostConfig`管理面，禁止恢复独立`capability/configcap`或根`Admin().Config()`入口；同步`apps/lina-core/pkg/plugin`中英文 README，删除`capability/configcap/`目录说明并扩展`hostconfigcap`职责说明。
- `i18n`：仅修改技术 README、OpenSpec 文档、Go 接口名和注释，不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：本次不新增数据读取、写入、列表、详情、导出、下载、聚合统计或下拉接口；运行时配置管理 adapter 仍按既有租户上下文读取平台和租户行，并在写入前锁定当前可见行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新或集群一致性机制；写入路径继续在事务成功后 bump 既有运行时配置共享 revision，确认无新的缓存一致性机制影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引或迁移文件；没有手写软删除、时间字段或自增主键写入变更，确认无数据库结构和 SQL 幂等性影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或长期维护开发工具；验证使用既有`go test`、`rg`和`openspec`入口，确认无开发工具跨平台影响。
- 测试策略：本次为后端内部插件能力契约和管理目录迁移，无前端 UI、路由、表单、用户可观察页面流程或 E2E 资产变化，不触发 E2E；通过 Go 编译测试、治理静态检索和 OpenSpec 严格校验覆盖回归。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；`capabilityhost.New`仍在启动期用同一`tenantFilterSvc`构造运行时配置管理 adapter，并通过`adminDirectory.HostConfig()`发布。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/database.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`；后端 Go 实现按`goframe-v2`技能检查。
- 静态检索：`rg -n 'lina-core/pkg/plugin/capability/configcap|pkg/plugin/capability/configcap|capability/configcap/|AdminServices\.Config|Admin\(\)\.Config\(|\bBatchGetConfig\b|\bSetConfigJSON\b|\bconfigcap\.' apps/lina-core -g '*.go' -g '*.md'`无输出；公开 README 不再列出`capability/configcap`，OpenSpec 仅保留禁止恢复旧入口的治理描述；`test ! -e apps/lina-core/pkg/plugin/capability/configcap`通过；`test ! -e apps/lina-core/internal/service/plugin/internal/capabilityhost/internal/configcap`通过。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/capability/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-12 `lina-review`反馈级审查记录：

- 范围：公开`capability`聚合接口、`hostconfigcap`普通/管理双面契约、`capabilityhost`管理目录装配、运行时配置管理 adapter、测试替身、反射治理测试、插件包 README 和增量规范。
- 结论：未发现阻塞问题；`capability/configcap`公开包和内部独立 adapter 已移除，根`AdminServices.Config()`不再存在，可信源码插件管理入口统一为`AdminServices.HostConfig()`；普通`Services.HostConfig()`仍保持只读宿主配置能力。
- 性能与数据边界：运行时配置批量读取仍使用`WhereIn`集合化查询并保留平台/租户 fallback；写入仍先锁定当前可见行并在事务内 bump 共享 revision，未引入 N+1、跨租户可见性扩张或新缓存一致性路径。
- 规则影响确认：无 HTTP API、前端 UI、SQL/DAO、运行时用户可见文案、开发工具或 E2E 资产变更；相关无影响判断已在 FB-12 执行记录中列明。
- 剩余风险：稳定规范中的旧配置语义需在本变更归档时由 OpenSpec 合并流程统一收敛；当前活跃增量规范已明确禁止恢复独立`configcap`和根`Admin().Config()`入口。

FB-13执行记录：

- 根因：动态插件的`HostConfig()`、`Manifest()`和`Jobs()`曾作为`guest.Services`下的独立 guest host-call 接口暴露，其中`Jobs().Register(...)`属于插件启动/发现期的声明能力，不是业务运行期领域能力；同时动态路由声明仍使用孤立路由注册入口，和源码插件`Declarations`用一个启动期对象管理基础声明能力的设计不一致。
- 实现：新增动态插件声明期契约，统一提供`Routes()`和`Jobs()`声明 facade，让构建期路由解析、宿主 Jobs discovery 和测试可以复用同一个启动期入口。动态插件示例迁移为`RegisterPlugin(plugin pluginbridge.Declarations)`，路由声明改为`plugin.Routes().Group(...)`，Jobs 声明改为`plugin.Jobs().Register(...)`。
- 运行时领域目录：`pluginbridge.Services.HostConfig()`改为返回`hostconfigcap.Service`，`Manifest()`改为返回`manifestcap.Service`，`Jobs()`改为返回`jobcap.Service`；底层`HostConfig()`、`Manifest()`传输 helper 只作为 adapter 内部实现保留，不再作为运行时业务目录的平行领域接口。运行时`Jobs()`只保留批量读取等普通领域能力，声明期`Register`只存在于`Declarations.Jobs()`。
- 构建工具：`linactl`动态 WASM builder 改为解析`RegisterPlugin(plugin pluginbridge.Declarations)`中的`plugin.Routes().Group(...)`调用，同时保留对`RouteDeclarations`参数形态的 AST 解析能力。实现仍使用 Go AST、`filepath`和标准库路径处理，不引入平台专属脚本或 shell 语义。
- 文档与规范：更新`apps/lina-core/pkg/plugin`中英文 README 和本变更增量规范，明确动态插件启动期基础声明能力由`Declarations`承载，运行时业务代码通过`pluginbridge.Services`复用领域`*cap.Service`契约，禁止恢复运行时`Jobs().Register`、`HostConfigHostService`、`ManifestHostService`等平行公共领域接口。补回`pkg/plugin/pluginbridge/pluginbridge.go`包注释入口，使根包继续作为动态插件桥接命名空间和治理扫描锚点。
- `i18n`：仅修改技术 README、OpenSpec 文档、Go 注释和示例代码，不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计或下拉接口；`HostConfig`、`Manifest`和`Jobs`运行时调用仍进入既有领域 capability 和 host service 授权路径，`jobs.register`仍只在受控 discovery 执行源中收集当前插件自身声明，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；动态 Jobs 声明发现继续复用既有插件 manifest、授权快照、启停状态、job handler registry 和 scheduler 链路。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：修改`hack/tools/linactl/internal/wasmbuilder`中的 Go 工具实现；路由声明解析继续基于 Go AST 和标准库文件系统能力，默认开发路径无新增 Bash、PowerShell 或平台专属命令依赖。
- 测试策略：本次为后端插件桥接公开契约、构建期解析和动态示例启动入口变化，无前端 UI、页面路由、表单、表格或端到端用户工作流变化，不触发新增 E2E；通过 guest、WASM builder、动态插件后端、宿主插件能力大范围 Go 测试、静态扫描和 OpenSpec 严格校验覆盖。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；`Declarations`是 guest 侧声明 facade，默认`Jobs` facade 仍通过既有`jobs.register`host-call 进入当前 host-driven discovery collector，运行时领域能力仍由既有启动期共享`capability.Services`和 WASM dispatcher 提供。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI、SQL/DAO 和缓存规则新增实现影响。
- 静态检索：`rg -n 'JobsHostService|New\(\)\.Jobs\(\)\.Register|Default\(\)\.Jobs\(\)\.Register|guestServices\.Jobs\(\)\.Register|func RegisterRoutes|RegisterRoutes\(registrar bridgeguest\.DynamicRouteRegistrar\)|routeRegisterFunctionName' apps/lina-core apps/lina-plugins/linapro-demo-dynamic hack/tools/linactl/internal/wasmbuilder -S`仅剩 WASM host dispatcher 内部`dispatchJobsHostService`命中，未发现旧运行时公共 Jobs 注册、旧`RegisterRoutes`入口或旧 route register 常量。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/guest -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-13 `lina-review`反馈级审查记录：

- 范围：`pluginbridge.Declarations`声明期契约、`pluginbridge.Services`运行时领域目录、`HostConfig`/`Manifest`/`Jobs`guest adapter、`linactl`动态 WASM builder 路由解析、`linapro-demo-dynamic`动态插件启动入口、`pluginbridge`根包注释入口、插件 README 和本 OpenSpec 任务记录。
- 范围来源：`git status --short`、`git ls-files --others --exclude-standard`、`git -C apps/lina-plugins status/diff/ls-files`和`openspec status --change consolidate-plugin-domain-capability-boundaries --json`；当前工作区存在大量既有脏变更，FB-13 审查仅覆盖本反馈增量。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 结论：未发现阻塞问题；`Declarations`承载启动期`Routes`和`Jobs`声明，`Services`不再暴露运行时`JobsHostService`、`HostConfigHostService`或`ManifestHostService`返回值；底层`HostConfig()`和`Manifest()`host-call helper 仅作为 adapter 传输实现保留，未形成平行运行时领域目录。
- 架构与插件边界：动态插件启动期基础能力已与运行时领域能力分离，和源码插件`Declarations`边界保持一致；新增 facade 是当前声明期变化点，不是无需求支撑的转发型抽象。
- 数据权限、缓存和数据库：无新增数据操作接口、SQL、DAO、缓存、快照或失效机制；`jobs.register`仍仅在受控 discovery 中收集当前插件声明，普通运行期领域调用继续走既有 capability 授权和 WASM dispatcher。
- 开发工具跨平台：`linactl`修改位于 Go 工具实现，使用 Go AST、`filepath`和标准库能力，无新增平台专属脚本或 shell 依赖。
- 测试与 E2E：本次无前端 UI、页面路由、表单、表格或用户可观察端到端工作流变化，不触发新增 E2E；已用 Go 测试、静态扫描和 OpenSpec 严格校验覆盖。剩余风险仅是本轮未重新执行既有动态插件浏览器 E2E。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/database.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`。

FB-14执行记录：

- 根因：动态插件声明期能力已经与运行时领域能力拆开，但公开入口仍要求插件作者导入`pkg/plugin/pluginbridge/guest`，而源码插件入口是根级`pluginhost`包。动态插件规则和用户认知都更倾向于“动态插件入口就是`pluginbridge`”，继续保留公开`guest`子包会造成入口命名不对称。
- 实现：将`apps/lina-core/pkg/plugin/pluginbridge/guest`下的公开 SDK 源文件迁移到`pkg/plugin/pluginbridge`根包，包名改为`pluginbridge`；将`guest/internal/domainhostcall`迁移为`pluginbridge/internal/domainhostcall`；全仓生产 Go、`linactl`WASM builder、动态示例插件和宿主级插件 E2E 嵌入 Go 代码改为导入`lina-core/pkg/plugin/pluginbridge`。
- 治理测试：更新`capability`边界扫描和`pluginbridge/internal/hostservice`descriptor 同步测试，使治理目标变为禁止旧`guest`公开子包、禁止 root `pluginbridge`重新别名低层`protocol`符号、禁止`recordstore`直接导入`pluginbridge`，同时允许`pluginbridge`根包承载动态 SDK。
- 文档与规范：更新`apps/lina-core/pkg/plugin`中英文 README、`linapro-demo-dynamic`中英文 README、本变更`proposal.md`、`design.md`和增量规范，明确源码插件入口是`pluginhost`，动态插件入口是`pluginbridge`，低层协议目录仍由`pluginbridge/protocol`和`pluginbridge/contract`承载。
- `i18n`：仅修改技术文档、OpenSpec 文档、Go 注释、导入路径和测试嵌入代码，不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计或下拉接口；动态插件运行时能力仍通过既有`hostServices`授权、`CapabilityContext`和领域`*cap.Service`执行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；仅调整 SDK 包路径，确认无缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：修改`hack/tools/linactl/internal/wasmbuilder`中的 Go 生成器和测试，继续使用 Go AST、`filepath`和标准库文件系统能力，无新增平台专属脚本、shell 依赖、`Makefile`或 CI 入口。
- E2E：仅更新`hack/tests/e2e/extension/plugin`中嵌入 Go 源码片段的导入别名和路径，不新增测试用例、不改 TC 编号和目录结构、不改变浏览器业务断言；本轮未启动 Playwright 浏览器服务栈，使用静态检索确认旧`bridgeguest`和`pluginbridge/guest`清空。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；`pluginbridge`根包仍是 guest 侧 SDK，运行时 host service 分发继续由既有 WASM dispatcher 和启动期共享`capability.Services`提供。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`；后端 Go 实现按`goframe-v2`技能检查，E2E 资产修改按`lina-e2e`技能检查。确认无 HTTP API、前端 UI、SQL/DAO 和缓存规则新增实现影响。
- 静态检索：`rg -n 'lina-core/pkg/plugin/pluginbridge/guest|pkg/plugin/pluginbridge/guest|pluginbridge/guest|guest/internal/domainhostcall|bridgeguest' apps/lina-core apps/lina-plugins/linapro-demo-dynamic hack/tools/linactl -S`仅剩`guest_router.go`和`guest_types_aliases.go`文件名级防回归命中，无旧包路径或导入残留；`find apps/lina-core/pkg/plugin/pluginbridge -maxdepth 2 -type d`确认公开`guest`目录已删除；`rg -n 'bridgeguest|pluginbridge/guest' hack/tests/e2e/extension/plugin/TC004-runtime-wasm-host-services.ts hack/tests/e2e/extension/plugin/TC005-runtime-wasm-host-services-low-priority.ts`无输出。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/capability -run TestRepositoryPluginCapabilityBoundaries -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-14 `lina-review`反馈级审查记录：

- 范围来源：已按`lina-review`要求查看父仓库`git status --short`、`git ls-files --others --exclude-standard`，并展开`apps/lina-plugins`子仓库的`status`、`diff --name-only`和未跟踪文件；当前工作区存在大量其他活跃变更，本次审查只覆盖 FB-14 触达的`pluginbridge`公开入口迁移、动态示例插件导入适配、`linactl`动态 WASM builder、宿主级插件 E2E 嵌入 Go 片段、README 与本 OpenSpec 记录。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/i18n.md`；后端 Go 审查按`goframe-v2`技能执行，E2E 资产审查按`lina-e2e`技能执行。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 发现的问题：未发现阻塞问题；静态检索确认生产代码、动态示例插件、`linactl`构建器和宿主级 E2E 嵌入 Go 片段无旧`pluginbridge/guest`导入或`bridgeguest`别名残留，`apps/lina-core/pkg/plugin/pluginbridge`下已无公开`guest`目录。
- 规则域结论：插件和架构边界通过，动态插件公开入口已与源码插件`pluginhost`入口保持根包对称；后端 Go 通过，未新增运行期依赖、服务定位器或临时`New()`服务图；文档通过，中英文 README 与 OpenSpec 已同步；开发工具跨平台通过，`linactl`仍使用 Go AST、`filepath`和标准库；测试策略通过，本次仅改 E2E 嵌入 Go 源码片段，不新增用例或改变浏览器断言。
- 无影响判断：无 HTTP API、前端 UI、SQL/DAO、缓存机制、翻译资源、运行时用户可见文案或新增数据权限路径影响；动态插件运行时数据访问仍通过既有`hostServices`授权、`CapabilityContext`和领域`*cap.Service`执行。
- 复验：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- 剩余风险：未重新启动 Playwright 浏览器服务栈执行既有宿主级插件 E2E；本轮以 Go 测试、静态检索和 OpenSpec strict 校验覆盖该纯导入路径迁移。

FB-15执行记录：

- 根因：`pluginbridge/protocol`作为公开协议目录已经被文档和调用方视为 payload DTO 与 codec 的稳定入口，但实际实现仍放在`pluginbridge/internal/hostservice`，`protocol`通过类型别名和函数别名重新导出内部实现；这会让公共协议所有权和内部授权治理所有权混在一起。
- 实现：将动态`host service`payload struct、wire payload struct、marshal 和 unmarshal 实现迁移到`pkg/plugin/pluginbridge/protocol`的`protocol_hostservice_*_codec.go`文件；将对应 codec 单元测试迁移到`protocol`包；删除`pluginbridge/internal/hostservice`中的 codec 实现、codec 测试和 codec 专用 helper；`protocol_hostservice_types.go`仅保留`HostServiceSpec`和`HostServiceResourceSpec`清单声明类型别名。
- 治理测试：更新`pluginbridge/internal/hostservice`descriptor 同步测试，使其同时识别 public protocol codec 函数实现，并新增检查阻断`protocol_hostservice_*_codec.go`再次导入`pluginbridge/internal/hostservice`或别名导出内部 codec 函数。
- 文档与规范：更新本变更`proposal.md`、`design.md`和增量规范，明确`protocol`公开暴露协议名并拥有 payload DTO 和 codec，`internal/hostservice`拥有 descriptor、授权推导、资源形态、清单规范化和校验治理；同步更新`apps/lina-core/pkg/plugin`中英文 README 的边界说明。
- 验证补充修复：运行`cd apps/lina-core && go test ./internal/cmd -count=1`时发现当前工作区既有`pluginbridge/guest`到根包迁移后，panic allowlist 仍指向旧`pluginbridge/guest_router.go`路径；已将 allowlist 路径修正为当前`pluginbridge/pluginbridge_router.go`，未改变 panic 许可数量、函数名或原因。
- `i18n`：仅修改技术 README、OpenSpec 文档、Go 注释、内部测试和协议 codec 归属，不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计或下拉接口；仅迁移既有 payload codec owner，动态插件运行时访问仍通过既有`hostServices`授权、`CapabilityContext`和领域`*cap.Service`执行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；cache host service payload wire 格式保持原实现语义，确认无缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或跨平台执行入口；验证使用既有`go test`、`rg`和`openspec`命令，确认无开发工具跨平台影响。
- 测试策略：本次为后端内部插件协议 codec 所有权迁移，无前端 UI、页面路由、表单、表格或端到端用户工作流变化，不触发新增 E2E；通过协议 codec 单元测试、descriptor 治理测试、WASM dispatcher 包测试、插件服务大范围 Go 测试和 OpenSpec strict 校验覆盖。
- DI 来源：未新增运行期依赖 owner、构造参数、全局目录、服务定位器或临时`New()`服务图；payload codec 迁移不改变启动装配、WASM host service 注入或`capability.Services`共享实例路径。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`；后端 Go 实现按`goframe-v2`技能检查。
- 静态检索：`rg -n "pluginbridge/internal/hostservice|= hostservice\\.(MarshalHostService|UnmarshalHostService)" apps/lina-core/pkg/plugin/pluginbridge/protocol -g 'protocol_hostservice_*_codec.go'`无输出；`find apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice -maxdepth 1 -type f -name '*codec*.go' -print`无输出；`rg -n "concrete payload ownership remains in the internal hostservice|direct aliases and must stay behavior-free|owned by hostservice|codec alias" apps/lina-core/pkg/plugin/pluginbridge/protocol apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/consolidate-plugin-domain-capability-boundaries`无输出。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/internal/hostservice -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-15 `lina-review`反馈级审查记录：

- 范围来源：已按`lina-review`要求查看父仓库`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change consolidate-plugin-domain-capability-boundaries --json`，并展开`apps/lina-plugins`子仓库状态；当前工作区存在大量其他活跃变更，本次审查只覆盖 FB-15 触达的`protocol`codec 迁移、`internal/hostservice`descriptor 治理测试、`apps/lina-core/pkg/plugin`中英文 README、OpenSpec 记录和`internal/cmd`panic allowlist 路径修正。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`；后端 Go 审查按`goframe-v2`技能执行。
- 发现的问题：未发现阻塞问题；审查中发现文档初稿把 service/method 常量所有权过度描述为`protocol`拥有，已收窄为`protocol`公开暴露协议名并拥有 payload DTO 和 codec，`internal/hostservice`继续拥有 descriptor 与校验治理。
- 规则域结论：OpenSpec、架构、插件、后端 Go、文档和测试均通过；未发现`protocol_hostservice_*_codec.go`重新导入内部`hostservice`或别名内部 codec 函数；内部`hostservice`目录已无 codec 源文件。
- 无影响判断：无 HTTP API、前端 UI、SQL/DAO、数据库迁移、运行时用户可见文案、翻译资源、缓存机制、数据权限路径或新增 DI 依赖影响；动态插件运行时数据访问仍通过既有`hostServices`授权、`CapabilityContext`和领域`*cap.Service`执行。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/internal/hostservice -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
- 剩余风险：本次未启动浏览器服务栈执行既有插件 E2E；FB-15 为内部协议 codec 所有权迁移，无用户可观察页面或端到端流程变化，已用 Go 测试、静态检索和 OpenSpec strict 校验覆盖。

FB-16执行记录：

- 根因：`DynamicPlugin`承载的是动态插件声明期能力 facade，不是运行期插件实例。继续使用该名称会模糊`pluginbridge.Services`运行期能力目录与声明期`Routes()`、`Jobs()`入口之间的边界，也不利于后续与源码插件声明入口形成一致心智模型。
- 实现：将公开声明期契约统一命名为`pluginbridge.Declarations`；构造入口统一为`NewDeclarations`；选项和子 facade 命名统一为`DeclarationOption`、`WithDeclarationRoutes`、`WithDeclarationJobs`、`RouteDeclarations`和`JobDeclarations`。`RegisterPlugin(plugin pluginbridge.Declarations)`作为动态插件声明期入口，`pluginbridge.Services`继续只承载运行期普通领域能力。
- 同步范围：同步更新`linactl`动态 WASM builder 的 AST 识别，使其解析`Declarations`参数和`plugin.Routes().Group(...)`调用；同步更新动态示例插件、宿主级 E2E 嵌入 Go 片段、`pluginbridge`单元测试、`apps/lina-core/pkg/plugin`中英文 README、`linapro-demo-dynamic`中英文 README 和本变更增量规范。
- `i18n`：仅修改技术 README、OpenSpec 文档、Go 注释、公开 Go 标识符和测试嵌入代码；不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计、下拉候选或授权关系变更；动态插件运行时能力仍通过既有`hostServices`授权、`CapabilityContext`和领域`*cap.Service`执行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；声明期 facade 重命名不改变插件状态、授权快照、Jobs 声明发现或调度链路，确认无缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：修改`hack/tools/linactl/internal/wasmbuilder`中的 Go 工具实现，继续使用 Go AST、`filepath`和标准库文件系统能力；未新增 Bash、PowerShell、平台脚本、`Makefile`或 CI 入口。
- 测试策略：本次为后端插件 SDK 命名和构建期解析契约变更，无前端 UI、页面路由、表单、表格或端到端用户工作流变化，不新增 E2E；通过 Go 包测试、动态示例插件后端测试、静态检索和 OpenSpec strict 校验覆盖。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；`Declarations`是 guest 侧声明 facade，运行时领域能力仍由既有启动期共享`capability.Services`和 WASM dispatcher 提供。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI、SQL/DAO 新增实现影响。
- 静态检索：`rg -n "\bDynamicPlugin\b|NewDynamicPlugin|WithDynamicPlugin|DynamicRouteRegistrar|DynamicPluginRoutes|DynamicPluginJobs|DynamicPluginOption" apps/lina-core apps/lina-plugins hack/tools/linactl hack/tests/e2e -S`无输出；`rg -n "RegisterPlugin\([^\n]*(DynamicPlugin|DynamicRouteRegistrar)|NewDynamicPlugin|WithDynamicPlugin" openspec/changes/consolidate-plugin-domain-capability-boundaries apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md apps/lina-plugins/linapro-demo-dynamic/README.md apps/lina-plugins/linapro-demo-dynamic/README.zh-CN.md -S`无输出。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-16 `lina-review`反馈级审查记录：

- 范围来源：已按`lina-review`要求查看父仓库`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change consolidate-plugin-domain-capability-boundaries --json`，并展开`apps/lina-plugins`子仓库状态和未跟踪文件；当前工作区存在大量其他活跃变更，本次审查只覆盖 FB-16 触达的`pluginbridge`声明期契约、`linactl`动态 WASM builder、动态示例插件声明入口、README、宿主级 E2E 嵌入 Go 片段和本 OpenSpec 记录。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`；后端 Go 审查按`goframe-v2`技能执行。
- 插件本地规范：`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`不存在，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 发现的问题：未发现阻塞问题；`pluginbridge.Declarations`、`NewDeclarations`、`RouteDeclarations`和`JobDeclarations`命名已成为公开声明期入口，`pluginbridge.Services`仍只承载运行期普通领域能力；`linactl`通过 Go AST 解析`RegisterPlugin(plugin pluginbridge.Declarations)`和`plugin.Routes().Group(...)`，`astTypeName`支持选择器类型名。
- 规则域结论：OpenSpec、架构、插件、后端 Go、文档、测试、`i18n`、缓存一致性和开发工具规则均通过；确认无 HTTP API、前端 UI、SQL/DAO、数据库迁移、运行时用户可见文案、翻译资源、缓存机制、数据权限路径或新增 DI 依赖影响。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过；旧公开 SDK 名静态检索无输出。
- 剩余风险：未重新启动浏览器服务栈执行既有插件 E2E；本轮为 Go SDK 命名和构建期解析契约变更，已用 Go 测试、动态示例插件后端测试、静态检索和 OpenSpec strict 校验覆盖。

FB-17执行记录：

- 根因：源码插件公开声明期 facade 仍命名为`SourcePlugin`，容易把启动期声明对象误解为运行期插件实例；动态插件已统一为`pluginbridge.Declarations`，源码插件需要使用同样的声明期命名模型，与运行期`pluginhost.Services`区分。
- 实现：将源码插件作者侧声明期接口重命名为`pluginhost.Declarations`，构造入口重命名为`pluginhost.NewDeclarations()`；`RegisterSourcePlugin(plugin pluginhost.Declarations)`继续作为源码插件注册函数；`SourcePluginDefinition`、生命周期输入、升级输入和卸载输入等宿主读模型或运行期输入名称保持不变。
- 同步范围：更新`pkg/plugin/pluginhost`公开接口、注册表签名、单元测试、宿主测试 fixture、`i18n`和`apidoc`相关测试中的源码插件构造、内置源码插件`backend/plugin.go`入口、`apps/lina-core/pkg/plugin`中英文 README、本变更增量规范和任务记录。
- `i18n`：仅修改技术 README、OpenSpec 文档、Go 标识符、测试 fixture 和源码插件注册入口；不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计、下拉候选或授权关系变更；源码插件运行时能力仍通过`pluginhost.Services`、`CapabilityContext`和既有领域`*cap.Service`执行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；声明期 facade 重命名不改变插件状态、插件启停快照、资源同步、Jobs 调度或回调执行链路，确认无缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或跨平台执行入口；验证使用既有`go test`、`rg`和`openspec`命令，确认无开发工具跨平台影响。
- 测试策略：本次为后端插件 SDK 命名和源码插件注册契约变更，无前端 UI、页面路由、表单、表格或端到端用户工作流变化，不新增 E2E；通过 Go 包测试、源码插件后端测试、宿主启动绑定测试、静态检索和 OpenSpec strict 校验覆盖。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；`Declarations`仅为源码插件声明期 facade，运行期服务目录仍由启动期共享`capability.Services`和`pluginhost.Services`提供。
- 插件本地规范：`linapro-ops-demo-guard`、`linapro-monitor-online`、`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-demo-source`、`linapro-monitor-server`、`linapro-monitor-loginlog`、`linapro-org-core`、`linapro-tenant-core`和`linapro-ai-core`插件根目录均不存在`AGENTS.md`，按仓库顶层`AGENTS.md`和`.agents/rules/plugin.md`执行。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI、SQL/DAO 和开发工具新增实现影响。
- 静态检索：`rg -n 'pluginhost\.NewSourcePlugin\(|NewSourcePlugin\(|pluginhost\.SourcePlugin\b|type SourcePlugin interface' apps/lina-core apps/lina-plugins openspec/changes/consolidate-plugin-domain-capability-boundaries -S`仅剩`pkg/i18nresource.SourcePlugin`元数据接口；`rg -n '\bSourcePlugin\b' apps/lina-core/pkg/plugin/pluginhost apps/lina-plugins -g '*.go'`无输出；README 和 OpenSpec 中旧公开 SDK 名检索仅剩 FB-17 任务标题中的旧名说明。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginhost -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/i18n ./internal/service/apidoc -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；源码插件`linapro-monitor-online`、`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-demo-source`、`linapro-monitor-server`、`linapro-monitor-loginlog`、`linapro-org-core`、`linapro-tenant-core`和`linapro-ai-core`的`GOWORK=off go test ./backend/... -count=1`均通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。`linapro-ops-demo-guard`尝试独立执行`GOWORK=off go test ./backend/... -count=1`未通过，原因是其`go.mod`既有状态缺少`lina-core`本地`require/replace`映射，无法解析已存在的`lina-core/pkg/...`导入；本轮未改该模块元数据，已用静态检索覆盖其注册入口重命名。

FB-17 `lina-review`反馈级审查记录：

- 范围来源：已按`lina-review`要求查看父仓库`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change consolidate-plugin-domain-capability-boundaries --json`，并展开`apps/lina-plugins`子仓库状态和未跟踪文件；当前工作区存在大量其他活跃变更，本次审查只覆盖 FB-17 触达的`pluginhost`声明期契约、宿主测试 fixture、源码插件注册入口、README、增量规范和本 OpenSpec 记录。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`；后端 Go 审查按`goframe-v2`技能执行。
- 发现的问题：未发现阻塞问题；`pluginhost.Declarations`和`NewDeclarations`已成为源码插件声明期入口，`RegisterSourcePlugin`参数已改为`pluginhost.Declarations`，`pluginhost.Services`仍只承载运行期源码插件能力目录。
- 规则域结论：OpenSpec、架构、插件、后端 Go、文档、测试、`i18n`和缓存一致性规则均通过；确认无 HTTP API、前端 UI、SQL/DAO、数据库迁移、运行时用户可见文案、翻译资源、缓存机制、数据权限路径、开发工具入口或新增 DI 依赖影响。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginhost -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/i18n ./internal/service/apidoc -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；九个具备本地`lina-core`映射的源码插件后端测试通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过；旧公开 SDK 名静态检索无输出。
- 剩余风险：`linapro-ops-demo-guard`独立后端测试因既有模块依赖映射缺失未能运行，属于本反馈外的模块配置问题；本轮没有启动浏览器服务栈执行既有 E2E，因为变更不涉及前端 UI 或端到端用户路径。

FB-18执行记录：

- 根因：FB-17 已将源码插件顶层声明期入口收敛为`pluginhost.Declarations`，但其声明期子 facade 仍命名为`SourcePluginAssets`、`SourcePluginLifecycle`、`SourcePluginHooks`、`SourcePluginHTTP`、`SourcePluginJobs`和`SourcePluginGovernance`。这与`pluginbridge.Declarations`下的`RouteDeclarations`、`JobDeclarations`命名方式不一致，也会继续把“声明期子能力”误读为“源码插件实体”。
- 实现：将源码插件声明期子接口重命名为`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`GovernanceDeclarations`；`pluginhost.Declarations`的`Assets()`、`Lifecycle()`、`Hooks()`、`HTTP()`、`Jobs()`和`Governance()`返回值同步更新；`sourcePlugin`内部字段、方法签名和单元测试断言同步调整。
- 保留边界：`SourcePluginDefinition`、`SourcePlugin*Input`、`SourcePlugin*Handler`、`NewSourcePluginLifecycleCallbackAdapter`和源码插件 registry 查询函数名称保持不变，因为它们表达宿主读模型、生命周期输入、回调处理器或源码插件 registry 语义，不属于声明期子 facade。
- 规范与文档：更新本变更增量规范，要求源码插件声明期子 facade 使用`*Declarations`命名，并限制`SourcePlugin*`前缀只用于非声明期 facade 契约；已审查`apps/lina-core/pkg/plugin`中英文 README，现有 README 只描述`Declarations`方法，不列子接口类型名，因此无需同步修改。
- `i18n`：仅修改 OpenSpec 文档、Go 标识符、Go 注释和测试断言；不新增或修改运行时用户可见文案、菜单、路由、按钮、表单、表格、错误消息、语言包、翻译缓存或 API 文档源文本，确认无运行时`i18n`资源影响。
- 数据权限：不新增数据读取、写入、列表、详情、导出、下载、聚合统计、下拉候选或授权关系变更；源码插件运行时能力仍通过`pluginhost.Services`、`CapabilityContext`和既有领域`*cap.Service`执行，确认无新增数据权限绕行。
- 缓存一致性：未新增缓存、快照、失效、刷新、revision 或集群一致性机制；声明期接口命名不改变插件状态、启停快照、资源同步、Jobs 调度或回调执行链路，确认无缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、迁移、软删除或时间字段逻辑，确认无数据库结构、SQL 幂等性和查询性能影响。
- 开发工具跨平台：未修改`Makefile`、脚本、`linactl`、CI、代码生成入口或跨平台执行入口；验证使用既有`go test`、`rg`和`openspec`命令，确认无开发工具跨平台影响。
- 测试策略：本次为后端插件 SDK 命名治理，无前端 UI、页面路由、表单、表格或端到端用户工作流变化，不新增 E2E；通过`pluginhost`单元测试、宿主插件服务测试、启动装配测试、源码插件后端编译、静态检索和 OpenSpec strict 校验覆盖。
- DI 来源：未新增运行期依赖 owner、构造参数、服务定位器、全局目录或临时`New()`服务图；声明期 facade 仅记录源码插件编译期注册结果，运行期服务目录仍由启动期共享`capability.Services`和`pluginhost.Services`提供。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI、SQL/DAO、缓存实现、数据权限实现、运行时`i18n`资源和开发工具新增实现影响。
- 静态检索：`rg -n '\b(SourcePluginAssets|SourcePluginLifecycle|SourcePluginHooks|SourcePluginHTTP|SourcePluginJobs|SourcePluginGovernance)\b' apps/lina-core apps/lina-plugins openspec/changes/consolidate-plugin-domain-capability-boundaries -S`仅命中 FB-18 记录中对旧名的根因说明和扫描命令本身，源码、README 和增量规范无旧声明期子接口名残留；`rg -n '\b(AssetDeclarations|LifecycleDeclarations|HookDeclarations|HTTPDeclarations|JobDeclarations|GovernanceDeclarations)\b' apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md apps/lina-core/pkg/plugin/pluginhost openspec/changes/consolidate-plugin-domain-capability-boundaries -S`仅命中新公开接口、测试断言和本变更规范记录。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginhost -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；源码插件`linapro-monitor-online`、`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-demo-source`、`linapro-monitor-server`、`linapro-monitor-loginlog`、`linapro-org-core`、`linapro-tenant-core`和`linapro-ai-core`的`GOWORK=off go test ./backend/... -count=1`均通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。`linapro-ops-demo-guard`未作为本轮验证项重跑，原因是其既有`go.mod`缺少`lina-core`本地`require/replace`映射，上一轮已确认为非本反馈引入的模块配置问题；本轮静态检索已覆盖其未引用旧声明期子接口名。

FB-18 `lina-review`反馈级审查记录：

- 范围来源：已按`lina-review`要求查看父仓库`git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change consolidate-plugin-domain-capability-boundaries --json`，并展开`apps/lina-plugins`子仓库状态和未跟踪文件；当前工作区存在大量其他活跃变更，本次审查只覆盖 FB-18 触达的`pluginhost`声明期子接口、对应单元测试、增量规范和本 OpenSpec 记录。
- 审查文件：`apps/lina-core/pkg/plugin/pluginhost/pluginhost.go`、`pluginhost_source_plugin.go`、`pluginhost_source_plugin_contract.go`、`pluginhost_source_plugin_test.go`、`openspec/changes/consolidate-plugin-domain-capability-boundaries/specs/plugin-host-domain-capabilities/spec.md`和`tasks.md`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`.agents/rules/dev-tooling.md`；后端 Go 审查按`goframe-v2`技能执行。
- 发现的问题：未发现阻塞问题；`pluginhost.Declarations`下的声明期子接口已统一为`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`GovernanceDeclarations`，命名模型与`pluginbridge.Declarations`保持一致。`SourcePluginDefinition`和`SourcePlugin*Input/Handler`继续保留源码插件读模型和回调输入语义，未被误改。
- 规则域结论：OpenSpec、架构、插件、后端 Go、文档和测试规则通过；本轮不涉及 HTTP API、SQL/DAO、前端 UI、运行时用户可见文案、翻译资源、缓存机制、数据权限路径、开发工具入口或新增 DI 依赖。已按 OpenSpec 审查要求记录`i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响判断。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginhost -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；九个具备本地`lina-core`映射的源码插件后端测试通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过；旧声明期子接口名静态检索确认源码、README 和增量规范无残留。
- 剩余风险：`linapro-ops-demo-guard`独立后端测试因既有模块依赖映射缺失未作为本轮验证项重跑，属于本反馈外的模块配置问题；本轮没有启动浏览器服务栈执行既有 E2E，因为变更不涉及前端 UI 或端到端用户路径。

FB-19执行记录：

- 根因：WASM `org`和`tenant`领域 host service 已通过`hostServices`授权控制插件是否可调用方法，但用户作用域方法直接信任 payload 中的任意`userID`，未先确认目标用户是否对当前插件调用 actor 可见。这样会让动态插件在已获`org`或`tenant`能力后，通过`ListUserDeptAssignments`、`GetUserDept*`、`ListUserTenants`、`ValidateUserInTenant`或`ValidateSwitch`探测范围外用户的组织和租户信息。
- 实现：新增`ensureHostCallUsersVisible`共享 helper，基于当前`hostCallContext`构造`CapabilityContext`，通过同一插件作用域的`Users().EnsureUsersVisible`校验目标用户；`org`批量与单用户方法、`tenant`用户租户列表/校验/切换方法均在调用领域服务前执行该校验；不可见时返回`HostCallStatusCapabilityDenied`并停止后续调用。
- 测试：新增`TestHandleHostServiceInvokeOrgRejectsInvisibleTargetUser`和`TestHandleHostServiceInvokeTenantRejectsInvisibleTargetUser`，覆盖目标用户不可见时拒绝访问，并断言`tenant`后续服务未被调用；正常路径测试同步提供用户能力替身，确认目标用户可见时原调用继续成功。
- `i18n`：仅调整后端权限边界和测试，不新增运行时 UI 文案、菜单、路由、API 文档源文本、语言包或翻译缓存；错误 payload 使用既有 host-call 结构化 fallback，不新增语言资源，确认无运行时`i18n`资源影响。
- 数据权限：有影响。目标用户可见性从调用方自律变为 host service 前置校验，读边界由`usercap.Service.EnsureUsersVisible`统一承载，任一目标不可见时整体拒绝，符合插件通过宿主发布服务访问数据时与宿主 API 等价的数据权限边界。
- 缓存一致性：未新增缓存、快照、失效、revision 或集群一致性机制；目标用户可见性仍由启动期共享的用户能力服务和其既有数据权限策略决定，确认无新增缓存一致性影响。
- 数据库/SQL：未新增或修改 SQL、DAO、DO、Entity、索引、软删除或时间字段逻辑；新增校验复用既有用户能力契约，不直接引入数据库查询路径。
- 开发工具跨平台：未修改脚本、`Makefile`、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：本次为后端内部插件桥接数据权限修复，无前端页面或端到端用户路径变化，不新增 E2E；使用 WASM host service 单元测试和 race 测试覆盖。
- DI 来源：未新增运行期依赖；可见性校验复用已由启动期注入并按插件作用域绑定的`capability.Services.Users()`共享实例，不在调用路径临时`New()`服务图。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/i18n.md`、`.agents/rules/database.md`、`.agents/rules/documentation.md`；后端 Go 实现按`goframe-v2`技能检查。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test -race ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-20执行记录：

- 根因：host-call envelope 只有`status`和`payload`，失败时`payload`直接写入字符串；WASM dispatcher 多处使用`err.Error()`构造响应，既无法稳定传递`errorCode`、`messageKey`和参数，也容易让动态插件依赖自由文本。
- 实现：在`pluginbridge/internal/hostcall`新增`HostCallErrorPayload`，失败 payload 统一编码为 JSON，包含`errorCode`、`messageKey`、`messageParams`和英文`fallback`；保留外层 envelope wire 字段不变，新增`NewHostCallErrorResponseFromError`和`UnmarshalHostCallErrorPayload`；当错误链包含`bizerr.Error`时保留其结构化元数据，否则按 host-call status 生成稳定默认错误码和消息 key。WASM 侧新增`hostCallErrorFromError`，并将 WASM 包内`err.Error()`、`callErr.Error()`、`decodeErr.Error()`和`execErr.Error()`直传 host-call 错误响应的路径统一改为传递`error`。
- 测试：新增`TestHostCallErrorPayloadIsStructured`和`TestHostCallErrorPayloadPreservesBizerrMetadata`，覆盖普通 host-call 错误不再是裸字符串，以及`bizerr`元数据可跨 bridge 保留。
- `i18n`：有影响评估但无需新增资源。错误 payload 现在具备稳定`messageKey`和英文 fallback，为后续 guest/UI 本地化提供结构化边界；本次未新增宿主 API 文档源文本、前端文案或语言包条目。
- 数据权限：间接有利于数据权限拒绝表达；不改变数据查询或写入边界，FB-19 负责实际可见性校验。
- 缓存一致性：未新增缓存或失效机制，确认无缓存一致性影响。
- 数据库/SQL：未涉及数据库结构、DAO、SQL 或索引，确认无数据库影响。
- 开发工具跨平台：未修改开发工具或脚本，确认无跨平台影响。
- 测试策略：内部协议行为变更使用`pluginbridge`协议单元测试和 WASM 包测试覆盖，不触发 E2E。
- DI 来源：未新增运行期服务依赖；协议层只依赖`bizerr`元数据提取，不引入服务定位或运行期配置。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI、SQL/DAO 和缓存实现影响。
- 静态检索：`rg "NewHostCallErrorResponse\\([^\\n]+(err|callErr|decodeErr|execErr)\\.Error\\(\\)\\)" apps/lina-core/internal/service/plugin/internal/wasm -n`无输出。
- 验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-21执行记录：

- 根因：WASM host service 的领域服务、插件配置 factory、host config service 和 manifest factory 分散在多个可变包级变量中，启动期虽然按共享实例传入，但 dispatcher 读取的是无同步保护的全局状态，测试也需要直接保存和恢复这些变量；这不符合 WASM host service 显式依赖和并发安全要求。
- 实现：新增`hostServiceRuntime`不可变快照和`atomic.Pointer`发布机制；新增`ConfigureHostServiceRuntime`一次性接收启动期共享`capability.Services`、插件配置 factory、host config service 和 manifest factory；`ConfigureWasmHostServices`改为一次发布完整快照；保留窄的`Configure*`测试/局部配置入口，但内部通过 CAS 更新快照字段；各 dispatcher 只从当前快照读取依赖。
- 测试：更新 host config、manifest、plugin config 和 domain service 测试 helper，保存/恢复运行期快照而不是直接写旧包级变量；更新`ConfigureWasmHostServicesRequiresExplicitDependencies`断言，确认统一运行期入口仍拒绝缺失依赖。
- `i18n`：不新增运行时用户可见文案、语言包、API 文档源文本或翻译缓存；确认无`i18n`资源影响。
- 数据权限：不改变数据权限语义；运行期快照确保用户、组织、租户等能力继续复用启动期共享服务实例，避免调用路径临时构造绕过数据权限状态。
- 缓存一致性：有影响评估。该改造不新增缓存，但收敛了持有运行期状态和快照的服务实例读取方式，动态插件 host service 继续复用启动期共享后端和缓存敏感服务，避免每次调用临时创建仅当前节点可见的默认实例。
- 数据库/SQL：未涉及 SQL、DAO、DO、Entity、索引、软删除或时间字段，确认无数据库影响。
- 开发工具跨平台：未修改脚本、`Makefile`、`linactl`、CI 或代码生成入口，确认无开发工具跨平台影响。
- 测试策略：后端运行期依赖装配变更使用包级 Go 测试、启动绑定包测试和 race 测试覆盖，不触发 E2E。
- DI 来源：运行期依赖 owner 为`internal/service/plugin.ConfigureWasmHostServices`调用方的 HTTP 启动装配；创建位置仍是宿主启动期共享服务图；传递路径为启动装配 → `plugin.ConfigureWasmHostServices` → `wasm.ConfigureHostServiceRuntime` → `hostServiceRuntime`原子快照；所有字段均复用启动期共享实例或 factory，无调用路径临时`New()`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`；后端 Go 实现按`goframe-v2`技能检查。确认无 HTTP API、前端 UI 和 SQL/DAO 影响。
- 验证：`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`cd apps/lina-core && go test -race ./internal/service/plugin/internal/wasm -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。

FB-22执行记录：

- 根因：`TestHandleHostServiceInvokeStorageConcurrentDispatchIsRaceSafe`会并发触发`capability.ServicesForPlugin`，测试替身`capabilityHostServiceTestServices.ForPlugin()`在共享父对象上写入`lastPlugin`，其他断言读取同一字段，导致 race 测试失败。该竞争属于测试替身共享状态问题，但会掩盖真实 dispatcher race。
- 实现：将测试替身的插件作用域记录改为`capabilityHostServiceScopeRecorder`，用互斥锁保护共享记录；`ForPlugin()`不再修改父对象字段，而是在返回的 scoped copy 上写入不可变`scopedPlugin`用于 AI 子服务注入；相关断言改为读取 recorder。并发测试不再读写未保护字段。
- `i18n`：仅修改测试 helper 和断言，不涉及运行时文案或语言资源，确认无`i18n`影响。
- 数据权限：不改变生产数据权限逻辑；测试 helper 变化只保证并发验证可信。
- 缓存一致性：不改变缓存、快照或集群一致性机制，确认无缓存影响。
- 数据库/SQL：不涉及 SQL 或 DAO，确认无数据库影响。
- 开发工具跨平台：不涉及脚本或工具入口，确认无开发工具跨平台影响。
- 测试策略：该任务本身是测试治理修复，使用`go test -race ./internal/service/plugin/internal/wasm -count=1`作为主要验证；相关普通 Go 测试同步通过。
- DI 来源：未新增生产运行期依赖；测试替身新增 recorder 仅作用于单元测试内部，不进入启动装配或生产调用路径。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/plugin.md`、`.agents/rules/i18n.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`；后端 Go 测试按`goframe-v2`技能检查。
- 验证：`cd apps/lina-core && go test -race ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/... -count=1`通过；`openspec validate consolidate-plugin-domain-capability-boundaries --strict`通过。
