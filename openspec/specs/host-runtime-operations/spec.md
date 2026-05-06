# 宿主运行时操作

## 目的

定义宿主在健康探测、优雅关闭、受保护上传访问、HTTP 入口组织、配置服务边界和运行时默认值方面的操作行为。

## 需求

### 需求：宿主必须提供匿名健康探测端点

宿主 SHALL 在公共路由组下提供 `GET /api/v1/health`，无需登录状态即可访问。该端点必须通过标准 API DTO 和控制器流程暴露，返回服务自检结果，并包含一个轻量级数据库探测。数据库不可用或探测超时时，必须返回 HTTP `503` 和稳定的脱敏不可用原因。健康时，必须返回 HTTP `200` 和 `{"status":"ok","mode":"<single|master|slave>"}`。探测超时 SHALL 由配置键 `health.timeout` 控制，默认 `5s`，解析为 `time.Duration`。

#### 场景：数据库健康时健康探测返回 200

- **当** 调用方匿名访问 `GET /api/v1/health`
- **且** 数据库可达且探测在超时内完成
- **则** 端点返回 `200`，响应体包含 `status="ok"` 加当前部署 `mode`
- **且** 端点不要求 `Authorization` 头

#### 场景：数据库不可用时健康探测返回 503

- **当** 调用方匿名访问 `GET /api/v1/health`
- **且** 数据库探测未在 `health.timeout` 内返回
- **则** 端点返回 `503`，响应体包含 `status="unavailable"` 加稳定的脱敏原因
- **且** 原始数据库、网络或模式错误必须仅记录日志，不得返回给匿名调用方

#### 场景：健康探测不挂载在受保护路由下

- **当** 服务启动且路由注册完成时
- **则** `/api/v1/health` 必须注册在公共路由组下
- **且** Auth 和 Permission 中间件不得拦截该路由

### 需求：宿主必须复用 GoFrame HTTP 优雅关闭

宿主 SHALL 对 `SIGTERM`、`SIGINT` 和类似关闭信号使用 GoFrame `Server.Run()` 的内置进程信号处理和 HTTP 优雅关闭行为。`internal/cmd/cmd_http.go` 不得再次注册 `os/signal` 或重新实现 HTTP 服务器关闭循环。GoFrame `Server.Run()` 返回后，宿主必须按此顺序清理拥有的运行时资源：停止 cron 调度器、停止集群服务、关闭数据库连接池。宿主拥有的清理必须受 `shutdown.timeout` 约束，默认 `30s`；`shutdown.timeout` 必须使用带单位的字符串配置值，解析为 `time.Duration`。

#### 场景：SIGTERM 复用 GoFrame HTTP 关闭

- **当** 宿主进程收到 `SIGTERM` 时
- **则** HTTP 服务器关闭必须由 GoFrame `Server.Run()` 内置信号处理触发
- **且** `cmd_http.go` 不得再注册额外的 `signal.NotifyContext` 或等效的 `os/signal` 监听器
- **且** Cron 调度器必须在 `Server.Run()` 返回后停止接受新触发并等待进行中的任务
- **且** 集群服务必须在 Cron 调度器关闭后停止
- **且** 数据库连接池必须在 Cron 关闭后关闭
- **且** 宿主拥有的运行时清理必须在 `shutdown.timeout` 内完成

#### 场景：拥有的运行时清理超时

- **当** 关闭超过 `shutdown.timeout` 时
- **则** 宿主必须记录超时警告并返回错误
- **且** 进程不得永久挂起

### 需求：上传文件访问端点必须归属文件模块并使用宿主统一授权

宿主 SHALL 通过文件 API DTO 和文件控制器声明 `GET /api/v1/uploads/*`，并在受保护路由组下注册到文件控制器。统一 Auth 和 Permission 中间件必须处理认证和权限检查。匿名调用方不得直接访问上传文件。端点权限标签 SHALL 与文件模块菜单/按钮权限对齐。实现必须从 URL 中的相对存储路径查询文件元数据，并通过文件服务存储后端读取文件流。不得在 `internal/cmd/cmd_http.go` 中拼接本地上传目录或直接访问本地文件系统路径。

#### 场景：未认证访问被拒绝

- **当** 匿名调用方请求 `GET /api/v1/uploads/<path>` 时
- **则** 宿主必须返回标准未认证响应，如 401 或等效业务码
- **且** 文件内容不得出现在响应体中

#### 场景：已认证但无权限的调用方被拒绝

- **当** 无文件读取权限的已认证调用方请求 `GET /api/v1/uploads/<path>` 时
- **则** 宿主必须返回标准禁止响应，如 403 或等效业务码

#### 场景：有权限的已认证调用方获取文件

- **当** 有文件读取权限的已认证调用方请求 `GET /api/v1/uploads/<path>`
- **且** 文件存在时
- **则** 宿主返回 `200` 和文件内容
- **且** 文件内容通过文件服务和存储后端读取，而非通过 `cmd_http.go` 中的本地路径处理器

### 需求：宿主必须移除空的审计占位包

宿主源码不得保留 `apps/lina-core/pkg/auditi18n/` 和 `apps/lina-core/pkg/audittype/` 等零文件占位目录。真正的审计日志能力可通过单独迭代引入，但主代码库不得保留暗示审计能力已存在的占位符。

#### 场景：仓库不包含空的审计占位目录

- **当** 检查 `apps/lina-core/pkg/` 时
- **则** `auditi18n` 和 `audittype` 目录不得存在
- **或** 如果存在，必须包含至少一个有效的 `.go` 文件

### 需求：HTTP 入口代码必须按职责拆分

宿主 SHALL 保持 `apps/lina-core/internal/cmd/cmd_http.go` 聚焦于 HTTP 命令入口编排。HTTP 运行时服务构建、API 路由绑定、前端静态资源服务、宿主 OpenAPI 绑定和启动后生命周期钩子必须在同一包中的独立命名源文件中维护，避免一个 HTTP 入口文件承载多个基础设施实现细节。

#### 场景：HTTP 入口文件保持轻量编排

- **当** 维护者打开 `apps/lina-core/internal/cmd/cmd_http.go` 时
- **则** 文件应仅包含 `HttpInput`、`HttpOutput` 和 `Main.Http` 启动编排
- **且** 具体的路由绑定、运行时构建、静态资源服务和 OpenAPI 处理器必须位于独立的 `cmd_http_*.go` 文件中

### 需求：配置服务接口必须组合分类

顶层宿主配置 `Service` 接口 SHALL 通过嵌入组合更窄的基于职责的接口。认证、登录、集群、前端、国际化、cron、宿主运行时、交付元数据、插件、上传和运行时参数同步的配置能力必须在命名清晰的分类接口中维护，避免在顶层 `Service` 上直接累积每个方法。

#### 场景：Service 接口组合分类接口

- **当** 维护者审查 `apps/lina-core/internal/service/config/config.go` 时
- **则** `Service` 接口必须嵌入多个分类接口
- **且** 每个分类接口中的方法必须在声明旁保留职责注释
- **且** `serviceImpl` 必须继续实现完整的 `Service` 契约

### 需求：调度器默认时区必须可配置

宿主 SHALL 将 `cron_managed_jobs.go` 中的硬编码默认时区替换为配置键 `scheduler.defaultTimezone`，默认为 `UTC`。源码不得保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 等硬编码常量。

#### 场景：缺失配置使用 UTC

- **当** 配置文件未声明 `scheduler.defaultTimezone` 时
- **则** 宿主在注册内置任务时使用 `UTC` 作为默认时区

#### 场景：自定义配置的时区生效

- **当** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"` 时
- **则** 宿主在注册内置任务时使用 `Asia/Shanghai`
