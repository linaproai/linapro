## Requirements

### Requirement: Dynamic Plugins Acquire Named Lock Resources

The system SHALL provide lock service reusing host distributed lock with ticket-based renew/release.

#### Scenario: Plugin acquires authorized lock
- **WHEN** a plugin calls lock service acquire
- **THEN** the host applies lease and timeout policies
- **AND** binds the logical lock name to plugin isolation
- **AND** returns a lock ticket

#### Scenario: Plugin renews or releases lock
- **WHEN** a plugin renews or releases a held lock
- **THEN** the host validates ticket and lock-resource match
- **AND** only operates on the current plugin's valid locks

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

### Requirement: 插件锁必须复用宿主启动期共享锁后端

系统 SHALL 将宿主启动期创建或注入的共享锁后端用于源码插件和动态插件锁领域服务。插件锁实现 MUST NOT 在插件注册、请求处理、hook 回调、jobs 回调、WASM host call 或锁方法调用路径中临时创建独立 locker、coordination provider、Redis client 或进程内锁图。
