## Context

LinaPro 已经有一条清晰的 HTTP 启动编排路径：`cmd.Http` 创建 `configSvc`，`newHTTPRuntime` 构造长生命周期服务，`bindHostAPIRoutes` 绑定宿主控制器，源码插件通过 `pluginhost.HTTPRegistrar`、`CronRegistrar`、Hook 和 host service 与宿主协作。

问题在于这条启动编排还没有成为完整的运行期共享服务图。大量 Controller、Middleware、Service、`pkg/pluginservice/*` 适配器和源码插件控制器仍在自身 `New()` 或 `NewV1()` 中再次调用其他 `service.New()`，导致缓存敏感路径上的实例不一致。

API 响应直接嵌入数据库实体，如用户响应嵌入 `entity.SysUser`，文件响应嵌入 `entity.SysFile`，可能暴露密码字段、软删除字段、存储路径等实现细节。`apps/lina-core/api` 中存在多处语义相同但分散定义的枚举和值类型。

仓库同时维护 PostgreSQL 与 SQLite 两条运行路径，`pkg/dbdriver` 注册 PG/SQLite 驱动，`pkg/dialect` 按链接分发方言，初始化命令通过转译器把 PostgreSQL 源 SQL 转译到 SQLite，启动钩子在 SQLite 下强制关闭集群模式。

## Goals / Non-Goals

**Goals:**

- 后端生产代码中 Controller、Middleware、Service、插件适配器和 WASM host service 的接口型依赖通过构造函数参数逐项显式传入，禁止使用聚合依赖结构体整体传递。
- 消除缓存敏感路径中的孤立实例，确保认证、权限、租户、session、插件运行时、运行时配置、i18n、notify、cachecoord 和插件管理复用启动期同一套实例。
- API 响应使用独立 DTO，不直接嵌入数据库实体。
- 公共枚举契约按稳定边界提取到 `pkg/` 小组件。
- 运行时数据库支持收敛到 PostgreSQL 14+，移除 SQLite 支持。
- 将规则写入项目规范、OpenSpec 规范和 lina-review 审查标准。

**Non-Goals:**

- 不引入 `wire`、`dig`、`fx` 或其他通用 DI 框架。
- 不新增单独的宿主私有组装层或全局 service locator。
- 不引入 MySQL、Redis-only 或其他数据库支持。
- 不改变业务 API 语义、权限模型或前端 UI。
- 不新增 `pkg/enums` 或其他大一统枚举包。

## Decisions

### 1. 显式依赖注入是唯一生产构造规则

生产后端组件的构造函数必须逐项接收接口型依赖。接口型依赖一律拆分为独立构造函数参数，即使依赖数量超过 2 个也不得改用聚合结构体承载。纯值配置可以使用专门配置结构体，但不得混入接口型运行期依赖。

### 2. 复用现有启动编排作为构造边界

宿主构造边界保持在 `cmd_http_runtime.go`、`cmd_http_routes.go`、`pluginhost` registrar 回调和 WASM host service 注册位置。不新增独立应用组装包或全局 registry。

### 3. 高风险缓存一致性路径必须共享实例

auth、session、role、plugin、config、i18n、cachecoord、kvcache、locker、notify、pkg/pluginservice/* 适配器和 WASM host service 等持有缓存或运行期状态的组件必须复用启动期共享实例。

### 4. 源码插件通过 registrar 获取宿主发布依赖

`pluginhost.HTTPRegistrar` 暴露稳定的宿主依赖目录 `HostServices`，源码插件路由注册从该目录获取 bizctx、config、i18n、notify、auth、session、pluginstate 等宿主能力。

### 5. 初始化与注册 API 返回 error

宿主与源码插件的运行时初始化、注册 API、回调注册、路由/Cron/中间件注册 API 在依赖缺失或校验失败时必须返回 error，由调用方显式处理。是否中止进程必须由调用栈最上层入口决定。

### 6. API 响应使用独立 DTO

控制器在响应边界显式映射允许暴露的字段，避免把实体指针直接传入 API 响应。用户响应不包含 password 等内部字段。

### 7. 公共枚举契约按稳定边界提取

`pkg/listorder` 承载排序方向，`pkg/tenantoverride` 承载租户覆盖模式，`pkg/statusflag` 承载通用 0/1 状态标志。菜单类型复用 `pkg/menutype`，插件桥接常量复用 `pkg/pluginbridge`。领域私有枚举留在所属领域包内。

### 8. SQLite 链接快速失败

`sqlite:` 链接在 `dialect.From` 和驱动注册边界明确拒绝，不保留只读兼容层。保留 `pkg/dialect` 抽象但只保留 PostgreSQL 具体实现。

### 9. 分阶段迁移

按风险优先级推进：规范和审查标准先落地 → 迁移 middleware、auth、role 等高风险宿主服务 → 迁移宿主 Controller → 迁移 pkg/pluginservice/* 和 source plugin registrar → 迁移源码插件 → 迁移 WASM host service → 收紧静态扫描。

## Risks / Trade-offs

- 构造函数签名大面积变化 → 分阶段迁移，每阶段只覆盖一组依赖链。
- 循环依赖暴露出来 → 使用窄接口和职责拆分降低耦合。
- 测试构造成本上升 → 提供测试 helper 和 fake dependencies，但保持生产路径显式。
- 插件注册回调接口扩展影响所有源码插件 → 先在 registrar 中新增能力，再逐个插件迁移。
- 现有本地开发者使用 SQLite 文件启动会失败 → 配置模板、README 和错误信息统一指向 PostgreSQL。
- Go module 依赖可能仍通过间接路径保留 SQLite 包 → 运行 `go mod tidy` 并扫描残留。
