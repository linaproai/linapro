## Why

当前源码插件和动态插件的安装、启用流程主要依赖系统启动后的管理操作，这会让一些必须在宿主启动即就绪的插件无法自动进入可用状态，例如演示控制、启动期治理扩展或需要在首个请求前注册能力的插件。我们需要一个足够简单的启动期自动启用机制，并且该配置必须直接落在宿主服务主配置文件中，避免为少量“开机即启用”的插件引入过重的策略模型。

## What Changes

- 在宿主服务主配置文件中新增简化配置 `plugin.autoEnable`，仅使用插件 ID 列表声明哪些插件需要在系统启动时自动启用。
- 将“自动启用”语义定义为“若插件尚未安装，则先安装；随后启用”，避免继续暴露 `manual` / `installed` / `enabled` 多档启动策略配置。
- 在宿主 HTTP 启动链路中增加“插件启动期 bootstrap”阶段：先完成插件发现与注册表同步，再仅对 `plugin.autoEnable` 命中的插件推进安装与启用。
- 调整源码插件生命周期语义：源码插件被发现后默认仍保持“仅发现”状态，只有管理员显式操作或命中 `plugin.autoEnable` 时，才会自动安装并启用。
- 复用动态插件既有的 `desired_state/current_state` 与 reconciler 机制；对于声明受治理宿主服务资源的动态插件，启动自动启用只复用既有授权快照，不再把复杂授权结构塞进主配置文件。
- 明确启动失败处理：既然 `plugin.autoEnable` 是显式声明的开机必需插件列表，则命中列表的插件缺失或启用失败时，宿主直接 fail-fast，避免系统以“关键插件未启用”的状态对外服务。

## Capabilities

### New Capabilities
- `plugin-startup-bootstrap`: 定义宿主在启动阶段按静态配置自动安装、启用并等待插件收敛的能力，覆盖源码插件、动态插件、缺失处理、失败策略和动态插件授权快照。

### Modified Capabilities
- `plugin-manifest-lifecycle`: 调整源码插件发现后的默认生命周期语义，并补充启动期策略可触发插件安装/启用的规则。

## Impact

- 受影响代码主要位于 `apps/lina-core/internal/cmd/cmd_http.go`、`apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/plugin/` 及其 `internal/runtime` / `internal/catalog` 子模块。
- 需要扩展宿主主配置文件模板 `apps/lina-core/manifest/config/config.template.yaml` 中的插件配置结构，并补充对应配置读取与校验。
- 需要新增插件启动期 bootstrap 的单元测试 / 集成测试，覆盖 source、dynamic、既有授权快照、cluster 和 fail-fast 场景。
- 不涉及现有对外 RESTful API 变更，重点是宿主启动行为、配置模型和插件生命周期收敛逻辑的增强。
