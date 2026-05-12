## Why

当前项目的开发与交付入口主要依赖 `make`、POSIX Shell 语法以及 `lsof`、`awk`、`sed`、`nohup`、`kill` 等 Linux/macOS 常见工具。Windows 默认不内置 GNU Make，也不保证存在 Bash/POSIX 工具链，导致 Windows 用户即使安装了 Go、Node.js、Docker 等核心依赖，仍难以顺利执行项目常用命令。

本变更旨在提供一套跨平台的一等公民命令入口，让 Windows、macOS、Linux 用户可以使用一致的开发体验，同时保留现有 `make` 入口作为兼容层。

## What Changes

- 新增跨平台开发命令能力，以 Go 工具承载常用任务编排逻辑，避免在核心路径中依赖 POSIX Shell。
- 在项目根目录提供 Windows `make.cmd` 薄包装入口，使 `cmd.exe` 用户可使用 `make dev`、`make build` 等近似体验，PowerShell 用户可使用 `.\make dev`，必要时也可显式使用 `.\make.cmd dev`。
- 保留现有 `Makefile`，但逐步将复杂目标改为调用跨平台 Go CLI，降低 shell 方言和平台工具差异影响。
- 将 `dev`、`stop`、`status`、`build`、`wasm`、`init`、`mock`、`test`、`test-go`、`help`、`cli.install` 等常用目标纳入跨平台入口规划。
- 对仍需调用的外部专业工具保持包装而非重写，例如 `go`、`pnpm`、`docker`、`kubectl`、`gf`、`playwright`。
- 更新 `.github/workflows/` 下相关 GitHub Actions，在 Windows runner 中验证跨平台命令入口与 `make.cmd` 的基本命令可用性。
- 更新根目录文档中的命令说明，明确跨平台推荐入口、Windows `cmd.exe` 与 PowerShell 使用方式，以及兼容 `make` 入口。

## Capabilities

### New Capabilities

- `cross-platform-dev-commands`: 定义项目跨平台开发命令入口、Windows `make.cmd` 兼容入口、命令参数兼容、外部工具调用边界、测试与文档要求。

### Modified Capabilities

- 无。

## Impact

- 影响根目录 `Makefile`、`hack/makefiles/`、`hack/scripts/`、`apps/lina-core/Makefile`、`apps/lina-plugins/Makefile` 等开发工具入口。
- 可能新增 `hack/tools/linactl` 或等价 Go CLI，并更新 `go.work`、工具 README、根目录 README/README.zh-CN。
- 可能新增根目录 `make.cmd`，用于 Windows 命令行兼容入口。
- 影响 `.github/workflows/` 下与构建、测试、工具验证相关的 GitHub Actions，需要增加 Windows 基本命令验证。
- 不涉及后端运行时 HTTP API、数据库 Schema、权限模型、数据权限、业务缓存、插件运行时契约或前端用户界面。
- i18n 影响：本变更只涉及开发工具命令与技术文档，不新增运行时菜单、按钮、表单、接口消息或插件 manifest 文案；无需修改运行时 i18n JSON、manifest i18n 或 apidoc i18n 资源。
- 缓存一致性影响：本变更不新增或修改业务运行时缓存；无需设计分布式缓存失效或跨实例一致性机制。
