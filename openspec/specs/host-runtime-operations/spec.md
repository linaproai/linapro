# 宿主运行时操作

## Purpose

定义宿主在业务健康边界、优雅关闭、受保护上传访问、HTTP 入口组织、配置服务边界和运行时默认值方面的操作行为。

## Requirements

### Requirement:宿主不得内建业务健康检查接口

宿主 SHALL NOT 提供内建匿名业务健康检查 HTTP API、健康检查 DTO、健康检查控制器或`health.timeout`等专用配置入口。具体业务应用、交付镜像或业务插件需要健康检查时，必须在自身边界定义业务健康语义和路由，避免`lina-core`将数据库探测、集群角色或业务可用性固化为框架级契约。

#### Scenario:主框架不注册健康检查端点

- **当** LinaPro 宿主启动并完成静态 API 路由注册时
- **则** 主框架不得注册`GET /api/v1/health`
- **且** 主框架不得保留`api/health`或`internal/controller/health`生产代码

#### Scenario:主框架不读取健康检查配置

- **当** 宿主配置服务初始化时
- **则** 配置服务不得暴露`GetHealth`或`HealthConfig`
- **且** 交付配置模板不得声明`health.timeout`

#### Scenario:业务自行定义健康语义

- **当** 业务应用需要对外提供健康检查时
- **则** 业务必须在自身应用、交付层或插件边界实现健康检查
- **且** 业务健康接口不得依赖主框架内建的`/api/v1/health`

#### Scenario:保留已认证运行诊断

- **当** 需要验证集群 coordination 状态时
- **则** 已认证调用方可以继续通过系统信息能力读取 coordination 诊断
- **且** 该诊断不得退化为匿名业务健康检查 API

### Requirement:宿主必须复用 GoFrame HTTP 优雅关闭

宿主 SHALL 对 `SIGTERM`、`SIGINT` 和类似关闭信号使用 GoFrame `Server.Run()` 的内置进程信号处理和 HTTP 优雅关闭行为。`internal/cmd/cmd_http.go` 不得再次注册 `os/signal` 或重新实现 HTTP 服务器关闭循环。GoFrame `Server.Run()` 返回后，宿主必须按此顺序清理拥有的运行时资源：停止 cron 调度器、停止集群服务、关闭数据库连接池。宿主拥有的清理必须受 `shutdown.timeout` 约束，默认 `30s`；`shutdown.timeout` 必须使用带单位的字符串配置值，解析为 `time.Duration`。

#### Scenario:SIGTERM 复用 GoFrame HTTP 关闭

- **当** 宿主进程收到 `SIGTERM` 时
- **则** HTTP 服务器关闭必须由 GoFrame `Server.Run()` 内置信号处理触发
- **且** `cmd_http.go` 不得再注册额外的 `signal.NotifyContext` 或等效的 `os/signal` 监听器
- **且** Cron 调度器必须在 `Server.Run()` 返回后停止接受新触发并等待进行中的任务
- **且** 集群服务必须在 Cron 调度器关闭后停止
- **且** 数据库连接池必须在 Cron 关闭后关闭
- **且** 宿主拥有的运行时清理必须在 `shutdown.timeout` 内完成

#### Scenario:拥有的运行时清理超时

- **当** 关闭超过 `shutdown.timeout` 时
- **则** 宿主必须记录超时警告并返回错误
- **且** 进程不得永久挂起

### Requirement:上传文件访问端点必须归属文件模块并使用宿主统一授权

宿主 SHALL 通过文件 API DTO 和文件控制器声明 `GET /api/v1/uploads/*`，并在受保护路由组下注册到文件控制器。统一 Auth 和 Permission 中间件必须处理认证和权限检查。匿名调用方不得直接访问上传文件。端点权限标签 SHALL 与文件模块菜单/按钮权限对齐。实现必须从 URL 中的相对存储路径查询文件元数据，并通过文件服务存储后端读取文件流。不得在 `internal/cmd/cmd_http.go` 中拼接本地上传目录或直接访问本地文件系统路径。

#### Scenario:未认证访问被拒绝

- **当** 匿名调用方请求 `GET /api/v1/uploads/<path>` 时
- **则** 宿主必须返回标准未认证响应，如 401 或等效业务码
- **且** 文件内容不得出现在响应体中

#### Scenario:已认证但无权限的调用方被拒绝

- **当** 无文件读取权限的已认证调用方请求 `GET /api/v1/uploads/<path>` 时
- **则** 宿主必须返回标准禁止响应，如 403 或等效业务码

#### Scenario:有权限的已认证调用方获取文件

- **当** 有文件读取权限的已认证调用方请求 `GET /api/v1/uploads/<path>`
- **且** 文件存在时
- **则** 宿主返回 `200` 和文件内容
- **且** 文件内容通过文件服务和存储后端读取，而非通过 `cmd_http.go` 中的本地路径处理器

### Requirement:宿主必须移除空的审计占位包

宿主源码 SHALL NOT 保留 `apps/lina-core/pkg/auditi18n/` 和 `apps/lina-core/pkg/audittype/` 等零文件占位目录。真正的审计日志能力可通过单独迭代引入，但主代码库不得保留暗示审计能力已存在的占位符。

#### Scenario:仓库不包含空的审计占位目录

- **当** 检查 `apps/lina-core/pkg/` 时
- **则** `auditi18n` 和 `audittype` 目录不得存在
- **或** 如果存在，必须包含至少一个有效的 `.go` 文件

### Requirement:HTTP 入口代码必须按职责拆分

宿主 SHALL 保持 `apps/lina-core/internal/cmd/cmd_http.go` 聚焦于 HTTP 命令入口编排。HTTP 运行时服务构建、API 路由绑定、前端静态资源服务、宿主 OpenAPI 绑定和启动后生命周期钩子必须在同一包中的独立命名源文件中维护，避免一个 HTTP 入口文件承载多个基础设施实现细节。

#### Scenario:HTTP 入口文件保持轻量编排

- **当** 维护者打开 `apps/lina-core/internal/cmd/cmd_http.go` 时
- **则** 文件应仅包含 `HttpInput`、`HttpOutput` 和 `Main.Http` 启动编排
- **且** 具体的路由绑定、运行时构建、静态资源服务和 OpenAPI 处理器必须位于独立的 `cmd_http_*.go` 文件中

### Requirement:配置服务接口必须组合分类

顶层宿主配置 `Service` 接口 SHALL 通过嵌入组合更窄的基于职责的接口。认证、登录、集群、前端、国际化、cron、宿主运行时、交付元数据、插件、上传和运行时参数同步的配置能力必须在命名清晰的分类接口中维护，避免在顶层 `Service` 上直接累积每个方法。

#### Scenario:Service 接口组合分类接口

- **当** 维护者审查 `apps/lina-core/internal/service/config/config.go` 时
- **则** `Service` 接口必须嵌入多个分类接口
- **且** 每个分类接口中的方法必须在声明旁保留职责注释
- **且** `serviceImpl` 必须继续实现完整的 `Service` 契约

### Requirement:调度器默认时区必须可配置

宿主 SHALL 将 `cron_managed_jobs.go` 中的硬编码默认时区替换为配置键 `scheduler.defaultTimezone`，默认为 `UTC`。源码不得保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 等硬编码常量。

#### Scenario:缺失配置使用 UTC

- **当** 配置文件未声明 `scheduler.defaultTimezone` 时
- **则** 宿主在注册内置任务时使用 `UTC` 作为默认时区

#### Scenario:自定义配置的时区生效

- **当** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"` 时
- **则** 宿主在注册内置任务时使用 `Asia/Shanghai`
