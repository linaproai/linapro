# 改进 make agents 展示内容任务清单

## 1. 交互选择展示

- [x] 1.1 将 `make agents` / `linactl agents` 的 Agent 选择项改为只展示人类可读名称。
- [x] 1.2 将单选提示说明改为 `Arrow keys to navigate, Enter to confirm, Esc to cancel, any key to search.`。

## 2. 执行结果展示

- [x] 2.1 为聚合 `agents` 命令新增简化资源结果表格，按 `RESOURCE`、`STATUS`、`DETAIL` 展示。
- [x] 2.2 聚合 `agents` 命令直接调用各资源 apply API，避免重复输出三段资源明细表；资源子命令保留原详细表格。
- [x] 2.3 保留冲突、mismatch、error 等关键异常提示，但避免冗余路径/分类列和重复 Summary 文本。

## 3. 验证与审查

- [x] 3.1 新增或更新 `hack/tools/linactl` 单元测试覆盖简化选择项和聚合结果表格。
- [x] 3.2 运行 `cd hack/tools/linactl && go test ./... -count=1`。
- [x] 3.3 运行 `openspec validate improve-agents-display --strict`。
- [x] 3.4 记录 i18n、缓存、数据权限、REST API、E2E、开发工具跨平台影响。
- [x] 3.5 完成 `lina-review` 审查。

## Feedback

- [x] **FB-1**: `agents.prompts.link` 为 Claude Code 创建 `.claude/commands/opsx -> .agents/prompts/opsx` 子目录软链，导致已有且更合理的 `.claude/commands -> ../.agents/prompts` 父级目录软链被误判为 `conflict`。
- [x] **FB-2**: `make agents agent=ClaudeCode` 应与 `make agents agent=claude-code` 等价，且 Make 参数应优先使用小写 `agent/action/force`，大写 `AGENT/ACTION/FORCE` 仅作为兼容别名。
- [x] **FB-3**: `make agents` 系列目标会回显 `cd hack/tools/linactl && go run . ...` 委托命令，导致终端输出包含非业务噪音。

## Verification Notes

- 交互选择展示：`selectableAgent.optionLabel()` 只返回 `DisplayName` 或回退到 `Name`，不再拼接内部标识、资源分类或状态摘要；`PromptSingleSelection` 描述文案已更新为 `Arrow keys to navigate, Enter to confirm, Esc to cancel, any key to search.`。
- 执行结果展示：`dispatchAgentSetup` 直接调用 `skills.ApplyLink/ApplyUnlink`、`prompts.ApplyLink/ApplyUnlink`、`md.ApplyLink/ApplyUnlink`，并渲染 `RESOURCE`、`STATUS`、`DETAIL` 三列表；`agents.skills.*`、`agents.prompts.*`、`agents.md.*` 子命令仍使用原有 `common.Render` 详细表格。
- 单元测试：新增 `TestSelectableAgentOptionLabelUsesDisplayNameOnly`，更新 `TestDispatchAgentSetupSkipsUnregisteredResources` 断言聚合输出包含简化表格且不包含 `Summary:`、`== skills ==`、`PROJECT PATH`、`CATEGORY`。
- 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- 验证通过：`openspec validate improve-agents-display --strict`。
- 验证通过：`git diff --check -- hack/tools/linactl/command_agents.go hack/tools/linactl/command_agents_test.go hack/tools/linactl/internal/agents/common/common_interactive_huh.go openspec/changes/improve-agents-display`。
- FB-1 修复：`prompts` 注册表改为父级目录软链，四个初始 Agent 均以 `.agents/prompts` 为 `SourcePath`；Claude Code 的 `ProjectPath` 为 `.claude/commands`，Codex 为 `.codex/prompts`，Cursor 为 `.cursor/commands`，Gemini CLI 为 `.gemini/commands`。
- FB-1 测试：更新 `internal/agents/prompts` 单元测试，断言 `claude-code` 创建 `.claude/commands -> ../.agents/prompts`，且 `.claude/commands/opsx` 可通过父级软链访问。
- FB-1 文档：同步更新 `hack/tools/linactl/README.md`、`README.zh-CN.md`、`proposal.md`、`design.md` 和增量规范，明确 `agents.prompts` 管理 Agent prompts 根目录软链，而不是单个 `opsx` 子目录软链。
- FB-1 验证通过：`cd hack/tools/linactl && go test ./internal/agents/prompts -count=1`。
- FB-1 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- FB-1 验证通过：`go run . agents.prompts.link agent=claude-code` 输出 `claude-code  .agents/prompts  .claude/commands  link  ok`。
- FB-1 验证通过：`go run . agents agent=claude-code` 输出 `prompts   applied  ok`。
- FB-1 验证通过：`openspec validate improve-agents-display --strict`。
- FB-1 验证通过：`git diff --check -- hack/tools/linactl openspec/changes/improve-agents-display`。
- FB-2 修复：新增 Agent 名称归一化，使用 `gstr.CaseKebab` 将 `ClaudeCode`、`Claude Code`、`claude_code` 等输入统一为注册表使用的 `claude-code` 形式；聚合命令和各资源子命令共享该解析规则。
- FB-2 修复：`linactl` 参数 key 解析改为大小写不敏感，`AGENT=ClaudeCode` 和 `agent=ClaudeCode` 进入同一个 `agent` 参数槽；Makefile 优先读取小写 `agent/action/force`，并保留大写 `AGENT/ACTION/FORCE` 兼容历史用法。
- FB-2 测试：新增 `NormalizeAgentName`、`ParseSelectors`、`ResolveTargets`、`validateSingleAgentName`、`runAgents` 一键模式和大写参数 key 解析单元测试，覆盖 `ClaudeCode -> claude-code`、`Claude Code -> claude-code`、`claude_code -> claude-code`、`AGENT=ClaudeCode` 等场景。
- FB-2 文档：同步更新 `hack/tools/linactl/README.md`、`README.zh-CN.md`、`proposal.md`、`design.md` 和增量规范，明确小写参数为推荐写法，大写 Make 变量仅兼容保留。
- FB-2 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- FB-2 验证通过：`go run . agents agent=ClaudeCode` 输出 `Agent: Claude Code`，并显示 `skills/prompts/md` 均为 `applied ok`。
- FB-2 验证通过：`go run . agents AGENT=ClaudeCode` 输出与小写参数一致。
- FB-2 验证通过：`make -n agents agent=ClaudeCode` 展开为 `cd hack/tools/linactl && go run . agents agent=ClaudeCode`。
- FB-2 验证通过：`make -n agents AGENT=claude-code ACTION=unlink FORCE=1` 展开为 `cd hack/tools/linactl && go run . agents agent=claude-code action=unlink force=1`。
- FB-2 验证通过：`openspec validate improve-agents-display --strict`。
- FB-2 验证通过：`git diff --check -- hack/tools/linactl hack/makefiles/agents.mk openspec/changes/improve-agents-display`。
- FB-3 修复：`hack/makefiles/agents.mk` 中 `agents`、`agents.skills.*`、`agents.prompts.*`、`agents.md.*` 目标统一使用 `@$(LINACTL)`，屏蔽 Make recipe 自身的委托命令回显，同时保留 `linactl` 的业务输出。
- FB-3 规范：增量规范新增 Make 包装层不回显 `cd hack/tools/linactl && go run . ...` 委托命令行的场景。
- FB-3 影响分析：本次只修改 Make 包装层和 OpenSpec 文档，不改变 `linactl` 命令参数、软链状态机、文件系统写入规则或子命令业务输出。
- FB-3 验证通过：静态扫描 `rg -n '^\t\$\(LINACTL\)' hack/makefiles Makefile` 未发现未静默的 `$(LINACTL)` recipe。
- FB-3 验证通过：`make agents agent=claude-code` 输出只包含 `Agent: Claude Code`、`Action: link` 和资源汇总表，不再包含 `cd hack/tools/linactl && go run . ...`。
- FB-3 验证通过：`make agents.prompts.link agent=claude-code` 输出只包含详细资源表，不再包含 `cd hack/tools/linactl && go run . ...`。
- FB-3 验证通过：`cd hack/tools/linactl && go test ./... -count=1`。
- FB-3 验证通过：`openspec validate improve-agents-display --strict`。
- FB-3 Review：已按 `lina-review` 口径完成审查。审查范围覆盖 `hack/makefiles/agents.mk`、`openspec/changes/improve-agents-display/tasks.md`、`openspec/changes/improve-agents-display/specs/agents-multi-resource/spec.md`；确认 Make 包装层仅屏蔽 recipe 回显，不吞掉 `linactl` 业务输出；未修改后端运行时 Go 生产代码、REST API、前端页面、运行时 i18n、缓存或数据权限逻辑。
- i18n 影响：本次仅修改开发工具 CLI 英文提示和 OpenSpec 文档，不修改前端运行时语言包、宿主/插件 `manifest/i18n` 或 apidoc i18n JSON；不需要新增运行时翻译键。
- 缓存一致性影响：本次不新增或修改运行时业务缓存、缓存键、失效触发、分布式协调或跨实例一致性逻辑。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口、服务数据访问路径、插件宿主服务适配器或聚合统计，不影响角色数据权限边界。
- REST API 影响：本次不新增或修改 REST API。
- E2E 影响：本次为开发工具 CLI 展示改进，不涉及用户可观察前端页面、路由、表单、表格或端到端业务流程；使用 `linactl` 单元测试和 OpenSpec 校验覆盖。
- Go 生产代码影响：本次修改 `hack/tools/linactl` 开发工具 Go 代码，不涉及 `apps/lina-core` 后端运行时生产包、Controller 构造、路由绑定、启动编排或 API 接口签名；已运行 `cd hack/tools/linactl && go test ./... -count=1` 覆盖变更包。开发工具/脚本影响已通过同一 Go 工具测试和 `git diff --check` 验证。
- Review：已按 `lina-review` 口径完成审查。审查范围来源包括 `git status --short`、`git ls-files --others --exclude-standard`、未跟踪 `openspec/changes/improve-agents-display/` 文件、`hack/tools/linactl` 相关 diff、OpenSpec strict 校验、Go 测试、真实仓库 `go run . agents.prompts.link agent=claude-code`、`go run . agents agent=claude-code`、`go run . agents agent=ClaudeCode`、`go run . agents AGENT=ClaudeCode` smoke、Make dry-run 和 whitespace 检查。确认聚合命令展示简化未破坏子命令详细输出；`prompts` 注册表已改为父级目录软链，当前 `.claude/commands -> ../.agents/prompts` 被识别为 `ok`，不会再因 `.claude/commands/opsx` 可访问而误报 `conflict`；Agent 选择器归一化与小写参数优先策略均有测试和 smoke 覆盖；严重问题 0，警告 0。
