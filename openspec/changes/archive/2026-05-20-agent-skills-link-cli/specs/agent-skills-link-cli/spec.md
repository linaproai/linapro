## ADDED Requirements

### Requirement: 仓库必须提供`Agent`项目路径软链管理命令

系统 SHALL 在`linactl`中提供`skills.link`和`skills.unlink`命令，并在根`Makefile`中提供等价的`make skills.link`和`make skills.unlink`包装入口，用于在仓库根目录内统一管理`Agent`项目路径与`.agents/skills/`之间的软链。命令 SHALL 仅操作仓库内文件系统，不得修改`HOME`目录或任何系统全局路径。

#### Scenario: 默认列出所有受支持`Agent`的当前状态

- **WHEN** 开发者运行`make skills.link`或`linactl skills.link`且未提供`AGENT`参数
- **THEN** 命令以表格形式输出每个受支持`Agent`的`AGENT`、`PROJECT PATH`、`CATEGORY`、`STATUS`和`DETAIL`列
- **AND** 命令不创建、不修改、不删除任何文件系统对象
- **AND** 输出末尾给出后续可执行的提示

#### Scenario: 选择性创建受支持`Agent`项目路径软链

- **WHEN** 开发者运行`make skills.link AGENT=<name|all|csv>`或`linactl skills.link agent=<name|all|csv>`
- **THEN** 命令对所选`Agent`按内置映射表的项目路径，在仓库根创建`<项目路径> -> .agents/skills`相对软链
- **AND** 已存在且指向`.agents/skills`的软链输出`ok`且不重新创建
- **AND** `agent=all`对所有`link`类`Agent`生效，不包含`native`类与`rootCollision`类

#### Scenario: 选择性移除受支持`Agent`项目路径软链

- **WHEN** 开发者运行`make skills.unlink AGENT=<name|all|csv>`或`linactl skills.unlink agent=<name|all|csv>`
- **THEN** 命令仅在目标是软链且`Readlink`命中`.agents/skills`时执行`Remove`并输出`removed`
- **AND** 目标是软链但指向其他位置时输出`skipped (foreign target)`且不删除
- **AND** 目标是普通目录或文件时输出`skipped (not a managed link)`且不删除
- **AND** 目标不存在时输出`absent`

### Requirement: `Agent`映射表必须与`vercel-labs/skills`官方项目路径对齐

系统 SHALL 在`hack/tools/linactl/internal/skilllink/`下集中维护一份`AgentSpec`映射表，与`vercel-labs/skills`官方支持的`Agent`项目路径列表保持一致，并显式区分`native`、`link`、`rootCollision`三类。任何新增、修改、移除条目 SHALL 通过本地映射表与单元测试体现，禁止从`docs`目录或外部网络运行时解析。

#### Scenario: `native`类`Agent`不需要项目软链

- **WHEN** 开发者对`amp`、`kimi-cli`、`replit`、`universal`、`antigravity`、`cline`、`dexto`、`warp`、`codex`、`cursor`、`deepagents`、`firebender`、`gemini-cli`、`github-copilot`、`opencode`等`native`类`Agent`运行`skills.link`
- **THEN** 命令输出`native`状态
- **AND** 命令不创建任何软链或目录

#### Scenario: `link`类`Agent`创建相对软链到`.agents/skills`

- **WHEN** 开发者对`link`类`Agent`（如`claude-code`、`codebuddy`、`windsurf`、`qoder`、`roo`、`goose`、`devin`、`droid`、`forgecode`、`tabnine-cli`等）运行`skills.link`
- **THEN** 命令在仓库根创建`<项目路径> -> .agents/skills`相对软链
- **AND** 软链路径使用`filepath.Rel`从软链所在目录推导，确保仓库整体复制或重命名后仍然有效

#### Scenario: `rootCollision`类`Agent`默认拒绝在仓库根创建`skills/`

- **WHEN** 开发者对`openclaw`等`rootCollision`类`Agent`运行`skills.link`且未提供`force=1`
- **THEN** 命令输出`skipped (root collision)`并保留`hint`说明需要显式`force=1`
- **AND** 仅当显式提供`force=1`且仓库根的`skills/`不存在或已是指向`.agents/skills`的软链时，才创建对应软链

### Requirement: 软链冲突必须可预测且不破坏既有真实目录

系统 SHALL 在执行`skills.link`和`skills.unlink`时使用`os.Lstat`/`os.Readlink`显式判定目标类型，对不同状态分别处理，且禁止在任何情况下递归删除真实目录或文件。`force=1` SHALL 仅作用于“已是软链但指向非`.agents/skills`”的情况。

#### Scenario: 已存在指向其他位置的软链

- **WHEN** `skills.link`检测到目标已存在且为软链但`Readlink`未命中`.agents/skills`
- **THEN** 默认输出`mismatch`并不修改任何内容
- **AND** 当显式提供`force=1`时，命令`Remove`旧软链后重建并输出`rebuilt`
- **AND** 命令仅删除软链本身，不递归访问软链指向的真实目录

#### Scenario: 目标已是普通目录或文件

- **WHEN** `skills.link`检测到目标已存在且不是软链
- **THEN** 命令输出`conflict`并以非零退出码结束（当且仅当本次执行存在`error`类状态）
- **AND** 命令在任何情况下都不删除真实目录或文件，包括提供`force=1`时

#### Scenario: `unlink`严格匹配受管软链

- **WHEN** `skills.unlink`检测到目标
- **THEN** 仅当目标是软链且`Readlink`命中`.agents/skills`时执行`Remove`
- **AND** 任何其他情况都不修改文件系统

### Requirement: 命令在终端下必须提供交互式选择

系统 SHALL 在`skills.link`和`skills.unlink`未提供`agent`参数且标准输入连接到真实终端时，进入交互式选择模式；在`CI`、管道或非终端输入上下文中保持非交互行为。系统 SHALL 同时提供`linactl skills`和`make skills`聚合入口：终端下展示`link`/`unlink`/`quit`操作菜单并分发到对应子命令的交互式流程；非终端上下文打印用法指引而不阻塞。交互式模式 SHALL 仅依赖`Go`标准库（`os.File.Stat()`+`ModeCharDevice`），不得引入第三方交互框架。

#### Scenario: TTY 下`make skills`展示操作菜单

- **WHEN** 开发者在终端下运行`make skills`或`linactl skills`
- **THEN** 命令展示编号操作菜单：`[1] link`、`[2] unlink`、`[q] quit`
- **AND** 输入`1`或`link`进入`skills.link`交互式选择流程
- **AND** 输入`2`或`unlink`进入`skills.unlink`交互式选择流程
- **AND** 输入`q`、`quit`或空行取消
- **AND** 输入其他值返回错误并退出

#### Scenario: 非终端下`make skills`打印用法指引

- **WHEN** 开发者在`CI`或管道中运行`make skills`或`linactl skills`
- **THEN** 命令打印`linactl skills.link`/`skills.unlink`和对应`make`目标的用法说明
- **AND** 命令以零退出码结束，便于`make help`等聚合入口正常工作

#### Scenario: TTY 下默认进入交互式

- **WHEN** 开发者在终端下运行`make skills.link`或`linactl skills.link`且未传入`agent`参数
- **THEN** 命令以 3 列网格展示`link`类 Agent 候选，每个单元格包含编号、单字符状态符号（`[+]`linked/`[~]`mismatch/`[.]`absent/`[!]`conflict/`[*]`root-collision/`[?]`error）和 Agent 名称
- **AND** 网格上方输出图例行，将状态符号映射到完整状态语义
- **AND** 网格 + 图例 + 标题 + 提示总行数控制在 24 行终端可一屏完整显示
- **AND** 命令读取一行以逗号分隔的选择
- **AND** 输入`all`选择全部候选；输入`q`、`quit`或空行取消
- **AND** 若所选 Agent 包含`mismatch`状态，命令必须再次询问是否使用`FORCE=1`重建
- **AND** `native`和`rootCollision`类 Agent 不出现在交互候选清单中

#### Scenario: TTY 下`skills.unlink`仅列出受管软链

- **WHEN** 开发者在终端下运行`make skills.unlink`或`linactl skills.unlink`且未传入`agent`参数
- **THEN** 命令仅展示当前项目路径已是指向`.agents/skills`的软链的 Agent
- **AND** 候选清单为空时输出`No managed agent skill symlinks were found`并以零退出码结束
- **AND** 输入选择后命令复用`skills.unlink`的非交互执行路径，仍然遵循“仅移除受管软链”的规则

#### Scenario: 非终端上下文保持非交互

- **WHEN** 开发者在`CI`或管道中运行`make skills.link`或`linactl skills.link`且未传入`agent`参数
- **THEN** 命令输出只读状态列表并退出，不提示任何输入
- **AND** `make skills.unlink`或`linactl skills.unlink`在未传入`agent`参数且非终端时返回非零退出码并提示需要显式`agent`参数

### Requirement: 命令必须使用`Go`标准库实现并跨平台一致

系统 SHALL 使用`os.Symlink`、`os.Readlink`、`os.Lstat`、`os.Remove`、`os.MkdirAll`、`filepath.Rel`、`filepath.Clean`等`Go`标准库实现命令逻辑，禁止调用`ln`、`mklink`、`bash`、`cmd.exe`或其他平台专属命令。`Windows`平台软链不可用时 SHALL 给出明确指引。

#### Scenario: 命令实现不依赖外部命令

- **WHEN** 评审`hack/tools/linactl/command_skills.link.go`、`command_skills.unlink.go`和`hack/tools/linactl/internal/skilllink/`
- **THEN** 实现仅使用`Go`标准库和项目已有的`internal`组件
- **AND** 实现不调用`exec.Command("ln", ...)`、`exec.Command("mklink", ...)`、`bash -c`或`cmd /c`等子进程

#### Scenario: `Windows`下软链不可用时给出指引

- **WHEN** 在`Windows`平台执行`skills.link`且`os.Symlink`返回权限错误
- **THEN** 命令对相关行输出`error`状态及包含“需要开发者模式或管理员”的提示
- **AND** 命令以非零退出码结束，便于`CI`断言
