## ADDED Requirements

### Requirement: 插件公开能力服务必须归属`*cap`组件包

系统 SHALL 要求`apps/lina-core/pkg/plugin/capability`下对插件公开的具体能力服务接口归属职责明确的领域命名空间或`*cap`组件包。`capability.Services`普通消费面 MUST 只返回各领域命名空间、`*cap.Service`或等价窄接口，不得返回`contract.*Service`具体服务接口。

#### Scenario: 根能力目录返回具体服务

- **WHEN** 开发者查看`capability.Services`
- **THEN** 每个普通能力方法返回对应领域命名空间、`*cap`组件包的服务接口或等价窄接口
- **AND** `APIDoc`、`Auth`、`BizCtx`、`Cache`、`HostConfig`、`I18n`、`Manifest`和`Route`不得再返回`contract.*Service`
- **AND** 根目录不得继续暴露`Config()`、`PluginConfig()`、`PluginLifecycle()`或`PluginState()`

#### Scenario: 认证授权能力族入口

- **WHEN** 插件需要访问认证 token handoff 或授权能力
- **THEN** 根能力目录只暴露`Services.Auth()`认证授权能力族入口
- **AND** token 生命周期能力通过`Services.Auth().Token()`访问，接口归属`pkg/plugin/capability/authcap/token`
- **AND** 授权查询能力通过`Services.Auth().Authz()`访问，接口归属`pkg/plugin/capability/authcap/authz`
- **AND** 根能力目录不得继续并列暴露`Services.Authz()`

#### Scenario: 静态检索发现旧具体服务引用

- **WHEN** 静态检索发现生产代码通过`capability.Services`公开面新增`contract.*Service`具体能力引用
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到对应`*cap`组件包

### Requirement: Services 方法名必须按领域消费语义命名

系统 SHALL 将`capability.Services`方法名视为插件开发者看到的领域入口，而不是 Go 包名的机械映射。资源集合或领域命名空间 MAY 使用复数入口，例如`Users()`、`Jobs()`、`Plugins()`；单一上下文、配置能力或专有能力 SHOULD 使用单数或专有名词，例如`Tenant()`、`BizCtx()`、`HostConfig()`、`AI()`。Go 组件包名 MUST 继续使用单数领域名加`cap`后缀，例如`usercap`、`jobcap`、`plugincap`。

#### Scenario: 用户领域入口命名

- **WHEN** 插件读取用户领域普通能力
- **THEN** 根入口使用`Services.Users()`
- **AND** 返回接口归属`usercap.Service`
- **AND** 不得为了匹配包名而强制改为容易表示当前用户对象的`Services.User()`

#### Scenario: 插件领域入口命名

- **WHEN** 插件读取插件自身配置、插件状态、生命周期或插件治理投影
- **THEN** 根入口使用`Services.Plugins()`作为插件领域命名空间
- **AND** 返回接口归属`plugincap.Service`
- **AND** 不得为了匹配包名而强制改为容易表示当前插件对象的`Services.Plugin()`

### Requirement: 插件相关能力必须收口到 Plugins 命名空间

系统 SHALL 将插件自身配置、插件状态、插件生命周期和插件治理投影收口到`Services.Plugins()`插件领域命名空间。根`capability.Services` MUST NOT 继续暴露`Config()`、`PluginConfig()`、`PluginLifecycle()`或`PluginState()`。

#### Scenario: 插件读取自身配置

- **WHEN** 源码插件需要读取当前插件自身配置
- **THEN** 插件通过`Services.Plugins().Config()`获取插件作用域配置服务
- **AND** 不得通过根`Services.Config()`或根`Services.PluginConfig()`读取

#### Scenario: 插件读取启用状态

- **WHEN** 源码插件需要读取插件启用状态或 provider 可用性
- **THEN** 插件通过`Services.Plugins().State()`获取插件状态服务
- **AND** 不得通过根`Services.PluginState()`读取

#### Scenario: 插件触发生命周期治理

- **WHEN** 源码插件需要执行插件生命周期前置校验或通知
- **THEN** 插件通过`Services.Plugins().Lifecycle()`获取生命周期服务
- **AND** 不得通过根`Services.PluginLifecycle()`调用

### Requirement: 配置公开面只能包含插件自身配置和宿主配置

系统 SHALL 将插件公开配置能力限定为两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主授权开放配置。根`Services.Config()` MUST NOT 作为普通插件公开入口存在。

#### Scenario: 插件读取宿主配置

- **WHEN** 插件需要读取宿主授权开放配置项
- **THEN** 插件通过`Services.HostConfig()`访问
- **AND** `HostConfig()`不得读取当前插件私有`config.yaml`

#### Scenario: 插件读取自身配置

- **WHEN** 插件需要读取自身`config.yaml`或 artifact 内配置
- **THEN** 插件通过`Services.Plugins().Config()`访问
- **AND** `Plugins().Config()`不得读取宿主配置树或运行时配置中心数据

### Requirement: 租户过滤不得进入普通租户消费面

系统 SHALL 将源码插件自有表`tenant_id`过滤接口归属到`tenantcap.PluginTableFilterService`，但该接口 MUST 只通过`pluginhost.Services.TenantFilter()`等源码插件专用受控接缝暴露。普通`capability.Services.Tenant()` MUST 只返回`tenantcap.Service`普通租户消费面，不得提供`Filter()`、`Apply(...)`或任何携带`*gdb.Model`、SQL 片段、DAO、query builder 的方法。

#### Scenario: 源码插件过滤自有表

- **WHEN** 源码插件需要给插件自有表查询追加当前租户过滤
- **THEN** 插件通过`pluginhost.Services.TenantFilter()`获取`tenantcap.PluginTableFilterService`
- **AND** 该接口可以接收`*gdb.Model`并追加约定`tenant_id`谓词
- **AND** 该接口不得通过`capability.Services.Tenant()`普通入口暴露

#### Scenario: 普通插件读取租户能力

- **WHEN** 插件通过`Services.Tenant()`读取租户能力
- **THEN** 返回值只允许包含`Current`、`EnsureTenantVisible`、`ListUserTenants`等普通租户能力
- **AND** 不得包含`Filter()`、`Apply(...)`或插件自有表查询构造器

### Requirement: 公共原语包不得承载具体能力服务

系统 SHALL 允许`capability`维护一个小型公共原语包，用于承载跨领域值对象、分页结果、批量结果、能力上下文和能力状态。该公共原语包 MUST NOT 定义具体能力`Service`、`AdminService`、factory、provider adapter 或 host service handler。

#### Scenario: 领域组件使用公共原语

- **WHEN** `usercap`、`dictcap`、`aicap`或其他领域组件需要`CapabilityContext`、`BatchResult`或`PageRequest`
- **THEN** 它们可以导入公共原语包
- **AND** 公共原语包只提供值对象和通用结果结构

#### Scenario: 新增具体能力接口

- **WHEN** 系统新增插件可见宿主能力接口
- **THEN** 该接口必须放入对应`*cap`组件包
- **AND** 不得放入公共原语包或恢复`contract`万能聚合包

### Requirement: 认证授权子能力必须收敛到`authcap`能力族

系统 SHALL 将认证 token handoff 与授权能力作为`authcap`能力族维护。`authcap`根包只承载聚合入口，子领域`authcap/token`维护租户 token、tenant switch 和 impersonation token 契约，子领域`authcap/authz`维护权限投影、权限检查和角色授权管理契约。源码插件业务服务 MUST 通过构造函数接收所需的窄子领域接口，不得为了目录收敛而长期保存整个`authcap.Service`。

#### Scenario: token 子领域

- **WHEN** 源码插件需要选择租户、切换租户或签发 impersonation token
- **THEN** 依赖`authcap/token.Service`
- **AND** 不得导入旧`pkg/plugin/capability/authcap`作为 token 窄服务包

#### Scenario: 授权子领域

- **WHEN** 源码插件需要批量读取权限投影、检查权限或判断平台管理员
- **THEN** 依赖`authcap/authz.Service`
- **AND** 不得导入旧`pkg/plugin/capability/authzcap`

#### Scenario: 认证授权管理子目录

- **WHEN** 受信任源码插件需要角色权限管理命令
- **THEN** 通过管理服务目录的认证授权子目录获取`authcap/authz.AdminService`
- **AND** 管理根目录不得继续并列暴露`AdminServices.Authz()`

### Requirement: 旧`capability/contract`具体服务聚合必须删除

系统 SHALL 删除或清空`capability/contract`作为具体服务聚合包的职责。迁移完成后，生产代码、官方插件和测试替身 MUST 不再导入`lina-core/pkg/plugin/capability/contract`获取具体能力服务接口；若仍需公共原语，必须导入新的公共原语包。

#### Scenario: 官方源码插件导入旧`contract`

- **WHEN** 官方源码插件生产代码导入`lina-core/pkg/plugin/capability/contract`
- **THEN** 静态检索、Go 编译门禁或审查必须阻断
- **AND** 插件必须改为依赖对应领域`*cap`包或公共原语包

#### Scenario: 宿主适配器导入旧`contract`

- **WHEN** `hostservices`、`WASM`host service 或启动装配代码继续使用`contract.ConfigService`、`contract.ManifestService`或等价具体服务接口
- **THEN** 变更不得通过审查
- **AND** 适配器必须迁移到`plugincap`配置子服务、`manifestcap`或其他目标组件
