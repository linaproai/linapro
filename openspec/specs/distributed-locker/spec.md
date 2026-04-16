# distributed-locker Specification

## Purpose
TBD - created by archiving change distributed-locker. Update Purpose after archive.
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

