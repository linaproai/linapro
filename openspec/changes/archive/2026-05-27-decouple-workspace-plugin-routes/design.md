## Context

当前宿主 HTTP 启动顺序中，宿主 API 与动态插件 `/x` 分发器先注册，源码插件随后通过 `pluginhost.HTTPRegistrar` 注册 GoFrame 路由，最后宿主把前端静态资源与 SPA fallback 绑定到 `/*`。由于默认管理工作台前端路由根节点也是 `/`，且 SPA fallback 会把未匹配路径回退到管理台 `index.html`，默认工作台实际占用了根前端资源边界。后续实现需要把 `/x` 从动态插件专属分发入口调整为统一插件 API 命名空间，避免源码插件与动态插件 API 使用两套前缀。

源码插件已经具备代码级 HTTP 路由注册能力，也已经通过 `plugin.yaml` 的 `menus` 声明工作台菜单与权限。问题不在于缺少“门户”功能，而在于默认管理工作台与根路径、静态资源 fallback、工作台路由 base 强绑定，导致源码插件无法闭环维护自己的公开 HTTP 页面、静态资源或 SPA fallback。

本设计把默认管理工作台视为内建工作台应用，给它一个可配置入口路径；源码插件 HTTP 路由仍然是插件内部实现细节，主框架不感知这些路由属于公开页面、静态资源还是工作台 API。工作台可见性和权限仍只由 `plugin.yaml` 菜单治理。对于无需插件自行编写静态路由的公开文件，源码插件和动态插件可以通过 `plugin.yaml` 中的 `public_assets` 声明交由宿主统一托管到 `/x-assets/{plugin-id}/{version}/...`，但该声明只表达可公开静态资源目录和其在插件资产命名空间内的挂载点，不表达 HTTP 路由、门户路由或工作台页面语义。本变更不考虑兼容旧 `/plugin-assets` 入口。

## Goals / Non-Goals

**Goals:**

- 让默认管理工作台通过可配置的 `basePath` 访问，默认值确定为 `/admin`，不再默认独占 `/`；部署方可在独立管理后台域名或根路径由工作台独占的部署中显式配置为 `/`；`basePath` 以启动期配置为权威，不作为运行期热修改配置。
- 让源码插件可以通过现有 GoFrame 路由注册机制闭环注册 `/`、`/portal/*`、`/assets/*` 或其他非保留路径。
- 保持插件 HTTP 路由以代码注册为唯一事实来源；不在 `plugin.yaml` 中声明公开路由、门户路由、工作台 API 路由或路由分组语义。
- 支持源码插件和动态插件声明可由宿主统一托管的公开静态资源目录，宿主只保证声明路径不能逃逸插件自身资源边界，是否公开插件内某个目录由 `public_assets` 声明负责。
- 保持 `plugin.yaml` 的 `menus` 作为工作台菜单、权限和动态路由治理的事实来源。
- 利用 GoFrame 路由注册冲突作为源码插件 HTTP 路由冲突检测机制；宿主不维护单独的源码插件 HTTP 路由清单。
- 将源码插件和动态插件 API 统一到 `/x/{plugin-id}/api/v1` 插件 API 命名空间，让用户、前端 helper、网关、审计、OpenAPI 和测试使用同一插件 API base 规则；`/x/**` 只承载插件 API，不承载插件公开页面、门户、静态资源或自管 fallback。
- 确保已注册宿主控制面 API 路由、统一插件 API 命名空间 `/x`、统一插件资产命名空间 `/x-assets`、管理工作台 basePath 等宿主边界优先绑定并不可被源码插件公开路由覆盖；源码插件 API 不再使用宿主 `/api/v1` 命名空间，源码插件公开页面、门户、静态资源或自管 fallback 仍可注册其他非保留路径。

**Non-Goals:**

- 不实现具体门户、CMS、站点渲染、静态文件服务或 SSR 能力。
- 不要求源码插件使用 `/x-assets/{plugin-id}/{version}/...` 托管自己的静态资源；源码插件可以自行决定响应内容和资源路径。`/x-assets` 仅作为可选的声明式 public asset 托管入口，而不是源码插件公开前端或 HTTP 路由的唯一入口；旧 `/plugin-assets` 入口不作为兼容保留项。
- 不把源码插件 HTTP 路由自动投影为菜单、权限、OpenAPI 或资源引用。
- 不把源码插件公开页面、门户、静态资源或自管 fallback 投影为 OpenAPI；只有插件 API 前缀下符合文档条件的 API 路由可以参与 OpenAPI 投影。
- 不重构现有 `plugin.yaml` 菜单声明模型。
- 不要求主框架理解插件内部路由分组语义，例如“门户路由”或“工作台路由”。

## Decisions

1. 默认管理工作台使用可配置的 `basePath`，默认值为 `/admin`。

   配置校验必须拒绝空值、通配符、相对路径和宿主保留命名空间，但不得拒绝根路径 `/`。当部署方给管理后台配置独立域名，或明确让工作台独占当前域名根路径时，`workspace.basePath=/` 是有效配置。工作台 `basePath` 以启动期配置为权威，后端 fallback、前端 router base、构建或启动配置、开发代理和 E2E 访问地址必须使用同一值；该配置不作为运行期热修改项处理，修改后需要重新启动对应服务并重新加载前端资源。默认值仍为 `/admin`，以便默认部署继续把根路径留给源码插件公开路由或宿主自定义公开页面。

2. 后端 SPA fallback 只服务管理工作台 basePath。

   当前 catch-all fallback 仍可作为最终静态资源处理器存在，但它必须先判断请求路径是否属于工作台 basePath。属于工作台路径时返回管理台静态资源或 `index.html`；不属于时不得吞掉请求，应让前面已经注册的宿主 API、动态插件、源码插件路由决定响应，最终未匹配请求返回 404 或 GoFrame 默认结果。

3. 源码插件 HTTP 路由不进入 manifest 声明。

   用户指出代码可能与声明不一致，因此插件 HTTP 路由以 `registerRoutes` 中的 GoFrame 注册行为为准。`plugin.yaml` 继续声明插件身份、版本、依赖、租户治理和工作台菜单；不新增 `public_routes`、`portal_routes` 或资源路径声明。

4. 主框架不感知源码插件 HTTP 路由分组语义。

   源码插件可在内部按插件 API、公开页面、门户、静态资源等分组组织路由，但宿主只提供注册器和插件启停包裹，不判断非 API 路由是公开页面、工作台页面、静态资源还是其他用途。插件 API 必须使用统一插件 API 命名空间 `/x/{plugin-id}/api/v1`；工作台权限治理只发生在菜单与后端 API 自身中间件上。

5. 插件 API 统一使用 `/x/{plugin-id}/api/v1/...` 前缀，且 `/x/**` 只用于插件 API。

   源码插件仍然拥有自由的 GoFrame 路由注册能力，宿主不强制其所有 HTTP 路由进入 `/x`。但源码插件对外暴露的插件 API 必须进入统一插件 API 命名空间 `/x/{plugin-id}/api/v1`，避免源码插件 API 与宿主控制面 `/api/v1`、公开页面、门户或静态资源路径混用。`/x/**` 是宿主保留的统一插件 API 命名空间，源码插件不得在 `/x/{plugin-id}` 下注册非 API 路由，例如 `/x/{plugin-id}/assets/*`、`/x/{plugin-id}/portal/*` 或 `/x/{plugin-id}/health`；公开页面、门户、静态资源或自管 fallback 必须使用 `/`、`/portal/*`、`/assets/*`、`/x-assets` 或其他非保留路径。源码插件 registrar 应提供 `APIPrefix()` 或等价方法，返回 `/x/{plugin-id}/api/v1`。源码插件开发文档、测试夹具和官方示例必须使用：

   ```text
   /x/{plugin-id}/api/v1/{resource}
   ```

   这里的 `api/v1` 表示插件自己的 HTTP API 版本，不表示宿主控制面 `/api/v1` 版本。源码插件若要注册公开页面、门户、静态资源或根路径 SPA fallback，仍应根据自身需求使用 `/`、`/portal/*`、`/assets/*` 或其他非保留路径，而不是被强制放入插件 API 前缀或 `/x` 命名空间。

6. 动态插件 API 复用统一插件 API 命名空间 `/x/{plugin-id}/api/v1/...`。

   `/x` 不再表达动态插件专属数据面，而是统一插件 API 命名空间。动态插件仍由宿主根据 `{plugin-id}` 和活动 runtime artifact 执行桥接分发；动态插件 route contract 的 `path` 必须声明为 `/api/v1/...`，宿主拼接后形成：

   ```text
   /x/{plugin-id}/api/v1/{resource}
   ```

   源码插件与动态插件在 URL 上共享相同插件 API 规则，执行模型仍不同：源码插件 API 由 GoFrame route handler 和插件自行挂载的中间件处理，动态插件 API 由宿主根据 route contract 匹配并桥接到 runtime artifact。动态插件内部 route contract 仍以方法与内部 path 作为匹配依据，`access`、`permission`、路径参数、OpenAPI 投影和宿主鉴权继续基于 route contract 执行。现有动态插件示例、菜单引用、OpenAPI 快照、E2E 和开发文档需要从 `/x/{plugin-id}/{resource}` 迁移为 `/x/{plugin-id}/api/v1/{resource}`；官方源码插件示例也必须迁移到 `/x/{plugin-id}/api/v1/{resource}`。

7. 路由冲突检测复用 GoFrame 注册冲突，并由 registrar 明确拒绝 `/x` 非 API 注册。

   GoFrame 在默认路由注册冲突时会报错或导致启动失败，这比维护一份宿主侧源码插件路由清单更贴近事实来源。宿主需要保证注册顺序让保留边界先绑定：宿主已注册 API 路由、统一插件 API 命名空间 `/x`、管理工作台 basePath 和插件资产命名空间必须先于可冲突的源码插件公开路由或在注册器中明确拒绝被覆盖。`/x/{plugin-id}/api/v1/...` 是插件 API 的强制路径，源码插件 API 必须注册到自身插件 ID 对应的 `/x/{plugin-id}/api/v1` 子树；动态插件分发器不得以 `/x/*` 通配方式吞掉已注册源码插件 API，而应只在动态插件 ID 子树或未被源码插件具体路由占用的路径上执行动态 route contract 分发。源码插件不得把插件 API 继续挂在宿主 `/api/v1` 下，也不得在 `/x` 下注册非 API 路由；源码插件公开页面、门户、静态资源和自管 fallback 仍可使用其他非保留路径。

8. 工作台菜单仍由 `plugin.yaml` 管理。

   插件贡献到管理工作台的导航入口、按钮权限、父子关系、排序和可见性仍通过 `menus` 同步到 `sys_menu`，参与角色授权与租户/插件启停过滤。HTTP 路由注册不会自动创建或修改菜单。

9. 前端路由 base 与 API base 分离。

   管理工作台页面 URL 受 `workspace.basePath` 影响，宿主控制面 API 仍使用 `/api/v1`，源码插件和动态插件 API base 均使用 `/x/{plugin-id}/api/v1`。前端启动、登录跳转、刷新、E2E base URL、插件前端 API helper 和开发代理需要显式处理这些 base path，不能假设工作台一定部署在 `/` 或 `/admin`，也不能把宿主 `/api/v1` 当成任何插件 API 的默认 base。开发模式下以后端公开地址作为统一访问入口，后端在工作台 `basePath` 下代理或转发到 Vite dev server；默认 `/admin` 配置下，开发者访问 `/admin` 时进入管理工作台，访问 `/` 或插件公开路由时仍到达后端源码插件路由，Vite 不得以根路径 SPA fallback 吞掉这些请求；显式配置为 `/` 时，根路径由工作台占用，源码插件公开根路由冲突必须在注册阶段暴露。

10. `/x-assets` 支持声明式 public asset 托管。

   动态插件现有 `/plugin-assets/{plugin-id}/{version}/...` frontend assets 访问路径迁移到统一的声明式 public asset 模型 `/x-assets/{plugin-id}/{version}/...`，不再把运行时 artifact 中的 frontend assets 视为天然公开，也不保留旧入口兼容。源码插件和动态插件都必须在 `plugin.yaml` 根字段 `public_assets` 中声明可公开目录及挂载点，宿主采用严格路径 schema 校验并只暴露声明匹配的资源。声明项使用 `source` 指向插件内相对目录或动态 artifact asset 前缀，使用可选 `mount` 指定其在 `/x-assets/{plugin-id}/{version}/` 下的相对挂载路径；`mount` 为空或 `/` 时，声明目录内文件直接挂载到版本根。声明项还可以使用可选 `index` 指定访问该挂载目录本身时返回的默认文件，未配置时默认 `index.html`；`index` 必须是安全的相对文件名，不得是目录、绝对路径、URL、通配符或包含 `../` 的路径。示例：`source: frontend/public`、`mount: /` 时，`frontend/public/logo.png` 映射为 `/x-assets/{plugin-id}/{version}/logo.png`；`index: index.htm` 时，访问版本根或对应挂载目录会解析为声明 source 下的 `index.htm`。源码插件从 `plugin.Assets().UseEmbeddedFiles(...)` 注册的 embedded FS 或插件根目录中读取声明目录；动态插件从运行时 artifact 中匹配 `source` 前缀的 frontend asset 集合读取。声明必须拒绝空 `source`、绝对路径、包含 `../` 的路径、重复或互相覆盖的 `mount`、不存在的 `source`、符号链接逃逸插件资源边界以及无法安全规范化的路径；Content-Type 按扩展名或等价 MIME 推断能力确定。宿主不根据 `backend`、`manifest`、`plugin.yaml` 等目录或文件名维护 public asset 黑名单，也不把声明限制到 `frontend/public` 白名单；插件作者配置的插件内目录即为发布授权边界。

   public asset URL 以插件 ID 和版本为缓存边界，同一 `{plugin-id, version}` 下资源内容必须稳定。源码插件或动态插件变更 public asset 内容时必须升级 `plugin.yaml` 版本，或在设计中引入等价内容版本机制。`public_assets` 默认匿名可读，但必须满足插件已安装、已启用和当前租户可用；请求缺少租户上下文且无法判定租户时，按全局启用状态判断，插件未安装、未启用或当前租户不可用时返回 404。租户专属、用户专属或需要鉴权的文件不应放入 `public_assets`，应由插件自管 HTTP 路由并自行挂载鉴权与租户中间件。

11. 动态插件工作台页面继续由 `system/plugin/dynamic-page` 承载。

   动态插件既有 hosted frontend 菜单迁移为继续使用 `system/plugin/dynamic-page` 等工作台承载组件，由菜单 query 或等价参数引用 `/x-assets/{plugin-id}/{version}/...` 下的声明式 public asset 入口。菜单 `path` 不直接作为 `/x-assets/.../index.html` 的工作台路由，避免绕过工作台动态页治理、iframe/embedded-mount 访问模式、刷新匹配和权限入口。前端热升级、iframe 刷新、embedded-mount 加载和动态页匹配逻辑需要把旧 `/plugin-assets` 识别规则迁移到 `/x-assets`。

12. 插件 API 的 OpenAPI 投影只覆盖 API 路由。

   源码插件通过 `APIPrefix()` 注册、且符合 GoFrame DTO 文档形态的插件 API 可以投影为 OpenAPI；动态插件 route contract 继续按 `/x/{plugin-id}/api/v1/...` 公开路径投影。源码插件公开页面、门户、静态资源、自管 fallback 以及 `public_assets` 文件访问不得自动进入 OpenAPI，也不得因为可通过 HTTP 访问而生成权限节点或资源引用。

## Risks / Trade-offs

- [Risk] 源码插件注册 `/` 后遮挡其他插件或宿主未来公开能力。
  Mitigation：复用 GoFrame 路由冲突失败作为检测机制，并将宿主保留命名空间写入规范和测试；插件间冲突在启动注册阶段暴露。

- [Risk] 管理工作台 basePath 与插件路由、宿主已注册 API 路由、`/x` 或 `/x-assets` 冲突。
  Mitigation：配置加载时校验 basePath，不允许空值、通配符、相对路径、保留命名空间或明显不规范路径；源码插件 API 必须使用 `/x/{plugin-id}/api/v1`，实际冲突由 GoFrame 注册结果暴露，官方示例必须同步迁移。

- [Risk] 前端构建 base、router base 和后端 fallback 口径不一致。
  Mitigation：以启动期工作台 `basePath` 配置作为权威来源，同步投影到前端构建/启动配置、后端 fallback 和开发代理；新增前后端集成测试覆盖刷新和登录跳转；插件前端调用应使用插件 API base helper 区分宿主 `/api/v1` 与插件 `/x/{plugin-id}/api/v1`。

- [Risk] 将 `/x/{plugin-id}/api/v1` 误认为源码插件所有 HTTP 路由的强制前缀。
  Mitigation：接口名使用 `APIPrefix()` 或等价名称，文档明确它只用于插件 API 路由；公开页面、门户、静态资源和插件自管 fallback 仍由源码插件自由注册。

- [Risk] 统一 `/x` 后动态插件通配分发吞掉源码插件 API，或源码插件在 `/x` 下注册非 API 路由导致语义混乱。
  Mitigation：实现必须把 `/x` 定义为统一插件 API 命名空间，动态插件分发器只接管动态插件 ID 子树或未被源码插件具体路由占用的动态插件路径，registrar 必须拒绝源码插件 `/x` 下非 `/x/{own-plugin-id}/api/v1/...` 的注册；新增路由优先级测试覆盖源码插件 `/x/{source-plugin-id}/api/v1/...` 能命中源码插件 handler，动态插件 `/x/{dynamic-plugin-id}/api/v1/...` 仍能命中 runtime dispatcher，未知插件 ID 返回明确 404。

- [Risk] 源码插件 HTTP 路由被误认为不需要鉴权。
  Mitigation：文档明确 HTTP 路由是插件实现细节；插件公开路由、工作台 API 路由和静态资源路由的鉴权策略由插件代码通过中间件自行决定。工作台菜单权限不自动保护任意 HTTP 路由。

- [Risk] 插件 public asset 声明误暴露插件内敏感文件。
  Mitigation：`public_assets` 被视为插件作者的显式发布授权，宿主不按目录名判断敏感性；插件审查应关注声明目录是否适合匿名公开。宿主强制校验 `source`、`mount`、`index` 不能使用 URL、绝对路径、`../`、通配符或符号链接逃逸插件资源边界；动态插件 artifact 中 SQL、生命周期、host-service、路由契约、i18n 和 apidoc 仍保留在独立 section，不会被 `/x-assets` resolver 读取。

- [Risk] `/x-assets/{plugin-id}/{version}/...` 版本 URL 下资源内容变化导致浏览器或 CDN 缓存陈旧。
  Mitigation：要求同一插件版本下 public asset 内容不可变；资源变化必须 bump 插件版本或采用明确内容版本机制。实现与审查中需要覆盖源码插件和动态插件的版本边界，并将动态插件既有 frontend assets 路径迁移到 `plugin.yaml public_assets` 声明。

- [Risk] public assets 被误用于租户专属或鉴权资源。
  Mitigation：`public_assets` 默认匿名可读但受插件启用和租户可用性治理约束；租户专属、用户专属或需要鉴权的资源必须由插件自管 HTTP 路由并挂载鉴权/租户中间件。测试需要覆盖无租户上下文、插件禁用和租户不可用时的 404 行为。

- [Risk] i18n 或缓存治理被误触发为全量改造。
  Mitigation：本变更只调整工作台入口配置、路由 fallback 和 public asset 托管契约，不新增运行时业务缓存；若新增配置界面文案则同步 i18n，否则在任务记录中明确无运行时文案影响。public asset 托管不改变运行时 i18n 资源加载边界；只有插件显式把插件内 i18n 目录声明为 `public_assets` 时，宿主才会按普通静态资源发布该声明目录。

## Implementation Plan

1. 新增工作台 basePath 配置与校验，默认值为 `/admin`，允许 `/`，并拒绝空值、通配符、相对路径和宿主保留命名空间。
2. 调整后端前端静态资源 handler，使管理台静态资源和 SPA fallback 只匹配 basePath。
3. 调整前端 router base、登录跳转、刷新路径、public frontend 配置消费和开发代理；开发模式以后端公开地址作为统一入口，后端在工作台 basePath 下代理 Vite。
4. 保持源码插件 HTTP registrar 的代码注册模型，不新增 manifest 路由声明；新增 `APIPrefix()` 或等价方法，补充测试证明源码插件可注册根路径、插件 API 必须使用 `/x/{plugin-id}/api/v1`，且主框架不感知源码插件非 API 路由分组。
5. 调整 `/x` 动态插件分发边界、动态插件示例和 route contract 测试，使动态插件 API 公开路径使用 `/x/{plugin-id}/api/v1/...`，且不吞掉源码插件同一统一插件 API 命名空间下的具体路由。
6. 增加插件 `plugin.yaml public_assets` 声明契约和 resolver 边界，验证源码插件 embedded FS 与动态插件 artifact 仅暴露声明的 public assets，且声明路径不能通过 `../`、绝对路径、URL 或符号链接逃逸插件自身资源边界；同步迁移动态插件既有 frontend assets 访问路径到该声明式模型，并保持 `system/plugin/dynamic-page` 工作台承载模型。
7. 更新 E2E 与开发文档中的默认管理台访问路径、源码/动态插件 API base 和插件 public asset 托管说明。
8. 运行 Go 路由绑定测试、前端单元测试、相关 E2E、`openspec validate`。

## Open Questions

- 无。2026-05-21 已澄清开发模式以后端公开地址作为统一入口，后端在工作台 basePath 下代理或转发到 Vite dev server。
