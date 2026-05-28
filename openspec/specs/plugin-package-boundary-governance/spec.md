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

