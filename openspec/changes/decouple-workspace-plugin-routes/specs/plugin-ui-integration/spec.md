## ADDED Requirements

### Requirement: 源码插件 HTTP 路由不得被强制绑定到管理工作台 UI 模型

系统 SHALL 将源码插件通过代码注册的 HTTP 路由视为插件实现细节。管理工作台只通过 `plugin.yaml` 的 `menus` 声明感知插件贡献的工作台导航和权限；源码插件公开页面、门户、静态资源和自管 fallback 不得自动投影为菜单、工作台页面、权限节点、OpenAPI 文档或插件资源引用。源码插件在统一插件 API 前缀下注册且符合文档条件的 API 路由 MAY 进入 OpenAPI 投影，但该能力不得扩展为对所有 HTTP 路由的自动投影。

#### Scenario: 插件公开 HTTP 路由不进入菜单治理
- **WHEN** 源码插件通过代码注册 `/` 或 `/portal/*` 等公开 HTTP 路由
- **THEN** 管理工作台菜单列表不得因为这些 HTTP 路由自动新增菜单
- **AND** 角色授权不得因为这些 HTTP 路由自动新增权限节点

#### Scenario: 插件工作台菜单仍由 manifest 管理
- **WHEN** 源码插件需要向管理工作台贡献导航入口
- **THEN** 插件仍必须通过 `plugin.yaml` 的 `menus` 声明菜单、按钮权限、排序和父子关系
- **AND** 主框架按菜单声明同步 `sys_menu` 并参与角色授权

#### Scenario: 示例插件菜单归入扩展中心
- **WHEN** 官方源码插件示例和动态插件示例安装后同步工作台菜单
- **THEN** 源码插件示例和动态插件示例菜单必须挂载到宿主 `extension` 扩展中心目录下
- **AND** 扩展中心目录下的可见菜单顺序必须为插件管理、源码插件示例、动态插件示例

#### Scenario: 菜单声明不代表 HTTP 路由存在
- **WHEN** `plugin.yaml` 声明某个工作台菜单路径
- **THEN** 主框架只将该路径作为管理工作台动态路由和导航资源
- **AND** 主框架不得据此推断源码插件已经注册了同路径 HTTP handler

#### Scenario: 公开 HTTP 路由不进入 OpenAPI
- **WHEN** 源码插件通过代码注册 `/portal/*`、`/assets/*` 或自管 SPA fallback
- **THEN** 主框架不得因为这些路由可通过 HTTP 访问而自动生成 OpenAPI 路径
- **AND** 只有统一插件 API 前缀下符合 GoFrame DTO 文档形态的源码插件 API MAY 参与 OpenAPI 投影

### Requirement: 源码插件前端实现必须允许脱离管理工作台 SPA bundle

系统 SHALL 允许源码插件在其 HTTP handler 中自行返回页面、静态资源、SPA fallback 或其他内容。主框架不得要求源码插件公开页面或静态资源必须参与默认管理工作台 SPA bundle。源码插件和动态插件 MAY 在 `plugin.yaml` 的 `public_assets` 中声明由宿主托管的公开静态资源目录；该声明只表达 public asset 托管，不表达 HTTP 路由、门户路由或工作台页面语义。

#### Scenario: 源码插件自行返回页面内容
- **WHEN** 源码插件 HTTP handler 返回 HTML、JSON、文件流或其他响应
- **THEN** 主框架不得要求这些响应来自管理工作台前端构建产物
- **AND** 主框架不得要求插件为自行注册的 HTTP handler 声明对应静态资源目录

#### Scenario: 管理工作台继续消费 manifest 菜单页面
- **WHEN** 源码插件的 `plugin.yaml` 菜单指向 `system/plugin/dynamic-page` 或其他工作台组件
- **THEN** 管理工作台继续按菜单和权限加载对应页面
- **AND** 这不影响同一插件通过 HTTP 路由闭环维护公开页面或静态资源

#### Scenario: `/x-assets` 不是源码插件公开前端的强制入口
- **WHEN** 源码插件选择通过自己的 HTTP 路由服务静态资源
- **THEN** 系统不得强制其使用 `/x-assets/{plugin-id}/{version}/...`
- **AND** `/x-assets` 只能作为宿主提供的可选插件 `public_assets` 托管入口
- **AND** 系统不得保留 `/plugin-assets` 兼容入口

### Requirement: 插件公开静态资源必须通过声明式 public asset 托管

系统 SHALL 允许源码插件和动态插件在 `plugin.yaml` 的 `public_assets` 根字段中声明可公开的静态资源目录，并由宿主在 `/x-assets/{plugin-id}/{version}/...` 下统一托管这些 public assets。宿主不得因为插件注册了 embedded files、运行时 artifact 或 manifest 资源而默认公开整个插件资源包；只有 `public_assets` 明确声明的插件自有资源才能被托管。每个声明项 MUST 使用 `source` 指向插件内相对目录或动态 artifact asset 前缀，并 MAY 使用 `mount` 指定在 `/x-assets/{plugin-id}/{version}/` 下的相对挂载路径；`mount` 为空或 `/` 时，声明目录内文件直接挂载到版本根。每个声明项 MAY 使用 `index` 指定访问该挂载目录本身时默认返回的文件名；未配置 `index` 时默认使用 `index.html`。宿主 SHALL 使用严格路径 schema 校验 `public_assets`，非法声明必须阻止插件校验、安装或启用。宿主 SHALL NOT 按目录名维护治理黑名单或前端目录白名单；插件配置声明的插件内目录即为发布授权边界。本变更 SHALL NOT 保留 `/plugin-assets` 旧路径兼容。

#### Scenario: 源码插件声明 public assets
- **WHEN** 源码插件通过 `plugin.Assets().UseEmbeddedFiles(...)` 注册 embedded files
- **AND** 插件在 `plugin.yaml` 的 `public_assets` 中声明 `source: frontend/public` 和 `mount: /`
- **THEN** 宿主从该 embedded filesystem 中读取声明目录，并将 `frontend/public/logo.png` 映射为 `/x-assets/{plugin-id}/{version}/logo.png`
- **AND** 未声明目录中的文件不得被该路由访问

#### Scenario: 动态插件声明 public assets
- **WHEN** 动态插件 artifact 中包含 frontend assets
- **AND** 插件在 `plugin.yaml` 的 `public_assets` 中声明 `source` 前缀和可选 `mount`
- **THEN** 宿主通过 `/x-assets/{plugin-id}/{version}/...` 仅提供声明匹配的 public asset 集合
- **AND** SQL、生命周期、host-service、路由契约、i18n 和 apidoc 资源不得被 public asset resolver 读取

#### Scenario: 动态插件既有 frontend assets 迁移到声明式模型
- **WHEN** 动态插件 artifact 中存在 frontend asset，但 `plugin.yaml` 未在 `public_assets` 中声明对应 `source`
- **THEN** `/x-assets/{plugin-id}/{version}/...` 不得因为 artifact 中存在该 frontend asset 而自动返回资源
- **AND** 依赖该资源的动态插件菜单或页面必须迁移为引用 `public_assets` 声明后生成的访问路径

#### Scenario: public asset 声明以插件自身资源边界为安全边界
- **WHEN** 插件声明 `manifest`、`backend` 或其他插件自身目录为 public assets
- **THEN** 宿主必须按插件作者的显式声明发布这些目录中的资源
- **AND** 当 `source` 使用 `../`、绝对路径、URL 或符号链接逃逸插件自身资源边界时，宿主必须拒绝该声明或使插件校验失败

#### Scenario: public asset 声明使用严格 schema
- **WHEN** 插件声明空 `source`、绝对路径、包含 `../` 的路径、不存在的 `source`、符号链接逃逸插件资源边界、重复或互相覆盖的 `mount`
- **THEN** 宿主必须拒绝该声明并使插件校验、安装或启用失败
- **AND** 宿主不得静默忽略非法声明项
- **AND** 可访问文件的 Content-Type 必须按文件扩展名或等价 MIME 推断能力确定

#### Scenario: public asset 目录访问使用声明的 index 文件
- **WHEN** 插件在 `public_assets` 中声明 `source: frontend/public`、`mount: /docs` 和 `index: index.htm`
- **THEN** 浏览器访问 `/x-assets/{plugin-id}/{version}/docs` 或 `/x-assets/{plugin-id}/{version}/docs/` 时，宿主必须返回 `frontend/public/index.htm`
- **AND** 当 `index` 未配置时，宿主必须继续默认返回声明目录下的 `index.html`
- **AND** `index` 必须是安全的相对文件名，不能是目录、绝对路径、URL、通配符或包含 `../` 的路径

#### Scenario: 插件版本作为 public asset 缓存边界
- **WHEN** public asset 通过 `/x-assets/{plugin-id}/{version}/...` 访问
- **THEN** 同一 `{plugin-id, version}` 下的资源内容必须保持稳定
- **AND** 插件变更 public asset 内容时必须升级插件版本或使用明确的内容版本机制

#### Scenario: 插件不可用时 public assets 不可访问
- **WHEN** 插件未安装、未启用或当前租户不可用
- **THEN** `/x-assets/{plugin-id}/{version}/...` 默认返回 404
- **AND** 请求缺少租户上下文且无法判定租户时，系统按插件全局启用状态判断可用性
- **AND** 租户专属、用户专属或需要鉴权的文件不得通过 `public_assets` 发布，应由插件自管 HTTP 路由并自行挂载鉴权与租户中间件

### Requirement: 动态插件工作台页面必须继续通过动态页组件承载公开资源

系统 SHALL 将动态插件 hosted frontend 菜单迁移为继续使用 `system/plugin/dynamic-page` 或等价工作台承载组件，并通过菜单 query 或等价参数引用 `/x-assets/{plugin-id}/{version}/...` 下的声明式 public asset 入口。管理工作台不得把 `/x-assets/{plugin-id}/{version}/index.html` 这类资源 URL 直接作为动态插件菜单路由路径来替代动态页组件。

#### Scenario: 动态插件菜单通过动态页组件引用 x-assets
- **WHEN** 动态插件菜单需要加载 embedded-mount 或 iframe 前端资源
- **THEN** 菜单组件仍指向 `system/plugin/dynamic-page` 或等价动态页组件
- **AND** 菜单 query 或等价参数引用 `/x-assets/{plugin-id}/{version}/...` 下的声明式 public asset URL
- **AND** 前端刷新、iframe、embedded-mount 和插件页面热升级匹配逻辑必须识别 `/x-assets` 资源 URL

#### Scenario: 动态插件菜单不得直接使用 x-assets 作为工作台路由
- **WHEN** 动态插件菜单 path 直接声明为 `/x-assets/{plugin-id}/{version}/index.html`
- **THEN** 系统不得把该 path 当作管理工作台动态路由来替代 `system/plugin/dynamic-page`
- **AND** 插件开发文档和测试夹具必须使用动态页组件加资源 URL 参数的迁移形态
