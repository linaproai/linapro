## ADDED Requirements

### Requirement: 单机模式会话有效性必须以 sys_online_session 为权威

系统 SHALL 在单机模式下使用`sys_online_session`作为用户 session 有效性的权威来源。进程内 JWT revoke 状态仅用于当前进程快速拒绝已撤销 token，不得作为服务重启后的登录态权威。退出、强制下线、租户切换和 impersonation revoke 等会话失效路径 MUST 删除对应`sys_online_session`记录或确保完整认证入口无法再通过该 session。

#### Scenario: 服务重启后已退出 token 被拒绝

- **WHEN** 用户退出登录后宿主单机进程重启
- **AND** 客户端继续使用旧 access token 访问受保护 API
- **THEN** 完整认证链查询`sys_online_session`时找不到对应 token_id
- **AND** 请求被拒绝
- **AND** 系统不得因为进程内 revoke 缓存丢失而放行该 token

#### Scenario: 服务重启后未超时 session 继续有效

- **WHEN** 宿主单机进程重启前`sys_online_session`中存在未超时 session
- **AND** 客户端使用签名有效且未过期的 access token 访问受保护 API
- **THEN** 完整认证链基于`sys_online_session`确认 session 存在且未超时
- **AND** 请求可以继续处理
- **AND** 系统刷新`last_active_time`时遵守既有写入节流策略

#### Scenario: 强制下线后 token 被拒绝

- **WHEN** 管理员强制下线某个在线 session
- **AND** 系统删除该 session 对应的`sys_online_session`记录
- **THEN** 客户端后续使用该 token 访问受保护 API 被拒绝
- **AND** 拒绝结果不依赖`sys_kv_cache`

## MODIFIED Requirements

### Requirement:认证中间件
系统 SHALL 提供认证中间件来保护需要登录才能访问的 API。中间件必须同时验证 JWT 签名、token kind、受控`clientType`、revoke 快速状态和宿主维护的会话有效性，不得依赖`linapro-monitor-online`是否安装。完整认证入口 MUST 校验`sys_online_session`中对应 token_id 是否存在、租户是否匹配且 session 未超时；低层 JWT 解析方法不得作为公开`Service`契约暴露，也不得被调用方单独当作完整登录态判断。

#### Scenario:在线用户插件未安装时访问受保护接口
- **WHEN** 请求头携带有效的`Authorization: Bearer <token>`，`sys_online_session`中存在对应的会话记录，且`linapro-monitor-online`未安装或未启用
- **THEN** 中间件仍通过 UPDATE 操作将会话的`last_active_time`更新为当前时间
- **AND** 请求正常处理，用户信息注入请求上下文

#### Scenario:缺少在线会话记录时拒绝访问
- **WHEN** 请求头携带签名有效且未过期的 access token
- **AND** `sys_online_session`中不存在该 token_id 对应记录
- **THEN** 中间件拒绝请求
- **AND** token 绑定的角色访问上下文被失效
- **AND** 系统不得仅凭 JWT 签名放行请求

#### Scenario:在线会话已超时时拒绝访问
- **WHEN** 请求头携带签名有效且未过期的 access token
- **AND** `sys_online_session.last_active_time`早于配置的会话超时时间窗口
- **THEN** 中间件拒绝请求
- **AND** 该 session 可由会话存储或清理任务自然删除

### Requirement: 集群模式 token 撤销必须使用 Redis
系统 SHALL 在集群模式下使用 Redis coordination KV 存储 JWT revoke 状态。revoked token key 的 TTL MUST 等于 token 剩余有效期。单机模式不得依赖`sys_kv_cache`存储 JWT revoke 状态，单机 revoke 仅作为进程内快速拒绝缓存；完整 session 有效性由`sys_online_session`校验兜底。

#### Scenario: 登出写入 Redis revoke
- **WHEN** 用户在集群模式下调用 logout
- **THEN** 系统写入 Redis revoke key
- **AND** key TTL 不超过 JWT 剩余有效期
- **AND** 后续使用该 token 的请求被所有节点拒绝

#### Scenario: 切换租户撤销旧 token
- **WHEN** 用户调用 switch-tenant 并获得新 token
- **THEN** 系统将旧 token ID 写入 Redis revoke store
- **AND** 旧 token 在所有节点立即不可用

#### Scenario: 单机登出不写 sys_kv_cache revoke
- **WHEN** 用户在单机模式下调用 logout
- **THEN** 系统删除对应`sys_online_session`记录
- **AND** 系统可写入进程内 revoke 快速拒绝缓存
- **AND** 系统不得要求`sys_kv_cache`存在才能完成登出

### Requirement: token 撤销读取失败必须 fail-closed
系统 SHALL 在集群模式认证链中读取 Redis revoke 状态。读取失败时 MUST fail-closed，不得仅凭 JWT 签名放行。单机模式读取进程内 revoke 状态不涉及外部 Redis 故障，但完整认证链仍 MUST 校验`sys_online_session`。

#### Scenario: Redis revoke 读取失败
- **WHEN** 请求携带签名有效的 JWT
- **AND** Redis revoke store 读取失败
- **THEN** 认证链拒绝请求
- **AND** 返回结构化认证失败或服务不可用错误

#### Scenario: 单机进程内 revoke 未命中但 session 缺失
- **WHEN** 请求携带签名有效的 JWT
- **AND** 单机进程内 revoke 缓存未命中
- **AND** `sys_online_session`中不存在对应 token_id
- **THEN** 认证链拒绝请求
- **AND** 系统不得把进程内 revoke 未命中解释为完整登录态有效

### Requirement: pre_token 必须使用 Redis 单次 TTL 状态

系统 SHALL 在集群模式下使用 Redis 存储`pre_token`、候选租户、用户会话`clientType`和 single-use 消费状态。`pre_token`MUST 短期有效且只能使用一次。单机模式 SHALL 使用宿主共享`kvcache`的`memory`后端保存`pre_token`短期状态，不得依赖`sys_kv_cache`。

#### Scenario: pre_token 首次选择租户
- **WHEN** 用户使用`clientType=mobile`完成密码验证并获得有效`pre_token`
- **AND** 用户使用该`pre_token`调用 select-tenant
- **THEN** 系统原子消费 Redis 中的`pre_token`
- **AND** 签发正式 JWT
- **AND** 正式 JWT 和在线会话的`clientType`均为`mobile`
- **AND** 后续同一`pre_token`不可再次使用

#### Scenario: pre_token 重放
- **WHEN** 客户端第二次使用同一`pre_token`
- **THEN** 系统拒绝请求
- **AND** 不签发正式 JWT

#### Scenario: 单机 pre_token 使用 memory TTL
- **WHEN** 宿主以单机模式运行且用户完成需要选择租户的密码验证
- **THEN** 系统将`pre_token`短期状态写入共享`memory`后端
- **AND** `pre_token`到期后由`memory`TTL 处理为不可用
- **AND** 系统不得要求`sys_kv_cache`存在

### Requirement: 认证短期状态必须集中使用 coordination KV
登录验证码、一次性 token、认证 challenge 或后续同类短 TTL 安全状态 SHALL 在集群模式下使用 coordination KV，在单机模式下使用宿主共享`kvcache`的`memory`后端。模块不得自行创建 Redis key、直接访问 Redis client、自行创建独立内存缓存实例或写入`sys_kv_cache`。

#### Scenario: 新增一次性认证 challenge
- **WHEN** 后续认证流程需要保存一次性 challenge
- **THEN** 实现通过宿主共享`kvcache.Service`写入带 TTL 的状态
- **AND** 集群模式下该服务背后使用 coordination KV
- **AND** 单机模式下该服务背后使用`memory`
- **AND** 认证模块不直接导入 Redis client 或自行创建独立内存缓存实例
