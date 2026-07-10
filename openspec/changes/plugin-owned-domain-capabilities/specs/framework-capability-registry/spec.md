## MODIFIED Requirements

### Requirement: 框架能力必须按领域归属独立维护

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

### Requirement: 插件必须通过 Provider Factory 或 Capability Descriptor 声明能力实现

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

## ADDED Requirements

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
