## Context

`LinaPro`仓库通过`.agents/skills/`统一维护`Agent Skills`内容，并在`docs/quick/4000-agent-tools.md`中梳理了主流`AI Coding`工具的技能目录与项目规范读取路径。多个工具（如`Claude Code`、`CodeBuddy`、`Windsurf`、`Qoder`、`Roo Code`、`Goose`等）默认只从各自专属项目路径加载技能，需要在仓库根创建`.<tool>/skills -> .agents/skills`项目级软链。仓库当前仅手工维护了`.claude/skills -> ../.agents/skills`和`.claude/commands -> ../.agents/prompts`两条；其余工具均未建立软链，且没有统一的命令查询和重建状态。

`vercel-labs/skills`官方文档给出了完整的`Agent`项目路径与全局路径表。本次需要在`linactl`内置该表的**项目路径子集**，按`make`/`linactl`命令统一管理仓库内软链，避免开发者手写`ln -s`，并保证跨平台一致性。

## Goals / Non-Goals

**Goals:**

- 提供`make skills.link`/`make skills.unlink`和`linactl skills.link`/`linactl skills.unlink`两套等价入口，统一管理仓库内`Agent`项目路径软链。
- 仅在仓库根目录范围内操作；不修改`HOME`目录或任何系统全局路径。
- 命令默认非交互；通过`AGENT=<name|all|csv>`参数选择目标，`FORCE=1`覆盖指向错误源的旧软链。
- 内置`Agent`映射表与`vercel-labs/skills`官方项目路径表对齐，并显式区分`native`/`link`/`root-collision`三种类别。
- 软链冲突处理可预测、可逆、不破坏既有真实目录或文件。
- 使用`Go`标准库实现，跨`Windows`/`Linux`/`macOS`一致行为；`Windows`下软链不可用时给出明确指引。

**Non-Goals:**

- 不实现交互式选择`UI`、不引入第三方交互库（如`survey`/`promptui`）。
- 不创建`AGENTS.md → CLAUDE.md / GEMINI.md / QWEN.md / QOQDER.md`等文件级软链（项目已经通过版本化软链或独立文件维护）。
- 不管理`~/.claude/skills`、`~/.codebuddy/skills`等`HOME`目录的全局路径。
- 不替代`skills-lock.json`从远端`gogf/skills`等仓库拉取`Skill`源码的能力，也不下载`vercel-labs/skills`仓库本身。
- 不变更`.agents/skills/`下技能内容或目录结构。

## Decisions

### 1. 命令入口与参数

新增两个`linactl`命令：

- `linactl skills.link [agent=<name|all|csv>] [force=1]`：
  - 终端上下文且未传入`agent`参数时，进入交互式选择模式：列出`link`类候选、读取以逗号分隔的编号选择，支持`all`/`q`，并在选中的 Agent 包含`mismatch`时询问是否启用`FORCE=1`重建。
  - 非终端上下文（CI、管道）且未传入`agent`参数时，回退到只读状态列表。
  - 显式`agent=...`时直接非交互执行；`force=1`仅对“已存在但指向非`.agents/skills`”的软链允许覆盖重建。
- `linactl skills.unlink [agent=<name|all|csv>]`：
  - 终端上下文且未传入`agent`参数时，进入交互式选择模式，仅列出当前已是受管软链的 Agent；候选清单为空时输出提示并以零退出码结束。
  - 非终端上下文且未传入`agent`参数时返回非零退出码并提示必须显式传入。
  - 仅在目标是软链且`Readlink`命中`.agents/skills`时才`Remove`，绝不递归删除真实目录或文件。

`Makefile`包装：

- `make skills.link [AGENT=...] [FORCE=1]`
- `make skills.unlink [AGENT=...]`

包装层只做参数转发，不内置默认`AGENT`。`make`包装层执行后`stdin`仍然连接到调用者终端，因此终端下`make skills.link`等价于直接在终端运行`linactl skills.link`，进入交互式选择。

交互式实现仅依赖`Go`标准库：通过`os.File.Stat()`返回值的`ModeCharDevice`位检测真实终端，使用`bufio.Reader.ReadString('\n')`读取一行选择，通过编号清单 + 逗号分隔输入实现多选。不引入`survey`/`promptui`等第三方交互库，符合开发工具最简跨平台原则。

候选清单使用 3 列网格 + 单字符状态符号布局：39 个`link`类 Agent 经过列优先排列后只占 13 行，加上标题、图例和提示行总计 17 行，能够在 24 行终端中一屏完整显示。状态符号定义为`[+]`linked、`[~]`mismatch（含`skipped-foreign`/`skipped-not-managed`）、`[.]`absent（含`native`占位）、`[!]`conflict、`[*]`root-collision、`[?]`error，全部使用 ASCII 字符避免`Unicode`/`emoji`在不同终端下的字符宽度问题。完整状态文字与`ProjectPath`仍可通过非交互列表（`make skills.link`无参数 + 非 TTY）查询，避免在窄屏下展开列宽。

备选方案是单命令`make skills.sync`一次性同步所有`link`类`Agent`，但这会让`unlink`和`force-rebuild`语义混在一起，不利于调试和审查。

### 2. `Agent`映射表内置

`hack/tools/linactl/internal/skilllink/agents.go`内置一份按`vercel-labs/skills`官方项目路径整理的映射表，每条记录：

```go
type AgentSpec struct {
    Name        string // CLI 参数名，例如 claude-code
    DisplayName string // 输出表中显示用，例如 Claude Code
    ProjectPath string // 项目路径，例如 .claude/skills
    Category    Category // native / link / rootCollision
}
```

`Category`：

- `native`：`ProjectPath == .agents/skills`，无需软链。`amp`、`kimi-cli`、`replit`、`universal`、`antigravity`、`cline`、`dexto`、`warp`、`codex`、`cursor`、`deepagents`、`firebender`、`gemini-cli`、`github-copilot`、`opencode`。
- `link`：`ProjectPath`不同于`.agents/skills`且不是仓库根。占绝大多数`Agent`，例如`claude-code → .claude/skills`、`codebuddy → .codebuddy/skills`、`qoder → .qoder/skills`等。
- `rootCollision`：`ProjectPath == skills`（仅`openclaw`）。默认拒绝创建，需要`AGENT=openclaw FORCE=1`并且`skills`不存在或已是指向`.agents/skills`的软链才允许。

`trae`和`trae-cn`项目路径相同（`.trae/skills`），合并为同一目标，但允许通过`AGENT=trae,trae-cn`独立寻址（去重后只创建一次）。

备选方案是从`docs/quick/4000-agent-tools.md`运行时解析；但`docs`格式只是说明性表格，不是稳定数据源，且前述文档与`vercel-labs/skills`存在条目差异，运行时解析会增加耦合面。

### 3. 软链路径与方向

固定为`<项目路径> -> ../.agents/skills`相对软链。例如：

- `.claude/skills -> ../.agents/skills`
- `.codebuddy/skills -> ../.agents/skills`
- `.tabnine/agent/skills -> ../../.agents/skills`（按深度生成相对路径）

实现使用`filepath.Rel`从软链所在目录推导出相对源路径，确保仓库整体复制或重命名后软链仍然有效。

### 4. 创建/检测/移除规则

```
状态                    动作
未创建                   os.MkdirAll 父目录 + os.Symlink，输出 created
软链 → .agents/skills    输出 ok
软链 → 其他              输出 mismatch；FORCE=1 时 os.Remove + 重建，输出 rebuilt
普通目录或文件           输出 conflict；任何情况下都不删除，要求人工处理
父目录创建失败           输出 error 并返回错误
Windows os.Symlink 失败  输出 error，提示开启开发者模式或以管理员运行
```

`unlink`：

- 软链 → `.agents/skills`：`os.Remove`，输出 removed
- 软链 → 其他：输出 skipped (foreign target)，不删除
- 普通目录或文件：输出 skipped (not a managed link)，不删除
- 不存在：输出 absent

`Readlink`返回值与目标路径比对前，统一使用`filepath.Clean`+`filepath.IsAbs`判断后转换为相对仓库根的标准化形式，避免`./`、`../`差异导致误判。

### 5. 输出格式

固定列表化输出，列：`AGENT` `PROJECT PATH` `CATEGORY` `STATUS` `DETAIL`。`status`枚举：

- `link`命令：`native` / `created` / `ok` / `mismatch` / `rebuilt` / `conflict` / `skipped` / `error`
- `unlink`命令：`native` / `removed` / `skipped` / `absent` / `error`

输出按`AGENT`字典序排序，最后追加 hint：

- 存在`mismatch`时提示`run with FORCE=1 to rebuild mismatched links`
- 存在`conflict`时提示需要人工处理具体路径
- 存在`Windows`权限错误时提示开发者模式或管理员

`error`分类必须导致命令以非零退出码结束，便于`CI`夹具断言。

### 6. 与`.gitignore`协同

`.<tool>/skills`是开发者本地工具适配产物，不需要进入版本库。命令首次执行成功后，由开发者自行决定是否提交`AGENT`目录；同时本次变更在`.gitignore`中添加常见`AGENT`目录的`/skills`忽略规则（仅忽略软链本身，不影响`.claude/commands`等已有真实目录）。具体规则在实现阶段按内置映射表生成。

### 7. 跨平台与`linactl`组织规范

- 命令文件：`hack/tools/linactl/command_skills.link.go`、`command_skills.unlink.go`，按`linactl`命令文件命名规范使用点分段后缀。
- 子组件：`hack/tools/linactl/internal/skilllink/`，仅包含`Go`实现与单元测试，不依赖任何外部工具或平台命令。
- 完全使用`os.Symlink`/`os.Readlink`/`os.Lstat`/`os.Remove`/`os.MkdirAll`/`filepath.Rel`/`filepath.Clean`，不调用`ln`/`mklink`/`cmd.exe`/`bash`。
- `Windows`下捕获`os.Symlink`错误，输出包含“需要开发者模式或管理员”的明确提示文案，避免静默失败。

## Risks / Trade-offs

- `vercel-labs/skills`未来可能新增`Agent` → 通过将映射表集中放入`internal/skilllink/agents.go`并补单元测试，确保新增条目在评审中可见。
- `Agent`项目路径变更（例如`droid → .factory/skills`这类与工具名不同的路径） → 在映射表中显式记录`DisplayName`和`ProjectPath`字段，避免代码硬编码工具名拼接路径。
- 普通目录冲突 → 永不自动删除，仅给出 conflict 提示和处理建议。
- `Windows`权限受限 → 命令明确报错并指引开发者模式/管理员，不退化为平台专属脚本。
- 用户误以为命令会同步`HOME`目录 → 在`README`和命令`Description`中明确仅管理仓库内项目路径软链。
- 旧手工创建的`.claude/skills -> ../.agents/skills`软链 → 命令检测到等价目标后输出`ok`，不重新创建，保持幂等。

## Migration Plan

1. 实现`internal/skilllink`子组件：`AgentSpec`、`Category`、内置映射表、`Plan/Apply/Unlink`核心函数与单元测试。
2. 实现`linactl skills.link`/`linactl skills.unlink`命令文件，调用子组件并按列表化输出渲染状态。
3. 在`command.go`注册两条命令；按命令文件命名规范放置`command_skills.link.go`和`command_skills.unlink.go`。
4. 新增`hack/makefiles/skills.mk`，并在根`Makefile`中`include`。
5. 在`.gitignore`中按内置`Agent`映射表追加常见`AGENT`目录`/skills`忽略规则。
6. 更新`hack/tools/linactl/README.md`、`README.zh-CN.md`，描述命令、参数与示例。
7. 运行`cd hack/tools/linactl && go test ./... -count=1`、`cd hack/tools/linactl && go run . skills.link`、`cd hack/tools/linactl && go run . skills.link agent=claude-code`、`cd hack/tools/linactl && go run . skills.unlink agent=claude-code`等命令烟测。
8. 运行`openspec validate agent-skills-link-cli --strict`（如`openspec`工具可用）或等价静态校验，并执行`/lina-review`。

## Open Questions

无。
