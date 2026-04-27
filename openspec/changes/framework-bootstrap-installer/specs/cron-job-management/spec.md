## MODIFIED Requirements

### Requirement: 日志清理策略

系统 SHALL 提供全局默认清理策略与任务级覆盖，由系统内置定时任务定期执行清理。内置清理任务 SHALL 由宿主源码注册，并在服务启动时投影同步到 `sys_job`；交付型 SQL SHALL NOT 向 `sys_job` 写入 `host:cleanup-job-logs` 的初始化种子数据。

#### Scenario: 全局默认策略

- **WHEN** 任务 `log_retention_override` 为空
- **THEN** 该任务日志 SHALL 按系统参数 `cron.log.retention` 的策略清理

#### Scenario: 任务级覆盖

- **WHEN** 任务 `log_retention_override` 配置为 `{mode: days, value: 60}` 或 `{mode: count, value: 500}`
- **THEN** 系统 SHALL 按任务级策略清理该任务日志，忽略全局默认

#### Scenario: 不清理策略

- **WHEN** 策略 `mode=none`
- **THEN** 系统 SHALL 不清理该任务对应日志

#### Scenario: 内置清理任务

- **WHEN** 宿主服务启动并执行内置任务同步
- **THEN** 系统 SHALL 将 `host:cleanup-job-logs` handler 任务投影到 `sys_job`
- **AND** 默认 `cron_expr` 为每日凌晨触发
- **AND** 该任务 `is_builtin=1`
- **AND** 交付型 SQL 不包含该任务的 `sys_job` 初始化种子数据
