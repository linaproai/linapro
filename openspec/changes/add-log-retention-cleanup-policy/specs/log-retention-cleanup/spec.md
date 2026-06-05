## ADDED Requirements

### Requirement: 日志最长保留天数运行时参数

系统 SHALL 在系统参数设置中提供受保护运行时参数 `sys.log.retentionDays`，单位为天，默认值为 `90`，用于约束宿主与官方源码插件日志数据的最长存储时间。

#### Scenario: 初始化全局日志保留参数

- **WHEN** 执行宿主初始化 SQL
- **THEN** `sys_config` MUST 包含 `key=sys.log.retentionDays` 的内建参数
- **AND** 参数默认值 MUST 为 `90`
- **AND** 参数名称和备注 MUST 说明单位为天，以及该参数约束日志最长存储时间

#### Scenario: 校验日志保留天数

- **WHEN** 管理员新增、修改或导入 `sys.log.retentionDays`
- **THEN** 系统 MUST 只接受正整数值
- **AND** 空值、零、负数或非数字值 MUST 被拒绝
- **AND** 参数变更 MUST 复用 protected runtime parameter 快照与共享修订号失效机制

#### Scenario: 插件读取全局日志保留天数

- **WHEN** 源码插件需要读取 `sys.log.retentionDays`
- **THEN** 插件 MUST 通过宿主发布的配置能力读取运行时有效值
- **AND** 插件 MUST NOT 导入 `apps/lina-core/internal/service/config` 或宿主内部 DAO、DO、Entity

### Requirement: 日志清理任务必须遵守全局最长保留天数

系统 SHALL 让系统监控在线会话投影、系统监控登录日志、任务调度执行日志和智能中心调用日志的自动清理任务遵守 `sys.log.retentionDays`。

#### Scenario: 任务调度执行日志应用全局上限

- **WHEN** 任务调度执行日志清理任务触发
- **THEN** 系统 MUST 删除 `start_at` 早于当前时间减去 `sys.log.retentionDays` 的执行日志
- **AND** 系统 MUST 再应用既有 `cron.log.retention` 和任务级 `log_retention_override` 的额外清理策略
- **AND** 任务级 `none` 策略 MUST NOT 使执行日志超过全局最长保留天数

#### Scenario: 在线会话投影清理不误删有效会话

- **WHEN** 在线会话清理任务触发
- **THEN** 系统 MUST 使用会话超时和 `sys.log.retentionDays` 中更严格的时间边界清理历史在线会话记录
- **AND** 系统 MUST NOT 仅因为会话创建时间超过 `sys.log.retentionDays` 就删除仍处于有效状态的活跃会话

#### Scenario: 登录日志插件自动清理过期日志

- **WHEN** `linapro-monitor-loginlog` 插件已安装启用且内建清理任务触发
- **THEN** 插件 MUST 按 `login_time` 删除早于当前时间减去 `sys.log.retentionDays` 的登录日志
- **AND** 清理 MUST 使用数据库侧范围删除，不得逐行循环删除

#### Scenario: 智能中心调用日志插件自动清理过期日志

- **WHEN** `linapro-ai-core` 插件已安装启用且内建清理任务触发
- **THEN** 插件 MUST 按 `created_at` 删除早于当前时间减去 `sys.log.retentionDays` 的调用日志
- **AND** 清理 MUST 使用数据库侧范围删除，不得逐行循环删除

### Requirement: 日志清理验证与性能边界

系统 SHALL 为日志最长保留天数和自动清理行为提供自动化验证，并说明数据权限、缓存一致性和查询性能边界。

#### Scenario: 清理任务具备自动化测试

- **WHEN** 修改日志保留参数或清理任务实现
- **THEN** 宿主和插件后端单元测试 MUST 覆盖默认值、非法值、过期删除和未过期保留
- **AND** OpenSpec 严格校验 MUST 通过

#### Scenario: 清理任务避免无约束扫描和循环删除

- **WHEN** 自动清理任务执行
- **THEN** 清理条件 MUST 使用日志时间字段在数据库侧过滤
- **AND** 高频日志表 MUST 具备支撑时间范围删除的索引
- **AND** 实现 MUST NOT 对每条日志单独查询或删除

