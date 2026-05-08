## Why

`apps/lina-core/pkg/pluginbridge` 当前将桥接 ABI、编解码、WASM 产物解析、host call、host service 协议和 guest SDK 聚合在同一个 Go 包根目录下，生产源码文件数量较多，用户阅读时难以分辨哪些是稳定合约、哪些是宿主内部协议、哪些是插件开发者应直接使用的 helper。

本变更将 `pluginbridge` 按类似 `pkg/pluginservice` 的子组件方式重构，使公开能力边界按职责呈现，同时保留现有稳定调用路径的兼容 facade，降低动态插件开发者和宿主维护者的理解成本。

## What Changes

- 将 `pkg/pluginbridge` 重构为职责明确的公开子组件包，例如 `contract`、`codec`、`artifact`、`hostcall`、`hostservice`、`guest`。
- 保留一个薄的根包 `pluginbridge` facade，通过 type alias、const alias 和 wrapper 函数维持现有宿主、插件样例和动态插件 guest 代码的稳定 import 路径。
- 固定子组件之间的依赖方向，避免 Go 包循环：底层合约与协议包不得反向依赖高层 guest SDK 或宿主实现。
- 将纯实现细节收敛到对应子组件的 `internal` 包中，例如 protobuf wire 工具、WASM section 低层读取工具、guest DTO 绑定辅助等。
- 调整现有宿主和样例插件 import，优先让内部宿主代码使用更精确的子组件包；插件兼容路径继续可用。
- 更新测试覆盖，确保重构前后的 ABI 常量、序列化字节、WASM section 读取、host service payload 编解码和 guest helper 行为不变。
- 不改变 REST API、数据库结构、插件清单语义、动态插件运行时协议语义或用户可见功能。

## Capabilities

### New Capabilities

- `pluginbridge-subcomponent-architecture`: 定义 `pkg/pluginbridge` 子组件化后的公开包结构、依赖边界、兼容 facade 和验证要求。

### Modified Capabilities

- `plugin-runtime-loading`: WASM 自定义段解析能力仍由 pluginbridge 体系集中提供，但不再要求实现固定放在根包文件 `pluginbridge_wasm_section.go` 中。

## Impact

- 受影响后端代码：
  - `apps/lina-core/pkg/pluginbridge/`
  - `apps/lina-core/internal/service/plugin/internal/runtime/`
  - `apps/lina-core/internal/service/plugin/internal/wasm/`
  - `apps/lina-core/internal/service/i18n/`
  - `apps/lina-core/internal/service/apidoc/`
  - `apps/lina-core/pkg/plugindb/`
  - 动态插件样例 `apps/lina-plugins/plugin-demo-dynamic/`
- 受影响测试：
  - `apps/lina-core/pkg/pluginbridge/...`
  - `apps/lina-core/internal/service/plugin/internal/runtime/...`
  - `apps/lina-core/internal/service/plugin/internal/wasm/...`
  - `apps/lina-core/pkg/plugindb/...`
  - `apps/lina-plugins/plugin-demo-dynamic` 的普通 Go 测试与 `wasip1/wasm` 构建验证。
- i18n 影响：
  - 本变更不新增、修改或删除用户可见前端文案、菜单、按钮、表单、表格或 API DTO 文档源文本；预计不需要维护运行时 i18n、插件 manifest i18n 或 apidoc i18n 资源。
- 缓存一致性影响：
  - 本变更不新增业务缓存，不改变插件运行时缓存、i18n 资源缓存或 WASM 编译缓存的权威数据源、失效机制和集群一致性模型。
