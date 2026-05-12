## ADDED Requirements

### Requirement: 集群模式分布式锁必须使用 coordination lock
系统 SHALL 在 `cluster.enabled=true` 时使用 coordination provider 的 lock store 实现分布式锁。当前 Redis provider MUST 使用带 TTL 的原子锁，不得使用 PostgreSQL `sys_locker` 作为集群锁事实源。

#### Scenario: 集群模式获取锁
- **WHEN** 集群模式节点获取分布式锁 `plugin:demo:install`
- **THEN** 系统通过 Redis lock store 创建带 TTL 的锁
- **AND** 返回包含 owner token 的 lock instance
- **AND** 不写入 `sys_locker` 作为该锁的事实源

#### Scenario: 单机模式保持轻量锁语义
- **WHEN** `cluster.enabled=false`
- **THEN** 系统可继续使用本地或现有 SQL locker 实现
- **AND** 不要求 Redis lock store 存在

### Requirement: 锁释放必须校验 owner token
系统 SHALL 在释放 Redis 锁时校验 owner token。非持有者、过期 handle 或旧 owner token MUST 不得删除新持有者的锁。

#### Scenario: 旧 handle 释放失败
- **WHEN** 节点 A 获取锁后租约过期
- **AND** 节点 B 获取同名锁
- **AND** 节点 A 使用旧 handle 调用 Unlock
- **THEN** 节点 B 的锁仍然存在
- **AND** 节点 A 收到未持有锁或等价错误

#### Scenario: 持有者释放成功
- **WHEN** 锁持有者使用当前 owner token 调用 Unlock
- **THEN** Redis 中对应锁 key 被删除
- **AND** 其他节点可随后获取该锁

### Requirement: 锁续约必须校验 owner token
系统 SHALL 在续约 Redis 锁时校验 owner token。只有当前持有者可以延长租约。

#### Scenario: 持有者续约成功
- **WHEN** 锁持有者在 TTL 到期前调用 Renew
- **THEN** 系统延长锁 TTL
- **AND** lock instance 继续有效

#### Scenario: 非持有者续约失败
- **WHEN** 锁已被其他节点持有
- **AND** 旧持有者调用 Renew
- **THEN** 系统返回续约失败
- **AND** 不修改当前锁 TTL

### Requirement: 分布式锁应预留 fencing token
coordination lock store SHALL 在接口层预留 fencing token，用于后续需要防止旧 leader 写入的场景。Redis provider MAY 通过独立递增计数器生成 fencing token。

#### Scenario: 获取锁返回 fencing token
- **WHEN** 节点成功获取支持 fencing 的锁
- **THEN** lock instance 包含单调递增 fencing token
- **AND** 后续需要严格防旧写入的模块可记录该 token

