## ADDED Requirements

### Requirement: 插件资源型基础能力必须收敛为领域能力

系统 SHALL 将插件可消费的`cache`、`lock`和`storage`能力发布为`pkg/plugin/capability`下的领域契约。源码插件 MUST 通过`pluginhost.Services`消费这些领域能力；动态插件 MUST 通过`pluginbridge/guest`消费实现同一领域接口的 guest adapter。`pluginbridge`协议和`hostServices`声明只拥有动态插件 transport、授权和 payload 编解码职责，不得成为`cache`、`lock`或`storage`业务接口 owner。

#### Scenario: 源码插件消费资源型基础能力

- **WHEN** 源码插件在 route、hook、jobs 或生命周期回调中需要缓存、锁或对象存储能力
- **THEN** 宿主通过`pluginhost.Services`提供`cachecap.Service`、`lockcap.Service`和`storagecap.Service`
- **AND** 插件业务服务应注入所需的最窄领域接口
- **AND** 插件不得接收宿主内部`kvcache.Service`、`hostlock.Service`、存储 provider、物理路径或底层客户端

#### Scenario: 动态插件消费资源型基础能力

- **WHEN** 动态插件业务代码调用`guest.Services.Cache()`、`guest.Services.Lock()`或`guest.Services.Storage()`
- **THEN** guest 侧返回值必须实现对应`cachecap.Service`、`lockcap.Service`或`storagecap.Service`
- **AND** 公共 guest API 不得向业务代码暴露`protocol.HostServiceCacheValue`、`protocol.HostServiceLockAcquireResponse`、`protocol.HostServiceStorageObject`或等价 transport DTO 作为领域返回值

#### Scenario: 动态插件资源能力经过授权后进入领域服务

- **WHEN** 动态插件通过`hostServices`调用`cache`、`lock`或`storage`方法
- **THEN** WASM host service 分发器必须先校验能力、方法和资源授权快照
- **AND** 校验通过后必须通过`capabilityServicesForHostCall`获取插件作用域领域服务
- **AND** 分发器不得直接调用底层缓存、锁或存储实现

### Requirement: 源码插件资源能力默认全信任但必须作用域隔离

系统 SHALL 将源码插件视为可信插件形态，源码插件消费`cache`、`lock`和`storage`时不需要在`plugin.yaml hostServices`中声明资源边界。即便源码插件默认全信任，领域服务 MUST 仍按当前插件 ID 和租户上下文隔离内部 cache key、lock name 和 storage object key。

#### Scenario: 源码插件不声明资源边界

- **WHEN** 源码插件未在`plugin.yaml hostServices`中声明`cache`、`lock`或`storage`资源
- **THEN** 宿主仍允许该源码插件通过`pluginhost.Services`调用对应领域服务
- **AND** 宿主不得要求源码插件经过动态插件安装授权确认

#### Scenario: 源码插件资源作用域隔离

- **WHEN** 两个源码插件使用相同 logical cache namespace、lock name 或 storage path
- **THEN** 宿主为两个插件生成不同的内部资源身份
- **AND** 一个源码插件不得读取、续租、释放或删除另一个源码插件的资源

### Requirement: WASM 资源能力配置必须复用领域能力目录

系统 SHALL 要求动态插件`cache`、`lock`和`storage`分发复用启动期注入的同一个`capability.Services`目录。WASM 运行时 MUST NOT 继续发布或使用`ConfigureCacheHostService`、`ConfigureLockHostService`、`ConfigureStorageHostService`或等价的资源能力专用底层配置入口。

#### Scenario: 启动期配置动态资源能力

- **WHEN** 宿主启动期配置动态插件 WASM host services
- **THEN** 启动装配只需要为`cache`、`lock`和`storage`调用`ConfigureDomainHostServices(capability.Services)`
- **AND** 不再额外注入`kvcache.Service`、`hostlock.Service`或 storage config reader 到 WASM 分发层

#### Scenario: 治理扫描检查专用配置入口

- **WHEN** 生产代码新增`ConfigureCacheHostService`、`ConfigureLockHostService`、`ConfigureStorageHostService`或等价专用入口
- **THEN** 治理验证必须失败
- **AND** 变更必须改为通过领域能力目录配置资源能力
