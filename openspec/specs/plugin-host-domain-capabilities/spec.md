# plugin-host-domain-capabilities Specification

## Purpose
TBD - created by archiving change add-plugin-host-domain-capabilities. Update Purpose after archive.
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
- **AND** 插件自身配置、宿主配置和运行时配置必须使用明确命名区分，避免`Config`、`RuntimeConfig`和`HostConfig`同时表达同一配置领域
- **AND** 通知、会话等领域不得同时暴露旧`contract.*Service`和新领域`*cap.Service`

#### Scenario: 普通插件消费领域能力

- **WHEN** 普通源码插件或动态插件调用领域能力
- **THEN** 它只能获得领域`Service`或对应 guest 代理
- **AND** 返回值使用领域`DTO`、值对象、批量投影或结构化响应
- **AND** 不返回`*gdb.Model`、`DAO`、`DO`、`Entity`、HTTP 请求对象或内部 service 实例

#### Scenario: 宿主内部注入数据范围

- **WHEN** 宿主领域适配器需要在数据库查询阶段注入租户、组织或数据权限过滤
- **THEN** 它使用宿主内部`ScopeService`或等价窄接口
- **AND** 该接口不得通过`capability.Services`、`pluginhost.Services`普通消费面或动态 guest 目录暴露给插件

### Requirement: 领域能力调用必须携带 CapabilityContext

系统 SHALL 为所有插件可见领域能力调用构造`CapabilityContext`。该上下文 MUST 至少包含`pluginID`、actor、tenant、调用来源、系统调用标识、授权快照和审计信息。缺少 actor 的敏感领域调用 MUST 默认拒绝；系统 actor MUST 由宿主创建，插件不得自行伪造。

#### Scenario: 请求型插件调用领域能力

- **WHEN** 登录用户触发插件路由并调用领域能力
- **THEN** 宿主将当前用户、租户、插件`ID`、路由来源和授权快照写入`CapabilityContext`
- **AND** 领域方法基于该上下文执行数据权限、租户和审计治理

#### Scenario: 系统型插件调用领域能力

- **WHEN** 插件在生命周期、hook、provider 回调或定时任务中调用管理领域方法
- **THEN** 宿主必须显式创建系统 actor 并写入`CapabilityContext`
- **AND** 领域方法不得把缺少用户上下文的调用自动提升为全量权限

### Requirement: 源码插件管理能力必须通过 AdminServices 目录提供

系统 SHALL 允许源码插件通过`pluginhost.Services.Admin()`获取完整类型化`AdminService`目录。源码插件不需要维护字符串式管理能力授权声明，但`AdminService`方法 MUST 执行`CapabilityContext`解析、租户边界、目标数据边界、状态机、数量上限、系统 actor 和审计治理。

#### Scenario: 源码插件调用管理方法

- **WHEN** 源码插件需要执行用户状态变更、授权关系变更、插件治理或其他管理动作
- **THEN** 插件通过`pluginhost.Services.Admin().<Domain>()`获取对应`AdminService`
- **AND** 方法调用进入领域 owner 的命令实现
- **AND** 实现记录插件`ID`、actor、来源、目标资源、结果状态和审计原因

#### Scenario: 源码插件业务服务保存依赖

- **WHEN** 源码插件业务服务只需要用户展示投影
- **THEN** 其构造函数只接收`usercap.Service`等最窄领域接口
- **AND** 不得为了局部调用长期保存完整`pluginhost.Services`目录

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

### Requirement: 插件可见 ID 必须使用领域命名类型

系统 SHALL 在插件可见领域契约中使用领域命名`ID`类型，并在动态协议中统一编码为字符串。领域实现可以在内部将领域`ID`映射为数据库主键、业务键或组合键，但插件契约 MUST NOT 暴露数据库自增主键类型作为长期边界。

#### Scenario: 动态插件传递用户 ID

- **WHEN** 动态插件调用`usercap`批量读取用户投影
- **THEN** 请求载荷中的用户标识使用字符串编码的`usercap.UserID`
- **AND** 宿主适配器内部完成领域`ID`到`sys_user`主键或业务键的解析

#### Scenario: 源码插件编译期使用领域 ID

- **WHEN** 源码插件调用`tenantcap`、`plugincap`或`filecap`
- **THEN** 方法签名使用对应领域的命名`ID`类型
- **AND** 插件代码不得依赖数据库整型主键进行跨领域组合

### Requirement: 批量读取不得泄露不可见目标原因

系统 SHALL 要求`BatchGet*`类领域读取只返回当前上下文可见的`Items`和不可解释的`MissingIDs`。`MissingIDs` MUST 不区分真实不存在和不可见。命令场景 MUST 使用`Ensure*`类方法，默认任一不可见、不可用或越权目标导致整体失败，除非领域规范显式定义逐项处理语义、响应结构和审计方式。

#### Scenario: 批量读取包含不可见用户

- **WHEN** 插件批量读取多个用户，其中部分用户超出当前租户或数据权限范围
- **THEN** 响应只返回可见用户投影
- **AND** 不可见用户进入`MissingIDs`
- **AND** 响应不得区分这些用户是不存在还是不可见

#### Scenario: 命令校验包含不可见目标

- **WHEN** 插件调用`EnsureVisible`或等价命令前置校验
- **THEN** 任一不可见目标导致整体失败
- **AND** 错误使用结构化业务错误和审计摘要表达

### Requirement: 高频领域方法必须具备有界性能契约

系统 SHALL 为列表、搜索、批量详情、树形数据、下拉候选、聚合统计、标签解析和工作台聚合类领域方法定义分页、数量上限、字段投影、数据库侧过滤和批量装配策略。领域实现 MUST 避免随返回行数、树节点数、插件数、权限项数或关联对象数线性增长的`N+1`查询。

#### Scenario: 插件列表装配创建人标签

- **WHEN** 插件列表需要展示当前页记录的创建人标签
- **THEN** 插件先收集当前页创建人领域`ID`
- **AND** 调用`usercap.BatchGet`一次性获取可见投影
- **AND** 不得循环调用单项用户详情方法

#### Scenario: 插件读取树形候选

- **WHEN** 插件读取部门树、权限树或任务分组树
- **THEN** 请求必须包含有界根节点、深度、分页、懒加载或最大节点数策略
- **AND** 宿主在数据库查询阶段应用租户和数据权限过滤

### Requirement: 关键运行时数据缓存必须使用共享修订号和事务后失效

系统 SHALL 对权限、角色关系、用户角色关系、租户成员关系、插件状态、插件资源引用、动态路由、字典、组织树、运行时配置和授权`hostConfig`等关键运行时数据使用共享修订号和事务后失效。单机模式 MAY 使用本地缓存实现，但 MUST 复用同一修订号抽象；集群模式 MUST 接入共享后端、事件广播、分布式缓存或等价协调机制。

#### Scenario: 角色授权关系变更

- **WHEN** 领域管理方法成功替换角色权限或用户角色关系
- **THEN** 宿主在事务提交成功后推进权限相关共享修订号
- **AND** 所有实例能够基于修订号失效或重建权限快照

#### Scenario: 缓存后端不可用

- **WHEN** 领域投影读取发现缓存后端不可用
- **THEN** 领域实现按规范回源重建或返回明确的能力不可用错误
- **AND** 不得在集群模式下退化为仅当前节点可见的本地状态

### Requirement: 领域能力必须提供稳定 i18n 标签语义

系统 SHALL 要求领域能力默认返回稳定值和`labelKey`。当领域能力需要返回`label`时，MUST 按当前请求 locale 解析，并同时保留`labelKey`。字典、菜单、权限、插件资源、状态和错误消息等用户可见文本 MUST 遵守宿主或插件的`i18n`启用边界。

#### Scenario: 字典标签解析

- **WHEN** 插件调用`dictcap.ResolveLabels`
- **THEN** 响应返回字典值、稳定`labelKey`和按当前 locale 解析的可选`label`
- **AND** 插件可使用`labelKey`在前端或自身运行时继续解析展示文案

#### Scenario: 未启用 i18n 的插件消费领域标签

- **WHEN** 未启用`i18n`的插件消费宿主领域能力
- **THEN** 宿主仍按宿主 locale 和领域资源返回稳定`labelKey`
- **AND** 不要求该插件补齐自身`manifest/i18n`资源

### Requirement: 插件宿主领域能力迁移必须有治理扫描

系统 SHALL 提供静态治理扫描或等价验证，阻断插件生产代码重新生成、导入或查询宿主核心表，阻断旧领域接口、旧动态`host service`方法和动态`data`服务核心表授权。测试、Mock、安装 SQL 和迁移 SQL 例外 MUST 被限定在对应目录和职责内。

#### Scenario: 生产代码引用宿主核心表 DAO

- **WHEN** 插件生产 Go 代码导入或调用`dao.SysUser`、`dao.SysDictData`、`shared.TableSysUser`或等价宿主核心表入口
- **THEN** 治理扫描失败
- **AND** 变更必须迁移到对应领域能力或记录受控非生产例外

#### Scenario: 动态插件声明核心表 data 授权

- **WHEN** 动态插件 manifest 或授权快照声明`data`服务访问`sys_user`、`sys_role`、`sys_plugin`或其他核心表
- **THEN** 构建、安装或启用校验失败
- **AND** 宿主不得把该声明写入运行时授权快照

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

#### Scenario: Provider 插件实现组织能力

- **WHEN** 源码 provider 插件需要实现组织 scope 或 provider runtime
- **THEN** 它 import `pkg/plugin/capability/orgcap/orgspi`
- **AND** `orgspi.ProviderEnv`可以引用`tenantspi.PluginTableFilterService`
- **AND** 普通`orgcap.Service`不暴露`*gdb.Model`或 provider SPI

### Requirement: 动态路由和 API 文档能力不得暴露 HTTP 框架对象

系统 SHALL 要求普通领域能力契约使用`context.Context`、路径、方法、DTO 或中立值对象传递请求相关信息，不得在普通`capability/**`父包中暴露`*ghttp.Request`或`*ghttp.HandlerItemParsed`。

#### Scenario: 动态路由元数据读取

- **WHEN** 插件通过`routecap.Service.DynamicRouteMetadata`读取当前动态路由元数据
- **THEN** 方法签名接收`context.Context`
- **AND** 宿主适配器在内部从上下文恢复 HTTP 请求并读取元数据
- **AND** `routecap`父包不 import `ghttp`

#### Scenario: API 文档 operation key 派生

- **WHEN** 插件或源码插件中间件需要派生 API 文档 operation key
- **THEN** 它使用不依赖`ghttp`的路径和方法派生入口
- **AND** 需要`*ghttp.HandlerItemParsed`的宿主私有逻辑不得放入`apidoccap`普通契约包

### Requirement: 数据权限过滤迁移必须保持数据库侧注入语义

系统 SHALL 将租户与组织 scope 接口迁移视为类型归属重构。租户过滤、组织部门过滤、用户 membership 过滤和插件自有表租户过滤 MUST 继续在数据库查询阶段注入约束，不得因为 SPI 拆分退化为内存过滤、放开全量数据或改变拒绝策略。

#### Scenario: 宿主列表查询使用迁移后的租户 scope

- **WHEN** 宿主列表、下拉候选或批量读取路径需要按当前租户过滤
- **THEN** 调用方使用`tenantspi.ScopeService`或等价窄接口在`*gdb.Model`上追加数据库过滤
- **AND** 不得先读取全量结果再在内存中过滤
- **AND** 过滤前后的租户可见性语义保持不变

#### Scenario: 宿主列表查询使用迁移后的组织 scope

- **WHEN** 宿主列表、下拉候选或批量读取路径需要按部门或用户组织范围过滤
- **THEN** 调用方使用`orgspi.ScopeService`或等价窄接口在`*gdb.Model`上追加数据库过滤
- **AND** 不得通过关联名称、分页总数或候选项泄露范围外数据存在性
- **AND** 过滤前后的组织数据权限语义保持不变

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
- **AND** 源码插件声明期子 facade 必须使用`*Declarations`命名，例如`AssetDeclarations`、`LifecycleDeclarations`、`HookDeclarations`、`HTTPDeclarations`、`JobDeclarations`和`AccessDeclarations`
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

### Requirement: 插件资源型基础能力必须收敛为领域能力

系统 SHALL 将插件可消费的`cache`、`lock`和`storage`能力发布为`pkg/plugin/capability`下的领域契约。源码插件 MUST 通过`pluginhost.Services`消费这些领域能力；动态插件 MUST 通过`pluginbridge/guest`消费实现同一领域接口的 guest adapter。`pluginbridge`协议和`hostServices`声明只拥有动态插件 transport、授权和 payload 编解码职责，不得成为`cache`、`lock`或`storage`业务接口 owner。

#### Scenario: 源码插件消费资源型基础能力

- **WHEN** 源码插件在 route、hook、jobs 或生命周期回调中需要缓存、锁或对象存储能力
- **THEN** 宿主通过`pluginhost.Services`提供`cachecap.Service`、`lockcap.Service`和`storagecap.Service`
- **AND** 插件业务服务应注入所需的最窄领域接口
- **AND** 插件不得接收宿主内部`kvcache.Service`、`hostlock.Service`、存储 provider、物理路径或底层客户端

#### Scenario: 动态插件消费资源型基础能力

- **WHEN** 动态插件业务代码调用`guest.Services.Cache()`、`guest.Services.Lock()`或`guest.Services.Storage()`
- **THEN** guest 侧返回值必须实现对应`cachecap.Service`、`lockcap.Service`或`storagecap.Service`
- **AND** 公共 guest API 不得向业务代码暴露`protocol.HostServiceCacheValue`、`protocol.HostServiceLockAcquireResponse`、`protocol.HostServiceStorageObject`或等价 transport DTO 作为领域返回值

#### Scenario: 动态插件资源能力经过授权后进入领域服务

- **WHEN** 动态插件通过`hostServices`调用`cache`、`lock`或`storage`方法
- **THEN** WASM host service 分发器必须先校验能力、方法和资源授权快照
- **AND** 校验通过后必须通过`capabilityServicesForHostCall`获取插件作用域领域服务
- **AND** 分发器不得直接调用底层缓存、锁或存储实现

### Requirement: 源码插件资源能力默认全信任但必须作用域隔离

系统 SHALL 将源码插件视为可信插件形态，源码插件消费`cache`、`lock`和`storage`时不需要在`plugin.yaml hostServices`中声明资源边界。即便源码插件默认全信任，领域服务 MUST 仍按当前插件 ID 和租户上下文隔离内部 cache key、lock name 和 storage object key。

#### Scenario: 源码插件不声明资源边界

- **WHEN** 源码插件未在`plugin.yaml hostServices`中声明`cache`、`lock`或`storage`资源
- **THEN** 宿主仍允许该源码插件通过`pluginhost.Services`调用对应领域服务
- **AND** 宿主不得要求源码插件经过动态插件安装授权确认

#### Scenario: 源码插件资源作用域隔离

- **WHEN** 两个源码插件使用相同 logical cache namespace、lock name 或 storage path
- **THEN** 宿主为两个插件生成不同的内部资源身份
- **AND** 一个源码插件不得读取、续租、释放或删除另一个源码插件的资源

### Requirement: WASM 资源能力配置必须复用领域能力目录

系统 SHALL 要求动态插件`cache`、`lock`和`storage`分发复用启动期注入的同一个`capability.Services`目录。WASM 运行时 MUST NOT 继续发布或使用`ConfigureCacheHostService`、`ConfigureLockHostService`、`ConfigureStorageHostService`或等价的资源能力专用底层配置入口。

#### Scenario: 启动期配置动态资源能力

- **WHEN** 宿主启动期配置动态插件 WASM host services
- **THEN** 启动装配只需要为`cache`、`lock`和`storage`调用`ConfigureDomainHostServices(capability.Services)`
- **AND** 不再额外注入`kvcache.Service`、`hostlock.Service`或 storage config reader 到 WASM 分发层

#### Scenario: 治理扫描检查专用配置入口

- **WHEN** 生产代码新增`ConfigureCacheHostService`、`ConfigureLockHostService`、`ConfigureStorageHostService`或等价专用入口
- **THEN** 治理验证必须失败
- **AND** 变更必须改为通过领域能力目录配置资源能力

### Requirement: Files 领域必须只表示宿主文件中心投影

系统 SHALL 将`Files()`领域能力限定为宿主文件中心资源投影和可见性校验能力。`Files()`方法 MUST 基于宿主文件领域 owner 的数据权限、租户边界和存在性不泄露策略返回文件投影或执行可见性确认，不得承担插件私有对象存储的内容读写生命周期。

#### Scenario: 插件读取宿主文件投影

- **WHEN** 源码插件或动态插件需要展示用户已上传到宿主文件中心的文件名称、MIME、大小或业务场景
- **THEN** 插件必须调用`Files().BatchGet`读取当前上下文可见的文件投影
- **AND** 响应必须使用`filecap.FileProjection`等领域 DTO
- **AND** 响应不得向插件暴露宿主文件中心`DAO`、`DO`、`Entity`、本地绝对路径或存储 provider 私有 key

#### Scenario: 插件校验宿主文件可见性

- **WHEN** 插件业务命令引用一批宿主文件中心文件 ID
- **THEN** 插件必须在命令执行前调用`Files().EnsureVisible`或等价领域校验
- **AND** 任一文件不存在、不可见或越过租户和数据权限边界时，命令必须整体拒绝
- **AND** 错误不得区分目标文件是真实不存在还是当前调用方不可见

#### Scenario: 动态插件声明文件领域服务

- **WHEN** 动态插件需要读取或校验宿主文件中心投影
- **THEN** 插件必须在`plugin.yaml hostServices`中声明`service: files`和所需文件领域方法
- **AND** 该声明不得使用`resources.paths`表达对象存储路径
- **AND** 宿主分发必须进入`filecap.Service`，不得转发到`storagecap.Service`

#### Scenario: 插件私有附件不使用 Files

- **WHEN** 插件需要保存、下载、删除、列出或清理插件私有附件对象
- **THEN** 插件必须使用`Storage()`领域能力
- **AND** 插件不得通过`Files()`领域方法把插件私有对象伪装为宿主文件中心资源

### Requirement: 插件领域能力 Go 方法命名必须避免重复领域名

系统 SHALL 为 `pkg/plugin/capability/*cap` 下的公开 Go 接口使用动作式命名。当 `Service` 所在目录已经表达主资源时，主资源方法 MUST 省略重复领域名，优先使用 `BatchGet`、`Search`、`Current`、`EnsureVisible`、`Delete`、`Run`、`Revoke`、`SetStatus`、`SetEnabled` 等短名。仅当方法目标是子资源、复合对象或短名会造成歧义时，才允许保留限定词，例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource`。动态 `host service` 的 wire method 名称 MAY 保持显式原样，不受本规则影响。

#### Scenario: 主资源方法使用动作名

- **WHEN** 插件通过 `usercap.Service`、`filecap.Service`、`sessioncap.Service` 或 `plugincap.Service` 调用主资源能力
- **THEN** 方法名 SHOULD 直接表达动作，而不是重复资源名
- **AND** 例如 `Users().Search`、`Files().EnsureVisible`、`Sessions().BatchGet`、`Plugins().BatchGet` 这类命名方式是允许的

#### Scenario: 子资源方法保留限定词

- **WHEN** 方法目标是权限、租户关系、部门树、来源维度或其他复合对象
- **THEN** 方法名 MAY 保留限定词以避免歧义
- **AND** 例如 `BatchGetPermissions`、`ListUserTenants`、`SearchDepartments`、`DeleteBySource` 这类命名方式仍然符合规范

#### Scenario: 动态 wire method 名称保持显式

- **WHEN** 动态插件通过 `pluginbridge` 或 WASM host service 调用宿主领域能力
- **THEN** wire method 名称 SHOULD 继续保持显式资源名，例如 `users.batch_get`、`messages.batch_get`
- **AND** 该 wire 命名不要求与 typed Go 方法名完全一致

