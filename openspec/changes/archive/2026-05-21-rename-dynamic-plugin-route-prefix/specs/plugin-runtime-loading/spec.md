## ADDED Requirements

### Requirement: 动态插件数据面路由必须使用独立宿主命名空间

系统 SHALL 使用 `/x/{pluginId}/...` 作为动态插件数据面路由的 canonical 公开入口。宿主只将 `/x` 识别为动态插件分发命名空间，并只从路径中解析 `pluginId`；`{pluginId}` 之后的路径 SHALL 完全归插件路由声明所有。宿主 MUST NOT 将动态插件数据面路由固定在宿主控制面 `/api/v1` 前缀下，也不得限制插件在自身路径中声明 `/api/v1`、`/api/v2`、`/graphql` 或其他插件自有路径结构。

#### Scenario: 插件声明自己的 API 版本

- **WHEN** 动态插件 `plugin-a` 声明内部路由 `/api/v2/items`
- **THEN** 宿主以 `/x/plugin-a/api/v2/items` 作为 canonical 公开路径
- **AND** 宿主不得生成 `/api/v1/extensions/plugin-a/api/v2/items` 作为 canonical 公开路径

#### Scenario: 插件声明非 REST 版本路径

- **WHEN** 动态插件 `plugin-a` 声明内部路由 `/graphql`
- **THEN** 宿主以 `/x/plugin-a/graphql` 分发该请求
- **AND** 宿主不得要求插件路径包含宿主 API 版本段

### Requirement: 动态插件旧扩展路由不得继续作为分发入口

系统 MUST NOT 继续接受旧 `/api/v1/extensions/{pluginId}/...` 作为动态插件数据面分发入口。OpenAPI 投影、插件资源列表、示例插件前端和新文档 MUST 使用 `/x/{pluginId}/...` 作为公开路径。

#### Scenario: 旧扩展路径不再分发动态插件请求

- **WHEN** 客户端请求 `/api/v1/extensions/plugin-a/backend-summary`
- **THEN** 宿主不得按动态插件 `plugin-a` 的 `/backend-summary` 内部路由执行请求

#### Scenario: 新文档只展示新路径

- **WHEN** 宿主生成动态插件 OpenAPI 文档或插件资源列表
- **THEN** 动态插件公开路径以 `/x/{pluginId}/...` 开头
- **AND** 新生成内容不得把 `/api/v1/extensions/{pluginId}/...` 作为动态插件公开路径

### Requirement: 动态插件根级路由必须保留宿主 HTTP 治理链路

系统 SHALL 在根级 `/x` 动态插件路由上复用宿主统一 HTTP 治理链路。请求在进入动态插件 bridge 执行前 MUST 经过响应包装、CORS、请求体限制、业务上下文初始化、运行时 freshness 检查、动态插件路由准备、登录鉴权和权限校验。路由前缀迁移 MUST NOT 绕过插件启用状态、运行时修订号、数据权限上下文或审计元数据构建。

#### Scenario: 未认证用户访问需要登录的插件路由

- **WHEN** 未认证用户请求 `/x/plugin-a/private-summary`
- **AND** 动态插件路由声明需要登录访问
- **THEN** 宿主拒绝该请求
- **AND** 拒绝结果使用宿主统一响应格式

#### Scenario: 插件禁用后新前缀不可继续暴露能力

- **WHEN** 动态插件 `plugin-a` 被禁用
- **THEN** 后续访问 `/x/plugin-a/backend-summary` 不得继续执行该插件 bridge 路由

#### Scenario: 动态路由元数据使用实际命中路径

- **WHEN** 请求通过 `/x/plugin-a/backend-summary` 命中动态插件路由
- **THEN** 传递给动态插件 bridge 和宿主中间件的 public path 元数据反映实际命中的 `/x/plugin-a/backend-summary`
