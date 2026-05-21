# agent-skills-link-cli 增量规范（命令名迁移）

## MODIFIED Requirements

### Requirement: 仓库必须提供`Agent`项目路径软链管理命令

系统 SHALL 在`linactl`中提供`agents.skills.link`和`agents.skills.unlink`命令，并在根`Makefile`中提供等价的`make agents.skills.link`和`make agents.skills.unlink`包装入口（通过`hack/makefiles/agents.mk`提供），用于在仓库根目录内统一管理`Agent`项目路径与`.agents/skills/`之间的软链。命令 SHALL 仅操作仓库内文件系统，不得修改`HOME`目录或任何系统全局路径。原`skills`/`skills.link`/`skills.unlink`命令以及`hack/makefiles/skills.mk` SHALL 被删除，且不保留为别名。

#### Scenario: 默认列出所有受支持`Agent`的当前状态

- **WHEN** 开发者运行`make agents.skills.link`或`linactl agents.skills.link`且未提供`AGENT`参数
- **THEN** 命令以表格形式输出每个受支持`Agent`的`AGENT`、`PROJECT PATH`、`CATEGORY`、`STATUS`和`DETAIL`列
- **AND** 命令不创建、不修改、不删除任何文件系统对象
- **AND** 输出末尾给出后续可执行的提示

#### Scenario: 选择性创建受支持`Agent`项目路径软链

- **WHEN** 开发者运行`make agents.skills.link AGENT=<name|all|csv>`或`linactl agents.skills.link agent=<name|all|csv>`
- **THEN** 命令对所选`Agent`按内置映射表的项目路径，在仓库根创建`<项目路径> -> .agents/skills`相对软链
- **AND** 已存在且指向`.agents/skills`的软链输出`ok`且不重新创建
- **AND** `agent=all`对所有`link`类`Agent`生效，不包含`native`类与`rootCollision`类

#### Scenario: 选择性移除受支持`Agent`项目路径软链

- **WHEN** 开发者运行`make agents.skills.unlink AGENT=<name|all|csv>`或`linactl agents.skills.unlink agent=<name|all|csv>`
- **THEN** 命令仅在目标是软链且`Readlink`命中`.agents/skills`时执行`Remove`并输出`removed`
- **AND** 目标是软链但指向其他位置时输出`skipped (foreign target)`且不删除
- **AND** 目标是普通目录或文件时输出`skipped (not a managed link)`且不删除
- **AND** 目标不存在时输出`absent`

#### Scenario: 旧`skills`命令已被删除

- **WHEN** 开发者运行`make skills`、`make skills.link`、`make skills.unlink`、`linactl skills`、`linactl skills.link`或`linactl skills.unlink`
- **THEN** 命令以错误退出（`make`层面：`No rule to make target`；`linactl`层面：`unknown command`）
- **AND** 仓库内的`hack/makefiles/skills.mk`、`command_skills.go`、`command_skills.link.go`、`command_skills.unlink.go`、`internal/skilllink/`不再存在

### Requirement: 命令在终端下必须提供交互式选择

系统 SHALL 在`agents.skills.link`和`agents.skills.unlink`未提供`agent`参数且标准输入连接到真实终端时，进入交互式选择模式；在`CI`、管道或非终端输入上下文中保持非交互行为。系统 SHALL 同时通过`linactl agents`和`make agents`聚合入口在`TTY`下展示资源 → 动作 → `Agent`三层菜单，并在选择`skills` → `link`/`unlink`时分发到`agents.skills.link` / `agents.skills.unlink`的交互式流程；非终端上下文打印用法指引而不阻塞。交互式模式 SHALL 仅依赖`Go`标准库（`os.File.Stat()`+`ModeCharDevice`），不得引入第三方交互框架。

#### Scenario: TTY 下`make agents`三层菜单进入`skills`流程

- **WHEN** 开发者在终端下运行`make agents`或`linactl agents`并依次选择`skills` → `link`
- **THEN** 命令进入`agents.skills.link`的交互式 3 列网格选择流程
- **AND** 同样的入口选择`skills` → `unlink`时进入`agents.skills.unlink`的受管软链选择流程
- **AND** 第二层选`q`或`back`时返回第一层资源选择
- **AND** 第一层选`q`、`quit`或空行时退出整个聚合菜单

#### Scenario: 非终端下`make agents`打印用法指引

- **WHEN** 开发者在`CI`或管道中运行`make agents`或`linactl agents`
- **THEN** 命令打印`agents.skills.link`/`agents.skills.unlink`等子命令及对应`make`目标的用法说明
- **AND** 命令以零退出码结束，便于`make help`等聚合入口正常工作

#### Scenario: TTY 下`agents.skills.link`默认进入交互式

- **WHEN** 开发者在终端下运行`make agents.skills.link`或`linactl agents.skills.link`且未传入`agent`参数
- **THEN** 命令首先打印一份覆盖注册表全部条目（含`native`、`link`、`rootCollision`三类）的状态总览表，使用与非交互模式相同的`Render`格式，让开发者可见所有 Agent 的当前状态
- **AND** 状态总览之后命令以 3 列网格展示`link`类 Agent 候选，每个单元格包含编号、单字符状态符号（`[+]`linked/`[~]`mismatch/`[.]`absent/`[!]`conflict/`[*]`root-collision/`[?]`error）和 Agent 名称
- **AND** 网格上方输出图例行，将状态符号映射到完整状态语义
- **AND** 网格 + 图例 + 标题 + 提示总行数控制在 24 行终端可一屏完整显示
- **AND** 命令读取一行以逗号分隔的选择
- **AND** 输入`all`选择全部候选；输入`q`、`quit`或空行取消
- **AND** 若所选 Agent 包含`mismatch`状态，命令必须再次询问是否使用`FORCE=1`重建
- **AND** `native`和`rootCollision`类 Agent 不出现在交互候选清单中

#### Scenario: TTY 下`agents.skills.unlink`仅列出受管软链

- **WHEN** 开发者在终端下运行`make agents.skills.unlink`或`linactl agents.skills.unlink`且未传入`agent`参数
- **THEN** 命令仅展示当前项目路径已是指向`.agents/skills`的软链的 Agent
- **AND** 候选清单为空时输出`No managed agent skill symlinks were found`并以零退出码结束
- **AND** 输入选择后命令复用`agents.skills.unlink`的非交互执行路径，仍然遵循"仅移除受管软链"的规则

#### Scenario: 非终端上下文保持非交互

- **WHEN** 开发者在`CI`或管道中运行`make agents.skills.link`或`linactl agents.skills.link`且未传入`agent`参数
- **THEN** 命令输出只读状态列表并退出，不提示任何输入
- **AND** `make agents.skills.unlink`或`linactl agents.skills.unlink`在未传入`agent`参数且非终端时返回非零退出码并提示需要显式`agent`参数

### Requirement: 命令必须使用`Go`标准库实现并跨平台一致

系统 SHALL 使用`os.Symlink`、`os.Readlink`、`os.Lstat`、`os.Remove`、`os.MkdirAll`、`filepath.Rel`、`filepath.Clean`等`Go`标准库实现命令逻辑，禁止调用`ln`、`mklink`、`bash`、`cmd.exe`或其他平台专属命令。`Windows`平台软链不可用时 SHALL 给出明确指引。

#### Scenario: 命令实现不依赖外部命令

- **WHEN** 评审`hack/tools/linactl/command_agents.skills.link.go`、`command_agents.skills.unlink.go`和`hack/tools/linactl/internal/agents/skills/`
- **THEN** 实现仅使用`Go`标准库和项目已有的`internal`组件
- **AND** 实现不调用`exec.Command("ln", ...)`、`exec.Command("mklink", ...)`、`bash -c`或`cmd /c`等子进程

#### Scenario: `Windows`下软链不可用时给出指引

- **WHEN** 在`Windows`平台执行`agents.skills.link`且`os.Symlink`返回权限错误
- **THEN** 命令对相关行输出`error`状态及包含"需要开发者模式或管理员"的提示
- **AND** 命令以非零退出码结束，便于`CI`断言
