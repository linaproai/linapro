# plugin-package-boundary-governance 规范增量

## ADDED Requirements

### Requirement: `pkg/plugin`顶层组件依赖方向必须固定并受治理验证

系统 SHALL 固定`pkg/plugin`三个公开顶层组件的依赖方向：`capability`是最底层契约层，其非测试代码 MUST NOT import `pkg/plugin/pluginbridge`或`pkg/plugin/pluginhost`的任何子包；`pluginhost`非测试代码 MUST NOT import `pkg/plugin/pluginbridge`的任何子包。`pluginhost`与`pluginbridge`共享的契约类型 MUST 下沉到`capability`公共原语包，不得由其中一方 import 另一方获得。该依赖方向 MUST 由随`go test`执行的治理测试持续验证。

#### Scenario: 治理测试验证 capability 依赖边界

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `capability/**`非测试源文件不存在`lina-core/pkg/plugin/pluginbridge`或`lina-core/pkg/plugin/pluginhost`前缀的 import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 治理测试验证 pluginhost 依赖边界

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `pluginhost/**`非测试源文件不存在`lina-core/pkg/plugin/pluginbridge`前缀的 import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 源码插件升级回调使用中立 manifest 快照契约

- **WHEN** `pluginhost`为源码插件升级回调发布 manifest 快照契约
- **THEN** typed manifest snapshot 类型定义位于`pkg/plugin/capability/capmodel`
- **AND** `pluginbridge/contract`通过类型别名复用同一定义，动态插件生命周期请求的 JSON wire 格式保持不变
- **AND** `pluginhost`不因 manifest 快照 import `pluginbridge/contract`

#### Scenario: 测试代码跨边界验证豁免

- **WHEN** `capability`或`pluginhost`的`_test.go`文件为集成验证 import `pluginbridge`
- **THEN** 治理测试不将其判定为违规
- **AND** 豁免仅适用于测试代码，不适用于任何生产源文件

## MODIFIED Requirements

### Requirement: `pluginhost`、`pluginbridge`和`capability`必须职责分离

系统 SHALL 在`pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API，`pluginbridge`只负责动态插件 ABI、WASM transport、公开协议出口和动态插件专属 guest SDK，`capability`只负责插件消费宿主能力的稳定目录、公共原语和`*cap`能力契约。`capability`下的具体能力组件 MUST 使用职责明确的`*cap`包；公共原语包不得成为具体能力服务聚合点。仅服务动态插件的 guest SDK（如 record store）MUST 归属`pluginbridge`，不得放入`capability`。

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

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户或业务上下文等宿主能力
- **THEN** 插件使用`pkg/plugin/capability`、对应`pkg/plugin/capability/<domain>cap`能力组件或其 guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 record store 能力

- **WHEN** 动态插件需要使用 ORM-style record store facade、typed record store plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/pluginbridge/recordstore`
- **AND** Go guest 能力目录通过`RecordStore()`返回该 facade
- **AND** 不得继续通过`pkg/plugin/capability/recordstore`或顶层`pkg/plugindb`暴露该能力

### Requirement: 插件可导入契约不得放入`pkg/plugin/internal`

系统 SHALL 将`pkg/plugin/internal`限制为`pkg/plugin/...`公共组件自身共享的内部实现。源码插件、动态插件 guest 代码或宿主插件运行时需要直接 import 的接口、DTO、guest SDK、能力目录、公共原语和测试可见契约 MUST 位于公开子包中。

#### Scenario: 动态插件导入 guest SDK

- **WHEN** 动态插件 guest 代码需要访问宿主能力 guest client
- **THEN** guest SDK 位于`pkg/plugin/capability/guest`或其他公开子包
- **AND** guest SDK 不得位于`pkg/plugin/internal/guest`

#### Scenario: 动态插件导入 record store SDK

- **WHEN** 动态插件 guest 代码需要构造受治理表记录查询、变更或事务
- **THEN** record store SDK 位于`pkg/plugin/pluginbridge/recordstore`
- **AND** record store SDK 不得位于`pkg/plugin/pluginbridge/internal`、`pkg/plugin/capability/internal`或`pkg/plugin/internal`

#### Scenario: 源码插件导入业务上下文能力

- **WHEN** 源码插件服务需要读取当前请求身份、租户或代管上下文
- **THEN** 业务上下文能力位于`pkg/plugin/capability/bizctxcap`
- **AND** 业务上下文公共值对象位于公共原语包
- **AND** 不得放入`pkg/plugin/internal/bizctx`或旧`pkg/plugin/capability/contract`

#### Scenario: 宿主插件运行时复用内部实现

- **WHEN** `apps/lina-core/internal/service/plugin/...`需要复用插件运行时内部实现
- **THEN** 该实现必须位于宿主`internal/service/plugin/internal/...`或其他宿主可导入边界
- **AND** 不得放入`pkg/plugin/internal`后再由宿主运行时直接导入

### Requirement: capability 子包命名必须表达能力组件职责

系统 SHALL 要求`apps/lina-core/pkg/plugin/capability`下的插件公开能力组件使用`*cap`命名方式表达领域能力职责。除`guest`、`internal`和公共原语包等明确非具体能力组件外，公开能力组件 MUST 使用`<domain>cap`包名。

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
