# service-dependency-injection-governance Specification

## Purpose
TBD - created by archiving change explicit-service-dependency-injection. Update Purpose after archive.
## Requirements
### Requirement: 后端组件必须通过显式依赖注入管理运行期依赖
系统 SHALL 要求宿主和源码插件的生产后端组件通过构造函数参数逐项显式接收运行期依赖。Controller、Middleware、Service、插件宿主服务适配器和 WASM host service MUST 不在业务构造函数、请求处理、插件回调或 host service 调用路径中隐式创建关键服务依赖，MUST NOT 通过聚合依赖结构体整体传递多个接口型运行期依赖。

#### Scenario: 服务构造函数逐项接收接口依赖
- **WHEN** 宿主服务需要访问配置、插件、权限、租户、会话、缓存协调或 i18n 等运行期依赖
- **THEN** 构造函数在签名中逐项接收这些接口型依赖
- **AND** 构造函数不得在依赖缺失时静默调用其他关键服务的 `New()` 补齐依赖

#### Scenario: 禁止聚合结构体隐藏接口依赖
- **WHEN** 后端组件需要接收多个接口对象、服务对象或宿主能力适配器
- **THEN** 这些接口型依赖必须拆分为独立构造函数参数
- **AND** 不得通过 `Dependencies`、`Deps`、`Options` 或等价聚合结构体整体传递
- **AND** 依赖新增、删除或替换必须能通过 Go 编译错误暴露所有未同步调用点

#### Scenario: 控制器构造函数接收服务依赖
- **WHEN** 宿主或源码插件控制器依赖一个或多个服务组件
- **THEN** 控制器构造函数通过参数接收这些服务实例
- **AND** 控制器构造函数不得自行创建缓存敏感或运行期状态敏感服务实例

#### Scenario: 请求路径不得临时创建关键服务
- **WHEN** HTTP handler、中间件、插件回调或 WASM host service 正在处理一次运行期调用
- **THEN** 该路径复用构造时注入的依赖
- **AND** 该路径不得临时调用关键服务 `New()` 创建新的服务图

### Requirement: 不得通过通用容器或全局 service locator 规避显式依赖
系统 SHALL 在不引入通用 DI 容器、全局 service locator 或新增宿主私有组装层的前提下完成依赖管理。启动期已有编排、路由绑定和插件 registrar SHALL 作为显式构造边界。

#### Scenario: 启动编排持有共享实例
- **WHEN** HTTP runtime 构造宿主长生命周期服务
- **THEN** 这些服务由现有启动编排结构持有并向路由绑定、插件注册和 host service 配置传递
- **AND** 业务组件不得通过全局 registry 在运行期查询依赖

#### Scenario: 禁止新增通用 DI 容器
- **WHEN** 开发者为后端依赖管理设计方案
- **THEN** 方案不得引入第三方或自研通用 DI 容器
- **AND** 依赖关系必须保持 Go 类型签名可见

### Requirement: 缓存敏感组件必须共享运行期实例或共享后端
系统 SHALL 对所有持有缓存、派生状态、失效观察状态、session/token 状态、插件运行时状态、运行时配置快照、权限快照或跨实例协调依赖的组件强制共享同一运行期实例或同一共享后端。

#### Scenario: 中间件复用认证和权限服务实例
- **WHEN** 宿主认证、租户和权限中间件被构造
- **THEN** 中间件接收启动期构造的 `auth`、`role`、`tenant`、`config`、`i18n`、`bizctx` 和 `plugin` 依赖
- **AND** 中间件不得自行创建另一套认证、权限、租户或插件服务图

#### Scenario: 插件管理和插件运行时复用同一插件服务
- **WHEN** 插件管理控制器、插件 HTTP route dispatcher、插件 runtime cache、source route registrar 或动态插件 host service 需要插件治理状态
- **THEN** 它们复用启动期同一个插件服务实例或该实例发布的窄接口
- **AND** 不得创建会持有独立 enabled snapshot、route binding、frontend bundle、runtime i18n 或 revision observer 的插件服务实例

#### Scenario: 缓存协调后端在集群模式下保持一致
- **WHEN** `cluster.enabled=true` 且组件需要 cachecoord、kvcache、lock、session hot state 或 token state
- **THEN** 该组件使用启动期注入的 coordination-backed 服务或同一共享 coordination 后端
- **AND** 不得退回到仅当前节点可见的本地默认实例

### Requirement: 源码插件必须通过宿主发布依赖获取宿主能力
系统 SHALL 通过源码插件 registrar 或等价宿主发布上下文向源码插件提供稳定的宿主服务目录。源码插件 Controller 和 Service MUST 通过该目录接收宿主能力适配器，不得在插件生产路径中自行构造宿主内部服务图。

#### Scenario: 源码插件注册 HTTP 路由
- **WHEN** 源码插件在 `http.route.register` 回调中构造控制器和服务
- **THEN** 插件从 registrar 暴露的宿主服务目录获取 `bizctx`、`config`、`i18n`、`notify`、`auth`、`session`、`pluginstate` 等宿主能力
- **AND** 插件业务服务通过显式依赖接收这些能力

#### Scenario: 插件宿主服务适配器由宿主构造
- **WHEN** 源码插件需要使用 `pkg/pluginservice/*` 发布的宿主能力
- **THEN** 适配器实例由宿主运行期构造并通过 registrar 传递
- **AND** 插件生产路径不得调用无参 adapter 构造函数创建孤立宿主服务图

### Requirement: 初始化与注册 API 必须返回错误给调用方决策
系统 SHALL 要求宿主和源码插件的运行时初始化、源码插件注册、registrar、回调注册、路由注册、Cron 注册和中间件注册 API 在依赖缺失、注册参数非法、配置来源缺失、后端创建失败或校验失败时返回 `error`。这些 API MUST NOT 在内部直接 `panic` 处理可预期错误；是否中止进程、忽略或降级必须由调用栈最上层入口显式决定。

#### Scenario: 源码插件注册 API 返回错误
- **WHEN** 源码插件声明无效 extension point、无效执行模式、nil callback 或重复注册
- **THEN** `pluginhost` 注册 API 返回 `error`
- **AND** API 内部不得直接 `panic`

#### Scenario: 顶层静态注册入口选择失败退出
- **WHEN** 源码插件包级 `init` 调用注册 API 收到错误
- **THEN** 该顶层静态注册入口可以显式 `panic`
- **AND** panic 治理扫描 MUST 将该调用识别为顶层入口收到错误后的失败退出
- **AND** 识别方式可以是宿主精确 allowlist 条目，也可以是官方插件工作区对 `backend/plugin.go` `init` 注册 fail-fast 模式的自动归类
- **AND** 官方插件集合变化时，不得要求维护按插件 ID 枚举的宿主 allowlist 清单

#### Scenario: 运行期回调缺少宿主依赖
- **WHEN** HTTP、Cron、Hook 或中间件注册回调在执行期发现宿主发布依赖缺失
- **THEN** 回调返回 `error`
- **AND** 宿主调用方决定阻断启动、记录失败或执行其他降级策略

### Requirement: 依赖注入规则必须纳入项目规范和 lina-review 审查
系统 SHALL 将显式依赖注入、隐式构造禁止、初始化/注册错误返回和缓存敏感共享实例要求写入项目规范与 `lina-review` 审查标准。审查 MUST 覆盖宿主、源码插件、插件 host service、WASM host service 和测试验证。

#### Scenario: 审查后端实现变更
- **WHEN** `lina-review` 审查任何后端 Go 变更
- **THEN** 审查检查新增或修改的组件是否通过显式依赖注入管理运行期依赖
- **AND** 审查标记生产路径中的关键服务隐式构造

#### Scenario: 审查初始化和注册错误处理
- **WHEN** `lina-review` 审查运行时初始化、源码插件注册、registrar、回调注册或启动装配变更
- **THEN** 审查确认可预期失败通过 `error` 返回给调用方
- **AND** 审查标记 API 内部直接 `panic` 处理可预期错误的实现

#### Scenario: 审查聚合接口依赖结构体
- **WHEN** `lina-review` 审查构造函数或依赖注入设计
- **THEN** 审查标记通过聚合结构体整体传递多个接口型运行期依赖的实现
- **AND** 审查要求将接口型依赖拆分为独立构造函数参数

#### Scenario: 审查缓存敏感组件
- **WHEN** `lina-review` 审查涉及认证、权限、session、插件、配置、i18n、cachecoord、kvcache、lock、notify 或 host service 的变更
- **THEN** 审查要求说明共享实例或共享后端如何保证状态一致
- **AND** 若变更无缓存影响，审查结论必须明确说明

#### Scenario: 静态扫描阻止回归
- **WHEN** 变更完成验证
- **THEN** 项目执行静态扫描或等价治理验证，识别非测试、非启动构造边界中对关键服务 `New()` 的调用
- **AND** 任何新增违规调用必须修复或记录明确豁免理由

### Requirement: 插件服务边界收敛必须保持启动期显式依赖注入

系统 SHALL 在收敛插件服务边界时继续使用启动期显式依赖注入。`plugin`根包新增的任何 facade 构造入口 MUST 逐项接收接口型运行期依赖并委托内部子组件，不得通过聚合依赖结构体、全局 service locator、隐式`New()`或包级默认实例补齐`auth`、`session`、`plugin`、`i18n`、`cachecoord`、`kvcache`、`notify`、`orgcap`、`tenantcap`等关键依赖。

#### Scenario: NewHostServices 构造源码插件能力目录
- **WHEN** `plugin`根包提供源码插件 host services 构造 facade
- **THEN** 该 facade 的签名逐项接收所需宿主服务实例
- **AND** 它不得使用`Dependencies`、`Deps`或`Options`等聚合结构体承载多个接口型依赖
- **AND** 它不得在依赖缺失时临时创建关键服务实例

#### Scenario: ConfigureWasmHostServices 保持共享后端
- **WHEN** 宿主启动配置动态插件 WASM host service
- **THEN** 配置入口继续使用启动期传入的共享 cache、lock、notify、config、host services 和 manifest/config factory
- **AND** 迁移 hostservices 或 runtimecache 包路径不得导致 WASM host service 回退到包级默认孤立实例

### Requirement: 插件内部子组件导入边界必须可静态验证

系统 SHALL 通过静态检索、编译门禁或治理测试验证插件服务边界收敛后的导入方向。宿主启动层和插件外部调用方 MUST 只依赖`internal/service/plugin`根 facade 或`plugin/runtimecache`等明确允许的受控子包，不得直接依赖`plugin/internal/<subcomponent>`实现包。

#### Scenario: 启动层不导入 plugin internal 子组件
- **WHEN** 审查`apps/lina-core/internal/cmd`生产 Go 代码
- **THEN** 不得发现它导入`lina-core/internal/service/plugin/internal/`
- **AND** 它通过`lina-core/internal/service/plugin`根 facade 获取插件服务、host services 构造和 WASM host service 配置入口

#### Scenario: 旧独立插件 service 包无生产导入
- **WHEN** 审查迁移后的生产 Go 代码
- **THEN** 不得发现生产代码 import `lina-core/internal/service/pluginhostservices`
- **AND** 不得发现生产代码 import `lina-core/internal/service/pluginruntimecache`

#### Scenario: 子组件不扩大导出面规避循环依赖
- **WHEN** 插件内部实现迁入`plugin/internal/<subcomponent>`
- **THEN** 子组件只导出父组件或授权边界内调用所需的窄契约
- **AND** 不得为了测试便利、临时复用或规避循环依赖暴露缓存快照、DAO、DO、Entity、私有配置或运行时状态结构

### Requirement: Capability Provider Manager 必须由宿主显式持有

系统 SHALL 要求框架 capability provider manager 由宿主启动装配或插件能力宿主装配层创建、持有并通过构造函数显式注入。能力包父包或 SPI 子包 MUST NOT 通过包级`defaultManager`、全局 service locator、隐式`New()`或旧`Provide()`函数保存 provider factory 注册表。

#### Scenario: 宿主创建租户 Provider Manager

- **WHEN** 宿主构造租户 capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[tenantspi.ProviderEnv]`实例
- **AND** 租户 service 构造函数显式接收该 manager
- **AND** `tenantcap`或`tenantspi`包级作用域不存在`defaultManager`

#### Scenario: 宿主创建组织 Provider Manager

- **WHEN** 宿主构造组织 capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[orgspi.ProviderEnv]`实例
- **AND** 组织 service 构造函数显式接收该 manager
- **AND** `orgcap`或`orgspi`包级作用域不存在`defaultManager`

#### Scenario: 宿主创建 AI 文本 Provider Manager

- **WHEN** 宿主构造文本`AI` capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[aitext.ProviderEnv]`实例
- **AND** 文本`AI`service 构造函数显式接收该 manager
- **AND** `aitext`包级作用域不存在`defaultManager`

### Requirement: Source Plugin Provider 声明必须进入 registrar 生命周期

系统 SHALL 要求源码插件 provider factory 声明通过`pluginhost.Declarations`进入源码插件 registrar 生命周期。宿主 MUST 能从`SourcePluginDefinition`读取 provider factory 声明，并在插件能力宿主装配阶段注册到共享 provider manager；注册 API MUST 返回`error`给调用方决策，不得在可预期失败时 panic。

#### Scenario: 插件声明 Provider Factory

- **WHEN** 源码插件在`backend/plugin.go`中声明组织、租户或文本`AI`provider factory
- **THEN** 它调用`pluginhost.Declarations`提供的强类型 provider 声明方法
- **AND** 声明方法校验 nil factory、重复声明或非法插件 ID 时返回`error`
- **AND** 插件入口自行决定是否在顶层注册失败时 panic

#### Scenario: 宿主注册 Provider Factory 到共享 Manager

- **WHEN** 宿主读取一个源码插件定义中的 provider factory 声明
- **THEN** 宿主将该 factory 与声明插件 ID 注册到对应共享 manager
- **AND** provider 使用路径继续通过插件 enabled snapshot 判断该插件是否可用
- **AND** 宿主不得在业务请求路径临时创建新的 provider manager

#### Scenario: DI 来源检查覆盖 Provider Manager

- **WHEN** OpenSpec 任务完成 provider manager 迁移
- **THEN** 任务记录必须说明 manager 的 owner、创建位置、传递路径、共享实例或共享后端策略
- **AND** 若没有新增缓存、数据权限或运行期依赖语义变化，也必须记录无影响判断

### Requirement: 动态领域能力配置必须复用启动期共享服务目录

系统 SHALL 要求动态插件普通领域能力配置复用启动期构造的同一个`capability.Services`目录。`ConfigureWasmHostServices` MUST 逐项接收并传递启动期共享依赖，MUST NOT 为普通领域能力创建新的服务图、领域专用全局目录、通用 service locator 或聚合依赖结构体。

#### Scenario: ConfigureWasmHostServices 配置普通领域能力

- **WHEN** 宿主启动配置动态插件`WASM host service`
- **THEN** 普通领域能力通过启动期共享的`capability.Services`实例一次性注入
- **AND** `WASM`分发层不得为`AI`、`User`、`Org`、`Tenant`或其他普通领域维护第二个共享实例来源

#### Scenario: 缺失领域能力目录

- **WHEN** `ConfigureWasmHostServices`收到`nil`领域能力目录
- **THEN** 配置入口必须返回错误
- **AND** 不得用包级默认实例、空实现或临时`New()`补齐运行期依赖

#### Scenario: 测试配置动态领域能力

- **WHEN** 单元测试需要验证动态领域 host service 分发
- **THEN** 测试必须构造自包含的`capability.Services`替身并调用`ConfigureDomainHostServices`
- **AND** 涉及包级状态时必须保存原值并通过`t.Cleanup`恢复

### Requirement: 缓存敏感服务后端选择必须来自启动期显式装配

系统 SHALL 要求宿主启动期根据拓扑显式创建缓存敏感服务的共享实例或共享后端。生产路径 MUST NOT 依赖包级默认 provider、进程级可变默认值或构造函数隐式 fallback 来决定 `kvcache`、插件缓存、WASM cache host service 或源码插件缓存 facade 的后端类型。

#### Scenario: HTTP 启动期创建共享 kvcache 服务

- **WHEN** 宿主 HTTP runtime 初始化共享 `kvcache.Service`
- **THEN** 启动装配根据 `cluster.enabled` 和 coordination 初始化结果显式选择 provider
- **AND** 使用该 provider 创建一个共享 `kvcache.Service`
- **AND** 后续插件 host service、源码插件缓存 facade 和 WASM cache dispatcher 复用该共享实例

#### Scenario: 生产路径不依赖默认 provider 选择后端

- **WHEN** 审查 HTTP 启动装配、插件 host service 配置或 WASM host service 配置
- **THEN** 不得发现通过进程级默认 provider 隐式选择 `kvcache` 后端的生产接线
- **AND** 测试 helper 若使用默认 provider 必须保存并恢复全局状态

### Requirement: 消费方专用窄接口必须基于复杂度收益定义

系统 SHALL 要求仅服务特定消费场景且不能由既有稳定契约清晰表达的窄依赖接口定义在消费者包内，而不是由生产者组件预先导出。生产者组件 SHOULD 默认只暴露自身完整`Service`契约、稳定产品契约或受治理的公开能力契约。消费方 SHOULD 优先复用目标组件已有`Service`或稳定契约；当该契约已经能满足消费方需要，且不会引入不稳定实现细节或额外运行期依赖时，不应为了单个方法重复定义本地窄接口。

#### Scenario: 插件 host config 只读取原始配置值

- **WHEN** 插件 host config 或 plugin config 适配器只需要读取原始宿主配置值
- **THEN** 读取接口由该插件消费方包定义
- **AND** `config`生产者组件不为该消费场景导出额外窄接口

#### Scenario: 插件能力宿主只需要认证或权限子集

- **WHEN** 插件能力宿主适配器只需要认证服务的会话撤销、租户令牌签发或角色服务的访问快照方法
- **THEN** 该窄接口由插件能力宿主或插件 service facade 包定义
- **AND** `auth`和`role`生产者组件不为该消费场景导出专用接口类型

#### Scenario: 既有 Service 已能清晰满足消费者

- **WHEN** 控制器或普通 service 只消费目标组件已有`Service`中的方法
- **AND** 直接依赖该`Service`不会暴露被调用模块的内部实现、缓存快照或私有模型
- **THEN** 消费方应直接依赖目标组件`Service`
- **AND** 不应再为同一目标组件方法重复声明只转述方法签名的本地窄接口

#### Scenario: 分类接口只用于生产者 Service 自组合

- **WHEN** 一个导出接口只被同包默认`Service`内嵌，没有独立跨包替换、测试或稳定产品契约用途
- **THEN** 该接口的方法应直接声明在默认`Service`中
- **AND** 不应保留仅用于分类命名的导出接口

#### Scenario: 稳定产品契约保留在能力 owner 包

- **WHEN** 一个接口表达稳定运行期产品契约、公共插件协议、共享状态后端或明确的跨组件能力 owner
- **THEN** 该接口可以继续由能力 owner 包定义
- **AND** 审查必须确认它不是单一消费者的临时窄依赖

### Requirement: 插件服务消费者必须依赖最小稳定契约

系统 SHALL 要求宿主内部消费者通过构造函数接收启动期共享插件服务实例发布的最小稳定契约。普通消费者只需要启动、运行时、job、状态、provider env 或租户生命周期等部分能力时，MUST 依赖消费者本地私有窄接口；若消费者是插件管理控制器、HTTP 启动上下文、`RuntimeDelegate`或其他必须跨启动阶段传递统一插件服务入口的边界，MUST 只引用导出的`plugin.Service`。消费者 MUST NOT 引用`plugin`包导出的 facet 类型，`plugin`包拆分出来的 facet MUST 保持私有。

#### Scenario: 定时任务组件依赖插件 job 能力

- **WHEN** `cron`或`jobhandler`需要同步插件声明的定时任务
- **THEN** 它们通过构造函数接收插件 job 和生命周期 observer 所需的最小契约
- **AND** 测试替身只需要实现这些实际使用的方法
- **AND** 该最小契约在消费者包内私有声明，不依赖`plugin`包导出的 facet 类型

#### Scenario: API 文档组件依赖插件路由能力

- **WHEN** `apidoc`需要读取源码插件 route binding 或投影动态路由
- **THEN** 它依赖插件路由和状态读取所需的窄契约
- **AND** 不得依赖插件安装、卸载、上传或租户 lifecycle 等无关方法
- **AND** 该窄契约在消费者包内私有声明，不依赖`plugin`包导出的 facet 类型

#### Scenario: Runtime delegate 依赖插件状态和 hook 能力

- **WHEN** 启动期 cycle breaker 在插件服务完成构造前被传入认证、菜单、角色或 capability provider
- **THEN** delegate 通过`plugin.Service`绑定启动期共享插件服务实例
- **AND** delegate 绑定后只调用 hook、菜单过滤、状态读取、provider env 和租户生命周期等当前需要的方法
- **AND** 不得为该单一 cycle breaker 额外保留`runtimeDelegateService`等重复窄接口

#### Scenario: 共享实例语义保持不变

- **WHEN** 消费者从完整`plugin.Service`迁移到 facet 或窄接口
- **THEN** 注入对象仍然来自启动期构造的同一个插件服务实例
- **AND** 不得创建新的插件服务实例、缓存快照、route binding 状态或 runtime frontend bundle 缓存

#### Scenario: HTTP 启动上下文保存插件服务

- **WHEN** HTTP 启动装配需要在多个启动阶段、路由绑定 helper 或 controller 构造中复用插件服务
- **THEN** `httpRuntime`只保存单个`pluginSvc pluginsvc.Service`字段作为统一入口
- **AND** 不得在`httpRuntime`中为管理、启动、运行时 HTTP、集成或 job facet 增加多个插件服务字段
- **AND** 路由、frontend asset 和 hook helper 从`httpRuntime`接收插件服务时应继续使用该统一`pluginsvc.Service`入口，不得额外声明插件 runtime HTTP 或 integration 分类接口

#### Scenario: 插件管理控制器使用统一插件服务入口

- **WHEN** 插件管理控制器通过`NewV1`接收插件服务依赖
- **THEN** 构造函数和控制器字段应直接使用`pluginsvc.Service`统一入口
- **AND** 不得在 controller 包中额外声明重复的插件 management 分类接口
- **AND** 单元测试替身可以嵌入`pluginsvc.Service`并只覆盖测试路径实际调用的方法

