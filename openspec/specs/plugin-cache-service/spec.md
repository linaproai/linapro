# plugin-cache-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
## Requirements
### Requirement: 动态插件通过命名缓存空间访问基于 MEMORY 表的宿主分布式缓存

系统 SHALL 为动态插件提供受治理的缓存服务，插件只能通过宿主授权的命名缓存空间访问基于 MySQL `MEMORY` 表的宿主分布式缓存，而不能直接获取本机缓存实现或其他底层缓存客户端。

#### Scenario: 插件访问授权缓存空间

- **WHEN** 插件调用缓存服务执行`get`、`set`、`delete`、`incr`或`expire`
- **THEN** 宿主仅允许访问当前插件已授权的`host-cache`资源
- **AND** 宿主按该缓存空间的命名规则和 TTL 策略执行操作
- **AND** 宿主将缓存数据落到共享数据库中的 `MEMORY` 缓存表，而不是宿主进程内本机缓存

#### Scenario: 插件写入超出字段长度限制的缓存值

- **WHEN** 插件调用缓存服务写入超出命名空间、缓存键或缓存值长度上限的数据
- **THEN** 宿主返回显式错误
- **AND** 宿主不得截断写入
- **AND** 宿主不得写入任何部分数据

#### Scenario: 插件尝试访问未授权缓存空间

- **WHEN** 插件调用一个未授权的缓存空间
- **THEN** 宿主拒绝该调用
- **AND** 宿主不向 guest 暴露底层缓存连接信息

