## Why

源码插件目前可以通过 `HostServices()` 使用配置、鉴权、上下文、通知、租户过滤等宿主能力，但无法以稳定的公共契约复用宿主统一 KV cache。插件如果自行构造缓存或直接依赖宿主内部 `kvcache`，会破坏插件隔离、租户边界、集群模式缓存后端选择和后续治理空间。

本变更将宿主缓存能力以源码插件专用 facade 暴露出来，让源码插件获得与动态插件一致的受治理缓存能力，同时避免泄露宿主内部 key 编码、owner 类型、coordination backend 或底层缓存客户端。

## What Changes

- 在源码插件 `HostServices` 服务目录中增加插件可见的缓存服务入口，供 HTTP registrar、Cron registrar 和 hook payload 使用。
- 新增源码插件缓存公共契约，提供 `Get`、`Set`、`Delete`、`Incr`、`Expire` 等基础操作，并使用 `time.Duration` 表达 TTL。
- 宿主实现一个按 `pluginID` 绑定的缓存适配器，将插件侧的 `namespace + key` 映射到内部 `kvcache.Service`，并强制使用 `OwnerTypePlugin`。
- 缓存 key 必须由宿主根据当前插件 ID、命名空间、逻辑 key 和当前租户上下文生成，插件不得获得或传入内部 owner key。
- 源码插件缓存复用启动期注入的同一个 `kvCacheSvc`，在 `cluster.enabled=false` 时沿用单机 SQL/local 路径，在 `cluster.enabled=true` 时沿用 coordination KV 后端。
- 保持缓存为有损缓存，不允许源码插件将其作为权限、配置、插件状态、租户隔离、业务权威数据或关键缓存修订号的事实源。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-cache-service`: 扩展现有插件缓存服务规范，使源码插件也可以通过受治理的宿主缓存 facade 使用插件私有 KV cache，并继承租户隔离、命名空间隔离、有损缓存和集群后端选择语义。

## Impact

- 影响 `apps/lina-core/pkg/pluginhost` 的 `HostServices` 公共接口，以及 HTTP/Cron/hook 注册路径中的服务目录传递方式。
- 影响 `apps/lina-core/pkg/pluginservice/contract` 和 `apps/lina-core/internal/service/pluginhostservices`，需要新增源码插件可见的 cache contract 与宿主适配器。
- 影响 `apps/lina-core/internal/cmd/cmd_http_runtime.go`，需要将启动期 `kvCacheSvc` 显式注入源码插件 host service 目录。
- 影响官方源码插件的编译面：所有测试替身和实现 `pluginhost.HostServices` 的结构需要补齐 `Cache()` 方法。
- 需要新增或更新 Go 单元测试，覆盖插件 ID 绑定、租户 key 隔离、TTL、`incr`、`expire`、nil/缺失依赖错误和集群后端复用。
- i18n 影响预计为无：本变更不新增用户可见 UI 文案、菜单、按钮、表单、manifest i18n 或 apidoc 文案；若实现中新增调用端可见业务错误码，必须同步维护错误 i18n。
