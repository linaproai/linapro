## Why

当前动态插件公开路由固定在 `/api/v1/extensions/{pluginId}/...` 下，使插件业务 API 被宿主控制面 API 版本 `/api/v1` 强行约束。动态插件应拥有自己的数据面路径和版本语义，例如 `/api/v2/...`、`/graphql` 或其他插件自定义路径，而宿主只应固定用于识别动态插件的入口命名空间。

本次变更将动态插件公开路由的 canonical 前缀调整为 `/x/{pluginId}/...`，解除宿主 API 版本与插件 API 版本之间的耦合。

## What Changes

- **BREAKING**：将动态插件数据面路由从 `/api/v1/extensions/{pluginId}/...` 调整为 `/x/{pluginId}/...`，不保留旧路径兼容入口。
- 宿主只保留 `/x/{pluginId}` 作为动态插件识别和治理入口，`{pluginId}` 之后的路径完全归插件所有。
- 动态插件可在自身路径内声明版本，例如 `/x/{pluginId}/api/v1/...`、`/x/{pluginId}/api/v2/...`，宿主不得把插件版本绑定到宿主 `/api/v1`。
- OpenAPI 投影、插件资源列表、动态插件 demo 前端和新测试以 `/x/...` 作为 canonical public path。
- 动态插件路由移出 `/api/v1` 后，仍必须复用宿主统一 HTTP 治理链路，包括响应包装、CORS、请求体限制、上下文初始化、动态插件鉴权和权限校验。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-runtime-loading`：动态插件运行时公开路由前缀从宿主版本化 API 命名空间迁移到独立 `/x/{pluginId}` 数据面命名空间，并定义新路径和治理链路要求。

## Impact

- 后端：`apps/lina-core/internal/cmd` 路由绑定、动态插件 runtime 路由匹配、OpenAPI 投影、apidoc 路径识别、插件资源列表 public path 生成。
- 插件：`apps/lina-plugins/linapro-demo-dynamic` 示例前端、后端 API 文档注释、manifest apidoc i18n 资源和插件自有 E2E。
- 工具与测试：动态插件 runtime 单元测试、控制器测试、pluginbridge 编解码测试、性能审计脚本中的动态插件路径识别。
- API 契约：动态插件公开路径变更为 `/x/{pluginId}/...`；旧 `/api/v1/extensions/{pluginId}/...` 不再可用。
- i18n：需要同步更新 apidoc i18n JSON 中描述动态插件公开路径的文案，避免继续提示 `/api/v1/extensions`。
- 缓存一致性：本次不新增缓存；动态插件启用、禁用、升级、运行时修订号和派生缓存失效机制保持不变，路由前缀迁移不得绕过既有插件 runtime freshness 检查。
