# 改进 make agents 展示内容

## Why

`make agents` 当前的交互选择项把内部标识、展示名、三类资源分类和软链状态全部拼在同一行，例如 `claude-code (Claude Code) — skills: link[+], prompts: link[!], md: link[+]`。这些信息对调试有帮助，但对日常选择流程过重，用户只需要先选择要配置的 Agent。

聚合命令执行完成后还会分别输出 `skills`、`prompts`、`md` 三段资源明细表，再追加一段文字 Summary。对一次只配置一个 Agent 的聚合入口而言，这些路径、分类和来源列信息重复且占屏，冲突或成功结果也不够聚焦。

## What Changes

- 简化 `make agents` / `linactl agents` 的 TTY Agent 选择列表：只展示 Agent 的人类可读名称，例如 `Claude Code`、`Codex`、`Cursor`。
- 将单选提示说明改为 `Arrow keys to navigate, Enter to confirm, Esc to cancel, any key to search.`。
- 聚合命令不再委托三个子命令分别渲染完整资源明细表，而是直接执行对应资源的 apply API，并输出一张按资源汇总的简化表格。
- 保留 `agents.skills.*`、`agents.prompts.*`、`agents.md.*` 子命令原有详细表格输出，避免破坏需要排查具体路径的高级入口。
- 修正 `agents.prompts` 的 Claude Code 映射：使用 `.claude/commands -> ../.agents/prompts` 这类父级目录软链，而不是 `.claude/commands/opsx -> .agents/prompts/opsx` 子目录软链。
- 对命令行 Agent 名称增加归一化处理，使 `agent=ClaudeCode`、`agent=Claude Code`、`agent=claude_code` 与 `agent=claude-code` 等价。
- Make 参数统一推荐小写 `agent/action/force`，继续兼容历史大写变量 `AGENT/ACTION/FORCE`。

## Impact

- 影响 `hack/tools/linactl` 的 `agents` 聚合命令展示层、单选 TTY helper 文案、Agent 选择器解析和 Make 参数包装。
- 不改变 `skills`、`md` 资源注册表、软链状态机、文件系统写入规则或子命令详细输出；`prompts` 资源注册表调整为父级目录软链。
- 不涉及后端运行时服务、REST API、前端页面、运行时 i18n 资源、业务缓存、数据权限或 E2E 页面流程。
