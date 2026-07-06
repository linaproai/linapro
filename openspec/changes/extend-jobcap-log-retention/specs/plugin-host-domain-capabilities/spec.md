## ADDED Requirements

### Requirement: `Jobs` 领域能力创建、更新和查询必须支持任务级日志清理策略

系统 SHALL 允许源码插件和动态插件通过 `jobcap.Service.Create` 与 `jobcap.Service.Update` 在 `SaveInput` 中传入可选的任务级日志清理策略，并在 `jobcap.Service.Get`、`jobcap.Service.BatchGet` 与 `jobcap.Service.List` 返回的 `JobInfo` 中读取该任务级策略投影。该策略的模式和值 MUST 与宿主定时任务管理 API 的 `logRetentionOverride` 语义一致，并由宿主任务 owner 持久化到 `sys_job.log_retention_override`。未传入或未持久化策略时，任务 MUST 跟随系统参数 `sys.cron.log.retention`。策略为 `mode=none` 时，该任务 MUST 不按任务级策略清理日志。

#### Scenario: 源码插件创建任务时传入日志清理策略

- **WHEN** 源码插件通过 `Services().Jobs().Create` 创建定时任务
- **AND** `SaveInput.LogRetentionOverride` 为 `{mode:"days",value:60}` 或 `{mode:"count",value:500}`
- **THEN** 宿主 MUST 将该策略传递给定时任务 owner
- **AND** 后续日志清理 MUST 使用该任务级策略覆盖全局默认策略

#### Scenario: 动态插件更新任务时传入无清理策略

- **WHEN** 动态插件通过已授权的 `service: jobs`、`method: jobs.update` 更新定时任务
- **AND** 请求载荷中的 `LogRetentionOverride` 为 `{mode:"none",value:0}`
- **THEN** WASM host service MUST 将该策略解码为 `jobcap.SaveInput`
- **AND** 宿主 MUST 将该策略传递给定时任务 owner
- **AND** 后续日志清理 MUST 不按任务级策略清理该任务日志

#### Scenario: 未传入任务级策略时跟随系统默认

- **WHEN** 源码插件或动态插件创建、更新定时任务
- **AND** `SaveInput.LogRetentionOverride` 为空
- **THEN** 宿主 MUST 清空任务级覆盖
- **AND** 后续日志清理 MUST 按系统参数 `sys.cron.log.retention` 执行

#### Scenario: 插件查询任务时返回日志清理策略投影

- **WHEN** 源码插件或动态插件通过 `jobcap.Service.Get`、`jobcap.Service.BatchGet` 或 `jobcap.Service.List` 查询可见定时任务
- **AND** 目标任务的 `sys_job.log_retention_override` 为 `{mode:"days",value:60}`、`{mode:"count",value:500}` 或 `{mode:"none",value:0}`
- **THEN** 返回的 `JobInfo.LogRetentionOverride` MUST 包含对应的模式和值
- **AND** 若目标任务没有任务级策略，`JobInfo.LogRetentionOverride` MUST 为空，表示跟随系统默认策略
- **AND** 查询实现 MUST 通过现有可见性过滤和同一次任务投影读取该字段，不得为每条任务额外查询

#### Scenario: 非法策略复用任务 owner 校验

- **WHEN** 插件传入不支持的日志清理模式，或 `days`、`count` 模式的值小于等于 `0`
- **THEN** 宿主 MUST 返回结构化业务错误
- **AND** 不得创建或更新目标定时任务
