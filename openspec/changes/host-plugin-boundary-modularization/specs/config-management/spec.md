## ADDED Requirements

### Requirement: 日志 TraceID 输出开关支持静态配置与动态系统参数

系统 SHALL 默认关闭日志中的 TraceID 输出，并同时提供 `config.yaml` 静态开关与受保护系统参数 `sys.logger.traceID.enabled` 的运行时覆盖能力。

#### Scenario: 未显式开启时日志默认不输出 TraceID
- **WHEN** `logger.extensions.traceIDEnabled` 未在配置文件中声明，且 `sys.logger.traceID.enabled` 不存在或保持默认继承模式
- **THEN** 宿主日志与 HTTP Server 日志默认不输出 TraceID 字段

#### Scenario: 配置文件显式开启 TraceID 输出
- **WHEN** `logger.extensions.traceIDEnabled` 在配置文件中被显式设置为 `true`
- **AND** `sys.logger.traceID.enabled` 不存在或取值为 `inherit`
- **THEN** 宿主日志与 HTTP Server 日志输出 TraceID 字段

#### Scenario: 系统参数动态覆盖静态配置
- **WHEN** 管理员将 `sys.logger.traceID.enabled` 更新为 `true` 或 `false`
- **THEN** 宿主使用最新系统参数作为日志 TraceID 输出开关
- **AND** 运行中的节点在受保护参数快照刷新后无需重启即可生效

#### Scenario: 初始化内置日志 TraceID 参数元数据
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 中包含 `sys.logger.traceID.enabled` 记录
- **AND** 该参数说明支持 `inherit`、`true`、`false` 三种取值
