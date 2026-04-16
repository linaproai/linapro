## MODIFIED Requirements

### Requirement: 不活跃会话自动清理
系统 SHALL 提供定时任务自动清理长时间未操作的在线会话，防止会话表无限增长。超时阈值和清理频率 MUST 支持通过 duration 字符串配置文件调整。

#### Scenario: 定时清理超时会话
- **WHEN** 定时清理任务执行时（默认每 5 分钟一次）
- **THEN** 系统 MUST 查询 `sys_online_session` 表中 `last_active_time` 距当前时间超过超时阈值（默认 24 小时）的记录，并将其删除

#### Scenario: 超时阈值可通过新配置调整
- **WHEN** 管理员在 `config.yaml` 中设置 `session.timeout=24h`
- **THEN** 系统 MUST 使用该 duration 值作为会话超时阈值，不设置时默认为 24 小时

#### Scenario: 清理频率可通过新配置调整
- **WHEN** 管理员在 `config.yaml` 中设置 `session.cleanupInterval=5m`
- **THEN** 系统 MUST 使用该 duration 值作为清理任务执行间隔，不设置时默认为 5 分钟
