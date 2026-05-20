## ADDED Requirements

### Requirement: 集群模式 token 撤销必须使用 Redis
系统 SHALL 在集群模式下使用 Redis coordination KV 存储 JWT revoke 状态。revoked token key 的 TTL MUST 等于 token 剩余有效期。

#### Scenario: 登出写入 Redis revoke
- **WHEN** 用户在集群模式下调用 logout
- **THEN** 系统写入 Redis revoke key
- **AND** key TTL 不超过 JWT 剩余有效期
- **AND** 后续使用该 token 的请求被拒绝

#### Scenario: 切换租户撤销旧 token
- **WHEN** 用户调用 switch-tenant 并获得新 token
- **THEN** 系统将旧 token ID 写入 Redis revoke store
- **AND** 旧 token 在所有节点立即不可用

### Requirement: token 撤销读取失败必须 fail-closed
系统 SHALL 在集群模式认证链中读取 Redis revoke 状态。读取失败时 MUST fail-closed，不得仅凭 JWT 签名放行。

#### Scenario: Redis revoke 读取失败
- **WHEN** 请求携带签名有效的 JWT
- **AND** Redis revoke store 读取失败
- **THEN** 认证链拒绝请求
- **AND** 返回结构化认证失败或服务不可用错误

### Requirement: pre_token 必须使用 Redis 单次 TTL 状态
系统 SHALL 在集群模式下使用 Redis 存储 `pre_token`、候选租户和 single-use 消费状态。`pre_token` MUST 短期有效且只能使用一次。

#### Scenario: pre_token 首次选择租户
- **WHEN** 用户使用有效 `pre_token` 调用 select-tenant
- **THEN** 系统原子消费 Redis 中的 `pre_token`
- **AND** 签发正式 JWT
- **AND** 后续同一 `pre_token` 不可再次使用

#### Scenario: pre_token 重放
- **WHEN** 客户端第二次使用同一 `pre_token`
- **THEN** 系统拒绝请求
- **AND** 不签发正式 JWT

### Requirement: 认证短期状态必须集中使用 coordination KV
登录验证码、一次性 token、认证 challenge 或后续同类短 TTL 安全状态 SHALL 在集群模式下使用 coordination KV。模块不得自行创建 Redis key 或直接访问 Redis client。

#### Scenario: 新增一次性认证 challenge
- **WHEN** 后续认证流程需要保存一次性 challenge
- **THEN** 实现通过 coordination KV 写入带 TTL 的状态
- **AND** 不直接导入 Redis client

