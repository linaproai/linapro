# 插件缓存服务规范

## Purpose

定义动态插件受治理的宿主缓存访问，包括授权命名空间、有损缓存语义、后端/提供者抽象、关键修订号分离、原子递增和过期清理行为。
## Requirements
### Requirement:动态插件通过基于易失性缓存表的授权命名空间访问宿主分布式缓存

系统 SHALL 为动态插件提供受治理的缓存服务。插件只能通过宿主授权的命名缓存命名空间访问宿主通用 KV 缓存基础，不得直接接收本地缓存实现或其他低级缓存客户端。通用缓存模块 SHALL 通过后端/提供者抽象隐藏底层实现。当前单机默认后端为基于 PostgreSQL SQL 表的实现（包名 `sqltable`，常量 `BackendSQLTable`，字符串值 `"sql-table"`），并依赖应用层 TTL 与定时清理任务处理过期条目；集群模式使用 coordination KV backend。所有后端 SHALL 被视为有损缓存，不得作为权限、配置、插件稳定状态、缓存修订号或任何其他可靠业务状态的权威来源。

#### Scenario:插件访问授权的缓存命名空间

- **当** 插件调用缓存服务执行 `get`、`set`、`delete`、`incr` 或 `expire` 时
- **则** 宿主仅允许访问当前插件授权的 `host-cache` 资源
- **且** 宿主根据该缓存命名空间的命名规则和后端无关的 TTL 策略执行操作
- **且** 单机默认 `sqltable` 后端将缓存数据存储在共享 PostgreSQL 数据库的 `sys_kv_cache` 表中，而非宿主进程本地缓存

#### Scenario:数据库缓存表丢失后插件缓存作为未命中处理

- **当** `sys_kv_cache` 表内容因人为清理或重建数据库而不存在时
- **则** 插件缓存读取作为缓存未命中处理
- **且** 系统不得依赖 `sys_kv_cache` 恢复关键业务状态或缓存修订号

#### Scenario:插件写入超过字段长度限制的缓存值

- **当** 插件调用缓存服务写入超过命名空间、缓存键或缓存值长度限制的数据时
- **则** 宿主返回明确错误
- **且** 宿主不得截断写入
- **且** 宿主不得写入部分数据

#### Scenario:插件尝试访问未授权的缓存命名空间

- **当** 插件调用未授权的缓存命名空间时
- **则** 宿主拒绝调用
- **且** 宿主不向 guest 暴露底层缓存连接信息

### Requirement:插件缓存不得协调关键修订号

系统 SHALL 使用独立的持久化修订号机制来协调权限、配置和插件运行时等关键缓存域。不得在 `sys_kv_cache` 中存储这些域的共享修订号。

#### Scenario:发布关键缓存修订号

- **当** 权限、运行时配置或插件运行时关键缓存域发布修订号时
- **则** 系统写入持久化修订号存储
- **且** 系统不得将该关键缓存域修订号写入 `sys_kv_cache`

#### Scenario:插件缓存清除不影响关键缓存协调

- **当** `sys_kv_cache` 因数据库重启或缓存清理变空时
- **则** 已提交的关键缓存修订号仍可从持久化修订号存储读取
- **且** 节点仍可判断本地权限、配置和插件运行时缓存是否需要刷新

### Requirement:插件缓存递增在缓存存活期间必须是原子的

系统 SHALL 保证同一插件缓存键的 `incr` 在共享数据库和缓存表存活期间线性递增。数据库重建或缓存表内容被人为删除后，后续递增可能从新缓存值重新开始。`incr` 实现 SHALL 使用方言中性的 CAS 重试模式：读取当前整数快照；若行不存在，则用幂等插入初始化缺失整数行为 0 后重新读取；若现有行不是整数，则返回结构化错误且不修改原值；随后执行带 `value_int=<snapshot>` 条件的参数化 `UPDATE` 写入新值。实现不得使用任何数据库专用的原子自增技巧（如 MySQL `LAST_INSERT_ID(value_int + delta)`、`RETURNING` 等），不得为了原子递增修改 `sys_kv_cache` 表结构。遇到快照竞争、PostgreSQL serialization failure 或 deadlock 等可重试冲突时，系统 SHALL 对单次 `incr` 执行有限退避重试；超过重试上限后返回明确错误。

#### Scenario:多节点并发递增同一缓存键

- **当** 多个节点并发对同一插件缓存键执行 `incr` 时
- **则** 每次成功调用返回唯一的递增整数值
- **且** 最终缓存值等于初始值加上所有成功递增的总和
- **且** 任何节点不得通过读-修改-写竞争丢失递增

#### Scenario:并发首次递增缺失缓存键

- **当** 同一插件缓存键当前不存在且多个调用方并发执行 `incr(delta=1)` 时
- **则** 后端通过幂等插入确保存在 `value_kind=int, value_int=0` 的记录
- **且** 每个成功调用随后通过带 `value_int=<snapshot>` 条件的单条 CAS `UPDATE` 写入递增后的新值
- **且** 返回值从 1 开始线性递增，不出现重复值或丢失递增
- **且** 若数据库在并发首次初始化过程中返回可重试锁冲突，宿主在有限退避后重试该次递增

#### Scenario:首次递增使用 delta 作为初始结果

- **当** 插件对不存在的缓存键执行 `incr(delta=5)` 时
- **则** 宿主返回整数值 `5`
- **且** 数据库中保存的 `value_int` 为 `5`
- **且** 后续 `incr(delta=2)` 返回 `7`

#### Scenario:递增非整数缓存值

- **当** 插件对现有字符串缓存键执行 `incr` 时
- **则** 宿主返回结构化错误
- **且** 原始缓存值保持不变
- **且** 不执行 CAS `UPDATE` 修改该行

#### Scenario:incr 实现不依赖数据库专用原子技巧

- **当** 缓存服务执行 `incr` 操作时
- **则** SQL 执行路径不包含 `LAST_INSERT_ID(...)` / `RETURNING` / 其他数据库专用原子函数
- **且** 单次成功的 `incr` 操作通过读取快照、必要时的缺失键幂等插入、类型校验和单条 CAS `UPDATE` 完成

### Requirement:插件缓存过期清理必须避免热路径全表扫描

读取插件缓存时，系统 SHALL 仅执行只读查询。不得仅因缓存条目过期就在查询请求中删除数据。过期清理必须由后端在读取结果上的过期过滤和后台批量清理或写路径替换处理。`sqltable` 后端在 PostgreSQL 单机模式下必须提供后台批量清理能力，因为 `sys_kv_cache` 不支持自动过期淘汰。

#### Scenario:读取过期的缓存键

- **当** 插件读取过期的缓存键时
- **则** 宿主返回缓存未命中
- **且** 宿主不得在该查询请求中删除缓存行
- **且** 宿主不得要求此次读取为任何命名空间清理过期缓存

#### Scenario:后台批量清理移除过期缓存

- **当** 过期缓存批量清理任务触发时
- **则** 系统删除过期缓存行
- **且** `sqltable` 后端 SHALL 提供内置定时任务，每小时调用一次 `CleanupExpired`
- **且** 集群模式下，任务不得在多个节点间产生不受控的重复压力
- **且** 不需要外部过期清理的后端（如 Redis）可将 `CleanupExpired` 实现为空操作，无需投射 SQL 表清理任务

### Requirement: 集群模式插件缓存必须使用 coordination KV backend
系统 SHALL 在 `cluster.enabled=true` 时使用 coordination KV backend 承载 host/plugin KV cache。coordination KV backend MUST 通过统一 coordination provider 创建，不得由插件缓存服务自行创建 Redis client。

#### Scenario: 集群模式写入插件缓存
- **WHEN** 插件在集群模式下调用 cache set
- **THEN** 系统将值写入 coordination KV backend
- **AND** key 包含租户、owner type、owner key、namespace 和 logical key
- **AND** 不写入 `sys_kv_cache` 作为集群 KV cache 主实现

#### Scenario: 单机模式继续使用 SQL table backend
- **WHEN** `cluster.enabled=false`
- **THEN** 插件缓存可继续使用 SQL table backend
- **AND** 不要求 coordination KV backend 存在

### Requirement: coordination KV 插件缓存必须使用后端原生 TTL
coordination KV backend SHALL 使用后端原生 TTL 处理缓存过期。coordination KV backend MUST 返回 `RequiresExpiredCleanup=false`，并且不得注册 SQL 过期清理任务。当前集群 coordination backend 为 Redis 时，该 TTL 由 Redis 原生过期能力负责。

#### Scenario: coordination KV TTL 到期
- **WHEN** 插件写入 TTL 为 `5s` 的缓存值
- **AND** 5 秒后读取该 key
- **THEN** 返回缓存未命中
- **AND** 不需要后台 SQL cleanup 才能过期

#### Scenario: coordination KV backend 不注册 KV cleanup job
- **WHEN** 宿主以集群模式和 coordination KV backend 启动
- **THEN** 内置 `host:kvcache-cleanup-expired` 不因 coordination KV backend 注册
- **AND** Redis 过期由 Redis 自身负责

### Requirement: coordination KV 插件缓存递增必须原子
coordination KV backend SHALL 使用 coordination KV 原子递增能力实现 `incr`。并发成功递增不得丢失。

#### Scenario: 多节点并发 coordination KV incr
- **WHEN** 多个节点并发对同一插件缓存 key 执行 `incr(delta=1)`
- **THEN** 每次成功调用返回唯一整数
- **AND** 最终值等于成功调用次数

#### Scenario: 递增字符串值
- **WHEN** 插件对已有字符串缓存值执行 `incr`
- **THEN** 系统返回结构化类型错误
- **AND** 原始字符串值不被修改

### Requirement: 插件缓存仍为有损缓存
无论 backend 是 Redis 还是 SQL table，插件缓存 SHALL 仍被视为有损缓存。系统 MUST 不依赖插件缓存作为权限、配置、插件稳定状态或业务权威数据源。

#### Scenario: Redis key 被清理
- **WHEN** Redis 中某插件缓存 key 被 TTL 或运维清理移除
- **THEN** 插件读取该 key 返回缓存未命中
- **AND** 系统不得因此丢失权威业务状态

### Requirement: coordination KV 插件缓存故障不得伪装写入成功
当 coordination KV backend 写入、删除、递增或过期操作失败时，系统 SHALL 返回结构化错误。系统 MUST 不得在 coordination KV 写失败时向插件报告成功。

#### Scenario: coordination KV 写失败
- **WHEN** 插件调用 cache set
- **AND** coordination KV 返回连接错误
- **THEN** host service 返回错误响应
- **AND** 插件可根据错误决定重试或降级

### Requirement: 源码插件必须通过插件作用域缓存 facade 使用宿主 KV cache

系统 SHALL 通过源码插件 `HostServices` 服务目录提供受治理的缓存 facade。源码插件只能通过插件可见的 `namespace`、逻辑 `key` 和 TTL 使用缓存，不得接收宿主内部 `kvcache.Service`、`OwnerType`、编码后的 owner key、coordination KV、Redis client、SQL table backend 或 provider。

#### Scenario: 源码插件通过 HostServices 获取缓存服务

- **WHEN** 源码插件在 HTTP registrar、Cron registrar 或 hook payload 中访问 `HostServices().Cache()`
- **THEN** 系统返回当前插件作用域绑定的缓存服务
- **AND** 该服务仅接受插件可见的 `namespace`、逻辑 `key`、缓存值和 TTL 参数
- **AND** 该服务不暴露内部缓存 backend、owner type 或底层客户端

#### Scenario: 源码插件缓存服务缺失

- **WHEN** 源码插件调用路径未配置缓存服务
- **THEN** 系统不得在调用路径中临时创建新的宿主缓存服务图
- **AND** 调用方必须获得明确错误或 nil 服务并由插件代码显式处理

### Requirement: 源码插件缓存 key 必须由宿主按插件和租户作用域生成

系统 SHALL 在源码插件缓存 facade 内部根据当前 `pluginID`、`namespace`、逻辑 `key` 和当前租户上下文生成内部缓存 key。源码插件 MUST NOT 传入或覆盖 `pluginID`、owner key、owner type 或租户 key。

#### Scenario: 同一命名空间下不同源码插件缓存隔离

- **WHEN** 源码插件 `plugin-a` 和 `plugin-b` 都写入 `namespace=profile` 且 `key=current`
- **THEN** 系统为两个插件生成不同的内部缓存 key
- **AND** `plugin-a` 读取不到 `plugin-b` 的缓存值
- **AND** `plugin-b` 读取不到 `plugin-a` 的缓存值

#### Scenario: 当前租户上下文下写入源码插件缓存

- **WHEN** 源码插件在租户 `1001` 的请求上下文中写入缓存
- **THEN** 系统生成包含租户 `1001`、插件 ID、命名空间和逻辑 key 的内部缓存 key
- **AND** 其他租户上下文读取同一插件、同一命名空间和同一逻辑 key 时不得命中该租户缓存

#### Scenario: 无租户上下文下写入源码插件缓存

- **WHEN** 源码插件在无租户上下文的启动期、平台级任务或测试调用中写入缓存
- **THEN** 系统生成平台级插件缓存 key
- **AND** 该 key 仍必须包含插件 ID、命名空间和逻辑 key

### Requirement: 源码插件缓存必须复用宿主启动期注入的共享缓存服务

系统 SHALL 将 HTTP 启动期创建的共享 `kvCacheSvc` 注入源码插件缓存 facade。源码插件缓存 facade MUST 复用该共享实例或其共享 backend，不得在插件注册、请求处理、hook 回调、cron 回调或缓存方法调用路径中调用 `kvcache.New()` 创建独立缓存服务图。

#### Scenario: 单机模式源码插件缓存使用单机后端

- **WHEN** `cluster.enabled=false` 且源码插件调用缓存 `set`
- **THEN** 源码插件缓存 facade 通过启动期注入的共享 `kvCacheSvc` 执行写入
- **AND** 系统可使用 SQL table backend 或宿主单机缓存策略
- **AND** 不要求 coordination KV backend 存在

#### Scenario: 集群模式源码插件缓存使用 coordination KV backend

- **WHEN** `cluster.enabled=true` 且源码插件调用缓存 `set`
- **THEN** 源码插件缓存 facade 通过启动期注入的共享 `kvCacheSvc` 执行写入
- **AND** 该共享服务使用宿主统一 coordination provider 背后的 coordination KV backend
- **AND** 源码插件缓存 facade 不自行解析 Redis 配置或创建 Redis client

### Requirement: 源码插件缓存操作必须保持有损缓存和 TTL 语义

系统 SHALL 将源码插件缓存视为有损缓存。源码插件缓存 MUST NOT 被用作权限、配置、插件稳定状态、租户隔离、业务权威数据、关键缓存修订号或跨实例一致性协调的事实源。源码插件缓存 TTL MUST 使用 `time.Duration` 语义表达，负 TTL 必须返回明确错误。

#### Scenario: 源码插件读取不存在或已过期的缓存

- **WHEN** 源码插件读取不存在或已过期的缓存 key
- **THEN** 系统返回缓存未命中
- **AND** 系统不得要求调用方把该缓存值视为权威业务状态

#### Scenario: 源码插件设置带 TTL 的缓存

- **WHEN** 源码插件写入缓存值并传入正数 TTL
- **THEN** 系统按后端无关的 TTL 语义设置过期时间
- **AND** 单机 SQL table backend 通过既有过期判断和清理任务处理过期
- **AND** 集群 coordination KV backend 通过后端原生 TTL 处理过期

#### Scenario: 源码插件传入负 TTL

- **WHEN** 源码插件调用 `set`、`incr` 或 `expire` 并传入负 TTL
- **THEN** 系统返回明确错误
- **AND** 系统不得写入或修改对应缓存值

### Requirement: 源码插件缓存写入失败不得伪装成功

系统 SHALL 在源码插件缓存写入、删除、递增或过期操作失败时返回错误。系统 MUST NOT 在共享缓存 backend、coordination KV、SQL table 或 key 校验失败时向源码插件报告成功。

#### Scenario: 源码插件缓存写入 backend 失败

- **WHEN** 源码插件调用 `set`
- **AND** 共享缓存 backend 返回连接、校验或持久化错误
- **THEN** 源码插件缓存 facade 返回错误
- **AND** 系统不得向插件返回成功写入的缓存快照

#### Scenario: 源码插件递增字符串缓存值

- **WHEN** 源码插件对现有字符串缓存值执行 `incr`
- **THEN** 系统返回结构化类型错误
- **AND** 原始字符串值不得被修改

