## 1. 实施准备

- [x] 1.1 实施前重新读取本变更命中的项目规则文件，并使用`goframe-v2`技能约束后端 Go 实现。
- [x] 1.2 静态检索`sys_kv_cache`、`NewSQLTableProvider`、`BackendSQLTable`、`RequiresExpiredCleanup`、`host:kvcache-cleanup-expired`和低层 JWT 解析调用点，确认代码、SQL、测试和文档影响面。
- [x] 1.3 记录本变更影响判断：属于`lina-core`宿主通用能力；无新增 HTTP API；无新增前端 UI；数据权限不新增读取或写操作接口；`i18n`仅受 KV cleanup 内置任务文案删除影响。

## 2. 单机 memory 后端

- [x] 2.1 在`apps/lina-core/internal/service/kvcache`实现`memory`backend/provider，当前实现基于 GoFrame`gcache`，支持`Get`、`GetInt`、`Set`、`Delete`、`Incr`、`Expire`和`CleanupExpired`空操作。
- [x] 2.2 保持`BuildCacheKey`、`OwnerType`、租户、owner key、`namespace`和 logical key 约束不变，并延续负 TTL、值类型错误、key/value 长度校验和写入失败返回错误语义。
- [x] 2.3 为`memory`后端补充自包含单元测试，覆盖 TTL 过期、缓存未命中、字符串值、整数递增、并发递增、递增字符串报错、负 TTL 和进程内有损缓存语义。
- [x] 2.4 将`kvcache.New()`默认 provider 调整为`memory`，同时保持生产路径通过启动期显式 provider 注入共享`kvcache.Service`。

## 3. 启动拓扑和共享实例

- [x] 3.1 修改 HTTP 启动期`newHTTPKVCacheProvider`：`cluster.enabled=false`选择`memory`provider，`cluster.enabled=true`继续选择 coordination KV provider。
- [x] 3.2 确认源码插件缓存 facade、动态插件 cache host service、WASM host service、auth pre-token/revoke 和宿主模块均复用启动期共享`kvCacheSvc`，不得在运行路径自行创建独立缓存服务图。
- [x] 3.3 补充启动期 provider 选择测试，覆盖单机`memory`、集群 coordination KV、集群缺失 coordination fail-fast、两个后端均不注册 KV cleanup job。
- [x] 3.4 在任务记录中写明 DI 来源检查：`memory`provider 的 owner、创建位置、传递路径、共享实例复用策略和集群分支不退回本地缓存的依据。

## 4. 认证和会话权威源

- [x] 4.1 调整单机 JWT revoke store，使单机仅使用进程内快速拒绝缓存；集群继续使用 Redis coordination KV，并保持读取失败 fail-closed。
- [x] 4.2 收敛完整认证入口，确保中间件、插件 route auth、refresh、tenant switch、logout、force logout 和 impersonation revoke 等路径在需要完整登录态时校验`sys_online_session`。
- [x] 4.3 新增更清晰的完整认证方法，并将低层 JWT 解析保持为内部实现，避免调用方把仅 JWT 解析结果当作完整登录态。
- [x] 4.4 补充认证单元测试，覆盖单机进程重启后 revoke 缓存丢失但 session 缺失仍拒绝、session 未超时继续有效、退出和强制下线删除 session 后拒绝、集群 revoke 读取失败 fail-closed。
- [x] 4.5 补充静态检索或测试证据，证明生产调用方只能通过完整认证入口获取 access-token claims，不暴露低层 JWT 解析方法。

## 5. 删除 SQL table cache 和数据库交付物

- [x] 5.1 删除`apps/lina-core/internal/service/kvcache/internal/sqltable`实现、`BackendSQLTable`常量、SQL table provider 和相关测试路径。
- [x] 5.2 在用户明确认可无兼容性负担后，直接从宿主 SQL 基线删除`sys_kv_cache`相关索引、表和交付注释依赖，不新增过渡清理 SQL 文件。
- [x] 5.3 运行`make db.init`和`make dao`，同步删除或更新`SysKvCache`相关 DAO/DO/Entity 生成物，不手工维护生成代码。
- [x] 5.4 更新或删除依赖`sys_kv_cache`DDL 的测试、mock、打包 manifest 和多进程测试断言，改为验证`memory`或 coordination KV 分支。
- [x] 5.5 在任务记录中说明 SQL 幂等性、数据分类、自增主键、软删除语义和索引影响：本变更删除缓存表，不新增业务查询表或数据权限索引。

## 6. 定时任务、i18n 和治理资源

- [x] 6.1 移除`host:kvcache-cleanup-expired`handler、默认内置作业投影、相关错误码和测试；保留`host:session-cleanup`。
- [x] 6.2 同步清理或调整宿主`manifest/i18n`、API 文档资源和打包资源中 KV cleanup 任务名称、描述或错误文案。
- [x] 6.3 确认没有新增前端页面、菜单、按钮、表格或路由；若仅删除后台内置任务文案，记录无新增 UI 文案影响。
- [x] 6.4 更新易失性表治理实现或测试中的自然过期清单，使其只包含`sys_online_session`和`sys_locker`，不得继续列出`sys_kv_cache`。

## 7. 验证和审查

- [x] 7.1 运行`openspec validate replace-single-node-kvcache-with-gcache --strict`并记录结果。
- [x] 7.2 运行后端单元测试：至少覆盖`kvcache`、`auth`、`session`、`middleware`、`plugin`cache host service、`cron`和`httpstartup`相关包。
- [x] 7.3 运行宿主启动绑定编译门禁，默认执行`cd apps/lina-core && go test ./internal/cmd -count=1`，或记录等价覆盖命令。
- [x] 7.4 运行静态检索，确认生产代码和交付 SQL 不再依赖`sys_kv_cache`、`NewSQLTableProvider`、`BackendSQLTable`或 SQL table backend。
- [x] 7.5 运行或记录`i18n`资源校验、SQL 静态检查、DAO 生成结果检查和相关打包资源检查。
- [x] 7.6 完成`lina-review`审查，确认缓存一致性、DI 来源、数据库迁移、数据权限、`i18n`、测试覆盖和集群模式不退回本地缓存均满足规则。

## Feedback

- [x] **FB-1**: 在无兼容性负担前提下清理旧 SQL 基线，不再通过新增 SQL 文件删除`sys_kv_cache`
- [x] **FB-2**: 收窄 auth 服务导出面，移除公开低层 JWT 解析入口以避免绕过`sys_online_session`完整校验
- [x] **FB-3**: 移除 kvcache 进程级可变默认 provider，保持启动期显式注入和固定本地默认
- [x] **FB-4**: 将单机 kvcache 公开后端命名和文件名从`gcache`收敛为`memory`
- [x] **FB-5**: 移除 memory 后端固定 LRU 容量，要求缓存写入、递增和过期更新必须使用正 TTL
- [x] **FB-6**: 将 kvcache 具体 backend 实现封装为`internal/memory`和`internal/coordkv`子组件

## 实施记录

- 规则读取：本轮实施和审查前已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/database.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/data-permission.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`、`.agents/rules/plugin.md`和`.agents/rules/dev-tooling.md`；后端实现使用`goframe-v2`技能约束。
- 影响判断：本变更属于`apps/lina-core`宿主通用缓存、认证和启动拓扑能力；未新增或修改 HTTP API、前端页面、菜单、按钮、表格或路由；未新增数据读取或写操作接口，数据权限无新增接入点；`i18n`影响仅为删除宿主内置 KV cleanup 任务文案。
- 缓存一致性：单机`cluster.enabled=false`由启动期创建`memory`provider，缓存权威源为调用方业务数据或`sys_online_session`等既有权威表，缓存本身为进程内有损状态；一致性模型为本进程内可见、TTL 原生过期、进程重启后未命中；最大可接受陈旧时间为调用方设置的 TTL；故障降级为未命中或返回结构化错误。`memory`后端要求写入、递增和过期更新使用正 TTL，不创建永不过期缓存条目。集群`cluster.enabled=true`继续使用 coordination KV backend，要求 coordination 服务存在，启动期缺失时 fail-fast，不退回本地缓存。
- DI 来源检查：`memory`provider 的 owner 是 HTTP 启动拓扑装配；创建位置为`apps/lina-core/internal/cmd/internal/httpstartup/http_runtime.go`中的`newHTTPKVCacheProvider`；传递路径为`newHTTPRuntime`创建`kvCacheSvc := kvcache.New(kvcache.WithProvider(kvCacheProvider))`后注入`auth.New`、`pluginsvc.NewHostServices`、源码插件 cache facade、动态插件 cache host service 和 WASM host service；生产路径复用启动期同一个`kvCacheSvc`，不在中间件、插件请求路径或 host service 调用路径自行创建独立服务图。
- 认证权威源：新增`AuthenticateAccessToken`作为完整登录态入口，低层 JWT 解析仅作为`auth`包内部实现，负责 JWT 签名、access kind、`clientType`和 revoke 快速状态。中间件、tenant switch bearer 入口和 impersonation revoke 均校验`sys_online_session`；tenant switch bearer 入口兼容`Authorization` header 值和裸 token，并有单测覆盖；refresh 继续通过`TouchOrValidate`校验 session；登出、强制下线和 tenant switch revoke 删除对应 session 后，单机进程重启导致 revoke 快速缓存丢失也不会放行缺失 session 的 token。
- 数据库和 DAO：已按用户反馈直接从`apps/lina-core/manifest/sql/010-runtime-host-services.sql`删除`sys_kv_cache`相关建表、索引和注释，不再新增`013`过渡清理 SQL；不含 Seed DML 或 Mock 数据，不写入自增主键，不涉及软删除字段，不新增查询表、数据权限索引或高频业务查询索引。已执行`make db.init confirm=init rebuild=true`和`make dao`；`make dao`不会自动删除缺失表的历史生成文件，因此手动删除旧`SysKvCache` DAO/DO/Entity 生成物，并保留生成器对`sys_online_session`产生的字段顺序更新。
- 定时任务和资源：已移除`host:kvcache-cleanup-expired`handler、默认内置作业投影、相关错误码和测试路径；保留`host:session-cleanup`。已清理`apps/lina-core/manifest/i18n/en-US/job.json`和`apps/lina-core/manifest/i18n/zh-CN/job.json`中的 KV cleanup 文案。已执行`make pack.assets`，packed manifest 未检出旧`sys_kv_cache`、`kvcache-cleanup-expired`、`NewSQLTableProvider`或`BackendSQLTable`引用。
- 静态检索：`rg -n "NewSQLTableProvider|BackendSQLTable|internal/sqltable|CodeCronKVCacheDependencyMissing" apps/lina-core`无结果；`rg -n "sys_kv_cache" apps/lina-core`无结果；`rg -n "host:kvcache-cleanup-expired|kvcache-cleanup-expired" apps/lina-core`仅命中负向测试断言；生产代码不再暴露或调用公开低层 JWT 解析方法，该解析仅保留为`auth`包内部实现。
- 反馈修复影响分析：FB-1 修改宿主 SQL 基线和 packed manifest，不新增 Seed DML、Mock 数据、自增主键写入、软删除字段或高频查询索引；FB-2 仅收窄`auth.Service`导出面，生产完整认证继续通过`AuthenticateAccessToken`校验`sys_online_session`；FB-3 移除 kvcache 进程级可变默认 provider，`kvcache.New()`固定本地`memory`兜底，生产路径仍由 HTTP 启动期显式注入 provider。三项反馈均无新增 HTTP API、前端 UI、插件目录文件、数据权限读写接口或开发工具脚本；`i18n`资源无新增影响。
- FB-4 影响分析：本次反馈修正单机 kvcache 公开后端命名和文件名，`BackendName`从`gcache`收敛为`memory`，构造入口从`NewGCacheProvider()`收敛为`NewMemoryProvider()`，源码文件改为`kvcache_memory_backend.go`和`kvcache_memory_backend_test.go`；GoFrame`gcache`仅作为内部实现依赖保留。该反馈不新增 HTTP API、SQL、前端 UI、数据权限读写接口、`i18n`资源或开发工具脚本；缓存一致性语义不变，仍为单机进程内有损缓存、TTL 原生过期、集群不退回本地缓存。
- FB-4 验证结果：`cd apps/lina-core && go test ./internal/service/kvcache ./internal/cmd/internal/httpstartup -count=1`通过；`cd apps/lina-core && go test ./internal/service/kvcache ./internal/service/auth ./internal/service/session ./internal/service/middleware ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./internal/service/cron ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`通过；`openspec validate replace-single-node-kvcache-with-gcache --strict`通过；`git diff --check`通过；静态检索确认生产代码中`BackendGCache`、`NewGCacheProvider`、旧`gCache`内部类型、旧`kvcache_gcache_*`文件名和`"gcache"`后端值均无残留，`gcache`仅作为 GoFrame 内部实现依赖或 FB-4 历史说明保留。
- FB-5 影响分析：本次反馈移除`memory`后端固定 LRU 容量，`NewMemoryProvider().NewBackend()`改为使用无容量参数的`gcache.New()`，并将`kvcache.Service`、`memory`后端、coordination KV 后端、源码插件 cache facade 和 WASM cache host service 的写入、递增、过期更新语义收紧为必须传入正 TTL。新增`KV_CACHE_EXPIRE_SECONDS_REQUIRED`结构化错误和中英文`i18n`资源；同步更新`apps/lina-core/pkg/plugin`README、源码插件 cache capability 注释、WASM host service 协议注释和 OpenSpec 规范。该反馈不新增 HTTP API、SQL、前端 UI、数据权限读写接口或开发工具脚本；缓存一致性影响为强化生命周期约束，权威数据源、单机进程内有损缓存、集群 coordination KV 和启动期共享实例策略不变。数据权限无新增读写边界；DI 无新增运行期依赖；开发工具跨平台无影响。
- FB-5 验证结果：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm ./internal/service/kvcache -count=1`通过；`cd apps/lina-core && go test ./internal/service/kvcache ./internal/service/auth ./internal/service/session ./internal/service/middleware ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./internal/service/cron ./internal/cmd/internal/httpstartup ./internal/cmd ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`通过；`openspec validate replace-single-node-kvcache-with-gcache --strict`通过；`git diff --check`通过；`make i18n.check`通过，仅报告既有模块级`$t()`警告；`make pack.assets`通过且未产生额外 packed 文件差异；静态检索确认`defaultMemoryLRUCap`、`lruCap`、带容量参数的`gcache.New(...)`、`ttl=0`旧语义说明、`no expiration`、`never expire`和固定 LRU 旧语义在本变更范围无残留，生产代码未发现`kvcache.Set/Incr/Expire(..., 0)`调用。
- FB-6 影响分析：本次反馈将`apps/lina-core/internal/service/kvcache`中的具体后端实现拆分到`internal/memory`和`internal/coordkv`，并新增`internal/contract`承载避免 Go import cycle 的窄共享契约；根包继续通过类型别名和`NewMemoryProvider`、`NewCoordinationKVProvider`facade 暴露既有调用面，key 解析、TTL 校验和后端 payload 转换收敛到内部子组件。该反馈不改变缓存权威源、一致性模型、TTL、错误码、启动期 provider 选择、HTTP API、SQL、前端 UI、数据权限读写接口、`i18n`资源或开发工具脚本；缓存一致性语义不变，单机仍为进程内有损缓存，集群仍使用 coordination KV 且不退回本地缓存。DI 无新增运行期依赖，测试迁移为内部子包外部测试，通过公开 kvcache 契约验证两个后端。
- FB-6 验证结果：`cd apps/lina-core && go test ./internal/service/kvcache/... -count=1`通过；`cd apps/lina-core && go test ./internal/service/kvcache/... ./internal/service/auth ./internal/service/session ./internal/service/middleware ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./internal/service/cron ./internal/cmd/internal/httpstartup ./internal/cmd ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`通过；`openspec validate replace-single-node-kvcache-with-gcache --strict`通过；`git diff --check`通过；静态检索确认`internal/contract`、`internal/memory`和`internal/coordkv`只在`kvcache`组件授权边界内使用，根包不再保留具体 backend 实现。
- 验证结果：`make db.init confirm=init rebuild=true`通过，fresh init 仅执行`001`至`012`SQL；`make dao`通过且不再生成`SysKvCache`相关文件；`make pack.assets`通过并同步 packed SQL；`rg -n "sys_kv_cache|013-replace-single-node-kvcache-with-gcache|NewSQLTableProvider|BackendSQLTable|internal/sqltable|DefaultProvider|SetDefaultProvider|processDefaultProvider|func \\(s \\*serviceImpl\\) ParseToken|ParseToken\\(" apps/lina-core`无结果；`openspec validate replace-single-node-kvcache-with-gcache --strict`通过；`git diff --check`通过；`cd apps/lina-core && go test ./internal/service/kvcache ./internal/service/auth ./internal/service/session ./internal/service/middleware ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./internal/service/cron ./internal/cmd/internal/httpstartup ./internal/cmd ./pkg/dialect -count=1`通过；`make i18n.check`通过，仅报告既有模块级`$t()`警告。
