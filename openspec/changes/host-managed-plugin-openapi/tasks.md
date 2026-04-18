## 1. OpenSpec 文档

- [x] 1.1 创建 `host-managed-plugin-openapi` active change，并补齐 `proposal.md`、`design.md`、`tasks.md`
- [x] 1.2 为 `system-api-docs`、`plugin-runtime-loading`、`plugin-manifest-lifecycle` 编写本次变更的 delta specs

## 2. 源码插件路由归属采集

- [x] 2.1 改造 `pluginhost.RouteRegistrar`，引入宿主可观测的源码插件 `RouteGroup` facade，停止向源码插件暴露原生 `*ghttp.RouterGroup`
- [x] 2.2 为源码插件路由注册增加 `SourceRouteBinding` 采集模型，自动记录 `pluginID`、方法、路径、处理器与 DTO `g.Meta` 文档元数据
- [x] 2.3 保持源码插件中间件注册方式不变，确保插件仍然可以自行组合宿主发布的原始 `ghttp.HandlerFunc`
- [x] 2.4 迁移 `plugin-demo-source` 到新的源码插件路由注册 facade

## 3. 宿主自建 `/api.json`

- [x] 3.1 在 HTTP 启动链路中停用 GoFrame 内建 `/api.json` 自动产物，并绑定宿主自建 OpenAPI handler
- [x] 3.2 实现宿主 OpenAPI builder：从宿主真实路由表构建静态接口文档，并排除源码插件路由、动态插件固定分发入口与非业务路由
- [x] 3.3 实现源码插件路由按启用状态投影到宿主 OpenAPI 文档
- [x] 3.4 保持动态插件路由按启用状态继续投影到宿主 OpenAPI 文档

## 4. 测试与验证

- [x] 4.1 补充 Go 单测，覆盖源码插件路由归属采集、宿主 OpenAPI 构建、源码插件禁用后文档移除及动态插件投影回归
- [x] 4.2 新增或更新 Playwright 用例，验证源码插件启用/禁用后系统接口文档中对应接口的展示变化
- [x] 4.3 运行相关 Go 测试与 Playwright 用例，确认 `/api.json` 自建链路和插件文档投影没有引入回归

## Feedback

- [x] **FB-1**: 按职责拆分 `plugin` 与 `integration` 的 `Service` 接口，并通过接口嵌入组合总接口
