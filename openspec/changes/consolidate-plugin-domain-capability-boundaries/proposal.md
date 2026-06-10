## Why

当前领域能力的公共契约已经集中在`pkg/plugin/capability`，但宿主实现、动态`WASM host service`配置和动态 guest 侧能力代理仍存在命名不清、遗留专用入口和局部重复接口，导致源码插件与动态插件共享同一领域语义的事实不够直观。现在需要在继续扩展插件生态前收敛能力边界，避免后续新增领域能力时沿用分散入口。

## What Changes

- 将`apps/lina-core/internal/service/plugin/internal/hostservices`重命名或迁移为职责更明确的领域能力实现组件`internal/service/plugin/internal/capabilityhost`，并保持由`internal/service/plugin`根包提供窄 facade。
- 保持`pkg/plugin/capability`作为唯一领域能力契约入口，保持`pkg/plugin/pluginhost`作为源码插件固定消费入口。
- 将动态插件普通领域能力在宿主侧收敛为单一`ConfigureDomainHostServices(capability.Services)`配置入口，移除`AI`、`User`、`Org`、`Tenant`等领域专用`Configure*HostService`全局入口和 fallback 目录。
- 保持`pkg/plugin/pluginbridge/protocol`作为动态`hostServices`公开协议、payload DTO 和 codec owner，保持`pkg/plugin/pluginbridge/internal/hostservice`作为 descriptor、授权和清单治理目录，并明确二者都不拥有领域业务契约。
- 将动态 guest 侧领域能力代理固定在`pkg/plugin/pluginbridge`公共入口和`pluginbridge/internal/domainhostcall`内部实现；`AI`guest client 应复用或实现`capability/aicap.Service`，不再维护一套与`aicap`平行的 guest 专用领域接口。
- 补充治理测试和静态扫描，阻止领域能力实现、WASM 配置入口、guest 领域代理和协议描述源再次分叉。
- 不改变动态插件`hostServices`的 service/method 字符串、payload wire 格式、安装授权语义或源码插件 registrar 语义。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-domain-capabilities`：明确领域能力实现、源码插件消费入口和动态插件消费入口的固定归属。
- `plugin-host-service-extension`：要求动态普通领域`host service`统一使用单一领域能力目录配置入口。
- `pluginbridge-subcomponent-architecture`：明确`pluginbridge`协议目录和 guest 领域代理的职责边界，收敛`AI`guest 代理。
- `plugin-capability-boundary-governance`：增加防止领域能力边界再次分叉的治理验证。
- `service-dependency-injection-governance`：要求迁移后仍复用启动期共享`capability.Services`实例，不新增领域专用全局服务目录。

## Impact

- 影响`apps/lina-core/internal/service/plugin/internal/hostservices`及其测试、导入路径和根 facade 注释。
- 影响`apps/lina-core/internal/service/plugin/internal/wasm`中的普通领域 host service 配置、分发测试和测试工具配置。
- 影响`apps/lina-core/pkg/plugin/pluginbridge`与`pluginbridge/internal/domainhostcall`中的动态领域能力代理组织，尤其是`AI`能力。
- 影响`apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice`与`protocol`的治理测试覆盖路径，但不改变动态`hostServices`协议职责边界。
- 需要更新`apps/lina-core/pkg/plugin`相关 README 中能力边界说明。
- 不涉及数据库迁移、HTTP API 路由、前端 UI、运行时文案、菜单、SQL、插件 manifest 资源格式或动态插件 wire 兼容性。
