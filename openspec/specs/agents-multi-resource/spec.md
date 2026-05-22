# agents-multi-resource Specification

## Purpose

定义`LinaPro`仓库内面向多资源的`Agent`项目桥接命令树（`agents.<resource>.<action>`）的开发工具能力边界，统一管理`skills`、`prompts`、`md`三类项目资源在仓库根的软链。规范覆盖`agents`聚合菜单、`agents.skills` / `agents.prompts` / `agents.md`三组子命令、聚合命令在`TTY`下的三层交互菜单、命令仅操作仓库根的作用域约束，以及`md`资源单文件软链状态机的不变量。

## Requirements
### Requirement: 仓库必须提供面向多资源的`Agent`项目桥接命令树

系统 SHALL 在`linactl`中提供以`agents`为顶层、以资源类型为第二段、以动作为第三段的命令树，统一管理`skills`、`prompts`、`md`三类项目资源在仓库根的软链。命令树 SHALL 至少包含：`agents`（聚合菜单）、`agents.skills.link`、`agents.skills.unlink`、`agents.prompts.link`、`agents.prompts.unlink`、`agents.md.link`、`agents.md.unlink`。根`Makefile`通过`hack/makefiles/agents.mk`提供同名`make`目标。命令 SHALL 仅操作仓库内文件系统，不得修改`HOME`目录或任何系统全局路径，不得引入`--root`参数。

#### Scenario: 命令树顶层覆盖三类资源

- **WHEN** 开发者运行`linactl help`或`make help`
- **THEN** 输出包含`agents`、`agents.skills.link`、`agents.skills.unlink`、`agents.prompts.link`、`agents.prompts.unlink`、`agents.md.link`、`agents.md.unlink`七条命令
- **AND** 不再出现`skills`、`skills.link`、`skills.unlink`三条旧命令

#### Scenario: 命令仅操作仓库根

- **WHEN** 开发者从仓库子目录或非仓库目录运行`linactl agents.skills.link`
- **THEN** 命令仍以仓库根为基准解析所有目标路径
- **AND** 命令不接受任何用于改变作用根的参数（如`--root`）

### Requirement: `agents.skills`子命令必须等价迁移自既有`skills`命令

系统 SHALL 在`agents.skills.link`和`agents.skills.unlink`下完整保留`agent-skills-link-cli`既有规范的所有`Scenario`行为，包括状态表渲染、`agent=<name|all|csv>`选择器、`force=1`重建、`rootCollision`默认拒绝、`mismatch`提示、`TTY`三列网格交互、`Go`标准库实现等约束。从用户视角看，`make agents.skills.link AGENT=...`与原`make skills.link AGENT=...`产生**完全相同**的状态机行为，仅命令名发生变化。

#### Scenario: `agents.skills.link`复用`skills`资源注册表

- **WHEN** 开发者运行`make agents.skills.link AGENT=all`
- **THEN** 命令对所有`link`类`Agent`创建`<项目路径> -> .agents/skills`相对软链
- **AND** `native`类`Agent`输出`native`、`rootCollision`类`Agent`输出`skipped (root collision)`，行为与原`skills.link`一致

#### Scenario: `agents.skills.unlink`严格匹配受管软链

- **WHEN** 开发者运行`make agents.skills.unlink AGENT=<name>`
- **THEN** 命令仅在目标是软链且`Readlink`命中`.agents/skills`时执行`Remove`
- **AND** 任何其他情况都不修改文件系统，行为与原`skills.unlink`一致

### Requirement: `agents.prompts`子命令必须基于显式声明的源路径管理目录级软链

系统 SHALL 在`agents.prompts.link`和`agents.prompts.unlink`下管理每个`Agent`显式声明的`SourcePath` → `ProjectPath`目录级软链，源路径不再固定为`.agents/skills`，而是由`prompts`资源的`AgentSpec`中`SourcePath`字段逐项指定。注册表 SHALL 至少覆盖`claude-code`、`cursor`、`codex`、`gemini-cli`四个`Agent`，每个`Agent`的`SourcePath`必须位于仓库内`.agents/`下且为已存在的真实目录。

#### Scenario: 每个`prompts`类`Agent`显式声明源路径

- **WHEN** 评审`internal/agents/prompts/`下的`Agent`注册表
- **THEN** 每个`AgentSpec`同时声明`SourcePath`、`ProjectPath`和`Category`
- **AND** `claude-code`条目的`SourcePath`等于`.agents/prompts/opsx`，`ProjectPath`等于`.claude/commands/opsx`

#### Scenario: 状态表渲染包含`SOURCE`列

- **WHEN** 开发者运行`make agents.prompts.link`且未提供`AGENT`参数
- **THEN** 命令输出表格包含`AGENT`、`SOURCE`、`PROJECT PATH`、`CATEGORY`、`STATUS`、`DETAIL`六列
- **AND** `SOURCE`列展示该条目的`SourcePath`字段值

#### Scenario: 软链创建使用相对源路径

- **WHEN** 开发者运行`make agents.prompts.link AGENT=claude-code`
- **THEN** 命令在仓库根创建`.claude/commands/opsx`软链
- **AND** 软链目标为`filepath.Rel`从软链所在目录到`SourcePath`推导出的相对路径

#### Scenario: 受管软链严格匹配源路径

- **WHEN** 开发者运行`make agents.prompts.unlink AGENT=<name>`
- **THEN** 命令仅在目标是软链且`Readlink`命中该`Agent`声明的`SourcePath`时执行`Remove`
- **AND** 目标软链指向其他位置时输出`skipped-foreign`且不删除

### Requirement: `agents.md`子命令必须实现单文件软链状态机

系统 SHALL 在`agents.md.link`和`agents.md.unlink`下管理`Agent`私有项目规范文件到仓库根`AGENTS.md`的**单文件**软链。状态机 SHALL 与目录级软链共用同一套`Status`枚举（`native`/`ok`/`created`/`rebuilt`/`mismatch`/`conflict`/`removed`/`skipped-foreign`/`skipped-not-managed`/`absent`/`error`），并保持以下不变量：

1. 永远不删除真实文件或目录；`force=1`仅作用于"已是软链且指向其他位置"的情况。
2. 目标存在但不是软链时输出`conflict`，与目录级状态机的"真实目录冲突"语义对齐。
3. `native`类`Agent`不建链，仅在状态表展示。
4. 注册表无`rootCollision`类条目（`AGENTS.md`本身位于仓库根，不存在对称冲突）。

#### Scenario: `link`类`Agent`创建私有文件 → `AGENTS.md`软链

- **WHEN** 开发者运行`make agents.md.link AGENT=claude-code`
- **THEN** 命令在仓库根创建`CLAUDE.md`软链
- **AND** 软链目标为相对路径`AGENTS.md`（同目录下）
- **AND** 状态输出为`created`或在已存在正确软链时输出`ok`

#### Scenario: `native`类`Agent`不建链

- **WHEN** 开发者运行`make agents.md.link AGENT=codex`
- **THEN** 命令输出`native`状态
- **AND** 命令不创建、修改、删除任何文件系统对象

#### Scenario: 目标已是真实文件时拒绝覆盖

- **WHEN** `agents.md.link`检测到目标存在且不是软链（如开发者本地手动编辑了`CLAUDE.md`）
- **THEN** 命令输出`conflict`并以非零退出码结束
- **AND** 命令在任何情况下都不删除真实文件，包括提供`force=1`时

#### Scenario: 目标是指向其他位置的软链时按`force`决定

- **WHEN** `agents.md.link`检测到目标已存在且为软链但`Readlink`未命中`AGENTS.md`
- **THEN** 默认输出`mismatch`并不修改任何内容
- **AND** 当显式提供`force=1`时，命令`Remove`旧软链后重建并输出`rebuilt`

#### Scenario: `agents.md.unlink`严格匹配受管软链

- **WHEN** 开发者运行`make agents.md.unlink AGENT=<name>`
- **THEN** 命令仅在目标是软链且`Readlink`命中`AGENTS.md`时执行`Remove`
- **AND** 目标是真实文件时输出`skipped-not-managed`且不删除

### Requirement: `agents.md`和`agents.skills`的`TTY`交互模式必须先打印含`native`类的完整状态总览

系统 SHALL 在`agents.md.link`和`agents.skills.link`的`TTY`交互模式下，**进入候选`grid`选择之前**先调用`PlanList`渲染一份覆盖注册表全部条目（含`native`、`link`、`rootCollision`三类）的状态总览表，使用与非交互模式相同的`Render`格式，让开发者明确看到`native`类`Agent`原生支持，无需手工建链。状态总览仅用于展示，不进入候选选择；候选`grid`仍只列出可建链的`link`类`Agent`。`agents.md.unlink`和`agents.skills.unlink`不受此约束，它们的`TTY`候选`grid`仅列出当前为受管软链（`StatusOK`）的`Agent`，与既有行为保持一致。

#### Scenario: `agents.md.link`交互模式先渲染含 native 的状态总览

- **WHEN** 开发者在终端下运行`make agents.md.link`且未传入`AGENT=`
- **THEN** 命令首先打印一份覆盖注册表全部条目的状态总览表，`native`类条目以`native`状态出现
- **AND** 表格之后再进入候选`grid`选择`link`类`Agent`
- **AND** `grid`本身仍只列出`link`类`Agent`，不允许选中`native`条目

#### Scenario: `agents.skills.link`交互模式先渲染含 native 的状态总览

- **WHEN** 开发者在终端下运行`make agents.skills.link`且未传入`AGENT=`
- **THEN** 命令首先打印一份覆盖注册表全部条目的状态总览表，`native`类条目（如`cursor`、`gemini-cli`、`codex`）以`native`状态出现
- **AND** 表格之后再进入候选`grid`选择`link`类`Agent`

### Requirement: 聚合命令必须在`TTY`下提供三层交互菜单

系统 SHALL 在`linactl agents`和`make agents`下提供`TTY`聚合菜单：第一层选择资源类型（`skills`/`prompts`/`md`/`quit`），第二层选择动作（`link`/`unlink`/`back`），第三层进入对应子命令的既有交互式选择流程（复用`agents.<resource>.<action>`的`TTY`分支）。在`CI`、管道或非`TTY`输入上下文中，命令 SHALL 打印用法指引而不阻塞。

#### Scenario: TTY 下三层菜单依次选择

- **WHEN** 开发者在终端下运行`make agents`并依次输入`1`（资源选`skills`）、`1`（动作选`link`）
- **THEN** 命令进入`agents.skills.link`的交互式选择流程，与直接运行`make agents.skills.link`的`TTY`体验一致
- **AND** 在第二层选`q`或`back`时返回第一层
- **AND** 在第一层选`q`、`quit`或空行时退出整个聚合菜单

#### Scenario: 非 TTY 下打印用法指引

- **WHEN** 开发者在`CI`或管道中运行`make agents`或`linactl agents`
- **THEN** 命令打印`agents.skills.link`/`agents.skills.unlink`/`agents.prompts.link`/`agents.prompts.unlink`/`agents.md.link`/`agents.md.unlink`六条子命令及对应`make`目标的用法说明
- **AND** 命令以零退出码结束，便于`make help`等聚合入口正常工作

#### Scenario: 资源类型展示与子命令对齐

- **WHEN** 开发者在终端下运行`make agents`
- **THEN** 第一层菜单选项编号、资源名（`skills`/`prompts`/`md`）和子命令名第二段保持一致
- **AND** 资源未来扩展（如新增`mcp`资源）时，仅需在聚合菜单注册表新增一项即可被自动展示

### Requirement: 实现必须按"通用能力 + 资源专属逻辑"分层组织

系统 SHALL 在`hack/tools/linactl/internal/agents/`下按以下结构组织实现，确保新增第四类资源不需要修改公共状态机或交互层：

- `internal/agents/common/`：承载与资源类型无关的能力，包括`Status`枚举、`Result`类型、表格渲染、提示行、选择器解析、目标策略、目录级与文件级共享的软链状态机辅助、交互式终端检测、提示流程、候选网格渲染、状态符号映射。
- `internal/agents/skills/`、`internal/agents/prompts/`、`internal/agents/md/`：每个子包承载该资源类型的`Agent`注册表、`Inspect`、`ApplyLink`、`ApplyUnlink`、`LinkCandidates`、`UnlinkCandidates`，并复用`common`提供的能力。

#### Scenario: 通用能力集中在`common`子包

- **WHEN** 评审`internal/agents/common/`
- **THEN** 该子包导出`Status`枚举、`Result`类型、`Render`、`EmitHints`、`HasError`、`ParseSelectors`、`SelectorAll`、`PromptSelection`、`PromptYesNo`、`IsInteractiveTerminal`、`ReadLine`、`linkMatchesSource`等公共`API`
- **AND** 该子包不导入任何资源专属注册表

#### Scenario: 每资源子包仅导入 common

- **WHEN** 评审`internal/agents/{skills,prompts,md}/`
- **THEN** 每个子包对外只依赖`internal/agents/common`，不互相依赖
- **AND** 每个子包提供本资源的`Agent`注册表与`Inspect`/`ApplyLink`/`ApplyUnlink`/`LinkCandidates`/`UnlinkCandidates`五个公共入口

#### Scenario: 表格渲染支持额外列

- **WHEN** `prompts`资源调用`common.Render`
- **THEN** 调用方传入一列`SOURCE`扩展列规格
- **AND** `skills`和`md`资源调用`common.Render`时不传扩展列，输出列与历史行为一致

### Requirement: 命令必须使用`Go`标准库实现并跨平台一致

系统 SHALL 使用`os.Symlink`、`os.Readlink`、`os.Lstat`、`os.Remove`、`os.MkdirAll`、`filepath.Rel`、`filepath.Clean`等`Go`标准库实现命令逻辑，禁止调用`ln`、`mklink`、`bash`、`cmd.exe`或其他平台专属命令。`Windows`平台软链不可用时 SHALL 给出明确指引。该约束同时适用于目录级和单文件软链。

#### Scenario: 实现不依赖外部命令

- **WHEN** 评审`hack/tools/linactl/command_agents*.go`和`hack/tools/linactl/internal/agents/`
- **THEN** 实现仅使用`Go`标准库和项目已有的`internal`组件
- **AND** 实现不调用`exec.Command("ln", ...)`、`exec.Command("mklink", ...)`、`bash -c`或`cmd /c`等子进程

#### Scenario: `Windows`下软链不可用时给出指引

- **WHEN** 在`Windows`平台执行`agents.<resource>.link`且`os.Symlink`返回权限错误
- **THEN** 命令对相关行输出`error`状态及包含"需要开发者模式或管理员"的提示
- **AND** 命令以非零退出码结束，便于`CI`断言

#### Scenario: 单文件软链跨平台行为一致

- **WHEN** 在 Linux、macOS、Windows 上执行`agents.md.link AGENT=claude-code`
- **THEN** 三个平台均创建`CLAUDE.md`软链，目标为相对路径`AGENTS.md`
- **AND** `Windows`平台由`os.Symlink`根据目标已存在的真实文件类型自动选择文件软链
