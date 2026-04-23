## Why

当前源码插件和动态插件的安装、启用流程主要依赖系统启动后的管理操作，这会让一些必须在宿主启动即就绪的插件无法自动进入可用状态，例如演示控制、启动期治理扩展或需要在首个请求前注册能力的插件。我们需要一个配置驱动的启动期插件 bootstrap 机制，让宿主在启动过程中按明确策略自动完成安装、启用与就绪收敛，同时避免把所有源码插件都一刀切默认启用。

## What Changes

- 新增宿主级 `plugin.startup` 静态配置，允许按 `pluginId` 声明插件启动目标状态（`manual` / `installed` / `enabled`）、是否必需，以及动态插件所需的宿主服务授权快照。
- 在宿主 HTTP 启动链路中增加“插件启动期 bootstrap”阶段：先完成插件发现与注册表同步，再按策略分别推进源码插件和动态插件的安装/启用。
- 调整源码插件生命周期语义：源码插件被发现后默认仍保持“仅发现”状态，只有管理员显式操作或启动期策略命中时，才会自动安装或启用。
- 复用动态插件既有的 `desired_state/current_state` 与 reconciler 机制，在启动期支持把动态插件收敛到 `installed` 或 `enabled`；当动态插件声明受治理的宿主服务资源时，支持通过启动配置一并提供授权快照。
- 明确启动失败处理：非必需插件的缺失或收敛失败仅记录告警并允许宿主降级启动；必需插件失败时允许宿主 fail-fast，避免系统以“看似启动成功但关键插件缺席”的状态对外服务。

## Capabilities

### New Capabilities
- `plugin-startup-bootstrap`: 定义宿主在启动阶段按静态配置自动安装、启用并等待插件收敛的能力，覆盖源码插件、动态插件、缺失处理、失败策略和动态插件授权快照。

### Modified Capabilities
- `plugin-manifest-lifecycle`: 调整源码插件发现后的默认生命周期语义，并补充启动期策略可触发插件安装/启用的规则。

## Impact

- 受影响代码主要位于 `apps/lina-core/internal/cmd/cmd_http.go`、`apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/plugin/` 及其 `internal/runtime` / `internal/catalog` 子模块。
- 需要扩展 `apps/lina-core/manifest/config/config.template.yaml` 中的插件配置结构，并补充对应配置读取与校验。
- 需要新增插件启动期 bootstrap 的单元测试 / 集成测试，覆盖 source、dynamic、授权、cluster 降级和 fail-fast 场景。
- 不涉及现有对外 RESTful API 变更，重点是宿主启动行为、配置模型和插件生命周期收敛逻辑的增强。
