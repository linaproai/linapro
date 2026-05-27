# 改进 make agents 展示内容设计

## Scope

本次只调整 `linactl agents` / `make agents` 聚合入口的展示层：

- Agent 选择项只展示人类可读名称。
- 聚合执行结果只展示资源级汇总表。
- `agents.skills.*`、`agents.prompts.*`、`agents.md.*` 子命令继续保留详细表格，用于排查路径、来源、分类和状态机细节。
- `agents.prompts` 的默认注册表从单个 catalog 子目录软链调整为 Agent prompts 根目录软链。
- Agent 命令行选择器增加归一化处理，小写 `agent/action/force` 成为推荐 Make 参数，大写 `AGENT/ACTION/FORCE` 继续兼容。

## Implementation

`selectableAgent` 新增 `optionLabel()`，统一生成聚合选择列表的显示文本。该方法优先返回 `DisplayName`，缺失时回退到内部 `Name`，不拼接资源分类、路径或运行时状态。

`PromptSingleSelection` 的描述文案更新为固定句式：`Arrow keys to navigate, Enter to confirm, Esc to cancel, any key to search.`。

`dispatchAgentSetup` 不再调用 `executeAgentsSkillsLink`、`executeAgentsPromptsLink`、`executeAgentsMdLink` 等会渲染详细表格的子命令执行函数，而是直接调用各资源包的 `ApplyLink` / `ApplyUnlink` API，收集每个资源的 `common.Result` 并投影为 `RESOURCE`、`STATUS`、`DETAIL` 三列表。

`prompts` 注册表中的四个初始 Agent 均使用 `.agents/prompts` 作为 `SourcePath`。Claude Code 对应 `.claude/commands`，Cursor 对应 `.cursor/commands`，Codex 对应 `.codex/prompts`，Gemini CLI 对应 `.gemini/commands`。这样 `.agents/prompts/opsx` 会自然暴露为对应 Agent 根目录下的 `opsx` catalog，且已有父级软链会被识别为 `ok`。

`common.NormalizeAgentName()` 使用 GoFrame `gstr.CaseKebab` 将用户输入归一化为注册表使用的标准标识。`ClaudeCode`、`Claude Code`、`claude_code` 和 `claude-code` 都解析为 `claude-code`；`ParseSelectors()` 与 `ResolveTargets()` 共享该规则，因此聚合入口和资源子命令保持一致。

`linactl` 参数 key 解析统一转为小写并兼容横线/下划线，避免 `AGENT` 与 `agent` 在直接调用 `linactl` 时表现不一致。Makefile 包装层优先读取小写 `agent/action/force`，大写变量作为历史兼容入口保留。

## Compatibility

聚合命令仍沿用既有排序规则、link/unlink 行为、`force` 语义和软链状态机。子命令的详细输出保持不变，因此需要完整路径和冲突细节的高级用法不受影响。历史大写 Make 变量仍可使用，但文档和用法提示优先展示小写参数。

## Risks

聚合表将 `conflict`、`mismatch`、`native`、`absent` 等非写入结果归为 `skipped`，并在 `DETAIL` 中保留原始状态。这样会减少视觉噪声，但调用者如果需要完整逐项状态表，应继续使用资源子命令。
