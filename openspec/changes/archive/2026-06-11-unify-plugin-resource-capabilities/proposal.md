## Why

当前动态插件的`cache`、`lock`和`storage`能力仍以资源型`hostServices`直连接口暴露，源码插件只能消费其中一部分能力，导致同一宿主基础能力在源码插件和动态插件之间存在两套调用面。项目正在收敛插件领域能力边界，需要将这三类高频资源能力统一为领域契约，同时保留动态插件安装授权和资源治理。

## What Changes

- 将插件`Cache`、`Lock`和`Storage`统一纳入`pkg/plugin/capability`领域能力目录，源码插件通过`pluginhost.Services`消费，动态插件通过`pluginbridge/guest`消费同一组领域接口。
- 修改动态插件`cache`、`lock`和`storage`分发路径，使其在完成`hostServices`能力、方法和资源授权后进入插件作用域的领域服务，而不是直接调用底层`kvcache`、`hostlock`或本地文件目录实现。
- 为`Storage`新增`Provider`设计，主框架默认提供本地磁盘 provider，后续允许源码插件注册 OSS、MinIO、S3 等存储 provider；存储 provider 负责对象存储实现，`storagecap.Service`负责插件作用域、租户作用域和动态授权边界。
- 保留动态插件的`storage`路径资源授权，以及`cache`、`lock`资源引用授权；源码插件默认全信任，不要求在`plugin.yaml hostServices`中声明资源边界，但仍由领域服务按`pluginID`和租户上下文隔离。
- **BREAKING**：动态插件公开 guest API 不再返回`protocol.*`类型的`CacheHostService`、`LockHostService`和`StorageHostService`；调用方必须迁移到`cachecap.Service`、`lockcap.Service`和`storagecap.Service`领域接口。
- **BREAKING**：删除动态插件 WASM 分发层的`ConfigureCacheHostService`、`ConfigureLockHostService`和`ConfigureStorageHostService`专用入口，统一通过`ConfigureDomainHostServices(capability.Services)`注入领域能力目录。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-host-domain-capabilities`：将`cache`、`lock`和`storage`从资源型直连 host service 收敛为插件可见领域能力，并明确源码插件与动态插件的统一消费边界。
- `plugin-cache-service`：动态插件缓存调用改为通过`cachecap.Service`领域契约进入共享插件缓存 facade，源码插件与动态插件共享 TTL、有损缓存和集群后端语义。
- `plugin-lock-service`：新增源码插件可消费的`lockcap.Service`领域契约，并要求动态插件锁分发复用同一领域服务。
- `plugin-storage-service`：新增`storagecap.Service`和`storagecap.Provider`要求，存储能力改为基于主框架统一对象存储抽象，默认本地磁盘 provider，后续支持插件形式提供 OSS 等存储 provider。

## Impact

- 影响`apps/lina-core/pkg/plugin/capability`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge/guest`、`pkg/plugin/pluginbridge/internal/hostservice`、`internal/service/plugin/internal/capabilityhost`和`internal/service/plugin/internal/wasm`。
- 影响动态插件示例`linapro-demo-dynamic`对`Storage`、`Cache`和`Lock` guest API 的调用方式，以及相关 host service 单元测试、stub 测试和治理测试。
- 影响启动装配和测试 fixture：WASM 资源能力不再逐项注入底层服务，而是复用启动期构造的`capability.Services`共享目录。
- 不新增 HTTP API、前端 UI、数据库表或运行时用户可见文案；`i18n`无运行时资源影响。
