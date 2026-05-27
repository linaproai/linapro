## MODIFIED Requirements

### Requirement: 插件宿主服务适配器必须由宿主运行期统一构造

系统 SHALL 由宿主运行期统一构造并发布源码插件和动态插件 host service 适配器。适配器 MUST 复用启动期共享的宿主服务实例或共享后端，MUST 不在插件调用路径中自行构造孤立宿主服务图。

#### Scenario: 源码插件使用宿主能力目录

- **WHEN** 源码插件调用`pkg/plugin/capability`发布的宿主能力
- **THEN** 该能力目录由宿主运行期构造并通过源码插件 registrar 或 callback 输入传递给插件
- **AND** 能力目录复用宿主共享的 auth、session、notify、config、i18n、pluginstate、org、tenant、cache 或其他依赖
- **AND** 插件生产路径不得无参创建该能力目录或对应适配器

#### Scenario: 动态插件 host service 调用共享宿主能力

- **WHEN** 动态插件通过统一 host service 协议调用 cache、lock、notify、config、runtime、storage 或 data 等宿主能力
- **THEN** host service handler 使用插件 runtime 注入的共享宿主服务或共享后端
- **AND** handler 不得在每次调用中创建独立 cache、lock、notify、config 或 plugin service 实例

#### Scenario: WASM host service 配置入口由启动期注入

- **WHEN** 宿主启动并初始化 WASM host service
- **THEN** 启动路径显式配置 cache、lock、notify、storage、config、runtime 和 framework capability 等 host service 的共享依赖
- **AND** 包级默认实例不得在生产启动后继续作为实际运行依赖

## ADDED Requirements

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
