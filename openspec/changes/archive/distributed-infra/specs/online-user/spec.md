## ADDED Requirements

### Requirement: 在线用户列表必须使用 PostgreSQL 投影并可结合 Redis hot state
系统 SHALL 继续以 `sys_online_session` 作为在线用户管理查询投影。集群模式下请求热状态存储在 Redis，但在线用户列表的数据权限过滤、分页、搜索和治理字段 MUST 通过 PostgreSQL 投影完成。

#### Scenario: 集群模式查询在线用户列表
- **WHEN** 管理员在集群模式下查询在线用户列表
- **THEN** 系统从 `sys_online_session` 投影查询可见会话
- **AND** 查询继续接入 tenantcap 和 datascope
- **AND** 返回 token_id、用户名、部门、IP、浏览器、操作系统、登录时间和最后活跃时间

#### Scenario: Redis hot state 与投影短暂不一致
- **WHEN** Redis session 已过期但 PostgreSQL 投影尚未清理
- **THEN** 清理任务最终删除投影
- **AND** 认证链以 Redis hot state 为请求有效性权威

### Requirement: 在线用户强退必须撤销 Redis hot state
系统 SHALL 在强制下线时删除 Redis session hot key 并写入 Redis revoke 状态。仅删除 PostgreSQL 投影不得视为完成强退。

#### Scenario: 强退后所有节点拒绝 token
- **WHEN** 管理员强制下线 token T
- **THEN** 系统删除 Redis session hot key
- **AND** 系统写入 Redis revoke key
- **AND** 任意节点收到 token T 的后续请求时拒绝

### Requirement: 在线用户清理必须区分 hot state 和投影
系统 SHALL 使用 Redis TTL 管理 session hot state 过期，并保留 PostgreSQL 投影清理任务。投影清理任务 MUST 幂等。

#### Scenario: 投影清理删除过期会话
- **WHEN** PostgreSQL 投影中存在 `last_active_time` 超过会话超时阈值的记录
- **THEN** 清理任务删除该投影
- **AND** 重复执行清理任务不会产生错误

