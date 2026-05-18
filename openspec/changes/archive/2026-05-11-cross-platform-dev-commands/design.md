## Context

当前根目录 `Makefile` 将开发任务拆分到 `hack/makefiles/*.mk`，并在多个目标中嵌入 POSIX Shell 逻辑。`dev`、`stop`、`status`、`build`、`wasm`、`test-go`、`help` 等目标依赖 Bash 语法或 Linux/macOS 工具链；`hack/scripts/*.sh` 也承担了资源复制、进程停止、smoke 测试等逻辑。Windows 默认没有 GNU Make，也不保证存在 Bash、`lsof`、`awk`、`sed`、`nohup`、`kill` 等命令。

仓库已经存在 `hack/tools/build-wasm`、`hack/tools/image-builder`、`hack/tools/runtime-i18n` 等 Go 工具，说明将长生命周期的开发工具沉淀到 `hack/tools/` 与现有工程治理方式一致。本变更应延续该模式，用 Go 承载跨平台任务编排，保留 `Makefile` 作为兼容入口。

## Goals / Non-Goals

**Goals:**

- 提供不依赖 GNU Make 的跨平台开发命令入口。
- 让 Windows 用户可以通过 `go run ./hack/tools/linactl <command>` 执行常用任务。
- 提供根目录 `make.cmd`，使 Windows `cmd.exe` 用户可用 `make <target>`，PowerShell 用户可用 `.\make <target>`，必要时也可显式使用 `.\make.cmd <target>`。
- 保持 Linux/macOS 用户现有 `make <target>` 工作流可用。
- 将复杂 shell 编排逐步迁入 Go 工具，减少平台专属命令依赖。
- 在 `.github/workflows/` 的相关 GitHub Actions 中增加 Windows runner 基本命令验证，防止 Windows 入口退化。
- 保持外部专业工具边界清晰，只包装调用，不重写 `go`、`pnpm`、`docker`、`kubectl`、`gf`、`playwright`。

**Non-Goals:**

- 不重写 GoFrame CLI、Docker、kubectl、pnpm、Playwright 或 Go toolchain。
- 不把所有开发任务迁移到 Node.js 任务系统。
- 不要求 Windows 用户安装 GNU Make、Git Bash、MSYS2 或 Cygwin。
- 不改变后端运行时 API、数据库结构、权限模型、插件运行时契约或前端用户界面。
- 不新增运行时 i18n 文案或业务缓存行为。

## Decisions

### 1. 使用 Go CLI 作为跨平台主入口

新增或扩展 `hack/tools/linactl`，提供统一子命令，例如 `dev`、`stop`、`status`、`build`、`wasm`、`init`、`mock`、`test`、`test-go`、`help`。Go CLI 负责跨平台路径处理、文件复制、进程启动、HTTP readiness、端口检测、日志文件、子命令执行和错误输出。

选择 Go 的原因：

- 项目后端与现有工具链已经以 Go 为基础。
- Go 标准库对文件系统、进程、HTTP、路径和环境变量处理具备跨平台能力。
- Windows 用户在后端开发前已经需要 Go，避免为任务入口再引入 Node 依赖安装。
- 现有 `hack/tools/*` 已采用 Go 工具模式，新增工具不会引入额外治理风格。

备选方案：

- **Node.js CLI**：适合前端任务，但会让后端初始化、状态检查、服务停止等任务依赖 Node 环境与前端包管理，不适合作为全局唯一入口。
- **Taskfile/just/Mage**：可改善任务入口，但仍会引入新的全局工具安装要求；其中 Taskfile/just 的命令体如果继续写 shell 仍无法彻底解决兼容问题，Mage 与 Go CLI 能力接近但会增加新的入口约定。

### 2. `make.cmd` 只作为 Windows 薄包装

根目录提供 `make.cmd`，内容只转发参数到 Go CLI。该文件不承载业务逻辑，不复制复杂判断，不维护另一套任务实现。

设计约定：

- `cmd.exe` 用户可在项目根目录执行 `make dev`。
- PowerShell 用户使用 `.\make dev`；如需避免与本机已安装的其他 `make` 命令混淆，可显式使用 `.\make.cmd dev`。
- `make.cmd` 透传所有参数，例如 `make init confirm=init rebuild=true` 应转为 `linactl init confirm=init rebuild=true`。
- 不默认新增 `make.ps1`，避免 PowerShell Execution Policy 带来的额外说明与维护成本。

### 3. `Makefile` 保留为兼容层

现有 `make <target>` 入口不立即移除。目标实现逐步瘦身为调用 Go CLI，例如 `go run ./hack/tools/linactl build`。这样 Linux/macOS 用户习惯、CI 和文档中的旧入口可以平滑迁移。

迁移顺序应按风险和收益分批：

1. 低风险工具目标：`help`、`status`、`prepare-packed-assets`、`wasm` 插件发现。
2. 开发服务目标：`dev`、`stop`，替换 `nohup`、`lsof`、`pgrep`、`kill` 等逻辑。
3. 构建目标：`build`，在 Go 中编排前端构建、资源复制、Wasm 构建、后端多平台构建。
4. 验证目标：`test-go`、`test-scripts`，将 shell 测试逐步改为 Go 测试或 Go smoke 工具。
5. 子模块目标：`apps/lina-core` 与 `apps/lina-plugins` 的 Makefile 入口薄包装化。

### 4. 兼容 make 风格参数

为了降低迁移成本，Go CLI 需要支持现有 make 风格的 `key=value` 参数，例如：

- `init confirm=init rebuild=true`
- `build platforms=linux/amd64,linux/arm64 verbose=1`
- `wasm p=plugin-demo-dynamic`

CLI 可在内部将 `key=value` 归一化为选项结构；后续也可以支持标准 `--key=value` 写法，但不得要求用户一次性修改已有命令习惯。

### 5. 测试策略以工具行为为中心

该变更属于开发工具链治理，不涉及用户可观察页面，因此不需要 E2E 测试。应优先使用 Go 单元测试和命令级 smoke 测试覆盖：

- 参数解析和 make 风格参数兼容。
- 文件复制结果，例如 packed manifest 只包含应嵌入资源。
- 插件扫描与 dynamic wasm 构建列表识别。
- 服务状态检测与 HTTP readiness。
- Windows 包装脚本只做参数透传。
- `Makefile` 目标是否调用对应 Go CLI。

### 6. GitHub Actions 必须覆盖 Windows 基本命令

`.github/workflows/` 下与构建、测试、工具验证相关的 workflow 需要增加 Windows runner 验证，至少覆盖不会依赖数据库、Docker daemon 或长驻开发服务的基本命令，例如：

- `go run ./hack/tools/linactl help`
- `go run ./hack/tools/linactl status`
- `go run ./hack/tools/linactl prepare-packed-assets`
- `go run ./hack/tools/linactl wasm` 或指定插件的 dry-run/轻量验证模式
- `.\make help`
- `.\make status`

Windows CI 验证应优先使用 `windows-latest`，并明确 shell。`cmd.exe` 场景用于验证 `make.cmd` 可直接作为 `make` 入口；PowerShell 场景用于验证 `.\make` 主用法，必要时保留 `.\make.cmd` 显式后缀用法。对于需要外部服务、数据库或长驻进程的命令，首期应使用轻量 smoke 或 dry-run 模式，避免引入高成本、不稳定的 CI 依赖。

## Risks / Trade-offs

- **风险：一次性迁移所有目标导致回归面过大** → 分批迁移，先将低风险目标移入 Go CLI，再处理进程管理和构建编排。
- **风险：Go CLI 与 Makefile 同时存在造成行为分叉** → Makefile 和 `make.cmd` 只做薄包装，真实逻辑只保留在 Go CLI。
- **风险：Windows 进程树停止语义与 POSIX 不同** → Go CLI 需要尽量以 PID 文件、端口检测、命令特征和子进程启动句柄组合处理；必要时明确 Windows 降级策略，只停止由当前工具启动且可识别的进程。
- **风险：`go run` 每次编译影响命令启动速度** → 初期接受该成本，后续可提供 `go install ./hack/tools/linactl` 或缓存二进制入口，但不作为首期硬要求。
- **风险：`make.cmd` 与真实 GNU Make 命名冲突** → 仅在 Windows 项目根目录作为本地脚本使用；Linux/macOS 仍使用真实 Makefile。
- **风险：文档中双入口说明造成困惑** → 文档按“跨平台推荐入口”和“兼容入口”分组，Windows 示例优先展示 `go run` 与 `make.cmd`。
- **风险：Windows CI 过慢或依赖缺失导致误报** → 首期只验证基本命令、入口转发、文件操作和轻量 smoke；重型构建、数据库初始化、Docker 镜像和开发服务可在后续矩阵中逐步扩展。

## Migration Plan

1. 新增 `hack/tools/linactl` 的命令框架、README 与中英文说明。
2. 新增根目录 `make.cmd`，透传参数到 `linactl`。
3. 迁移低风险目标并补充测试。
4. 迁移开发服务与构建相关目标并补充 smoke 验证。
5. 将根目录 `Makefile` 与子模块 `Makefile` 改为薄包装。
6. 更新 `.github/workflows/`，增加 Windows 基本命令验证。
7. 更新根目录 README/README.zh-CN 和相关工具文档。
8. 运行 Go 工具测试、现有构建/初始化验证、GitHub Actions 本地静态检查和 OpenSpec 校验。

回滚策略：保留当前 Makefile 目标结构直到对应 Go CLI 目标验证完成；若某个目标迁移失败，可单独恢复该目标的旧实现，不影响其他已迁移目标。

## Open Questions

- 首期是否需要提供预编译 `linactl` 二进制下载，还是仅支持 `go run ./hack/tools/linactl`。
- Windows CI 首期是否只覆盖轻量命令，还是同时覆盖 `build` 的最小构建路径。
- `stop` 在 Windows 上是否接受仅停止本工具启动且有 PID 文件记录的进程，还是需要实现更激进的端口占用进程识别。
