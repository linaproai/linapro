## Context

`hack/tools/linactl/internal/skilllink`目前以**单一职责**形式存在：一个`AgentSpec`只承载`skills`资源的目标路径，源路径由`SourceDir`常量固定为`.agents/skills`，整个状态机只处理"目录到目录"的相对软链。这一设计在`agent-skills-link-cli`变更落地时刻意选择了"`MVP`不做多资源抽象"，注释写得很清楚：

```
hack/tools/linactl/internal/skilllink/skilllink_agents.go:13
const SourceDir = ".agents/skills"
```

随着仓库引入了第二类需要桥接的资源（`.agents/prompts/opsx/` 已经手工挂在 `.claude/commands/opsx/`）和第三类资源（`AGENTS.md` 通过 `CLAUDE.md` 软链供 `Claude Code` 读取），继续在`skilllink`中横向扩展会带来三个具体问题：

1. **数据模型撑不下**：`AgentSpec`没有源路径字段，加进去就会污染所有`skills`类条目；如果按资源类型加多个字段集合，整个注册表会变成稀疏矩阵。
2. **状态机不够通用**：`linkMatchesSource`、`createLink`、`applyOneUnlink`隐含了"目录到目录"假设，对`AGENTS.md → CLAUDE.md`这种**单文件软链**需要不同的目标存在性判定（`Lstat`仍可用，但`MkdirAll`变成`MkdirAll(filepath.Dir(target), ...)`且不需要为空目录建链）。
3. **命令与`Make`目标命名空间会爆炸**：每加一类资源就多出三个`linactl`命令、三个`make`目标、三段`README`，`commandRegistry()`迅速变得臃肿，维护代价非线性增长。

`agent-skills-link-cli`已建立的`native`/`link`/`rootCollision`三态分类、`Inspect`/`PlanList`/`ApplyLink`/`ApplyUnlink`四个公开入口、`Status`枚举集合、`PromptSelection`/`PromptYesNo`交互式流程都是**与资源类型无关**的工程经验，理应被多资源场景复用。本设计的核心任务是把"通用能力"和"资源专属逻辑"清晰地切开，并在新的`agents`命令树下统一暴露。

## Goals / Non-Goals

**Goals:**

1. 把`skilllink`重构为面向多资源的`agents`框架，使新增第四类资源（如`mcp.json`、`apidoc i18n`等）只需新增一个`internal/agents/<resource>/`子包并在`commandRegistry()`补三个命令，不需要改动公共状态机或交互层。
2. 在`linactl`命令面建立`agents.<resource>.{link,unlink}`命令树，资源类型作为命令名的一部分（而非参数），保留`make help`/`linactl help`一眼可见每条命令意图的可发现性。
3. 提供`linactl agents`聚合菜单，`TTY`下三层交互（资源 → 动作 → `Agent`）一次性完成跨资源的桥接维护；非`TTY`下打印用法指引而不阻塞`make help`。
4. 全量迁移`skills`类既有行为，确保`agents.skills.link`与原`skills.link`在所有现有`Scenario`下行为完全一致（包括`force=1`重建、`rootCollision`默认拒绝、`mismatch`提示、`TTY`三列网格交互等）。
5. 删除旧`skills`命令树（破坏性变更），在`README`迁移指引中明确说明对应的新命令名，让仓库内不残留旧引用。
6. 单文件软链状态机在语义上与目录级状态机保持一致（同一套`Status`枚举、同样的"`force`只作用于已是软链且指向其他位置"约束、同样不删除真实文件）。

**Non-Goals:**

1. 不引入"跨资源批量"快捷命令（如`agents.link --all-resources`）；多资源协同通过聚合菜单完成。
2. 不引入第三方交互式`TUI`框架；交互层继续仅依赖`Go`标准库。
3. 不改变`.agents/skills/`、`.agents/prompts/`、`AGENTS.md`本身的内容、目录结构或加载顺序。
4. 不下载或同步外部内容；不替代`skills-lock.json`从远端拉取`Skill`包的现有流程。
5. 不支持`--root=<path>`参数，工具仅在仓库根目录运行；嵌套子仓库（如`linapro-site`）由用户在该目录内独立运行同名工具。
6. 不为`AGENTS.md`资源实现"自动写入文件内容"等非软链式桥接（仅维护单文件软链）。

## Decisions

### D1：包结构采用"通用子包 + 每资源子包"

**决定**：

```
hack/tools/linactl/internal/agents/
├── common/        # 与资源类型无关的能力
├── skills/        # skills 资源的注册表 + 资源专属逻辑（极薄）
├── prompts/       # prompts 资源的注册表 + 资源专属逻辑（极薄）
└── md/            # md 资源的注册表 + 单文件软链状态机
```

`common`抽取的能力包括：

- `Status`枚举集合（`StatusNative`/`StatusOK`/`StatusCreated`/`StatusRebuilt`/`StatusMismatch`/`StatusConflict`/`StatusSkippedRootCollision`/`StatusRemoved`/`StatusSkippedForeignTarget`/`StatusSkippedNotManaged`/`StatusAbsent`/`StatusError`）。
- `Result`类型与`HasError`/`IsError`辅助。
- 表格渲染`Render`和提示行`EmitHints`（**注意**：列名仍叫`AGENT`/`PROJECT PATH`/`CATEGORY`/`STATUS`/`DETAIL`，但`SOURCE`列在`prompts`资源下需要额外展示，故`Render`接受可选列扩展回调）。
- 选择器解析`ParseSelectors`/`SelectorAll`/`hasAll`、`targetPolicy`/`resolveTargets`（泛化为按`Category`策略选过滤）。
- 交互式：`IsInteractiveTerminal`/`ReadLine`/`PromptSelection`/`PromptYesNo`/`renderCandidateGrid`/`statusGlyph`。
- 通用文件系统辅助：`linkMatchesSource(repoRoot, link, sourceAbs)`/`pathsEqual`/`equalFoldPath`/`symlinkErrorDetail`。

每资源子包只承载：

- 该资源的`AgentSpec`注册表（`agents []AgentSpec`，`init()`排序+路径规范化）。
- 资源类型常量（`ResourceKind = "skills" | "prompts" | "md"`）。
- 资源专属的`Inspect`/`ApplyLink`/`ApplyUnlink`包装（内部委托给`common`，仅在源路径解析或文件级 vs. 目录级状态机选择上有差异）。
- 资源专属的`LinkCandidates`/`UnlinkCandidates`（同样委托`common`）。

**为什么不直接做一个统一`AgentSpec{Resources map[ResourceKind]Binding}`？**

考虑过的备选方案：

```go
type AgentSpec struct {
    Name      string
    Resources map[ResourceKind]ResourceBinding
}
```

驳回原因：

1. 注册表会变成稀疏矩阵——`amp`只有`skills`，`claude-code`三类都有，`junie`可能只有`md`，遍历时大量空判分支。
2. 每资源子命令仍要从同一份大注册表里筛选，不如直接让每资源拥有自己的注册表语义清晰。
3. 跨资源协同场景（聚合菜单）需要的不是"`Agent`视角的全部资源"，而是"按资源类型逐项展示状态"，三个独立注册表反而更贴合交互。

### D2：单文件软链状态机与目录级状态机共用同一套`Status`

**决定**：`md`资源的状态机直接复用`common`提供的目录级状态机骨架，仅在以下三个点上做单文件特化：

1. `Lstat`目标后判定`info.Mode()`时，**目录级**要求"非 symlink 即真实目录"，**文件级**要求"非 symlink 即真实文件"；两者都返回`StatusConflict`但`Detail`文案不同。
2. `createLink`：目录级在`createLink`中`MkdirAll(filepath.Dir(target), 0o755)`；文件级也是同一行（创建父目录），两者实现一致，差异只在`os.Symlink`的目标值。
3. `linkMatchesSource`：函数签名扩展为接受`sourceAbs string`参数（而不是从`SourceDir`常量推断），让目录级和文件级共用比较逻辑。

这样`md`资源不需要单独实现一套状态机，只需要在它的`Inspect`/`ApplyLink`/`ApplyUnlink`里把"`source`是文件还是目录"作为参数传给`common`的统一函数。

### D3：命令树形态与命名

**决定**：

| 命令 | 用途 |
|---|---|
| `linactl agents` | 聚合菜单（TTY 三层交互；非 TTY 打印 usage） |
| `linactl agents.skills.link` / `linactl agents.skills.unlink` | `skills`资源 |
| `linactl agents.prompts.link` / `linactl agents.prompts.unlink` | `prompts`资源 |
| `linactl agents.md.link` / `linactl agents.md.unlink` | `md`资源 |

`make`目标按 1:1 映射，参数仍走`AGENT=<name|all|csv>`和`FORCE=1`。聚合命令本身没有`agent`/`force`参数（聚合菜单内部分发）。

**驳回的备选**：

- `agents.link resource=skills|prompts|md|all`（参数化）：用户决定不要，原因是命令树更直观，`make help`一行就能看出意图。
- `agents.link.skills`：把`link`放在中间会让"动作前缀"分散到尾部，与现有`make`目标命名风格（`test.host`/`test.plugins`/`plugins.install`等）的"主词在前"模式冲突。

### D4：聚合菜单的三层交互流程

**决定**：

```
$ make agents
What resource do you want to manage?
  [1] skills    Agent skills directory bridge
  [2] prompts   Agent commands/prompts directory bridge
  [3] md        AGENTS.md project guide file bridge
  [q] quit
> 1

What do you want to do?
  [1] link    Create symlinks for the chosen resource
  [2] unlink  Remove managed symlinks for the chosen resource
  [q] back
> 1

# 进入 agents.skills.link 的交互流程（与现有 skills.link 完全一致的 3 列网格）
```

第二层选`[q] back`返回第一层；第一层选`[q] quit`退出。第三层（`Agent`选择）复用每资源子包的`LinkCandidates`/`UnlinkCandidates`和`common.PromptSelection`。

**驳回的备选**：

- 扁平 6 项菜单（直接列 `skills.link / skills.unlink / prompts.link / ...`）：用户选了三层菜单，扁平方式在资源类型扩展到 4-5 类时会让一屏装不下。
- "默认全部资源走一遍"：会让`force=1`等危险操作扩散到所有资源，不安全。

### D5：注册表初始内容

`skills`：1:1 迁移，零行为差异。

`prompts`：仅纳入有可验证证据的 4 个`Agent`，每个显式声明`SourcePath`：

| Agent | SourcePath | ProjectPath | Category |
|---|---|---|---|
| `claude-code` | `.agents/prompts/opsx` | `.claude/commands/opsx` | `link` |
| `cursor` | `.agents/prompts/opsx` | `.cursor/commands/opsx`（社区约定）| `link` |
| `codex` | `.agents/prompts/opsx` | `.codex/prompts/opsx` | `link` |
| `gemini-cli` | `.agents/prompts/opsx` | `.gemini/commands/opsx` | `link` |

未来扩展时新增条目即可，不影响其他资源。

`md`：覆盖全套已知`Agent`，分两态：

- `link`类（私有文件 → `AGENTS.md`）：`claude-code`→`CLAUDE.md`、`gemini-cli`→`GEMINI.md`、`qwen-code`→`QWEN.md`、`junie`→`.junie/guidelines.md`、`windsurf`→`.windsurfrules`、`augment`→`.augment-guidelines`、`continue`→`.continuerules`等约 8-10 项。
- `native`类（在状态表中标 native，不建链）：`codex`、`cursor`、`amp`、`opencode`、`cline`、`warp`、`replit`、`antigravity`、`deepagents`、`dexto`、`firebender`、`github-copilot`、`kimi-cli`、`universal`等约 14 项。
- `rootCollision`类：`md`资源**没有**`rootCollision`场景（`AGENTS.md`本身就是仓库根文件，不存在与第三方仓库根冲突的对称问题），注册表无此类别条目。

### D6：破坏性变更与文档迁移指引

旧命令`skills*`和`hack/makefiles/skills.mk`一并删除，**不**保留为别名。这要求：

1. `hack/tools/linactl/README.md` 与 `README.zh-CN.md`新增"迁移指引"段落，列对照表（`make skills.link` → `make agents.skills.link`等）。
2. `linapro-site`下的文档同步更新（仅声明影响面，实际更新通过`linapro-site`独立仓提交，因为该仓已不再被主仓索引追踪）。
3. `/lina-review`审查时通过静态扫描确认仓库内不再出现`skills.link`/`skills.unlink`/`make skills`字符串引用（除"迁移指引"段落和归档变更目录内的历史文档外）。
4. `openspec/specs/agent-skills-link-cli/spec.md`通过`MODIFIED Requirements`增量同步命令名变更，确保规范基线与实现保持一致。

### D7：`Render`列扩展机制

`prompts`资源比`skills`/`md`多一个有意义的`SOURCE`列（每条目源路径不同）。设计`common.Render`接受可选的`extraColumns`参数：

```go
type ColumnSpec struct {
    Header string
    Value  func(Result) string
}
func Render(out io.Writer, results []Result, extraColumns ...ColumnSpec) error
```

`skills`和`md`资源传零额外列；`prompts`资源传一列`SOURCE`。这样表格渲染器单一实现，无需为每个资源各写一份。

### D8：`Status`枚举的归属

`Status`从`skilllink`迁入`common`时**保持类型名不变**（仍叫`Status`），不引入`ResourceStatus`/`AgentStatus`这类新名字，避免不必要的`API`噪音。所有子包`import`同一个`common.Status`类型并复用其常量。

## Risks / Trade-offs

- **风险：破坏性删除旧命令导致开发者本地工作流报错** → 缓解：在`README`迁移指引中提供命令对照表；`/lina-review`审查时全仓扫描旧命令字符串；本次变更涉及命令面变更，开发者执行`git pull`后第一次运行`make skills.link`会得到`make: *** No rule to make target 'skills.link'`，错误明确，不会沉默失败。
- **风险：`prompts`资源源路径需要每个`Agent`显式声明，注册表维护成本高** → 缓解：初始只纳入 4 个有充分证据的`Agent`；新增条目要求在`PR`描述中提供`Agent`官方文档链接证明该路径正确。
- **风险：单文件软链在`Windows`上的行为与目录级软链不一致**（`Windows`区分文件软链与目录软链） → 缓解：`os.Symlink`在`Windows`上根据目标是否存在自动选择文件/目录软链类型；`AGENTS.md`链接创建时目标必然是已存在的真实文件，类型推断正确；测试覆盖`Windows runner`的`md`资源链接创建。
- **风险：聚合菜单三层交互在`SSH`/`ssh -t`等非完整`TTY`场景下可能误判** → 缓解：复用既有`IsInteractiveTerminal`实现（`os.File.Stat()`+`ModeCharDevice`），与现有`skills`命令的判定保持一致，不引入额外终端检测库。
- **权衡：每资源独立注册表无法直接回答"`claude-code`涉及哪些资源"** → 接受：聚合菜单按资源维度展开，开发者交互意图就是"维护某资源类型的所有相关`Agent`"，而不是"维护某`Agent`的所有资源"；如果未来确实需要"`Agent`视角"视图，再加一个`agents.list`只读命令即可，不影响现有结构。
- **权衡：删除`skills.mk`、新增`agents.mk`后，`hack/makefiles/`目录文件名不再与`linactl`命令第一段精确对应**（`agents.mk`包含`agents`、`agents.skills.*`、`agents.prompts.*`、`agents.md.*`所有目标） → 接受：`Makefile include`粒度按"顶层命令组"划分更合理，否则会出现`agents.skills.mk`/`agents.prompts.mk`/`agents.md.mk`三个文件的过度拆分。

## Migration Plan

1. **代码迁移**：先把`skilllink`内部代码搬到`internal/agents/skills/`并改`import`路径，确认`go build ./hack/tools/linactl/...`通过、所有现有测试通过。
2. **抽公共**：把通用能力提到`internal/agents/common/`，`skills`子包改为`import common`，再次跑测试确认行为零变化。
3. **新增`prompts`和`md`子包**：实现注册表、`Inspect`/`Apply*`、`*Candidates`，附测试。
4. **CLI 与 Make**：删除旧`command_skills*.go`/`skills.mk`，新增`command_agents*.go`/`agents.mk`，更新`commandRegistry()`、根`Makefile` `include`、`README`迁移指引段落。
5. **聚合菜单**：实现`runAgents`三层交互，跨子包调用`*Candidates`。
6. **OpenSpec 规范同步**：在本变更的`specs/agent-skills-link-cli/spec.md`增量中`MODIFIED`命令名相关`Requirement`；在`specs/agents-multi-resource/spec.md`新增整体能力规范。
7. **执行验证**：`go test ./hack/tools/linactl/... -count=1`、`go run ./hack/tools/linactl test.scripts`、`make agents`/`make agents.skills.link`/`make agents.md.link AGENT=claude-code`本地烟测；`/lina-review`扫描旧命令字符串残留。

**回滚策略**：本次重构在单一`PR`内完成，回滚直接`git revert`即可恢复`skilllink`包与`skills.mk`。变更不涉及数据库、运行时配置或外部依赖，无状态需要清理。

## Open Questions

1. `cursor`/`codex`/`gemini-cli`的`prompts`目录路径是否已在各家官方文档中稳定下来？若实施时发现某`Agent`官方路径仍在快速变化，应在该`Agent`条目前打`TODO`注释并暂不纳入注册表，避免错误链接误导用户。
2. `linapro-site`相关文档更新是否需要在主仓变更的`tasks.md`中显式列出？倾向**列出但标记为"主仓只声明影响面，实际提交在`linapro-site`独立仓"**，由`/lina-review`时核对是否同步推进。
