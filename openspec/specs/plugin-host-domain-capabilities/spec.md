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

### Requirement:批量读取不得泄露不可见目标原因

系统 SHALL 为插件可见目录的可见性校验统一使用单一`EnsureVisible(ctx, ids []...) error`方法。普通消费面 MUST NOT 同时暴露单个可见性方法和批量可见性方法；如果调用场景只有一个 ID，也应通过单一批量签名传入长度为 1 的切片。该约束旨在让单条校验、列表装配和批量拒绝语义保持一致，避免同一目录出现并列的单条/批量接口形态。

#### Scenario: 定义组织或租户目录可见性校验

- **WHEN** 系统定义组织、租户或其他插件可见目录的可见性校验
- **THEN** 目录接口 MUST 只暴露一个`EnsureVisible`方法
- **AND** 该方法的参数 MUST 为 ID 切片
- **AND** 不得再新增`EnsureVisibleMany`入口

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

- **WHEN** 插件通过`routecap.Service.GetMetadata`读取当前动态路由元数据
- **THEN** 方法签名接收`context.Context`
- **AND** 返回值使用`routecap.Metadata`
- **AND** 宿主适配器在内部从上下文恢复 HTTP 请求并读取元数据
- **AND** `routecap`父包不 import `ghttp`

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
- **AND** 响应必须使用`filecap.FileInfo`等领域 DTO
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

### Requirement: 阶段一后续候选与搜索能力必须有界发布
系统 SHALL 补齐阶段 1.5 的普通领域候选、搜索和可见性能力，包括扩展`Users.Search`稳定过滤、新增`Dict.ListValues`、`Files.Search`、`Jobs.Search`、`Jobs.EnsureVisible`、`Sessions.BatchGetUserOnlineStatus`和`Sessions.EnsureVisible`。这些方法 MUST 定义分页或数量上限、稳定排序、数据库侧租户和数据权限过滤、动态授权资源和超限错误。

#### Scenario: 用户搜索使用稳定过滤
- **WHEN** 插件按关键字、状态、租户或启用状态搜索用户候选
- **THEN** 系统在数据库查询阶段应用租户和数据权限过滤
- **AND** 响应只返回有界用户投影，不诱导逐项用户详情补查

#### Scenario: 字典候选分页
- **WHEN** 插件按字典类型读取字典值候选
- **THEN** 系统按类型、租户覆盖、状态和排序返回有界分页或 limit 结果
- **AND** 不得无上限返回全部字典值

#### Scenario: 在线状态批量读取
- **WHEN** 插件批量判断多个用户是否在线
- **THEN** 系统使用 session owner 的批量路径或投影查询完成
- **AND** 不得对每个用户执行一次在线会话查询

### Requirement: 组织租户和插件治理投影必须安全降级
系统 SHALL 补齐阶段 2 的`Org`、`Tenant`和`Plugins`投影能力。组织能力 MUST 支持批量用户组织档案、受限部门树、部门搜索、岗位候选和部门/岗位可见性校验；租户能力 MUST 支持当前租户详情、批量租户投影、租户搜索、批量用户租户列表和批量租户可见性校验；插件治理能力 MUST 支持当前插件投影、插件搜索、分页租户插件列表和能力状态批量读取。Provider 缺失或禁用时 MUST 返回规范定义的空投影、空页、不可用错误或中性状态，不得放开全量数据。多个同一 singleton 能力 provider 同时可服务时 MUST 返回`CodeCapabilityProviderConflict`，能力状态 reason MUST 为`provider_conflict`。

#### Scenario: 组织 provider 禁用
- **WHEN** 插件读取用户组织档案且组织 provider 未启用
- **THEN** 系统返回空组织档案或明确能力不可用结果
- **AND** 不得通过宿主内部组织表或未授权 provider 数据兜底

#### Scenario: 领域 provider 多实例冲突
- **WHEN** 同一 singleton 领域能力存在多个可服务 provider
- **THEN** 系统返回`CodeCapabilityProviderConflict`
- **AND** 能力状态 reason 为`provider_conflict`

#### Scenario: 租户批量投影
- **WHEN** 插件按租户 ID 集合读取租户投影
- **THEN** 系统只返回当前 actor 可见的租户
- **AND** 不存在、不可见和租户外目标统一进入`MissingIDs`

#### Scenario: 插件治理分页搜索
- **WHEN** 插件按插件 ID、名称、类型或启用状态搜索插件
- **THEN** 系统使用插件治理读模型、缓存快照或集合化查询返回分页结果
- **AND** 不得为每个插件重复扫描 manifest 或重复读取启用状态

### Requirement: 插件私有资源批量能力必须保持插件和租户作用域
系统 SHALL 补齐阶段 3 的插件私有资源批量能力，包括`Storage.BatchStat`、`Storage.ListCursor`、`Storage.DeleteMany`、`Cache.GetMany`、`Cache.SetMany`、`Cache.DeleteMany`、`Manifest.GetMany`、`Manifest.List`以及动态 runtime state 多键读写删。所有方法 MUST 受当前插件 ID、租户、资源授权、路径或 key 数量上限约束，并复用既有共享后端或 owner 实例。

#### Scenario: Storage 批量元数据读取
- **WHEN** 动态插件批量读取多个私有对象元数据
- **THEN** 系统只返回当前插件和租户作用域下且资源授权允许的路径
- **AND** 不得暴露宿主物理路径或 provider 私有 key

#### Scenario: Cache 多键写入
- **WHEN** 插件批量写入缓存键
- **THEN** 系统复用既有缓存后端和命名空间隔离
- **AND** 缓存仍为非权威加速数据，不改变权限、配置或业务记录权威来源

#### Scenario: Runtime state 多键删除
- **WHEN** 动态插件批量删除运行态 key
- **THEN** 系统只删除当前插件和租户作用域下的 key
- **AND** 删除不存在 key 不得泄露其他插件状态空间

### Requirement: 通知和 AI 状态能力必须类型化并可批量降级
系统 SHALL 补齐阶段 4 的通知类型化和`AI`状态能力。通知读取 MUST 返回稳定类型化投影，支持按业务来源批量读取和可见性校验；`AI`能力 MUST 支持文本方法级状态和跨子能力方法状态批量读取。结果不得暴露 provider 密钥、模型映射、供应商配置或通知内部存储结构。

#### Scenario: 通知按来源批量读取
- **WHEN** 插件按`SourceType + SourceIDs`批量读取消息
- **THEN** 系统返回当前 actor 可见的类型化消息投影
- **AND** 不得用`map[string]any`暴露未治理字段

#### Scenario: AI 方法状态批量读取
- **WHEN** 插件批量读取多个`AI`子能力方法状态
- **THEN** 系统返回可用性、禁用原因或能力不可用的结构化状态
- **AND** 不返回 provider 配置、API key 或模型路由内部细节

### Requirement: 剩余阶段完成记录必须覆盖影响和验证证据
系统 SHALL 在任务记录和审查结论中证明阶段 1.5 至阶段 5 全部已实现或被规范明确排除，并记录`i18n`、缓存一致性、数据权限、数据库、开发工具、测试和 E2E 的影响判断。

#### Scenario: 完整方案完成审计
- **WHEN** 本变更准备标记完成
- **THEN** 任务记录必须逐项列出`localdocs`阶段 1.5 至阶段 5 的实现、排除或延后依据
- **AND** 验证证据必须覆盖 OpenSpec strict、目标 Go 测试、启动绑定 smoke、静态边界检索和`lina-review`

### Requirement: 插件领域能力扩展必须先冻结阶段矩阵

系统 SHALL 在新增或动态发布插件领域能力方法前冻结方法发布、错误语义、规模上限和授权资源四类矩阵。未完成冻结的方法 MUST 只保留为路线图，不得进入当批普通能力实现或动态`host service`实现。

#### Scenario: 方法缺少发布矩阵

- **WHEN** 变更计划新增插件可消费领域方法
- **THEN** OpenSpec 设计必须声明该方法是源码插件专属、同步动态发布、延后发布还是不发布
- **AND** 未声明发布决策的方法不得修改`capability.Services`或动态`hostServices`目录

#### Scenario: 动态方法缺少授权资源矩阵

- **WHEN** 变更计划把普通领域方法发布为动态`host service`
- **THEN** OpenSpec 设计必须声明对应`service`、`method`、`resource`和`plugin.yaml hostServices`授权形态
- **AND** 未声明授权资源的方法不得进入动态协议 catalog、guest client 或 dispatcher

#### Scenario: 高频方法缺少规模上限

- **WHEN** 变更计划新增批量、搜索、候选、树形、聚合或资源枚举类领域方法
- **THEN** OpenSpec 设计必须声明输入数量、分页、key/path 长度或总字节数上限
- **AND** 超限行为必须映射为结构化能力错误

### Requirement: 阶段一必须只发布冻结的高频只读领域能力

系统 SHALL 将`expand-plugin-domain-capabilities`第一批实现范围限定为当前用户投影、用户批量解析、批量权限判断、字典值可见性校验和当前在线会话投影。候选搜索、组织/租户投影、插件治理搜索、私有资源批量、通知类型化和`AI`方法状态 MUST 保留为后续阶段，除非另行更新 OpenSpec 并完成阶段矩阵冻结。

#### Scenario: 阶段一实现当前用户投影

- **WHEN** 源码插件或已授权动态插件调用用户当前投影方法
- **THEN** 系统返回当前 actor 可见的用户最小投影
- **AND** 缺少用户 actor 或系统上下文调用必须 fail-closed

#### Scenario: 阶段一实现用户批量解析

- **WHEN** 插件按用户 ID、用户名、手机号或邮箱批量解析用户
- **THEN** 系统在数据库查询阶段应用租户和数据权限过滤
- **AND** 不存在、不可见、租户外或未授权目标统一进入`MissingIDs`
- **AND** 实现不得对每个解析键执行一次用户详情查询

#### Scenario: 阶段一实现批量权限判断

- **WHEN** 插件一次判断多个权限 key
- **THEN** 系统返回每个权限 key 的布尔结果
- **AND** 实现必须复用权限快照、集合化权限服务或等价批量路径
- **AND** 不得循环调用单权限判断作为常规实现

#### Scenario: 阶段一发布字典值可见性校验

- **WHEN** 插件在写入或执行动作前校验一组字典值
- **THEN** 系统按字典类型和值集合执行可见性校验
- **AND** 任一值不存在或不可见时整体拒绝
- **AND** 错误不得区分不存在和不可见

#### Scenario: 阶段一实现当前在线会话投影

- **WHEN** 插件请求当前在线会话投影
- **THEN** 系统只返回当前 token 对应的会话投影
- **AND** 缺少请求型 token/session 上下文时必须 fail-closed
- **AND** 实现不得扫描全部在线会话来推断当前用户最新会话

### Requirement: 阶段一领域能力必须记录影响和验证证据

系统 SHALL 在任务记录和审查结论中记录阶段一领域能力对`i18n`、缓存一致性、数据权限、数据库、开发工具、测试和 E2E 的影响判断，并提供匹配验证证据。

#### Scenario: 无 HTTP API 或 UI 变化

- **WHEN** 阶段一只修改 Go 领域契约、动态 host service 协议和 README
- **THEN** 任务记录必须说明无静态 HTTP API、前端 UI、插件清单、语言包和 E2E 影响
- **AND** 仍必须运行 OpenSpec strict 校验、相关 Go 包测试和 README 静态检查

#### Scenario: 缓存和数据权限影响

- **WHEN** 阶段一方法复用权限、字典、用户或 session owner 数据
- **THEN** 任务记录必须说明权威数据源、共享实例或快照复用方式
- **AND** 数据权限读取或校验路径必须通过测试、静态检索或审查证据证明未退化为内存过滤或`N+1`

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

### Requirement: `Jobs` 领域能力创建、更新和查询必须支持任务级日志清理策略

系统 SHALL 允许源码插件和动态插件通过 `jobcap.Service.Create` 与 `jobcap.Service.Update` 在 `SaveInput` 中传入可选的任务级日志清理策略，并在 `jobcap.Service.Get`、`jobcap.Service.BatchGet` 与 `jobcap.Service.List` 返回的 `JobInfo` 中读取该任务级策略投影。该策略的模式和值 MUST 与宿主定时任务管理 API 的 `logRetentionOverride` 语义一致，并由宿主任务 owner 持久化到 `sys_job.log_retention_override`。未传入或未持久化策略时，任务 MUST 跟随系统参数 `sys.cron.log.retention`。策略为 `mode=none` 时，该任务 MUST 不按任务级策略清理日志。

#### Scenario: 源码插件创建任务时传入日志清理策略

- **WHEN** 源码插件通过 `Services().Jobs().Create` 创建定时任务
- **AND** `SaveInput.LogRetentionOverride` 为 `{mode:"days",value:60}` 或 `{mode:"count",value:500}`
- **THEN** 宿主 MUST 将该策略传递给定时任务 owner
- **AND** 后续日志清理 MUST 使用该任务级策略覆盖全局默认策略

#### Scenario: 动态插件更新任务时传入无清理策略

- **WHEN** 动态插件通过已授权的 `service: jobs`、`method: jobs.update` 更新定时任务
- **AND** 请求载荷中的 `LogRetentionOverride` 为 `{mode:"none",value:0}`
- **THEN** WASM host service MUST 将该策略解码为 `jobcap.SaveInput`
- **AND** 宿主 MUST 将该策略传递给定时任务 owner
- **AND** 后续日志清理 MUST 不按任务级策略清理该任务日志

#### Scenario: 未传入任务级策略时跟随系统默认

- **WHEN** 源码插件或动态插件创建、更新定时任务
- **AND** `SaveInput.LogRetentionOverride` 为空
- **THEN** 宿主 MUST 清空任务级覆盖
- **AND** 后续日志清理 MUST 按系统参数 `sys.cron.log.retention` 执行

#### Scenario: 插件查询任务时返回日志清理策略投影

- **WHEN** 源码插件或动态插件通过 `jobcap.Service.Get`、`jobcap.Service.BatchGet` 或 `jobcap.Service.List` 查询可见定时任务
- **AND** 目标任务的 `sys_job.log_retention_override` 为 `{mode:"days",value:60}`、`{mode:"count",value:500}` 或 `{mode:"none",value:0}`
- **THEN** 返回的 `JobInfo.LogRetentionOverride` MUST 包含对应的模式和值
- **AND** 若目标任务没有任务级策略，`JobInfo.LogRetentionOverride` MUST 为空，表示跟随系统默认策略
- **AND** 查询实现 MUST 通过现有可见性过滤和同一次任务投影读取该字段，不得为每条任务额外查询

#### Scenario: 非法策略复用任务 owner 校验

- **WHEN** 插件传入不支持的日志清理模式，或 `days`、`count` 模式的值小于等于 `0`
- **THEN** 宿主 MUST 返回结构化业务错误
- **AND** 不得创建或更新目标定时任务

