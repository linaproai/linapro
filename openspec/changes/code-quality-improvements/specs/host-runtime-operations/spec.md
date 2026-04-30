## ADDED Requirements

### Requirement: 宿主必须提供匿名健康探针接口
宿主 SHALL 在公开路由分组下提供 `GET /api/v1/health`，无需登录态即可访问。该接口 MUST 通过标准 API DTO 与 controller 暴露，响应中返回服务自检结果，并 MUST 包含一次轻量数据库探活；当数据库不可达或探活超时时 MUST 返回 `503` 状态码并附带稳定、脱敏的不可用原因；正常时 MUST 返回 `200` 与 `{"status":"ok","mode":"<single|master|slave>"}` 结构。探活超时时长 SHALL 由配置 `health.timeout` 控制，默认 `5s`，并通过 `time.Duration` 解析。

#### Scenario: 健康探针在数据库正常时返回 200
- **WHEN** 调用方匿名访问 `GET /api/v1/health`
- **AND** 数据库可达且探活在超时时间内完成
- **THEN** 接口返回 `200`，响应体包含 `status="ok"` 与当前部署模式 `mode`
- **AND** 接口不要求 `Authorization` 头

#### Scenario: 健康探针在数据库不可达时返回 503
- **WHEN** 调用方匿名访问 `GET /api/v1/health`
- **AND** 数据库探活在 `health.timeout` 内未返回
- **THEN** 接口返回 `503`，响应体包含 `status="unavailable"` 与稳定脱敏原因
- **AND** 原始数据库、网络或 schema 错误 MUST 仅记录到日志，不得直接返回给匿名调用方

#### Scenario: 健康探针不挂载到受保护路由分组
- **WHEN** 服务启动并完成路由注册
- **THEN** `/api/v1/health` 路由 MUST 注册在公开路由分组中
- **AND** Auth 与 Permission 中间件 MUST NOT 拦截该路由

### Requirement: 宿主必须复用 GoFrame HTTP 优雅关停能力
宿主 SHALL 使用 GoFrame `Server.Run()` 内置的进程信号监听与 HTTP graceful shutdown 能力处理 `SIGTERM`、`SIGINT` 等关闭信号，`internal/cmd/cmd_http.go` MUST NOT 重复注册 `os/signal` 或重复实现 HTTP Server shutdown 循环。GoFrame `Server.Run()` 返回后，宿主 MUST 按"关停 cron 调度器 → 停止集群服务 → 关闭数据库连接池"的顺序清理自有运行期资源。宿主自有运行期资源清理 MUST 受 `shutdown.timeout` 配置约束，默认 `30s`；`shutdown.timeout` MUST 使用带单位字符串配置并由 `time.Duration` 解析。

#### Scenario: 收到 SIGTERM 后复用 GoFrame 关闭 HTTP Server
- **WHEN** 宿主进程收到 `SIGTERM`
- **THEN** HTTP Server shutdown MUST 由 GoFrame `Server.Run()` 内置信号处理流程触发
- **AND** `cmd_http.go` MUST NOT 同时注册额外的 `signal.NotifyContext` 或等价 `os/signal` 监听
- **AND** Cron 调度器 MUST 在 `Server.Run()` 返回后停止接受新触发并等待在途任务结束
- **AND** 集群服务 MUST 在 Cron 调度器关停后停止
- **AND** 数据库连接池在 cron 关停后关闭
- **AND** 宿主自有运行期资源清理在 `shutdown.timeout` 内完成

#### Scenario: 自有运行期资源清理超时
- **WHEN** 关停过程中超过 `shutdown.timeout`
- **THEN** 宿主 MUST 打印超时 warning 并返回错误
- **AND** 进程不得永久挂起

### Requirement: 上传文件访问接口必须内聚到文件模块并经过宿主统一鉴权
宿主 SHALL 通过文件 API DTO 与文件 controller 声明 `GET /api/v1/uploads/*` 访问接口，并随文件 controller 注册到受保护路由分组下，由统一的 Auth 与 Permission 中间件处理鉴权与权限校验，匿名调用方 MUST NOT 直接访问已上传文件。该接口的权限标签 SHALL 与文件模块菜单/按钮权限保持一致。接口实现 MUST 根据 URL 中的相对存储路径查询文件元数据，并通过文件 service 的 storage backend 读取文件流；实现 MUST NOT 在 `internal/cmd/cmd_http.go` 中直接拼接本地上传目录或直接访问本地文件系统路径。

#### Scenario: 未登录访问被拒绝
- **WHEN** 匿名调用方请求 `GET /api/v1/uploads/<path>`
- **THEN** 宿主 MUST 返回标准未认证响应（401 或同等业务错误码）
- **AND** 文件内容 MUST NOT 出现在响应体中

#### Scenario: 已登录但无权限调用方被拒绝
- **WHEN** 已登录但缺少文件读取权限的调用方请求 `GET /api/v1/uploads/<path>`
- **THEN** 宿主 MUST 返回标准无权限响应（403 或同等业务错误码）

#### Scenario: 已登录且有权限调用方正常获取文件
- **WHEN** 已登录且具备文件读取权限的调用方请求 `GET /api/v1/uploads/<path>`
- **AND** 该文件存在
- **THEN** 宿主返回 `200` 与文件内容
- **AND** 文件内容通过文件 service 与 storage backend 读取，而不是通过 `cmd_http.go` 中的本地路径 handler 读取

### Requirement: 宿主必须移除空置的 audit 占位包
宿主源码 MUST NOT 保留 `apps/lina-core/pkg/auditi18n/` 与 `apps/lina-core/pkg/audittype/` 这类零文件占位目录。审计日志能力的真实实现 MAY 通过独立迭代重新引入，但 MUST NOT 在主线代码库中留下"已存在审计能力"的误导。

#### Scenario: 仓库不包含空 audit 占位目录
- **WHEN** 检查 `apps/lina-core/pkg/` 目录
- **THEN** `auditi18n` 与 `audittype` 目录 MUST NOT 存在
- **OR** 如果存在 MUST 至少包含一个生效的 `.go` 文件

### Requirement: HTTP 入口代码必须按职责拆分
宿主 SHALL 保持 `apps/lina-core/internal/cmd/cmd_http.go` 只承担 HTTP 命令入口编排职责。HTTP 运行期服务构造、API 路由绑定、前端静态资源服务、host OpenAPI 绑定和启动后 lifecycle hook MUST 拆分到同包下按职责命名的独立源码文件维护，避免单个 HTTP 入口文件同时承载多类基础设施实现细节。

#### Scenario: HTTP 入口文件保持轻量编排
- **WHEN** 维护者打开 `apps/lina-core/internal/cmd/cmd_http.go`
- **THEN** 文件中 SHOULD 只包含 `HttpInput`、`HttpOutput` 与 `Main.Http` 启动编排
- **AND** 具体路由绑定、runtime 构造、静态资源服务与 OpenAPI handler MUST 位于独立的 `cmd_http_*.go` 文件中

### Requirement: 配置服务接口必须按类别组合
宿主配置 service 的顶层 `Service` 接口 SHALL 通过 embed 方式组合按职责拆分出的窄接口。鉴权、登录、集群、前端、i18n、cron、host runtime、交付元数据、插件、上传和运行时参数同步等配置能力 MUST 分别由命名清晰的分类接口维护，避免在顶层 `Service` 中直接堆叠全部方法。

#### Scenario: Service 接口通过分类接口组合
- **WHEN** 维护者查看 `apps/lina-core/internal/service/config/config.go`
- **THEN** `Service` 接口 MUST embed 多个分类接口
- **AND** 各分类接口中的方法 MUST 保留紧邻方法声明的职责注释
- **AND** `serviceImpl` MUST 继续实现完整 `Service` 契约

### Requirement: 调度器默认时区必须可配置
宿主 SHALL 把 `cron_managed_jobs.go` 中内置任务的默认时区从硬编码常量改为配置读取，配置键为 `scheduler.defaultTimezone`，默认值为 `UTC`。源码 MUST NOT 再保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 这类硬编码常量。

#### Scenario: 配置缺省时使用 UTC
- **WHEN** 配置文件未声明 `scheduler.defaultTimezone`
- **THEN** 宿主在注册内置任务时使用 `UTC` 作为默认时区

#### Scenario: 配置自定义时区生效
- **WHEN** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"`
- **THEN** 宿主在注册内置任务时使用 `Asia/Shanghai` 作为默认时区
