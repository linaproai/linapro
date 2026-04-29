## ADDED Requirements

### Requirement: sys_online_session 必须包含 last_active_time 索引
系统 SHALL 在 `sys_online_session` 表上维护 `KEY idx_last_active_time (last_active_time)` 索引，以支撑按 `last_active_time` 范围清理不活跃会话的查询路径，避免全表扫描。

#### Scenario: 索引存在
- **WHEN** `make init` 完成数据库初始化
- **THEN** `SHOW INDEX FROM sys_online_session` 结果中 MUST 出现 `idx_last_active_time`，索引列为 `last_active_time`

#### Scenario: 不活跃会话清理走索引
- **WHEN** 服务执行 `WHERE last_active_time < ?` 类型清理查询
- **THEN** 数据库 MUST 选中 `idx_last_active_time` 索引以避免全表扫描
