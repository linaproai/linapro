## Context

当前宿主通过 GoFrame 默认能力在启动后生成 `/api.json`。这意味着：

- 文档数据源来自 GoFrame 扫描到的最终路由表
- 源码插件只要把路由注册到宿主路由表中，就会自动进入文档
- 宿主只能在 GoFrame 自动产物之上做“增强”，而不能精确控制某一类插件路由是否进入文档

与此同时，源码插件的真实设计边界与动态插件不同：

- 源码插件允许注册任意合法路由地址，不要求固定前缀
- 源码插件通过注册代码自行组合和维护中间件
- 源码插件不应在 `plugin.yaml` 中重复声明路由，否则会形成双份事实来源

因此，本次设计的核心不是“统一源码插件和动态插件的路由形态”，而是“统一宿主对接口文档生成权和路由归属的掌控方式”。

## Goals / Non-Goals

**Goals:**

- 由宿主服务自建 `/api.json`，不再直接使用 GoFrame 默认 OpenAPI 输出
- 让宿主能够在不依赖 URL 前缀的前提下，明确知道每条源码插件路由属于哪个插件
- 保持源码插件路由只在代码中定义，不引入 `plugin.yaml` 路由重复声明
- 保持源码插件中间件链仍由插件自己维护，宿主不解释其中的业务语义
- 让系统接口文档按当前插件启用状态展示源码插件和动态插件路由
- 为源码插件路由采集和宿主 OpenAPI 构建补齐可回归测试

**Non-Goals:**

- 不要求源码插件把后端路由统一迁移到固定前缀
- 不把源码插件中间件收敛为动态插件那种宿主受控执行链
- 不在 `plugin.yaml` 中增加 `routes` 字段或其他等价路由清单
- 不尝试从源码插件中间件链反推 `public/login/permission` 等文档语义
- 不为 `func(*ghttp.Request)` 这类原始 handler 自动补全文档；没有 DTO 元数据的原始 handler 暂不自动投影到 OpenAPI

## Decisions

### 1. 宿主接管 `/api.json` 输出，不再直接使用 GoFrame 默认 OpenAPI 产物

宿主启动时保存当前配置中的 `openapiPath` 作为系统接口文档地址，然后显式关闭 GoFrame 默认 `OpenApiPath`，改为绑定宿主自己的 handler 输出 OpenAPI 文档。

宿主自建文档的来源分为三部分：

- 宿主静态接口：从当前宿主路由表中扫描并调用 `goai.Add`
- 源码插件接口：从宿主在注册时采集的源码插件路由绑定记录投影
- 动态插件接口：继续从当前已启用动态插件的 route contracts 投影

这样宿主获得完整的“是否进入文档”的决策权，而不再依赖 GoFrame 先自动生成、宿主后置修改的模式。

### 2. 源码插件路由保持代码定义，宿主在注册时自动采集归属

源码插件路由的唯一事实来源是注册代码及其 DTO `g.Meta`，不在 `plugin.yaml` 中重复维护。

为此，源码插件注册接口从：

- `register func(group *ghttp.RouterGroup)`

调整为：

- `register func(group RouteGroup)`

`RouteGroup` 仍提供接近 GoFrame 的注册体验，例如：

- `Group`
- `Middleware`
- `Bind`
- `BindMethod`
- `GET/POST/PUT/DELETE`

但所有绑定动作都先经过宿主包装层。宿主在这些绑定动作里自动记录：

- `pluginID`
- 最终 `method`
- 最终 `path`
- 处理器引用
- DTO `g.Meta` 中的标签、摘要、描述、权限等文档元数据

这样源码插件仍然是代码定义路由，但宿主已经能稳定知道路由归属。

### 3. 源码插件中间件链保持插件自维护，宿主不解释其业务语义

源码插件和动态插件的中间件边界不同，本次保持这种差异：

- 动态插件中间件链仍由宿主统一控制
- 源码插件仍可通过宿主发布的原始 `ghttp.HandlerFunc` 自由组合中间件

宿主只把源码插件中间件调用原样透传到底层 GoFrame 路由组，不把它收敛为“带标识的中间件描述符”，也不尝试根据中间件链推导访问语义。

这意味着源码插件的文档语义必须来自处理器 DTO `g.Meta` 本身，而不是来自中间件推断。

### 4. 只有标准 DTO 路由会自动进入源码插件 OpenAPI 投影

宿主自动采集并投影到 OpenAPI 的源码插件路由，限定为 GoFrame 标准业务处理器签名：

```go
func(ctx context.Context, req *Req) (res *Res, err error)
```

并要求 `Req` 通过 `g.Meta` 提供 `path`、`method` 及其他文档元数据。

对 `func(*ghttp.Request)` 这类原始 handler：

- 宿主允许注册
- 宿主记录其路由归属，避免被误认为宿主静态接口
- 但默认不自动投影到 OpenAPI

这样可以避免再次引入“为文档额外维护一份元数据”的双轨模型。

### 5. 宿主静态接口文档仍从实际路由表生成，但会排除插件路由

宿主静态接口的 OpenAPI 生成不再依赖 GoFrame 默认 `server.openapi`，但仍以“真实已注册路由”为准，避免手工维护控制器清单。

具体做法：

- 读取 `server.GetRoutes()`
- 过滤中间件、hook、静态兜底、上传静态文件、宿主自建 `/api.json` handler 和动态插件固定分发入口
- 根据源码插件路由绑定记录排除所有插件路由
- 对剩余可文档化业务 handler 调用 `goai.Add`

这样宿主静态接口的文档仍与真实路由表一致，同时不会把插件路由混入宿主静态接口集合。

### 6. 插件路由文档按当前启用状态投影

插件路由文档投影分成两类：

- **源码插件**：读取宿主采集到的 `SourceRouteBinding`，仅为当前启用的源码插件投影
- **动态插件**：继续读取当前已启用动态插件的 `manifest.Routes` 进行投影

源码插件禁用后，真实路由仍然可能存在于宿主路由表中，但：

- 请求访问会被插件状态守卫拦住
- 宿主自建 `/api.json` 不再投影这些路由

从而保证“系统接口页面是否展示”与“插件当前治理状态”一致。

### 7. 路由冲突继续以宿主真实注册失败为准，不新增前缀型约束

由于源码插件允许使用任意合法路由地址，本次不新增路径前缀约束。

冲突治理策略保持简单：

- 源码插件绑定的真实路由仍注册到宿主路由树
- 若与宿主静态接口、其他源码插件路由或动态插件固定入口产生重复，沿用 GoFrame 的重复路由检测直接失败
- 宿主的源码插件路由绑定记录同样按 `method + path` 保持唯一，避免文档投影出现多份归属

这样既保留源码插件路径灵活性，也不引入额外的路由命名制度。

## Architecture

```text
source plugin code
  -> pluginhost.RouteRegistrar
     -> host RouteGroup facade
        -> plugin-owned middleware registration
        -> host-captured route binding records
        -> actual GoFrame route binding

host /api.json request
  -> host OpenAPI builder
     -> scan host static routes from server.GetRoutes()
     -> exclude source plugin routes using binding registry
     -> project enabled source plugin route bindings
     -> project enabled dynamic plugin route contracts
     -> return merged OpenAPI document
```

## Risks / Trade-offs

- **[源码插件注册接口有兼容性变更]**：源码插件不再直接拿 `*ghttp.RouterGroup`，需要迁移到宿主 facade
- **[原始 handler 不自动进文档]**：这是为了避免在 `plugin.yaml` 或其他位置重复维护第二份文档元数据；若后续确有需求，可在注册代码处增加显式元数据能力
- **[源码插件中间件语义不受宿主解释]**：宿主无法从中间件链自动推断公开/鉴权语义，因此标准 DTO `g.Meta` 质量会更重要
- **[静态接口文档从真实路由扫描构建]**：需要额外过滤非业务路由与插件路由，否则会把内部兜底或固定分发入口误加到系统接口页面

## Migration Plan

1. 在宿主 HTTP 启动链路中切换到宿主自建 `/api.json`
2. 在 `pluginhost` 中引入新的源码插件 `RouteGroup` facade 和路由绑定采集结构
3. 迁移现有源码插件示例到新的注册接口
4. 为插件服务增加源码插件路由绑定读取与插件路由 OpenAPI 投影能力
5. 补充 Go 单测与 Playwright 用例，验证系统接口页面在源码插件启停前后展示变化正确

## Open Questions

- 当前无阻塞实现的开放问题；若后续需要为原始 `ghttp.Request` handler 提供文档化能力，应在注册代码处补充显式 route metadata，而不是把路由清单迁移到 `plugin.yaml`
