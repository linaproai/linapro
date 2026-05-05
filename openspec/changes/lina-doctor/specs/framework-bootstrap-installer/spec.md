## MODIFIED Requirements

### Requirement: 安装入口必须提供一致的跨平台能力

系统 SHALL 在`hack/scripts/install/bootstrap.sh`提供单一快速安装入口，并将同一文件内容发布到`https://linapro.ai/install.sh`，使用户可以通过`curl -fsSL https://linapro.ai/install.sh | bash`在受支持平台下载`LinaPro`仓库源码。`bootstrap.sh`必须是完全自包含的单文件`bash`脚本，职责仅限于下载仓库源码与打印下一步指引：

1. 探测操作系统（macOS、Linux、Windows Git Bash 或 WSL），不支持的平台快速失败。
2. 从`LINAPRO_VERSION`或 GitHub`releases/latest`重定向解析目标版本。
3. 按解析出的标签把仓库克隆到目标目录。
4. 打印明确的后续步骤，指引用户调用`lina-doctor`准备环境，并在环境就绪后执行`make init`初始化项目。

`bootstrap.sh`不得分发到平台专属安装脚本。`bootstrap.sh`不得执行环境前置检查、包安装命令、`go mod download`、`pnpm install`、`make init`、`make mock`、端口占用探测，或任何超出克隆源码与打印下一步指引的动作。Windows 用户必须在`Git Bash`（随`Git for Windows`提供）或`WSL`中执行同一条命令。

#### Scenario: macOS 用户运行统一安装命令

- **WHEN** 用户在`macOS`上运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `bootstrap.sh`通过`uname -s`探测到`Darwin`
- **AND** 按解析出的标签把仓库克隆到目标目录
- **AND** 打印下一步指引，要求用户先调用`lina-doctor`，再执行`make init`
- **AND** 不调用任何其他平台脚本或初始化命令

#### Scenario: Linux 用户运行统一安装命令

- **WHEN** 用户在`Linux`上运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `bootstrap.sh`通过`uname -s`探测到`Linux`
- **AND** 按解析出的标签把仓库克隆到目标目录
- **AND** 打印下一步指引，要求用户先调用`lina-doctor`，再执行`make init`
- **AND** 不调用任何其他平台脚本或初始化命令

#### Scenario: Windows 用户在 Git Bash 中运行统一安装命令

- **WHEN** 用户在 Windows 的`Git Bash`中运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `bootstrap.sh`通过`uname -s`探测到`MINGW*`、`MSYS*`或`CYGWIN*`
- **AND** 按解析出的标签把仓库克隆到目标目录
- **AND** 打印下一步指引，要求用户先调用`lina-doctor`，再执行`make init`
- **AND** 不调用任何其他平台脚本或初始化命令

#### Scenario: Windows 用户在 PowerShell 中运行统一安装命令

- **WHEN** 用户在没有可用`bash`的原生`PowerShell`或`CMD`中运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** 命令因为`bash`不在`PATH`中而失败
- **AND** 系统文档指引用户切换到`Git Bash`或`WSL`后重试

### Requirement: 安装入口必须支持常用环境变量覆盖

系统 SHALL 支持以下环境变量控制安装行为，并为每个变量记录默认值：

| Variable | Default | Effect |
| --- | --- | --- |
| `LINAPRO_VERSION` | 从 GitHub 解析 | 要安装的目标版本标签 |
| `LINAPRO_DIR` | `./linapro` | 克隆目标目录 |
| `LINAPRO_SHALLOW` | 未设置 | 设置后使用`--depth 1`执行浅克隆 |
| `LINAPRO_FORCE` | 未设置 | 设置为`1`时，在脚本的安全检查通过后允许替换非空目标目录 |

`LINAPRO_NON_INTERACTIVE`不再需要，因为简化后的`bootstrap.sh`不会要求用户做任何交互确认。`LINAPRO_SKIP_MOCK`不再被识别，因为`bootstrap.sh`不再触发`make mock`；用户需要时自行显式执行`make mock`。

#### Scenario: 浅克隆开关生效

- **WHEN** 用户设置`LINAPRO_SHALLOW=1`运行安装命令
- **THEN** `bootstrap.sh`调用`git clone --branch <tag> --depth 1`
- **AND** 打印提示，说明后续首次使用`lina-upgrade`技能前需要执行`git fetch --unshallow`

#### Scenario: 非空目标目录必须显式 force

- **WHEN** 解析出的`LINAPRO_DIR`已经存在且非空
- **AND** `LINAPRO_FORCE`未设置为`1`
- **THEN** `bootstrap.sh`在克隆前失败
- **AND** 打印指引，要求用户选择其他`LINAPRO_DIR`或使用`LINAPRO_FORCE=1`重试

#### Scenario: LINAPRO_SKIP_MOCK 不再被识别

- **WHEN** 用户设置`LINAPRO_SKIP_MOCK=1`运行安装命令
- **THEN** `bootstrap.sh`忽略该变量
- **AND** 仍然只完成克隆步骤，不执行`make init`或`make mock`

## REMOVED Requirements

### Requirement: 安装后输出必须包含环境健康检查与下一步指引

**Reason**: 环境健康检查已整体迁移到`lina-doctor`技能，由该技能覆盖完整的`LinaPro`编码与测试工具链诊断和修复。在`prereq.sh`中继续保留一份纯提示式检查，会造成同一逻辑双份维护，并让安装路径重新耦合到技能路径。

**Migration**: 本变更后，`bootstrap.sh`不再运行任何环境健康检查。安装器不再打印此前的平台专属安装提示（如`brew`、`apt-get`、`winget`）。用户通过 AI 工具（`Claude Code`、`Codex`等）调用`lina-doctor`技能，获得完整环境诊断与逐步自动安装。`bootstrap.sh`的下一步输出会明确把`lina-doctor`作为克隆完成后的即时动作。

### Requirement: 平台专属安装脚本必须共享公共 helper 库

**Reason**: 安装入口已收敛为单一自包含`bootstrap.sh`，平台专属安装脚本不再存在，因此安装路径也不再需要共享 helper 库。仍被 doctor 技能需要的纯工具函数（`version_ge`、`confirm`、`retry`、`detect_os`）迁移到`.claude/skills/lina-doctor/lib/_common.sh`，仅供 doctor 技能消费。

**Migration**: 从仓库移除`hack/scripts/install/install-macos.sh`、`hack/scripts/install/install-linux.sh`、`hack/scripts/install/install-windows.sh`、`hack/scripts/install/checks/prereq.sh`和`hack/scripts/install/lib/_common.sh`。doctor 技能仍需使用的 helper 函数迁移到`.claude/skills/lina-doctor/lib/_common.sh`。`hack/scripts/install/checks/`与`hack/scripts/install/lib/`目录整体删除。

## ADDED Requirements

### Requirement: Bootstrap 输出必须指引用户通过 lina-doctor 准备环境

系统 SHALL 在克隆成功后打印多行“下一步”消息，指引用户：

1. 进入克隆后的项目目录（如`cd ./linapro`）。
2. 通过 AI 工具调用`lina-doctor`技能，检查并安装开发依赖（如`ask Claude Code "run lina-doctor to set up my LinaPro environment"`）。
3. 在`lina-doctor`报告成功后执行`make init`，再执行`make dev`启动开发环境。

输出必须包含克隆项目的绝对路径、默认管理员账号（`admin`/`admin123`），以及后续命令的推荐顺序。输出不得打印任何环境工具安装提示（如`brew install go`）；环境安装指引只由`lina-doctor`技能负责。

#### Scenario: 克隆成功后下一步消息指向 lina-doctor

- **WHEN** `bootstrap.sh`按解析出的标签完成`LinaPro`仓库的`git clone`
- **THEN** stdout 包含克隆后的绝对项目路径
- **AND** stdout 包含在运行`make init`前调用`lina-doctor`技能的指引
- **AND** stdout 不包含任何平台专属包安装命令

#### Scenario: 下一步输出在不同平台保持一致

- **WHEN** `bootstrap.sh`在 macOS、Linux 或 Windows Git Bash 中成功完成
- **THEN** 下一步消息文本在各平台保持一致（仅项目路径可不同）
- **AND** 消息以相同的规范顺序引用`lina-doctor`和`make init`
