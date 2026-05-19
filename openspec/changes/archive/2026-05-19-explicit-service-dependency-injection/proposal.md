## Why

当前宿主 Controller、Middleware、插件服务适配器和部分插件后端在运行期各自调用 `service.New()` 构造独立服务图，导致服务依赖关系不透明，并在权限、会话、插件状态、运行时配置、i18n 和缓存协调等高风险路径上存在实例不一致隐患。需要以显式依赖注入收敛后端依赖管理，并把该约束纳入项目规范和 `lina-review` 审查标准，避免后续迭代继续引入隐式服务构造。

## What Changes

- **BREAKING**：后端宿主和源码插件的 Controller、Middleware、Service、插件宿主服务适配器和 WASM host service 不再在业务构造函数或请求/回调路径中隐式创建关键依赖服务。
- **BREAKING**：`Service.New()`、`Controller.NewV1()` 等构造入口需要改为逐项接收接口型显式依赖；无状态服务也由调用方统一持有共享实例。禁止通过 `Dependencies`、`Deps`、`Options` 等聚合结构体整体传递多个接口对象，以便依赖变化通过编译错误暴露所有未同步调用点。
- 不新增通用 DI 容器、不新增新的宿主私有组装层；继续复用现有启动编排和路由注册位置，通过显式传参完成依赖收敛。
- 高风险缓存一致性路径必须复用同一套运行期服务实例，包括认证中间件、权限服务、插件管理、插件运行时缓存、运行时配置、i18n、session hot state、source plugin registrar 和 WASM host service。
- 为源码插件和动态插件 host service 提供宿主发布的依赖入口，插件控制器和服务通过注册回调上下文接收依赖，而不是自行构造宿主服务适配器。
- 更新 `AGENTS.md` 项目规范和 `.agents/skills/lina-review/SKILL.md` 审查标准，要求审查所有后端变更的依赖显式注入、共享实例和缓存一致性影响。
- 增加静态扫描、单元测试或等价治理验证，防止非测试/非组装边界继续直接调用关键服务 `New()` 创建孤立实例。

## Capabilities

### New Capabilities
- `service-dependency-injection-governance`: 定义宿主、源码插件、动态插件 host service 和审查流程的显式依赖注入、共享实例和隐式构造禁止规则。

### Modified Capabilities
- `backend-conformance`: 将现有控制器和服务层构造约束升级为显式依赖注入规范，禁止业务组件内部隐式构造关键依赖。
- `distributed-cache-coordination`: 补充缓存敏感服务必须复用运行期同一服务实例的要求，避免集群模式下本地状态、修订号或派生缓存分裂。
- `plugin-http-slot-extension`: 补充源码插件 HTTP、全局中间件和 Cron 注册回调应通过宿主发布依赖完成控制器和服务构造。
- `plugin-host-service-extension`: 补充插件 host service 适配器必须由宿主运行期统一构造并复用，不得在调用路径内临时创建孤立宿主服务图。

## Impact

- 影响 `apps/lina-core/internal/controller/**`、`apps/lina-core/internal/service/**`、`apps/lina-core/internal/cmd/**` 和 `apps/lina-core/pkg/pluginservice/**` 的构造方式。
- 影响 `apps/lina-core/internal/service/middleware`、`plugin`、`auth`、`role`、`session`、`i18n`、`config`、`notify`、`cachecoord`、`pluginruntimecache`、WASM host service 等缓存或运行时状态敏感组件。
- 影响 `apps/lina-plugins/**/backend/plugin.go`、源码插件 Controller 和 Service 的依赖传递方式。
- 影响 `.agents/skills/lina-review/SKILL.md`、`AGENTS.md` 和 OpenSpec 规范文档。
- 不涉及数据库 schema、REST API 语义、前端 UI、运行时语言包或 SQL seed 数据变更；若实现过程中发现用户可见错误文案或 apidoc 变化，应按 i18n 规范同步维护。
