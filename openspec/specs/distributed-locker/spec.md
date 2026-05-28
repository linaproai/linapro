# 分布式锁规范

## Purpose
待定 - 由归档变更 distributed-locker 创建。归档后更新目的。
## Requirements
### Requirement: 分布式锁获取
系统 SHALL 提供分布式锁获取功能，支持基于唯一名称的锁抢占。

#### Scenario: 首次获取锁成功
- **WHEN** 节点尝试获取一个不存在的锁
- **THEN** 系统创建锁记录并返回锁实例

#### Scenario: 锁已被其他节点持有
- **WHEN** 节点尝试获取一个已被持有且未过期的锁
- **THEN** 系统返回失败，不创建锁实例

#### Scenario: 获取已过期的锁
- **WHEN** 节点尝试获取一个已过期的锁
- **THEN** 系统更新锁记录的过期时间和持有者，返回新的锁实例

### Requirement: 分布式锁释放
系统 SHALL 提供分布式锁释放功能，允许锁持有者主动释放锁。

#### Scenario: 释放自己持有的锁
- **WHEN** 锁持有者调用 Unlock 方法
- **THEN** 系统将锁的过期时间设置为当前时间，允许其他节点获取

#### Scenario: 释放非自己持有的锁
- **WHEN** 非锁持有者尝试释放锁
- **THEN** 系统不执行任何操作（或返回错误）

### Requirement: 租约续期
系统 SHALL 提供租约续期功能，允许锁持有者延长锁的有效期。

#### Scenario: 续期成功
- **WHEN** 锁持有者在锁未过期时调用 Renew 方法
- **THEN** 系统更新锁的过期时间为当前时间加上租约时长

#### Scenario: 续期失败
- **WHEN** 锁已被其他节点抢占或过期时调用 Renew 方法
- **THEN** 系统返回错误，表示续期失败

### Requirement: 锁状态检查
系统 SHALL 提供锁状态检查功能，判断当前节点是否持有特定锁。

#### Scenario: 检查自己持有的锁
- **WHEN** 锁持有者检查锁状态
- **THEN** 系统返回 true，表示当前节点持有该锁

#### Scenario: 检查其他节点持有的锁
- **WHEN** 非锁持有者检查锁状态
- **THEN** 系统返回 false，表示当前节点不持有该锁

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

### Requirement: 动态插件协调器集群模式必须使用 per-plugin 分布式锁

系统 SHALL 在 `cluster.enabled=true` 时，使用现有 distributed locker 或 coordination lock 为动态插件协调器提供 per-plugin 互斥。锁名称 MUST 包含稳定插件 ID，并且同一插件的生命周期 SQL、迁移账本、治理资源同步、发布状态切换和 runtime revision 发布 MUST 只由持锁节点执行。未获得锁的节点 MUST 跳过当前插件并等待后续 revision、event 或 safety sweep。

#### Scenario: 集群节点竞争同一插件协调锁
- **WHEN** 集群节点 A 和 B 同时尝试协调动态插件 P
- **THEN** 只有成功获取 P 的 per-plugin 分布式锁的节点执行 P 的共享生命周期副作用
- **AND** 未获得锁的节点不执行 P 的 SQL、菜单同步或发布状态写入

#### Scenario: 锁名称按插件隔离
- **WHEN** 节点同时协调动态插件 P 和 Q
- **THEN** P 的协调锁名称与 Q 的协调锁名称不同
- **AND** P 的锁竞争不得阻止 Q 在独立锁下收敛

#### Scenario: 单机模式不强制依赖 coordination lock
- **WHEN** `cluster.enabled=false`
- **THEN** 动态插件协调器可以使用进程内互斥或单机锁分支保护本节点并发
- **AND** 系统不得要求 Redis coordination lock 存在

