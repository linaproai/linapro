## ADDED Requirements

### Requirement: 调度任务默认时区必须可配置
系统 SHALL 把内置 cron 任务的默认时区改为通过配置 `scheduler.defaultTimezone` 读取，默认值为 `UTC`。源码 MUST NOT 保留 `defaultManagedJobTimezone = "Asia/Shanghai"` 类硬编码常量。

#### Scenario: 缺省配置使用 UTC 时区
- **WHEN** 配置文件未声明 `scheduler.defaultTimezone`
- **AND** 服务启动并注册内置任务
- **THEN** 内置任务 MUST 以 `UTC` 作为默认时区

#### Scenario: 自定义时区生效
- **WHEN** 配置文件设置 `scheduler.defaultTimezone: "Asia/Shanghai"`
- **AND** 服务启动并注册内置任务
- **THEN** 内置任务 MUST 以 `Asia/Shanghai` 作为默认时区

### Requirement: sys_job 表禁止使用外键约束
系统 SHALL 移除 `sys_job` 表的 `fk_sys_job_group_id` 外键约束，改为在应用层维护 `group_id` 与 `sys_job_group` 的引用一致性，并在 `sys_job` 上保留 `KEY idx_group_id (group_id)` 以支撑按分组的查询/级联清理路径。仓库其他关联表均依靠应用层一致性，本表 MUST 与该约定保持一致以避免高并发任务调度场景下额外的外键锁开销。

#### Scenario: sys_job 表不再包含外键约束
- **WHEN** `make init` 完成数据库初始化
- **THEN** `sys_job` 表 MUST NOT 包含 `fk_sys_job_group_id` 或任何指向 `sys_job_group` 的 `FOREIGN KEY` 约束

#### Scenario: sys_job 仍保留 group_id 索引
- **WHEN** `make init` 完成数据库初始化
- **THEN** `SHOW INDEX FROM sys_job` 结果中 MUST 出现 `idx_group_id`，索引列为 `group_id`

#### Scenario: 写入路径校验 group_id 引用一致性
- **WHEN** 上层调用方创建或更新 `sys_job` 记录并提供 `group_id`
- **THEN** Service 层 MUST 校验 `group_id` 引用的分组存在
- **AND** 校验失败 MUST 返回 `bizerr` 业务错误而非依赖数据库外键拦截
