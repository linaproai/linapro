## ADDED Requirements

### Requirement: capability 子包命名必须表达能力组件职责

系统 SHALL 要求`apps/lina-core/pkg/plugin/capability`下的插件公开能力组件使用`*cap`命名方式表达领域能力职责。除`guest`、`recordstore`、`internal`和公共原语包等明确非具体能力组件外，公开能力组件 MUST 使用`<domain>cap`包名。

#### Scenario: AI 能力组件命名

- **WHEN** 系统发布`AI`能力族聚合入口
- **THEN** Go 包路径使用`pkg/plugin/capability/aicap`
- **AND** 不得继续使用`pkg/plugin/capability/ai`作为公开能力组件包

#### Scenario: 认证授权能力族命名

- **WHEN** 系统发布认证 token 与授权能力
- **THEN** Go 包路径使用`pkg/plugin/capability/authcap`作为认证授权能力族聚合入口
- **AND** token 子领域位于`pkg/plugin/capability/authcap/token`
- **AND** 授权子领域位于`pkg/plugin/capability/authcap/authz`
- **AND** 不得继续使用根`pkg/plugin/capability/authcap`作为 token 窄服务包或使用`pkg/plugin/capability/authzcap`作为授权能力包

#### Scenario: 插件配置能力归属

- **WHEN** 源码插件或动态插件读取当前插件作用域静态配置
- **THEN** Go 入口使用`Services.Plugins().Config()`
- **AND** 插件配置接口归属`plugincap`插件领域子服务
- **AND** 不得继续使用容易与运行时配置领域混淆的根`Services.Config()`或`capability/config`公开包

#### Scenario: 租户过滤能力归属

- **WHEN** 源码插件需要使用插件自有表租户过滤能力
- **THEN** 过滤接口归属`pkg/plugin/capability/tenantcap`下的源码插件专用接口
- **AND** 不得继续保留`pkg/plugin/capability/tenantfilter`公开包
- **AND** 不得新增独立`pkg/plugin/capability/tenantfiltercap`能力组件包

#### Scenario: 公共原语包例外

- **WHEN** 包只承载跨领域值对象、分页结构、批量结果或能力上下文
- **THEN** 该包可以不使用`*cap`后缀
- **AND** 它不得定义具体能力服务接口

### Requirement: 旧非`*cap`能力包不得作为生产入口保留

系统 SHALL 在迁移完成后删除旧非`*cap`公开能力包入口。生产代码、官方插件、动态插件样例和测试替身 MUST 不再导入`capability/ai`、`capability/bizctx`、`capability/config`、`capability/hostconfig`、`capability/manifest`、`capability/pluginlifecycle`、`capability/pluginstate`或`capability/tenantfilter`作为具体能力组件；同时不得新增`capability/tenantfiltercap`独立能力组件，也不得通过根`Services.Config()`、`Services.PluginConfig()`、`Services.PluginLifecycle()`或`Services.PluginState()`访问插件相关能力。

生产代码、官方插件、动态插件样例和测试替身也 MUST 不再导入`capability/authzcap`，不得继续把根`capability/authcap`作为 token 窄服务包使用。认证 token 能力必须迁移到`capability/authcap/token`，授权能力必须迁移到`capability/authcap/authz`，根`capability/authcap`只保留能力族聚合入口。

#### Scenario: 静态检索发现旧包导入

- **WHEN** 静态检索发现生产代码导入旧非`*cap`具体能力包或调用旧根能力入口
- **THEN** 验证或审查必须失败
- **AND** 调用方必须迁移到目标`*cap`包或`Services.Plugins()`子领域

#### Scenario: 测试替身实现旧接口

- **WHEN** 测试替身继续实现旧包路径下的具体服务接口
- **THEN** Go 编译门禁必须暴露该遗漏
- **AND** 测试替身必须改为实现新`*cap`接口

## MODIFIED Requirements

### Requirement: `pluginhost`、`pluginbridge`和`capability`必须职责分离

系统 SHALL 在`pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API，`pluginbridge`只负责动态插件 ABI、WASM transport 和公开协议出口，`capability`只负责插件消费宿主能力的稳定目录、公共原语和`*cap`能力契约。`capability`下的具体能力组件 MUST 使用职责明确的`*cap`包；公共原语包不得成为具体能力服务聚合点。

#### Scenario: 源码插件注册贡献

- **WHEN** 源码插件需要注册路由、hook、cron、生命周期回调或 provider factory
- **THEN** 插件使用`pkg/plugin/pluginhost`
- **AND** `pluginhost`不得拥有宿主能力消费目录实现

#### Scenario: 动态插件执行 bridge 调用

- **WHEN** 动态插件需要声明 route、处理 WASM request envelope 或调用 host call transport
- **THEN** 插件使用`pkg/plugin/pluginbridge/guest`和`pkg/plugin/pluginbridge/protocol`
- **AND** `pluginbridge`不得定义业务能力可用性、provider 激活、数据权限降级或配置读取语义
- **AND** `pluginbridge/guest`不得承载`RuntimeHostService`、`StorageHostService`、`ConfigHostService`、`DataHostService`等宿主能力 client 实现
- **AND** `pluginbridge`根包不得重新导出协议 DTO、常量、codec、guest helper 或`Runtime()`、`Data()`、`RecordStore()`、`Cron()`等能力 client facade

#### Scenario: 插件消费宿主能力

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户、record store 或业务上下文等宿主能力
- **THEN** 插件使用`pkg/plugin/capability`、对应`pkg/plugin/capability/<domain>cap`能力组件或其 guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 record store 能力

- **WHEN** 动态插件需要使用 ORM-style record store facade、typed record store plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/capability/recordstore`
- **AND** Go guest 能力目录通过`RecordStore()`返回该 facade
- **AND** 不得继续通过顶层`pkg/plugindb`暴露该能力

### Requirement: 插件可导入契约不得放入`pkg/plugin/internal`

系统 SHALL 将`pkg/plugin/internal`限制为`pkg/plugin/...`公共组件自身共享的内部实现。源码插件、动态插件 guest 代码或宿主插件运行时需要直接 import 的接口、DTO、guest SDK、能力目录、公共原语和测试可见契约 MUST 位于公开子包中。

#### Scenario: 动态插件导入 guest SDK

- **WHEN** 动态插件 guest 代码需要访问宿主能力 guest client
- **THEN** guest SDK 位于`pkg/plugin/capability/guest`或其他公开子包
- **AND** guest SDK 不得位于`pkg/plugin/internal/guest`

#### Scenario: 动态插件导入 record store SDK

- **WHEN** 动态插件 guest 代码需要构造受治理表记录查询、变更或事务
- **THEN** record store SDK 位于`pkg/plugin/capability/recordstore`
- **AND** record store SDK 不得位于`pkg/plugin/capability/internal`或`pkg/plugin/internal`

#### Scenario: 源码插件导入业务上下文能力

- **WHEN** 源码插件服务需要读取当前请求身份、租户或代管上下文
- **THEN** 业务上下文能力位于`pkg/plugin/capability/bizctxcap`
- **AND** 业务上下文公共值对象位于公共原语包
- **AND** 不得放入`pkg/plugin/internal/bizctx`或旧`pkg/plugin/capability/contract`

#### Scenario: 宿主插件运行时复用内部实现

- **WHEN** `apps/lina-core/internal/service/plugin/...`需要复用插件运行时内部实现
- **THEN** 该实现必须位于宿主`internal/service/plugin/internal/...`或其他宿主可导入边界
- **AND** 不得放入`pkg/plugin/internal`后再由宿主运行时直接导入
