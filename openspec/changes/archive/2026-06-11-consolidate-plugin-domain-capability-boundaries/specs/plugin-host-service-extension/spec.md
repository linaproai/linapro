## REMOVED Requirements

### Requirement: 源码插件宿主服务适配器必须归属 plugin 内部 hostservices 子组件

**Reason**: `hostservices`命名会和动态插件`plugin.yaml hostServices`协议混淆，不能准确表达该组件实际负责宿主侧`capability.Services`实现。

**Migration**: 使用新增的`源码插件宿主服务适配器必须归属 plugin 内部 capabilityhost 子组件`要求；实现从`internal/service/plugin/internal/hostservices`迁移到`internal/service/plugin/internal/capabilityhost`，根 facade 和运行期语义保持不变。

## ADDED Requirements

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

系统 SHALL 要求动态插件普通领域`hostServices.service`协议名与`pkg/plugin/capability.Services`领域目录名称保持一致。集合型领域 MUST 使用复数领域名：`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`。对应能力字符串 MUST 分别使用`host:users`、`host:files`、`host:jobs`、`host:notifications`、`host:plugins`和`host:sessions`。项目不保留旧单数 service 别名。

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

- **WHEN** 动态插件声明`auth`、`authz`、`apidoc`、`bizctx`、`dict`、`i18n`、`infra`、`route`、`ai`、`org`或`tenant`
- **THEN** service 继续使用该领域命名空间名称
- **AND** 不得为了形式统一将其机械复数化

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

## MODIFIED Requirements

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
