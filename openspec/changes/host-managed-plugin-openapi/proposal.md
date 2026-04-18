## Why

当前“系统接口”页面依赖 GoFrame 内建的 `/api.json` 自动产物。这个链路对宿主静态接口没有问题，但对插件路由存在两个长期缺口：

- 源码插件路由是通过注册代码直接写入宿主路由表的，GoFrame 自动文档并不知道这些路由属于哪个插件，也无法按插件启用状态决定是否展示
- 如果要求开发者再去 `plugin.yaml` 中重复声明一份路由清单，会增加维护成本，并且极易与实际注册代码不一致

现在需要把 `/api.json` 的生成权收回到宿主服务，并让源码插件路由在“注册时自动采集归属和文档元数据”，从而在不限制源码插件路由地址、不牺牲源码插件中间件灵活性的前提下，实现按插件启用状态投影接口文档。

## What Changes

- 停止直接使用 GoFrame 内建 `/api.json` 自动产物，改为由宿主服务自建 OpenAPI 文档输出
- 新增宿主管理的系统接口文档构建链路：宿主静态接口通过宿主扫描路由并调用 `goai.Add` 生成，插件接口通过宿主掌握的插件路由绑定记录投影
- 改造源码插件路由注册接口，不再向插件暴露原生 `*ghttp.RouterGroup`，而是暴露宿主可观测的 `RouteGroup` facade
- 源码插件路由仍然只在代码中定义；宿主在注册时自动采集 `pluginID`、方法、路径、处理器及 DTO `g.Meta` 元数据，不要求在 `plugin.yaml` 中重复声明路由
- 源码插件路由地址不做前缀约束，仍然允许注册任意合法地址；宿主通过显式归属记录而不是路径前缀推断路由归属
- 源码插件中间件注册和组合方式保持插件自维护；宿主不把源码插件中间件收敛为动态插件那种受控中间件描述符模型
- 系统接口文档仅展示当前启用的插件路由：源码插件按宿主采集的路由绑定投影，动态插件继续按当前已启用 release 的路由合同投影
- 补充 Go 单测与 E2E，覆盖源码插件路由归属采集、宿主 `/api.json` 生成、插件启停后的文档展示变化

## Capabilities

### Modified Capabilities

- `system-api-docs`: 系统接口文档改为由宿主自建 `/api.json`，并按插件启用状态投影源码插件与动态插件路由
- `plugin-runtime-loading`: 源码插件路由改为在注册时由宿主自动采集路由归属与文档元数据，同时保持任意路由地址和插件自维护中间件
- `plugin-manifest-lifecycle`: `plugin.yaml` 继续保持精简，不新增源码插件路由声明字段

## Impact

- **后端 HTTP 启动链路**: `apps/lina-core/internal/cmd/cmd_http.go` 需要关闭 GoFrame 内建 OpenAPI 输出并绑定宿主自建 `/api.json` handler
- **插件宿主接口**: `apps/lina-core/pkg/pluginhost/` 需要引入源码插件路由采集能力和新的 `RouteGroup` facade
- **插件服务**: `apps/lina-core/internal/service/plugin/` 需要暴露源码插件路由绑定查询和插件路由 OpenAPI 投影能力
- **系统接口文档**: 新增宿主 OpenAPI 构建服务，用于合并宿主静态接口、源码插件接口和动态插件接口
- **源码插件示例**: `apps/lina-plugins/plugin-demo-source/backend/plugin.go` 需要迁移到新的宿主路由注册 facade
- **测试**: 新增/更新 Go 单测与 Playwright 用例，确保源码插件路由归属、文档生成和启停联动行为可回归验证
