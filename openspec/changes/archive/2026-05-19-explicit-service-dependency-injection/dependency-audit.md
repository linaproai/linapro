## 依赖构造基线审计

本文件记录 `explicit-service-dependency-injection` 迭代开始时发现的隐式服务构造、缓存一致性风险和允许局部构造候选。它是迁移任务的事实基线，不表示这些调用已经合规。

## 1. 隐式构造分类清单

### 宿主 Controller

| 分类 | 文件 | 当前构造点 | 风险 |
|------|------|------------|------|
| 认证控制器 | `apps/lina-core/internal/controller/auth/auth_new.go` | `hostconfig.New()`、`pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`authsvc.NewWithConfig(...)`、`bizctx.New()` | 登录、token、插件登录 Hook、租户能力与 HTTP runtime 中的共享插件/配置实例可能不一致 |
| 用户控制器 | `apps/lina-core/internal/controller/user/user_new.go` | `pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`usersvc.New(...)`、`role.New(pluginSvc)`、`menu.New(pluginSvc)`、`i18n.New()` | 用户管理、角色权限、组织能力、i18n 和数据权限实例分裂 |
| 菜单控制器 | `apps/lina-core/internal/controller/menu/menu_new.go` | `pluginsvc.New(nil)`、`menu.New(pluginSvc)`、`role.New(pluginSvc)`、`bizctx.New()` | 菜单/权限过滤和插件 enabled snapshot 可能与插件 runtime 不一致 |
| 角色控制器 | `apps/lina-core/internal/controller/role/role_new.go` | `pluginsvc.New(nil)`、`role.New(pluginSvc)` | token access cache、插件权限过滤和数据权限缓存可能分裂 |
| 插件控制器 | `apps/lina-core/internal/controller/plugin/plugin_new.go` | `pluginsvc.New(topology)`、`bizctx.New()`、`role.New(pluginSvc)` | 插件管理控制器可能持有不同 plugin service，导致 enabled snapshot、runtime cache、source route bindings 不一致 |
| 文件控制器 | `apps/lina-core/internal/controller/file/file_new.go` | `pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`filesvc.New(...)` | 文件数据权限、组织能力和插件状态读取可能分裂 |
| i18n/config/dict/publicconfig/usermsg 控制器 | `apps/lina-core/internal/controller/**/_new.go` | 各自无参 `NewV1()` 内部构造 service | 运行时配置、i18n bundle、字典和用户消息相关缓存无法统一确认实例来源 |
| 部分已注入控制器 | `health/job/joblog/jobgroup/jobhandler/sysinfo` | 已接收部分依赖 | 风格不统一，仍需统一纳入显式依赖注入审查 |

### 宿主 Service 与 Middleware

| 分类 | 文件 | 当前构造点 | 风险 |
|------|------|------------|------|
| 认证服务 | `apps/lina-core/internal/service/auth/auth.go` | `config.New()`、`pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`role.New(pluginSvc)`、`tenantcap.New(pluginSvc)`、`session.NewDBStore()` | JWT、session hot state、revoke/pre-token、租户选择、权限预热与插件登录 Hook 可能使用不同事实源 |
| 认证 token store | `apps/lina-core/internal/service/auth/auth_revoke.go`、`auth_pre_token.go` | `kvcache.New()` | 集群模式 token state 必须使用 coordination KV，局部默认实例可能退回本地/SQL 默认 |
| Middleware | `apps/lina-core/internal/service/middleware/middleware.go` | `config.New()`、`pluginsvc.New(nil)`、`auth.New(nil)`、`bizctx.New()`、`i18n.New()`、`role.New(pluginSvc)`、`tenantcap.New(pluginSvc)` | 认证、权限、租户、请求 locale、session store 和插件 enabled snapshot 的最高风险路径 |
| 角色服务 | `apps/lina-core/internal/service/role/role.go`、`role_data_scope.go` | `bizctx.New()`、`config.New()`、`datascope.New(...)` | token access cache、数据权限和配置缓存可能分裂 |
| 数据权限服务 | `apps/lina-core/internal/service/datascope/datascope.go` | `bizctx.New()`、`orgcap.New(nil)` | 数据权限 fallback 可能与当前请求上下文和组织能力实例不一致 |
| 用户/文件/任务服务 | `apps/lina-core/internal/service/user/**`、`file/**`、`jobmgmt/**` | `auth.New(...)`、`bizctx.New()`、`role.New(nil)`、`orgcap.New(nil)`、`datascope.New(...)` | 数据权限、组织能力、会话/权限缓存和审计上下文实例来源不透明 |
| i18n/config/sysinfo/usermsg/notify | `apps/lina-core/internal/service/**` | `config.New()`、`bizctx.New()`、`pluginstate.New()` 等 | runtime bundle、配置快照、用户通知与插件状态缓存实例分裂 |
| Host lock | `apps/lina-core/internal/service/hostlock/hostlock.go` | `locker.New()` | 集群模式下锁必须走共享 coordination lock，本地默认构造风险高 |
| Cron | `apps/lina-core/internal/service/cron/cron.go` | `config.New()`、`kvcache.New()`、`pluginsvc.New(clusterSvc)`、`role.New(pluginSvc)` | 定时任务、KV cleanup、插件 lifecycle watcher 和权限缓存实例来源不透明 |
| 插件根服务 | `apps/lina-core/internal/service/plugin/plugin.go` | 内部构造 config、bizctx、catalog、runtime、integration、frontend、openapi、i18n | 插件子服务已集中在 plugin service 内，但外部多次 `pluginsvc.New` 会生成多套 plugin runtime state |

### `pkg/pluginservice/*`

| 分类 | 文件 | 当前构造点 | 风险 |
|------|------|------------|------|
| 插件 auth adapter | `apps/lina-core/pkg/pluginservice/auth/auth.go` | `pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`internalauth.NewTenantTokenIssuer(...)` | 源码插件租户选择和切换 token 时可能使用独立 auth/plugin/orgcap 实例 |
| 插件 session adapter | `apps/lina-core/pkg/pluginservice/session/session.go` | `pluginsvc.New(nil)`、`orgcap.New(pluginSvc)`、`internalauth.New(...)`、`role.New(pluginSvc)`、`datascope.New(...)`、`tenantcap.New(pluginSvc)` | 在线用户列表/强退的数据权限、租户可见性、revoke 和 session store 可能分裂 |
| 插件 notify adapter | `apps/lina-core/pkg/pluginservice/notify/notify.go` | `internalnotify.New(pluginstate.New())` | 通知投递依赖的插件状态可能与 runtime enabled snapshot 不一致 |
| 插件 i18n/config/bizctx/pluginstate/apidoc/route | `apps/lina-core/pkg/pluginservice/**` | 各自构造内部服务 | 源码插件面向宿主能力调用时缺少统一宿主发布服务目录 |
| tenantfilter | `apps/lina-core/pkg/pluginservice/tenantfilter/tenantfilter.go` | `bizctx.New().Current(ctx)` | 热路径上隐式构造 bizctx adapter，需改为上下文读取或注入 |

### 源码插件 Controller / Service / `backend/plugin.go`

| 分类 | 文件 | 当前构造点 | 风险 |
|------|------|------------|------|
| 路由注册回调 | `apps/lina-plugins/*/backend/plugin.go` | `controller.NewV1()`、`NewControllerV1()`、`middlewaresvc.New()`、`monitorsvc.New()` | 插件 route 注册不接收宿主发布依赖，控制器和服务自行生成依赖图 |
| content-notice | `apps/lina-plugins/content-notice/backend/internal/service/notice/notice.go` | `bizctx.New()`、`notify.New()` | 通知和业务上下文 host adapter 可能孤立 |
| multi-tenant | `apps/lina-plugins/multi-tenant/backend/internal/service/**` | `bizctx.New()`、`config.New()`、`tenantplugin.New()`、`resolverconfig.New()` | 租户解析、配置读取、租户插件状态和 impersonation token 逻辑需要共享宿主上下文 |
| org-center | `apps/lina-plugins/org-center/backend/internal/controller/**` | `deptsvc.New()`、`postsvc.New()` | 当前多为插件自有无状态服务，但仍需通过显式依赖证明无宿主缓存影响 |
| monitor 系列 | `monitor-online`、`monitor-loginlog`、`monitor-operlog`、`monitor-server` | 控制器和服务无参构造 | 在线用户、登录/操作日志、服务器监控需区分插件自有纯查询和宿主 session/config 依赖 |
| demo 插件 | `demo-control`、`plugin-demo-source`、`plugin-demo-dynamic` | 控制器、service 和 middleware 无参构造 | 示例代码会被复制，必须展示新的显式依赖风格 |

### WASM host service

| 分类 | 文件 | 当前构造点 | 风险 |
|------|------|------------|------|
| Cache host service | `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_cache.go` | 包级 `kvcache.New()`；`ConfigureCacheHostService(nil)` 恢复默认 | 集群模式可能绕过 coordination KV，动态插件 cache 与宿主 runtime 不一致 |
| Lock host service | `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_lock.go` | 包级 `hostlock.New()`；`ConfigureLockHostService(nil)` 恢复默认 | 集群模式可能绕过 coordination lock，插件锁跨节点失效 |
| Notify host service | `apps/lina-core/internal/service/plugin/internal/wasm/hostfn_service_notify.go` | 包级 `notifysvc.New()` | 动态插件 notify 可能读取孤立 plugin state |
| Storage/config host service | `hostfn_service_storage.go`、`hostfn_service_config.go` | 包级 `config.New()`、局部 `configsvc.New()` | 上传限制、存储配置、运行时配置读取可能与 HTTP runtime 不一致 |

## 2. 高风险缓存一致性路径

| 路径 | 权威数据源 | 必须共享的实例/后端 | 当前隐式构造点 | 可能不一致结果 | 故障策略 | 迁移优先级 |
|------|------------|--------------------|----------------|----------------|----------|------------|
| auth/session | PostgreSQL `sys_online_session` 管理投影、coordination KV session hot state、coordination KV token state | `auth.Service`、`session.Store`、`kvcache/coordination` backend | `auth.New`、`middleware.New`、`pluginservice/session.New` | logout/强退只影响部分实例；Redis fail-closed 被本地默认绕过 | token/session state 读取失败 fail-closed | P0 |
| role/datascope | PostgreSQL role/menu/user-role 权威表、`permission-access` revision | `role.Service`、`datascope.Service`、`cachecoord.Service` | Controller、user/file/jobmgmt data scope helper、pluginservice/session | 权限缓存未失效、数据权限范围不一致 | 权限 freshness 不可确认时 fail-closed | P0 |
| runtime config | PostgreSQL `sys_config`、`runtime-config` revision | `config.Service`、`cachecoord/coordination` backend | `auth.NewWithConfig` 以外路径、middleware、cron、i18n/sysinfo | session timeout、JWT expire、黑名单读取不一致 | 配置不可确认时返回结构化错误 | P0 |
| plugin runtime | PostgreSQL plugin registry/release/state、`plugin-runtime` revision/event | `plugin.Service`、runtime cache、integration shared state | Controller、auth、role、menu、pluginservice、middleware 多处 `pluginsvc.New` | disabled 插件路由/菜单/权限仍可见；source route bindings 分裂 | conservative-hide 或启动失败 | P0 |
| i18n runtime bundle | manifest/DB/content 权威资源、i18n cachecoord revision | `i18n.Service`、runtime bundle cache | middleware、controller、plugin、service 多处 `i18n.New` | locale 或错误文案缓存更新不一致 | 缓存不可确认时按现有 i18n fallback 或错误策略 | P1 |
| kvcache/locker/cachecoord | coordination Redis 或单机 SQL/local 后端 | `coordination.Service`、`kvcache.Provider`、`locker.Service`、`cachecoord.Service` | auth token store、cron、hostlock、WASM host service | 集群模式退化为本地或 SQL 默认，跨节点失效失败 | 写失败返回错误，安全路径 fail-closed | P0 |
| notify | PostgreSQL notify tables、插件状态、租户上下文 | `notify.Service`、`pluginstate` adapter、`bizctx.Service` | content-notice、pluginservice/notify、WASM notify | 已禁用插件状态或租户上下文读取不一致 | 写失败返回结构化错误 | P1 |
| source route bindings | plugin manifest + source plugin registration + enabled snapshot | `plugin.Service.integrationSvc` shared state | 每次 `pluginsvc.New` 生成新 integration shared state | OpenAPI、诊断、插件路由列表与实际绑定不一致 | 启动期失败或 conservative-hide | P0 |

## 3. 允许局部构造候选

以下候选需要在实现阶段逐项确认。允许局部构造的前提是：不持有缓存、订阅、session/token、插件状态、运行时配置快照、跨实例协调状态，不参与用户可见错误本地化，也不访问宿主发布服务。

| 候选 | 位置 | 允许理由 | 后续要求 |
|------|------|----------|----------|
| 插件自有纯 DAO 查询服务 | `org-center` 的 `dept.New()`、`post.New()` 等 | 当前服务实现主要包装插件自有 DAO 查询和事务逻辑，不持有宿主缓存状态 | 构造函数仍应显式接收未来新增的宿主能力依赖；扫描 allowlist 需写明无状态原因 |
| DTO 转换 helper 和导出 helper | Controller 同包私有函数、Excel helper | 纯函数，无运行期状态 | 不纳入 service `New()` allowlist |
| GoFrame DAO 构造函数 | `internal/dao/internal/*` | 生成代码，由 `gf gen dao` 管理 | 静态扫描排除 DAO 生成目录 |
| 测试 fake / fixture 构造 | `*_test.go`、`internal/service/plugin/internal/testutil` | 测试需要独立 fake 依赖 | 测试必须自包含并恢复全局状态 |
| 插件 demo 的纯示例业务服务 | `plugin-demo-source`、`plugin-demo-dynamic` 的无宿主能力 service | 可作为过渡候选 | 示例最终应展示显式依赖风格，避免被复制出隐式模式 |

## 4. 首批迁移建议

1. 先改 `auth.Service`、`middleware.Service` 和 `cmd_http_runtime.go`，因为它们影响认证、权限、session 和配置热路径。
2. 随后改宿主 `auth/user/role/menu/plugin/file/i18n/config` Controller，使路由绑定持有共享实例。
3. 再改 `pkg/pluginservice/*` 和 `pluginhost` registrar，给源码插件提供宿主发布服务目录。
4. 最后迁移源码插件控制器/服务和 WASM host service 包级默认实例。
