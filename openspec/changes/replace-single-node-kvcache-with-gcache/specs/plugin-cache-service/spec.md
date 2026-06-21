## REMOVED Requirements

### Requirement:动态插件通过基于易失性缓存表的授权命名空间访问宿主分布式缓存

**Reason**: 单机模式不再使用`sys_kv_cache`数据库表承载缓存。缓存表语义会把有损缓存变成数据库热路径，并要求额外 SQL 清理任务；集群模式已经由 Redis coordination KV 提供跨实例缓存能力。

**Migration**: 动态插件继续通过授权的 cache host service 访问缓存，不得直接访问`sys_kv_cache`。宿主单机模式迁移到`memory`进程内后端；集群模式继续使用 coordination KV backend。

## ADDED Requirements

### Requirement:动态插件通过授权命名空间访问宿主有损缓存

系统 SHALL 为动态插件提供受治理的缓存服务。插件只能通过宿主授权的命名缓存命名空间访问宿主通用 KV 缓存基础，不得直接接收本地缓存实现、Redis client、底层进程内缓存实例或其他低级缓存客户端。通用缓存模块 SHALL 通过后端/提供者抽象隐藏底层实现。单机默认后端 SHALL 为`memory`进程内缓存；集群模式 SHALL 使用 coordination KV backend。所有后端 SHALL 被视为有损缓存，不得作为权限、配置、插件稳定状态、缓存修订号或任何其他可靠业务状态的权威来源。

#### Scenario:插件访问授权的缓存命名空间

- **WHEN** 插件调用缓存服务执行`get`、`set`、`delete`、`incr`或`expire`
- **THEN** 宿主仅允许访问当前插件授权的`host-cache`资源
- **AND** 宿主根据该缓存命名空间的命名规则和后端无关的 TTL 策略执行操作
- **AND** 单机默认`memory`后端将缓存数据存储在宿主进程内存中，不写入`sys_kv_cache`

#### Scenario:单机进程重启后插件缓存作为未命中处理

- **WHEN** 宿主以单机模式重启且进程内`memory`内容丢失
- **THEN** 插件缓存读取作为缓存未命中处理
- **AND** 系统不得依赖进程内缓存恢复关键业务状态或缓存修订号

#### Scenario:插件写入超过字段长度限制的缓存值

- **WHEN** 插件调用缓存服务写入超过命名空间、缓存键或缓存值长度限制的数据
- **THEN** 宿主返回明确错误
- **AND** 宿主不得截断写入
- **AND** 宿主不得写入部分数据

#### Scenario:插件尝试访问未授权的缓存命名空间

- **WHEN** 插件调用未授权的缓存命名空间
- **THEN** 宿主拒绝调用
- **AND** 宿主不向 guest 暴露底层缓存连接信息

## MODIFIED Requirements

### Requirement:插件缓存不得协调关键修订号

系统 SHALL 使用独立的持久化修订号机制来协调权限、配置和插件运行时等关键缓存域。不得在`sys_kv_cache`、进程内`memory`插件缓存或普通 plugin cache namespace 中存储这些域的共享修订号。

#### Scenario:发布关键缓存修订号

- **WHEN** 权限、运行时配置或插件运行时关键缓存域发布修订号
- **THEN** 系统写入持久化修订号存储或集群 coordination revision store
- **AND** 系统不得将该关键缓存域修订号写入`sys_kv_cache`、进程内`memory`插件缓存或普通 plugin cache namespace

#### Scenario:插件缓存清除不影响关键缓存协调

- **WHEN** 单机进程重启、`memory`被清空、Redis key 过期或运维清理插件缓存
- **THEN** 已提交的关键缓存修订号仍可从其权威修订号存储读取
- **AND** 节点仍可判断本地权限、配置和插件运行时缓存是否需要刷新

### Requirement:插件缓存递增在缓存存活期间必须是原子的

系统 SHALL 保证同一插件缓存键的`incr`在当前缓存后端的存活窗口内线性递增。单机`memory`后端 MUST 在同一宿主进程内保证并发成功递增不丢失；进程重启后，后续递增 MAY 从新缓存值重新开始。集群 coordination KV backend MUST 使用后端原子递增能力保证跨节点并发成功递增不丢失。`incr`不得被用作业务权威计数。

#### Scenario:单机进程内并发递增同一缓存键

- **WHEN** 宿主以单机模式运行且多个 goroutine 并发对同一插件缓存键执行`incr(delta=1)`
- **THEN** 每次成功调用返回唯一的递增整数值
- **AND** 最终缓存值等于初始值加上所有成功递增的总和
- **AND** 宿主不得通过读-修改-写竞争丢失递增

#### Scenario:集群多节点并发递增同一缓存键

- **WHEN** 宿主以集群模式运行且多个节点并发对同一插件缓存键执行`incr(delta=1)`
- **THEN** 每次成功调用返回唯一的递增整数值
- **AND** 最终缓存值等于初始值加上所有成功递增的总和
- **AND** 任一节点不得通过本地进程缓存完成该递增

#### Scenario:首次递增使用 delta 作为初始结果

- **WHEN** 插件对不存在的缓存键执行`incr(delta=5)`
- **THEN** 宿主返回整数值`5`
- **AND** 缓存后端保存的整数值为`5`
- **AND** 后续`incr(delta=2)`返回`7`

#### Scenario:递增非整数缓存值

- **WHEN** 插件对现有字符串缓存键执行`incr`
- **THEN** 宿主返回结构化错误
- **AND** 原始缓存值保持不变
- **AND** 宿主不得把字符串值隐式转换为整数

#### Scenario:单机进程重启后递增重新开始

- **WHEN** 宿主以单机模式重启且原`memory`整数缓存已丢失
- **THEN** 插件再次对同一缓存键执行`incr(delta=1)`可返回`1`
- **AND** 系统不得把该缓存递增值解释为业务权威计数

### Requirement:插件缓存过期清理必须避免热路径全表扫描

读取插件缓存时，系统 SHALL 仅执行只读查询或后端内存读取。不得仅因缓存条目过期就在查询请求中执行数据库删除。过期清理必须由后端在读取结果上的过期过滤、后端原生 TTL、进程内淘汰或写路径替换处理。单机`memory`后端和集群 coordination KV backend SHALL 使用后端 TTL 语义处理过期，并返回`RequiresExpiredCleanup=false`；宿主不得为这些后端注册`host:kvcache-cleanup-expired`内置定时任务。

#### Scenario:读取过期的缓存键

- **WHEN** 插件读取过期的缓存键
- **THEN** 宿主返回缓存未命中
- **AND** 宿主不得在该查询请求中删除数据库缓存行
- **AND** 宿主不得要求此次读取为任何命名空间执行全表过期清理

#### Scenario:单机 memory TTL 到期

- **WHEN** 宿主以单机模式写入 TTL 为`5s`的插件缓存值
- **AND** 5 秒后读取该 key
- **THEN** 返回缓存未命中
- **AND** 不需要后台 SQL cleanup 才能过期

#### Scenario:不注册 KV cleanup job

- **WHEN** 宿主使用单机`memory`后端或集群 coordination KV backend 启动
- **THEN** 内置`host:kvcache-cleanup-expired`不注册
- **AND** 默认作业清单不投射 KV cache SQL 过期清理任务

### Requirement: 集群模式插件缓存必须使用 coordination KV backend
系统 SHALL 在`cluster.enabled=true`时使用 coordination KV backend 承载 host/plugin KV cache。coordination KV backend MUST 通过统一 coordination provider 创建，不得由插件缓存服务自行创建 Redis client。

#### Scenario: 集群模式写入插件缓存
- **WHEN** 插件在集群模式下调用 cache set
- **THEN** 系统将值写入 coordination KV backend
- **AND** key 包含租户、owner type、owner key、namespace 和 logical key
- **AND** 不写入`sys_kv_cache`作为集群 KV cache 主实现

#### Scenario: 单机模式使用 memory backend
- **WHEN** `cluster.enabled=false`
- **THEN** 插件缓存使用`memory`进程内 backend
- **AND** 不要求 coordination KV backend 存在
- **AND** 不要求`sys_kv_cache`数据库表存在

### Requirement: coordination KV 插件缓存必须使用后端原生 TTL
coordination KV backend SHALL 使用后端原生 TTL 处理缓存过期。coordination KV backend MUST 返回`RequiresExpiredCleanup=false`，并且不得注册 SQL 过期清理任务。当前集群 coordination backend 为 Redis 时，该 TTL 由 Redis 原生过期能力负责。

#### Scenario: coordination KV TTL 到期
- **WHEN** 插件写入 TTL 为`5s`的缓存值
- **AND** 5 秒后读取该 key
- **THEN** 返回缓存未命中
- **AND** 不需要后台 SQL cleanup 才能过期

#### Scenario: coordination KV backend 不注册 KV cleanup job
- **WHEN** 宿主以集群模式和 coordination KV backend 启动
- **THEN** 内置`host:kvcache-cleanup-expired`不因 coordination KV backend 注册
- **AND** Redis 过期由 Redis 自身负责

### Requirement: coordination KV 插件缓存递增必须原子
coordination KV backend SHALL 使用 coordination KV 原子递增能力实现`incr`。并发成功递增不得丢失。

#### Scenario: 多节点并发 coordination KV incr
- **WHEN** 多个节点并发对同一插件缓存 key 执行`incr(delta=1)`
- **THEN** 每次成功调用返回唯一整数
- **AND** 最终值等于成功调用次数

#### Scenario: 递增字符串值
- **WHEN** 插件对已有字符串缓存值执行`incr`
- **THEN** 系统返回结构化类型错误
- **AND** 原始字符串值不被修改

### Requirement: 插件缓存仍为有损缓存
无论 backend 是 Redis coordination KV 还是单机`memory`，插件缓存 SHALL 仍被视为有损缓存。系统 MUST 不依赖插件缓存作为权限、配置、插件稳定状态或业务权威数据源。

#### Scenario: Redis key 被清理
- **WHEN** Redis 中某插件缓存 key 被 TTL 或运维清理移除
- **THEN** 插件读取该 key 返回缓存未命中
- **AND** 系统不得因此丢失权威业务状态

#### Scenario: memory key 被清理
- **WHEN** 单机进程内某插件缓存 key 被 TTL 或进程重启移除
- **THEN** 插件读取该 key 返回缓存未命中
- **AND** 系统不得因此丢失权威业务状态

### Requirement: coordination KV 插件缓存故障不得伪装写入成功
当 coordination KV backend 写入、删除、递增或过期操作失败时，系统 SHALL 返回结构化错误。系统 MUST 不得在 coordination KV 写失败时向插件报告成功。

#### Scenario: coordination KV 写失败
- **WHEN** 插件调用 cache set
- **AND** coordination KV 返回连接错误
- **THEN** host service 返回错误响应
- **AND** 插件可根据错误决定重试或降级

### Requirement: 源码插件缓存必须复用宿主启动期注入的共享缓存服务

系统 SHALL 将 HTTP 启动期创建的共享`kvCacheSvc`注入源码插件缓存 facade。源码插件缓存 facade MUST 复用该共享实例或其共享 backend，不得在插件注册、请求处理、hook 回调、cron 回调或缓存方法调用路径中调用`kvcache.New()`创建独立缓存服务图。

#### Scenario: 单机模式源码插件缓存使用 memory 后端

- **WHEN** `cluster.enabled=false`且源码插件调用缓存`set`
- **THEN** 源码插件缓存 facade 通过启动期注入的共享`kvCacheSvc`执行写入
- **AND** 该共享服务使用宿主单机`memory`后端
- **AND** 不要求 coordination KV backend 或`sys_kv_cache`存在

#### Scenario: 集群模式源码插件缓存使用 coordination KV backend

- **WHEN** `cluster.enabled=true`且源码插件调用缓存`set`
- **THEN** 源码插件缓存 facade 通过启动期注入的共享`kvCacheSvc`执行写入
- **AND** 该共享服务使用宿主统一 coordination provider 背后的 coordination KV backend
- **AND** 源码插件缓存 facade 不自行解析 Redis 配置或创建 Redis client

### Requirement: 源码插件缓存操作必须保持有损缓存和 TTL 语义

系统 SHALL 将源码插件缓存视为有损缓存。源码插件缓存 MUST NOT 被用作权限、配置、插件稳定状态、租户隔离、业务权威数据、关键缓存修订号或跨实例一致性协调的事实源。源码插件缓存 TTL MUST 使用`time.Duration`语义表达，`set`、`incr`和`expire`必须传入正 TTL，零值或负 TTL 必须返回明确错误；单机`memory`后端不得通过任意固定 LRU 容量替代缓存条目的过期生命周期。

#### Scenario: 源码插件读取不存在或已过期的缓存

- **WHEN** 源码插件读取不存在或已过期的缓存 key
- **THEN** 系统返回缓存未命中
- **AND** 系统不得要求调用方把该缓存值视为权威业务状态

#### Scenario: 源码插件设置带 TTL 的缓存

- **WHEN** 源码插件写入缓存值并传入正数 TTL
- **THEN** 系统按后端无关的 TTL 语义设置过期时间
- **AND** 单机`memory`后端通过进程内 TTL 处理过期
- **AND** 集群 coordination KV backend 通过后端原生 TTL 处理过期

#### Scenario: 源码插件传入非正 TTL

- **WHEN** 源码插件调用`set`、`incr`或`expire`并传入零值或负 TTL
- **THEN** 系统返回明确错误
- **AND** 系统不得写入或修改对应缓存值

#### Scenario: memory 后端不使用固定 LRU 容量作为生命周期兜底

- **WHEN** 宿主以单机模式创建`memory`后端
- **THEN** 后端不设置任意固定 LRU 容量作为缓存过期语义
- **AND** 缓存条目的生命周期由调用方传入的正 TTL 和后端 TTL 机制决定
- **AND** 系统不得接受永不过期插件缓存写入

### Requirement: 源码插件缓存写入失败不得伪装成功

系统 SHALL 在源码插件缓存写入、删除、递增或过期操作失败时返回错误。系统 MUST NOT 在共享缓存 backend、coordination KV、`memory`或 key 校验失败时向源码插件报告成功。

#### Scenario: 源码插件缓存写入 backend 失败

- **WHEN** 源码插件调用`set`
- **AND** 共享缓存 backend 返回连接、校验或运行时错误
- **THEN** 源码插件缓存 facade 返回错误
- **AND** 系统不得向插件返回成功写入的缓存快照

#### Scenario: 源码插件递增字符串缓存值

- **WHEN** 源码插件对现有字符串缓存值执行`incr`
- **THEN** 系统返回结构化类型错误
- **AND** 原始字符串值不得被修改

### Requirement: 源码插件和动态插件缓存必须共享缓存语义

系统 SHALL 要求源码插件和动态插件使用同一组插件缓存语义，包括插件 ID 隔离、租户上下文隔离、有损缓存定位、TTL、原子递增、单机`memory`后端、集群 coordination KV backend 和写入失败返回错误。

#### Scenario: 源码插件和动态插件同名缓存隔离

- **WHEN** 源码插件和动态插件使用相同`namespace`和 logical`key`
- **THEN** 宿主按各自插件 ID 生成不同内部 cache key
- **AND** 两个插件不得互相读取缓存值

#### Scenario: 单机模式动态插件缓存使用共享后端

- **WHEN** `cluster.enabled=false`且动态插件写入缓存
- **THEN** 动态插件缓存通过启动期共享`kvcache.Service`背后的`memory`后端写入
- **AND** 分发器不得自行创建独立内存缓存实例、Redis client 或 SQL table backend

#### Scenario: 集群模式动态插件缓存使用共享后端

- **WHEN** `cluster.enabled=true`且动态插件写入缓存
- **THEN** 动态插件缓存通过启动期共享`kvcache.Service`背后的 coordination KV backend 写入
- **AND** 分发器不得自行创建 Redis client 或 SQL table backend

### Requirement: 宿主共享 kvcache 服务必须显式选择拓扑后端

系统 SHALL 在 HTTP 启动期显式创建宿主共享`kvcache.Service`。单机模式使用`memory`provider；集群模式使用 coordination KV provider。该共享服务 MUST 被注入源码插件缓存 facade、动态插件 cache host service 和其他宿主插件缓存调用路径；这些路径不得各自调用`kvcache.New()`并依赖进程默认 provider。

#### Scenario: 单机模式显式使用 memory provider

- **WHEN** `cluster.enabled=false`且宿主启动创建共享`kvcache.Service`
- **THEN** 启动装配使用`kvcache.NewMemoryProvider()`或等价`memory`provider
- **AND** 不要求 coordination KV backend 存在
- **AND** 不要求`sys_kv_cache`数据库表存在

#### Scenario: 集群模式显式使用 coordination KV provider

- **WHEN** `cluster.enabled=true`且 coordination 服务已初始化
- **THEN** 启动装配使用`kvcache.NewCoordinationKVProvider(coordinationSvc)`或等价 coordination KV provider
- **AND** cache host service 写入、删除、递增和过期操作使用 coordination KV backend

#### Scenario: 缺失集群 coordination 依赖时不退回默认后端

- **WHEN** `cluster.enabled=true`但共享 coordination KV provider 无法创建
- **THEN** 启动或配置入口返回明确错误
- **AND** 系统不得静默退回`memory`provider、SQL table provider 或包级默认 provider
