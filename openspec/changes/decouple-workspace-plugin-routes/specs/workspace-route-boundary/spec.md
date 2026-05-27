## ADDED Requirements

### Requirement: 默认管理工作台入口必须可配置且不默认独占根路径

系统 SHALL 提供默认管理工作台入口路径配置，并将管理工作台 SPA 的静态资源服务与 fallback 限定在该入口路径及其子路径内。默认管理工作台入口路径默认值为 `/admin`；部署方 MAY 将入口路径配置为根路径 `/`，用于管理后台拥有独立域名或根路径由工作台独占的部署场景。工作台入口路径 SHALL 以启动期配置为权威，后端 fallback、前端 router base、构建或启动配置、开发代理和测试入口必须使用同一值；系统不得把该配置作为运行期热修改项处理。

#### Scenario: 管理工作台通过配置入口访问
- **WHEN** 工作台入口路径配置为 `/admin`
- **THEN** 浏览器访问 `/admin` 或 `/admin/*` 时系统服务默认管理工作台 SPA
- **AND** 管理工作台刷新子路由时仍返回工作台 `index.html`

#### Scenario: 根路径不再被工作台 fallback 吞掉
- **WHEN** 工作台入口路径配置为 `/admin`
- **AND** 没有任何宿主或源码插件路由匹配 `/`
- **THEN** 请求 `/` 不得返回默认管理工作台 SPA
- **AND** 系统返回未匹配路由结果

#### Scenario: 独立管理后台域名使用根路径工作台入口
- **WHEN** 部署方将工作台入口路径配置为 `/`
- **THEN** 浏览器访问 `/` 或工作台客户端路由时系统服务默认管理工作台 SPA
- **AND** 宿主 API `/api/**`、统一插件 API `/x/**`、插件资产 `/x-assets/**` 和已注册的更具体宿主或源码插件路由不得被工作台 SPA fallback 吞掉
- **AND** 若源码插件也尝试注册与根路径工作台入口冲突的 `/` 或 `/*` 路由，系统必须在路由注册或启动阶段显式失败

#### Scenario: 工作台入口路径不是运行期热修改项
- **WHEN** 部署方修改工作台入口路径配置
- **THEN** 修改后的入口路径必须在对应服务重新启动并重新加载前端资源后生效
- **AND** 系统不得要求已创建的前端 router 在不刷新或不重启的情况下热切换 basePath
- **AND** public frontend 运行时配置不得成为工作台 router base 的唯一权威来源

### Requirement: 宿主保留命名空间和统一插件 API 命名空间必须优先于源码插件公开路由

系统 SHALL 保护宿主已注册控制面 API 路由、统一插件 API 命名空间、插件资产入口和管理工作台入口路径。源码插件通过 GoFrame 注册 HTTP 路由时不得覆盖已注册宿主路由、其他插件已注册路由、动态插件 API 分发、插件资产入口或管理工作台入口；冲突必须在路由注册或启动阶段显式失败。源码插件和动态插件 API 均 SHALL 使用 `/x/{plugin-id}/api/v1/...` 作为公开 API 路径；`/x` 表示统一插件 API 命名空间，不再表示动态插件专属数据面。源码插件 API 不得继续挂载在宿主控制面 `/api/v1` 命名空间下；源码插件不得在 `/x` 下注册非 API 路由。源码插件公开页面、门户、静态资源或自管 fallback MAY 注册 `/`、`/portal/*`、`/assets/*` 或其他非保留 HTTP 路由。`/x-assets` 由宿主保留，用于源码插件和动态插件 `plugin.yaml public_assets` 声明式 public assets 的统一托管；本变更不得保留 `/plugin-assets` 兼容入口。

#### Scenario: 保留 API 命名空间优先
- **WHEN** 源码插件尝试注册与某个已注册 `/api/v1` 宿主控制面 API 具有相同 HTTP 方法和路径的路由
- **THEN** 系统在启动或路由注册阶段失败
- **AND** 不允许源码插件覆盖宿主控制面 API

#### Scenario: 源码插件 API 使用统一插件 API 前缀
- **WHEN** 源码插件 ID 为 `my-plugin`
- **AND** 源码插件通过 registrar 的 API 前缀注册资源路由 `/reports`
- **THEN** 系统完整 API 路径为 `/x/my-plugin/api/v1/reports`
- **AND** 同一源码插件的 API 路由不得挂载到 `/api/v1` 宿主控制面命名空间
- **AND** 该 API 前缀不得阻止同一源码插件注册公开页面、门户、静态资源或其他非保留 HTTP 路由

#### Scenario: 统一插件 API 命名空间优先
- **WHEN** 源码插件尝试注册其他插件 ID 子树下的 `/x/{plugin-id}/api/v1/...` 路由
- **THEN** 系统在启动或路由注册阶段失败
- **AND** 不允许源码插件覆盖其他源码插件 API 或动态插件 API 分发入口

#### Scenario: `/x` 下非 API 路由被拒绝
- **WHEN** 源码插件尝试注册 `/x/my-plugin/assets/logo.png`、`/x/my-plugin/portal` 或 `/x/my-plugin/health` 等非 `/x/{own-plugin-id}/api/v1/...` 路由
- **THEN** 系统在启动或路由注册阶段失败
- **AND** 源码插件公开页面、门户、静态资源或自管 fallback 必须改用其他非保留路径或 `/x-assets`

#### Scenario: `/x` 命名空间保留给插件 API
- **WHEN** 宿主工作台入口或插件资产入口配置尝试占用 `/x`
- **THEN** 系统在配置加载或启动校验阶段失败
- **AND** `/x/{plugin-id}/api/v1/...` 保留作为统一插件 API 命名空间

#### Scenario: 管理工作台入口命名空间优先
- **WHEN** 工作台入口路径配置为 `/admin`
- **AND** 源码插件尝试注册与 `/admin` 或 `/admin/*` 冲突的路由
- **THEN** 系统在启动或路由注册阶段失败
- **AND** 不允许源码插件覆盖默认管理工作台入口

### Requirement: 源码插件 HTTP 路由必须由插件代码闭环维护

系统 SHALL 允许源码插件通过现有 HTTP registrar 注册非保留路径的 HTTP 路由。主框架不得要求源码插件在 `plugin.yaml` 中声明公开页面路由、门户路由、工作台 API 路由或路由分组语义。插件 MAY 声明可由宿主托管的 public asset 目录，但该声明不得替代或推断源码插件的 HTTP 路由注册行为。源码插件在 `/x/{plugin-id}/api/v1/...` 下注册且符合文档条件的 API 路由 MAY 参与 OpenAPI 投影；公开页面、门户、静态资源、自管 fallback 和 `/x-assets` 文件访问不得自动投影为 OpenAPI。

#### Scenario: 源码插件注册根路径
- **WHEN** 工作台入口路径配置为 `/admin`
- **AND** 源码插件通过代码注册 `/` 路由
- **THEN** 系统允许该路由注册并由插件 handler 处理请求
- **AND** 主框架不要求 `plugin.yaml` 为该路由提供声明

#### Scenario: 源码插件自行组织静态资源路由
- **WHEN** 源码插件通过代码注册 `/assets/*` 或其他非保留静态资源路由
- **THEN** 系统允许插件 handler 自行决定返回内容、缓存策略和资源路径
- **AND** 主框架不从该 HTTP 路由推断资源目录或静态文件清单
- **AND** 这不影响同一插件额外声明 public assets 并由宿主通过 `/x-assets/{plugin-id}/{version}/...` 托管

#### Scenario: registrar 暴露统一插件 API 前缀
- **WHEN** 源码插件通过 HTTP registrar 注册插件 API
- **THEN** registrar SHALL 提供 `APIPrefix()` 或等价方法返回 `/x/{plugin-id}/api/v1`
- **AND** 方法名称和文档必须表达这是插件 API 路由强制前缀，不是所有源码插件 HTTP 路由的强制前缀

#### Scenario: 主框架不感知插件路由分组语义
- **WHEN** 源码插件在代码中按公开页面、工作台 API 或静态资源组织多个 route group
- **THEN** 主框架只提供路由挂载与插件启停边界
- **AND** 主框架不得把这些 route group 自动分类为门户路由、工作台路由或权限资源

#### Scenario: 只有插件 API 路由可进入 OpenAPI
- **WHEN** 源码插件同时注册 `/x/my-plugin/api/v1/reports` API 路由和 `/portal/my-plugin` 公开页面路由
- **THEN** 只有符合 GoFrame DTO 文档形态的 `/x/my-plugin/api/v1/reports` MAY 被投影到 OpenAPI
- **AND** `/portal/my-plugin` 不得因为可通过 HTTP 访问而自动进入 OpenAPI、菜单、权限节点或插件资源引用

### Requirement: 源码插件 HTTP 路由冲突必须以 GoFrame 注册结果为准

系统 SHALL 以 GoFrame 实际路由注册结果作为源码插件 HTTP 路由冲突检测依据。主框架不得维护一份独立的源码插件 HTTP 路由声明清单来替代实际注册行为。

#### Scenario: 插件间路由冲突启动失败
- **WHEN** 两个源码插件注册相同 HTTP 方法和路径的路由
- **THEN** GoFrame 路由注册冲突必须导致启动或注册失败
- **AND** 系统不得静默选择其中一个插件路由

#### Scenario: 代码注册是唯一事实来源
- **WHEN** 源码插件代码没有注册某个 HTTP 路由
- **THEN** 主框架不得因为配置或 manifest 中存在菜单而假定该 HTTP 路由存在
- **AND** 对该路径的请求只按实际已注册路由处理

### Requirement: 动态插件 API 公开路径必须使用统一插件 API 命名空间

系统 SHALL 将动态插件 route contract 映射到 `/x/{plugin-id}/api/v1/...` 公开 API 路径。`/x` 前缀表示统一插件 API 命名空间，`{plugin-id}` 表示插件身份，`api/v1` 表示插件 API 版本。动态插件 route contract 仍 SHALL 以插件声明的内部 route path、HTTP method、access 和 permission 作为匹配与鉴权事实来源。动态插件分发器不得以 `/x/*` 通配方式吞掉已注册源码插件 API；当 `{plugin-id}` 对应源码插件时，请求 SHALL 按源码插件 GoFrame route handler 处理。

#### Scenario: 动态插件 route contract 映射为公开 API 路径
- **WHEN** 动态插件 ID 为 `my-plugin`
- **AND** route contract 声明 `path: /api/v1/reports/{id}` 和 `method: GET`
- **THEN** 系统公开路径为 `GET /x/my-plugin/api/v1/reports/{id}`
- **AND** 宿主传递给动态插件的 route snapshot 仍包含声明的 route path、内部 path、公开 path 和路径参数

#### Scenario: 源码插件与动态插件共享统一插件 API 命名空间
- **WHEN** 用户查看 `/x/source-plugin/api/v1/reports` 和 `/x/dynamic-plugin/api/v1/reports`
- **THEN** 两者均使用统一插件 API 命名空间
- **AND** 源码插件请求由源码插件 GoFrame handler 执行
- **AND** 动态插件请求由宿主动态运行时桥接分发执行

#### Scenario: 动态插件旧式无 api 版本路径需要迁移
- **WHEN** 动态插件 route contract 或菜单示例仍引用 `/x/{plugin-id}/{resource}` 形式的旧公开 API 路径
- **THEN** 本变更实现中必须迁移为 `/x/{plugin-id}/api/v1/{resource}`
- **AND** OpenAPI 投影、前端调用、示例插件和 E2E 断言必须使用新的动态插件 API 路径
