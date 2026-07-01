## ADDED Requirements

### Requirement: 插件可见方法必须具备方法级治理元数据

系统 SHALL 为每个插件可见领域方法记录方法级治理元数据。治理元数据 MUST 至少包含风险等级、授权资源、上下文要求、数据权限、性能边界和缓存影响。风险等级仅用于授权展示、升级风险、测试和审查门禁，不得作为动态插件可声明性的功能开关。

#### Scenario: 新增插件可见领域方法

- **WHEN** 系统新增或修改一个插件可见领域方法
- **THEN** OpenSpec 设计或对应方法矩阵必须记录该方法的风险等级、授权资源、上下文要求、数据权限、性能边界和缓存影响
- **AND** 不得使用`risk`决定该方法是否自动进入动态插件可声明集合

#### Scenario: 动态插件声明方法

- **WHEN** 动态插件声明一个领域`service + method`
- **THEN** 宿主只根据动态`host service registry`是否存在该方法判断能否进入声明和授权流程
- **AND** 宿主不得要求插件管理者维护`dynamic-auth`、`source-only`、`reserved`或等价方法可声明性字段

### Requirement: 动态插件领域方法必须由 registry 注册事实决定可声明性

系统 SHALL 使用动态`host service registry`作为动态插件领域方法可声明性的唯一事实来源。已注册方法可以被动态插件在`plugin.yaml hostServices`中声明，并在安装、启用和运行时接受授权校验；未注册方法 MUST 在构建、安装、启用或运行时被拒绝，且不得进入领域 owner。

#### Scenario: 动态插件声明已注册方法

- **WHEN** 动态插件在`plugin.yaml hostServices`中声明已注册的`service + method`
- **THEN** 宿主允许该声明进入安装或启用授权确认流程
- **AND** 运行时仍必须校验授权快照和资源范围

#### Scenario: 动态插件声明未注册方法

- **WHEN** 动态插件声明或调用未注册到动态`host service registry`的`service + method`
- **THEN** 构建、安装、启用或运行时校验必须拒绝该方法
- **AND** 请求不得进入`capability.Services`或领域 owner 业务实现

#### Scenario: 源码插件调用未注册动态方法

- **WHEN** 源码插件调用某个未注册为动态 host service 的统一`Service`方法
- **THEN** 源码插件可以按类型化 Go 契约调用该方法
- **AND** 源码插件调用必须复用标准业务`ctx`，方法内部从`ctx`读取并执行租户、数据权限、状态机和缓存治理

## MODIFIED Requirements

### Requirement: 领域能力必须定义清晰接口分层

系统 SHALL 为每个宿主领域定义职责明确的接口边界。插件可见消费面 MUST 使用单一`Service`作为领域入口表达读取、候选、批量投影、统计、标签解析、校验、执行、创建、更新、删除、状态变更和授权关系变更等完整领域用例。稳定子资源或关联关系 MAY 在同一领域`Service`下通过命名子能力聚合，例如`Assignment()`。系统 MUST NOT 通过插件可见`AdminService`、`Services.Admin()`或等价管理目录表达风险边界。宿主内部数据库范围注入 MUST 使用`ScopeService`等内部接口，不得暴露给普通插件。

#### Scenario: 公开能力目录不重复暴露同一领域

- **WHEN** 源码插件或动态插件通过`capability.Services`、`pluginhost.Services`或动态 guest 目录获取宿主能力
- **THEN** 同一领域只能暴露一个插件可见稳定`Service`入口
- **AND** 写入、删除、状态变更和执行类动作必须并入对应领域`Service`
- **AND** 插件自身配置、静态宿主配置和`sys_config`必须使用明确命名区分，插件作用域配置通过`Config`表达，静态宿主配置通过`HostConfig`表达，`sys_config`通过`HostConfig().SysConfig()`表达
- **AND** 通知、会话等领域不得同时暴露旧`contract.*Service`和新领域`*cap.Service`

#### Scenario: 普通插件消费领域能力

- **WHEN** 普通源码插件或动态插件调用领域能力
- **THEN** 它只能获得领域`Service`或对应 guest 代理
- **AND** 返回值使用领域`DTO`、值对象、批量投影或结构化响应
- **AND** 动态 guest 代理不得返回`*gdb.Model`、`DAO`、`DO`、`Entity`、HTTP 请求对象或内部 service 实例
- **AND** 源码插件同进程查询构造器辅助只能作为对应领域子能力暴露，不得通过`pluginhost.Services`顶层分散入口或动态 WASM 协议暴露

#### Scenario: 用户角色关联关系归入用户子领域

- **WHEN** 源码插件需要替换一个用户的角色关联关系
- **THEN** 插件通过`Services.Users().Assignment().ReplaceRoles(...)`或注入的`usercap.AssignmentService`调用
- **AND** `usercap.Service`顶层只保留用户基础读写、状态和凭证类方法
- **AND** 后续新增用户与角色关联关系方法必须继续聚合在`Assignment()`子领域
- **AND** 动态插件只有在该子领域方法被注册为动态`host service`后才能声明和调用

#### Scenario: 租户公共能力保持只读和校验边界

- **WHEN** 源码插件或动态插件消费`Tenant().Directory()`和`Tenant().Membership()`
- **THEN** `Directory()`只暴露`Get`、`BatchGet`、`List`和`EnsureVisible`
- **AND** 租户创建、更新、状态变更和删除必须留在租户 owner 的插件 API、内部 service 或受控宿主流程中，不得通过通用`tenantcap.DirectoryService`发布为插件公共写入能力
- **AND** `Membership()`只暴露`ListByUser`和`Validate`
- **AND** 用户租户成员关系替换必须留在宿主内部`tenantspi.UserMembershipService`或租户 owner 内部 service 中，不得通过通用`tenantcap.MembershipService`发布为插件公共写入能力

### Requirement: 治理能力必须内聚到对应领域组件

系统 SHALL 将插件治理能力和普通能力统一归属到对应领域能力组件。插件生命周期归属`Plugins().Lifecycle()`，插件启用状态读取归属`Plugins().State()`，插件配置归属`Plugins().Config()`。`pluginhost.Services` MUST NOT 暴露顶层`PluginLifecycle()`或`PluginState()`方法，也 MUST NOT 作为新增治理能力的事实 owner。租户插件启停、租户插件默认供给和插件表租户过滤归属`Tenant()`领域；保留的源码插件租户快捷入口只能委托到`Tenant()`子能力。动态插件可以声明和调用已注册的治理领域方法，未注册方法 MUST 被构建、安装、启用和运行时校验拒绝。

#### Scenario: 源码插件调用插件生命周期治理

- **WHEN** 源码插件需要执行插件生命周期治理检查或通知
- **THEN** 它应通过`Services.Plugins().Lifecycle()`或注入的`plugincap.LifecycleService`调用
- **AND** `pluginhost.Services`不得暴露`PluginLifecycle()`同级顶层入口

#### Scenario: 源码插件读取插件启用状态

- **WHEN** 源码插件需要读取插件启用状态
- **THEN** 它应通过`Services.Plugins().State()`或注入的`plugincap.StateService`调用
- **AND** `pluginhost.Services`不得暴露`PluginState()`同级顶层入口

#### Scenario: 动态插件声明插件治理方法

- **WHEN** 动态插件在`plugin.yaml hostServices`中声明插件治理领域方法
- **THEN** 宿主必须先确认该方法已由`Plugins()`领域 owner 注册到动态`host service registry`
- **AND** 安装或启用授权后运行时仍校验`service + method + resource`
- **AND** 方法实现必须进入`plugincap.Service`或对应领域 owner 适配器，而不是进入`pluginhost`或`pluginbridge`平行业务接口

#### Scenario: 源码插件应用插件表租户过滤

- **WHEN** 源码插件在插件自有表查询中需要注入租户过滤
- **THEN** 它应通过`Services.Tenant().Filter().Context()`读取租户过滤上下文
- **AND** `Services.Tenant().Filter()` MUST 静态返回普通`tenantcap.FilterService`
- **AND** 需要直接修改`*gdb.Model`时 MUST 调用`tenantspi.ApplyPluginTableFilter(ctx, filter, model, qualifier)`
- **AND** `pluginhost.Services` MUST NOT 定义独立`TenantService`镜像或顶层`TenantTableFilter()`入口
- **AND** `*gdb.Model`只允许停留在源码插件同进程 Go SPI 或宿主内部适配器中，不得进入普通`tenantcap.FilterService`或动态插件 wire 协议
- **AND** 动态插件的插件表租户隔离必须通过`RecordStore`或领域 host service 的可序列化参数完成

#### Scenario: 宿主内部注入数据范围

- **WHEN** 宿主领域适配器需要在数据库查询阶段注入租户、组织或数据权限过滤
- **THEN** 它使用宿主内部`ScopeService`或等价窄接口
- **AND** 该接口不得通过`capability.Services`、`pluginhost.Services`普通消费面或动态 guest 目录暴露给插件

### Requirement: 插件可见目录的可见性校验必须统一为批量签名

系统 SHALL 为插件可见目录的可见性校验统一使用单一`EnsureVisible(ctx, ids []...) error`方法。普通消费面 MUST NOT 同时暴露单个可见性方法和批量可见性方法；如果调用场景只有一个 ID，也应通过单一批量签名传入长度为 1 的切片。该约束旨在让单条校验、列表装配和批量拒绝语义保持一致，避免同一目录出现并列的单条/批量接口形态。

#### Scenario: 定义组织或租户目录可见性校验

- **WHEN** 系统定义组织、租户或其他插件可见目录的可见性校验
- **THEN** 目录接口 MUST 只暴露一个`EnsureVisible`方法
- **AND** 该方法的参数 MUST 为 ID 切片
- **AND** 不得再新增`EnsureVisibleMany`入口

### Requirement: 动态路由和 API 文档能力不得暴露 HTTP 框架对象

系统 SHALL 要求普通领域能力契约使用`context.Context`、路径、方法、DTO 或中立值对象传递请求相关信息，不得在普通`capability/**`父包中暴露`*ghttp.Request`或`*ghttp.HandlerItemParsed`。

#### Scenario: 动态路由元数据读取

- **WHEN** 插件通过`routecap.Service.GetMetadata`读取当前动态路由元数据
- **THEN** 方法签名接收`context.Context`
- **AND** 返回值使用`routecap.Metadata`
- **AND** 宿主适配器在内部从上下文恢复 HTTP 请求并读取元数据
- **AND** `routecap`父包不 import `ghttp`

### Requirement: 动态插件领域方法必须通过安装授权快照调用

系统 SHALL 要求动态插件在`plugin.yaml hostServices`中声明已注册到动态`host service registry`的领域`service + method`，并在安装或启用阶段由宿主确认授权后形成运行时授权快照。运行时调用 MUST 在 dispatcher 内校验授权快照中的领域服务、方法和资源或投影范围，不得把授权摘要作为额外领域上下文传入 owner。集合型领域的协议 service 名 MUST 与`capability.Services`领域目录名称保持一致，例如用户、文件、任务、通知、插件治理和在线会话领域分别使用`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`。安装授权替代插件级菜单/RBAC 方法校验，但不得替代领域数据权限、租户边界、状态机和数量上限校验。

#### Scenario: 动态插件调用已授权领域读取方法

- **WHEN** 动态插件声明并获得授权调用`service: users`和`method: users.batch_get`
- **THEN** 运行时 host service 分发器允许请求进入`usercap.Service`
- **AND** `usercap.Service`仍从标准业务`ctx`读取当前用户、租户和数据权限上下文，并过滤租户、数据权限和可见字段

#### Scenario: 动态插件调用已授权领域管理方法

- **WHEN** 动态插件声明并获得授权调用已注册的领域管理方法
- **THEN** 运行时不再额外校验当前用户是否拥有某个工作台菜单或按钮权限
- **AND** 领域方法仍必须校验目标资源可见性、状态机、租户边界和批量上限

#### Scenario: 动态插件调用未注册领域方法

- **WHEN** 动态插件声明、安装、启用或调用未注册到动态`host service registry`的领域方法
- **THEN** 宿主必须拒绝该方法
- **AND** 不得因为该方法存在于 Go 统一`Service`接口中就允许动态插件调用

### Requirement: 领域能力边界必须具有固定归属

系统 SHALL 将插件可消费领域能力固定为四类边界：`pkg/plugin/capability`拥有领域契约，真实领域 owner 拥有业务实现，`pkg/plugin/pluginhost`拥有源码插件消费入口，`pkg/plugin/pluginbridge`和动态 host service 协议拥有动态插件 transport 与 guest 消费入口。`internal/service/plugin/internal/capabilityhost`和 WASM host service 只承担标准业务上下文桥接、动态授权、编解码和错误映射职责，不得成为跨领域业务实现 owner。任何新增领域能力 MUST 先进入对应`*cap`契约，再由领域 owner 实现和动态 transport 适配，不得在`WASM`分发、`pluginbridge`协议目录或动态插件公共 SDK 中单独定义一套平行业务接口。

#### Scenario: 新增宿主领域能力契约

- **WHEN** 系统新增一个插件可消费的宿主领域能力
- **THEN** 领域接口、`DTO`、领域`ID`和降级语义必须定义在`pkg/plugin/capability/<domain>cap`或等价领域命名空间
- **AND** `pluginbridge`不得成为该领域业务接口的 owner

#### Scenario: 宿主实现领域能力

- **WHEN** 宿主启动期构造`capability.Services`
- **THEN** 具体业务实现必须归属真实领域 owner 或 owner 发布的稳定适配契约
- **AND** `internal/service/plugin/internal/capabilityhost`只允许作为动态调用薄适配层
- **AND** `capabilityhost`不得直接访问其他领域`DAO`、`DO`、`Entity`、私有缓存、私有 provider 或内部 helper
- **AND** `internal/cmd`不得直接导入`plugin/internal/capabilityhost`

#### Scenario: 源码插件消费领域能力

- **WHEN** 源码插件通过 registrar、hook、route 或 jobs callback 获取宿主能力
- **THEN** 插件只能通过`pkg/plugin/pluginhost.Services`获取插件作用域的统一`capability.Services`和源码插件专用能力
- **AND** 插件业务服务必须继续注入所需的最窄`*cap.Service`或源码插件专用接口
- **AND** 不得再注入或保存`AdminService`

#### Scenario: 源码插件声明启动期能力

- **WHEN** 源码插件需要声明嵌入文件、生命周期回调、后端钩子、HTTP 路由、内置 Jobs 或治理过滤器
- **THEN** 插件必须通过`pluginhost.NewDeclarations()`创建声明期 facade
- **AND** 通过`RegisterSourcePlugin(plugin pluginhost.Declarations)`注册声明结果
- **AND** 源码插件声明期子 facade 必须使用`*Declarations`命名，例如`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`AccessDeclarations`
- **AND** `SourcePlugin*`前缀只用于源码插件读模型、生命周期输入、回调处理器或其他非声明期 facade 契约
- **AND** `pluginhost.Declarations`不得作为运行时领域能力挂载到`pluginhost.Services`

#### Scenario: 动态插件消费领域能力

- **WHEN** 动态插件通过`hostServices`调用普通领域能力
- **THEN** 宿主侧分发必须先校验动态 registry、授权快照和资源范围，再进入启动期共享的`capability.Services`
- **AND** guest 侧公共入口必须位于`pkg/plugin/pluginbridge`
- **AND** 普通领域 hostcall 代理实现必须位于`pluginbridge/internal/domainhostcall`或等价 internal 子组件
- **AND** 框架领域的可用性与诊断状态必须通过对应 owner 领域读取，不得由`plugins`领域聚合`org`、`tenant`或`AI`状态
- **AND** 普通领域协议 service 名必须与`capability.Services`领域目录名称保持一致

#### Scenario: 动态插件声明内置 Jobs

- **WHEN** 动态插件需要声明宿主管理的内置定时任务
- **THEN** guest 侧必须通过动态插件声明期对象的`Jobs()`facade 提交`jobs.register`声明
- **AND** 宿主只在 Jobs 发现执行源中接收声明
- **AND** `jobs.register`不得作为运行期`jobcap.Service`方法暴露给源码插件或动态插件业务服务
- **AND** 不得重新引入`cron`领域对象、`CronHostService`或`cron.register`协议

#### Scenario: 动态插件声明启动期能力

- **WHEN** 动态插件需要声明构建期路由分组、内置 Jobs 或后续生命周期声明能力
- **THEN** 插件必须通过`RegisterPlugin(plugin pluginbridge.Declarations)`使用声明期 facade 表达
- **AND** `Declarations.Routes()`和`Declarations.Jobs()`不得作为运行时领域能力挂载到`pluginbridge.Services`
- **AND** 运行时业务服务必须继续通过`pluginbridge.Services`获取普通`*cap.Service`领域能力

#### Scenario: 动态插件消费插件领域能力

- **WHEN** 动态插件通过`pluginbridge.Services.Plugins()`获取插件领域能力
- **THEN** 返回值必须实现`plugincap.Service`
- **AND** `Config()`、`Registry()`和`State()`必须归属同一个`plugins`领域对象
- **AND** `Lifecycle()`和状态读取治理必须归属同一个`plugins`领域对象，动态是否可用由 registry 注册事实和授权快照决定
- **AND** 公共`guest`包不得再声明与`plugincap.Service`平行的`PluginService`接口

## REMOVED Requirements

### Requirement: 源码插件管理能力必须通过 AdminServices 目录提供

**Reason**: 插件可见领域能力统一收敛到每个领域一个`Service`入口，`AdminServices`和`Services.Admin()`会保留两套风险表达方式，并与动态插件的`service + method + resource`授权模型重复。

**Migration**: 删除`capability.AdminServices`、各领域`AdminService`和`pluginhost.Services.Admin()`；将原管理方法并入对应领域`Service`，调用方注入最窄统一`*cap.Service`，领域方法仅依赖标准业务`ctx`中的当前用户、租户、数据权限和系统调用标识执行治理。
