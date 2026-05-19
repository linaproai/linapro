## ADDED Requirements

### Requirement: 集群模式会话热状态必须存储在 Redis
系统 SHALL 在 `cluster.enabled=true` 且 `cluster.coordination=redis` 时，将请求路径使用的在线会话热状态存储到 Redis。PostgreSQL `sys_online_session` SHALL 保留为在线用户管理、数据权限过滤、登录信息展示和清理投影。

#### Scenario: 登录写入 Redis hot state 和 PostgreSQL 投影
- **WHEN** 用户登录成功并签发正式 JWT
- **THEN** 系统写入 Redis session hot key
- **AND** Redis session TTL 等于有效会话超时时长
- **AND** 系统写入或更新 `sys_online_session` 投影行

#### Scenario: 受保护请求验证 Redis hot state
- **WHEN** 已登录用户访问受保护 API
- **AND** JWT 签名有效
- **THEN** 认证链读取 Redis session hot key
- **AND** 仅当 Redis session 存在且未被撤销时请求继续处理

### Requirement: 会话热状态必须保留租户和 token 维度
Redis session key SHALL 至少包含 tenant ID 和 token ID。会话 payload SHALL 包含 user ID、username、tenant ID、login time、last active time 和必要客户端上下文。

#### Scenario: 同一用户多个租户会话隔离
- **WHEN** 用户 U 同时持有租户 A token 和租户 B token
- **THEN** Redis 中存在两个不同 session hot key
- **AND** 删除租户 A token 不影响租户 B token

#### Scenario: token 租户不匹配
- **WHEN** JWT claims 中 `TenantId=1001`
- **AND** Redis session hot key 仅存在于 `tenant=1002`
- **THEN** 认证链拒绝请求

### Requirement: 会话活跃时间写回 PostgreSQL 必须节流
系统 SHALL 在 Redis 中维护请求热路径 last active，并按节流窗口写回 PostgreSQL 投影。认证中间件 MUST 避免每个请求都更新 `sys_online_session`。

#### Scenario: 更新窗口内不写 PostgreSQL
- **WHEN** 同一 token 在短更新窗口内连续访问受保护 API
- **THEN** Redis session TTL 可被刷新
- **AND** 系统不得为每次请求写入 `sys_online_session.last_active_time`

#### Scenario: 更新窗口后写回投影
- **WHEN** Redis session 的 last active 已超过 PostgreSQL 投影写回窗口
- **THEN** 系统更新 `sys_online_session.last_active_time`
- **AND** 后续在线用户列表可看到新的活跃时间

### Requirement: 强制下线必须同时撤销 Redis 和 PostgreSQL 投影
系统 SHALL 在强制下线时先校验调用方对目标会话的可见性，再撤销 Redis session hot key、写入 token revoke 状态并删除或标记 PostgreSQL 投影。

#### Scenario: 强制下线成功
- **WHEN** 管理员强制下线其数据权限范围内的 token
- **THEN** 系统删除对应 Redis session hot key
- **AND** 系统写入 revoked token 状态直到 JWT 自然过期
- **AND** 系统删除或标记 `sys_online_session` 投影
- **AND** 后续使用该 token 的请求返回 401

#### Scenario: 拒绝强制下线不可见会话
- **WHEN** 调用方尝试强制下线数据权限范围外的 token
- **THEN** 系统拒绝操作
- **AND** 不删除 Redis session hot key
- **AND** 不修改 `sys_online_session` 投影

### Requirement: Redis 会话故障必须 fail-closed
当集群模式下 Redis session hot state 不可读取时，认证链 SHALL fail-closed。系统不得因为 Redis 不可用而仅凭 JWT 签名放行业务请求。

#### Scenario: Redis 读取失败
- **WHEN** 已登录用户访问受保护 API
- **AND** Redis session 读取返回连接错误
- **THEN** 认证链拒绝请求
- **AND** 返回结构化认证或服务不可用错误
- **AND** 系统记录可观测日志

### Requirement: PostgreSQL 投影清理必须保留
系统 SHALL 保留不活跃会话投影清理任务。集群模式下 Redis 负责 hot state TTL，PostgreSQL 清理任务负责移除过期或孤儿投影。

#### Scenario: Redis session 已过期但投影仍存在
- **WHEN** Redis session key 因 TTL 到期消失
- **AND** `sys_online_session` 中仍有对应投影
- **THEN** 清理任务最终删除该投影
- **AND** 在线用户列表不应长期显示已过期会话

