# 任务清单：拆分插件消费契约与 Provider SPI

## 1. `tenantspi`拆分

- [x] 1.1 创建`pkg/plugin/capability/tenantcap/tenantspi`，迁移`Provider`、`ProviderFactory`、`ProviderEnv`、`ProviderRuntime`、`Resolver`、`RequestResolver`、`ResolverChain`、`ScopeService`、`UserMembershipProvider`、`RuntimeService`和`PluginTableFilterService`相关实现，父级`tenantcap`仅保留普通消费契约、DTO、错误码和常量
- [x] 1.2 更新宿主 tenant scope、plugin table filter、request resolver、`pluginhost.Services.TenantFilter()`和相关测试替身 import，确保数据权限过滤语义、数据库侧注入时机和拒绝策略不变
- [x] 1.3 运行覆盖 tenant 能力与 scope 的 Go 测试，记录数据权限影响判断和无运行时行为变更结论

## 2. `orgspi`拆分

- [x] 2.1 创建`pkg/plugin/capability/orgcap/orgspi`，迁移`Provider`、`ProviderFactory`、`ProviderEnv`、`ProviderRuntime`、`ScopeService`、`RuntimeService`和组织 host implementation，父级`orgcap`仅保留普通消费契约、DTO、错误码和常量
- [x] 2.2 将`orgspi.ProviderEnv.TenantFilter`切换为`tenantspi.PluginTableFilterService`，更新宿主 org scope、datascope、user、notify、session 等调用点和测试替身 import
- [x] 2.3 运行覆盖 org 能力、datascope 和相关列表过滤路径的 Go 测试，记录数据权限无语义变化结论

## 3. 普通能力包去`ghttp`/`gdb`

- [x] 3.1 将`routecap.Service.DynamicRouteMetadata(*ghttp.Request)`改为`DynamicRouteMetadata(context.Context)`，宿主`capabilityhost`适配器内部从`context.Context`恢复请求元数据，更新 guest 实现和调用方
- [x] 3.2 删除`apidoccap.BuildOperationKeyFromHandler`或迁入宿主侧私有 helper，唯一调用方改用不依赖`ghttp`的 operation key 派生路径，并验证结果等价
- [x] 3.3 静态检索确认非`*spi`的`capability/**`生产代码无`gdb`和`ghttp` import

## 4. Provider 注册机制收敛

- [x] 4.1 在`pluginhost.Declarations`新增强类型 provider 声明分组，提供`ProvideTenant`、`ProvideOrg`、`ProvideAIText`等注册方法，`SourcePluginDefinition`可读取声明出的 factory
- [x] 4.2 宿主插件能力装配层创建并持有 tenant/org/AI text provider manager 实例，通过构造函数显式注入对应 service，并从 source plugin definition 注册 provider factory
- [x] 4.3 迁移`linapro-tenant-core`、`linapro-org-core`、`linapro-ai-core`的`backend/plugin.go`到`pluginhost.Declarations`provider 声明入口，修改前检查各插件根目录`AGENTS.md`
- [x] 4.4 删除`tenantcap`、`orgcap`、`aitext`中的包级`defaultManager`和旧`Provide()`入口，静态检索确认旧入口无生产和测试残留；任务记录补充 DI 来源检查：owner、创建位置、传递路径、共享实例策略

## 5. 治理测试与文档同步

- [x] 5.1 扩展`pkg/plugin/plugin_boundary_test.go`：非`*spi`的`capability/**`生产代码不得 import `gdb/ghttp`，`pluginbridge/**`生产代码不得 import 任意`*spi`子包
- [x] 5.2 正反向验证治理测试：当前代码通过；临时构造违规 import 确认测试能捕获并输出违规文件与 import 路径，验证后删除临时文件
- [x] 5.3 更新`pkg/plugin/README.md`与`README.zh-CN.md`，说明普通消费契约、源码插件 provider SPI、动态插件 guest SDK 的归属边界与新增能力判定标准

## 6. 验证与收尾

- [x] 6.1 宿主执行`go build ./...`、`go vet ./...`、`go test ./... -count=1`；至少单独运行`go test ./pkg/plugin/... -count=1`、tenant/org/datascope 相关测试和`internal/cmd`启动绑定测试
- [x] 6.2 对涉及的源码插件和动态插件样例执行普通 Go 构建；对动态插件样例执行项目正式`wasip1`构建路径，确认 guest 编译闭包不含`*spi`
- [x] 6.3 静态检索确认`tenantcap.Provide`、`orgcap.Provide`、`aitext.Provide`、`defaultManager`、旧 SPI 类型路径和普通`capability`父包`gdb/ghttp` import 无残留
- [x] 6.4 记录影响分析：无 HTTP API/DTO/路由变更、无 SQL/数据库变更、无前端变更、无运行时 i18n 文案变更、无缓存一致性语义变更、数据权限仅 import 与类型归属迁移、无新增运行期依赖语义但 provider manager owner 从包级单例迁移为宿主持有共享实例
- [x] 6.5 运行`openspec validate split-plugin-capability-consumer-spi --strict`通过，调用`lina-review`完成变更审查

## 执行记录

- 数据权限影响：本变更只迁移`tenant`、`org`相关 scope、provider 和 plugin table filter 的类型归属与 import 路径；`Apply`、`ApplyUserTenantScope`、`ApplyUserDeptScope`、`BatchGetScoped`和`ListPageScoped`等数据库侧过滤注入时机、拒绝策略和降级语义不变。
- DI 来源检查：provider manager 的 owner 为 HTTP 启动期插件能力装配；创建位置为`apps/lina-core/internal/cmd/internal/httpstartup/http_runtime.go`的`newHTTPRuntime`；传递路径为`RegisterSourcePluginProviderFactories(tenantProviderManager, orgProviderManager, aiTextProviderManager)`注册 source plugin definition 中的 factory，再将同一批 manager 注入`tenantspi.New`、`orgspi.New`、`aitext.New`；共享策略为每个 capability family 在一次宿主启动中复用一个启动期共享实例，不再使用能力包级`defaultManager`。
- 插件本地规范检查：修改`apps/lina-plugins/<plugin-id>/`文件前执行`find apps/lina-plugins -maxdepth 2 -name AGENTS.md -print`，未发现插件根目录本地`AGENTS.md`。
- 无影响判断：无 HTTP API/DTO/路由契约变更；无 SQL、DAO、DO、Entity 或数据库迁移变更；无前端页面和用户可观察 UI 变更；无运行时用户可见文案或`i18n`资源变更；无缓存权威源、失效路径、跨实例同步或一致性语义变更；无开发工具入口变更。
- `lina-review`修复记录：审查发现`role`服务仍通过`pluginSvc`临时构造空`orgspi.Manager`派生组织能力状态，导致部门数据范围可用性判断没有复用启动期共享 provider manager；已改为`role.New`显式接收启动期共享`orgCapSvc`，并同步启动装配与测试构造。
- 验证证据：`go test ./pkg/plugin/... -count=1`、宿主 tenant/org/datascope 和启动绑定相关测试、`go build ./...`、`go vet ./...`、`go test ./... -count=1`均通过；涉及的源码插件和动态插件样例`GOWORK=off go test ./... -count=1`通过；`make -C apps/lina-plugins wasm p=linapro-demo-dynamic`生成`temp/output/linapro-demo-dynamic.wasm`；治理测试正反向验证完成，临时违规文件已删除；`openspec validate split-plugin-capability-consumer-spi --strict`通过。
