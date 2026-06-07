## ADDED Requirements

### Requirement: 宿主服务适配器必须适配到`*cap`能力组件

系统 SHALL 要求源码插件 hostservices directory、动态插件`WASM`host service handler 和 guest SDK 最终适配到`pkg/plugin/capability/<domain>cap`能力组件。适配层 MUST 不再依赖`capability/contract`具体能力服务接口作为运行时服务目录契约。

#### Scenario: 源码插件构造宿主服务目录

- **WHEN** 宿主启动期构造源码插件可消费的`capability.Services`
- **THEN** directory 字段和方法返回目标`*cap.Service`
- **AND** 认证授权能力以`authcap.Service`能力族聚合入口发布，token 子服务归属`authcap/token`，授权子服务归属`authcap/authz`
- **AND** 插件作用域配置 factory 通过`Plugins().Config()`对应子服务发布
- **AND** 其他插件作用域能力 factory 使用对应`manifestcap`或其他目标组件
- **AND** 源码插件专用`TenantFilter()`返回`tenantcap.PluginTableFilterService`，且不进入普通`capability.Services`

#### Scenario: 动态 host service 调用领域能力

- **WHEN** 动态插件通过`hostServices`调用`service + method`
- **THEN** `WASM`host service handler 校验授权后进入对应`*cap.Service`或`*cap.AdminService`
- **AND** handler 不得通过旧`contract.*Service`绕过目标能力组件

### Requirement: 动态插件不得直接消费源码插件租户过滤器

系统 SHALL 禁止动态插件 guest SDK、动态`hostServices`协议和`WASM`host service handler 暴露`tenantcap.PluginTableFilterService`、`*gdb.Model`、SQL 片段、DAO 或 query builder。动态插件租户隔离 MUST 通过普通`tenantcap.Service`读取当前租户或校验租户可见性，并由宿主 host service handler 在调用边界执行与宿主 API 等价的数据权限和租户边界过滤。

#### Scenario: 动态插件读取当前租户

- **WHEN** 动态插件需要知道当前请求租户
- **THEN** guest SDK 通过普通`Tenant().Current()`或等价 host service 调用读取
- **AND** 宿主返回租户 DTO 或租户 ID
- **AND** guest 不得获得`tenantcap.PluginTableFilterService`

#### Scenario: 动态插件调用宿主数据读取能力

- **WHEN** 动态插件通过用户、文件、通知、配置、插件治理或其他 host service 读取数据
- **THEN** 对应 handler 在宿主侧基于调用身份、`pluginID`、当前租户、授权快照和既有数据权限边界执行过滤
- **AND** 动态插件不得传入 SQL、DAO、`*gdb.Model`或 query builder 表达租户过滤

#### Scenario: 未来动态插件需要自有数据存储

- **WHEN** 动态插件需要维护插件自有持久化数据
- **THEN** 系统必须另行设计租户安全存储能力
- **AND** 该能力由宿主按`pluginID + tenantID`自动隔离
- **AND** 不得复用源码插件专用`TenantFilter()`作为动态插件数据访问协议

### Requirement: Go 包重命名不得改变动态插件协议

系统 SHALL 将本次`*cap`包重命名视为 Go 公共包边界重构。动态插件`plugin.yaml hostServices`声明、运行时授权快照、`service`字符串、`method`字符串、资源授权、protobuf envelope、错误 envelope 和审计语义 MUST 保持当前目标模型不变。

#### Scenario: 动态插件声明 AI 文本能力

- **WHEN** 动态插件声明`service: ai`和`method: text.generate`
- **THEN** 宿主仍按`service: ai`和`method: text.generate`校验授权
- **AND** Go 侧能力组件重命名为`aicap`不得要求插件清单改为`service: aicap`

#### Scenario: 动态插件读取插件配置

- **WHEN** 动态插件声明`service: config`和`methods: [get]`
- **THEN** 宿主仍按`config.get`授权快照执行校验
- **AND** Go 侧插件配置能力收口到`Plugins().Config()`不得改变动态协议服务名

#### Scenario: 动态插件读取宿主配置

- **WHEN** 动态插件声明`service: hostConfig`和授权配置 key
- **THEN** 宿主仍按宿主配置授权快照执行校验
- **AND** Go 侧宿主配置能力使用`HostConfig()`，不得与`Plugins().Config()`混用

### Requirement: 适配器迁移必须复用启动期共享实例

系统 SHALL 在`*cap`包重命名和接口迁移后继续由宿主运行期统一构造 host service 适配器。缓存敏感服务、插件状态、配置、`i18n`、会话、通知、组织、租户和插件生命周期等依赖 MUST 复用启动期共享实例或共享后端，不得因包迁移在插件调用路径中临时`New()`关键服务图。

#### Scenario: 源码插件读取缓存

- **WHEN** 源码插件通过迁移后的`cachecap.Service`读取插件作用域缓存
- **THEN** 该服务仍委托启动期传入的共享`kvcache`后端
- **AND** 不得在每次插件调用中创建独立缓存实例

#### Scenario: 动态插件读取插件状态

- **WHEN** 动态插件通过迁移后的`Plugins().State()`读取启用状态
- **THEN** handler 仍使用启动期注入的共享插件状态服务或共享快照
- **AND** 不得退化为仅当前节点可见的包级默认实例
