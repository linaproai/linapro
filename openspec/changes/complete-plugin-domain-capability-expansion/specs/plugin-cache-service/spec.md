## ADDED Requirements

### Requirement: Cache 必须支持插件私有多键读写删
系统 SHALL 提供`Cache.GetMany`、`Cache.SetMany`、`Cache.DeleteMany`以及动态`cache.get_many`、`cache.set_many`、`cache.delete_many`。这些方法 MUST 复用既有缓存后端、插件命名空间、租户作用域、TTL 语义和错误行为。

#### Scenario: 多键读取缓存
- **WHEN** 插件批量读取多个缓存 key
- **THEN** 系统返回当前插件命名空间下的 key/value 结果和缺失集合
- **AND** 不暴露其他插件或其他租户的缓存 key

#### Scenario: 多键写入缓存
- **WHEN** 插件批量写入多个缓存 key 和 TTL
- **THEN** 系统复用既有缓存后端写入
- **AND** 超过 key 数量、key 长度或总字节数上限时返回结构化超限错误

#### Scenario: 多键删除缓存
- **WHEN** 插件批量删除多个缓存 key
- **THEN** 系统只删除当前插件命名空间下的 key
- **AND** 删除缺失 key 不得泄露其他命名空间存在性

### Requirement: Cache 批量能力不得成为权威状态来源
系统 SHALL 明确 cache 批量能力仅用于插件私有非权威运行时加速数据。插件权限、租户、配置、插件状态、业务记录或审计状态 MUST NOT 以 cache 批量能力作为权威来源。

#### Scenario: 缓存后端不可用
- **WHEN** cache 多键方法发现共享缓存后端不可用
- **THEN** 系统返回结构化错误或沿用单键方法的降级语义
- **AND** 不得在集群模式下临时创建仅当前节点可见的默认缓存实例
