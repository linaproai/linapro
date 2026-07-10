# plugin-package-boundary-governance Specification

## Purpose
TBD - created by archiving change refactor-plugin-package-boundaries. Update Purpose after archive.
## Requirements
### Requirement: 插件公共命名空间必须集中到`pkg/plugin`

系统 SHALL 将插件相关公共 Go 组件集中到`apps/lina-core/pkg/plugin/`命名空间下。该命名空间下的公开顶层组件必须按职责拆分为源码插件贡献入口、动态插件桥接协议和插件消费宿主能力目录，不得继续在`apps/lina-core/pkg/`根层新增语义模糊的插件公共组件。

#### Scenario: 开发者定位插件公共入口

- **WHEN** 开发者需要查找 LinaPro 插件公共 Go 契约
- **THEN** 系统在`apps/lina-core/pkg/plugin/`下提供对应公共组件
- **AND** 不要求开发者在`pkg/pluginservice`、`pkg/pluginhost`、`pkg/pluginbridge`和其他顶层插件包之间猜测职责边界

#### Scenario: 新增插件公共组件

- **WHEN** 系统需要新增插件开发者或宿主插件运行时共享的公共契约
- **THEN** 该契约必须优先放入`apps/lina-core/pkg/plugin/`下职责明确的公开子包
- **AND** 不得新增`pluginservice`、`plugincommon`、`pluginutil`等语义模糊的顶层公共包

### Requirement: `pluginhost`、`pluginbridge`和`capability`必须职责分离

系统 SHALL 在`pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API，`pluginbridge`只负责动态插件 ABI、WASM transport 和公开协议出口，`capability`只负责插件消费宿主能力的稳定目录和能力契约。

#### Scenario: 源码插件注册贡献

- **WHEN** 源码插件需要注册路由、hook、cron、生命周期回调或 provider factory
- **THEN** 插件使用`pkg/plugin/pluginhost`
- **AND** `pluginhost`不得拥有宿主能力消费目录实现

#### Scenario: 动态插件执行 bridge 调用

- **WHEN** 动态插件需要声明 route、处理 WASM request envelope 或调用 host call transport
- **THEN** 插件使用`pkg/plugin/pluginbridge/guest`和`pkg/plugin/pluginbridge/protocol`
- **AND** `pluginbridge`不得定义业务能力可用性、provider 激活、数据权限降级或配置读取语义
- **AND** `pluginbridge/guest`不得承载`RuntimeHostService`、`StorageHostService`、`ConfigHostService`、`DataHostService`等宿主能力 client 实现
- **AND** `pluginbridge`根包不得重新导出协议 DTO、常量、codec、guest helper 或`Runtime()`、`Data()`、`Cron()`等能力 client facade

#### Scenario: 插件消费宿主能力

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户、data 或业务上下文等宿主能力
- **THEN** 插件使用`pkg/plugin/capability`或其 guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 data 能力

- **WHEN** 动态插件需要使用 ORM-style data facade、typed data plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/capability/data`
- **AND** 不得继续通过顶层`pkg/plugindb`暴露该能力

### Requirement: 插件可导入契约不得放入`pkg/plugin/internal`

系统 SHALL 将`pkg/plugin/internal`限制为`pkg/plugin/...`公共组件自身共享的内部实现。源码插件、动态插件 guest 代码或宿主插件运行时需要直接 import 的接口、DTO、guest SDK、能力目录和测试可见契约 MUST 位于公开子包中。

#### Scenario: 动态插件导入 guest SDK

- **WHEN** 动态插件 guest 代码需要访问宿主能力 guest client
- **THEN** guest SDK 位于`pkg/plugin/capability/guest`或其他公开子包
- **AND** guest SDK 不得位于`pkg/plugin/internal/guest`

#### Scenario: 动态插件导入 data SDK

- **WHEN** 动态插件 guest 代码需要构造受治理 data 查询、变更或事务
- **THEN** data SDK 位于`pkg/plugin/capability/data`
- **AND** data SDK 不得位于`pkg/plugin/capability/internal`或`pkg/plugin/internal`

#### Scenario: 源码插件导入业务上下文能力

- **WHEN** 源码插件服务需要读取当前请求身份、租户或代管上下文
- **THEN** 业务上下文能力位于`pkg/plugin/capability/bizctx`或`pkg/plugin/capability/contract`
- **AND** 不得放入`pkg/plugin/internal/bizctx`

#### Scenario: 宿主插件运行时复用内部实现

- **WHEN** `apps/lina-core/internal/service/plugin/...`需要复用插件运行时内部实现
- **THEN** 该实现必须位于宿主`internal/service/plugin/internal/...`或其他宿主可导入边界
- **AND** 不得放入`pkg/plugin/internal`后再由宿主运行时直接导入

### Requirement: capability 私有实现必须保留在`capability/internal`

系统 SHALL 保留`pkg/plugin/capability/internal`作为 capability 目录自己的私有实现边界。只服务 capability 的 provider registry、provider lazy loading、fallback、delegation、冲突检测和 capability 状态治理实现 MUST 位于`pkg/plugin/capability/internal`或 capability 子包私有实现中，不得迁入`pkg/plugin/internal`扩大给`pluginhost`或`pluginbridge`可导入。

#### Scenario: capability registry 只服务能力目录

- **WHEN** 系统维护 capability provider factory registry、懒加载 provider 实例或 provider 冲突检测
- **THEN** 相关实现位于`pkg/plugin/capability/internal/capabilityregistry`
- **AND** `pkg/plugin/pluginhost`和`pkg/plugin/pluginbridge`不得直接 import 该实现

#### Scenario: 跨公共组件共享实现

- **WHEN** 某个内部实现确实同时服务`pluginhost`、`pluginbridge`和`capability`中的多个公共组件
- **THEN** 该实现可以放入`pkg/plugin/internal`下职责明确的子包
- **AND** 不得把仅服务 capability 的实现迁入总`internal`以追求目录统一

### Requirement: 旧`pkg/pluginservice`公共入口必须删除

系统 SHALL 删除旧`apps/lina-core/pkg/pluginservice`公共入口，并以`apps/lina-core/pkg/plugin/capability`承载原插件消费宿主能力目录语义。迁移完成后生产代码、官方插件和动态插件样例不得继续 import 旧路径。

#### Scenario: 生产代码导入旧`pluginservice`

- **WHEN** 静态检索发现生产 Go 代码 import `lina-core/pkg/pluginservice`
- **THEN** 验证必须失败或审查必须阻断
- **AND** 调用方必须迁移到`lina-core/pkg/plugin/capability`

#### Scenario: 官方插件消费宿主能力

- **WHEN** 官方源码插件需要使用配置、业务上下文、组织、租户或 manifest 能力
- **THEN** 插件 import `lina-core/pkg/plugin/capability/**`
- **AND** 不得保留旧`lina-core/pkg/pluginservice/**`路径

### Requirement: 旧`pkg/plugindb`公共入口必须删除

系统 SHALL 删除`apps/lina-core/pkg/plugindb`公共入口，并以`apps/lina-core/pkg/plugin/capability/data`承载动态插件 data SDK、typed data plan 和宿主侧治理 facade。迁移完成后生产代码、官方插件、动态插件样例、CI smoke fixture 和开发工具测试不得继续 import 或复制旧路径。

`apps/lina-core/pkg/plugin/capability/data`组件自身 SHALL 使用与目录职责一致的`package data`、`data.go`主文件和`data_*.go`文件前缀，不得继续保留旧`plugindb`包名或旧组件名前缀。

#### Scenario: 动态插件导入 data SDK

- **WHEN** 动态插件需要访问受治理 data host service
- **THEN** 插件 import `lina-core/pkg/plugin/capability/data`
- **AND** 不得保留旧`lina-core/pkg/plugindb`路径
- **AND** 不得依赖该组件的旧`plugindb`包名

#### Scenario: 宿主 datahost 解析 typed data plan

- **WHEN** 宿主 datahost 需要解码 typed query plan、附加审计上下文或获取受治理 DB wrapper
- **THEN** 宿主 datahost import `lina-core/pkg/plugin/capability/data`
- **AND** 不得直接 import `pkg/plugin/capability/data/internal/**`

#### Scenario: 静态检索发现旧`plugindb`

- **WHEN** 静态检索发现生产 Go 代码、动态插件样例、CI smoke fixture 或`linactl`测试仍引用`apps/lina-core/pkg/plugindb`或`lina-core/pkg/plugindb`
- **THEN** 验证必须失败或审查必须阻断
- **AND** 调用方必须迁移到`apps/lina-core/pkg/plugin/capability/data`或`lina-core/pkg/plugin/capability/data`

### Requirement: 旧`pkg/sourceupgrade`公共入口必须删除

系统 SHALL 删除`apps/lina-core/pkg/sourceupgrade`公共入口。源码插件升级发现、对比、升级执行和结果状态属于宿主插件运行时内部治理，MUST 由`apps/lina-core/internal/service/plugin`及其内部 sourceupgrade 组件承载。

#### Scenario: 源码插件升级治理执行

- **WHEN** 管理员触发源码插件升级
- **THEN** 宿主通过插件管理 API 和`internal/service/plugin`服务方法执行升级治理
- **AND** 业务代码不得直接 import `lina-core/pkg/sourceupgrade`

#### Scenario: 源码插件声明升级资源

- **WHEN** 源码插件需要声明升级回调、升级 SQL 或生命周期资源
- **THEN** 插件通过`pluginhost`生命周期和插件 manifest 资源声明
- **AND** 插件不得依赖宿主内部 source upgrade scanner、executor 或公共`pkg/sourceupgrade`facade

### Requirement: 包边界迁移必须保持运行时语义不变

系统 SHALL 将本次迁移视为包命名和公共边界重构。除公开 import 路径和组件命名外，源码插件注册行为、动态插件 bridge 协议、`hostServices`授权快照、插件能力 DTO、数据权限边界、缓存失效策略和插件生命周期资源语义 MUST 保持不变。

#### Scenario: 动态插件 bridge 协议不变

- **WHEN** 动态插件通过 WASM ABI、host call 或 host service envelope 与宿主交互
- **THEN** protobuf 字段、service/method 字符串、授权快照和错误 envelope 保持兼容当前目标模型
- **AND** 仅 Go import 路径迁移到`pkg/plugin/pluginbridge`、`pkg/plugin/capability/guest`或`pkg/plugin/capability/data`

#### Scenario: 源码插件能力消费语义不变

- **WHEN** 源码插件通过能力目录读取配置、manifest、组织、租户或业务上下文
- **THEN** 返回 DTO、降级策略、数据权限和缓存一致性策略保持不变
- **AND** 仅公共包路径从`pkg/pluginservice`迁移到`pkg/plugin/capability`

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

#### Scenario: 租户过滤能力归属

- **WHEN** 源码插件需要使用插件自有表租户过滤能力
- **THEN** 过滤接口归属`pkg/plugin/capability/tenantcap`下的源码插件专用接口
- **AND** 不得继续保留`pkg/plugin/capability/tenantfilter`公开包
- **AND** 不得新增独立`pkg/plugin/capability/tenantfiltercap`能力组件包

### Requirement: 旧非 *cap 能力包不得作为生产入口保留

系统 SHALL 在迁移完成后删除旧非`*cap`公开能力包入口。生产代码、官方插件、动态插件样例和测试替身 MUST 不再导入旧非`*cap`具体能力包。同时不得新增`capability/tenantfiltercap`独立能力组件，也不得通过根`Services.Config()`、`Services.PluginConfig()`、`Services.PluginLifecycle()`或`Services.PluginState()`访问插件相关能力。

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

### Requirement: capability 普通契约与 SPI 子包边界必须受治理验证

系统 SHALL 要求`pkg/plugin/capability/**`普通生产契约保持无 GoFrame HTTP 和数据库 query builder 依赖。除路径段以`spi`结尾的源码插件 provider SPI 子包外，`capability/**`非测试生产代码 MUST NOT import `github.com/gogf/gf/v2/database/gdb`或`github.com/gogf/gf/v2/net/ghttp`。该约束 MUST 由随`go test`执行的治理测试持续验证。

#### Scenario: 治理测试验证普通 capability 不导入 gdb

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `capability/**`中非`*spi`子包的非测试源文件不存在`github.com/gogf/gf/v2/database/gdb` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 治理测试验证普通 capability 不导入 ghttp

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `capability/**`中非`*spi`子包的非测试源文件不存在`github.com/gogf/gf/v2/net/ghttp` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: SPI 子包允许宿主接缝类型

- **WHEN** `tenantspi`或`orgspi`需要表达数据库 scope helper、request resolver 或 provider runtime
- **THEN** 对应 SPI 子包可以 import `gdb`或`ghttp`
- **AND** 该豁免不得扩散到父级`tenantcap`、`orgcap`或其他普通能力包

### Requirement: pluginbridge 不得依赖源码插件 Provider SPI

系统 SHALL 将`pkg/plugin/pluginbridge`限定为动态插件 ABI、transport、公开协议和动态插件专属 guest SDK。`pluginbridge/**`非测试生产代码 MUST NOT import `pkg/plugin/capability/**`下路径段以`spi`结尾的源码插件 provider SPI 子包。该约束 MUST 由随`go test`执行的治理测试持续验证。

### Requirement: 跨插件 import 边界必须由治理扫描覆盖

系统 SHALL 提供静态治理扫描或等价验证，确保跨插件生产 Go import 只允许依赖目标插件的`backend/cap/...`公开契约。扫描 MUST 阻断跨插件 import `backend/internal/...`、`backend/internal/dao`、`backend/internal/model`、`backend/api`、controller、service 实现、私有 provider adapter 和`backend/pkg`领域能力入口。测试 fixture、代码生成输入和受控开发工具例外 MUST 在扫描规则中显式限定目录和用途。

#### Scenario: 生产代码 import owner internal

- **WHEN** 插件生产代码 import `lina-plugin-linapro-ai-core/backend/internal/service/ai`
- **THEN** 治理扫描 MUST 失败
- **AND** 调用方必须改为依赖`lina-plugin-linapro-ai-core/backend/cap/aicap/...`

#### Scenario: 测试例外受限

- **WHEN** 同插件内部测试需要访问`backend/internal`实现
- **THEN** 该 import MAY 被允许
- **AND** 例外不得允许其他插件生产代码跨插件访问 internal 实现
