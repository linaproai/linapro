## Why

当前插件领域能力同时存在`Service`和`AdminService`两套插件可见入口，源码插件需要理解`Services()`和`Services.Admin()`两个目录，动态插件又实际按`service + method + resource`授权，接口分层没有形成真实安全边界。项目没有历史兼容负担，应一次性把插件领域能力收敛到统一`Service`模型，降低插件开发复杂度，并让风险、授权、数据权限和缓存治理回到方法级契约。

## What Changes

- **BREAKING**：废除`capability.AdminServices`、各领域`AdminService`和`pluginhost.Services.Admin()`，原管理方法并入对应领域统一`Service`。
- **BREAKING**：动态 wire method 一次性标准化，不保留旧方法兼容别名。
- **BREAKING**：删除`pluginhost.Services.PluginLifecycle()`和`PluginState()`顶层入口，源码插件改用`Services.Plugins().Lifecycle()`和`Services.Plugins().State()`；保留的租户快捷入口仍必须委托到`Tenant()`领域。
- 动态插件能否声明和调用方法，仅由动态`host service registry`是否注册该方法表达；不新增`dynamic-auth`、`source-only`、`reserved`或等价方法可声明性字段。
- 动态插件调用普通能力或治理能力必须同时满足方法已注册、`plugin.yaml hostServices`已声明、安装或启用时已授权，以及运行时`service + method + resource`校验通过。
- 源码插件直接使用统一类型化`Service`，不经过`plugin.yaml hostServices`声明；领域方法仅依赖标准业务`ctx`中的当前用户、租户、权限和数据权限上下文，不再要求额外插件调用上下文。
- 领域能力实现归属真实领域 owner，owner 内部复用既有业务逻辑；`capabilityhost`和 WASM host service 仅保留标准业务上下文桥接、动态授权、编解码和错误映射。
- 补充插件可见方法的风险等级、授权资源、上下文要求、数据权限、性能边界和缓存影响元数据要求。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-domain-capabilities`：替换`Service`/`AdminService`分层要求，统一为每个领域一个插件可见`Service`入口，并调整实现 owner 与动态注册边界。
- `plugin-permission-governance`：删除源码插件通过`Services.Admin()`获取管理能力的规范，明确源码插件使用统一`Service`且仍受领域治理约束。
- `plugin-host-service-extension`：明确动态插件可声明性由`host service registry`注册事实表达，禁止新增额外方法可声明性字段，普通能力和治理能力均通过领域 owner 注册事实进入动态授权流程，未注册方法在声明、安装、启用和运行时均拒绝。
- `plugin-package-boundary-governance`：约束`capabilityhost`和 WASM host service 只作为薄适配层，业务实现不得集中复制在插件组件内。

## Impact

- 代码影响：`apps/lina-core/pkg/plugin/capability`、`apps/lina-core/pkg/plugin/pluginhost`、`apps/lina-core/pkg/plugin/pluginbridge`、`apps/lina-core/internal/service/plugin/internal/capabilityhost`、动态 WASM host service registry、guest client、dispatcher 和源码插件调用点。
- 文档影响：`apps/lina-core/pkg/plugin/README.md`、`apps/lina-core/pkg/plugin/README.zh-CN.md`和相关 OpenSpec 基线规范需要同步删除`AdminService`描述。
- 权限影响：插件安装授权、运行时授权和领域数据权限保持分离；动态插件方法授权不再映射工作台菜单按钮权限，但目标数据边界仍由领域方法执行。
- 缓存影响：权限、插件状态、运行时配置、字典、组织、租户等关键运行时数据写入方法必须记录权威数据源、事务后失效和跨实例同步策略。
- `i18n`影响：本迭代本身不新增运行时 UI 文案；实现阶段若修改 API 文档源文本、错误消息、菜单或插件清单文案，必须按宿主与插件边界补充`i18n`治理。
