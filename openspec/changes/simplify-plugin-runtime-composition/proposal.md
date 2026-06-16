## Why

当前主框架插件运行时组合仍保留了一些为打破启动循环而引入的宽泛 delegate、adapter 和进程级默认后端。它们在未绑定、缺失依赖或后端选择错误时容易静默成功，增加排障成本，也让缓存敏感服务的实例来源不够直观。

## What Changes

- 将插件运行时 delegate 的绑定状态改为显式可诊断，运行期不可安全跳过的操作在未绑定时返回明确错误。
- 收敛插件内部 cache notifier、dependency validator 和 source upgrade cache 适配器的空依赖语义，避免 nil service 路径伪装成功。
- 将 HTTP 启动期的 `kvcache` 后端选择改为显式创建共享 `kvcache.Service`，生产路径不再依赖进程级默认 provider 作为拓扑选择来源。
- 保留必要的启动组合接缝，不移除用于解决插件根服务构造循环的最小 delegate。
- 不新增 HTTP API、DTO、SQL、前端页面、插件协议或业务插件目录变更。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-service-layout`：补充插件服务内部 delegate 只能作为最小组合接缝，未绑定运行期调用不得静默成功。
- `service-dependency-injection-governance`：补充缓存敏感服务的启动装配必须显式选择共享后端，不得依赖进程级默认 provider。
- `plugin-cache-service`：补充宿主共享 `kvcache.Service` 必须由 HTTP 启动期按拓扑显式构造。
- `distributed-cache-coordination`：补充集群和单机缓存后端选择必须在拓扑感知构造边界完成。

## Impact

- 影响 `apps/lina-core/internal/service/plugin` 中运行时 delegate、内部 cache/upgrade 适配器和相关测试。
- 影响 `apps/lina-core/internal/cmd/internal/httpstartup` 中 `kvcache` 服务创建与插件服务启动装配。
- 影响 `apps/lina-core/internal/service/kvcache` 中默认 provider 的生产使用边界和测试覆盖。
- 涉及后端 Go、插件宿主运行时、缓存一致性、显式依赖注入和 OpenSpec 文档治理。
- 不影响数据权限、接口契约、数据库结构、前端 UI、用户可见文案或 `apps/lina-plugins/<plugin-id>/` 业务插件目录。
