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

系统 SHALL 对每一次宿主服务调用同时执行粗粒度 capability 校验和细粒度资源授权校验，但 capability 集合必须由`hostServices`声明自动推导，而不是要求插件作者在`plugin.yaml`中重复维护第二份`capabilities`列表；任一层不满足都必须拒绝调用。只读读取型服务的`methods` MUST 表达真实 host service 调用动作，SDK typed helper 不得作为独立授权方法进入声明或运行时快照。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者在动态插件清单中声明`hostServices`
- **THEN** 构建器校验 service、method、资源声明（如低优先级服务的逻辑`resourceRef`、`storage.resources.paths`、URL 模式、`resources.tables`、宿主公开配置 key 或 manifest 资源路径）和策略参数是否合法
- **AND** 宿主根据这些 methods 自动推导内部 capability 分类快照
- **AND** 将归一化后的宿主服务策略写入运行时产物
- **AND** 宿主装载产物后恢复为当前 release 的服务授权快照

#### Scenario: 缺少授权的宿主服务调用被拒绝

- **WHEN** 插件调用未声明的 service、method 或未授权的资源标识
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

### Requirement: 源码插件宿主服务适配器必须归属 plugin 内部 hostservices 子组件

系统 SHALL 将源码插件宿主服务适配器实现归属到`apps/lina-core/internal/service/plugin/internal/hostservices`子组件。该子组件负责把宿主启动期共享的`auth`、`apidoc`、`bizctx`、`datascope`、`i18n`、`notify`、`session`、`kvcache`、`orgcap`、`tenantcap`和插件生命周期能力适配为`pkg/plugin/capability.Services`。`apps/lina-core/internal/service/pluginhostservices`不得作为长期生产入口保留。

#### Scenario: 启动期构造源码插件宿主服务目录
- **WHEN** 宿主 HTTP runtime 需要构造源码插件可消费的宿主服务目录
- **THEN** 启动期通过`internal/service/plugin`根包暴露的窄构造入口创建`capability.Services`
- **AND** 该入口委托`plugin/internal/hostservices`完成具体适配
- **AND** `internal/cmd`不得直接导入`plugin/internal/hostservices`

#### Scenario: 适配器复用共享运行期实例
- **WHEN** `plugin/internal/hostservices`构造`capability.Services`
- **THEN** 所有接口型运行期依赖必须由启动期逐项显式传入
- **AND** 适配器不得在构造函数、插件回调路径或 host service 调用路径中创建独立的`auth`、`session`、`plugin`、`i18n`、`notify`、`kvcache`、`orgcap`或`tenantcap`服务实例

#### Scenario: 源码插件获取插件作用域能力
- **WHEN** 源码插件 registrar、hook、route 或 cron 回调需要插件作用域的 host services
- **THEN** 其获取的目录仍满足`capability.Services`和必要的`pluginhost.Services`契约
- **AND** cache、config 和 manifest 等插件作用域能力继续按插件 ID 绑定

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

