# service-dependency-injection-governance 规范增量

## ADDED Requirements

### Requirement: Capability Provider Manager 必须由宿主显式持有

系统 SHALL 要求框架 capability provider manager 由宿主启动装配或插件能力宿主装配层创建、持有并通过构造函数显式注入。能力包父包或 SPI 子包 MUST NOT 通过包级`defaultManager`、全局 service locator、隐式`New()`或旧`Provide()`函数保存 provider factory 注册表。

#### Scenario: 宿主创建租户 Provider Manager

- **WHEN** 宿主构造租户 capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[tenantspi.ProviderEnv]`实例
- **AND** 租户 service 构造函数显式接收该 manager
- **AND** `tenantcap`或`tenantspi`包级作用域不存在`defaultManager`

#### Scenario: 宿主创建组织 Provider Manager

- **WHEN** 宿主构造组织 capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[orgspi.ProviderEnv]`实例
- **AND** 组织 service 构造函数显式接收该 manager
- **AND** `orgcap`或`orgspi`包级作用域不存在`defaultManager`

#### Scenario: 宿主创建 AI 文本 Provider Manager

- **WHEN** 宿主构造文本`AI` capability host service
- **THEN** 宿主创建或传入共享的`capabilityregistry.Manager[aitext.ProviderEnv]`实例
- **AND** 文本`AI`service 构造函数显式接收该 manager
- **AND** `aitext`包级作用域不存在`defaultManager`

### Requirement: Source Plugin Provider 声明必须进入 registrar 生命周期

系统 SHALL 要求源码插件 provider factory 声明通过`pluginhost.Declarations`进入源码插件 registrar 生命周期。宿主 MUST 能从`SourcePluginDefinition`读取 provider factory 声明，并在插件能力宿主装配阶段注册到共享 provider manager；注册 API MUST 返回`error`给调用方决策，不得在可预期失败时 panic。

#### Scenario: 插件声明 Provider Factory

- **WHEN** 源码插件在`backend/plugin.go`中声明组织、租户或文本`AI`provider factory
- **THEN** 它调用`pluginhost.Declarations`提供的强类型 provider 声明方法
- **AND** 声明方法校验 nil factory、重复声明或非法插件 ID 时返回`error`
- **AND** 插件入口自行决定是否在顶层注册失败时 panic

#### Scenario: 宿主注册 Provider Factory 到共享 Manager

- **WHEN** 宿主读取一个源码插件定义中的 provider factory 声明
- **THEN** 宿主将该 factory 与声明插件 ID 注册到对应共享 manager
- **AND** provider 使用路径继续通过插件 enabled snapshot 判断该插件是否可用
- **AND** 宿主不得在业务请求路径临时创建新的 provider manager

#### Scenario: DI 来源检查覆盖 Provider Manager

- **WHEN** OpenSpec 任务完成 provider manager 迁移
- **THEN** 任务记录必须说明 manager 的 owner、创建位置、传递路径、共享实例或共享后端策略
- **AND** 若没有新增缓存、数据权限或运行期依赖语义变化，也必须记录无影响判断
