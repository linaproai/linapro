## Context

变更 A 已将`capability`固定为`pkg/plugin`底层契约层，并通过 import 边界测试阻断`capability`反向依赖`pluginbridge`或`pluginhost`。当前剩余问题集中在能力包内部边界：

- `tenantcap`和`orgcap`父包同时承载普通插件消费`Service`、provider SPI、宿主 scope helper 和 request resolver，导致`*gdb.Model`、`*ghttp.Request`进入普通插件可见契约包。
- `aitext`、`tenantcap`、`orgcap`各自维护包级`defaultManager`和`Provide()`，provider 声明入口不经过`pluginhost`源码插件注册 facade，运行期 manager 实例 owner 不清晰。
- `routecap.DynamicRouteMetadata(*ghttp.Request)`和`apidoccap.BuildOperationKeyFromHandler(*ghttp.HandlerItemParsed)`把`ghttp`泄漏到普通能力包。

本变更是纯后端和文档治理重构，不改变 wire 协议、HTTP API、数据库 schema、前端行为或运行时用户可见文案。项目无兼容性负担，因此不保留旧 import 路径或双轨过渡。

## Goals / Non-Goals

**Goals:**

- 让普通消费能力包保持无`gdb`、无`ghttp`、无 provider SPI 的纯契约形态。
- 将 tenant/org provider SPI 和宿主内部 scope 接缝迁入`tenantspi`、`orgspi`子包，源码插件 provider 与宿主内部调用显式 import SPI。
- 将 provider factory 声明入口收敛到`pluginhost.Declarations`，由宿主创建并持有`capabilityregistry.Manager`实例，再显式注入 tenant/org/AI text host service。
- 扩展治理测试，防止普通`capability`包重新引入`gdb`/`ghttp`，防止`pluginbridge`导入`*spi`子包。
- 保持租户、组织、AI provider 可用性、数据权限过滤和 fallback/delegation 行为不变。

**Non-Goals:**

- 不新增 HTTP API、DTO、路由或 OpenAPI 元数据。
- 不改变数据库 schema、seed/mock 数据、DAO/DO/Entity。
- 不调整动态插件 host service wire 字符串、payload 编解码或授权快照。
- 不为`aitext`创建空 SPI 子包；`aitext`没有`gdb`/`ghttp`泄漏，只参与 provider 注册机制收敛。
- 不归档或处理本仓库已有的其他完成态活跃变更。

## Decisions

### D-B1：父包保留消费契约，`*spi`子包承载 provider 与宿主 scope 接缝

`tenantcap`父包保留`Service`、DTO、错误码、常量、`ResolverResult`、`TenantFilterContext`等普通消费或中立值对象。`tenantspi`承载`Provider`、`ProviderFactory`、`ProviderEnv`、`ProviderRuntime`、`Resolver`、`RequestResolver`、`ResolverChain`、`ScopeService`、`UserMembershipProvider`、`RuntimeService`、`PluginTableFilterService`和 tenant host implementation。

`orgcap`父包保留`Service`、DTO、错误码、常量、`UserDeptAssignment`、`DeptTreeNode`、`PostOption`等消费契约。`orgspi`承载`Provider`、`ProviderFactory`、`ProviderEnv`、`ProviderRuntime`、`ScopeService`、`RuntimeService`和 org host implementation。

子包可以 import 父包来复用 DTO；父包不得 import 子包。`orgspi.ProviderEnv.TenantFilter`引用`tenantspi.PluginTableFilterService`，SPI 之间的依赖是源码插件/宿主内部接缝，不进入动态插件 guest 编译闭包。

**备选方案：**在父包保留 provider 接口，只把`gdb`方法拆成小接口。该方案无法阻止普通消费方继续看到 provider SPI，也无法通过简单 import 治理断言包纯净性，因此拒绝。

### D-B2：`routecap`使用`context.Context`，`apidoccap`删除 handler helper

`routecap.Service.DynamicRouteMetadata`改为接收`context.Context`。宿主`capabilityhost`适配器内部用`ghttp.RequestFromCtx(ctx)`恢复请求，再调用已有 runtime 元数据读取逻辑；动态插件 guest 实现本来忽略请求对象，改为`context.Context`不改变行为。

`apidoccap.BuildOperationKeyFromHandler`删除，唯一调用方改用`BuildOperationKeyFromPath`或宿主侧等价路径。若实现阶段发现`BuildOperationKeyFromPath`无法保持当前 operation key 语义，则把 handler 专用 helper 下沉到宿主或插件内部调用方，不留在`apidoccap`普通契约包。

**备选方案：**为`routecap`和`apidoccap`分别新增 SPI 子包。两者只有单个`ghttp`泄漏点，且可以直接收窄为现有中立参数；新增子包会增加抽象层和导入成本，因此拒绝。

### D-B3：provider 声明入口归属`pluginhost.Declarations`

`pluginhost.Declarations`新增强类型 provider 声明分组，例如`Capabilities()`或`Providers()`，并提供：

- `ProvideTenant(factory tenantspi.ProviderFactory) error`
- `ProvideOrg(factory orgspi.ProviderFactory) error`
- `ProvideAIText(factory aitext.ProviderFactory) error`

3 个官方 provider 源码插件在`backend/plugin.go`中通过该声明分组注册 provider factory，再调用`pluginhost.RegisterSourcePlugin(plugin)`。`tenantcap.Provide()`、`orgcap.Provide()`、`aitext.Provide()`和包级`defaultManager`删除。

provider manager 实例的 owner 为宿主插件能力装配层。宿主启动期构造 tenant/org/AI text host service 时创建或传入共享`capabilityregistry.Manager[Env]`实例，并通过构造函数显式注入到对应 service；`pluginhost`源码插件定义只保存声明出的 factory，宿主在扫描源码插件定义时把 factory 注册到共享 manager。任务记录必须说明 owner、创建位置、传递路径和共享实例策略。

**备选方案：**保留`Provide()`，只把`defaultManager`改为可注入全局变量。该方案仍保留全局可变状态和绕过 registrar 的声明入口，违反显式依赖注入治理，因此拒绝。

### D-B4：治理测试扩展到 SPI 和普通 capability 纯净性

在`pkg/plugin/plugin_boundary_test.go`扩展现有 AST import 检查：

- `capability/**`中非`*spi`子包、非测试生产代码不得 import `github.com/gogf/gf/v2/database/gdb`或`github.com/gogf/gf/v2/net/ghttp`。
- `pluginbridge/**`非测试生产代码不得 import 路径段以`spi`结尾的`pkg/plugin/capability/**`子包。
- 继续保留变更 A 中`capability`不得 import `pluginbridge/pluginhost`、`pluginhost`不得 import `pluginbridge`的要求。

测试代码豁免只适用于`_test.go`。正反向验证通过临时违规文件完成，验证后删除临时文件。

## Risks / Trade-offs

- `Risk`：源码插件和宿主调用点 import 扇出较大，容易遗漏测试替身或 provider adapter。`Mitigation`：使用`rg`定位旧类型和`Provide()`残留，编译门禁覆盖宿主、provider 插件和动态插件样例。
- `Risk`：provider manager 从包级单例迁移到宿主持有后，注册时序可能与现有包级`init`时序不同。`Mitigation`：provider factory 声明保存在 source plugin definition 中，宿主构造 capability host services 时集中注册；新增或更新测试验证 provider 可用状态、fallback 和冲突治理仍按插件 enabled snapshot 生效。
- `Risk`：租户/组织 scope 类型迁移影响数据权限路径。`Mitigation`：只改类型归属和 import，不改 scope 方法实现语义；运行 tenant/org/datascope 相关回归测试，并在任务记录中写明数据权限无语义变化。
- `Risk`：OpenSpec 基线已有旧`Provide()`要求。`Mitigation`：本变更通过 delta spec 明确新 registrar 声明入口，归档时同步基线。

## Migration Plan

1. 创建`tenantspi`和`orgspi`子包并迁移 SPI 类型、host implementation 与测试。
2. 更新宿主与源码插件调用方 import，保持`pluginhost.Services.TenantFilter()`同名入口。
3. 收窄`routecap`和`apidoccap`的`ghttp`依赖。
4. 扩展`pluginhost.Declarations`并迁移 3 个官方 provider 插件的 provider factory 声明入口。
5. 扩展治理测试、更新双语 README、运行全量验证和`lina-review`。

回滚策略为整体`git revert`；不涉及数据迁移或运行时配置迁移。

## Open Questions

无需要用户澄清的问题。实现阶段唯一需要复核的是`apidoccap.BuildOperationKeyFromPath`对`linapro-monitor-operlog`调用场景是否与旧 helper 等价；若不等价，按 D-B2 的宿主侧 helper 回退方案处理。
