## ADDED Requirements

### Requirement: 运行时配置 revision 必须使用 Redis coordination
系统 SHALL 在集群模式下通过 Redis revision/event 协调受保护运行时参数变更。`sys_config` 仍为权威数据源，Redis 仅承载 revision 和失效事件。

#### Scenario: 修改 JWT 过期配置
- **WHEN** 管理员修改 `sys.jwt.expire`
- **THEN** 系统提交 `sys_config` 权威数据
- **AND** 发布 `runtime-config` Redis revision
- **AND** 其他节点刷新运行时参数快照

#### Scenario: 修改会话超时配置
- **WHEN** 管理员修改 `sys.session.timeout`
- **THEN** 系统发布 `runtime-config` Redis revision
- **AND** 新请求使用更新后的会话超时策略

### Requirement: 运行时配置 freshness 不可确认时必须返回可见错误
系统 SHALL 在读取受保护运行时参数前确认本地快照 freshness。当 Redis revision 不可读取且本地快照超过最大陈旧窗口时，系统 MUST 返回结构化错误，不得静默使用陈旧配置。

#### Scenario: Redis runtime-config revision 不可读
- **WHEN** 请求路径需要读取受保护运行时参数
- **AND** Redis revision 不可读
- **AND** 本地运行时参数快照超过最大陈旧窗口
- **THEN** 系统返回结构化配置 freshness 错误
- **AND** 记录可观测日志

### Requirement: 单机运行时配置保持本地 revision
系统 SHALL 在单机模式下使用进程内 revision 管理运行时参数快照失效，不得要求 Redis。

#### Scenario: 单机修改运行时配置
- **WHEN** `cluster.enabled=false`
- **AND** 管理员修改受保护运行时参数
- **THEN** 系统更新进程内 revision
- **AND** 当前进程清理本地运行时参数快照

