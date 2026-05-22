# agents-multi-resource 实现任务清单

> 实施前提：本次变更涉及命令面破坏性变更与多包重构，建议在单一`PR`内完成所有任务。每个任务组完成后运行 `cd hack/tools/linactl && go test ./... -count=1` 烟测，避免后续任务在编译失败的基础上推进。

## 1. 准备：迁移现有 skilllink 到新包路径（行为零变化）

- [x] 1.1 在`hack/tools/linactl/internal/`下新建`agents/`目录，并新建子目录`common/`、`skills/`、`prompts/`、`md/`，确保`go build`阶段空目录不影响构建（暂不放代码）
- [x] 1.2 将`internal/skilllink/skilllink.go`包注释迁入`internal/agents/skills/skills.go`作为该子包主文件，更新包名为`skills`，包注释说明该子包承载`skills`资源的`Agent`注册表与桥接逻辑
- [x] 1.3 将`internal/skilllink/skilllink_agents.go`整体迁入`internal/agents/skills/skills_agents.go`，包名改为`skills`，类型与函数（`AgentSpec`/`Category`/`agents`/`SourceDir`/`Agents`/`FindAgent`）保持公开签名一致
- [x] 1.4 将`internal/skilllink/skilllink_apply.go`迁入`internal/agents/skills/skills_apply.go`，包名改为`skills`，公开`Inspect`/`PlanList`/`ApplyLink`/`ApplyUnlink`/`Result`/`Status`等签名暂保持不变（后续步骤再外提）
- [x] 1.5 将`internal/skilllink/skilllink_render.go`迁入`internal/agents/skills/skills_render.go`、`skilllink_selectors.go`迁入`skills_selectors.go`、`skilllink_interactive.go`迁入`skills_interactive.go`，包名统一改为`skills`
- [x] 1.6 将`internal/skilllink/skilllink_test.go`、`skilllink_interactive_test.go`整体迁入`internal/agents/skills/`下相同文件名，更新`package`声明为`skills`，导入路径同步调整
- [x] 1.7 删除`internal/skilllink/`整个目录
- [x] 1.8 全仓更新`linactl/internal/skilllink`导入路径为`linactl/internal/agents/skills`，包括`command_skills.go`、`command_skills.link.go`、`command_skills.unlink.go`等
- [x] 1.9 运行`cd hack/tools/linactl && go test ./... -count=1`，确认所有原 skills 测试通过、行为零变化

## 2. 抽取通用能力到 common 子包

- [x] 2.1 在`internal/agents/common/common.go`创建包主文件，包注释说明`common`是面向多资源的通用能力子包，承载`Status`枚举、`Result`类型、表格渲染、提示行、选择器解析、交互式终端检测与提示、目录级与文件级共享的软链状态机辅助
- [x] 2.2 将`Status`枚举集合、`Result`类型、`HasError`/`IsError`从`skills`子包迁入`common/common_status.go`；`skills`子包改为`type Status = common.Status`并重新导出常量，保持外部签名不破坏
- [x] 2.3 将`Render`/`EmitHints`从`skills`子包迁入`common/common_render.go`；扩展`Render`签名为`Render(out io.Writer, results []Result, extraColumns ...ColumnSpec) error`，新增`type ColumnSpec struct{ Header string; Value func(Result) string }`；`skills`和`md`资源调用时传零额外列，`prompts`资源调用时传一列`SOURCE`
- [x] 2.4 将`ParseSelectors`/`SelectorAll`/`hasAll`/`targetPolicy`/`resolveTargets`从`skills`子包迁入`common/common_selectors.go`；将`AgentSpec`抽象的部分（仅依赖`Name`和`Category`字段）下沉到`common`，每资源子包定义自己的具体`AgentSpec`实现`SpecLike`接口（`SpecName()`/`SpecCategory()`），`resolveTargets`基于该接口工作
- [x] 2.5 将`IsInteractiveTerminal`/`ReadLine`/`PromptSelection`/`PromptYesNo`/`SelectableEntry`/`renderCandidateGrid`/`statusGlyph`从`skills`子包迁入`common/common_interactive.go`，签名保持不变；`SelectableEntry`同样基于`SpecLike`接口承载`Spec`字段
- [x] 2.6 将`linkMatchesSource`/`pathsEqual`/`equalFoldPath`/`symlinkErrorDetail`从`skills`子包迁入`common/common_filesystem.go`；扩展`linkMatchesSource`签名为`linkMatchesSource(repoRoot, link, sourceAbs string) (bool, string, error)`，源路径不再从常量推断而是按调用方传入
- [x] 2.7 在`skills`子包`createLink`内部把原本读取`SourceDir`常量的逻辑改为读取`Spec.SourcePath`字段；`skills`子包`AgentSpec`新增`SourcePath`字段并默认填充为`.agents/skills`，对外行为保持一致
- [x] 2.8 运行`cd hack/tools/linactl && go test ./... -count=1`，确认抽取`common`后`skills`资源行为零变化、所有现有测试仍通过

## 3. 实现 prompts 资源子包

- [x] 3.1 在`internal/agents/prompts/prompts.go`创建包主文件，包注释说明`prompts`是面向多资源框架中`prompts`/`commands`类资源的子包，承载该资源的`Agent`注册表与桥接逻辑（注册表条目显式声明`SourcePath`）
- [x] 3.2 在`prompts/prompts_agents.go`实现`AgentSpec`结构（`Name`/`DisplayName`/`SourcePath`/`ProjectPath`/`Category`），并实现`SpecLike`接口；在`agents []AgentSpec`变量中初始化 4 项注册表：`claude-code`(`.agents/prompts/opsx`→`.claude/commands/opsx`)、`cursor`(`.agents/prompts/opsx`→`.cursor/commands/opsx`)、`codex`(`.agents/prompts/opsx`→`.codex/prompts/opsx`)、`gemini-cli`(`.agents/prompts/opsx`→`.gemini/commands/opsx`)，所有条目均为`Category=link`
- [x] 3.3 在`prompts/prompts_apply.go`实现`Inspect`/`PlanList`/`ApplyLink`/`ApplyUnlink`/`LinkRequest`/`UnlinkRequest`，行为复用`common`的目录级状态机；在调用`common.linkMatchesSource`时传入该条目的`SourcePath`绝对路径
- [x] 3.4 在`prompts/prompts_interactive.go`实现`LinkCandidates`/`UnlinkCandidates`，复用`common.SelectableEntry`和`common.PromptSelection`；与`skills`版本相同，但仅扫描`prompts`注册表
- [x] 3.5 在`prompts/prompts_test.go`新增单元测试，覆盖：注册表非空且`SourcePath`非空、`Inspect`对四种状态（native/link 已建/link absent/conflict）的返回正确性、`ApplyLink`创建相对软链且目标命中`SourcePath`、`ApplyUnlink`严格匹配`SourcePath`不删除指向其他位置的软链
- [x] 3.6 运行`cd hack/tools/linactl && go test ./internal/agents/prompts/... -count=1`，确认 prompts 子包独立通过

## 4. 实现 md 资源子包（单文件软链）

- [x] 4.1 在`internal/agents/md/md.go`创建包主文件，包注释说明`md`是面向多资源框架中`AGENTS.md`项目规范文件桥接的子包，承载单文件软链状态机与`Agent`注册表
- [x] 4.2 在`md/md_agents.go`实现`AgentSpec`结构（`Name`/`DisplayName`/`ProjectPath`/`Category`，`SourcePath`固定为`AGENTS.md`），并实现`SpecLike`接口；在`agents []AgentSpec`中初始化注册表：
  - `link`类约 8-10 项：`claude-code`→`CLAUDE.md`、`gemini-cli`→`GEMINI.md`、`qwen-code`→`QWEN.md`、`junie`→`.junie/guidelines.md`、`windsurf`→`.windsurfrules`、`augment`→`.augment-guidelines`、`continue`→`.continuerules`等
  - `native`类约 14 项：`codex`、`cursor`、`amp`、`opencode`、`cline`、`warp`、`replit`、`antigravity`、`deepagents`、`dexto`、`firebender`、`github-copilot`、`kimi-cli`、`universal`等
  - 不包含`rootCollision`类条目
- [x] 4.3 在`md/md_apply.go`实现`Inspect`/`PlanList`/`ApplyLink`/`ApplyUnlink`，复用`common`提供的目录级状态机骨架，但在`Lstat`后判断"非 symlink"分支时输出 detail 文案为`real file exists; resolve manually`（与目录级"real path exists"对齐但描述精确为文件）
- [x] 4.4 在`md/md_interactive.go`实现`LinkCandidates`/`UnlinkCandidates`，仅扫描`md`注册表
- [x] 4.5 在`md/md_test.go`新增单元测试，覆盖：`link`类创建`CLAUDE.md`相对软链且`Readlink`命中`AGENTS.md`、`native`类输出`StatusNative`不动文件系统、目标已是真实文件时输出`StatusConflict`且`force=1`不删除真实文件、目标是指向其他位置的软链时按`force`决定`mismatch`/`rebuilt`、`unlink`严格匹配指向`AGENTS.md`的软链
- [x] 4.6 运行`cd hack/tools/linactl && go test ./internal/agents/md/... -count=1`，确认 md 子包独立通过

## 5. 命令树重构与旧命令删除

- [x] 5.1 在`hack/tools/linactl/`新增`command_agents.skills.link.go`，迁移自原`command_skills.link.go`：函数名改为`runAgentsSkillsLink`/`runAgentsSkillsLinkInteractive`/`executeAgentsSkillsLink`，导入路径改为`linactl/internal/agents/skills`和`linactl/internal/agents/common`，行为完全一致
- [x] 5.2 新增`command_agents.skills.unlink.go`，迁移自原`command_skills.unlink.go`，函数与导入同上调整
- [x] 5.3 新增`command_agents.prompts.link.go`和`command_agents.prompts.unlink.go`，结构对照`skills`版本，调用`prompts`子包；`Render`调用传入`SOURCE`列；`runAgentsPromptsLinkInteractive`复用`common.PromptSelection`
- [x] 5.4 新增`command_agents.md.link.go`和`command_agents.md.unlink.go`，结构对照`skills`版本，调用`md`子包；`Render`调用不传额外列
- [x] 5.5 新增`command_agents.go`，承载`runAgents`聚合菜单：`TTY`下三层交互（资源 → 动作 → 进入对应子命令的 TTY 流程），非`TTY`下打印六条子命令用法指引，使用`writeLine`/`writeLines`复用既有辅助
- [x] 5.6 在`command.go`的`commandRegistry()`中注册七条新命令：`agents`、`agents.skills.link`、`agents.skills.unlink`、`agents.prompts.link`、`agents.prompts.unlink`、`agents.md.link`、`agents.md.unlink`，并删除原`skills`、`skills.link`、`skills.unlink`三条注册
- [x] 5.7 删除`command_skills.go`、`command_skills.link.go`、`command_skills.unlink.go`三个文件
- [x] 5.8 运行`cd hack/tools/linactl && go build ./... && go test ./... -count=1`，确认所有命令编译通过、注册表更新生效、原`skills`测试随文件删除而清理

## 6. Make 集成与根目录治理

- [x] 6.1 新增`hack/makefiles/agents.mk`，提供`agents`、`agents.skills.link`、`agents.skills.unlink`、`agents.prompts.link`、`agents.prompts.unlink`、`agents.md.link`、`agents.md.unlink`七个目标，模式参照原`skills.mk`使用`$(LINACTL)`和`$(if $(AGENT),agent=$(AGENT))`/`$(if $(FORCE),force=1)`
- [x] 6.2 删除`hack/makefiles/skills.mk`
- [x] 6.3 更新根`Makefile`的`include`段，将`include hack/makefiles/skills.mk`替换为`include hack/makefiles/agents.mk`
- [x] 6.4 更新`.gitignore`：将"Agent skill symlinks managed by `make skills.link`"分组注释更新为"Agent symlinks managed by `make agents.skills.link` / `make agents.prompts.link`"，新增`prompts`资源管理的软链路径忽略规则（如`/.claude/commands/opsx`、`/.cursor/commands/opsx`、`/.codex/prompts/opsx`、`/.gemini/commands/opsx`）；`md`资源链接（`CLAUDE.md`等）已被各自`Agent`私有目录规则或文件级规则覆盖，无需在本次新增
- [x] 6.5 在终端运行`make agents`确认聚合菜单可进入，运行`make agents.skills.link`/`make agents.prompts.link`/`make agents.md.link`分别确认状态表正确输出（不实际建链，仅 Plan 模式）

## 7. 文档与迁移指引

- [x] 7.1 更新`hack/tools/linactl/README.md`：移除原`skills.link`/`skills.unlink`章节；新增"`agents` 命令树"主章节，分三小节描述`agents.skills.*`/`agents.prompts.*`/`agents.md.*`；新增"迁移指引"段落，列对照表说明`make skills.link` → `make agents.skills.link`、`make skills.unlink` → `make agents.skills.unlink`、`make skills` → `make agents`，并明确旧命令已被删除而非保留为别名
- [x] 7.2 同步更新`hack/tools/linactl/README.zh-CN.md`，内容与英文版一一对应
- [x] 7.3 更新仓库根`AGENTS.md`：若文档中有提到`skills.link`/`skills.unlink`命令的位置（搜索`skills.link`/`skills.unlink`/`make skills`字符串），全部替换为新命令名；若没有则跳过该任务
- [x] 7.4 在`/lina-review`审查阶段全仓扫描旧命令字符串残留：执行`rg -n 'make skills(\.link|\.unlink|\b)' --glob '!openspec/changes/archive/**' --glob '!openspec/changes/agents-multi-resource/**'` 与 `rg -n 'linactl skills(\.link|\.unlink|\b)' --glob '!openspec/changes/archive/**' --glob '!openspec/changes/agents-multi-resource/**'`，确认无命中（归档目录与本变更目录的迁移指引可保留旧命令字符串）
- [x] 7.5 声明`linapro-site`文档影响面：在本任务记录中明确"`linapro-site/apps/lina-site/docs/docs/5000-tools/2000-agent-skills.md`和`linapro-site/apps/lina-site/docs/quick/4000-agent-tools.md`需要同步更新对应命令名"，但实际更新通过`linapro-site`独立仓提交，不在本主仓变更中进行

## 8. 端到端验证与编译门禁

- [x] 8.1 运行`cd hack/tools/linactl && go test ./... -count=1`，确认所有单元测试通过
- [x] 8.2 运行`cd /data/home/v_hlaghuang/project/public/linapro && go run ./hack/tools/linactl test.scripts`，确认仓库工具 smoke 测试通过
- [x] 8.3 在仓库根分别运行`make agents.skills.link AGENT=claude-code`、`make agents.prompts.link AGENT=claude-code`、`make agents.md.link AGENT=claude-code`，确认三类资源的链接操作均能成功创建预期软链；运行对应`unlink`命令确认能正确移除
- [x] 8.4 在仓库根运行`make agents.md.link AGENT=codex`，确认输出`native`状态且不创建任何文件系统对象
- [x] 8.5 在仓库根准备一个故意指向其他位置的`CLAUDE.md`软链（如指向`/tmp/foo`），运行`make agents.md.link AGENT=claude-code`确认输出`mismatch`不修改；再运行`make agents.md.link AGENT=claude-code FORCE=1`确认输出`rebuilt`并指向`AGENTS.md`
- [x] 8.6 在终端运行`make agents`三层菜单走通"`skills` → `link` → 选择 1 个 Agent"完整流程，确认与`make agents.skills.link`直接调用的体验一致
- [x] 8.7 运行`make help`确认输出包含七条新命令、不包含三条旧命令

## 9. OpenSpec 校验与归档前审查

- [x] 9.1 运行`openspec validate agents-multi-resource --strict`，确认变更通过 OpenSpec 校验
- [x] 9.2 运行`openspec list --json`确认本变更状态为`active`且任务清单可解析
- [x] 9.3 调用`/lina-review`技能进行变更审查，重点核对：通用能力是否真的下沉到`common`、每资源子包是否仅依赖`common`、命令命名是否严格按命令树形态、旧命令字符串是否已清理、文档迁移指引是否完整、跨平台 Go 标准库实现约束是否满足
- [x] 9.4 审查通过后，在用户确认本次迭代完成时通过`/opsx:archive`将本变更归档；归档时保持文档原始中文语言（与 propose 阶段 q14 决策一致），并将本变更对`agent-skills-link-cli`的修改同步合并到`openspec/specs/agent-skills-link-cli/spec.md`基线

## Feedback

- [x] **FB-1**: `agents.md.link`和`agents.skills.link`的 TTY 交互模式在进入候选`grid`前未渲染包含`native`类`Agent`的完整状态总览表，导致用户无法看到`native`类`Agent`的存在，与`md`资源 propose 阶段确认的"显示全部 native"诉求不符
- [x] **FB-2**: `md`注册表覆盖范围不全（22 项），相对于`skills`注册表（55 项）漏注册 33 个`Agent`（含`codebuddy`），需要逐个调研其对`AGENTS.md`的实际支持情况，将有可靠证据的`Agent`补全到`md`注册表，未能查到可靠证据的明确跳过
- [x] **FB-3**: 根据 CodeBuddy 官方文档（CodeBuddy 在 `CODEBUDDY.md` 不存在时会自动 fallback 读取 `AGENTS.md`），将`codebuddy`在`md`注册表中的分类从`link`改为`native`，并在`linactl README`（中英双版）补充对`fallback`机制的事实陈述说明
