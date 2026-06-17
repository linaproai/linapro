## Why

官方插件的 GoFrame 代码生成配置目前位于`apps/lina-plugins/<plugin-id>/backend/hack/config.yaml`，但该配置属于插件级开发工具配置，不属于插件后端运行时代码。将其迁移到插件根`hack/config.yaml`可以让插件工具资源集中在插件根目录，并避免`backend/`目录同时承载后端源码和插件级治理配置。

## What Changes

- 将官方插件代码生成配置从`backend/hack/config.yaml`迁移到插件根`hack/config.yaml`。
- **BREAKING**：`linactl ctrl`和`linactl dao`删除`p=`、`plugin=`和`target=`目标选择器，只保留`dir=`；旧参数必须返回清晰错误。
- `linactl`的 GoFrame 代码生成目标从单一目标目录升级为`workDir`和`configDir`，插件生成继续在`backend/`执行，但配置从插件根`hack/`读取。
- `plugins.check`扫描插件根`hack/config.yaml`，并阻断旧路径`backend/hack/config.yaml`。
- 同步更新 OpenSpec 基线、`.agents/rules/plugin.md`、插件 README、`linactl` README 和内置插件配置文件位置。
- 不保留双路径兼容，避免同一插件存在两个开发期代码生成配置事实源。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `linactl-build-tool-consolidation`：GoFrame 代码生成目标改为显式区分工作目录和配置目录，并收敛目标选择参数。
- `module-decoupling`：官方源码插件的独立代码生成配置位置改为插件根`hack/config.yaml`。
- `plugin-capability-boundary-governance`：插件治理扫描改为读取插件根`hack/config.yaml`，并将旧路径作为违规配置阻断。

## Impact

- 影响代码：`hack/tools/linactl`的 GoFrame 生成封装、命令参数解析、插件治理扫描和相关测试。
- 影响资源：官方插件的开发期`hack/config.yaml`位置。
- 影响文档：OpenSpec 基线、项目插件规则、插件目录 README、`linactl` README。
- 不影响运行时 HTTP API、DTO、前端页面、SQL schema、运行时配置、缓存、数据权限或业务服务语义。
