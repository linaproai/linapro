## ADDED Requirements

### Requirement: 日志 TraceID 输出开关仅由静态配置文件控制

系统 SHALL 默认关闭日志中的 TraceID 输出，并且仅允许通过 `config.yaml` 中的 `logger.extensions.traceIDEnabled` 静态开关控制是否输出 TraceID。

#### Scenario: 未显式开启时日志默认不输出 TraceID
- **WHEN** `logger.extensions.traceIDEnabled` 未在配置文件中声明
- **THEN** 宿主日志与 HTTP Server 日志默认不输出 TraceID 字段

#### Scenario: 配置文件显式开启 TraceID 输出
- **WHEN** `logger.extensions.traceIDEnabled` 在配置文件中被显式设置为 `true`
- **THEN** 宿主日志与 HTTP Server 日志输出 TraceID 字段

#### Scenario: 初始化内置参数时不再暴露 TraceID 系统参数
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 中不包含 `sys.logger.traceID.enabled` 记录
