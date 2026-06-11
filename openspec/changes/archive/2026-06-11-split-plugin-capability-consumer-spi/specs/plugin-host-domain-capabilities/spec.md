# plugin-host-domain-capabilities 规范增量

## ADDED Requirements

### Requirement: 普通领域能力契约必须与 Provider SPI 分离

系统 SHALL 将插件普通消费领域能力契约与源码插件 provider SPI、宿主内部 scope 接缝分离。普通`capability/<domain>cap`父包 MUST 只暴露普通消费`Service`、领域 DTO、值对象、错误码和常量；凡是需要`*gdb.Model`、`*ghttp.Request`、provider factory、provider runtime、provider env 或宿主内部 scope helper 的接口 MUST 放入对应`*spi`子包或宿主内部包。

#### Scenario: 普通插件消费租户能力

- **WHEN** 源码插件或动态插件通过`tenantcap.Service`消费租户能力
- **THEN** 该父包接口不暴露`*gdb.Model`、`*ghttp.Request`、provider factory 或 provider runtime
- **AND** 插件只看到租户 DTO、状态、批量投影、候选和普通消费方法

#### Scenario: Provider 插件实现租户能力

- **WHEN** 源码 provider 插件需要实现租户解析、membership、scope 或插件表过滤
- **THEN** 它 import `pkg/plugin/capability/tenantcap/tenantspi`
- **AND** 该 SPI 可以使用`*gdb.Model`或`*ghttp.Request`表达宿主内部数据库过滤和请求解析接缝
- **AND** 这些 SPI 不进入动态插件 guest SDK 或`hostServices`协议

#### Scenario: Provider 插件实现组织能力

- **WHEN** 源码 provider 插件需要实现组织 scope 或 provider runtime
- **THEN** 它 import `pkg/plugin/capability/orgcap/orgspi`
- **AND** `orgspi.ProviderEnv`可以引用`tenantspi.PluginTableFilterService`
- **AND** 普通`orgcap.Service`不暴露`*gdb.Model`或 provider SPI

### Requirement: 动态路由和 API 文档能力不得暴露 HTTP 框架对象

系统 SHALL 要求普通领域能力契约使用`context.Context`、路径、方法、DTO 或中立值对象传递请求相关信息，不得在普通`capability/**`父包中暴露`*ghttp.Request`或`*ghttp.HandlerItemParsed`。

#### Scenario: 动态路由元数据读取

- **WHEN** 插件通过`routecap.Service.DynamicRouteMetadata`读取当前动态路由元数据
- **THEN** 方法签名接收`context.Context`
- **AND** 宿主适配器在内部从上下文恢复 HTTP 请求并读取元数据
- **AND** `routecap`父包不 import `ghttp`

#### Scenario: API 文档 operation key 派生

- **WHEN** 插件或源码插件中间件需要派生 API 文档 operation key
- **THEN** 它使用不依赖`ghttp`的路径和方法派生入口
- **AND** 需要`*ghttp.HandlerItemParsed`的宿主私有逻辑不得放入`apidoccap`普通契约包

### Requirement: 数据权限过滤迁移必须保持数据库侧注入语义

系统 SHALL 将租户与组织 scope 接口迁移视为类型归属重构。租户过滤、组织部门过滤、用户 membership 过滤和插件自有表租户过滤 MUST 继续在数据库查询阶段注入约束，不得因为 SPI 拆分退化为内存过滤、放开全量数据或改变拒绝策略。

#### Scenario: 宿主列表查询使用迁移后的租户 scope

- **WHEN** 宿主列表、下拉候选或批量读取路径需要按当前租户过滤
- **THEN** 调用方使用`tenantspi.ScopeService`或等价窄接口在`*gdb.Model`上追加数据库过滤
- **AND** 不得先读取全量结果再在内存中过滤
- **AND** 过滤前后的租户可见性语义保持不变

#### Scenario: 宿主列表查询使用迁移后的组织 scope

- **WHEN** 宿主列表、下拉候选或批量读取路径需要按部门或用户组织范围过滤
- **THEN** 调用方使用`orgspi.ScopeService`或等价窄接口在`*gdb.Model`上追加数据库过滤
- **AND** 不得通过关联名称、分页总数或候选项泄露范围外数据存在性
- **AND** 过滤前后的组织数据权限语义保持不变

