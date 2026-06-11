## Why

`tenantcap`、`orgcap`当前同时承载普通插件消费契约与 provider 实现接缝，导致`*gdb.Model`和`*ghttp.Request`进入普通能力包；`aitext`、`tenantcap`、`orgcap`还通过包级`defaultManager`和`Provide()`维护可变注册表，绕过宿主启动装配和显式依赖注入治理。

本变更在已完成的`pkg/plugin`依赖方向修正基础上继续收敛能力边界：普通消费契约保持在父级`*cap`包，provider SPI 和宿主内部 scope 接缝迁入`*spi`子包，并将 provider 声明入口收敛到`pluginhost.Declarations`。

## What Changes

- **BREAKING**：将`tenantcap`和`orgcap`中的 provider SPI、scope helper、request resolver 与`PluginTableFilterService`迁移到`tenantcap/tenantspi`和`orgcap/orgspi`子包；父级`tenantcap`、`orgcap`仅保留普通消费`Service`、DTO、错误码和常量。
- **BREAKING**：将`pluginhost.Services.TenantFilter()`返回类型切换为`tenantspi.PluginTableFilterService`，源码插件继续通过同名入口获取租户过滤能力。
- **BREAKING**：将 provider factory 声明从`tenantcap.Provide()`、`orgcap.Provide()`和`aitext.Provide()`迁移到`pluginhost.Declarations`的强类型能力 provider 声明分组，删除包级`defaultManager`与`Provide()`入口。
- 将`routecap.DynamicRouteMetadata(*ghttp.Request)`改为`DynamicRouteMetadata(context.Context)`，由宿主适配器在内部从`context.Context`恢复请求元数据。
- 删除或迁移`apidoccap.BuildOperationKeyFromHandler(*ghttp.HandlerItemParsed)`，唯一调用方改用不依赖`ghttp`的 operation key 派生入口。
- 扩展`pkg/plugin`导入边界治理测试：非`*spi`的`capability/**`生产代码不得 import `gdb`或`ghttp`，`pluginbridge/**`生产代码不得 import 任何`*spi`子包。
- 同步`pkg/plugin`双语 README，明确普通消费契约、源码插件 provider SPI、动态插件 guest SDK 三类边界。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `framework-capability-registry`：框架能力 provider factory 的声明入口从能力包级`Provide()`迁移为源码插件 registrar 阶段的`pluginhost.Declarations`强类型声明；provider manager 由宿主装配持有并注入。
- `plugin-host-domain-capabilities`：普通插件消费能力不得暴露`gdb`、`ghttp`或 provider SPI；宿主内部 scope 接缝与源码插件 provider SPI 必须位于`*spi`子包。
- `plugin-package-boundary-governance`：补充`capability`父包与`*spi`子包边界，扩展 import 治理测试要求。
- `service-dependency-injection-governance`：补充 capability provider manager 不得由包级默认单例持有，必须由宿主启动装配创建并共享注入。

## Impact

- 影响 Go 包：`apps/lina-core/pkg/plugin/capability/{tenantcap,orgcap,routecap,apidoccap,aicap/aitext}`、`apps/lina-core/pkg/plugin/pluginhost`、`apps/lina-core/internal/service/plugin/internal/capabilityhost`及相关宿主调用方。
- 影响源码插件：`linapro-tenant-core`、`linapro-org-core`、`linapro-ai-core`的 provider 声明入口与 SPI import；使用`TenantFilter()`或租户/组织 scope 测试替身的官方插件按新类型更新。
- 无 HTTP API、DTO、路由、数据库 schema、前端页面或运行时用户可见文案变更。
- 数据权限影响：租户与组织 scope 过滤只迁移类型归属和注入路径，过滤语义、拒绝策略、数据库侧注入时机和降级行为保持不变；必须用既有 tenant/org/datascope 回归测试证明。
- 缓存一致性影响：不新增缓存；provider 可用性仍以既有插件 enabled snapshot 和`PluginStateService.IsProviderEnabled`为权威。provider manager 从包级单例改为宿主共享实例，任务记录必须包含 DI 来源检查。
