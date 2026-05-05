## ADDED Requirements

### Requirement: 环境医生技能必须可被 Claude Code 作为 skill 加载

系统 SHALL 在`.claude/skills/lina-doctor/`下提供`lina-doctor`技能，使`Claude Code`、`Codex`等 AI 工具可以加载并调用该技能，诊断和修复`LinaPro`开发环境。该技能必须包含`SKILL.md`入口文件，其 frontmatter 必须声明技能名称、单行中文描述，并覆盖至少以下双语触发词：`"environment check"`、`"install dependencies"`、`"fix environment"`、`"环境检查"`、`"安装依赖"`、`"工具链"`、`"缺少 go"`、`"缺少 node"`、`"缺少 openspec"`以及字面技能名`"lina-doctor"`。`SKILL.md`正文与所有`references/**`文档必须使用简体中文编写，以便维护者审阅；命令名、环境变量名、JSON 字段名和外部工具标识保持英文原文。该技能必须能够独立于`bootstrap.sh`安装入口被调用，不得硬连接到 install 路径。

#### Scenario: AI 工具通过自然语言调用技能

- **WHEN** 用户向 AI 工具提出“check my LinaPro environment”或“缺少 openspec 帮我装一下”
- **THEN** AI 工具加载`.claude/skills/lina-doctor/SKILL.md`
- **AND** 无需额外设置即可运行文档化的工作流

#### Scenario: 技能不耦合安装入口

- **WHEN** 用户首次运行`bash hack/scripts/install/bootstrap.sh`
- **THEN** 安装路径完成时不会调用`lina-doctor`
- **AND** 不运行`prereq.sh`、平台专属安装脚本、环境探测、包安装或项目初始化

### Requirement: 环境医生必须覆盖 LinaPro 编码与测试工具链

系统 SHALL 在 macOS、Linux、Windows Git Bash 或 WSL 上诊断并在用户确认后安装以下工具：

| Tool | Minimum Version | Install Channel Priority |
| --- | --- | --- |
| `go` | `1.22` | macOS: `brew`; Linux: `apt > dnf > yum > pacman`; Windows: `winget > scoop > choco` |
| `node` | `20.19` | 检测到 nvm / fnm / volta 时优先使用，否则使用操作系统原生包管理器 |
| `pnpm` | `8` | `npm i -g pnpm` |
| `git` | 只检查存在性 | 操作系统原生包管理器 |
| `make` | 只检查存在性 | 操作系统原生包管理器 |
| `openspec` | 只检查存在性 | macOS: `brew install openspec`，失败回退到`npm i -g @fission-ai/openspec@latest`; Linux/Windows: `npm i -g @fission-ai/openspec@latest` |
| `gf` (GoFrame CLI) | 只检查存在性 | `go install github.com/gogf/gf/v2/cmd/gf@latest` |
| `Playwright browsers` | 只检查存在性 | `cd hack/tests && pnpm exec playwright install` |
| `goframe-v2` Claude skill | 只检查存在性 | `npx skills add github.com/gogf/skills -g` |

doctor 不得安装或检查任何 MySQL server 或 MySQL client 工具。doctor 不得安装或检查`Docker`、`VS Code`、`Claude Code`、`Codex`或其他开发协作工具。

#### Scenario: 在干净机器上诊断全部九个工具

- **WHEN** 用户在只安装了`git`的新克隆仓库中调用`lina-doctor`
- **THEN** 诊断报告识别其余八个工具缺失或低于最低版本
- **AND** 报告中不列出`mysql`、`Docker`或任何 IDE 相关工具

#### Scenario: MySQL 明确不在范围内

- **WHEN** 用户要求`lina-doctor`“install MySQL”
- **THEN** 技能说明 MySQL 不属于本技能范围
- **AND** 指向人工安装 MySQL 的参考资料，不尝试安装

### Requirement: Doctor 必须运行诊断检查脚本并输出 JSON

系统 SHALL 提供可执行脚本`.claude/skills/lina-doctor/scripts/doctor-check.sh`，探测九个工具链目标、宿主操作系统、选定包管理器、用户 shell 和 PATH 问题，并向 stdout 输出单个 JSON 文档。该 JSON 对象必须包含`os`、`package_manager`、`shell`、`repo_root_detected`、`tools`、`path_issues`和`mirror_hints`字段。`tools`下的每个条目必须包含`present`（boolean）、`version`（string 或`null`）、`min_version`（string 或`null`）和`ok`（boolean）字段。该脚本必须是只读的，不得修改用户环境、安装任何工具，或写入`/tmp`及脚本自身工作目录之外的任何文件。

脚本必须在所有关键工具满足最低版本且可选目标已满足或被显式跳过时返回退出码`0`；在一个或多个关键工具缺失或低于最低版本时返回`1`；在仅存在可选问题（PATH 问题、镜像建议、缺失 Playwright browsers 或缺失`goframe-v2`技能）时返回`2`。关键工具为`go`、`node`、`pnpm`、`git`、`make`、`openspec`和`gf`；可选目标为`Playwright browsers`和`goframe-v2`。

#### Scenario: Check-only 模式输出结构化 JSON

- **WHEN** 用户运行`bash .claude/skills/lina-doctor/scripts/doctor-check.sh --check-only`
- **THEN** stdout 包含符合文档化 schema 的单个有效 JSON 文档
- **AND** stderr 不包含会破坏 JSON 消费方的诊断噪声
- **AND** 脚本不调用任何包管理器，也不修改仓库内任何文件

#### Scenario: 退出码反映严重度

- **WHEN** 用户缺少`pnpm`和`gf`
- **THEN** `doctor-check.sh`返回退出码`1`
- **AND** JSON 将`pnpm.ok`和`gf.ok`标记为`false`

### Requirement: Doctor 必须按拓扑顺序计算安装计划

系统 SHALL 生成安装计划，确保依赖前置工具的目标排在其前置工具之后。计划必须满足：

- 仅在`go`已满足或同计划中更早安装时安排`gf`
- 仅在`node`已满足或同计划中更早安装时安排`pnpm`
- 通过`npm`安装`openspec`时，仅在`node`已满足或同计划中更早安装时安排
- 仅在`pnpm`已满足或同计划中更早安装，并且已探测到仓库根目录时安排`Playwright browsers`

计划不得包含前置条件仍缺失且同计划中无法更早满足的工具。当某个平台没有可用安装通道导致前置条件无法满足时，计划必须排除所有依赖项，并在进入安装前向用户暴露阻塞问题。

#### Scenario: 典型计划中的拓扑顺序

- **WHEN** 用户缺少`go`、`gf`、`node`、`pnpm`和`Playwright browsers`
- **THEN** 计划将`go`排在`gf`之前
- **AND** 将`node`排在`pnpm`之前
- **AND** 将`pnpm`排在`Playwright browsers`之前

#### Scenario: 前置条件不可满足时排除依赖项

- **WHEN** 宿主 OS 不受支持且没有 Go 安装通道
- **THEN** 计划同时排除`go`和`gf`
- **AND** 在进入安装模式前打印 escalation 消息

### Requirement: Doctor 默认必须逐步请求用户确认

系统 MUST 在执行每条安装命令前请求用户显式确认（`y/N`）。提示必须展示即将执行的精确命令、选定包管理器，以及是否需要`sudo`或提升权限。Linux 包管理器命令可在计划中包含`sudo`，但该命令必须在执行前展示，并且只能在用户确认后执行，或在用户通过`LINAPRO_DOCTOR_NON_INTERACTIVE=1`明确选择非交互执行后执行。当设置`LINAPRO_DOCTOR_NON_INTERACTIVE=1`时，系统必须跳过逐项确认，但仍必须在执行前打印每条命令。系统不得静默提升权限，也不得在用户确认后再临时注入`sudo`。

#### Scenario: 交互模式逐步确认

- **WHEN** 用户在未设置`LINAPRO_DOCTOR_NON_INTERACTIVE`时调用`lina-doctor`
- **THEN** 系统为每个待安装工具打印计划命令并等待`y/N`输入
- **AND** 用户未确认时跳过该工具

#### Scenario: 非交互模式打印命令但跳过提示

- **WHEN** 用户设置`LINAPRO_DOCTOR_NON_INTERACTIVE=1`调用`lina-doctor`
- **THEN** 系统在执行前打印每条计划命令
- **AND** 不暂停等待用户输入
- **AND** 仍应用与交互模式相同的关键工具与可选目标失败策略

### Requirement: Doctor 必须自动探测平台包管理器

系统 SHALL 探测操作系统与可用包管理器，并按平台优先级选择一个包管理器。macOS 必须使用`brew`。Linux（默认包括 WSL）必须从`apt-get`、`dnf`、`yum`、`pacman`中选择第一个可用项。Windows Git Bash 必须从`winget`、`scoop`、`choco`中选择第一个可用的 Windows 包管理器。探测结果必须在任何安装开始前写入计划输出。没有找到受支持包管理器时，系统必须 escalation，不得尝试安装。运行在 WSL 中时，只有在用户明确确认希望安装到 Windows host 后，才允许通过 PowerShell 安装 Windows host 工具。

#### Scenario: macOS 使用 Homebrew

- **WHEN** `lina-doctor`在安装了`brew`的 macOS 上运行
- **THEN** 计划声明`package_manager: brew`
- **AND** 通过`brew install go`安装 Go

#### Scenario: Linux 存在 apt-get 时选择 apt-get

- **WHEN** `lina-doctor`在同时存在`apt-get`和`yum`的 Debian 系 Linux 上运行
- **THEN** 计划因优先级更高而选择`apt-get`
- **AND** 通过`sudo apt-get install -y make`安装 Make

#### Scenario: Windows Git Bash 使用 winget 并包装 PowerShell

- **WHEN** `lina-doctor`在 Windows 11 的 Git Bash 中运行且`winget`可用
- **THEN** 计划声明`package_manager: winget`
- **AND** 安装命令包装为`powershell.exe -NoProfile -Command "winget install <package>"`

#### Scenario: WSL 默认使用 Linux 包管理器

- **WHEN** `lina-doctor`在可用`apt-get`的 Ubuntu WSL 中运行
- **THEN** 计划声明`package_manager: apt-get`
- **AND** 安装命令默认作用于 WSL Linux 环境
- **AND** 除非用户明确请求，否则不通过 PowerShell 安装 Windows host 工具

### Requirement: Doctor 必须尊重已存在的 Node 版本管理器

系统 SHALL 通过探测`NVM_DIR`、`FNM_DIR`、`VOLTA_HOME`环境变量和`PATH`中的对应可执行项，识别`nvm`、`fnm`或`volta`。当任一版本管理器存在时，doctor 必须使用该版本管理器安装或切换 Node，而不是调用操作系统原生包管理器安装 Node。doctor 不得删除或替换任何既有版本管理器安装。

#### Scenario: 检测到 nvm 时通过 nvm 安装 Node

- **WHEN** `lina-doctor`检测到`$NVM_DIR`已设置且`nvm`可调用或`$NVM_DIR/nvm.sh`存在
- **THEN** 计划为 Node 安排加载`nvm.sh`后执行`nvm install 20 && nvm use 20`
- **AND** 不调用`brew install node`或`winget install OpenJS.NodeJS`

#### Scenario: 未检测到版本管理器时使用 OS 原生安装

- **WHEN** 未检测到`nvm`、`fnm`或`volta`
- **THEN** 计划通过平台包管理器安排 Node 安装

### Requirement: Doctor 必须自检 PATH 并暴露持久修复命令

系统 SHALL 在每个工具安装完成后重新检查该工具是否能通过`command -v`访问。当工具已经安装但不在`PATH`中时（常见于`$HOME/go/bin`中的`gf`或 npm global prefix 中的`pnpm`），系统必须：

1. 在当前 doctor 进程中把缺失目录导出到`PATH`，使后续步骤可使用该工具。
2. 打印一行用户可复制到当前 shell 的`export`命令。
3. 针对检测到的 shell rc 文件（`~/.zshrc`、`~/.bashrc`或 PowerShell`$PROFILE`）打印一行追加片段。
4. 不自动修改任何 rc 文件或 shell profile。

系统不得因 PATH 问题阻塞工作流；必须以 warning 记录并继续。

#### Scenario: gf 已安装但 GOBIN 不在 PATH 中

- **WHEN** `lina-doctor`完成`go install github.com/gogf/gf/v2/cmd/gf@latest`后，`command -v gf`仍无结果
- **THEN** 系统将`$HOME/go/bin`导出到当前进程的`PATH`供后续步骤使用
- **AND** 打印`export PATH="$HOME/go/bin:$PATH"`供用户复制
- **AND** 打印一行追加到用户`~/.zshrc`或`~/.bashrc`的片段
- **AND** 不修改任何 rc 文件

#### Scenario: 用户 rc 文件永不被自动写入

- **WHEN** 检测到 PATH 问题
- **THEN** 系统永不以写模式打开`~/.zshrc`、`~/.bashrc`或 PowerShell`$PROFILE`
- **AND** 即使用户已经确认安装步骤，也不会向这些文件追加内容

### Requirement: Doctor 必须探测网络镜像并仅作为建议暴露

系统 SHALL 探测`GOPROXY`、当前 npm registry（通过`npm config get registry`）和`PLAYWRIGHT_DOWNLOAD_HOST`。当任一值未设置、为空或为默认上游值时，doctor 必须在计划输出中以可选环境变量赋值的形式暴露镜像建议，但不得修改用户环境、shell rc 文件或任何全局配置。镜像建议必须同时说明默认上游选项与已知中国区镜像选项。

#### Scenario: GOPROXY 未设置时计划中出现镜像建议

- **WHEN** `lina-doctor`检测到`GOPROXY`为空或`off`
- **THEN** 计划输出包含可选建议`export GOPROXY=https://goproxy.cn,direct`
- **AND** 不自动导出`GOPROXY`

#### Scenario: 保留用户已有 GOPROXY

- **WHEN** 用户已有`GOPROXY=https://my-corp-proxy.example.com`
- **THEN** 计划输出不包含 GOPROXY 建议
- **AND** 不覆盖既有值

### Requirement: Doctor 必须在每个工具安装后立即验证

系统 SHALL 在每个安装步骤结束后立即重新调用工具级检查（`command -v <tool>`以及适用场景下的版本比较）。如果关键工具验证失败，系统必须在该处停止工作流并输出 escalation 报告，不得继续后续工具。如果可选目标验证失败，系统必须输出非阻塞 escalation 报告，并继续执行后续安装目标。

#### Scenario: 成功安装后验证通过

- **WHEN** `brew install go`返回退出码`0`且`go version`报告`go1.22.0`
- **THEN** 验证通过
- **AND** 工作流进入下一个计划工具

#### Scenario: 验证失败时停止工作流

- **WHEN** `npm i -g pnpm`报告成功，但之后`command -v pnpm`无结果
- **THEN** 工作流在安装`openspec`或`Playwright browsers`等依赖项之前停止
- **AND** 输出 escalation 报告

### Requirement: Doctor 必须在失败时输出结构化 escalation 报告

系统 SHALL 在任一安装步骤或验证失败时输出 escalation 报告，包含：

- 失败工具名
- 选定包管理器与失败的精确命令
- 失败命令 stdout/stderr 的末尾 50 行
- 从`network`、`permission`、`package_not_found`、`shim_conflict`、`unknown`中推断出的根因类别
- 平台相关的推荐人工动作

该报告必须可读，同时也应足够结构化，使 AI 工具能够无歧义提取失败工具、根因和推荐动作。

#### Scenario: go install gf 时发生网络失败

- **WHEN** `go install github.com/gogf/gf/v2/cmd/gf@latest`连接`proxy.golang.org`超时
- **THEN** escalation 报告识别失败工具为`gf`
- **AND** 将根因归类为`network`
- **AND** 建议重试前设置`GOPROXY=https://goproxy.cn,direct`

#### Scenario: npm install 发生权限失败

- **WHEN** `npm i -g pnpm`失败并输出`EACCES: permission denied`
- **THEN** escalation 报告将根因归类为`permission`
- **AND** 建议配置不需要 root 的 npm prefix，或在确需时使用`sudo`

### Requirement: Doctor 执行必须在重复运行时保持幂等

系统 SHALL 跳过最近一次`doctor-check.sh`输出中`ok`字段为`true`的工具。doctor 不得重新安装、升级或降级已经满足最低版本要求的工具。在完全满足的环境中重复调用`lina-doctor`，必须成功退出且不调用任何包管理器。

#### Scenario: 满足环境中 doctor 是 no-op

- **WHEN** 所有关键工具和可选目标都已经安装并满足最低版本或存在性要求
- **THEN** 工作流运行`doctor-check.sh`、打印“All tools satisfied”，并以退出码`0`退出
- **AND** 不调用任何包管理器

#### Scenario: 部分环境跳过已满足工具

- **WHEN** `go`、`node`、`pnpm`、`git`、`make`、`openspec`已经满足，但缺少`gf`和 Playwright browsers
- **THEN** 安装计划只包含`gf`和`Playwright browsers`
- **AND** 不重新安装已经满足的六个工具

### Requirement: Doctor 必须支持可选 Playwright browsers 步骤

系统 SHALL 默认把`Playwright browsers`（通过`pnpm exec playwright install`下载 Chromium）作为可安装目标。当设置`LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1`时，doctor 必须从计划中省略该目标。当未探测到仓库根目录（doctor 未运行在包含`apps/lina-core`、`apps/lina-vben`与`hack/tests/`的仓库中）时，必须自动跳过 Playwright 步骤。

doctor 必须把该步骤失败视为非阻塞。Playwright browser 安装或验证失败必须输出非阻塞 escalation 报告，不得将关键`LinaPro`编码环境标记为损坏。

#### Scenario: 默认安装包含 Playwright

- **WHEN** 用户在仓库根目录运行`lina-doctor`且未设置`LINAPRO_DOCTOR_SKIP_PLAYWRIGHT`
- **THEN** 计划包含`Playwright browsers`
- **AND** 在 pnpm 验证后运行`cd hack/tests && pnpm exec playwright install`

#### Scenario: 通过环境变量显式跳过

- **WHEN** 用户设置`LINAPRO_DOCTOR_SKIP_PLAYWRIGHT=1`运行`lina-doctor`
- **THEN** 计划不包含`Playwright browsers`
- **AND** 诊断 JSON 仍报告其存在或缺失状态

#### Scenario: 未探测到仓库时自动跳过 Playwright

- **WHEN** `lina-doctor`在`LinaPro`仓库根目录之外被调用
- **THEN** 计划排除`Playwright browsers`
- **AND** 诊断标记`repo_root_detected: false`

### Requirement: Doctor 必须通过 npx skills add 安装 goframe-v2 Claude skill

系统 SHALL 默认把`goframe-v2` Claude skill 作为可安装目标，安装命令为`npx skills add github.com/gogf/skills -g`。安装产物必须落到用户全局 Claude skills 目录（`~/.claude/skills/goframe-v2/`），而不是`LinaPro`仓库内的任何项目级目录。doctor 必须通过检查`~/.claude/skills/goframe-v2/SKILL.md`存在来验证安装成功。

当设置`LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1`时，doctor 必须从计划中省略该目标。doctor 必须把该步骤失败视为非阻塞，诊断继续后续目标，并输出 escalation 报告，明确`goframe-v2`步骤是非阻塞可选失败，而不是关键环境失败。

doctor 必须仅在`node`（因此也包含`npx`）验证完成后安排该目标，以满足拓扑安装顺序。

#### Scenario: 默认安装包含 goframe-v2

- **WHEN** 用户运行`lina-doctor`且未设置`LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL`
- **AND** `node`已经满足（预安装或本次运行中安装完成）
- **THEN** 计划将`goframe-v2`排在`node`之后
- **AND** 安装阶段运行`npx skills add github.com/gogf/skills -g`
- **AND** 安装后验证`~/.claude/skills/goframe-v2/SKILL.md`存在

#### Scenario: 通过环境变量显式跳过

- **WHEN** 用户设置`LINAPRO_DOCTOR_SKIP_GOFRAME_SKILL=1`运行`lina-doctor`
- **THEN** 计划不包含`goframe-v2`
- **AND** 诊断 JSON 仍报告其存在或缺失状态

#### Scenario: goframe-v2 安装失败是非阻塞的

- **WHEN** `npx skills add github.com/gogf/skills -g`失败（例如上游包不可用）
- **THEN** doctor 输出非阻塞 escalation 报告，识别`goframe-v2`为失败目标
- **AND** 继续后续安装目标（如`Playwright browsers`）
- **AND** 最终报告在“skipped or failed”部分包含`goframe-v2`，但不把整体环境标记为关键损坏

#### Scenario: goframe-v2 安装到用户全局而非项目本地

- **WHEN** doctor 成功运行`npx skills add github.com/gogf/skills -g`
- **THEN** skill 产物出现在`~/.claude/skills/goframe-v2/`
- **AND** 不出现在`LinaPro`仓库的`.claude/skills/goframe-v2/`
- **AND** 不污染`LinaPro` git 工作树

### Requirement: Doctor 工作流必须跨平台保持一致

系统 SHALL 在 macOS、Linux、Windows Git Bash 或 WSL 上产生等价的诊断语义。各平台安装命令必须来自`references/tool-matrix.md`中的规范映射。平台专属行为必须仅限于包管理器选择、Node 版本管理器处理和 Windows PowerShell 包装；工作流步骤、JSON schema、escalation 格式与确认策略必须在各平台保持一致。

#### Scenario: 各平台使用相同 JSON schema

- **WHEN** `doctor-check.sh`分别在三个平台上运行
- **THEN** 输出 JSON 拥有相同的顶层字段和每个工具的结构
- **AND** 只有`os`和`package_manager`字段值随平台变化

#### Scenario: 没有 Git Bash 的 Windows 用户收到明确失败信息

- **WHEN** 用户尝试从原生 PowerShell 或 CMD 调用`lina-doctor`
- **THEN** AI 工具告知用户`lina-doctor`需要在 Git Bash 或 WSL 中运行
- **AND** 不继续执行安装
