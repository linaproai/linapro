## MODIFIED Requirements

### Requirement: 安装入口必须提供一致的跨平台能力

系统 SHALL 在`hack/scripts/install/install.sh`提供单一快速安装入口，并将同一文件内容发布到`https://linapro.ai/install.sh`，使用户可以通过`curl -fsSL https://linapro.ai/install.sh | bash`在受支持平台下载`LinaPro`仓库源码。`install.sh`必须是完全自包含的单文件`bash`脚本，职责仅限于下载仓库源码与打印下一步指引：

1. 探测操作系统（macOS、Linux、Windows Git Bash 或 WSL），不支持的平台快速失败。
2. 从`LINAPRO_VERSION`或远程 Git 仓库标签列表解析目标版本。
3. 按解析出的标签把仓库克隆到目标目录，并保留可继续从`origin`拉取发布标签的 Git 仓库状态。
4. 打印明确的后续步骤，指引用户调用`lina-doctor`准备环境，并在环境就绪后执行`make init`初始化项目。

`install.sh`不得分发到平台专属安装脚本。`install.sh`不得执行环境前置检查、包安装命令、`go mod download`、`pnpm install`、`make init`、`make mock`、端口占用探测，或任何超出克隆源码与打印下一步指引的动作。Windows 用户必须在`Git Bash`（随`Git for Windows`提供）或`WSL`中执行同一条命令。

#### Scenario: macOS 用户运行统一安装命令

- **WHEN** 用户在`macOS`上运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `install.sh`通过`uname -s`探测到`Darwin`
- **AND** 按解析出的标签把仓库克隆到目标目录
- **AND** 打印下一步指引，要求用户先调用`lina-doctor`，再执行`make init`
- **AND** 不调用任何其他平台脚本或初始化命令

#### Scenario: Linux 用户运行统一安装命令

- **WHEN** 用户在`Linux`上运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `install.sh`通过`uname -s`探测到`Linux`
- **AND** 按解析出的标签把仓库克隆到目标目录
- **AND** 打印下一步指引，要求用户先调用`lina-doctor`，再执行`make init`
- **AND** 不调用任何其他平台脚本或初始化命令

#### Scenario: Windows 用户在 Git Bash 中运行统一安装命令

- **WHEN** 用户在 Windows 的`Git Bash`中运行`curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `install.sh`通过`uname -s`探测到`MINGW*`、`MSYS*`或`CYGWIN*`
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
| `LINAPRO_VERSION` | 从远程 Git 标签解析出的最高稳定版本 | 要安装的目标版本标签 |
| `LINAPRO_DIR` | `./linapro` | 克隆目标目录 |
| `LINAPRO_SHALLOW` | 未设置 | 设置后使用浅克隆；默认完整克隆以便后续直接通过 Git 标签升级 |
| `LINAPRO_FORCE` | 未设置 | 设置为`1`时，在脚本的安全检查通过后允许替换非空目标目录 |

`LINAPRO_NON_INTERACTIVE`不再需要，因为简化后的`install.sh`不会要求用户做任何交互确认。`LINAPRO_SKIP_MOCK`不再被识别，因为`install.sh`不再触发`make mock`；用户需要时自行显式执行`make mock`。

#### Scenario: 浅克隆开关生效

- **WHEN** 用户设置`LINAPRO_SHALLOW=1`运行安装命令
- **THEN** `install.sh`调用浅克隆参数执行`git clone`
- **AND** 打印提示，说明后续如果 Git 因浅历史限制无法切换到新标签，需要先执行`git fetch --unshallow --tags --force origin`

#### Scenario: 非空目标目录必须显式 force

- **WHEN** 解析出的`LINAPRO_DIR`已经存在且非空
- **AND** `LINAPRO_FORCE`未设置为`1`
- **THEN** `install.sh`在克隆前失败
- **AND** 打印指引，要求用户选择其他`LINAPRO_DIR`或使用`LINAPRO_FORCE=1`重试

#### Scenario: LINAPRO_SKIP_MOCK 不再被识别

- **WHEN** 用户设置`LINAPRO_SKIP_MOCK=1`运行安装命令
- **THEN** `install.sh`忽略该变量
- **AND** 仍然只完成克隆步骤，不执行`make init`或`make mock`

## REMOVED Requirements

### Requirement: 安装后输出必须包含环境健康检查与下一步指引

**Reason**: 环境健康检查已整体迁移到`lina-doctor`技能，由该技能覆盖完整的`LinaPro`编码与测试工具链诊断和修复。在`prereq.sh`中继续保留一份纯提示式检查，会造成同一逻辑双份维护，并让安装路径重新耦合到技能路径。

**Migration**: 本变更后，`install.sh`不再运行任何环境健康检查。安装器不再打印此前的平台专属安装提示（如`brew`、`apt-get`、`winget`）。用户通过 AI 工具（`Claude Code`、`Codex`等）调用`lina-doctor`技能，获得完整环境诊断与逐步自动安装。`install.sh`的下一步输出会明确把`lina-doctor`作为克隆完成后的即时动作。

### Requirement: 平台专属安装脚本必须共享公共 helper 库

**Reason**: 安装入口已收敛为单一自包含`install.sh`，平台专属安装脚本不再存在，因此安装路径也不再需要共享 helper 库。仍被 doctor 技能需要的纯工具函数（`version_ge`、`confirm`、`retry`、`detect_os`）迁移到`.claude/skills/lina-doctor/lib/_common.sh`，仅供 doctor 技能消费。

**Migration**: 从仓库移除`hack/scripts/install/install-macos.sh`、`hack/scripts/install/install-linux.sh`、`hack/scripts/install/install-windows.sh`、`hack/scripts/install/checks/prereq.sh`和`hack/scripts/install/lib/_common.sh`。doctor 技能仍需使用的 helper 函数迁移到`.claude/skills/lina-doctor/lib/_common.sh`。`hack/scripts/install/checks/`与`hack/scripts/install/lib/`目录整体删除。

## ADDED Requirements

### Requirement: 安装入口输出必须指引用户通过 lina-doctor 准备环境

系统 SHALL 在克隆成功后打印多行“下一步”消息，指引用户：

1. 进入克隆后的项目目录（如`cd ./linapro`）。
2. 通过 AI 工具调用`lina-doctor`技能，检查并安装开发依赖（如`ask Claude Code "run lina-doctor to set up my LinaPro environment"`）。
3. 在`lina-doctor`报告成功后执行`make init`，再执行`make dev`启动开发环境。

输出必须包含克隆项目的绝对路径、默认管理员账号（`admin`/`admin123`），以及后续命令的推荐顺序。输出不得打印任何环境工具安装提示（如`brew install go`）；环境安装指引只由`lina-doctor`技能负责。

#### Scenario: 克隆成功后下一步消息指向 lina-doctor

- **WHEN** `install.sh`按解析出的标签完成`LinaPro`仓库的`git clone`
- **THEN** stdout 包含克隆后的绝对项目路径
- **AND** stdout 包含在运行`make init`前调用`lina-doctor`技能的指引
- **AND** stdout 不包含任何平台专属包安装命令

#### Scenario: 下一步输出在不同平台保持一致

- **WHEN** `install.sh`在 macOS、Linux 或 Windows Git Bash 中成功完成
- **THEN** 下一步消息文本在各平台保持一致（仅项目路径可不同）
- **AND** 消息以相同的规范顺序引用`lina-doctor`和`make init`

### Requirement: 安装入口必须支持基于 Git 标签的后续升级

系统 SHALL 默认使用可持续升级的 Git 仓库克隆方式安装源码。未设置`LINAPRO_VERSION`时，安装器必须通过`git ls-remote --tags --refs`读取远程标签，并从符合`v<major>.<minor>.<patch>`格式的稳定标签中选择语义版本号最高的标签作为安装目标。安装完成后，目标目录必须保留`origin`远程地址，并配置为后续可以通过`git fetch --tags --force origin`拉取新发布标签。

安装器 SHOULD 默认执行完整克隆，以支持后续直接通过 Git 拉取新标签升级。仅当用户显式设置`LINAPRO_SHALLOW=1`时才使用浅克隆，并且必须提示浅克隆在首次升级时可能需要执行`git fetch --unshallow --tags --force origin`。

#### Scenario: 默认安装检出最高稳定标签

- **WHEN** 远程仓库同时存在`v0.0.1`、`v0.0.2`和非稳定标签`latest`
- **AND** 用户未设置`LINAPRO_VERSION`
- **THEN** `install.sh`选择`v0.0.2`作为安装版本
- **AND** 目标目录的`HEAD`检出到`v0.0.2`

#### Scenario: 安装后可拉取新标签升级

- **WHEN** 用户通过默认安装得到本地仓库
- **AND** `origin`后续新增发布标签`v0.0.3`
- **THEN** 用户可以在目标目录执行`git fetch --tags --force origin`
- **AND** 可以执行`git checkout --detach v0.0.3`切换到新标签内容
