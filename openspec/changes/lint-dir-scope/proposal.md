## Why

本地与 Agent 迭代修改单个后端组件后，当前`make lint`只能按`plugins=0|1`选择宿主或插件完整工作区，仍会对工作区内全部`Go module`执行`./...`扫描。全量静态检查反馈周期过长，降低了开发闭环效率，也削弱了“改完就 lint”门禁的可执行性。

## What Changes

- 为`linactl lint.go`与`make lint`/`make lint.go`增加可选`dir=<path>`参数，将扫描范围收敛到目标路径所属的单个`Go module`。
- 支持常见组件路径解析：`apps/lina-core`、`hack/tools/linactl`、`apps/lina-plugins/<plugin-id>`、`apps/lina-plugins/<plugin-id>/backend`，以及任意位于某`go.mod`之下的子目录。
- 未传`dir`时保持现有全量行为（`plugins=0|1`自动/显式模式不变），确保 CI 与审查门禁不受影响。
- 输出日志明确标识定向范围（`scope=dir`、目标 module），避免误以为已跑全量。
- 同步更新`linactl`中英文 README 与`.agents/rules/backend-go.md`中的 lint 使用说明。
- **不引入**包级`packages=`筛选（本期非目标）；**不改变** CI 默认全量策略。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `go-static-lint-governance`：增加基于`dir=`的组件/module 定向静态检查能力，并要求文档与日志标明定向范围。

## Impact

- 代码：`hack/tools/linactl/command_lint.go.go`、`hack/tools/linactl/command.go`、`hack/makefiles/lint.mk`、相关单测
- 文档：`hack/tools/linactl/README.md`、`README.zh-CN.md`、`.agents/rules/backend-go.md`
- 规范：`openspec/specs/go-static-lint-governance`
- 无运行时服务、API、数据库、i18n 资源、缓存与数据权限影响
- 跨平台：参数解析与路径解析继续走 Go 标准库，Makefile 仅透传`dir`
