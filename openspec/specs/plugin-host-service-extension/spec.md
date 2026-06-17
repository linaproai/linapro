# 插件宿主服务扩展规范

## Purpose
待定 - 由归档变更 dynamic-plugin-host-service-extension 创建。归档后更新目的。
## Requirements
### Requirement: 动态插件通过版本化宿主服务协议获取全部宿主能力

系统 SHALL 在保留`lina_env.host_call`入口的前提下，为动态插件提供版本化的宿主服务调用协议。动态插件访问宿主能力时 MUST 通过统一宿主服务通道进入`pluginservice`能力目录或其受控适配器，而不是继续线性增加新的专用 opcode，也不得让`pluginbridge`成为与源码插件平行的宿主能力语义 owner。

#### Scenario: Guest 调用结构化宿主服务

- **WHEN** guest SDK 发起一次宿主服务调用
- **THEN** 宿主通过统一请求 envelope 解析`service`、`method`、资源标识（如`storage.resources.paths`、URL 模式、`table`、pluginservice capability ID 或 manifest 资源路径）和请求载荷
- **AND** 宿主服务注册表定位对应的服务处理器
- **AND** 服务处理器委托到`pluginservice`能力目录、`orgcap.Service`、`tenantcap.Service`或其他受控宿主适配器
- **AND** 宿主以统一响应 envelope 返回业务结果或结构化错误

#### Scenario: 未知 service 或 method 被拒绝

- **WHEN** 插件调用一个宿主不支持的`service`或`method`
- **THEN** 宿主返回显式的“不支持”错误
- **AND** 宿主不得进入任何实际宿主服务逻辑

#### Scenario: 已有最小 Host Call 被统一重构

- **WHEN** 宿主收敛当前动态插件能力模型
- **THEN** 已实现的日志、状态和数据访问能力也通过统一宿主服务协议对 guest 暴露
- **AND** 宿主不得继续维护面向插件的平行公开协议

#### Scenario: 动态插件消费 Pluginservice Capability

- **WHEN** 动态插件通过 guest SDK 调用`framework.org.v1`
- **THEN** host service handler 校验动态插件的`hostServices`授权
- **AND** 调用进入`pluginservice.Services.Org()`对应的消费 service
- **AND** 如该动态插件需要硬依赖具体 provider 插件，则由既有`dependencies.plugins`在生命周期路径中校验
- **AND** `pluginbridge`仅承担 transport 和 payload 编解码职责

### Requirement: 宿主服务访问同时受宿主服务声明推导的能力分类和资源授权约束

系统 SHALL 对每一次宿主服务调用执行由`hostServices`声明自动推导的粗粒度 capability 校验。对于资源型 host service，系统还 SHALL 执行细粒度资源授权校验；对于`ai`这类方法授权型 host service，系统 MUST 只按`service + method`授权快照校验，不得要求插件作者声明或确认额外`resources`。只读读取型服务的`methods` MUST 表达真实 host service 调用动作，SDK typed helper 不得作为独立授权方法进入声明或运行时快照。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者在动态插件清单中声明`hostServices`
- **THEN** 构建器校验 service、method、资源声明（如`storage.paths`、URL 模式、`data.tables`、宿主公开配置 key 或 manifest 资源路径）和策略参数是否合法
- **AND** 对`service: ai`只校验声明的`methods`是否合法，并拒绝`resources`、`paths`、`tables`或`keys`
- **AND** 宿主根据这些 methods 自动推导内部 capability 分类快照
- **AND** 将归一化后的宿主服务策略写入运行时产物
- **AND** 宿主装载产物后恢复为当前 release 的服务授权快照

#### Scenario: 缺少授权的宿主服务调用被拒绝

- **WHEN** 插件调用未声明的 service、method 或资源型服务的未授权资源标识
- **THEN** 宿主返回显式拒绝错误
- **AND** 宿主不执行任何目标宿主服务逻辑

#### Scenario: typed helper 不作为授权方法声明

- **WHEN** 动态插件需要读取插件配置并使用 guest SDK 的`String`、`Bool`、`Int`、`Duration`或`Scan`helper
- **THEN** `plugin.yaml`只声明`service: config`和`methods: [get]`
- **AND** 宿主授权快照只记录`config.get`
- **AND** guest SDK helper 在插件侧或共享适配层基于`get`结果完成类型转换

### Requirement: 资源型宿主服务声明属于权限申请而非自动授权

系统 SHALL 将所有资源型 hostServices 声明解释为权限申请清单，而不是插件在运行时自动拥有的资源访问权；其中`storage`当前使用`resources.paths`，`network`使用 URL 模式，`data`当前使用`resources.tables`，`hostConfig`使用`resources.keys`，`manifest`使用`resources.paths`，其余低优先级服务（`cache`、`lock`、`notify`）继续沿用逻辑`resourceRef`规划，并分别表示缓存命名空间、逻辑锁名和通知通道标识。

#### Scenario: 清单声明宿主服务资源申请

- **WHEN** 开发者在动态插件清单中声明`storage.resources.paths`、`network`的 URL 模式、`data.resources.tables`、`hostConfig.resources.keys`、`manifest.resources.paths`，或声明`cache`、`lock`、`notify`等低优先级服务的逻辑`resourceRef`
- **THEN** 这些声明表示插件申请对应宿主资源权限
- **AND** 声明本身不得直接视为运行时已授权结果

#### Scenario: 运行时只认宿主确认后的授权快照

- **WHEN** 动态插件在运行时调用一个资源型宿主服务
- **THEN** 宿主仅按安装或启用阶段确认后的授权快照执行校验
- **AND** 未被宿主确认的声明必须被拒绝
- **AND** 插件不能通过“先声明后直接调用”的方式跳过宿主授权确认

### Requirement: 宿主服务调用携带执行上下文并支持审计

系统 SHALL 为每一次宿主服务调用附带统一的执行上下文，并记录最小充分的调用审计信息。

#### Scenario: 请求型上下文调用宿主服务

- **WHEN** 动态插件在处理路由请求期间调用宿主服务
- **THEN** 宿主向服务处理器传入`pluginId`、执行来源、当前路由标识、当前用户身份快照和数据范围快照
- **AND** 需要用户上下文的服务方法可以复用这些治理信息做校验

#### Scenario: 系统型上下文调用宿主服务

- **WHEN** 动态插件在 Hook、定时任务或生命周期流程中调用宿主服务
- **THEN** 宿主向服务处理器传入无用户身份的系统型执行上下文
- **AND** 要求用户上下文的服务方法必须拒绝该调用

#### Scenario: 宿主记录宿主服务调用摘要

- **WHEN** 一次宿主服务调用结束
- **THEN** 宿主记录`pluginId`、service、method、资源标识摘要（如`path`、URL 模式、`resourceRef`或`table`）、结果状态和耗时
- **AND** 失败调用保留错误摘要用于诊断
- **AND** 宿主默认不记录敏感请求体和敏感响应体原文

### Requirement: 插件宿主服务适配器必须由宿主运行期统一构造
系统 SHALL 由宿主运行期统一构造并发布源码插件和动态插件 host service 适配器。适配器 MUST 复用启动期共享的宿主服务实例或共享后端，MUST 不在插件调用路径中自行构造孤立宿主服务图。源码插件和动态插件访问同一宿主能力时 MUST 共享`pluginservice`能力目录语义，动态插件 host service handler 只作为 transport 适配层。

#### Scenario: 源码插件使用宿主服务适配器
- **WHEN** 源码插件调用`pkg/pluginservice/*`发布的宿主能力
- **THEN** 该能力适配器由宿主运行期构造并通过 registrar 传递给插件
- **AND** 适配器复用宿主共享的 auth、session、notify、config、i18n、pluginstate、orgcap、tenantcap 或其他依赖
- **AND** 插件生产路径不得无参创建该适配器

#### Scenario: 动态插件 host service 调用共享宿主能力
- **WHEN** 动态插件通过统一 host service 协议调用 cache、lock、notify、config、runtime、storage、data 或 pluginservice capability 等宿主能力
- **THEN** host service handler 使用插件 runtime 注入的共享宿主服务或共享后端
- **AND** handler 不得在每次调用中创建独立 cache、lock、notify、config、plugin service 或 capability manager 实例

#### Scenario: WASM host service 配置入口由启动期注入
- **WHEN** 宿主启动并初始化 WASM host service
- **THEN** 启动路径显式配置 cache、lock、notify、storage、config、runtime、orgcap、tenantcap 和 pluginservice 等 host service 的共享依赖
- **AND** 包级默认实例不得在生产启动后继续作为实际运行依赖

### Requirement: hostservices 子组件不得反向依赖 plugin 根包

系统 SHALL 保证`plugin/internal/hostservices`不导入`apps/lina-core/internal/service/plugin`根包。需要读取动态路由元数据或其他插件运行时上下文时，hostservices MUST 依赖真实 owner 的窄接口、窄 helper 或由`plugin`根包构造入口显式注入的 resolver。

#### Scenario: route adapter 解析动态路由元数据
- **WHEN** 源码插件通过宿主服务目录读取当前动态路由元数据
- **THEN** hostservices 通过运行时子组件的窄能力或显式注入的 resolver 完成读取
- **AND** hostservices 不得通过 import `internal/service/plugin`调用 facade alias

#### Scenario: 导入边界审查
- **WHEN** 审查 hostservices 迁移后的生产 Go 代码
- **THEN** 静态检索不得发现`plugin/internal/hostservices`导入`internal/service/plugin`
- **AND** 静态检索不得发现生产代码继续导入旧`internal/service/pluginhostservices`

### Requirement:能力目录普通消费面不得暴露宿主内部治理对象

系统 SHALL 将`pkg/plugin/capability.Services`定义为源码插件消费宿主能力的普通服务目录，并将`pkg/plugin/capability/guest.Directory`定义为动态插件消费宿主能力的 guest 目录。这些目录返回的普通消费接口 MUST 只暴露状态、DTO、批量投影、只读查询和可降级能力，MUST NOT 暴露`*gdb.Model`、`*ghttp.Request`、DAO、DO、Entity、宿主写入、数据范围注入、启动一致性或自动开通等内部治理能力。

#### Scenario:源码插件获取能力目录

- **WHEN** 源码插件通过 registrar 或 callback 输入获取宿主能力目录
- **THEN** 插件只能看到普通消费接口
- **AND** 不能通过该目录拿到底层数据库模型、HTTP 请求对象或宿主内部写入接口

#### Scenario:动态插件获取 guest 能力目录

- **WHEN** 动态插件通过`pkg/plugin/capability/guest`访问宿主能力
- **THEN** guest SDK 只提供经`hostServices`授权的 DTO 化 host service client，并通过`Directory.Data()`统一返回受治理的`capability/data` facade
- **AND** 不暴露`gdb.Model`、DAO、Entity 或宿主内部 service 实例

#### Scenario:普通插件需要新增宿主读能力

- **WHEN** 插件展示或编排场景需要读取新增宿主能力数据
- **THEN** 能力目录新增只读 DTO、批量投影或状态方法
- **AND** 不通过恢复旧宽接口、写入方法、数据范围注入方法或宿主内部对象满足该需求

### Requirement:组织和租户 capability 必须拆分普通消费面、provider 面和宿主内部治理面

系统 SHALL 将`orgcap`和`tenantcap`拆分为多个职责明确的接口。`capability.Services.Org()`、`capability.Services.Tenant()`、`guest.Directory.Org()`和`guest.Directory.Tenant()`只能返回普通消费接口；provider 实现、数据库范围注入、HTTP 租户解析、用户成员关系写入、租户插件自动开通和启动一致性检查必须通过独立接口表达。

#### Scenario:普通组织能力消费

- **WHEN** 插件或宿主普通业务需要读取组织能力状态、用户部门投影、部门树或岗位选项
- **THEN** 它使用`orgcap.Service`普通消费接口
- **AND** 该接口不包含数据库模型注入或用户组织关系写入方法

#### Scenario:宿主内部组织数据范围过滤

- **WHEN** 宿主需要按组织关系在数据库查询阶段注入数据范围
- **THEN** 它使用独立的组织范围治理接口
- **AND** 该接口不通过`capability.Services`或`guest.Directory`暴露给普通插件

#### Scenario:普通租户能力消费

- **WHEN** 插件或宿主普通业务需要读取当前租户、租户可用性、租户列表或租户可见性校验
- **THEN** 它使用`tenantcap.Service`普通消费接口
- **AND** 该接口不包含`*ghttp.Request`解析、数据库模型注入、用户租户关系写入或启动一致性方法

#### Scenario:宿主内部租户治理

- **WHEN** 宿主中间件、用户、角色、通知或插件运行时需要租户解析、数据过滤、成员关系写入、自动开通或启动一致性检查
- **THEN** 它使用对应的`RequestResolver`、`ScopeService`、`UserMembershipService`、`PluginProvisioningService`或`StartupConsistencyService`
- **AND** 这些接口通过构造函数显式注入，不从普通插件能力目录动态查找

### Requirement:Provider 构造环境必须强类型且按 capability 收窄

系统 SHALL 为每个 capability 定义自己的 provider 构造环境。provider factory MUST 接收强类型环境，环境字段只包含该 provider adapter 合法需要的宿主能力，MUST NOT 使用`any`传递完整能力目录。

#### Scenario:组织 provider 构造

- **WHEN** `linapro-org-core`或其他组织 provider 插件声明 provider factory
- **THEN** factory 接收`orgcap.ProviderEnv`等强类型环境
- **AND** 环境只包含组织 provider adapter 需要的宿主能力，例如租户过滤、i18n 或其他明确字段
- **AND** factory 不再对`env.Services`执行运行时类型断言
- **AND** 生产代码中不得继续保留`contract.ProviderEnv.Services`兼容字段或转发层

#### Scenario:租户 provider 构造

- **WHEN** `linapro-tenant-core`或其他租户 provider 插件声明 provider factory
- **THEN** factory 接收`tenantcap.ProviderEnv`等强类型环境
- **AND** 环境只包含租户 provider adapter 需要的宿主能力，例如业务上下文和插件生命周期服务
- **AND** factory 不得获得完整`capability.Services`后自行扩张依赖

### Requirement: pluginhost 不得拥有 HostServices 能力目录语义

系统 SHALL 将`pluginhost`限定为源码插件贡献入口。源码插件需要访问宿主能力时，registrar、callback payload 和测试替身 MUST 直接使用`pluginhost.Services`或命名为`Services()`的访问器；系统 MUST 删除`pluginhost.HostServices`、`ScopedHostServicesFactory`、`HostServicesForPlugin`和`HostServices()`等历史组件或方法。

#### Scenario:源码插件注册路由

- **WHEN** 源码插件路由注册回调需要宿主能力
- **THEN** registrar 暴露`Services()`方法返回`pluginhost.Services`
- **AND** 不再暴露`HostServices()`或`pluginhost.HostServices`
- **AND** 生产代码中旧`HostServices()`调用必须迁移完成或被治理扫描阻断

#### Scenario:源码插件 callback 获取宿主能力

- **WHEN** hook、cron 或 lifecycle callback 需要宿主能力
- **THEN** callback 输入直接提供`pluginhost.Services`源码插件运行期服务目录语义
- **AND** 不再通过`pluginhost.HostServicesForPlugin`完成 scoped 绑定

### Requirement: 配置和清单类只读宿主服务必须使用 get 方法声明

系统 SHALL 要求动态插件在`plugin.yaml`中为`config`、`hostConfig`和`manifest`等只读读取型宿主服务仅声明`methods: [get]`。宿主 MUST 拒绝这些服务上的写入、保存、热重载、完整快照、typed helper 或未知方法声明，并在运行时对每次`get`调用执行 service、method 和资源范围校验。

#### Scenario: 配置服务只声明 get

- **WHEN** 动态插件声明`service: config`且`methods: [get]`
- **THEN** 宿主接受声明并派生插件作用域配置读取能力
- **AND** 授权快照不包含`exists`、`string`、`bool`、`int`、`duration`或`scan`

#### Scenario: 宿主公开配置声明 key 白名单

- **WHEN** 动态插件声明`service: hostConfig`、`methods: [get]`和`resources.keys`
- **THEN** 宿主仅允许该插件读取声明且经宿主确认的公开宿主配置键
- **AND** 未声明或未确认的键在运行时被拒绝

#### Scenario: Manifest 声明资源路径白名单

- **WHEN** 动态插件声明`service: manifest`、`methods: [get]`和`resources.paths`
- **THEN** 宿主仅允许该插件读取声明且经宿主确认的当前插件`manifest/`资源路径
- **AND** 跨插件路径、绝对路径、URL 和路径穿越必须被拒绝

#### Scenario: 只读服务声明写入方法被拒绝

- **WHEN** 动态插件在`config`、`hostConfig`或`manifest`服务中声明`set`、`save`、`reload`或其他非`get`方法
- **THEN** 构建器或宿主清单校验拒绝该声明
- **AND** 运行时不得为该插件派生对应宿主服务能力

### Requirement: capability 必须作为源码插件和动态插件的统一能力集合

系统 SHALL 将`pkg/plugin/capability`定义为插件消费宿主能力的统一集合，并将根接口命名为`capability.Services`。源码插件 MUST 通过 registrar 或等价上下文获取 capability services；动态插件 MUST 通过 guest client 和 host service handler 进入同一组服务语义；两类插件不得分别使用不同组件暴露同一能力。

#### Scenario: 源码插件和动态插件读取同一能力

- **WHEN** 源码插件和动态插件分别消费插件作用域配置、宿主公开配置、manifest、数据服务或框架 capability
- **THEN** 二者共享同一 service 契约、授权边界、错误语义和降级策略
- **AND** 仅 transport 和运行时加载方式存在差异

#### Scenario: 新能力只注册到统一目录

- **WHEN** 系统新增一个插件可消费宿主能力
- **THEN** 该能力必须注册到`pkg/plugin/capability`根服务集合或其子包
- **AND** 动态插件的 host service handler 只把 bridge 请求映射到该统一服务集合

### Requirement: 动态 hostServices 与 capability services 必须语义分层

系统 SHALL 保留动态插件 manifest 中的`hostServices`作为授权声明和 bridge transport 调用面，同时将宿主能力语义归属到`pkg/plugin/capability`和`capability.Services`。`hostServices`不得被重新解释为 Go 公共能力集合名称，`capability`也不得绕过动态插件授权快照直接授予动态插件访问权。

#### Scenario: 动态插件声明 hostServices

- **WHEN** 动态插件在`plugin.yaml`中声明`hostServices`
- **THEN** 该声明表示动态插件申请调用的 service、method 和资源边界
- **AND** 宿主在安装、启用或升级阶段生成授权快照
- **AND** 该声明不改变`pkg/plugin/capability`中能力契约的 owner

#### Scenario: 动态插件通过 capability guest client 调用宿主能力

- **WHEN** 动态插件调用`pkg/plugin/capability/guest`中的能力 client
- **THEN** guest client 通过`pkg/plugin/pluginbridge/guest`raw transport 发起 host service 调用
- **AND** 宿主先校验`hostServices`授权快照
- **AND** 授权通过后再委托到同一`capability.Services`能力集合
- **AND** `RuntimeHostService`、`StorageHostService`、`ConfigHostService`、`DataHostService`等 guest 能力 client 接口、默认实例、WASI 实现和非 WASI stub 均归属`pkg/plugin/capability/guest`
- **AND** `pkg/plugin/pluginbridge/guest`只提供 raw host call transport、guest runtime 和 route binding，不拥有宿主能力 client 语义
- **AND** 能力 client 方法签名中的 host service DTO、cron 合约、日志等级和 codec 均直接使用`pkg/plugin/pluginbridge/protocol`，`capability/guest`不得重复导出这些协议别名

#### Scenario: 动态插件通过 data SDK 调用宿主 data 能力

- **WHEN** 动态插件调用`pkg/plugin/capability/data`中的 ORM-style data SDK
- **THEN** data SDK 通过`pkg/plugin/pluginbridge/guest`raw transport 和`pkg/plugin/pluginbridge/protocol`协议 DTO 发起 data host service 调用
- **AND** 宿主先校验`hostServices`中的 data 授权快照、资源表和方法范围
- **AND** 授权通过后再执行同一 data host service 治理路径

### Requirement: pluginservice 必须作为源码插件和动态插件的统一能力目录

系统 SHALL 将`pkg/pluginservice`定义为插件消费宿主能力的统一目录。源码插件 MUST 通过 registrar 或等价上下文获取`pluginservice.Services`；动态插件 MUST 通过 guest client 和 host service handler 进入同一组服务语义；两类插件不得分别使用不同组件暴露同一能力。

#### Scenario: 源码插件和动态插件读取同一能力

- **WHEN** 源码插件和动态插件分别消费插件作用域配置、宿主公开配置、manifest、数据服务或 pluginservice capability
- **THEN** 二者共享同一 service 契约、授权边界、错误语义和降级策略
- **AND** 仅 transport 和运行时加载方式存在差异

#### Scenario: 新能力只注册到统一目录

- **WHEN** 系统新增一个插件可消费宿主能力
- **THEN** 该能力必须注册到`pluginservice.Services`或其子目录
- **AND** 动态插件的 host service handler 只把 bridge 请求映射到该统一目录

### Requirement: 动态插件必须通过 ai.text.generate 调用文本 AI

系统 SHALL 在动态插件宿主服务体系中提供`ai`service family，并开放`text.generate`方法。动态插件 MUST 通过`hostServices`声明`service: ai`和对应`methods`申请文本`AI`调用能力，并由宿主授权快照确认后才能调用。

#### Scenario: 插件声明文本 AI 宿主服务

- **WHEN** 动态插件在`plugin.yaml`中声明`service: ai`和`methods: [text.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为文本`AI`调用权限申请
- **AND** 声明 MUST NOT 包含`resources`、`paths`、`tables`或`keys`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限，运行时只认宿主确认后的授权快照
- **AND** 运行时 MUST 将该方法映射为`capabilityType=text`与`capabilityMethod=generate`

#### Scenario: 未声明插件调用被拒绝

- **WHEN** 动态插件未声明或未获确认`ai.text.generate`授权却发起文本`AI`调用
- **THEN** 宿主 MUST 在执行`framework.ai.text.v1`或渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

### Requirement: ai.text.generate 调用契约必须支持文本参数和 thinkingEffort

系统 SHALL 为动态插件 `ai.text.generate` 定义 DTO 化调用载荷。载荷 MUST 支持 `purpose`、`tier`、`messages`、`maxOutputTokens`、`temperature`、可选 `thinkingEffort` 和短字符串 `metadata`，并与 `framework.ai.text.v1` 保持语义一致。

#### Scenario: 动态插件传入 thinkingEffort

- **WHEN** 动态插件调用 `ai.text.generate` 并传入 `thinkingEffort`
- **THEN** 宿主 MUST 校验该值属于 `low`、`medium`、`high`、`xhigh`、`max`
- **AND** 授权通过后 MUST 将该值传递给 `framework.ai.text.v1`
- **AND** 模型不支持该 effort 时 MUST 返回文本能力的结构化业务错误

#### Scenario: 动态插件不传 thinkingEffort

- **WHEN** 动态插件调用 `ai.text.generate` 未传 `thinkingEffort`
- **THEN** 宿主 MUST 保持字段为空并由档位默认值或模型默认行为决定实际 effort
- **AND** 宿主 MUST NOT 在 host service 层硬编码某个渠道专有 effort 值

#### Scenario: 动态插件 metadata 有界

- **WHEN** 动态插件在 `ai.text.generate` 中传入 `metadata`
- **THEN** 宿主 MUST 只接受短字符串键值用于调用来源、业务请求 ID 或审计关联
- **AND** 宿主 MUST 拒绝或截断超出契约边界的大段输入

### Requirement: ai.text.generate 必须记录最小审计信息

系统 SHALL 对动态插件发起的 `ai.text.generate` 调用记录最小宿主服务审计和智能中心调用日志。审计 MUST 支持诊断授权、耗时、状态和来源插件，但 MUST NOT 保存完整输入、完整输出、隐藏思考内容或密钥。

#### Scenario: 成功调用记录来源

- **WHEN** 动态插件通过 `ai.text.generate` 成功生成文本
- **THEN** 宿主服务审计 MUST 记录 `pluginId`、service、method、purpose 摘要、结果状态和耗时
- **AND** 智能中心调用日志 MUST 记录 `sourcePluginId`、`purpose`、档位、渠道模型投影、`thinkingEffort`、token 用量和耗时

#### Scenario: 失败调用记录脱敏错误

- **WHEN** 动态插件调用 `ai.text.generate` 失败
- **THEN** 宿主服务审计和智能中心调用日志 MUST 记录失败状态与脱敏错误摘要
- **AND** 审计和日志 MUST NOT 包含完整 `messages`、API key、认证头或渠道响应原文

### Requirement: ai service 必须拒绝首期未开放的方法

系统 SHALL 将 `ai` service family 设计为可扩展宿主服务族，但首期 MUST 只开放 `text.generate`。图片、音频、向量、重排、工具调用、流式输出或其他 `AI` 方法在未有正式规范和授权模型前 MUST 被拒绝。

#### Scenario: 未开放 image 方法被拒绝

- **WHEN** 动态插件声明或调用 `ai.image.generate`
- **THEN** 构建器、宿主清单校验或运行时 handler MUST 拒绝该方法
- **AND** 错误 MUST 指出该 `AI` host service method 当前不受支持

#### Scenario: 后续能力独立扩展

- **WHEN** 系统后续新增图片、音频或向量能力
- **THEN** 新方法 MUST 明确定义 `capabilityType`、`capabilityMethod`、资源授权、请求响应 DTO、审计字段和与框架 capability 的适配关系
- **AND** 新方法 MUST NOT 改变现有 `ai.text.generate` 的授权和同步文本响应语义

### Requirement: hostServices 必须支持领域服务和领域方法

系统 SHALL 允许动态插件通过`hostServices`声明宿主发布的领域服务和领域方法。领域协议服务名 MUST 使用语言无关的领域名，并且普通领域 service 名 MUST 与`pkg/plugin/capability.Services`领域目录名称保持一致；集合型领域使用`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`，命名空间型领域继续使用`authz`、`dict`、`org`、`tenant`、`ai`等领域名。领域协议名不得使用 Go 包名或宿主内部实现名。每个领域方法 MUST 映射到领域能力接口或受控领域适配器。

#### Scenario: 动态插件声明用户领域读取

- **WHEN** 动态插件在`plugin.yaml`中声明`service: users`和`methods: [users.batch_get, users.search]`
- **THEN** 宿主校验该领域服务和方法已发布
- **AND** 安装授权确认后将归一化声明写入运行时授权快照

#### Scenario: 动态插件调用未知领域方法

- **WHEN** 动态插件调用未发布、未声明或未授权的领域方法
- **THEN** 宿主返回能力拒绝或能力不可用错误
- **AND** 不进入任何领域业务逻辑

### Requirement: host service 调用必须传递 CapabilityContext

系统 SHALL 在每一次动态`hostServices`领域调用中构造并传递`CapabilityContext`。该上下文 MUST 包含插件`ID`、执行来源、actor、tenant、授权快照、资源或投影标识、系统调用标识和审计摘要。

#### Scenario: 请求型 host service 调用

- **WHEN** 动态插件在请求路由中调用领域`host service`
- **THEN** WASM host service handler 将当前用户、租户、插件`ID`、路由来源和授权快照传入领域适配器
- **AND** 领域适配器基于上下文执行数据权限和审计治理

#### Scenario: 系统型 host service 调用

- **WHEN** 动态插件在生命周期、hook 或定时任务中调用领域`host service`
- **THEN** WASM host service handler 使用宿主创建的系统 actor 构造上下文
- **AND** 需要用户上下文的领域方法必须拒绝或按领域定义的系统调用边界执行

### Requirement: 动态领域管理方法使用安装授权模型

系统 SHALL 允许动态插件在`hostServices`中声明宿主显式发布的领域管理方法。安装或启用阶段确认授权后，运行时不再额外校验当前用户是否拥有对应工作台菜单或按钮权限；领域管理方法 MUST 继续校验目标资源可见性、租户边界、数据权限、状态机、数量上限和审计来源。

#### Scenario: 动态插件调用授权管理方法

- **WHEN** 动态插件调用已授权的领域管理方法
- **THEN** host service handler 校验`service + method`存在于运行时授权快照
- **AND** 请求进入对应领域`AdminService`或命令适配器
- **AND** 领域命令执行目标边界、状态机和审计校验

#### Scenario: 动态插件越权访问目标资源

- **WHEN** 动态插件已获方法授权但请求操作跨租户、不可见或状态不允许的目标资源
- **THEN** 领域方法拒绝该操作
- **AND** 响应使用结构化业务错误
- **AND** 宿主记录失败审计摘要

### Requirement: 宿主服务适配器必须适配到`*cap`能力组件

系统 SHALL 要求源码插件 hostservices directory、动态插件`WASM`host service handler 和 guest SDK 最终适配到`pkg/plugin/capability/<domain>cap`能力组件。适配层 MUST 不再依赖`capability/contract`具体能力服务接口作为运行时服务目录契约。

#### Scenario: 源码插件构造宿主服务目录

- **WHEN** 宿主启动期构造源码插件可消费的`capability.Services`
- **THEN** directory 字段和方法返回目标`*cap.Service`
- **AND** 认证授权能力以`authcap.Service`能力族聚合入口发布，token 子服务归属`authcap/token`，授权子服务归属`authcap/authz`
- **AND** 插件作用域配置 factory 通过`Plugins().Config()`对应子服务发布
- **AND** 其他插件作用域能力 factory 使用对应`manifestcap`或其他目标组件
- **AND** 源码插件专用`TenantFilter()`返回`tenantcap.PluginTableFilterService`，且不进入普通`capability.Services`

#### Scenario: 动态 host service 调用领域能力

- **WHEN** 动态插件通过`hostServices`调用`service + method`
- **THEN** `WASM`host service handler 校验授权后进入对应`*cap.Service`或`*cap.AdminService`
- **AND** handler 不得通过旧`contract.*Service`绕过目标能力组件

### Requirement: 动态插件不得直接消费源码插件租户过滤器

系统 SHALL 禁止动态插件 guest SDK、动态`hostServices`协议和`WASM`host service handler 暴露`tenantcap.PluginTableFilterService`、`*gdb.Model`、SQL 片段、DAO 或 query builder。动态插件租户隔离 MUST 通过普通`tenantcap.Service`读取当前租户或校验租户可见性，并由宿主 host service handler 在调用边界执行与宿主 API 等价的数据权限和租户边界过滤。

#### Scenario: 动态插件读取当前租户

- **WHEN** 动态插件需要知道当前请求租户
- **THEN** guest SDK 通过普通`Tenant().Current()`或等价 host service 调用读取
- **AND** 宿主返回租户 DTO 或租户 ID
- **AND** guest 不得获得`tenantcap.PluginTableFilterService`

#### Scenario: 动态插件调用宿主数据读取能力

- **WHEN** 动态插件通过用户、文件、通知、配置、插件治理或其他 host service 读取数据
- **THEN** 对应 handler 在宿主侧基于调用身份、`pluginID`、当前租户、授权快照和既有数据权限边界执行过滤
- **AND** 动态插件不得传入 SQL、DAO、`*gdb.Model`或 query builder 表达租户过滤

#### Scenario: 未来动态插件需要自有数据存储

- **WHEN** 动态插件需要维护插件自有持久化数据
- **THEN** 系统必须另行设计租户安全存储能力
- **AND** 该能力由宿主按`pluginID + tenantID`自动隔离
- **AND** 不得复用源码插件专用`TenantFilter()`作为动态插件数据访问协议

### Requirement: Go 包重命名不得改变动态插件协议

系统 SHALL 将本次`*cap`包重命名视为 Go 公共包边界重构。动态插件`plugin.yaml hostServices`声明、运行时授权快照、`service`字符串、`method`字符串、资源授权、protobuf envelope、错误 envelope 和审计语义 MUST 保持当前目标模型不变。

#### Scenario: 动态插件声明 AI 文本能力

- **WHEN** 动态插件声明`service: ai`和`method: text.generate`
- **THEN** 宿主仍按`service: ai`和`method: text.generate`校验授权
- **AND** Go 侧能力组件重命名为`aicap`不得要求插件清单改为`service: aicap`

#### Scenario: 动态插件读取插件配置

- **WHEN** 动态插件声明`service: config`和`methods: [get]`
- **THEN** 宿主仍按`config.get`授权快照执行校验
- **AND** Go 侧插件配置能力收口到`Plugins().Config()`不得改变动态协议服务名

#### Scenario: 动态插件读取宿主配置

- **WHEN** 动态插件声明`service: hostConfig`和授权配置 key
- **THEN** 宿主仍按宿主配置授权快照执行校验
- **AND** Go 侧宿主配置能力使用`HostConfig()`，不得与`Plugins().Config()`混用

### Requirement: 适配器迁移必须复用启动期共享实例

系统 SHALL 在`*cap`包重命名和接口迁移后继续由宿主运行期统一构造 host service 适配器。缓存敏感服务、插件状态、配置、`i18n`、会话、通知、组织、租户和插件生命周期等依赖 MUST 复用启动期共享实例或共享后端，不得因包迁移在插件调用路径中临时`New()`关键服务图。

#### Scenario: 源码插件读取缓存

- **WHEN** 源码插件通过迁移后的`cachecap.Service`读取插件作用域缓存
- **THEN** 该服务仍委托启动期传入的共享`kvcache`后端
- **AND** 不得在每次插件调用中创建独立缓存实例

#### Scenario: 动态插件读取插件状态

- **WHEN** 动态插件通过迁移后的`Plugins().State()`读取启用状态
- **THEN** handler 仍使用启动期注入的共享插件状态服务或共享快照
- **AND** 不得退化为仅当前节点可见的包级默认实例

### Requirement: 动态插件 AI host service 必须支持多模态方法声明

系统 SHALL 扩展动态插件`service: ai`host service，使插件可以声明图片、向量、音频、视觉、文档、安全审核和视频方法。每个方法 MUST 映射到明确的`capabilityType + capabilityMethod`和独立授权分类；这些方法授权 MUST 仅通过`methods`表达，不得通过`resources`表达`purpose`或策略属性。

#### Scenario: 插件声明图片生成能力

- **WHEN** 动态插件在`plugin.yaml`中声明`service: ai`和`methods: [image.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为`image.generate`权限申请
- **AND** 运行时 MUST 将方法映射为`capabilityType=image`和`capabilityMethod=generate`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限

#### Scenario: 插件声明音频能力

- **WHEN** 动态插件声明`audio.transcribe`或`audio.synthesize`
- **THEN** 宿主 MUST 分别识别为不同 host service 方法
- **AND** 两个方法 MUST 使用独立 payload 契约和授权分类
- **AND** 两个方法 MUST NOT 要求插件声明`purpose`资源策略

#### Scenario: computer act 声明被拒绝

- **WHEN** 动态插件声明`computer.act`、`ui.operate`或等价 UI 控制方法
- **THEN** 清单校验或运行时 MUST 拒绝该声明
- **AND** 错误 MUST 表明该方法不属于本轮`ai`host service 支持范围

### Requirement: AI host service 大对象 payload 必须使用资产引用

系统 SHALL 要求动态插件多模态`ai`host service 使用`assetRef`或受控临时资产引用传递大对象。host service 请求和响应 MUST NOT 传输无上限 base64 或完整二进制内容。

#### Scenario: 图片输入使用资产引用

- **WHEN** 动态插件调用`vision.analyze`并提供图片输入
- **THEN** 请求 MUST 使用`assetRef`、mime 类型和大小投影引用图片
- **AND** 宿主 MUST 校验该资产引用对当前插件和请求上下文可访问

#### Scenario: 音频输出使用资产引用

- **WHEN** 动态插件调用`audio.synthesize`成功
- **THEN** 响应 MUST 返回输出音频的`assetRef`和摘要投影
- **AND** 响应 MUST NOT 返回完整音频 base64

### Requirement: AI host service 必须支持 provider operation 查询边界

系统 SHALL 允许动态插件在获得授权后使用 provider operation 查询方法跟踪渠道异步 operation。operation 查询 MUST 表达渠道协议状态，MUST NOT 表达业务任务状态。

#### Scenario: 视频生成返回 provider operation

- **WHEN** 动态插件调用`video.generate`
- **AND** 渠道返回异步 operation
- **THEN** host service 响应 MUST 返回不透明`operationRef`、状态、渠道模型投影、`nextPollAfterMs`和过期时间
- **AND** 响应 MUST NOT 返回业务任务 ID

#### Scenario: 查询 operation 状态

- **WHEN** 动态插件调用`video.operation.get`
- **AND** 插件已获该 operation 所属方法和资源授权
- **THEN** 宿主 MUST 返回 operation 当前状态或完成后的资产引用
- **AND** 宿主 MUST NOT 返回 provider 原始认证 URL、密钥或完整响应正文

#### Scenario: 未授权取消被拒绝

- **WHEN** 动态插件调用`video.operation.cancel`
- **AND** 授权资源未允许取消或 provider 不支持取消
- **THEN** 宿主 MUST 拒绝调用并返回结构化错误
- **AND** 宿主 MUST NOT 执行 provider 取消请求

### Requirement: 多模态 AI host service 必须记录最小审计

系统 SHALL 对动态插件多模态`ai`host service 调用记录最小审计信息。审计 MUST 支持诊断插件、方法、资源、耗时、状态和错误，但 MUST NOT 保存完整输入输出、大对象内容、渠道响应原文或密钥。

#### Scenario: 成功调用记录摘要

- **WHEN** 动态插件通过多模态`ai`host service 成功调用 provider
- **THEN** 宿主服务审计 MUST 记录`pluginId`、service、method、purpose、授权资源摘要、状态和耗时
- **AND** 智能中心调用日志 MUST 记录来源插件、能力方法、渠道模型投影、资产引用摘要和用量摘要

#### Scenario: 失败调用脱敏

- **WHEN** 多模态`ai`host service 调用失败
- **THEN** 审计和调用日志 MUST 记录失败状态、稳定错误码和脱敏错误摘要
- **AND** 审计和日志 MUST NOT 包含完整文件内容、音视频内容、API key、认证头或渠道响应原文

### Requirement: WASM host service 生产依赖必须由启动期显式配置

系统 SHALL 由宿主启动期统一配置 WASM host service dispatcher 所需的 cache、lock、notify、storage config、plugin config、manifest、AI、organization、tenant 和其他 host service 运行期依赖。生产代码中的 WASM host service 包级变量 MUST NOT 默认调用 `New()` 创建 cache、config、session、notify、lock、plugin service 或 capability manager 等关键服务实例。缺失配置 MUST 以显式初始化错误或 host call internal error 暴露。

#### Scenario: 启动期配置共享 cache host service

- **WHEN** 宿主 HTTP runtime 初始化动态插件 WASM host service
- **THEN** `cache` host service 使用启动期已创建并选定共享后端的 `kvcache.Service`
- **AND** WASM host service 包不得保留可被生产路径使用的默认 `kvcache.New()` fallback

#### Scenario: storage host service 使用启动期配置服务

- **WHEN** 动态插件调用 `storage` host service
- **THEN** storage dispatcher 使用启动期注入的插件动态存储配置 reader
- **AND** 不通过包级默认 `config.New()` 读取独立配置服务图

#### Scenario: 测试 fixture 显式注入 host service 依赖

- **WHEN** 单元测试或集成测试执行 WASM host service dispatcher
- **THEN** 测试 fixture 显式调用配置入口注入 fake、memory 或共享测试服务
- **AND** 测试不得依赖生产包级默认服务实例

### Requirement: host service 配置入口必须覆盖完整依赖集

系统 SHALL 为 WASM host service 提供一个启动期可调用的完整配置入口，覆盖所有已发布 host service dispatcher 的必需运行期依赖。新增 host service dispatcher 时，系统 MUST 同步更新配置入口和缺失配置测试，确保新依赖不会绕过启动装配。

#### Scenario: 新增 host service 后配置覆盖测试失败

- **WHEN** 开发者新增一个 WASM host service dispatcher
- **AND** 未把该 dispatcher 的必需依赖纳入统一配置入口或覆盖测试
- **THEN** 自动化测试失败
- **AND** 失败信息指向缺失的 host service 配置项

### Requirement: 动态插件 guest SDK 必须通过 AI 命名空间调用文本 AI

系统 SHALL 在动态插件 guest 侧通过`AI().Text()`暴露文本`AI`能力。guest SDK 的`AI().Text().GenerateText(...)`MUST 继续使用既有`ai.text.generate`host service 协议，并保持`host:ai:text`能力分类、类型化 DTO 和脱敏审计语义；请求资源 MUST NOT 再使用`purpose:<name>`表达授权用途。

#### Scenario: 动态插件通过 AI 命名空间生成文本

- **WHEN** 动态插件需要调用文本`AI`生成能力
- **THEN** guest 代码 MUST 通过`guest.Default().AI().Text().GenerateText(...)`或等价能力目录入口发起调用
- **AND** guest SDK MUST NOT 继续要求调用方使用根目录`AIText()`方法

#### Scenario: guest AI Text 调用进入既有 host service

- **WHEN** guest SDK 执行`AI().Text().GenerateText(...)`
- **THEN** SDK MUST 构造既有`service: ai`、`method: text.generate`host service 调用
- **AND** 请求 envelope 的`resourceRef` MUST 为空
- **AND** `purpose` MUST 仅作为请求 DTO 字段传递
- **AND** 宿主 MUST 在执行文本能力或渠道调用前完成 service、method 和来源身份校验

#### Scenario: 动态插件协议不因 Go 入口重构改变

- **WHEN** 系统将 guest 侧调用入口从`AIText()`重构为`AI().Text()`
- **THEN** 动态插件`plugin.yaml`中的`hostServices`声明格式 MUST 保持`service: ai`和`methods: [text.generate]`
- **AND** `host:ai:text`的能力分类和脱敏审计语义 MUST 保持不变

### Requirement: WASM host service dispatch 必须由显式 registry 驱动

系统 SHALL 将动态插件 WASM host service dispatch 收敛为显式注册的 registry 驱动结构。`wasm_host_service.go`入口 MUST 只负责 envelope 解码、调用上下文构造、授权校验、registry lookup 和统一错误响应；`internal/service/plugin/internal/wasm/hostservicedispatch`MUST 拥有 registry、handler context、注册校验和通用响应辅助。具体 service/method 处理逻辑 MAY 继续保留在`wasm`父包作为显式注册适配层，避免为了迁移目录扩大`hostCallContext`、运行时快照和插件执行状态的公开面；若后续领域 handler 迁移到子包，MUST 先抽取窄上下文契约并保持 DI 来源清晰。registry 注册 MUST 使用显式装配函数，不得使用`init()`隐式注册。

#### Scenario: 已注册 method 正常分发

- **WHEN** 动态插件调用一个已在 registry 注册且已授权的 service/method
- **THEN** `wasm_host_service.go`通过 registry lookup 定位 handler
- **AND** handler 或父包适配层接收统一 host call context、resource identifier、method 和 payload
- **AND** handler 返回统一 host call response envelope

#### Scenario: 未知 service 或 method 被拒绝

- **WHEN** 动态插件调用未在 registry 注册的 service/method
- **THEN** 宿主返回结构化“不支持”或“未找到”错误
- **AND** 宿主不得进入任何实际领域能力、数据访问、缓存、网络或外部资源调用

#### Scenario: 入口文件不维护 service 级 switch

- **WHEN** 静态检索`internal/service/plugin/internal/wasm/wasm_host_service.go`
- **THEN** 不得存在按 host service family 分发到`dispatch<X>HostService`的 service 级大 switch
- **AND** 新增领域 host service 不需要修改该入口文件的分发分支

#### Scenario: 注册方式保持显式依赖注入

- **WHEN** 宿主启动并构造 WASM host service registry
- **THEN** 注册入口显式接收 handler 需要的共享运行期依赖
- **AND** handler 不得通过`init()`、包级默认实例或调用关键服务`New()`自行获得依赖
- **AND** 缺失依赖必须在构造或注册阶段返回错误

### Requirement: 领域 dispatch handler 必须保持宿主治理边界

系统 SHALL 要求每个 host service dispatch handler 在 registry 驱动结构下继续保持既有授权、数据权限、租户边界、缓存一致性、审计和错误 envelope 语义。普通领域 handler 只负责 transport DTO 与`capability/<x>cap`领域契约之间的转换，不得直接依赖宿主 DAO、DO、Entity、私有缓存快照或未发布内部 service 实现。

#### Scenario: 数据访问能力通过等价数据权限边界

- **WHEN** 动态插件通过 host service handler 读取列表、详情、批量信息、候选项或执行写操作
- **THEN** handler 必须保持与宿主 API 等价的数据权限、租户边界和目标可见性校验
- **AND** 不得因为 registry 重构绕过授权快照、数据范围过滤或目标记录可见性检查

#### Scenario: 缓存敏感能力复用共享实例

- **WHEN** handler 访问 cache、session、权限快照、插件状态、运行时配置或其他缓存敏感能力
- **THEN** handler 必须复用启动期注入的共享服务实例或共享后端
- **AND** 不得在插件调用路径中创建仅当前节点可见的默认实例
- **AND** registry 重构不得改变缓存权威源、失效触发点、跨实例同步或可接受陈旧窗口

#### Scenario: 普通领域 handler 不暴露宿主内部模型

- **WHEN** 普通领域 handler 调用`usercap`、`dictcap`、`filecap`、`sessioncap`或其他`capability/<x>cap`契约
- **THEN** handler 只传递插件可见 DTO、投影或值对象
- **AND** 不得把`*gdb.Model`、DAO、DO、Entity、HTTP request 或宿主私有 service 实例暴露给 guest client

#### Scenario: 错误响应保持结构化

- **WHEN** handler 因授权、参数、数据权限、租户边界、缓存后端或领域服务失败返回错误
- **THEN** 宿主必须返回现有 host service 错误 envelope 或等价结构化错误
- **AND** 不得把裸 Go error 文本、敏感请求体、完整响应体或密钥写入插件可见响应

### Requirement: 源码插件宿主服务适配器必须归属 plugin 内部 capabilityhost 子组件

系统 SHALL 将源码插件宿主服务适配器实现归属到`apps/lina-core/internal/service/plugin/internal/capabilityhost`子组件。该子组件负责把宿主启动期共享的`auth`、`apidoc`、`bizctx`、`datascope`、`i18n`、`notify`、`session`、`kvcache`、`orgcap`、`tenantcap`和插件生命周期能力适配为`pkg/plugin/capability.Services`与`pkg/plugin/pluginhost.Services`。`internal/service/plugin/internal/hostservices`不得作为长期生产入口保留。

#### Scenario: 启动期构造源码插件能力目录

- **WHEN** 宿主 HTTP runtime 需要构造源码插件可消费的领域能力目录
- **THEN** 启动期通过`internal/service/plugin`根包暴露的窄构造入口创建`capability.Services`
- **AND** 该入口委托`plugin/internal/capabilityhost`完成具体适配
- **AND** `internal/cmd`不得直接导入`plugin/internal/capabilityhost`

#### Scenario: 适配器复用共享运行期实例

- **WHEN** `plugin/internal/capabilityhost`构造`capability.Services`
- **THEN** 所有接口型运行期依赖必须由启动期逐项显式传入
- **AND** 适配器不得在构造函数、插件回调路径或 host service 调用路径中创建独立的`auth`、`session`、`plugin`、`i18n`、`notify`、`kvcache`、`orgcap`或`tenantcap`服务实例

#### Scenario: 源码插件获取插件作用域能力

- **WHEN** 源码插件 registrar、hook、route 或 jobs 回调需要插件作用域的 host services
- **THEN** 其获取的目录仍满足`capability.Services`和必要的`pluginhost.Services`契约
- **AND** cache、config 和 manifest 等插件作用域能力继续按插件`ID`绑定

### Requirement: 动态普通领域 host service 必须共享单一领域能力目录

系统 SHALL 要求动态插件普通领域`host service`分发统一使用启动期注入的同一个`capability.Services`目录。`WASM`运行时 MUST 只通过`ConfigureDomainHostServices(capability.Services)`配置普通领域能力，不得为`AI`、`User`、`Org`、`Tenant`或其他领域继续新增领域专用`Configure*HostService`函数、领域专用包级服务目录或 fallback 能力目录。

#### Scenario: 启动期配置动态领域能力

- **WHEN** 宿主调用`ConfigureWasmHostServices`
- **THEN** 该入口只为普通领域能力调用一次`ConfigureDomainHostServices`
- **AND** `AI`、`User`、`Org`和`Tenant`动态分发均通过该共享目录按插件`ID`绑定后获取对应`*cap.Service`

#### Scenario: data host service 需要组织能力

- **WHEN** 动态`data`host service 为数据范围过滤需要组织能力
- **THEN** 它必须通过共享领域能力目录获取当前插件作用域的`Org()`服务
- **AND** 不得依赖组织领域专用全局变量或专用 Configure 入口

#### Scenario: 新增动态领域方法

- **WHEN** 开发者为动态插件新增一个普通领域`host service method`
- **THEN** 宿主分发代码必须复用`ConfigureDomainHostServices`维护的共享目录
- **AND** 不得新增与该领域同名的独立配置入口

### Requirement: 动态普通领域 host service 协议名必须与领域目录一致

系统 SHALL 要求动态插件普通领域`hostServices.service`协议名与已发布的动态领域目录名称保持一致。集合型领域 MUST 使用复数领域名：`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`。对应能力字符串 MUST 分别使用`host:users`、`host:files`、`host:jobs`、`host:notifications`、`host:plugins`和`host:sessions`。`i18n`不属于动态插件可声明的普通领域 host service；动态插件多语言资源由宿主统一管理。项目不保留旧单数 service 别名。

#### Scenario: 动态插件声明集合型领域服务

- **WHEN** 动态插件在`plugin.yaml`中声明用户、文件、任务、通知、插件治理或在线会话领域 host service
- **THEN** service 必须分别使用`users`、`files`、`jobs`、`notifications`、`plugins`或`sessions`
- **AND** 宿主 descriptor、授权快照、guest 调用和`WASM`dispatcher 必须使用同一 service 名

#### Scenario: 动态插件声明插件生命周期治理方法

- **WHEN** 动态插件在`plugin.yaml`中声明`service: plugins`和`method: lifecycle.tenant_delete.ensure`
- **THEN** 宿主校验该方法属于`plugins`领域已发布方法
- **AND** 运行时必须先校验`host:plugins`能力和授权快照中的精确 method
- **AND** 通过校验后才能进入`plugincap.LifecycleService`

#### Scenario: 动态插件声明单一命名空间领域服务

- **WHEN** 动态插件声明`auth`、`authz`、`apidoc`、`bizctx`、`dict`、`infra`、`route`、`ai`、`org`或`tenant`
- **THEN** service 继续使用该领域命名空间名称
- **AND** 不得为了形式统一将其机械复数化

#### Scenario: 动态插件声明 i18n host service

- **WHEN** 动态插件在`plugin.yaml`中声明`service: i18n`
- **THEN** host service 校验必须拒绝该声明
- **AND** public protocol alias、guest 公共目录和`WASM`dispatcher 不得暴露`i18n`服务

### Requirement: 插件自身配置读取必须归属 plugins 领域

系统 SHALL 将动态插件自身配置读取归属到`plugins`领域能力。动态插件公共入口 MUST 使用`pluginbridge.Services.Plugins().Config()`；`plugin.yaml hostServices`授权 MUST 使用`service: plugins`和`method: config.get`。系统 MUST NOT 继续发布`service: config`、`host:config`、公共`pluginbridge.ConfigHostService`或独立`dispatchConfigHostService`。

#### Scenario: 动态插件读取自身配置

- **WHEN** 动态插件声明`service: plugins`和`method: config.get`
- **AND** guest 侧调用`pluginbridge.Services.Plugins().Config().Get(ctx, key)`
- **THEN** 宿主必须校验`host:plugins`能力和授权快照中的`config.get`方法
- **AND** 通过`plugincap.ConfigService`读取当前插件作用域配置
- **AND** active artifact 默认配置必须继续通过`WithArtifactConfig`参与读取

#### Scenario: 动态插件声明旧 config host service

- **WHEN** 动态插件在`plugin.yaml`中声明`service: config`
- **THEN** host service 校验必须拒绝该声明
- **AND** public protocol alias、guest 公共目录和`WASM`总 dispatcher 不得再暴露独立`config`服务

### Requirement: 通知发送必须归属 notifications 领域

系统 SHALL 将动态插件通知读取和发送统一归属到`notifications`领域能力。读取消息 MUST 使用`messages.batch_get`；发送消息 MUST 使用`messages.send`。系统 MUST NOT 继续发布`service: notify`、`host:notify`或公共`pluginbridge.Notify()`。

#### Scenario: 动态插件发送通知

- **WHEN** 动态插件声明`service: notifications`、`method: messages.send`和授权的通知渠道资源引用
- **AND** guest 侧通过`pluginbridge.Services.Notifications().Send(ctx, capCtx, input)`发送通知
- **THEN** 宿主必须校验`host:notifications`能力、精确 method 和渠道资源引用
- **AND** 通过校验后才能进入通知领域发送能力

#### Scenario: 动态插件声明旧 notify host service

- **WHEN** 动态插件在`plugin.yaml`中声明`service: notify`
- **THEN** host service 校验必须拒绝该声明
- **AND** public protocol alias、guest 公共目录和`WASM`总 dispatcher 不得再暴露独立`notify`服务

### Requirement: 动态定时任务必须归属 jobs 领域

系统 SHALL 将动态插件定时任务的管理边界归属到`jobs`领域。动态插件 MUST NOT 通过`service: cron`、`host:cron`、公共`pluginbridge.Cron()`、`pluginbridge.Services.Cron()`、`CronHostService`或内部 reserved `cron.register`host-call 声明定时任务。动态插件需要交付内置定时任务时，MUST 使用`service: jobs`和`method: jobs.register`的发现期声明契约；声明结果 MUST 进入宿主 Jobs 管理投影、状态控制、handler 发布和调度执行链。

#### Scenario: 动态插件声明旧 cron host service

- **WHEN** 动态插件在`plugin.yaml`中声明`service: cron`
- **THEN** host service 校验必须拒绝该声明
- **AND** public protocol alias、guest 公共目录和`WASM`总 dispatcher 不得再暴露独立`cron`服务

#### Scenario: 动态插件通过旧 cron host-call 注册任务

- **WHEN** 动态插件通过旧`cron.register`host-call 或包级`pluginbridge.Cron()`尝试注册定时任务
- **THEN** 该调用路径在公开 guest SDK、public protocol 和`WASM`dispatcher 中必须不存在
- **AND** 系统不得触发动态插件 cron discovery 执行

#### Scenario: 动态插件通过 jobs 领域声明内置任务

- **WHEN** 动态插件声明`service: jobs`和`method: jobs.register`
- **AND** 宿主执行动态插件 Jobs 发现入口
- **AND** guest 侧通过`RegisterPlugin(plugin pluginbridge.Declarations)`中的`plugin.Jobs().Register(...)`提交任务声明
- **THEN** 宿主必须校验`host:jobs`能力和授权快照中的`jobs.register`方法
- **AND** 仅在 Jobs 发现执行源中接受该声明
- **AND** 声明必须通过 Jobs 合约校验后进入宿主管理投影和执行 handler 绑定

#### Scenario: 动态插件运行期尝试注册 Jobs

- **WHEN** 动态插件在普通路由、生命周期、hook 或任务执行期间调用`jobs.register`
- **THEN** 宿主必须拒绝该调用
- **AND** 不得修改 Jobs 管理投影或发布新的执行 handler

### Requirement: 源码插件定时任务注册必须归属 jobs 领域

系统 SHALL 将源码插件内置定时任务注册入口归属到`jobs`领域。源码插件 MUST 使用`pluginhost.Jobs().RegisterJobs(...)`、`ExtensionPointJobsRegister`和`JobsRegistrar`声明任务；系统 MUST NOT 继续发布`pluginhost.Cron()`、`RegisterCron`、`CronRegistrar`或`ExtensionPointCronRegister`作为源码插件公开注册契约。

#### Scenario: 源码插件声明内置定时任务

- **WHEN** 源码插件需要在启动期声明插件内置定时任务
- **THEN** 插件必须通过`pluginhost.Jobs().RegisterJobs(ExtensionPointJobsRegister, ...)`注册声明回调
- **AND** 回调接收的 registrar 必须是`JobsRegistrar`
- **AND** 宿主管理投影、执行 handler 引用和任务同步接口必须使用 Jobs 语义

#### Scenario: 源码插件尝试使用旧 cron 注册入口

- **WHEN** 源码插件代码引用`pluginhost.Cron()`、`RegisterCron`、`CronRegistrar`或`ExtensionPointCronRegister`
- **THEN** 这些公开标识符必须不存在
- **AND** 编译期必须暴露调用方未迁移的问题

### Requirement: AI host service 调用必须受 service、method 和 DTO 能力边界约束

系统 SHALL 对每一次`ai`host service 调用校验`service: ai`、声明的`method`、调用来源和请求 DTO 边界。`ai`host service MUST NOT 使用`resources`、`resourceRef`或`purpose:<name>`作为运行时授权条件；`purpose`、`tier`、`maxOutputTokens`、资产引用和其他方法参数 MUST 由请求 DTO 承载，并由对应`AI`子能力服务或`linapro-ai-core`治理。

#### Scenario: 方法授权后调用文本能力

- **WHEN** 动态插件已获`ai.text.generate`方法授权
- **AND** 请求 DTO 中提交`purpose`、`tier`、`messages`和`maxOutputTokens`
- **THEN** host service handler MUST 将请求转换为`AI().Text().GenerateText(...)`调用
- **AND** 宿主 MUST 使用 host-call 上下文中的`pluginID`注入来源插件身份
- **AND** 宿主 MUST NOT 按`purpose:<name>`资源授权或`resources.attributes`限制该请求

#### Scenario: 未授权方法被拒绝

- **WHEN** 动态插件未声明或未获确认`ai.document.cite`对应方法授权
- **THEN** 宿主 MUST 在执行`AI().Document().Cite(...)`或任何渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

#### Scenario: 请求参数由 AI 能力服务校验

- **WHEN** 动态插件已获`ai.text.generate`方法授权但请求 DTO 中`tier`非法
- **THEN** host service handler MUST 将请求交给类型化`AI`文本能力边界处理
- **AND** 错误 MUST 保持`AI`能力服务的 DTO 校验和 provider 可用性语义

### Requirement: 动态插件普通领域 host service 必须覆盖源码插件普通领域能力

系统 SHALL 让动态插件通过`hostServices`获得已发布动态普通领域能力的覆盖。动态插件领域 host service MUST 使用语言无关的领域服务名和方法名，MUST 使用`resourceKind: none`表达方法授权，运行时 MUST 从宿主注入的同一个`capability.Services`目录进入对应`*cap.Service`。动态插件协议 MUST NOT 暴露`AdminServices`目录、数据库查询构造器、`DAO/DO/Entity`、HTTP 请求对象、宿主内部 service 或`i18n`运行时翻译服务。

#### Scenario: 动态插件声明普通领域能力

- **WHEN** 动态插件声明`auth`、`authz`、`users`、`dict`、`files`、`sessions`、`jobs`、`infra`、`apidoc`、`bizctx`、`route`、`notifications`、`plugins`、`ai`、`org`或`tenant`等领域 host service
- **THEN** 宿主清单校验 MUST 识别对应领域方法已经发布
- **AND** 声明 MUST 只包含`methods`
- **AND** 声明 MUST NOT 包含`resources`、`paths`、`tables`或`keys`
- **AND** 运行时授权快照 MUST 只按`service + method`校验调用

#### Scenario: 动态插件调用普通领域能力

- **WHEN** 动态插件通过已授权领域方法读取用户、权限、字典、文件、会话、任务、基础设施、API 文档、业务上下文、路由元数据、通知消息、插件治理、AI、组织或租户投影
- **THEN** `WASM`host service handler MUST 构造`CapabilityContext`
- **AND** handler MUST 使用`Capability.ServicesForPlugin(..., pluginID)`取得插件绑定的能力目录
- **AND** 请求 MUST 进入对应`*cap.Service`普通消费面或该普通消费面拥有的子服务
- **AND** 领域实现 MUST 继续执行租户、数据权限、可见性、批量上限、缓存和`i18n`治理

#### Scenario: 动态插件不能通过普通领域面获得管理能力

- **WHEN** 动态插件声明普通领域 host service
- **THEN** 宿主 MUST NOT 因该声明暴露创建、更新、删除、状态变更、授权关系变更、执行任务、撤销会话或发送通知等管理动作
- **AND** 未来如需动态插件管理方法，MUST 通过显式发布的领域管理方法和独立授权语义进入

#### Scenario: 重叠动态能力收敛到领域能力

- **WHEN** 既有动态 host service 与普通领域能力语义重叠
- **THEN** 宿主实现 MUST 优先复用`capability.Services`中对应领域能力或插件绑定子服务
- **AND** 不得继续维护一套与领域能力平行且语义漂移的动态专用实现

### Requirement: WASM host service runtime 必须由实例持有

系统 SHALL 让动态插件 WASM host service 分发器通过显式 runtime 实例读取运行期依赖。WASM host service 的 domain capability directory、插件配置 factory、宿主配置 service、manifest service factory 和其他共享宿主能力 MUST 来自启动期构造并注入的实例，不得通过包级`Configure*`函数、包级`atomic.Pointer`快照或包级默认实例作为生产调用事实源。

#### Scenario: 宿主启动配置 WASM host service

- **WHEN** 宿主启动并初始化动态插件 runtime
- **THEN** 启动路径创建 WASM host service runtime 实例
- **AND** 该实例显式接收共享 capability directory、config factory、host config service 和 manifest factory
- **AND** 生产启动后不调用`wasm.Configure*`配置包级状态

#### Scenario: 动态插件执行 host call

- **WHEN** 动态插件通过统一 host service 协议执行 host call
- **THEN** 分发器从当前 runtime 实例读取依赖
- **AND** 不读取包级 snapshot
- **AND** 多个测试或多 service 实例可以持有彼此隔离的 WASM runtime

#### Scenario: 缺少 WASM runtime 依赖

- **WHEN** 启动期未提供 WASM host service 必需依赖
- **THEN** 构造函数返回错误
- **AND** 错误在启动装配阶段暴露
- **AND** 不允许请求路径首次 host call 时才因包级配置缺失失败

### Requirement: Host call 授权快照可以请求内复用但不得改变治理语义

系统 SHALL 允许`WASM`host service handler 在同一次 guest 执行中复用已构建的 host service 授权快照。复用 MUST 仅降低快照装配成本，不得改变当前 active release 授权来源、service/method/resource 校验、数据权限、租户边界、审计字段或错误 envelope。

#### Scenario: 同一次 guest 执行复用授权快照

- **WHEN** 动态插件在一次路由请求中连续调用多个 host service
- **THEN** 宿主可以复用本次`ExecuteBridge`入口构建的授权快照
- **AND** 每次 host call 仍校验 service、method 和资源标识是否已授权

#### Scenario: 授权收缩后新请求使用新快照

- **WHEN** 插件 active release 的 host service 授权被收缩并发布`plugin-runtime`修订号
- **THEN** 后续 guest 执行不得继续使用旧请求中的授权快照
- **AND** 未授权 service、method 或资源调用必须被拒绝

#### Scenario: 系统型调用不伪造用户上下文

- **WHEN** 动态插件在生命周期、hook 或 cron 中调用需要用户上下文的 host service
- **THEN** 即使授权快照命中，handler 也必须按领域契约拒绝或按系统调用边界处理
- **AND** 不得伪造请求型用户身份来绕过数据权限

### Requirement: WASM host service 公共 helper 必须归属公共层

系统 SHALL 将跨领域复用的 WASM host service helper 归属到`wasm`公共 host service 文件或`hostservicedispatch`公共层。具体领域文件 MUST 只承载该领域 service/method 的 transport 适配、授权前置和领域能力调用，不得承载`CapabilityContext`构造、统一 envelope 辅助或 registry 公共响应逻辑。

#### Scenario: 公共 capability context helper 不在用户领域文件

- **WHEN** 静态检索`apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_users.go`
- **THEN** 文件中不得定义`capabilityContextForHostCall`
- **AND** 用户领域 handler 仍通过同包公共 helper 构造`CapabilityContext`

#### Scenario: 新增普通领域 host service 不修改统一入口分发分支

- **WHEN** 开发者新增一个普通领域 host service 或 method
- **THEN** 开发者只需要新增或修改领域文件中的 handler 实现
- **AND** 开发者通过显式 registry 注册条目接入该 service/method
- **AND** `wasm_host_service.go`不得新增按 service family 直接分发的 switch 分支

#### Scenario: 公共 helper 不扩大运行期依赖来源

- **WHEN** 公共 helper 被迁移到公共层
- **THEN** helper 继续使用调用路径已有的`hostCallContext`和启动期注入能力
- **AND** 不得通过包级默认实例、`init()`或临时`New()`创建`auth`、`session`、`i18n`、`cache`或插件 runtime 依赖

### Requirement: 动态插件 i18n 资源必须由宿主管理

系统 SHALL 允许动态插件继续通过`manifest/i18n`交付多语言资源，但运行时资源发现、合并、缓存、失效和前端语言包分发 MUST 由宿主统一管理。动态插件后端 MUST NOT 通过 host service 读取 locale、翻译消息或检索 message key，也 MUST NOT 自行读取`manifest/i18n`资源完成运行时翻译。

#### Scenario: 动态插件交付多语言资源

- **WHEN** 动态插件 artifact 包含`manifest/i18n/<locale>/*.json`资源
- **THEN** 宿主资源扫描和多语言聚合流程负责发现并合并这些资源
- **AND** 动态插件不需要声明`service: i18n`

#### Scenario: 动态插件返回可本地化业务结果

- **WHEN** 动态插件后端需要返回用户可见状态、错误或提示
- **THEN** 响应应返回稳定 message key、message params、英文 fallback 或原始业务数据
- **AND** 最终展示本地化由宿主错误治理、前端运行时语言包或宿主统一展示层完成

