## Why

当宿主使用共享启动快照并在 `plugin.autoEnable` 中配置源码插件自动启用时，源码插件安装会写入数据库，但当前启动快照未同步刷新。后续启用阶段仍从快照读到 `installed=0`，导致启动失败并报出 `Plugin is not installed`。

## What Changes

- 修复源码插件启动自动安装后的启动快照同步，确保同一启动编排内后续启用检查读取到最新安装状态。
- 增加覆盖共享启动快照场景的后端单元测试，复现并防止 `plugin.autoEnable` 源码插件自动启用回归。
- 不改变插件配置结构、REST API、数据库结构或用户可见前端交互。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-startup-bootstrap`: 补充启动自动启用过程中生命周期写入必须同步更新共享启动快照的要求。

## Impact

- 受影响后端代码：
  - `apps/lina-core/internal/service/plugin/plugin_lifecycle_source.go`
  - `apps/lina-core/internal/service/plugin/plugin_auto_enable_test.go`
- 受影响测试：
  - `go test ./internal/service/plugin`
- i18n 影响：
  - 本变更不新增、修改或删除前端文案、菜单、API DTO 文档源文本、插件 manifest i18n 或 apidoc i18n 资源。
- 缓存一致性影响：
  - 本变更只修复单次启动编排内的短生命周期启动快照同步；权威数据源仍为数据库，不新增跨请求、跨进程或分布式缓存。
