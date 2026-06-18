## MODIFIED Requirements

### Requirement: 宿主必须复用 GoFrame HTTP 优雅关闭

宿主 SHALL 对 `SIGTERM`、`SIGINT` 和类似关闭信号使用 GoFrame `Server.Run()` 的内置进程信号处理和 HTTP 优雅关闭行为。`internal/cmd/cmd_http.go` 不得再次注册 `os/signal` 或重新实现 HTTP 服务器关闭循环。GoFrame `Server.Run()` 返回后，宿主必须按此顺序清理拥有的运行时资源：停止 cron 调度器、停止集群服务、停止协调服务、关闭数据库连接池。宿主拥有的运行时清理必须复用 GoFrame Server 已生效的 `server.gracefulShutdownTimeout` 作为超时预算，不得再通过顶层 `shutdown.timeout` 或等价 LinaPro 自定义配置重复维护优雅停止超时。

#### Scenario: SIGTERM 复用 GoFrame HTTP 关闭

- **当** 宿主进程收到 `SIGTERM` 时
- **则** HTTP 服务器关闭必须由 GoFrame `Server.Run()` 内置信号处理触发
- **且** `cmd_http.go` 不得再注册额外的 `signal.NotifyContext` 或等效的 `os/signal` 监听器
- **且** Cron 调度器必须在 `Server.Run()` 返回后停止接受新触发并等待进行中的任务
- **且** 集群服务必须在 Cron 调度器关闭后停止
- **且** 协调服务必须在集群服务关闭后停止
- **且** 数据库连接池必须在协调服务关闭后关闭
- **且** 宿主拥有的运行时清理必须在当前 GoFrame Server 的 `server.gracefulShutdownTimeout` 内完成

#### Scenario: 拥有的运行时清理超时

- **当** 关闭超过当前 GoFrame Server 的 `server.gracefulShutdownTimeout` 时
- **则** 宿主必须记录超时警告并返回错误
- **且** 进程不得永久挂起

#### Scenario: 停机配置不重复维护

- **当** 交付配置需要调整 HTTP 和宿主资源优雅停止超时时
- **则** 配置必须使用 GoFrame Server 原生的 `server.gracefulShutdownTimeout`
- **且** 交付配置不得再声明顶层 `shutdown.timeout`
