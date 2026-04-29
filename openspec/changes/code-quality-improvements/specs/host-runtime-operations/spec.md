## ADDED Requirements

### Requirement: 宿主必须提供匿名健康探针接口
宿主 SHALL 在公开路由分组下提供 `GET /api/v1/health`，无需登录态即可访问。该接口 MUST 在响应中返回服务自检结果，并 MUST 包含一次轻量数据库探活；当数据库不可达或探活超时时 MUST 返回 `503` 状态码并附带不可用原因；正常时 MUST 返回 `200` 与 `{"status":"ok","mode":"<single|master|slave>"}` 结构。探活超时时长 SHALL 由配置 `health.timeout` 控制，默认 `2s`，并通过 `time.Duration` 解析。

#### Scenario: 健康探针在数据库正常时返回 200
- **WHEN** 调用方匿名访问 `GET /api/v1/health`
- **AND** 数据库可达且探活在超时时间内完成
- **THEN** 接口返回 `200`，响应体包含 `status="ok"` 与当前部署模式 `mode`
- **AND** 接口不要求 `Authorization` 头

#### Scenario: 健康探针在数据库不可达时返回 503
- **WHEN** 调用方匿名访问 `GET /api/v1/health`
- **AND** 数据库探活在 `health.timeout` 内未返回
- **THEN** 接口返回 `503`，响应体包含 `status="unavailable"` 与简要原因

#### Scenario: 健康探针不挂载到受保护路由分组
- **WHEN** 服务启动并完成路由注册
- **THEN** `/api/v1/health` 路由 MUST 注册在公开路由分组中
- **AND** Auth 与 Permission 中间件 MUST NOT 拦截该路由

### Requirement: 宿主必须支持优雅关停
宿主 SHALL 在进程入口注册 `SIGTERM` 与 `SIGINT` 信号监听，收到信号后 MUST 按"停止接收新请求 → 关停 cron 调度器 → 关闭数据库连接池"的顺序执行关停。整个过程 MUST 受 `shutdown.timeout` 配置约束，默认 `30s`，超时后宿主 MAY 强制退出并记录 warning。`shutdown.timeout` MUST 使用带单位字符串配置并由 `time.Duration` 解析。

#### Scenario: 收到 SIGTERM 后按顺序关停
- **WHEN** 宿主进程收到 `SIGTERM`
- **THEN** HTTP Server MUST 先调用 `Shutdown(ctx)` 停止接收新连接
- **AND** Cron 调度器随后停止接受新触发并等待在途任务结束
- **AND** 数据库连接池在 cron 关停后关闭
- **AND** 全过程在 `shutdown.timeout` 内完成

#### Scenario: 优雅关停超时强制退出
- **WHEN** 关停过程中超过 `shutdown.timeout`
- **THEN** 宿主 MUST 打印超时 warning 并以非零状态码退出
- **AND** 进程不得永久挂起

### Requirement: 上传文件路由必须经过宿主统一鉴权
宿主 SHALL 把 `/api/v1/uploads/*` 路由注册到受保护路由分组下，由统一的 Auth 与 Permission 中间件处理鉴权与权限校验，匿名调用方 MUST NOT 直接访问已上传文件。该路由的权限标签 SHALL 与文件模块菜单/按钮权限保持一致。

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

### Requirement: 宿主必须移除空置的 audit 占位包
宿主源码 MUST NOT 保留 `apps/lina-core/pkg/auditi18n/` 与 `apps/lina-core/pkg/audittype/` 这类零文件占位目录。审计日志能力的真实实现 MAY 通过独立迭代重新引入，但 MUST NOT 在主线代码库中留下"已存在审计能力"的误导。

#### Scenario: 仓库不包含空 audit 占位目录
- **WHEN** 检查 `apps/lina-core/pkg/` 目录
- **THEN** `auditi18n` 与 `audittype` 目录 MUST NOT 存在
- **OR** 如果存在 MUST 至少包含一个生效的 `.go` 文件

### Requirement: 调度器默认时区必须可配置
宿主 SHALL 把 `cron_managed_jobs.go` 中内置任务的默认时区从硬编码常量改为配置读取，配置键为 `scheduler.defaultTimezone`，默认值为 `UTC`。源码 MUST NOT 再保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 这类硬编码常量。

#### Scenario: 配置缺省时使用 UTC
- **WHEN** 配置文件未声明 `scheduler.defaultTimezone`
- **THEN** 宿主在注册内置任务时使用 `UTC` 作为默认时区

#### Scenario: 配置自定义时区生效
- **WHEN** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"`
- **THEN** 宿主在注册内置任务时使用 `Asia/Shanghai` 作为默认时区
