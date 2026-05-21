## Why

官方插件生态已经同时覆盖源码插件、动态插件、插件治理、运行时 `i18n`、菜单权限、`cron`、动态资源和扩展路由。缺少统一前缀与稳定命名边界的插件 ID 会直接影响 URL、文件名、数据库键、菜单 key、权限字符串、`i18n` namespace、`apidoc` namespace 和宿主能力发现，官方插件与第三方插件之间的身份边界也不够清晰。

动态插件公开数据面长期挂在 `/api/v1/extensions/{pluginId}/...` 下，使宿主控制面 API 版本与插件自有 API 版本被迫耦合。插件作者无法自然表达 `/api/v2/...`、`/graphql` 等自身路径语义，OpenAPI、资源列表和示例也持续传播了错误的路径心智模型。

`linapro-ops-demo-guard` 用于演示环境的全局只读保护，一旦在普通管理环境中被页面误安装并启用，会直接阻断系统写操作和插件治理写操作。该能力应由运维通过 `plugin.autoEnable` 在宿主启动阶段显式引导，而不是暴露为常规页面安装入口。

与此同时，多租户工作台头部租户切换器迁移到插件插槽后，样式稳定性和演示数据完整性也需要补强：切换器应保持固定宽度、图标、截断和间距约束，演示账号需要具备可复现的数据隔离与多租户切换场景。

## What Changes

- 建立插件 ID 的基础安全契约，只在运行时强制校验非空、64 字符长度上限和 lowercase kebab-case，并将 `<author>-<domain>-<capability>` 保留为官方命名建议与仓库治理约定。
- 对官方插件实施统一规范化映射，全面更新目录名、`plugin.yaml`、源码注册、依赖声明、菜单、权限、`cron`、动态资源、`i18n`、`apidoc`、配置、SQL、测试和文档中的官方插件 ID；不保留旧 ID alias 或兼容查询。
- 将动态插件 canonical 公开路径切换为 `/x/{pluginId}/...`，移除旧 `/api/v1/extensions/{pluginId}/...` 分发入口，并保持宿主统一 HTTP 治理链路、动态插件鉴权、权限校验和运行时 freshness 检查不变。
- 为源码插件 `BeforeInstall` 生命周期暴露可信的启动自动启用安装上下文，使 `linapro-ops-demo-guard` 只能通过 `plugin.autoEnable` 在启动期间安装和启用，普通插件管理安装必须被拒绝，同时保留其只读守卫、最小会话白名单和插件治理写操作拦截语义。
- 稳定多租户工作台头部租户切换器的本地样式与图标装配，扩展 `E2E` 校验，并补齐 `tenant-user` 演示账号、租户关联和本租户权限场景，保证管理工作台演示链路完整可复现。

## Capabilities

### New Capabilities

- `plugin-id-governance`：定义插件 ID 运行时安全边界、官方插件规范化映射、运行时身份一致性和治理扫描要求。
- `plugin-startup-bootstrap`：为启动自动启用安装提供可信生命周期上下文，使源码插件可以区分启动引导安装与普通管理安装。

### Modified Capabilities

- `plugin-manifest-lifecycle`：将 manifest ID、源码注册 ID、动态 artifact ID、依赖声明和资源命名空间统一收敛到当前插件 ID 契约。
- `plugin-runtime-loading`：将动态插件公开数据面入口迁移到 `/x/{pluginId}/...`，允许插件在自身路径中定义版本和协议语义。
- `demo-control-guard`：将官方演示只读保护插件稳定为 `linapro-ops-demo-guard`，仅允许通过 `plugin.autoEnable` 启动自动启用安装，并继续阻断系统与插件治理写操作。
- `dashboard-workbench`：稳定多租户工作台头部租户切换器样式，补齐演示账号与多租户切换场景的验证覆盖。

## Impact

- **Breaking**：官方旧插件 ID 不再作为有效运行时身份；旧 `/api/v1/extensions/{pluginId}/...` 不再分发动态插件请求；`demo-control` 不再作为官方插件 ID 暴露，`linapro-ops-demo-guard` 也不得通过普通插件管理安装。
- 影响宿主插件治理、启动自动启用、动态插件 runtime 路由绑定、OpenAPI 投影、插件资源列表、`apidoc` 路径识别和官方插件能力常量。
- 影响官方插件目录、Go module/import、`plugin.yaml`、依赖声明、自有 SQL/DAO 产物、动态资源路径、菜单 key、权限字符串、`cron handlerRef`、运行时 `i18n`/`apidoc` namespace、默认配置和测试基线。
- 影响 `linapro-ops-demo-guard` 生命周期前置条件、运行时错误文案、本地化资源和相关单元测试；影响多租户工作台头部插槽组件、演示账号 mock 数据和既有 `E2E` 页面对象。
- 不引入新的业务数据访问接口或额外缓存类型；插件状态、菜单、路由、`cron`、`i18n` 和 `apidoc` 刷新继续按插件 ID scope 精确失效，数据权限边界保持既有治理策略。
