# plugin-package-boundary-governance Specification

## Purpose
TBD - created by archiving change refactor-plugin-package-boundaries. Update Purpose after archive.
## Requirements
### Requirement: 插件公共命名空间必须集中到`pkg/plugin`

系统 SHALL 将`lina-core`拥有的插件相关公共 Go 组件集中到`apps/lina-core/pkg/plugin/`命名空间下。该命名空间下的公开顶层组件必须按职责拆分为源码插件贡献入口、动态插件桥接协议和 core-owned 插件消费宿主能力目录，不得继续在`apps/lina-core/pkg/`根层新增语义模糊的插件公共组件。plugin-owned 非核心领域能力的公开契约不属于 core 公共命名空间，MUST 位于 owner 插件`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`，并受跨插件 import 边界治理。

#### Scenario: 开发者定位 core 插件公共入口

- **WHEN** 开发者需要查找 LinaPro 插件内核、源码插件贡献 API、动态桥接协议或 core-owned 宿主能力契约
- **THEN** 系统在`apps/lina-core/pkg/plugin/`下提供对应公共组件
- **AND** 不要求开发者在`pkg/pluginservice`、`pkg/pluginhost`、`pkg/pluginbridge`和其他顶层插件包之间猜测职责边界

#### Scenario: 开发者定位 plugin-owned 领域契约

- **WHEN** 开发者需要查找`linapro-ai-core`拥有的`AI`领域公开契约
- **THEN** 系统在`apps/lina-plugins/linapro-ai-core/backend/cap/aicap`下提供对应契约和 SDK
- **AND** 开发者不得在`lina-core/pkg/plugin/capability/aicap`继续寻找生产 owner 契约

#### Scenario: 新增 core 插件公共组件

- **WHEN** 系统需要新增插件开发者或宿主插件运行时共享的 core-owned 公共契约
- **THEN** 该契约必须优先放入`apps/lina-core/pkg/plugin/`下职责明确的公开子包
- **AND** 不得新增`pluginservice`、`plugincommon`、`pluginutil`等语义模糊的顶层公共包

#### Scenario: 新增 plugin-owned 领域公共契约

- **WHEN** 系统需要新增非核心领域 owner 插件公开契约
- **THEN** 该契约必须放入 owner 插件`backend/cap/<domain>cap`
- **AND** 不得为了方便跨插件 import 而放入 owner 插件`backend/internal`、`backend/pkg`或 core `pkg/plugin/capability`

### Requirement: `pluginhost`、`pluginbridge`和`capability`必须职责分离

系统 SHALL 在 core `pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API、源码插件获取统一服务目录和通用 capability descriptor 接收；`pluginbridge`只负责动态插件 ABI、WASM transport、公开协议出口、core-owned 动态插件公共入口和 owner-aware 通用 host service envelope；`capability`只负责 core-owned 插件消费宿主能力的稳定目录、公共原语和`*cap`能力契约。plugin-owned 非核心领域能力的公开契约、动态 guest SDK 和 provider SPI MUST 归属 owner 插件`backend/cap/<domain>cap`及其子包。仅服务动态插件的 core guest SDK（如 record store）MUST 归属`pluginbridge`，不得放入 owner 插件 cap；仅服务 owner 领域的 guest SDK MUST 归属 owner 插件 cap。

#### Scenario: 源码插件注册贡献

- **WHEN** 源码插件需要注册路由、hook、cron、生命周期回调或 core-owned provider factory
- **THEN** 插件使用`pkg/plugin/pluginhost`
- **AND** `pluginhost`不得拥有宿主能力业务实现
- **AND** `pluginhost`不得暴露`Admin()`或`AdminServices`管理目录
- **AND** `pluginhost.Services`顶层不得成为新增治理能力的事实 owner，治理能力必须委托到对应领域`Service`

#### Scenario: 源码插件注册 plugin-owned provider

- **WHEN** 源码插件需要声明`AI`或其他 plugin-owned 非核心领域 provider
- **THEN** 插件使用 owner 插件`backend/cap/<domain>cap/spi`或等价 helper 构造通用 descriptor
- **AND** `pluginhost`只接收通用 descriptor
- **AND** `pluginhost`不得 import owner 插件 cap 包来定义领域专属 facade

#### Scenario: 动态插件执行 bridge 调用

- **WHEN** 动态插件需要声明 route、处理 WASM request envelope 或调用 host call transport
- **THEN** 插件使用`pkg/plugin/pluginbridge/guest`和`pkg/plugin/pluginbridge/protocol`
- **AND** `pluginbridge`不得定义业务能力可用性、provider 激活、数据权限降级或配置读取语义
- **AND** `pluginbridge`根包不得重新导出协议 DTO、常量、codec、guest helper 或`Runtime()`、`Data()`、`RecordStore()`、`Cron()`等能力 client facade

#### Scenario: 动态插件消费 plugin-owned 能力

- **WHEN** 动态插件需要调用 owner 插件发布的`AI`能力
- **THEN** 插件使用 owner 插件`backend/cap/aicap/bridge`或等价公开 guest SDK
- **AND** SDK 只负责编码、声明 helper 和调用通用 host call，不得绕过`hostServices`授权、owner 依赖和宿主审计

#### Scenario: 插件消费 core-owned 宿主能力

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户或业务上下文等 core-owned 宿主能力
- **THEN** 插件使用`pkg/plugin/capability`、对应`pkg/plugin/capability/<domain>cap`能力组件或 core guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 record store 能力

- **WHEN** 动态插件需要使用 ORM-style record store facade、typed record store plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/pluginbridge/recordstore`
- **AND** Go guest 能力目录通过`RecordStore()`返回该 facade
- **AND** 不得继续通过`pkg/plugin/capability/recordstore`、owner 插件 cap 或顶层`pkg/plugindb`暴露该能力

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

#### Scenario: 动态插件 bridge 不导入 tenantspi

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `pluginbridge/**`非测试生产源文件不存在`pkg/plugin/capability/tenantcap/tenantspi` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 动态插件 bridge 不导入 orgspi

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `pluginbridge/**`非测试生产源文件不存在`pkg/plugin/capability/orgcap/orgspi` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 测试代码跨边界验证豁免

- **WHEN** `capability`、`pluginhost`或`pluginbridge`的`_test.go`文件为治理测试或集成验证 import SPI、`gdb`或`ghttp`
- **THEN** 治理测试不将其判定为违规
- **AND** 豁免仅适用于测试代码，不适用于任何生产源文件

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

