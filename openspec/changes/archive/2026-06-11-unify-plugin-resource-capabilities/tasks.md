## 1. 领域契约

- [x] 1.1 扩展`capability.Services`、`pluginhost.Services`和相关测试替身，加入`Lock()`和`Storage()`领域入口，并保持`Cache()`契约不变。
- [x] 1.2 新增`pkg/plugin/capability/lockcap`，定义`Service`、输入输出 DTO、票据语义、租约`time.Duration`边界和结构化错误码。
- [x] 1.3 新增`pkg/plugin/capability/storagecap`，定义`Service`、`Provider`、`ProviderFactory`、`ProviderRuntime`、`ProviderEnv`、对象 DTO、列表上限和 provider 状态 DTO。
- [x] 1.4 为`storagecap`实现 provider 注册与 active provider 选择逻辑，支持默认本地 provider 和显式 active provider plugin ID。

## 2. 宿主能力适配

- [x] 2.1 在`capabilityhost`中新增`lockcap.Service`适配器，复用启动期共享锁后端，按插件 ID 和租户上下文生成内部锁名。
- [x] 2.2 在`capabilityhost`中新增`storagecap.Service`适配器，完成 logical path 规范化、插件和租户作用域 object key 生成、对象大小与列表 limit 校验。
- [x] 2.3 实现主框架默认本地磁盘`storagecap.Provider`，并明确单机和集群模式语义。
- [x] 2.4 调整`NewHostServices`构造参数和`capabilityhost`目录结构，使源码插件可通过`pluginhost.Services.Lock()`和`Storage()`消费能力。
- [x] 2.5 调整 HTTP 启动装配，构造共享`lockcap`和`storagecap`能力并注入`capability.Services`，不得在插件调用路径创建独立服务图。

## 3. 动态插件桥接

- [x] 3.1 调整`ConfigureWasmHostServices`，删除`kvCacheSvc`、`lockSvc`和`configSvc`等`Cache`、`Lock`、`Storage`专用参数。
- [x] 3.2 删除或废弃`wasm.ConfigureCacheHostService`、`ConfigureLockHostService`和`ConfigureStorageHostService`生产入口，并让对应分发器通过`capabilityServicesForHostCall`获取领域服务。
- [x] 3.3 调整动态`cache`分发器，在授权校验后调用`cachecap.Service`并完成领域 DTO 与 wire payload 转换。
- [x] 3.4 调整动态`lock`分发器，在授权校验后调用`lockcap.Service`并完成票据、租约和过期时间 wire 转换。
- [x] 3.5 调整动态`storage`分发器，在授权 path 校验后调用`storagecap.Service`，不得直接访问本地目录或 provider object key。
- [x] 3.6 调整`pluginbridge/guest`，使`Cache()`、`Lock()`和`Storage()`返回`cachecap.Service`、`lockcap.Service`和`storagecap.Service`，并移除公开的`CacheHostService`、`LockHostService`和`StorageHostService`业务接口。

## 4. 示例插件和文档

- [x] 4.1 迁移`linapro-demo-dynamic`示例插件，使用新的领域接口调用`Cache`、`Lock`和`Storage`。
- [x] 4.2 更新动态插件 host service 示例清单和说明，保留`hostServices`授权声明但移除旧 guest protocol DTO 用法。
- [x] 4.3 检查并更新`apps/lina-core/pkg/plugin`相关 README 中关于源码插件、动态插件和资源能力边界的说明。

## 5. 测试和治理验证

- [x] 5.1 新增或更新`cachecap`动态桥接测试，覆盖动态授权拒绝、领域 DTO 返回、源码插件和动态插件缓存隔离、集群共享后端语义。
- [x] 5.2 新增`lockcap`单元测试和 WASM 分发测试，覆盖跨插件隔离、跨租户隔离、票据不匹配、未授权动态调用拒绝和后端故障。
- [x] 5.3 新增`storagecap`单元测试和 WASM 分发测试，覆盖路径规范化、目录穿越拒绝、授权 path 拒绝、源码插件全信任、provider 选择、本地 provider 和列表 limit。
- [x] 5.4 更新 guest stub、descriptor 覆盖测试和治理扫描，阻断重新引入`Cache`、`Lock`、`Storage`专用 WASM 底层配置入口或公开 protocol DTO 型 guest 业务接口。
- [x] 5.5 运行`openspec validate unify-plugin-resource-capabilities --strict`。
- [x] 5.6 运行覆盖变更包的 Go 测试，至少包括`apps/lina-core/pkg/plugin/capability/...`、`apps/lina-core/pkg/plugin/pluginbridge/...`、`apps/lina-core/internal/service/plugin/...`和受影响示例插件包。

## 执行记录

- 规则影响：命中 OpenSpec、插件、架构、缓存一致性、后端 Go、测试、文档规范；本次未新增或修改 HTTP API、数据库结构、前端 UI、运行时用户可见 i18n 文案或数据权限策略。
- 性能记录：`storagecap.Service.List`、guest `Storage().List`返回值和本地 provider 列表均有明确 limit；本地 provider 从请求 prefix 起始遍历，避免扫描插件存储根；卸载清理按 bounded list 循环删除授权 prefix 对象。
- 审查修复：`lina-review`过程中发现 guest storage list 响应未回填默认/最大 limit 语义，已补充共享 helper 和普通构建单元测试。
- 验证记录：已运行`openspec validate unify-plugin-resource-capabilities --strict`、`go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... -count=1`、`go test ./pkg/plugin/pluginbridge/guest -count=1`、`go test ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd/internal/httpstartup -count=1`、`go test ./internal/service/plugin/internal/runtime -count=1`、`GOWORK=off go test ./backend/internal/service/dynamic ./backend/internal/controller/dynamic -count=1`。
