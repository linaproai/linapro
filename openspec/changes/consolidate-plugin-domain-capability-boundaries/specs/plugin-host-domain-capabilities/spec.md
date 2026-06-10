## ADDED Requirements

### Requirement: 领域能力边界必须具有固定归属

系统 SHALL 将插件可消费领域能力固定为四类边界：`pkg/plugin/capability`拥有领域契约，`internal/service/plugin/internal/capabilityhost`拥有宿主实现，`pkg/plugin/pluginhost`拥有源码插件消费入口，`pkg/plugin/pluginbridge`和`pluginbridge`host service 协议拥有动态插件 transport 与 guest 消费入口。任何新增领域能力 MUST 先进入对应`*cap`契约，再由宿主实现和动态 transport 适配，不得在`WASM`分发、`pluginbridge`协议目录或动态插件公共 SDK 中单独定义一套平行业务接口。

#### Scenario: 新增宿主领域能力契约

- **WHEN** 系统新增一个插件可消费的宿主领域能力
- **THEN** 领域接口、`DTO`、领域`ID`和降级语义必须定义在`pkg/plugin/capability/<domain>cap`或等价领域命名空间
- **AND** `pluginbridge`不得成为该领域业务接口的 owner

#### Scenario: 宿主实现领域能力

- **WHEN** 宿主启动期构造`capability.Services`
- **THEN** 具体适配实现必须归属`internal/service/plugin/internal/capabilityhost`
- **AND** `internal/service/plugin`根包只提供窄 facade 给启动层调用
- **AND** `internal/cmd`不得直接导入`plugin/internal/capabilityhost`

#### Scenario: 源码插件消费领域能力

- **WHEN** 源码插件通过 registrar、hook、route 或 jobs callback 获取宿主能力
- **THEN** 插件只能通过`pkg/plugin/pluginhost.Services`获取插件作用域的`capability.Services`和源码插件专用能力
- **AND** 插件业务服务必须继续注入所需的最窄`*cap.Service`、`AdminService`或源码插件专用接口

#### Scenario: 源码插件声明启动期能力

- **WHEN** 源码插件需要声明嵌入文件、生命周期回调、后端钩子、HTTP 路由、内置 Jobs 或治理过滤器
- **THEN** 插件必须通过`pluginhost.NewDeclarations()`创建声明期 facade
- **AND** 通过`RegisterSourcePlugin(plugin pluginhost.Declarations)`注册声明结果
- **AND** 源码插件声明期子 facade 必须使用`*Declarations`命名，例如`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`GovernanceDeclarations`
- **AND** `SourcePlugin*`前缀只用于源码插件读模型、生命周期输入、回调处理器或其他非声明期 facade 契约
- **AND** `pluginhost.Declarations`不得作为运行时领域能力挂载到`pluginhost.Services`

#### Scenario: 动态插件消费领域能力

- **WHEN** 动态插件通过`hostServices`调用普通领域能力
- **THEN** 宿主侧分发必须进入启动期共享的`capability.Services`
- **AND** guest 侧公共入口必须位于`pkg/plugin/pluginbridge`
- **AND** 普通领域 hostcall 代理实现必须位于`pluginbridge/internal/domainhostcall`或等价 internal 子组件
- **AND** 普通领域协议 service 名必须与`capability.Services`领域目录名称保持一致

#### Scenario: 动态插件声明内置 Jobs

- **WHEN** 动态插件需要声明宿主管理的内置定时任务
- **THEN** guest 侧必须通过动态插件声明期对象的`Jobs()`facade 提交`jobs.register`声明
- **AND** 宿主只在 Jobs 发现执行源中接收声明
- **AND** 不得重新引入`cron`领域对象、`CronHostService`或`cron.register`协议

#### Scenario: 动态插件声明启动期能力

- **WHEN** 动态插件需要声明构建期路由分组、内置 Jobs 或后续生命周期声明能力
- **THEN** 插件必须通过`RegisterPlugin(plugin pluginbridge.Declarations)`使用声明期 facade 表达
- **AND** `Declarations.Routes()`和`Declarations.Jobs()`不得作为运行时领域能力挂载到`pluginbridge.Services`
- **AND** 运行时业务服务必须继续通过`pluginbridge.Services`获取普通`*cap.Service`领域能力

#### Scenario: 动态插件消费插件领域能力

- **WHEN** 动态插件通过`pluginbridge.Services.Plugins()`获取插件领域能力
- **THEN** 返回值必须实现`plugincap.Service`
- **AND** `Config()`、`Registry()`、`State()`和`Lifecycle()`必须归属同一个`plugins`领域对象
- **AND** 公共`guest`包不得再声明与`plugincap.Service`平行的`PluginService`接口

## MODIFIED Requirements

### Requirement: 动态插件领域方法必须通过安装授权快照调用

系统 SHALL 要求动态插件在`plugin.yaml hostServices`中声明领域`service + method`，并在安装或启用阶段由宿主确认授权后形成运行时授权快照。运行时调用 MUST 校验授权快照中的领域服务、方法和资源或投影范围。集合型领域的协议 service 名 MUST 与`capability.Services`领域目录名称保持一致，例如用户、文件、任务、通知、插件治理和在线会话领域分别使用`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`。安装授权替代插件级菜单/RBAC 方法校验，但不得替代领域数据权限、租户边界、状态机、数量上限和审计校验。

#### Scenario: 动态插件调用已授权领域读取方法

- **WHEN** 动态插件声明并获得授权调用`service: users`和`method: users.batch_get`
- **THEN** 运行时 host service 分发器允许请求进入`usercap.Service`
- **AND** `usercap.Service`仍按`CapabilityContext`过滤租户、数据权限和可见字段

#### Scenario: 动态插件调用已授权领域管理方法

- **WHEN** 动态插件声明并获得授权调用领域管理方法
- **THEN** 运行时不再额外校验当前用户是否拥有某个工作台菜单或按钮权限
- **AND** 领域命令仍必须校验目标资源可见性、状态机、租户边界、批量上限和审计来源
