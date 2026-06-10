## Context

`apps/lina-core`已经形成统一领域能力目录：`pkg/plugin/capability`定义领域契约，`internal/service/plugin/internal/capabilityhost`把启动期共享宿主服务适配为`capability.Services`和`pluginhost.Services`，动态插件通过`pluginbridge`协议进入宿主能力。

当前`Cache`、`Lock`和`Storage`仍处于不一致状态：

- `Cache`已有`cachecap.Service`，源码插件可通过`pluginhost.Services.Cache()`消费，但动态插件仍通过`guest.CacheHostService`和`protocol.HostServiceCacheValue`调用，宿主分发层直接注入`kvcache.Service`。
- `Lock`只有动态插件资源型 host service，宿主分发层直接注入`hostlock.Service`，源码插件没有统一领域消费入口。
- `Storage`只有动态插件资源型 host service，当前实现固化为动态插件本地目录树，依赖`config.GetPluginDynamicStoragePath`，没有主框架统一对象存储 provider 抽象，也没有源码插件消费入口。

本次用户已明确约束：

- `Storage`领域能力必须使用主框架统一抽象的存储能力，当前默认是单机本地磁盘存储，未来要支持多种分布式存储。
- `Storage`必须增加`Provider`设计，后续允许通过插件形式提供 OSS 等存储 provider。
- 源码插件不需要像动态插件那样在`plugin.yaml hostServices`声明资源边界，默认全信任。
- 本次只改`Cache`、`Lock`和`Storage`三个能力，不考虑兼容性，做完整重构。

## Goals / Non-Goals

**Goals:**

- 将`Cache`、`Lock`和`Storage`统一为`capability.Services`下的插件可见领域能力。
- 让源码插件通过`pluginhost.Services`消费`cachecap.Service`、`lockcap.Service`和`storagecap.Service`。
- 让动态插件通过`pluginbridge/guest`消费同一组领域接口，同时继续由`hostServices`安装授权快照限制方法和资源范围。
- 删除动态插件分发层对`kvcache.Service`、`hostlock.Service`和本地 storage config 的专用配置入口，统一复用启动期`capability.Services`共享目录。
- 为`Storage`建立`storagecap.Provider`和默认本地磁盘 provider，支持后续源码插件注册 OSS、MinIO、S3 等 provider。
- 明确`Storage` provider 只负责对象存储实现，插件隔离、租户隔离和动态资源授权由`storagecap.Service`或 transport adapter 负责。

**Non-Goals:**

- 不扩展到`runtime`、`network`、`data`、`manifest`、`hostConfig`等其他 host service。
- 不新增 HTTP API、前端页面、菜单、数据库表或用户可见文案。
- 不把源码插件改为需要在`plugin.yaml hostServices`声明`cache`、`lock`或`storage`资源。
- 不把插件`Storage`混同为宿主文件管理模块的`filecap`元数据能力；插件对象存储不直接暴露文件管理表、文件 ID 或上传下载工作台语义。
- 不在本次实现完整多 provider 治理 UI；provider 状态和选择先在后端能力与配置层闭环。

## Decisions

### 1. `Cache`、`Lock`和`Storage`成为普通插件领域能力

`capability.Services`增加或保留以下入口：

- `Cache() cachecap.Service`
- `Lock() lockcap.Service`
- `Storage() storagecap.Service`

源码插件从`pluginhost.Services`获得这些领域服务；动态插件从`guest.Services`获得实现相同领域接口的 guest adapter。动态`hostServices`协议继续保留`service: cache`、`service: lock`和`service: storage`作为授权和 transport service 名，但这些协议名不再拥有业务契约。

选择该方案而不是继续维护资源型 guest client，是因为已有领域能力治理要求新增插件可消费能力先进入`pkg/plugin/capability`。如果继续让动态插件直接依赖`protocol.*`DTO，会让源码插件与动态插件长期存在两套业务接口。

### 2. 动态授权保留在 transport 层，领域服务不接收授权快照

动态插件调用路径仍先经过`handleHostServiceInvoke`，校验：

- 插件是否具备`host:cache`、`host:lock`或`host:storage`能力。
- 请求 method 是否在安装或启用授权快照中。
- `storage`请求的 logical path 是否匹配授权 path。
- `cache`和`lock`请求的 resource ref 是否已授权。

校验通过后，WASM dispatcher 再调用`capabilityServicesForHostCall(hcc).Cache()`、`Lock()`或`Storage()`。源码插件默认全信任，不进入该授权校验路径，但领域服务仍按插件 ID 和租户上下文生成内部作用域。

选择该方案而不是把授权快照传入领域服务，是为了避免让`storagecap.Provider`或源码插件路径理解动态插件安装协议。动态授权是 transport 治理，领域服务负责插件作用域语义。

### 3. `Cache`复用现有`cachecap`，动态接口迁移到领域 DTO

`cachecap.Service`继续表达插件可见的`Get`、`Set`、`Delete`、`Incr`和`Expire`。动态 guest adapter 返回`cachecap.CacheItem`，TTL 使用`time.Duration`，不再向普通动态插件业务代码暴露`protocol.HostServiceCacheValue`或`expireSeconds`。

宿主动态分发层不再直接注入`kvcache.Service`。它在授权校验后调用插件作用域`cachecap.Service`，从而让源码插件和动态插件共享 key 生成、租户隔离、有损缓存、集群 coordination KV 和错误语义。

### 4. `Lock`新增`lockcap`，底层继续复用共享 locker

新增`pkg/plugin/capability/lockcap`，定义：

- `Service`：插件可见的`Acquire`、`Renew`和`Release`。
- `AcquireInput`、`AcquireOutput`：使用逻辑锁名、租约`time.Duration`、是否获得锁、票据和过期时间。
- 票据为宿主生成的不透明字符串，续租和释放必须校验票据与插件、租户和逻辑锁名匹配。

`capabilityhost`新增锁 adapter，复用启动期共享`locker.Service`或既有`hostlock.Service`后端，不在插件调用路径创建独立锁服务图。动态分发层在授权后调用同一`lockcap.Service`。

选择保留既有`hostlock`底层逻辑而不是重写锁实现，是因为当前锁服务已经完成插件 ID、租户、resource ref、ticket 和 lease 约束；本次需要的是领域契约和双插件形态统一。

### 5. `Storage`新增`storagecap.Service`和`storagecap.Provider`

新增`pkg/plugin/capability/storagecap`，分离消费面和 provider 面：

- `Service`面向源码插件和动态插件，提供`Put`、`Get`、`Delete`、`List`、`Stat`。
- `Provider`面向本地磁盘、OSS、MinIO、S3 等对象存储实现，提供对象读写、删除、列表和元数据读取。
- `ProviderFactory`允许源码插件通过`storagecap.Provide(pluginID, factory)`声明存储 provider。
- `ProviderRuntime`负责判断 provider 插件是否启用、读取 active provider 配置并构造`ProviderEnv`。

`Service`负责：

- 根据当前插件 ID、租户上下文和 logical path 生成 provider object key。
- 规范化路径，拒绝绝对路径、空路径、目录穿越和超过限制的路径。
- 执行对象大小、content type、overwrite、list limit 和元数据投影约束。
- 对动态插件调用，依赖 transport 层已完成的授权 path 校验；对源码插件调用，默认允许当前插件作用域下的所有 logical path。

`Provider`负责：

- 对 object key 执行底层存储读写和列表。
- 返回对象大小、content type、etag、更新时间、可见性等后端元数据。
- 不理解动态插件授权快照、源码插件信任策略、`hostServices`或工作台文件管理语义。

选择单独建立`storagecap.Provider`而不是直接复用当前`internal/service/file.Storage`，是因为现有`file.Storage`偏上传文件命名语义，只包含`Put/Get/Delete/Url`，缺少`List`、`Stat`、`Overwrite`和对象元数据能力。插件对象存储需要更接近对象存储底座；未来可让文件管理模块反向复用该底座，但不在本次扩大范围。

### 6. 默认本地磁盘 provider 由主框架提供

主框架提供一个默认本地磁盘 provider。默认 object key 建议使用稳定前缀：

```text
plugins/{pluginID}/tenant/{tenantID}/{logicalPath}
plugins/{pluginID}/platform/{logicalPath}
```

动态插件授权仍匹配`logicalPath`，不得匹配物理路径或 provider object key。宿主不得向 guest 返回本地绝对路径。

本地 provider 是主框架内建 fallback，不依赖插件启用状态。后续 OSS provider 由源码插件调用`storagecap.Provide`注册。

### 7. Provider 选择采用显式 active provider

第一版 provider 选择策略：

- 未配置 active provider 时，使用主框架内置`local` provider。
- 配置 active provider plugin ID 时，只有该插件启用且 provider 构造成功才使用该 provider。
- 如果配置了插件 provider 但插件未启用或构造失败，`storagecap.Service`返回明确能力不可用或 provider 错误，不静默回退本地磁盘。
- 多个插件注册 provider 时不自动冲突；只有配置选中的 provider 生效。`ProviderStatuses`可列出所有注册 provider 状态。

选择显式 active provider 而不是“启用 provider 中单活自动选择”，是因为 Storage 是持久数据落点。自动回退或自动切换可能导致数据分散在不同后端，风险高于普通可降级能力。

### 8. WASM 配置入口收敛

`ConfigureWasmHostServices`删除`kvCacheSvc`、`lockSvc`和`configSvc`参数，不再调用：

- `wasm.ConfigureCacheHostService`
- `wasm.ConfigureLockHostService`
- `wasm.ConfigureStorageHostService`

WASM 普通领域分发只通过`wasm.ConfigureDomainHostServices(capability.Services)`获得`Cache`、`Lock`和`Storage`。`config`、`manifest`、`hostConfig`等本次未收敛的资源型 service 保持现状。

替代方案是保留专用入口并内部转发到领域服务。该方案仍会让后续开发者误以为三类能力有独立底层注入通道，因此拒绝。

### 9. 动态 guest API 完整破坏式迁移

动态`guest.Services`中：

- `Cache()`返回`cachecap.Service`。
- `Lock()`返回`lockcap.Service`。
- `Storage()`返回`storagecap.Service`。

移除公共`CacheHostService`、`LockHostService`和`StorageHostService`接口，以及返回`protocol.*`DTO的业务调用面。协议编解码类型继续存在于`pluginbridge/protocol`，但只作为 transport 内部边界。

由于本项目没有兼容性负担，本次不保留旧别名。

## Risks / Trade-offs

- [Risk] `Storage` provider 选择错误可能导致数据写入错误后端。  
  Mitigation：active provider 显式配置；配置插件 provider 失败时返回错误，不自动回退本地；测试覆盖 provider 缺失、禁用和构造失败。

- [Risk] 本地磁盘 provider 在集群模式下不是天然共享存储。  
  Mitigation：设计和实现必须记录本地 provider 的集群语义；若`cluster.enabled=true`且未配置共享路径或分布式 provider，应提供明确诊断或按配置阻断，不能假装跨节点一致。

- [Risk] 动态 guest API 破坏式迁移会影响示例插件和测试。  
  Mitigation：同步迁移`linapro-demo-dynamic`、guest stub、WASM host service 测试和 descriptor 治理测试；不提供兼容层。

- [Risk] `storagecap.Provider`与现有`file.Storage`并存可能形成两个存储抽象。  
  Mitigation：`storagecap.Provider`定位为对象存储底座；`file.Storage`仍服务文件管理模块，本次不反向重构文件管理，后续可单独提案收敛。

- [Risk] 源码插件默认全信任可能绕过动态插件资源声明治理。  
  Mitigation：这是明确需求；领域服务仍强制插件 ID 和租户作用域隔离，并禁止源码插件接收 provider 物理路径或宿主内部存储客户端。

- [Risk] Cache 和 Lock 通过领域目录后可能重复执行授权或遗漏授权。  
  Mitigation：动态授权只在 transport 层执行一次；领域服务执行插件和租户隔离。测试分别覆盖未授权动态调用拒绝、源码插件全信任调用成功、跨插件和跨租户隔离。

## Migration Plan

1. 新增`lockcap`和`storagecap`领域契约，扩展`capability.Services`、`pluginhost.Services`和测试替身。
2. 在`capabilityhost`中新增锁和存储 adapter；`Cache`保留现有 adapter 并补齐动态复用路径。
3. 建立默认本地磁盘`storagecap.Provider`和 provider runtime/选择策略；默认未配置时使用内建本地 provider。
4. 调整 HTTP 启动装配，构造共享`storagecap.Service`和`lockcap.Service`，并传入`NewHostServices`。
5. 调整`ConfigureWasmHostServices`和`wasm`分发，删除`Cache`、`Lock`、`Storage`专用底层配置入口，改为通过`capabilityServicesForHostCall`获取领域服务。
6. 调整`pluginbridge/guest`，使`Cache()`、`Lock()`和`Storage()`返回领域接口，并将 WASI hostcall 实现迁入或对齐领域 adapter。
7. 更新`pluginbridge` descriptor、protocol alias 覆盖测试、guest stub 测试和 WASM host service 测试。
8. 迁移`linapro-demo-dynamic`示例插件到新领域接口。
9. 更新`apps/lina-core/pkg/plugin`相关 README，如边界说明需要同步。
10. 运行`openspec validate unify-plugin-resource-capabilities --strict`、相关 Go 包测试和必要静态检索。

回滚策略：由于不涉及数据库迁移，回滚以代码回滚为主；若已切换 Storage active provider 配置，回滚前必须确认对象数据落点，避免同一插件在两个后端间交叉写入。

## Impact Analysis

- `i18n`：不新增运行时用户可见文案、菜单、路由、API 文档源文本或语言包；仅新增 OpenSpec 文档。确认无运行时`i18n`资源影响。
- 缓存一致性：`Cache`继续复用启动期共享`kvcache.Service`和集群 coordination KV backend；不得在动态 hostcall 或源码插件路径临时创建新缓存服务。`Storage`本地 provider 需明确集群非共享语义或阻断策略。
- 数据权限：本次不新增宿主业务数据读取接口。动态插件资源授权仍由`hostServices`快照校验；源码插件默认全信任，但领域服务按插件 ID 和租户作用域隔离对象、缓存和锁资源。
- 接口性能：不新增列表 HTTP API。`Storage.List`必须有 limit 上限，provider 实现不得无界遍历；`Cache`和`Lock`均为单 key/resource 操作。
- 开发工具跨平台：不修改`Makefile`、脚本、`linactl`或 CI；确认无开发工具跨平台影响。
- 测试策略：属于后端插件桥接和能力边界行为变化，不新增 E2E。需要 Go 单元测试、guest stub 测试、WASM host service 测试、provider 选择测试和治理静态扫描。
- DI 来源：新增或修改的`lockcap.Service`、`storagecap.Service`、provider runtime 和默认 provider 必须由 HTTP 启动期构造并传入`capabilityhost`和 WASM domain services；不得在插件调用路径临时`New()`关键服务。

## Open Questions

无。用户已确认源码插件默认全信任、本次仅覆盖`Cache`、`Lock`、`Storage`，并要求完整破坏式重构。
