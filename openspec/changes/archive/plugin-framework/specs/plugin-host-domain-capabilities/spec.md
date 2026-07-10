# plugin-host-domain-capabilities Specification

## Purpose

定义插件访问宿主业务数据、治理状态和基础设施原语的领域能力总边界，包括领域接口、上下文、授权、数据权限、缓存一致性、`i18n`标签语义和迁移治理。

## Requirements

### Requirement: 插件宿主数据必须通过领域能力访问

系统 SHALL 将宿主核心表、官方能力插件表、宿主`DAO/DO/Entity`、私有缓存快照和宿主内部 service 视为领域 owner 私有实现。源码插件和动态插件访问宿主业务数据、治理状态或跨模块语义时，MUST 通过宿主发布的领域能力接口、`pluginhost.Services`或动态`hostServices`协议完成。

#### Scenario: 插件读取用户基础投影

- **WHEN** 插件需要展示用户名称、头像或状态
- **THEN** 插件调用`usercap`领域投影方法
- **AND** 插件不得生成、导入或查询`sys_user`的`DAO/DO/Entity`

#### Scenario: 插件读取官方能力插件表

- **WHEN** 插件需要读取组织、租户、`AI`、插件状态或其他官方能力插件内部数据
- **THEN** 插件调用对应`orgcap`、`tenantcap`、`ai`、`plugincap`或其他领域能力
- **AND** 插件不得通过动态`data`服务或源码插件`DAO`直接访问官方能力插件表

### Requirement: 领域能力必须定义清晰接口分层

系统 SHALL 为每个宿主领域定义职责明确的接口分层。普通消费面 MUST 使用`Service`表达读取、候选、批量投影、统计、标签解析、校验和低风险动作；管理面 MUST 使用`AdminService`或等价命令接口表达创建、更新、删除、状态变更、授权关系变更和高风险执行动作；宿主内部数据库范围注入 MUST 使用`ScopeService`等内部接口，不得暴露给普通插件。

#### Scenario: 公开能力目录不重复暴露同一领域

- **WHEN** 源码插件或动态插件通过`capability.Services`或`pluginhost.Services`获取宿主能力
- **THEN** 同一领域在普通消费面只能暴露一个稳定入口
- **AND** 写入、删除、状态变更和执行类动作必须通过`Services.Admin().<Domain>()`获取对应`AdminService`
- **AND** 插件自身配置、宿主配置和运行时配置必须使用明确命名区分
- **AND** 通知、会话等领域不得同时暴露旧`contract.*Service`和新领域`*cap.Service`

### Requirement: 领域能力调用必须携带 CapabilityContext

系统 SHALL 为所有插件可见领域能力调用构造`CapabilityContext`。该上下文 MUST 至少包含`pluginID`、actor、tenant、调用来源、系统调用标识、授权快照和审计信息。缺少 actor 的敏感领域调用 MUST 默认拒绝；系统 actor MUST 由宿主创建，插件不得自行伪造。

#### Scenario: 请求型插件调用领域能力

- **WHEN** 登录用户触发插件路由并调用领域能力
- **THEN** 宿主将当前用户、租户、插件`ID`、路由来源和授权快照写入`CapabilityContext`
- **AND** 领域方法基于该上下文执行数据权限、租户和审计治理

#### Scenario: 系统型插件调用领域能力

- **WHEN** 插件在生命周期、hook、provider 回调或定时任务中调用管理领域方法
- **THEN** 宿主必须显式创建系统 actor 并写入`CapabilityContext`

### Requirement: 源码插件管理能力必须通过 AdminServices 目录提供

系统 SHALL 允许源码插件通过`pluginhost.Services.Admin()`获取完整类型化`AdminService`目录。源码插件不需要维护字符串式管理能力授权声明，但`AdminService`方法 MUST 执行`CapabilityContext`解析、租户边界、目标数据边界、状态机、数量上限、系统 actor 和审计治理。

### Requirement: 动态插件领域方法必须通过安装授权快照调用

系统 SHALL 要求动态插件在`plugin.yaml hostServices`中声明领域`service + method`，并在安装或启用阶段由宿主确认授权后形成运行时授权快照。安装授权替代插件级菜单/RBAC 方法校验，但不得替代领域数据权限、租户边界、状态机、数量上限和审计校验。

### Requirement: 插件可见 ID 必须使用领域命名类型

系统 SHALL 在插件可见领域契约中使用领域命名`ID`类型，并在动态协议中统一编码为字符串。领域实现可以在内部将领域`ID`映射为数据库主键、业务键或组合键，但插件契约 MUST NOT 暴露数据库自增主键类型作为长期边界。

### Requirement: 批量读取不得泄露不可见目标原因

系统 SHALL 要求`BatchGet*`类领域读取只返回当前上下文可见的`Items`和不可解释的`MissingIDs`。`MissingIDs` MUST 不区分真实不存在和不可见。命令场景 MUST 使用`Ensure*`类方法，默认任一不可见、不可用或越权目标导致整体失败。

### Requirement: 高频领域方法必须具备有界性能契约

系统 SHALL 为列表、搜索、批量详情、树形数据、下拉候选、聚合统计、标签解析和工作台聚合类领域方法定义分页、数量上限、字段投影、数据库侧过滤和批量装配策略。领域实现 MUST 避免随返回行数、树节点数、插件数、权限项数或关联对象数线性增长的`N+1`查询。

### Requirement: 关键运行时数据缓存必须使用共享修订号和事务后失效

系统 SHALL 对权限、角色关系、用户角色关系、租户成员关系、插件状态、插件资源引用、动态路由、字典、组织树、运行时配置和授权`hostConfig`等关键运行时数据使用共享修订号和事务后失效。单机模式 MAY 使用本地缓存实现，但 MUST 复用同一修订号抽象；集群模式 MUST 接入共享后端、事件广播、分布式缓存或等价协调机制。

### Requirement: 领域能力必须提供稳定 i18n 标签语义

系统 SHALL 要求领域能力默认返回稳定值和`labelKey`。当领域能力需要返回`label`时，MUST 按当前请求 locale 解析，并同时保留`labelKey`。

### Requirement: 插件宿主领域能力迁移必须有治理扫描

系统 SHALL 提供静态治理扫描或等价验证，阻断插件生产代码重新生成、导入或查询宿主核心表，阻断旧领域接口、旧动态`host service`方法和动态`data`服务核心表授权。测试、Mock、安装 SQL 和迁移 SQL 例外 MUST 被限定在对应目录和职责内。

### Requirement: 普通领域能力契约必须与 Provider SPI 分离

系统 SHALL 将插件普通消费领域能力契约与源码插件 provider SPI、宿主内部 scope 接缝分离。普通`capability/<domain>cap`父包 MUST 只暴露普通消费`Service`、领域 DTO、值对象、错误码和常量；凡是需要`*gdb.Model`、`*ghttp.Request`、provider factory、provider runtime、provider env 或宿主内部 scope helper 的接口 MUST 放入对应`*spi`子包或宿主内部包。

#### Scenario: 普通插件消费租户能力

- **WHEN** 源码插件或动态插件通过`tenantcap.Service`消费租户能力
- **THEN** 该父包接口不暴露`*gdb.Model`、`*ghttp.Request`、provider factory 或 provider runtime
- **AND** 插件只看到租户 DTO、状态、批量投影、候选和普通消费方法

#### Scenario: Provider 插件实现租户能力

- **WHEN** 源码 provider 插件需要实现租户解析、membership、scope 或插件表过滤
- **THEN** 它 import `pkg/plugin/capability/tenantcap/tenantspi`
- **AND** 该 SPI 可以使用`*gdb.Model`或`*ghttp.Request`表达宿主内部数据库过滤和请求解析接缝
- **AND** 这些 SPI 不进入动态插件 guest SDK 或`hostServices`协议

### Requirement: 动态路由和 API 文档能力不得暴露 HTTP 框架对象

系统 SHALL 要求普通领域能力契约使用`context.Context`、路径、方法、DTO 或中立值对象传递请求相关信息，不得在普通`capability/**`父包中暴露`*ghttp.Request`或`*ghttp.HandlerItemParsed`。

### Requirement: 数据权限过滤迁移必须保持数据库侧注入语义

系统 SHALL 将租户与组织 scope 接口迁移视为类型归属重构。租户过滤、组织部门过滤、用户 membership 过滤和插件自有表租户过滤 MUST 继续在数据库查询阶段注入约束，不得因为 SPI 拆分退化为内存过滤、放开全量数据或改变拒绝策略。

### Requirement: 插件资源型基础能力必须收敛为领域能力

系统 SHALL 将插件可消费的`cache`、`lock`和`storage`能力发布为`pkg/plugin/capability`下的领域契约。源码插件 MUST 通过`pluginhost.Services`消费这些领域能力；动态插件 MUST 通过`pluginbridge`消费实现同一领域接口的 guest adapter。`pluginbridge`协议和`hostServices`声明只拥有动态插件 transport、授权和 payload 编解码职责，不得成为`cache`、`lock`或`storage`业务接口 owner。

#### Scenario: 源码插件消费资源型基础能力

- **WHEN** 源码插件在 route、hook、jobs 或生命周期回调中需要缓存、锁或对象存储能力
- **THEN** 宿主通过`pluginhost.Services`提供`cachecap.Service`、`lockcap.Service`和`storagecap.Service`
- **AND** 插件业务服务应注入所需的最窄领域接口
- **AND** 插件不得接收宿主内部`kvcache.Service`、`hostlock.Service`、存储 provider、物理路径或底层客户端

#### Scenario: 动态插件消费资源型基础能力

- **WHEN** 动态插件业务代码调用`guest.Services.Cache()`、`guest.Services.Lock()`或`guest.Services.Storage()`
- **THEN** guest 侧返回值必须实现对应`cachecap.Service`、`lockcap.Service`或`storagecap.Service`
- **AND** 公共 guest API 不得向业务代码暴露`protocol.HostServiceCacheValue`、`protocol.HostServiceLockAcquireResponse`、`protocol.HostServiceStorageObject`或等价 transport DTO 作为领域返回值

### Requirement: 源码插件资源能力默认全信任但必须作用域隔离

系统 SHALL 将源码插件视为可信插件形态，源码插件消费`cache`、`lock`和`storage`时不需要在`plugin.yaml hostServices`中声明资源边界。即便源码插件默认全信任，领域服务 MUST 仍按当前插件 ID 和租户上下文隔离内部 cache key、lock name 和 storage object key。

### Requirement: WASM 资源能力配置必须复用领域能力目录

系统 SHALL 要求动态插件`cache`、`lock`和`storage`分发复用启动期注入的同一个`capability.Services`目录。WASM 运行时 MUST NOT 继续发布或使用`ConfigureCacheHostService`、`ConfigureLockHostService`、`ConfigureStorageHostService`或等价的资源能力专用底层配置入口。

### Requirement: core-owned 与 plugin-owned 领域能力分类

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
