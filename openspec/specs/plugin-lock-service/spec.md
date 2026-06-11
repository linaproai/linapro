# 插件锁服务规范

## Purpose
待定 - 由归档变更 dynamic-plugin-host-service-extension 创建。归档后更新目的。
## Requirements
### Requirement: 动态插件通过命名锁资源获取宿主锁能力

系统 SHALL 为动态插件提供受治理的锁服务，插件只能对宿主授权的命名锁资源执行获取、续租和释放操作。

#### Scenario: 插件获取授权锁资源

- **WHEN** 插件调用锁服务获取一个已授权的`host-lock`资源
- **THEN** 宿主按该锁资源的租约和超时策略执行加锁
- **AND** 宿主将逻辑锁名自动绑定到插件隔离的实际锁名
- **AND** 宿主返回锁票据或失败结果

#### Scenario: 插件释放或续租锁资源

- **WHEN** 插件调用锁服务释放或续租一个已持有的锁
- **THEN** 宿主校验锁票据和锁资源匹配关系
- **AND** 仅对当前插件持有的有效锁执行续租或释放

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

### Requirement: 插件锁必须通过 lockcap 领域契约提供

系统 SHALL 定义`lockcap.Service`作为源码插件和动态插件共享的插件锁领域契约。该契约 MUST 提供获取、续租和释放锁的能力，并使用领域 DTO 表达逻辑锁名、租约、锁票据、是否获取成功和过期时间。公共插件调用面 MUST NOT 暴露`pluginbridge/protocol`锁响应 DTO 或宿主内部`hostlock.Service`。

#### Scenario: 源码插件获取插件锁服务

- **WHEN** 源码插件通过`pluginhost.Services.Lock()`获取锁服务
- **THEN** 宿主返回绑定当前插件 ID 的`lockcap.Service`
- **AND** 该服务不要求源码插件声明`hostServices`资源

#### Scenario: 动态插件获取插件锁服务

- **WHEN** 动态插件通过`guest.Services.Lock()`获取锁服务
- **THEN** guest API 返回实现`lockcap.Service`的客户端
- **AND** 插件业务代码不得依赖`protocol.HostServiceLockAcquireResponse`或等价 transport DTO

### Requirement: 插件锁必须按插件和租户作用域隔离

系统 SHALL 在`lockcap.Service`内部将插件可见 logical lock name 映射为包含插件 ID 和租户维度的内部锁名。不同插件或不同租户使用相同 logical lock name 时 MUST 不互相阻塞。无租户上下文时 MUST 使用平台级插件锁作用域。

#### Scenario: 不同插件同名锁隔离

- **WHEN** 插件`plugin-a`获取 logical lock `sync`
- **AND** 插件`plugin-b`获取 logical lock `sync`
- **THEN** 两个锁互不冲突
- **AND** 两个插件获得的票据不得互相续租或释放

#### Scenario: 不同租户同名锁隔离

- **WHEN** 同一插件在租户`1001`获取 logical lock `sync`
- **AND** 同一插件在租户`1002`获取 logical lock `sync`
- **THEN** 两个锁互不冲突
- **AND** 任一租户的票据不得释放另一租户的锁

### Requirement: 动态插件锁分发必须复用插件作用域 lockcap 服务

系统 SHALL 要求动态插件锁分发器在完成`hostServices`授权校验后调用当前插件作用域的`lockcap.Service`。分发器 MUST NOT 直接调用`hostlock.Service`、`locker.Service`、coordination lock store 或创建新的锁服务图。

#### Scenario: 动态插件获取授权锁资源

- **WHEN** 动态插件声明并获得授权访问`service: lock`的 logical lock resource ref
- **AND** 插件调用`acquire`
- **THEN** WASM 分发器调用`capabilityServicesForHostCall(hcc).Lock()`
- **AND** 内部锁名、租户隔离、租约边界和票据由`lockcap.Service`或其 adapter 处理

#### Scenario: 动态插件访问未授权锁资源

- **WHEN** 动态插件调用未授权的 logical lock resource ref
- **THEN** WASM 分发器在进入`lockcap.Service`之前拒绝调用
- **AND** 底层锁后端不得收到该请求

### Requirement: 插件锁必须复用宿主启动期共享锁后端

系统 SHALL 将宿主启动期创建或注入的共享锁后端用于源码插件和动态插件锁领域服务。插件锁实现 MUST NOT 在插件注册、请求处理、hook 回调、jobs 回调、WASM host call 或锁方法调用路径中临时创建独立 locker、coordination provider、Redis client 或进程内锁图。

#### Scenario: 集群模式插件锁使用 coordination lock

- **WHEN** `cluster.enabled=true`且插件调用`lockcap.Service.Acquire`
- **THEN** 系统通过宿主统一 coordination lock 后端获取锁
- **AND** 插件不得获得 Redis client 或底层锁存储连接

#### Scenario: 锁后端不可用

- **WHEN** 共享锁后端不可用且插件调用`acquire`、`renew`或`release`
- **THEN** `lockcap.Service`返回结构化错误
- **AND** 系统不得向插件报告锁获取、续租或释放成功

