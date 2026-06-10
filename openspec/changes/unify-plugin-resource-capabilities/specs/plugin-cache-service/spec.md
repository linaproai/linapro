## ADDED Requirements

### Requirement: 动态插件缓存必须通过 cachecap 领域契约调用

系统 SHALL 要求动态插件缓存 guest API 实现`cachecap.Service`领域契约。动态插件业务代码 MUST 使用`cachecap.CacheItem`和`time.Duration`语义表达缓存值与 TTL，不得依赖`pluginbridge/protocol`中的缓存 transport DTO、秒数字段或底层 host service envelope。

#### Scenario: 动态插件读取缓存领域对象

- **WHEN** 动态插件通过`guest.Services.Cache().Get(ctx, namespace, key)`读取缓存
- **THEN** guest API 返回`cachecap.CacheItem`、命中标记和错误
- **AND** 不返回`protocol.HostServiceCacheValue`

#### Scenario: 动态插件设置缓存 TTL

- **WHEN** 动态插件通过`guest.Services.Cache().Set(ctx, namespace, key, value, ttl)`设置缓存
- **THEN** guest API 接收`time.Duration`类型 TTL
- **AND** transport adapter 负责将 TTL 转换为 wire payload
- **AND** 插件业务代码不得直接操作 wire 层`expireSeconds`

### Requirement: 动态插件缓存分发必须复用插件作用域 cachecap 服务

系统 SHALL 要求动态插件缓存分发器在完成`hostServices`授权校验后调用当前插件作用域的`cachecap.Service`。分发器 MUST NOT 直接调用`kvcache.Service`、构造内部缓存 key、注入 owner type 或创建新的缓存服务图。

#### Scenario: 动态插件访问授权缓存资源

- **WHEN** 动态插件声明并获得授权访问`service: cache`的某个 resource ref
- **AND** 插件调用`get`、`set`、`delete`、`incr`或`expire`
- **THEN** WASM 分发器调用`capabilityServicesForHostCall(hcc).Cache()`
- **AND** 缓存 key 生成、租户隔离、TTL 和后端错误语义由`cachecap.Service`实现

#### Scenario: 动态插件访问未授权缓存资源

- **WHEN** 动态插件调用未授权的缓存 resource ref
- **THEN** WASM 分发器在进入`cachecap.Service`之前拒绝调用
- **AND** 底层缓存后端不得收到该请求

### Requirement: 源码插件和动态插件缓存必须共享缓存语义

系统 SHALL 要求源码插件和动态插件使用同一组插件缓存语义，包括插件 ID 隔离、租户上下文隔离、有损缓存定位、TTL、原子递增、集群 coordination KV backend 和写入失败返回错误。

#### Scenario: 源码插件和动态插件同名缓存隔离

- **WHEN** 源码插件和动态插件使用相同`namespace`和 logical `key`
- **THEN** 宿主按各自插件 ID 生成不同内部 cache key
- **AND** 两个插件不得互相读取缓存值

#### Scenario: 集群模式动态插件缓存使用共享后端

- **WHEN** `cluster.enabled=true`且动态插件写入缓存
- **THEN** 动态插件缓存通过启动期共享`kvcache.Service`背后的 coordination KV backend 写入
- **AND** 分发器不得自行创建 Redis client 或 SQL table backend
