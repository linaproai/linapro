## MODIFIED Requirements

### Requirement: 领域能力边界必须具有固定归属

系统 SHALL 将插件可消费领域能力固定为两类归属。core-owned 能力由`lina-core/pkg/plugin/capability`拥有契约，由真实宿主领域 owner 或官方基础能力适配器实现；plugin-owned 能力由对应 owner 插件的`backend/cap/<domain>cap`拥有契约、SDK、SPI 和版本策略。`pkg/plugin/pluginhost`只拥有源码插件声明和运行期服务目录接入，`pkg/plugin/pluginbridge`只拥有动态插件 ABI、transport、通用 host service envelope 和动态插件公共入口。`internal/service/plugin/internal/capabilityhost`和 WASM host service 只承担标准业务上下文桥接、动态授权、编解码、通用转发和错误映射职责，不得成为跨领域业务实现 owner。任何新增领域能力 MUST 先按归属矩阵进入 core-owned 或 plugin-owned 契约，再由真实 owner 实现和动态 transport 适配，不得在`WASM`分发、`pluginbridge`协议目录或动态插件公共 SDK 中单独定义一套平行业务接口。

#### Scenario: 新增宿主内核领域能力契约

- **WHEN** 系统新增一个插件运行、隔离、授权、资源访问或治理必需的宿主内核领域能力
- **THEN** 领域接口、`DTO`、领域`ID`和降级语义必须定义在`pkg/plugin/capability/<domain>cap`或等价 core-owned 领域命名空间
- **AND** `pluginbridge`不得成为该领域业务接口的 owner

#### Scenario: 新增插件拥有领域能力契约

- **WHEN** 系统新增一个非核心、变化快、由插件拥有实现的领域能力
- **THEN** 领域接口、`DTO`、领域`ID`、错误语义、动态 SDK 和 provider SPI 必须定义在 owner 插件的`backend/cap/<domain>cap`
- **AND** core 只通过通用 descriptor、依赖治理、授权快照和动态路由识别该能力
- **AND** core 不得新增该领域专属`*cap`包、provider facade、wire codec 或 dispatcher 分支

#### Scenario: 宿主实现 core-owned 领域能力

- **WHEN** 宿主启动期构造`capability.Services`
- **THEN** core-owned 能力的具体业务实现必须归属真实领域 owner 或 owner 发布的稳定适配契约
- **AND** `internal/service/plugin/internal/capabilityhost`只允许作为动态调用薄适配层
- **AND** `capabilityhost`不得直接访问其他领域`DAO`、`DO`、`Entity`、私有缓存、私有 provider 或内部 helper
- **AND** `internal/cmd`不得直接导入`plugin/internal/capabilityhost`

#### Scenario: owner 插件实现 plugin-owned 领域能力

- **WHEN** owner 插件实现 plugin-owned 领域能力
- **THEN** 业务逻辑、provider adapter、模型路由、调用日志和外部协议适配必须保留在该插件`backend/internal/service`或职责明确的内部包
- **AND** 其他插件只能依赖该插件`backend/cap/...`公开契约
- **AND** 其他插件不得 import 该插件`backend/internal`、`dao`、`do`、`entity`、controller 或私有缓存结构

#### Scenario: 源码插件消费领域能力

- **WHEN** 源码插件通过 registrar、hook、route 或 jobs callback 获取领域能力
- **THEN** core-owned 能力只能通过`pkg/plugin/pluginhost.Services`获取插件作用域的统一`capability.Services`和源码插件专用能力
- **AND** plugin-owned 能力必须通过 owner 插件公开 helper、显式注入的 owner 契约接口或受治理 capability descriptor 引用获取
- **AND** 插件业务服务必须继续注入所需的最窄`*cap.Service`、owner 契约接口或源码插件专用接口
- **AND** 不得再注入或保存`AdminService`

#### Scenario: 源码插件声明启动期能力

- **WHEN** 源码插件需要声明嵌入文件、生命周期回调、后端钩子、HTTP 路由、内置 Jobs、治理过滤器或 plugin-owned provider descriptor
- **THEN** 插件必须通过`pluginhost.NewDeclarations()`创建声明期 facade
- **AND** 通过`RegisterSourcePlugin(plugin pluginhost.Declarations)`注册声明结果
- **AND** 源码插件声明期子 facade 必须使用`*Declarations`命名，例如`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`AccessDeclarations`
- **AND** 非核心领域 provider 不得通过`pluginhost`新增领域专属`Provide<Domain>`方法，而必须由 owner helper 生成通用 descriptor
- **AND** `pluginhost.Declarations`不得作为运行时领域能力挂载到`pluginhost.Services`

#### Scenario: 动态插件消费领域能力

- **WHEN** 动态插件通过`hostServices`调用普通领域能力
- **THEN** 宿主侧分发必须先校验动态 registry、owner descriptor、授权快照和资源范围，再进入启动期共享的 core-owned 能力服务或 owner 插件注册的 plugin-owned handler
- **AND** guest 侧公共入口必须位于`pkg/plugin/pluginbridge`或 owner 插件公开的`backend/cap/<domain>cap/bridge`
- **AND** 普通领域 hostcall 代理实现必须位于`pluginbridge/internal/domainhostcall`、owner 插件 bridge SDK 或等价 internal 子组件
- **AND** 框架领域的可用性与诊断状态必须通过对应 owner 领域读取，不得由`plugins`领域聚合`org`、`tenant`或`AI`状态
- **AND** core-owned 普通领域协议 service 名必须与`capability.Services`领域目录名称保持一致
- **AND** plugin-owned 普通领域协议必须同时包含`owner`、`service`、`version`和`method`

#### Scenario: 动态插件声明内置 Jobs

- **WHEN** 动态插件需要声明宿主管理的内置定时任务
- **THEN** guest 侧必须通过动态插件声明期对象的`Jobs()`facade 提交`jobs.register`声明
- **AND** 宿主只在 Jobs 发现执行源中接收声明
- **AND** `jobs.register`不得作为运行期`jobcap.Service`方法暴露给源码插件或动态插件业务服务
- **AND** 不得重新引入`cron`领域对象、`CronHostService`或`cron.register`协议

#### Scenario: 动态插件声明启动期能力

- **WHEN** 动态插件需要声明构建期路由分组、内置 Jobs、owner 能力申请或后续生命周期声明能力
- **THEN** 插件必须通过`RegisterPlugin(plugin pluginbridge.Declarations)`和 manifest/构建产物表达声明
- **AND** `Declarations.Routes()`和`Declarations.Jobs()`不得作为运行时领域能力挂载到`pluginbridge.Services`
- **AND** 运行时业务服务必须继续通过`pluginbridge.Services`获取普通 core-owned 能力，通过 owner bridge SDK 获取 plugin-owned 能力

#### Scenario: 动态插件消费插件领域能力

- **WHEN** 动态插件通过`pluginbridge.Services.Plugins()`获取插件领域能力
- **THEN** 返回值必须实现`plugincap.Service`
- **AND** `Config()`、`Registry()`和`State()`必须归属同一个`plugins`领域对象
- **AND** `Lifecycle()`和状态读取治理必须归属同一个`plugins`领域对象，动态是否可用由 registry 注册事实和授权快照决定
- **AND** 公共`guest`包不得再声明与`plugincap.Service`平行的`PluginService`接口
