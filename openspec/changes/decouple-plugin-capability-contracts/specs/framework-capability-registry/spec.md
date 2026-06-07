## MODIFIED Requirements

### Requirement: 框架能力必须按领域归属独立 pluginservice 组件

系统 SHALL 通过`pkg/plugin/capability/orgcap`、`pkg/plugin/capability/tenantcap`、`pkg/plugin/capability/aicap`等独立`*cap`组件维护由宿主定义契约且由插件提供实现的框架能力。每个能力组件 MUST 直接维护自身 capability ID、版本、DTO、消费`Service`接口、provider factory 声明 facade、fallback/delegation 和必要错误类型；共享 provider registry、provider factory 声明、懒加载实例缓存、冲突检测和 manager 实现 MUST 放入`pkg/plugin/capability/internal/capabilityregistry`。系统 MUST NOT 再新增或保留`pkg/pluginservice`聚合包、`pkg/frameworkcap`聚合包、旧`pkg/orgcap`/`pkg/tenantcap`兼容包或宿主`internal/service/orgcap`、`internal/service/tenantcap`双重适配层。

#### Scenario: 组织能力由 orgcap 组件维护

- **WHEN** 系统定义`framework.org.v1`能力
- **THEN** 该能力的 capability ID、DTO、`Service`接口和`Provide(...)`声明 facade 位于`pkg/plugin/capability/orgcap`公开边界下
- **AND** fallback 和 delegation 位于`pkg/plugin/capability/orgcap`
- **AND** 共享 provider registry、懒加载实例缓存和冲突治理位于`pkg/plugin/capability/internal/capabilityregistry`
- **AND** 消费方不得依赖提供方插件的 provider adapter 或内部业务 service

#### Scenario: 租户能力由 tenantcap 组件维护

- **WHEN** 系统定义`framework.tenant.v1`能力
- **THEN** 该能力的 capability ID、DTO、`Service`接口和`Provide(...)`声明 facade 位于`pkg/plugin/capability/tenantcap`公开边界下
- **AND** 消费方通过显式注入的`tenantcap.Service`或`capability.Services.Tenant()`获取租户能力实例
- **AND** 系统不得要求消费方 import `pkg/frameworkcap`、`pkg/tenantcap`、`pkg/pluginservice`或宿主`internal/service/tenantcap`

#### Scenario: 源码插件租户过滤子接口由 tenantcap 维护

- **WHEN** 源码插件需要对插件自有表按当前租户追加`tenant_id`过滤
- **THEN** 过滤接口位于`pkg/plugin/capability/tenantcap`公开边界下
- **AND** 入口只通过`pluginhost.Services.TenantFilter()`暴露给源码插件
- **AND** 该接口不得成为`tenantcap.Service`普通租户消费面的一部分
- **AND** 动态插件 guest SDK 和`hostServices`协议不得暴露该接口

#### Scenario: 能力契约不泄漏内部模型

- **WHEN** capability service 返回组织、租户、用户或可见范围信息
- **THEN** 返回值使用该能力契约自有的 DTO、投影或值对象
- **AND** 返回值不得包含宿主或插件内部`DAO`、`DO`、`Entity`、`*gdb.Model`或私有缓存结构

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
