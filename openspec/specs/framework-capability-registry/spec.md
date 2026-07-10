# framework-capability-registry Specification

## Purpose
TBD - created by archiving change refine-plugin-capability-boundaries. Update Purpose after archive.
## Requirements
### Requirement: 框架能力必须按领域归属独立 pluginservice 组件

系统 SHALL 通过 core-owned 与 plugin-owned 两类路径维护插件可消费框架能力。core-owned 能力继续通过`pkg/plugin/capability/orgcap`、`pkg/plugin/capability/tenantcap`等独立`*cap`组件维护由宿主定义契约且由插件提供实现的框架能力。plugin-owned 非核心能力 MUST 通过 owner 插件`backend/cap/<domain>cap`维护 capability ID、版本、DTO、普通消费`Service`接口、fallback/delegation、错误语义、动态 guest SDK 和 provider SPI。共享 provider registry、provider factory 声明、懒加载实例缓存、冲突检测和 manager 实现可以继续由 core 承载，但 core 对 plugin-owned 能力只能感知通用 descriptor、owner 插件 ID、capability version、method、风险和资源形态。系统 MUST NOT 为 plugin-owned 非核心能力继续新增`lina-core/pkg/plugin/capability/<domain>cap`公开契约、领域专属 provider facade、动态 wire codec 或 WASM dispatcher 分支。

#### Scenario: 组织能力由 orgcap 组件维护

- **WHEN** 系统定义`framework.org.v1`能力
- **THEN** 该能力的 capability ID、DTO、普通消费`Service`接口和错误码位于`pkg/plugin/capability/orgcap`公开父包
- **AND** 组织 provider SPI、provider env、scope helper 和 host provider runtime 位于`pkg/plugin/capability/orgcap/orgspi`
- **AND** fallback 和 delegation 位于`pkg/plugin/capability/orgcap`或`pkg/plugin/capability/orgcap/orgspi`中与其职责一致的位置
- **AND** 共享 provider registry、懒加载实例缓存和冲突治理位于`pkg/plugin/capability/internal/capabilityregistry`
- **AND** 消费方不得依赖提供方插件的 provider adapter 或内部业务 service

#### Scenario: 租户能力由 tenantcap 组件维护

- **WHEN** 系统定义`framework.tenant.v1`能力
- **THEN** 该能力的 capability ID、DTO、普通消费`Service`接口和错误码位于`pkg/plugin/capability/tenantcap`公开父包
- **AND** 租户 provider SPI、request resolver、scope helper、membership provider、plugin table filter 和 host provider runtime 位于`pkg/plugin/capability/tenantcap/tenantspi`
- **AND** 消费方通过显式注入的`tenantcap.Service`或`capability.Services.Tenant()`获取租户普通消费能力实例
- **AND** 系统不得要求消费方 import `pkg/frameworkcap`、`pkg/tenantcap`、`pkg/pluginservice`或宿主`internal/service/tenantcap`

#### Scenario: AI 能力由 linapro-ai-core backend/cap 维护

- **WHEN** 系统定义`AI`文本、多模态或后续非核心`AI`能力
- **THEN** 该能力的 capability ID、DTO、普通消费`Service`接口、错误码、方法状态、provider SPI 和动态 guest SDK MUST 位于`apps/lina-plugins/linapro-ai-core/backend/cap/aicap`
- **AND** core 只能通过通用 capability descriptor 注册、发现、授权和路由该能力
- **AND** core 不得继续保留`pkg/plugin/capability/aicap`作为生产公开契约 owner

#### Scenario: 能力契约不泄漏内部模型

- **WHEN** capability service 返回组织、租户、用户、`AI`或可见范围信息
- **THEN** 返回值使用该能力契约自有的 DTO、投影或值对象
- **AND** 返回值不得包含宿主或插件内部`DAO`、`DO`、`Entity`、`*gdb.Model`、provider 密钥、模型路由表或私有缓存结构

### Requirement: 插件必须通过 Provider Factory 声明框架能力实现

系统 SHALL 要求源码插件在 registrar 阶段声明其对框架能力或 plugin-owned 能力的实现。core-owned 能力 MAY 继续通过`pluginhost.Declarations`的强类型 provider 声明 facade 声明，例如`plugin.Providers().ProvideOrg(...)`或`plugin.Providers().ProvideTenant(...)`。plugin-owned 非核心能力 MUST 通过通用 capability descriptor 声明，并由 owner 插件提供类型安全 helper 将领域 factory 包装为 descriptor；`pluginhost`不得为每个非核心领域新增`Provide<Domain>`方法。Provider 实例 MUST 由消费 service 或 owner handler 在使用时通过插件 enabled snapshot、owner 启用策略和 descriptor 版本判断可用。插件不得在路由注册、controller 构造、业务请求路径或能力包级`Provide()`函数中直接写入全局 provider 注册表。

#### Scenario: 源码插件声明组织能力 Provider

- **WHEN** `linapro-org-core`提供`framework.org.v1`实现
- **THEN** 插件入口通过`pluginhost.Declarations`的 provider 声明 facade 声明一个`orgspi.ProviderFactory`
- **AND** 消费 service 在调用组织能力时通过`PluginStateService.IsProviderEnabled(ctx, "linapro-org-core")`判断 provider 插件是否平台级可用
- **AND** provider 插件平台级可用时，`pkg/plugin/capability/internal/capabilityregistry`中的 manager 使用该插件声明的 factory 懒加载 provider 实例
- **AND** 路由注册回调不得直接调用全局`RegisterProvider(provider)`或旧能力包级`Provide()`完成激活

#### Scenario: 源码插件声明租户能力 Provider

- **WHEN** `linapro-tenant-core`提供`framework.tenant.v1`实现
- **THEN** 插件入口通过`pluginhost.Declarations`的 provider 声明 facade 声明一个`tenantspi.ProviderFactory`
- **AND** 宿主注册该 factory 时记录声明插件 ID，并在 provider 使用路径继续按插件 enabled snapshot 判断可用性
- **AND** 租户 scope、membership 和 request resolver 行为保持由 provider 实例实现

#### Scenario: 源码插件声明 AI 文本 Provider

- **WHEN** `linapro-ai-core`提供文本`AI`能力实现
- **THEN** 插件入口 MUST 通过`aicap`或`aicap/spi`提供的类型安全 helper 注册通用 capability descriptor
- **AND** descriptor MUST 记录 owner 插件 ID、`ai`或`ai.text`能力键、`v1`协议版本、方法、风险、资源形态和 provider factory
- **AND** `pluginhost`不得继续提供`ProvideAIText`专属方法
- **AND** provider 使用路径继续按 owner 插件 enabled snapshot、descriptor 版本和 provider 状态判断可用性

#### Scenario: 插件禁用后 Provider 不再被使用

- **WHEN** 提供 core-owned 或 plugin-owned capability 的插件被禁用、卸载或升级失败
- **THEN** 对应 enabled snapshot 或 owner 可用性检查返回 false
- **AND** 消费 service 或 owner handler 不再返回或调用该插件声明的 provider
- **AND** 消费 service 的`Available(ctx)`、方法状态或等价状态反映该能力不可用或降级状态

### Requirement: Provider 实现必须封装在提供方插件内部

系统 SHALL 将官方插件的 pluginservice capability provider adapter 视为插件内部实现。Provider adapter MUST 默认位于`apps/lina-plugins/<plugin-id>/backend/internal/provider/<capability>adapter/`，真实业务实现 MUST 位于`backend/internal/service/`；`backend/plugin.go`只负责声明路由、生命周期和 provider factory，不得承载 provider 业务实现。

#### Scenario: Provider Adapter 不作为公开包暴露

- **WHEN** 一个官方源码插件实现`framework.tenant.v1`
- **THEN** 该插件的 tenant provider adapter 位于`backend/internal/provider/tenantadapter/`
- **AND** 其他插件无法通过 Go import 直接依赖该 adapter
- **AND** 其他插件只能通过 pluginservice capability 消费 service 使用该能力

#### Scenario: 业务逻辑保留在插件 Service

- **WHEN** provider adapter 需要访问插件业务数据或领域逻辑
- **THEN** adapter 通过显式依赖调用同插件`backend/internal/service/`中的业务 service
- **AND** adapter 只做 pluginservice capability 契约到内部业务 service 的薄适配

### Requirement: 插件消费框架能力必须通过消费 Service

系统 SHALL 要求宿主模块、源码插件和动态插件通过`pkg/plugin/capability/<domain>cap`的消费 service 使用能力。消费方 MUST 不直接获取 provider 实例，也不得直接 import 提供方插件的 provider adapter、内部 service、DAO、Entity 或其他内部实现；消费方在需要硬阻断时 MUST 使用既有 provider 插件依赖，在可选使用能力时 MUST 支持运行时可用性降级语义。

#### Scenario: 源码插件消费组织能力

- **WHEN** 源码插件需要读取组织树或批量解析用户组织信息
- **THEN** 插件通过`capability.Services.Org()`或等价注入的`orgcap.Service`发起调用
- **AND** 插件不得 import `linapro-org-core/backend/internal/**`

#### Scenario: 动态插件消费组织能力

- **WHEN** 动态插件需要消费`framework.org.v1`
- **THEN** guest SDK 通过`capability/guest`发起版本化 host service 调用
- **AND** 宿主将调用路由到同一个`orgcap.Service`
- **AND** 调用必须满足动态插件`hostServices`授权；若消费方需要硬依赖具体 provider 插件，则由既有`dependencies.plugins`声明和生命周期校验表达

#### Scenario: 可选能力不可用时降级

- **WHEN** 插件可选使用`framework.org.v1`
- **AND** 当前环境没有可用组织能力 provider
- **THEN** 插件仍可启用
- **AND** 插件必须通过`Available(ctx)`或等价能力状态隐藏相关功能、返回零值或执行规范定义的降级行为

### Requirement: 框架能力必须治理 Provider 冲突和版本

系统 SHALL 为 pluginservice capability 定义 provider 选择、冲突处理和版本兼容规则。单例能力 MUST 在多个平台级可用插件同时声明同一个 capability provider 时进入明确冲突状态，或按规范定义的确定性优先级选择；不兼容契约变更 MUST 使用新的 capability 版本，不得破坏既有`v1`消费契约。

#### Scenario: 单例能力重复 Provider 被拒绝

- **WHEN** 两个已启用插件同时声明同一个单例 capability 的 provider
- **THEN** pluginservice capability manager 按平台级 provider 可用性快照进入明确冲突状态，或按该能力规范选择唯一 provider
- **AND** 错误或状态包含 capability ID 和冲突插件 ID

#### Scenario: 不兼容契约使用新版本

- **WHEN** `framework.org.v1`无法兼容新增的响应结构、错误语义或授权边界
- **THEN** 系统定义`framework.org.v2`
- **AND** 已声明依赖`framework.org.v1`的插件继续按`v1`契约运行或降级

### Requirement: 框架能力状态必须保持缓存一致性

系统 SHALL 将插件平台级 provider 可用性快照视为关键运行时数据。插件安装、启用、禁用、卸载、升级、同版本刷新、既有插件依赖状态变化或租户状态变化后，系统 MUST 刷新插件 enabled snapshot 并使`PluginStateService.IsProviderEnabled(ctx, pluginID)`反映最新状态；集群模式下不得只刷新当前节点本地内存。

#### Scenario: 插件启用后集群节点刷新能力状态

- **WHEN** 集群模式下启用一个提供`framework.tenant.v1`的插件
- **THEN** 主节点更新插件 enabled snapshot 权威数据并发布插件运行时修订或等价事件
- **AND** 其他节点观察到事件后刷新本地插件 enabled snapshot
- **AND** 后续 provider 使用路径通过`PluginStateService.IsProviderEnabled(ctx, pluginID)`读取刷新后的平台级可用性

#### Scenario: 只读能力检查不触发全局失效

- **WHEN** 插件或宿主只读取某个 capability 的`Available(ctx)`状态
- **THEN** 系统不得写入插件 registry 或清空无关插件、语言包、路由或前端 bundle 缓存
- **AND** 只读检查不得产生跨实例失效事件

### Requirement: 通用 capability descriptor 必须可治理和可测试

系统 SHALL 为通用 capability descriptor 提供结构化校验、重复注册检查、版本匹配、owner 插件启用状态检查、方法风险分类、资源形态校验和审计投影。descriptor 校验 MUST 在启动、源码插件同步、动态 artifact 加载和 owner 插件升级路径执行，并通过单元测试覆盖未知 owner、重复 method、版本不满足、依赖缺失、禁用 owner 和 provider 冲突。

#### Scenario: descriptor 注册缺少 owner

- **WHEN** 源码插件或动态 artifact 注册 capability descriptor 但`ownerPluginId`为空或不等于声明插件 ID
- **THEN** 系统 MUST 拒绝注册
- **AND** 错误必须包含声明插件 ID、owner 插件 ID 和 capability key

#### Scenario: descriptor 版本不满足消费声明

- **WHEN** 动态插件声明`owner: linapro-ai-core`、`service: ai`和`version: v1`
- **AND** owner 插件只发布不兼容版本
- **THEN** 安装、启用或运行时授权 MUST 被拒绝
- **AND** 拒绝结果必须进入依赖或 host service 授权诊断

