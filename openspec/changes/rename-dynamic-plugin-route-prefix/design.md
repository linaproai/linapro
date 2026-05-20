## Context

当前动态插件后端路由由宿主在 `/api/v1/extensions/{pluginId}/...` 下分发。这个路径同时承担了两类语义：`/api/v1` 表示宿主控制面 API 版本，`/extensions/{pluginId}` 表示动态插件数据面入口。两者耦合后，插件作者无法自然表达自己的 API 版本边界，例如 `/api/v2/...` 会变成 `/api/v1/extensions/{pluginId}/api/v2/...`，对调用端和文档都不清晰。

动态插件路由仍然必须由宿主管理入口、鉴权、权限、运行时 freshness、OpenAPI 投影和审计上下文，但宿主只需要识别“这是动态插件请求”和“属于哪个插件”。`{pluginId}` 后面的路径应当完全由插件声明和版本化。

## Goals / Non-Goals

**Goals:**

- 将动态插件 canonical 数据面入口定义为 `/x/{pluginId}/...`。
- 允许插件在自身路径中定义版本，例如 `/x/{pluginId}/api/v1/...` 或 `/x/{pluginId}/api/v2/...`。
- 不保留旧 `/api/v1/extensions/{pluginId}/...` 兼容入口，避免动态插件数据面继续受宿主 `/api/v1` 契约影响。
- 确保 `/x` 路由虽然移出 `/api/v1` 组，仍经过宿主统一 HTTP 治理链路和动态插件治理链路。
- 同步 OpenAPI、apidoc i18n、动态插件 demo、E2E 和审计脚本，使新路径成为唯一新文档和新示例口径。

**Non-Goals:**

- 不改变宿主控制面 API，例如插件安装、启用、禁用、资源列表仍在宿主 `/api/v1` 契约内。
- 不改变动态插件 manifest 中声明的内部路由语义、HTTP 方法、权限声明或 bridge 协议。
- 不改变 `/plugin-assets/<plugin-id>/<version>/...` 前端资产托管路径。
- 不引入可配置路由前缀；本次将 `/x` 作为稳定公开契约固化。
- 不重构动态插件缓存、运行时修订号、WASM 加载或 host service 协议。

## Decisions

1. `/x/{pluginId}` 是唯一动态插件数据面入口。

   备选方案是把前缀缩短为 `/api/v1/x/{pluginId}`。该方案虽然减少字符，但仍把插件数据面绑定到宿主 `/api/v1`，无法解决插件自有版本语义问题。因此选择根级 `/x/{pluginId}`。

2. 宿主只解析到 `pluginId`，剩余路径保持插件所有。

   宿主路由匹配只负责从 `/x/{pluginId}/...` 提取插件 ID 和内部路径，然后继续使用当前动态路由 manifest 匹配插件声明的 `route.path`。宿主不得把剩余路径解释为宿主 API 版本，也不得限制插件使用 `/api/v1`、`/api/v2`、`/graphql` 等路径段。

3. 复用现有动态插件 middleware，而不是创建绕过 `/api/v1` 的裸 handler。

   `/x` 路由需要挂载 `Response`、`CORS`、`RequestBodyLimit`、`Ctx` 等宿主通用中间件，再挂载 `PrepareDynamicRouteMiddleware` 和 `AuthenticateDynamicRouteMiddleware`。这样路径迁移不会绕过统一响应、跨域、请求大小限制、业务上下文、鉴权、权限和运行时 freshness。

4. OpenAPI 与资源列表只生成 `/x/...`。

   不保留旧路径可以避免调用端继续复制旧契约，并确保动态插件数据面语义只归属于 `/x`。

5. 旧入口直接移除，不使用 redirect 或双路由。

   动态插件路由包含 `GET`、`POST`、`PUT`、`DELETE` 等方法。重定向会引入方法保持、请求体重放和客户端差异；双路由会继续保留旧宿主版本语义。因此本变更直接迁移为破坏性路由契约调整。

## Risks / Trade-offs

- [Risk] 根级 `/x` 路径可能与未来宿主前端或反向代理路径冲突。
  Mitigation：将 `/x` 写入 OpenSpec 和路由测试，作为宿主保留命名空间；前端 SPA fallback 必须让 `/x` API 路由先匹配。

- [Risk] 路由移出 `/api/v1` 后遗漏通用中间件，导致响应格式、CORS、上下文或请求体限制不一致。
  Mitigation：路由绑定抽取可复用的 API middleware 装配，新增或更新 `internal/cmd` 测试覆盖 `/x` 分发链路。

- [Risk] 旧客户端依赖 `/api/v1/extensions` 后升级会失败。
  Mitigation：在提案、规范、接口文档、示例插件和测试中明确这是 breaking change，并统一迁移到 `/x/{pluginId}/...`。

- [Risk] apidoc 翻译键的动态插件路径识别依赖固定段位置。
  Mitigation：同步更新 apidoc 路径识别函数，识别 `/x/{pluginId}/...` 并生成相同插件维度翻译 key。

## Migration Plan

1. 增加动态插件 public prefix 常量 `/x`，并让 runtime 只匹配 `/x/{pluginId}/...`。
2. 将 OpenAPI 投影、插件资源列表和 demo 前端调用切换为 `/x/{pluginId}/...`。
3. 同步更新 apidoc i18n JSON 和 API DTO 文档示例，移除新文档中的 `/api/v1/extensions` 口径。
4. 更新单元测试、动态插件 E2E 和性能审计脚本，验证 `/x` 路径可用且旧路径不再作为动态插件入口。
5. 运行变更包 Go 测试、动态插件相关测试、E2E 或等价 smoke，并执行 `openspec validate`。

回滚策略：如 `/x` 发布后发现代理或部署冲突，可恢复动态插件路由绑定和 public path 生成到旧 `/api/v1/extensions` 前缀；不通过本变更同时维护两个入口。

## Open Questions

- 暂无。用户已明确不需要保留旧 `/api/v1/extensions/{pluginId}/...` 兼容入口。
