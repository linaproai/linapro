## ADDED Requirements

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
