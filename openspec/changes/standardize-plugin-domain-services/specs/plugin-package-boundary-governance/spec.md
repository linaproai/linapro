## ADDED Requirements

### Requirement: capabilityhost 和 WASM host service 必须保持薄适配边界

系统 SHALL 将`internal/service/plugin/internal/capabilityhost`、动态 WASM host service dispatcher 和相关 host service adapter 限定为插件调用适配层。适配层只允许承担标准业务上下文桥接、动态 registry 与授权校验、请求响应编解码和错误映射，不得承载其他领域业务实现，不得直接访问其他领域`DAO`、`DO`、`Entity`、私有缓存、私有 provider 或内部 helper。

#### Scenario: 动态插件调用用户领域方法

- **WHEN** 动态插件通过 host service 调用用户领域方法
- **THEN** WASM host service dispatcher 先完成 registry、授权、资源和 codec 校验
- **AND** 业务处理必须委托到用户领域 owner 发布的统一`Service`或稳定契约
- **AND** dispatcher 不得直接查询`sys_user`或导入用户领域内部 DAO

#### Scenario: 适配层需要新增领域能力

- **WHEN** 插件适配层需要支持新的领域业务能力
- **THEN** 该能力必须先由真实领域 owner 发布统一`*cap.Service`方法或稳定适配契约
- **AND** 适配层只能通过构造函数显式注入该契约
- **AND** 不得在适配层内部临时`New()`领域服务或复制业务逻辑

#### Scenario: 治理扫描检查适配层边界

- **WHEN** 执行插件包边界治理扫描
- **THEN** 扫描必须阻断`capabilityhost`或 WASM host service 生产代码导入其他领域`internal/dao`、`internal/model`或私有实现包
- **AND** 受控启动装配、测试 fixture 和代码生成入口必须在审查记录中说明例外边界

### Requirement: pluginhost 服务目录不得暴露 Admin 入口

系统 SHALL 将`pluginhost.Services`定义为源码插件获取统一`capability.Services`和源码插件专用能力的入口。`pluginhost.Services` MUST NOT 暴露`Admin()`、`AdminServices`或等价管理目录；源码插件管理动作 MUST 通过对应领域统一`Service`调用。

#### Scenario: 源码插件获取用户管理能力

- **WHEN** 源码插件需要创建、更新、删除用户或变更用户状态
- **THEN** 插件通过`pluginhost.Services.Users()`或注入的`usercap.Service`调用对应方法
- **AND** 插件不得通过`pluginhost.Services.Admin().Users()`获取管理目录

#### Scenario: 静态检索发现 Admin 入口

- **WHEN** 静态检索发现生产代码、官方插件、测试替身或 README 继续引用`Services.Admin()`、`AdminServices`或领域`AdminService`
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到统一领域`Service`

## MODIFIED Requirements

### Requirement: `pluginhost`、`pluginbridge`和`capability`必须职责分离

系统 SHALL 在`pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API 和源码插件获取统一服务目录，`pluginbridge`只负责动态插件 ABI、WASM transport、公开协议出口和动态插件专属 guest SDK，`capability`只负责插件消费宿主能力的稳定目录、公共原语和`*cap`能力契约。`capability`下的具体能力组件 MUST 使用职责明确的`*cap`包；公共原语包不得成为具体能力服务聚合点。仅服务动态插件的 guest SDK（如 record store）MUST 归属`pluginbridge`，不得放入`capability`。

#### Scenario: 源码插件注册贡献

- **WHEN** 源码插件需要注册路由、hook、cron、生命周期回调或 provider factory
- **THEN** 插件使用`pkg/plugin/pluginhost`
- **AND** `pluginhost`不得拥有宿主能力业务实现
- **AND** `pluginhost`不得暴露`Admin()`或`AdminServices`管理目录
- **AND** `pluginhost.Services`顶层不得成为新增治理能力的事实 owner，治理能力必须委托到对应领域`Service`

#### Scenario: 动态插件执行 bridge 调用

- **WHEN** 动态插件需要声明 route、处理 WASM request envelope 或调用 host call transport
- **THEN** 插件使用`pkg/plugin/pluginbridge/guest`和`pkg/plugin/pluginbridge/protocol`
- **AND** `pluginbridge`不得定义业务能力可用性、provider 激活、数据权限降级或配置读取语义
- **AND** `pluginbridge/guest`不得承载`RuntimeHostService`、`StorageHostService`、`ConfigHostService`、`DataHostService`等宿主能力 client 实现
- **AND** `pluginbridge`根包不得重新导出协议 DTO、常量、codec、guest helper 或`Runtime()`、`Data()`、`RecordStore()`、`Cron()`等能力 client facade

#### Scenario: 插件消费宿主能力

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户或业务上下文等宿主能力
- **THEN** 插件使用`pkg/plugin/capability`、对应`pkg/plugin/capability/<domain>cap`能力组件或其 guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 record store 能力

- **WHEN** 动态插件需要使用 ORM-style record store facade、typed record store plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/pluginbridge/recordstore`
- **AND** Go guest 能力目录通过`RecordStore()`返回该 facade
- **AND** 不得继续通过`pkg/plugin/capability/recordstore`或顶层`pkg/plugindb`暴露该能力

### Requirement: 旧非`*cap`能力包不得作为生产入口保留

系统 SHALL 在迁移完成后删除旧非`*cap`公开能力包入口。生产代码、官方插件、动态插件样例和测试替身 MUST 不再导入`capability/ai`、`capability/bizctx`、`capability/config`、`capability/hostconfig`、`capability/manifest`、`capability/pluginlifecycle`、`capability/pluginstate`或`capability/tenantfilter`作为具体能力组件；同时不得新增`capability/tenantfiltercap`独立能力组件，也不得通过根`Services.Config()`、`Services.PluginConfig()`、`Services.PluginLifecycle()`、`Services.PluginState()`、`Services.TenantPluginGovernance()`或`Services.TenantFilter()`访问插件相关治理能力。

生产代码、官方插件、动态插件样例和测试替身也 MUST 不再导入`capability/authzcap`，不得继续把根`capability/authcap`作为 token 窄服务包使用。认证 token 能力必须迁移到`capability/authcap/token`，授权能力必须迁移到`capability/authcap/authz`，根`capability/authcap`只保留能力族聚合入口。生产代码、官方插件、动态插件样例和测试替身还 MUST 不再引用`AdminService`、`AdminServices`或`Services.Admin()`作为插件可见能力入口。

#### Scenario: 静态检索发现旧包导入

- **WHEN** 静态检索发现生产代码导入旧非`*cap`具体能力包或调用旧根能力入口
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到目标`*cap`包或`Services.Plugins()`子领域

#### Scenario: 测试替身实现旧接口

- **WHEN** 测试替身继续实现旧包路径下的具体服务接口、领域`AdminService`或`AdminServices`目录
- **THEN** Go 编译门禁必须暴露该遗漏
- **AND** 测试替身必须改为实现新`*cap.Service`接口

#### Scenario: 静态检索发现旧 Admin 入口

- **WHEN** 静态检索发现生产代码、官方插件、动态插件样例、测试替身或文档继续引用`Services.Admin()`、`AdminServices`或领域`AdminService`
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到统一领域`Service`
