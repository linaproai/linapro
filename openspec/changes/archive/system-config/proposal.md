## Why

系统配置分组需要同时解决三类紧密相关的问题。

1. 宿主已经依赖 JWT 有效期、在线会话超时、上传大小、登录 IP 黑名单、登录页品牌和工作台主题等运行时配置，但配置管理此前缺少受保护的宿主参数注册表、统一校验、公共前端白名单出口，以及适合单机/集群场景的高频读取策略。与此同时，`sys.upload.maxSize` 的数据库初始值、配置模板与后端静态兜底值不一致，导致不同启动路径下默认行为分裂。

2. 系统 API 文档与插件治理缺少统一的宿主控制面。直接依赖 GoFrame 默认 `/api.json` 无法按插件启用状态投影接口，也无法可靠区分宿主路由与源码插件路由；如果把路由清单写回 `plugin.yaml`，又会引入重复声明。插件详情展示、日志治理、配置命名空间和归档规范也需要同步收敛。

3. 启动链路与登录后首页存在明显的重复 SQL。启动阶段会重复读取插件注册表、发布快照、菜单、资源引用和内置任务投影，清单无差异时仍可能进入空事务或写后回读；默认 SQL debug 还会放大日志噪音。登录后首页的并发请求则会对 `sys_online_session` 与 `sys_plugin_release` 产生重复读取，增加鉴权与插件运行时投影的数据库压力。

此外，`config-management` 组件单元测试覆盖率低于仓库门槛，缺少对配置快照、集群修订号、公共前端配置和插件路径等高风险分支的稳定回归保护。

## What Changes

- 在宿主配置层注册受保护的运行时参数和公共前端配置元数据，为 JWT、在线会话、上传限制、登录 IP 黑名单、登录页品牌和工作台主题提供统一的默认值、校验规则、导入保护和运行时读取入口。
- 让认证、在线会话、文件上传和前端启动真正消费运行时配置，并通过本地快照加共享修订号的方式降低高频读取开销，同时保留单机和集群场景下的差异化失效策略。
- 统一 `sys.upload.maxSize` 的数据库种子值、配置模板默认值和后端静态兜底值到 20 MB，并同步更新上传校验、友好错误提示与相关测试。
- 以宿主管理的 OpenAPI 构建器替代 GoFrame 默认 `/api.json`，在不增加 `plugin.yaml` 路由重复声明的前提下投影启用中的源码插件和动态插件接口，并排除内部或非业务路由。
- 在源码插件路由注册时捕获插件归属、HTTP 方法、路径和 DTO `g.Meta` 文档元数据，保持源码插件中间件编排仍由插件自行控制。
- 完善默认管理工作台中的插件详情弹窗、宿主服务授权展示、资源分组、空态表现和动态插件示例分页，同时补齐结构化日志开关、`extensions` 配置命名空间和相关后端注释治理。
- 引入一次 HTTP 启动编排内共享的 `StartupContext`，复用插件 catalog、菜单/资源引用和内置任务投影快照，减少启动阶段的重复查询。
- 将插件清单同步改为差异驱动的 no-op fast path：无差异时不写库、不启事务、不做写后回读；有差异时直接回写并同步更新启动快照。
- 优化内置定时任务启动注册，使用声明派生的投影快照直接注册运行时调度器，避免启动时再次从持久化表重复扫描同一批内置任务。
- 为启动阶段输出结构化摘要日志，并补充默认无 SQL 明细、启动快照复用、插件 no-op 同步和启动 smoke 回归验证。
- 优化登录后首页的在线会话校验与插件 release 读取复用，减少有效鉴权请求和插件运行时列表投影中的重复 SQL。
- 为配置管理、会话校验、插件 catalog 读取复用和启动编排补充单元测试与 smoke 验证，使 `apps/lina-core/internal/service/config` 的包级覆盖率达到交付门槛。

## Capabilities

### New Capabilities

- `startup-sql-efficiency`
- `login-home-sql-efficiency`

### Modified Capabilities

- `config-management`
- `online-user`
- `user-auth`
- `plugin-manifest-lifecycle`
- `plugin-runtime-loading`
- `plugin-startup-bootstrap`
- `plugin-ui-integration`
- `cron-job-management`
- `system-api-docs`
- `spec-governance`

## Impact

- 受影响后端模块包括 `apps/lina-core/internal/service/config`、`auth`、`session`、`file`、`cron`、`jobmgmt`、`plugin`、`apidoc` 及 `internal/cmd` 启动编排。
- 受影响插件接缝与配套组件包括 `apps/lina-core/pkg/pluginhost`、插件桥接层、插件数据库支持包和默认管理工作台中的插件治理界面。
- 受影响交付资源包括宿主初始化 SQL、配置模板、上传配置兜底值、结构化日志与扩展命名空间配置。
- 受影响验证范围包括配置管理覆盖率门禁、启动 smoke、会话校验测试、插件 catalog 复用测试、OpenSpec 审查和归档规范治理。
- 本次整合不引入新的公开 HTTP API 契约、数据库 schema 或运行时 i18n/apidoc 资源；缓存新增仅限启动链路或请求作用域内的短生命周期复用。
