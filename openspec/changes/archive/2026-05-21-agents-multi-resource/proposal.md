## Why

`LinaPro`已经通过`hack/tools/linactl`下的`skills.link`/`skills.unlink`命令统一管理`.agents/skills/`到各`AI Coding Agent`私有项目路径（如`.claude/skills/`、`.codebuddy/skills/`、`.windsurf/skills/`等）的软链。但`Agent Skills`并不是仓库与`Agent`之间唯一需要桥接的项目资源：

- **斜杠命令/Prompts**：仓库已在`.agents/prompts/opsx/`维护`OpenSpec`斜杠指令源文件，并已手工建立`.claude/commands/opsx/ -> .agents/prompts/opsx/`软链；其他主流`Agent`（如`Cursor`、`Codex`、`Gemini CLI`）也通过各自私有目录暴露斜杠命令或`prompts`扩展。当前没有任何工具帮助开发者维护这一类映射。
- **AGENTS.md 项目规范文件**：`AGENTS.md`已成为多家`Agent`厂商共同认可的项目级规范约定（`agents.md`），部分`Agent`原生读取仓库根的`AGENTS.md`（如`Codex`、`Cursor`、`Amp`、`OpenCode`、`Cline`、`Warp`等），部分`Agent`仍读取私有文件名（如`Claude Code` → `CLAUDE.md`、`Gemini CLI` → `GEMINI.md`、`Qwen Code` → `QWEN.md`、`Junie` → `.junie/guidelines.md`）。仓库当前仅通过手工`CLAUDE.md -> AGENTS.md`软链覆盖`Claude Code`，没有可扩展的工具支持其他`Agent`。

继续为每类资源单独造一套命令、单独维护注册表、单独写一份交互式选择器和软链状态机，会让`hack/tools/linactl`迅速膨胀且难以演进。本次需要把现有`skilllink`重构为面向多资源的`agents`框架，统一管理`skills`、`prompts`和`AGENTS.md`三类项目资源的软链桥接。

## What Changes

### 命令面整体重构（**BREAKING**）

- **BREAKING**：删除`linactl skills`、`linactl skills.link`、`linactl skills.unlink`命令以及对应的`make skills`、`make skills.link`、`make skills.unlink`目标和`hack/makefiles/skills.mk`。
- 新增以`agents`为顶层、以资源类型为第二段的命令树：
  - `linactl agents`：聚合菜单，`TTY`下提供三层交互（资源类型 → link/unlink 动作 → 具体`Agent`选择），非`TTY`下打印用法指引。
  - `linactl agents.skills.link` / `linactl agents.skills.unlink`：迁移自原`skills.link` / `skills.unlink`，行为与既有规范保持一致。
  - `linactl agents.prompts.link` / `linactl agents.prompts.unlink`：新增，管理`Agent`私有命令/`prompts`目录到`.agents/prompts/...`下显式声明源路径的软链。
  - `linactl agents.md.link` / `linactl agents.md.unlink`：新增，管理`Agent`私有项目规范文件到仓库根`AGENTS.md`的**单文件**软链。
- 同步在根`Makefile`新增`hack/makefiles/agents.mk`，提供等价的`make agents`、`make agents.skills.link`/`unlink`、`make agents.prompts.link`/`unlink`、`make agents.md.link`/`unlink`目标。

### 内部包结构重构

- 删除`hack/tools/linactl/internal/skilllink/`包，替换为`hack/tools/linactl/internal/agents/`下的多子包结构：
  - `internal/agents/common/`：抽取原`skilllink`中真正与资源类型无关的能力，包括`Status`枚举集合、表格渲染、交互式选择（`PromptSelection`/`PromptYesNo`/`IsInteractiveTerminal`/`ReadLine`）、目录级软链状态机（`linkMatchesSource`/`pathsEqual`/`symlinkErrorDetail`）、选择器解析（`ParseSelectors`/`SelectorAll`/`resolveTargets`/`targetPolicy`）。
  - `internal/agents/skills/`：迁移原`skilllink`的`Agent`注册表与目录级桥接逻辑，复用`common`。
  - `internal/agents/prompts/`：新增，目录级桥接逻辑复用`common`，但`Agent`注册表为每个`Agent`显式声明源路径（`SourcePath`）和目标路径（`ProjectPath`）。
  - `internal/agents/md/`：新增，承载**单文件**软链状态机（区别于目录级）和`AGENTS.md`资源`Agent`注册表。
- `hack/tools/linactl`下按命名规范新增`command_agents.go`（聚合菜单）、`command_agents.skills.link.go`、`command_agents.skills.unlink.go`、`command_agents.prompts.link.go`、`command_agents.prompts.unlink.go`、`command_agents.md.link.go`、`command_agents.md.unlink.go`，并删除`command_skills.go`、`command_skills.link.go`、`command_skills.unlink.go`。

### 注册表初始内容

- `skills`资源注册表与既有`skilllink_agents.go`完全一致，迁移过程中不调整任何条目（`native` 15 项 + `link` 39 项 + `rootCollision` 1 项）。
- `prompts`资源初始注册表覆盖：`claude-code`（源`.agents/prompts/opsx/`，目标`.claude/commands/opsx/`）、`cursor`、`codex`、`gemini-cli`，每个`Agent`显式声明源路径和目标路径，未来可逐步扩展。
- `md`资源注册表覆盖全套已知`Agent`：
  - `link`类（建立私有文件 → `AGENTS.md`软链）：`claude-code`→`CLAUDE.md`、`gemini-cli`→`GEMINI.md`、`qwen-code`→`QWEN.md`、`junie`→`.junie/guidelines.md`、`windsurf`→`.windsurfrules`、`augment`→`.augment-guidelines`、`continue`→`.continuerules`等；
  - `native`类（在状态表中显示但不建链）：`codex`、`cursor`、`amp`、`opencode`、`cline`、`warp`、`replit`、`antigravity`、`deepagents`、`dexto`、`firebender`、`github-copilot`、`kimi-cli`、`universal`等。

### 作用域与不变量

- 工具仅在仓库根目录内操作项目路径，不引入`--root`参数，不扫描嵌套`git`仓库。
- 单文件软链状态机与目录级软链状态机的语义保持完全一致：`native`/`ok`/`created`/`rebuilt`/`mismatch`/`conflict`/`removed`/`skipped-foreign`/`skipped-not-managed`/`absent`/`error`，且**永不删除真实文件或目录**，`force=1`仅作用于"目标已是软链但指向其他位置"的情况。
- 跨平台实现仅依赖`Go`标准库（`os.Symlink`/`os.Readlink`/`os.Lstat`/`os.Remove`/`os.MkdirAll`/`filepath.Rel`/`filepath.Clean`），不调用`ln`/`mklink`/`bash`/`cmd.exe`。

### 文档与历史治理

- 更新`hack/tools/linactl/README.md`和`README.zh-CN.md`：移除`skills.link`/`skills.unlink`章节，替换为`agents`命令树章节，并在迁移指引段落中明确告知`make skills.*`已被删除以及对应`make agents.skills.*`等价命令。
- 更新`linapro-site/apps/lina-site/docs/docs/5000-tools/2000-agent-skills.md`和`docs/quick/4000-agent-tools.md`：对齐新命令名，并新增`agents.prompts`、`agents.md`两节。
- `.gitignore`新增`agents.prompts`管理的软链路径忽略规则（如`/.claude/commands/opsx`）；`agents.md`不修改`.gitignore`，因为`AGENTS.md`本身被纳入版本控制，`CLAUDE.md`等链接文件已存在或可按`Agent`私有目录的现有规则忽略。

### 不在本次范围

- 不调整`.agents/skills/`、`.agents/prompts/`、`AGENTS.md`本身的内容、目录结构或加载顺序。
- 不下载或同步`vercel-labs/skills`内容；不替换`skills-lock.json`从远端拉取`Skill`包的现有流程。
- 不引入第三方交互式`TUI`框架；交互式实现继续仅依赖`Go`标准库。
- 不实现"跨资源批量"快捷命令（如`agents.link --all-resources`）；多资源协同通过`agents`聚合菜单的三层选择体验完成。
- 不为`AGENTS.md`资源建立"自动写入文件内容"等非软链式桥接（仅维护单文件软链）。

## Capabilities

### New Capabilities

- `agents-multi-resource`：定义`agents`命令树整体能力边界，包括三类资源（`skills`/`prompts`/`md`）共享的`native`/`link`/`rootCollision`分类、目录级与文件级软链状态机、`TTY`三层交互菜单、跨平台`Go`标准库实现约束以及`Agent`注册表治理规则。

### Modified Capabilities

- `agent-skills-link-cli`：原`skills.link`/`skills.unlink`命令面被新`agents.skills.link`/`agents.skills.unlink`命令替换，需在该规范下新增"命令名称从`skills.*`迁移到`agents.skills.*`"的`Requirement`变更，确保规范与实现保持一致；`make skills.*`目标和`hack/makefiles/skills.mk`一并删除。

## Impact

- 影响`hack/tools/linactl`（删除`command_skills*.go`和`internal/skilllink/`，新增`command_agents.go`/`command_agents.skills.*`/`command_agents.prompts.*`/`command_agents.md.*`和`internal/agents/{common,skills,prompts,md}/`）。
- 影响根`Makefile`、删除`hack/makefiles/skills.mk`、新增`hack/makefiles/agents.mk`。
- 影响`hack/tools/linactl/README.md`和`README.zh-CN.md`双语文档。
- 影响`linapro-site/apps/lina-site/docs/docs/5000-tools/2000-agent-skills.md`和`linapro-site/apps/lina-site/docs/quick/4000-agent-tools.md`（注：`linapro-site`已不在主仓索引中，本仓变更只声明影响面，文档实际更新可同步推到`linapro-site`独立仓）。
- 影响`.gitignore`：新增`agents.prompts`管理的软链路径忽略规则。
- 不涉及后端运行时服务、`REST API`、数据库结构、前端页面、用户可见运行时文案、运行时缓存、数据权限或`i18n`运行时资源。命令是开发期一次性`os.Symlink`/`os.Remove`操作，不引入运行时缓存、跨实例协调或集群一致性问题。
- **破坏性影响**：所有现有依赖`make skills.link`/`make skills.unlink`/`linactl skills.*`的本地工作流、文档、`CI`脚本和`Agent`使用习惯都需要切换到`make agents.skills.*`/`linactl agents.skills.*`；本次变更需在`README`迁移指引中明确说明，并通过`/lina-review`审查确认仓库内不再残留旧命令引用。
- 跨平台影响：所有目录级与文件级软链操作完全使用`Go`标准库实现；`Windows`平台软链不可用时复用现有错误指引（开发者模式或管理员）。
