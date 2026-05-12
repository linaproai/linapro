## ADDED Requirements

### Requirement: 插件锁服务在集群模式下必须使用 coordination lock
系统 SHALL 在集群模式下通过 coordination lock 为插件 host service 提供分布式锁能力。插件不得获得 Redis client 或底层锁存储连接。

#### Scenario: 插件获取分布式锁
- **WHEN** 动态插件调用 host lock service 获取锁 `daily-sync`
- **THEN** 宿主构造带插件 ID 和租户维度的 lock name
- **AND** 通过 Redis lock store 获取锁
- **AND** 返回受控 host service 响应

#### Scenario: 插件释放非自己持有的锁
- **WHEN** 插件尝试释放 owner token 不匹配的锁
- **THEN** 宿主拒绝释放
- **AND** 不删除其他节点或其他插件持有的锁

### Requirement: 插件锁名称必须隔离插件和租户
插件锁 key SHALL 包含插件 ID 和租户维度。不同插件或不同租户使用相同逻辑锁名时 MUST 不互相阻塞，除非显式使用平台级共享锁能力并通过权限审计。

#### Scenario: 不同租户同名锁隔离
- **WHEN** 插件 P 在租户 A 获取锁 `sync`
- **AND** 插件 P 在租户 B 获取锁 `sync`
- **THEN** 两个锁互不冲突

#### Scenario: 不同插件同名锁隔离
- **WHEN** 插件 P1 获取锁 `sync`
- **AND** 插件 P2 获取锁 `sync`
- **THEN** 两个锁互不冲突

### Requirement: 插件锁故障必须返回明确错误
当 Redis lock store 不可用时，插件锁服务 SHALL 返回结构化错误。系统不得向插件报告锁获取成功。

#### Scenario: Redis 锁服务不可用
- **WHEN** 插件调用 lock acquire
- **AND** Redis 返回连接错误
- **THEN** host service 返回错误响应
- **AND** 插件不获得锁 handle

