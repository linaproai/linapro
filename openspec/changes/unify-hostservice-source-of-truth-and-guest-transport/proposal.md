## Why

`pluginbridge`已经拥有结构化`hostservice`描述表，但双语`README`中的 host service 表格、guest client、宿主 dispatcher 和根目录 WASI 单例仍存在多处手写同步点。新增或调整一个 host service method 时仍可能遗漏文档、guest 或 dispatcher 某一侧，且 guest 侧同时存在根包手写`protowire`单例和`internal/domainhostcall`注入式客户端两套传输装配模式。

本变更把已有 descriptor 升级为文档和覆盖治理的强制事实源，并将 guest host service client 收敛到注入式单轨结构，降低动态插件能力扩展的长期维护扇出。

## What Changes

- 为`pluginbridge/internal/hostservice`新增 README 渲染器，用 descriptor 生成`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`中的 host service 表格，并通过漂移测试阻断未刷新文档的提交；不保留独立`go run`生成入口。
- 扩展 host service descriptor 治理测试，双向校验 descriptor、guest client selector、宿主 dispatcher service/method selector 和 dispatcher 文件集合，避免漏注册、多注册或孤儿 dispatcher。
- 将根目录仍残留的`pluginbridge_hostcall_*_wasip1.go`逐域单例客户端迁入`internal/domainhostcall`注入式客户端构造，`pluginbridge_directory.go`统一通过 invoker 装配。
- 删除根目录逐域 WASI 单例、adapter 和镜像 stub 残留；非 WASI stub 只保留传输层`InvokeHostService`统一 stub。`recordstore`保持现有注入式执行文件，因为它承载查询计划执行领域逻辑而非逐域客户端镜像。
- 保留现有 wire 格式、service/method 字符串、payload codec 和宿主 dispatcher 运行时行为；本变更只调整文档生成、治理测试和 guest 客户端内部结构。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `pluginbridge-subcomponent-architecture`：host service descriptor 必须驱动 README host service 表格、双向覆盖治理和 guest 注入式传输单轨结构。

## Impact

- 影响 Go 包：`apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice`、`internal/domainhostcall`、`pluginbridge`根目录 host call 文件、`pluginbridge_directory.go`及对应测试。
- 影响文档：`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`的 host service 表格改为生成区块，并保持中英文镜像事实一致。
- 影响开发工具：不新增独立生成器入口、脚本或默认开发命令；README 漂移治理由 Go 测试内的 descriptor 渲染器完成，不引入 shell-only 默认路径。
- 无 HTTP API、DTO、路由、OpenAPI 元数据、数据库 schema、SQL、前端页面或运行时用户可见文案变更。
- 数据权限影响：无数据访问路径或过滤语义变更；动态插件 data、tenant、org 等 host service 的授权、数据权限和拒绝策略保持不变。
- 缓存一致性影响：无缓存权威源、失效、刷新、跨实例同步或陈旧窗口变更；cache host service 的 wire 与宿主执行语义保持不变。
- `i18n`影响：无运行时语言包或用户可见 UI 文案变更；仅涉及双语技术文档同步生成，按文档镜像治理验证。
