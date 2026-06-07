## 1. 规则、范围和旧入口清单

- [x] 1.1 实施前重新读取`AGENTS.md`和命中的规则文件，至少覆盖`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`testing`，并评估`i18n`、缓存一致性、数据权限、SQL、HTTP API、前端 UI 和开发工具跨平台影响。
- [x] 1.2 使用`goframe-v2`技能作为 Go 后端实现约束，确认本变更不修改脚手架生成文件、`DAO/DO/Entity`、HTTP API DTO 或控制器骨架。
- [x] 1.3 静态扫描`capability/contract`内所有类型和函数，按公共原语、具体能力服务、factory、内部 helper、测试 helper 分类，形成迁移映射。
- [x] 1.4 静态扫描生产代码和测试代码中对`capability/contract`、旧非`*cap`包和`contract.*Service`的导入及方法签名使用，确认宿主、官方插件、guest SDK、`WASM`handler 和测试替身迁移范围。
- [x] 1.5 修改任一`apps/lina-plugins/<plugin-id>/`文件前，检查该插件根目录是否存在`AGENTS.md`普通文件或符号链接；存在时先读取并遵守插件本地规范。

## 2. 公共原语和领域组件拆分

- [x] 2.1 新增公共原语包，迁移`CapabilityContext`、`DomainID`、`CapabilityActor`、`CapabilityAuthorizationSnapshot`、`BatchResult`、`PageRequest`、`PageResult`、`LocalizedLabel`、`ProviderStatus`和`CapabilityStatus`等跨领域值对象。
- [x] 2.2 确认公共原语包不定义具体能力`Service`、`AdminService`、factory、provider adapter 或 host service handler，并补充静态检索或测试验证。
- [x] 2.3 将`APIDocService`迁移到`apidoccap`组件，保持 API 文档本地化能力语义不变。
- [x] 2.4 将`AuthService`迁移到`authcap`组件，保持租户 token handoff 和 impersonation token 语义不变。
- [x] 2.5 将`BizCtxService`、`CurrentContext`和上下文 helper 迁移到`bizctxcap`组件或公共原语边界，保持当前请求投影语义不变。
- [x] 2.6 将`CacheService`和`CacheItem`迁移到`cachecap`组件，继续绑定插件作用域并复用启动期共享缓存后端。
- [x] 2.7 将插件静态配置读取能力从`capability/config`和`contract.ConfigService`迁移到`plugincap`子领域，目标入口为`Services.Plugins().Config()`。
- [x] 2.8 将`HostConfigService`迁移到`hostconfigcap`组件，将`I18nService`迁移到`i18ncap`组件，将`ManifestService`迁移到`manifestcap`组件。
- [x] 2.9 将`PluginLifecycleService`、`PluginLifecycleRunner`和`PluginStateService`迁移到`plugincap`子领域，将`RouteService`迁移到`routecap`组件。
- [x] 2.10 将`tenantfilter`迁移到`tenantcap.PluginTableFilterService`源码插件专用接口，不新增`tenantfiltercap`组件，并确认该能力只通过`pluginhost.Services.TenantFilter()`等源码插件专用受控接缝暴露，不进入普通`capability.Services.Tenant()`。
- [x] 2.11 将`ai`能力聚合包重命名为`aicap`，同步更新`aiaudio`、`aicommon`、`aidocument`、`aiembedding`、`aiimage`、`aisafety`、`aitext`、`aivideo`和`aivision`子能力导入路径。

## 3. 服务目录、适配层和动态协议迁移

- [x] 3.1 调整`capability.Services`，使普通能力方法全部返回目标领域命名空间、`*cap.Service`或等价窄接口，不再返回`contract.*Service`；方法名按领域语义命名，资源集合领域保留`Users()`、`Jobs()`、`Plugins()`等复数入口，包名继续使用单数`*cap`。
- [x] 3.2 调整`capability.AdminServices`、`pluginhost.Services`和`ServicesForPlugin`相关类型，保持源码插件普通消费面、管理面和插件作用域绑定语义不变；`pluginhost.Services.TenantFilter()`返回`tenantcap.PluginTableFilterService`且不进入普通`capability.Services`。
- [x] 3.3 删除根`Config()`、`PluginConfig()`、`PluginLifecycle()`和`PluginState()`入口，确认插件自身配置通过`Services.Plugins().Config()`访问，宿主配置通过`Services.HostConfig()`访问。
- [x] 3.4 调整`apps/lina-core/internal/service/plugin/internal/hostservices`字段、构造函数、directory 方法和 scoped directory 方法，全部适配到新`*cap`组件和`Plugins()`子领域。
- [x] 3.5 调整`apps/lina-core/internal/service/plugin/internal/wasm`中 host service 配置入口和 handler，使动态调用进入新`*cap.Service`、`plugincap`子服务或`*cap.AdminService`，并保持`service + method`协议字符串不变；确认动态插件不暴露`tenantcap.PluginTableFilterService`、`*gdb.Model`、SQL 片段或 query builder，租户过滤由 handler 在宿主边界隐式执行。
- [x] 3.6 调整`pkg/plugin/capability/guest`目录和动态插件 guest SDK 入口，保持 host call envelope、授权快照、资源校验和错误 envelope 不变。
- [x] 3.7 调整`pkg/plugin/pluginbridge`中引用能力 DTO 或公共原语的编解码代码，只迁移 Go 导入路径，不改变 protobuf 字段或动态协议语义。
- [x] 3.8 调整`internal/cmd`启动装配和宿主插件运行时测试替身，确认缓存敏感服务继续复用启动期共享实例或共享后端。

## 4. 官方插件和测试替身迁移

- [x] 4.1 迁移官方源码插件生产代码中对`capability/contract`和旧非`*cap`包的导入，改为目标`*cap`包或公共原语包。
- [x] 4.2 迁移官方源码插件测试代码、fixture 和 fake service，确保它们实现新的`capability.Services`和窄领域接口。
- [x] 4.3 迁移动态插件样例、guest smoke fixture、`linactl`测试或 CI fixture 中的旧能力包导入。
- [x] 4.4 删除旧`capability/contract`具体服务聚合内容和旧非`*cap`公开能力包，不保留类型别名、转发函数或兼容空壳包。
- [x] 4.5 运行静态检索确认生产代码不再导入旧`capability/contract`具体能力服务、旧非`*cap`具体能力包或新增`tenantfiltercap`包，并确认源码插件租户过滤只依赖`tenantcap.PluginTableFilterService`。

## 5. 验证、影响记录和审查

- [x] 5.1 运行覆盖变更包的 Go 编译门禁，至少覆盖`apps/lina-core/pkg/plugin/capability/...`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge/...`、`internal/service/plugin/internal/hostservices`、`internal/service/plugin/internal/wasm`和`internal/cmd`。
- [x] 5.2 运行受影响官方源码插件后端 Go 测试；若插件为独立模块，使用其既有`GOWORK=off go test ./backend/... -count=1`或等价命令。
- [x] 5.3 运行静态检索或治理扫描，确认`contract.*Service`具体能力引用、旧非`*cap`能力包导入、旧`capability/ai`生产入口、根`Config()`、根`PluginConfig()`、根`PluginLifecycle()`、根`PluginState()`、独立`tenantfiltercap`和旧公共能力兼容层均无残留。
- [x] 5.4 记录 DI 来源检查，说明本变更未新增运行期依赖 owner；迁移后的 host service 适配器继续复用启动期共享实例或共享后端。
- [x] 5.5 记录影响分析：`i18n`、数据权限、缓存一致性、SQL、HTTP API、前端 UI、开发工具跨平台和 E2E 无行为影响；如实现中出现实际影响，按对应规则补充实现和验证。
- [x] 5.6 运行`openspec validate decouple-plugin-capability-contracts --strict`和变更范围`git diff --check`。
- [x] 5.7 完成实现和验证后调用`lina-review`，审查 OpenSpec 任务状态、宿主边界、插件边界、导入路径、公共原语最小化、动态协议不变、DI 来源和测试覆盖。

## 6. 执行记录

- 规则读取：本轮实现和审查前已重新读取`AGENTS.md`以及`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`api-contract`、`data-permission`、`cache-consistency`、`database`、`testing`、`i18n`、`frontend-ui`、`dev-tooling`规则；使用`goframe-v2`、`openspec-apply-change`、`karpathy-guidelines`和`lina-review`技能。
- 插件本地规范：执行`find apps/lina-plugins -maxdepth 2 -name AGENTS.md -print`未发现插件根目录`AGENTS.md`普通文件或符号链接，因此本变更内插件目录修改继续按顶层规范和命中规则执行。
- DI 来源检查：本变更未新增运行期依赖 owner；`hostservices`迁移后的领域适配器继续复用宿主启动期构造并传入的共享服务或共享后端，包括插件配置 factory、插件状态、插件生命周期、manifest 资源解析和`kvcache`后端。新增的 manifest 源码插件嵌入文件解析只通过既有`pluginhost`源码插件注册表读取当前`pluginID`的 embedded files，不创建新的服务图。
- 影响分析：无 HTTP API、DTO、OpenAPI 元数据、SQL、DAO/DO/Entity、前端页面、运行时用户文案、菜单、语言包或 E2E 用户路径变更；动态插件`service + method`协议字符串、授权快照和错误 envelope 保持不变。数据权限和租户隔离仍由既有 host service handler 与普通`tenantcap.Service`边界执行，源码插件专用`tenantcap.PluginTableFilterService`未进入普通`capability.Services.Tenant()`或动态插件 guest SDK。缓存一致性无新增失效或刷新机制，缓存敏感服务继续复用启动期共享实例或共享后端。开发工具跨平台入口无新增默认脚本；仅迁移测试和治理扫描中的导入路径。
- 验证命令：已运行覆盖`apps/lina-core/pkg/plugin/capability/...`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge/...`、`internal/service/plugin/internal/hostservices`、`internal/service/plugin/internal/wasm`和`internal/cmd`的 Go 测试；已使用临时`go.work`运行官方插件`./backend/...`测试；已运行旧包路径、旧根入口、`tenantfiltercap`和动态协议误改静态检索；已运行`openspec validate decouple-plugin-capability-contracts --strict`、根仓库`git diff --check`和`apps/lina-plugins`子仓库`git diff --check`。
- 替代验证说明：`apps/lina-plugins`聚合目录直接使用`GOWORK=off go test ./...`会命中当前工作区既有聚合`go.sum`缺少`github.com/redis/go-redis/v9`的问题。为避免把无关模块元数据变更混入本迭代，官方插件后端验证使用包含`apps/lina-core`和各插件模块的临时`go.work`执行逐插件`go test ./backend/... -count=1`。
- `lina-review`结论：未发现阻塞问题。审查过程中发现并修正了迁移后`aicap`、`bizctxcap`、`hostconfigcap`、`manifestcap`和`plugincap`部分文件顶部注释残留旧包名的问题；修正后重新运行注释静态扫描、OpenSpec 严格校验、diff 检查和相关 Go 测试均通过。

## Feedback

- [x] **FB-1**: 将认证 token 与授权能力收敛到`authcap`能力族子领域，避免根`Services`同时暴露`Auth()`和`Authz()`
- [x] **FB-2**: 将动态插件受治理数据 SDK 从`capability/data`重命名为`capability/recordstore`
- [x] **FB-3**: 将过轻`pkg/authtoken`公共包合并到`authcap/token`子领域
- [x] **FB-4**: 将`hostservices_domain_adapters.go`中的领域适配器实现拆分到各自`*cap`内部组件

### FB-1 执行记录

- 根因：认证 token handoff 与授权投影同属认证授权能力族，但旧结构在根`capability.Services`同时暴露`Auth()`和`Authz()`，并将 token 窄接口放在根`authcap`包、授权能力放在独立`authzcap`包，导致插件侧领域边界和维护入口发散。
- 修复：将 token 契约迁移到`pkg/plugin/capability/authcap/token`，将授权契约迁移到`pkg/plugin/capability/authcap/authz`，根`pkg/plugin/capability/authcap`只保留聚合入口`Service.Token()`、`Service.Authz()`和`AdminService.Authz()`；删除旧`authzcap`包，根`capability.Services`只暴露`Auth() authcap.Service`，根`capability.AdminServices`只暴露`Auth() authcap.AdminService`。宿主`hostservices`继续由启动期创建的 token adapter 和 authz domain adapter 组装`authcap.New(...)`，`linapro-tenant-core`改为注入`services.Auth().Token()`与`services.Auth().Authz()`窄接口，测试替身同步实现新聚合入口。
- 影响分析：无 HTTP API、DTO、OpenAPI 元数据、SQL、DAO/DO/Entity、前端页面、运行时用户文案、菜单、语言包、E2E 用户路径或开发工具跨平台入口变更。数据权限和租户隔离语义不变，授权查询和管理命令仍由原`authz`契约承载；缓存一致性无新增失效、刷新或共享状态影响；DI 无新增运行期服务 owner，聚合对象只组合既有启动期共享实例和适配器。
- 规则读取：已重新读取`AGENTS.md`以及`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`data-permission`、`cache-consistency`、`testing`、`i18n`规则；已使用`lina-feedback`、`goframe-v2`和`karpathy-guidelines`技能。
- 验证命令：`go test ./pkg/plugin/capability/... -count=1`、`go test ./internal/service/plugin/internal/hostservices -count=1`、`go test ./internal/service/plugin/internal/wasm -count=1`、`go test ./internal/service/plugin/internal/integration -count=1`、`go test ./internal/service/plugin -count=1`、`go test ./internal/cmd -count=1`、`go test ./internal/service/datascope -count=1`、`GOWORK=off go test ./backend/internal/controller/auth ./backend/internal/service/impersonate -count=1`、`openspec validate decouple-plugin-capability-contracts --strict`、根仓库`git diff --check`和`apps/lina-plugins`子仓库`git diff --check`。`GOWORK=off go test ./... -count=1`和`GOWORK=off go test ./backend -run TestNonExistent -count=1`在`linapro-tenant-core`中被既有`go.sum`缺少`github.com/redis/go-redis/v9`阻断，本次未修改插件依赖元数据；静态检索确认生产代码、官方插件和测试替身无旧`authzcap`导入、无根`services.Authz()`调用、无根`authcap` token DTO 使用残留。
- `lina-review`结论：反馈级审查未发现阻塞问题；审查范围限定为本次认证授权能力族收敛涉及的`authcap`、根`capability.Services`/`AdminServices`、`hostservices`装配、测试替身、`linapro-tenant-core`调用点和为恢复编译修正的`datascope`包名/测试引用。工作区存在大量同一 OpenSpec 下的既有改动，未作为本反馈逐项审查范围。

### FB-2 执行记录

- 根因：`pkg/plugin/capability/data`只服务动态插件受治理表记录访问，`data`命名过宽，容易被误解为通用数据层或宿主私有数据访问入口；该能力实际职责是面向动态插件的授权表记录查询、变更和事务 facade。
- 修复：将 SDK 包迁移到`apps/lina-core/pkg/plugin/capability/recordstore`，公开包名改为`recordstore`，typed plan 公开类型改为`QueryPlan`、`Filter`、`Order`、`Pagination`、`PlanAction`等 record store 本地名称；`capability/guest`目录入口由`Services.Data()`改为`Services.RecordStore()`；宿主`datahost`执行层和`linapro-demo-dynamic`样例改为导入新包并调用`RecordStore()`。
- 协议边界：本反馈只调整 Go SDK 命名和导入路径，动态插件 bridge 协议保持不变；`HostServiceData*`、`HostServiceMethodData*`、`service: data`和`host:data:*`继续作为底层 host-service 协议名称使用。
- 影响分析：无 HTTP API、DTO、OpenAPI 元数据、SQL、DAO/DO/Entity、前端页面、运行时用户文案、菜单、语言包、E2E 用户路径或开发工具跨平台入口变更。数据权限和租户隔离仍由既有 host data service handler、授权资源表和请求上下文执行，未新增数据操作语义；缓存一致性无新增失效、刷新或共享状态影响；DI 无新增运行期服务 owner。
- 插件本地规范：修改`apps/lina-plugins/linapro-demo-dynamic`前已检查该插件根目录，不存在`AGENTS.md`普通文件或符号链接。
- 规则读取：已重新读取`AGENTS.md`以及`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`testing`、`data-permission`、`i18n`规则；已使用`lina-feedback`和`goframe-v2`技能。
- 验证命令：`go test ./pkg/plugin/capability/recordstore ./pkg/plugin/capability/guest ./pkg/plugin/capability ./internal/service/plugin/internal/datahost ./pkg/plugin/pluginbridge/internal/hostservice -count=1`、`GOWORK=off go test ./backend/... -count=1`（在`apps/lina-plugins/linapro-demo-dynamic`执行）、静态检索确认旧`capability/data`导入、`Services.Data()`/`guestServices.Data()`调用和旧`Data*`公开 plan 名称无代码残留、`openspec validate decouple-plugin-capability-contracts --strict`、目标范围`git diff --check`和`apps/lina-plugins`目标范围`git diff --check`。
- `lina-review`结论：反馈级审查未发现阻塞问题；审查范围限定为本次`recordstore`重命名涉及的旧`capability/data`删除、新`capability/recordstore`SDK、`capability/guest`入口、宿主`datahost`适配、`linapro-demo-dynamic`样例调用点和 OpenSpec 记录。静态扫描仅在本执行记录中保留旧路径作为历史说明。

### FB-3 执行记录

- 根因：`pkg/authtoken`只承载 JWT token kind 和用户 session client type 契约，职责已归属`pkg/plugin/capability/authcap/token`，继续单独保留会形成过轻公共包。
- 修复：删除`apps/lina-core/pkg/authtoken`，将`KindAccess`、`KindRefresh`、`ClientType`和`ParseClientType`合并到`authcap/token`，宿主 auth service 与动态插件 route validator 均改为依赖该 token 子领域契约；auth 包仅保留错误适配，不保留`ClientType*`常量转发。
- 影响分析：无 HTTP API、DTO、OpenAPI 元数据、SQL、DAO/DO/Entity、前端页面、运行时用户文案、菜单、语言包、E2E 用户路径或开发工具跨平台入口变更。数据权限和租户隔离语义不变；动态插件 route validator 仍只校验 access token 与用户 session client type。缓存一致性无新增失效、刷新或共享状态影响；DI 无新增运行期服务 owner。
- 规则读取：已重新读取`AGENTS.md`以及`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`testing`规则；已使用`lina-feedback`、`lina-review`和`goframe-v2`技能。
- 验证命令：`go test ./internal/service/auth -count=1`、`go test ./internal/service/plugin/internal/runtime -count=1`、`go test ./pkg/plugin/capability/authcap/... -count=1`、静态检索确认`authtoken|pkg/authtoken|lina-core/pkg/authtoken`在`apps/lina-core`无代码残留且 auth 包不再定义`ClientType*`常量别名、`openspec validate decouple-plugin-capability-contracts --strict`。

### FB-4 执行记录

- 分诊：处理路径为`openspec-existing`，本反馈直接修正`decouple-plugin-capability-contracts`中`hostservices`适配层实现组织缺口；不需要更新增量规范，因为动态协议、公开`*cap`契约、数据权限、缓存一致性和用户可观察行为均不变。
- 根因：`apps/lina-core/internal/service/plugin/internal/hostservices/hostservices_domain_adapters.go`单文件聚合`usercap`、`authcap/authz`、`dictcap`、`filecap`、`sessioncap`、`configcap`、`notifycap`、`plugincap`、`jobcap`和`infracap`的具体实现，形成 1680 行跨领域实现文件。该结构让 hostservices 包同时承担装配目录和各领域数据库适配职责，不利于按能力 owner 审查数据权限、缓存失效和批量查询边界。
- 修复：删除原聚合文件，新增`hostservices/internal/{authzcap,configcap,dictcap,filecap,infracap,jobcap,notifycap,plugincap,sessioncap,usercap}`内部组件包；每个组件只实现对应公开`*cap.Service`/`AdminService`。跨领域共用的分页、领域 ID 解析、`MissingIDs`去重和共享修订号推进收敛到`hostservices/internal/domaincap`，该包不定义具体能力服务。`hostservices.New()`只负责显式装配这些组件，`directory`继续暴露原有`capability.Services`和`AdminServices`契约。
- DI 来源：未新增运行期依赖 owner，也未在业务路径临时`New()`关键服务图。新内部组件全部由`hostservices.New()`使用原启动期传入的`datascope.Service`、`tenantcap.RuntimeService`、`session.Store`、`NotifyPublisher`、`plugincap.StateService`、`PluginLifecycleRunner`、`i18n`适配器和`tenantcap.PluginTableFilterService`构造；缓存敏感能力仍复用启动期共享实例或共享后端。
- 数据权限影响：无新数据读取或写入语义。用户、文件、任务、配置和字典仍通过`tenantcap.PluginTableFilterService`在查询阶段过滤；会话仍通过`session.Store`的 scoped 查询和撤销前可见性校验；通知、插件启用、角色授权和运行时配置仍保持原有目标检查与不可见目标拒绝语义，批量读取继续用`MissingIDs`隐藏不存在和不可见差异。
- 缓存一致性影响：未新增缓存域或失效机制。授权、插件运行时、字典和运行时配置写路径继续通过共享修订号在事务内推进；拆分只改变实现归属，不改变单机或集群一致性策略。
- `i18n`影响：无运行时用户可见文案、菜单、路由、API 文档源文本、插件清单或语言包资源变更；字典 label 的可选翻译仍由`dictcap`组件通过既有`i18ncap.Service`执行。
- 开发工具跨平台影响：无脚本、`linactl`、CI 或构建入口变更；仅新增 Go 内部包和迁移 Go 测试。
- 测试策略：本次是内部实现重构，无前端 UI 或端到端用户路径变化，未触发 E2E；使用 Go 编译门禁、既有 hostservices 单元测试、新增`internal/sessioncap`转换测试、静态检索和 OpenSpec 严格校验覆盖。
- 规则读取：已重新读取`AGENTS.md`以及`openspec`、`architecture`、`plugin`、`backend-go`、`testing`、`data-permission`、`cache-consistency`、`i18n`规则；已使用`lina-feedback`、`lina-review`、`goframe-v2`和`karpathy-guidelines`技能。
- 验证命令：`cd apps/lina-core && go test ./internal/service/plugin/internal/hostservices/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/cmd -count=1`通过；`openspec validate decouple-plugin-capability-contracts --strict`通过；`git diff --check -- apps/lina-core/internal/service/plugin/internal/hostservices openspec/changes/decouple-plugin-capability-contracts/tasks.md`通过；静态检索确认旧适配器类型和旧`sessionAdapter`生产引用无残留。
