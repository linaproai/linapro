# Design

## 插件 ID 治理边界

插件 ID 是跨运行时边界的稳定身份。它会同时进入 URL path、动态资源路径、文件名、数据库键、菜单 key、权限字符串、`i18n` namespace、`apidoc` namespace 和宿主能力发现，因此运行时只强制最基础也最稳定的安全边界：非空、总长度不超过 64 字符、并且使用 lowercase kebab-case。

`<author>-<domain>-<capability>` 保留为官方插件命名建议和仓库治理约定，而不是宿主运行时硬编码的拒绝规则。这样既能为官方生态提供稳定分类方式，也不会把 domain 白名单、capability 保留字或旧官方 ID 拒绝表强塞进宿主运行时。

官方插件统一使用以下规范化映射作为唯一正向身份：

| 旧 ID | 新 ID |
| --- | --- |
| `content-notice` | `linapro-content-notice` |
| `monitor-loginlog` | `linapro-monitor-loginlog` |
| `monitor-operlog` | `linapro-monitor-operlog` |
| `monitor-online` | `linapro-monitor-online` |
| `monitor-server` | `linapro-monitor-server` |
| `multi-tenant` | `linapro-tenant-core` |
| `org-center` | `linapro-org-core` |
| `plugin-demo-dynamic` | `linapro-demo-dynamic` |
| `plugin-demo-source` | `linapro-demo-source` |
| `demo-control` | `linapro-ops-demo-guard` |

该映射是破坏式治理调整，不保留旧 ID alias、重定向或兼容查询。所有由插件 ID 派生的运行时身份都必须同步切换到当前插件 ID，包括插件状态表、发布表、迁移表、资源引用表、节点状态表、菜单、权限、`cron handlerRef`、动态资源路径、公开数据面路径、运行时 `i18n` 和 `apidoc` 命名空间。插件自有存储命名、DAO 生成配置和示例数据也应跟随新的 `plugin_id snake_case` 范围重建，避免长期保留新旧身份不一致的技术债。

## Manifest、生命周期与仓库治理一致性

插件治理的关键不是把所有命名策略都塞进运行时，而是把运行时基础校验和仓库一致性治理分层执行。manifest 加载、源码插件注册、动态插件 artifact 校验和插件依赖声明统一复用同一套插件 ID 基础安全校验，发现空 ID、超长 ID 或不安全字符时立即失败。

在此基础上，仓库治理扫描负责覆盖运行时难以完整表达的一致性约束：插件目录名等于 manifest ID，源码注册 ID 等于 manifest ID，菜单 key 使用 `plugin:<plugin-id>:` 前缀，运行时语言包使用 `plugin.<plugin-id>.` 前缀，`apidoc` 语言包使用 `plugins.<plugin_id_snake_case>.` 前缀，动态资源 URL 与公开数据面路径也必须使用当前插件 ID。这样可以让第三方插件继续在基础安全边界内自由命名，同时确保官方插件资产内部没有身份漂移。

## 动态插件公开路由治理

动态插件公开数据面入口固定为 `/x/{pluginId}/...`。宿主只负责识别 `/x` 命名空间并解析 `pluginId`，`{pluginId}` 之后的路径完全归插件声明所有。插件可以在内部路径中自行表达 `/api/v1/...`、`/api/v2/...`、`/graphql` 等版本或协议语义，宿主不再把插件数据面错误地绑定到控制面 `/api/v1` 前缀。

旧 `/api/v1/extensions/{pluginId}/...` 入口直接移除，不使用双路由或重定向。动态插件请求方法和请求体形态较复杂，保留兼容入口会继续传播错误路径语义，也会引入方法保持和请求体重放差异，因此本设计直接将 `/x/{pluginId}/...` 作为唯一 canonical 公开路径。

虽然 `/x` 路由移出了宿主 `/api/v1` 分组，但它仍必须复用宿主统一 HTTP 治理链路：响应包装、CORS、请求体限制、业务上下文初始化、动态插件路由准备、登录鉴权、权限校验、运行时 freshness 检查和审计元数据构建都必须保持一致。路径迁移不能绕过插件启用状态、运行时修订号和派生缓存失效逻辑。

## 启动自动启用安装上下文与演示保护插件访问控制

普通插件管理安装与启动自动启用安装最终都会汇入插件 facade 的内部安装路径。如果不给源码插件生命周期暴露可信来源，上层插件无法区分“运维在 `plugin.autoEnable` 中显式要求启动安装”和“管理员从页面手动安装”。因此安装选项需要携带仅限启动自动启用路径设置的上下文标记，并通过 `BeforeInstall` 生命周期回调向源码插件暴露。

该上下文只对 `plugin.autoEnable` 明确列出的目标插件生效。自动依赖预安装不能继承该标记，否则会把“被自动启用插件依赖”误解释为“该插件本身允许只在启动阶段安装”。

`linapro-ops-demo-guard` 基于这一上下文实现安装前置策略：仅当当前安装来自启动自动启用引导时才允许继续；普通插件管理安装必须返回 veto 原因，阻断安装 SQL、菜单同步和 installed 状态写入。这样可以保留演示环境通过运维配置自动进入全局只读模式的能力，同时避免普通管理环境误触发全局写保护。

## 演示保护行为与工作台体验稳定性

`linapro-ops-demo-guard` 的运行时职责保持不变：插件启用后按 HTTP 方法语义阻断 `POST`、`PUT`、`DELETE` 等写请求，保留 `GET`、`HEAD`、`OPTIONS` 查询能力，并继续对白名单会话路径放行，包括登录、令牌刷新、租户选择、租户切换和退出。插件治理写操作同样属于必须阻断的写路径，但插件管理查询仍应保持可读。

多租户工作台头部租户切换器已经迁移到官方租户插件的头部插槽，因此关键布局约束必须跟随组件本身收敛，而不能继续依赖宿主布局或外部样式扫描。固定宽度、右侧间距、单行截断、暗色主题样式和建筑图标需要通过插件本地 `scoped CSS` 与显式 `IconifyIcon` 保持稳定，避免运行时装配边界变化导致回归。

演示账号与租户关系同样属于插件治理可观察体验的一部分。`tenant-user` 需要具备除平台管理外的其余菜单权限、使用本租户数据权限，并关联 5 个活跃租户，这样登录后既能演示多租户选择，也能稳定复现本租户数据隔离行为。相关 `E2E` 页面对象和断言围绕这些最终行为建立，而不是依赖早期布局或单租户假设。
