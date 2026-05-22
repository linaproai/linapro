# agents-multi-resource 增量规范

## ADDED Requirements

### Requirement: 聚合命令展示必须保持简洁

系统 SHALL 保持`linactl agents`和`make agents`聚合入口的交互选择和执行结果简洁。Agent 选择阶段只展示用户需要识别的名称；执行完成后只展示按资源汇总的结果表。资源级路径、分类、来源和详细状态机信息 SHALL 保留在`agents.skills.*`、`agents.prompts.*`、`agents.md.*`子命令中，不在聚合入口重复展示。

#### Scenario: 聚合命令 Agent 选择列表只展示名称

- **WHEN** 开发者在终端下运行`make agents`或`linactl agents`
- **THEN** Agent 选择列表的每个候选项只展示该 Agent 的人类可读名称，例如`Claude Code`、`Codex`、`Cursor`
- **AND** 候选项不得拼接内部标识、资源分类、路径或状态摘要
- **AND** 单选提示说明为`Arrow keys to navigate, Enter to confirm, Esc to cancel, any key to search.`

#### Scenario: 聚合命令执行结果使用简化汇总表

- **WHEN** 开发者通过`make agents`或`linactl agents`聚合入口选择单个 Agent 并执行`link`或`unlink`
- **THEN** 命令输出一张按资源汇总的表格，至少包含`RESOURCE`、`STATUS`、`DETAIL`列
- **AND** 每个资源仅输出一行结果，展示该资源是`applied`、`skipped`或`failed`
- **AND** 命令不得重复输出`skills`、`prompts`、`md`三段完整资源明细表
- **AND** `agents.skills.*`、`agents.prompts.*`、`agents.md.*`子命令仍保留原有详细状态表输出

#### Scenario: Make 包装层不回显 linactl 委托命令

- **WHEN** 开发者运行`make agents`或`make agents.<resource>.<action>`系列目标
- **THEN** 终端输出应只包含`linactl`命令自身产生的业务结果
- **AND** Make 包装层不得额外回显`cd hack/tools/linactl && go run . ...`委托命令行

### Requirement: `agents.prompts`应管理 Agent 的 prompts 根目录软链

系统 SHALL 在`agents.prompts.link`和`agents.prompts.unlink`下管理每个 Agent 的 commands/prompts 根目录到仓库统一 prompts 根目录的软链。默认源目录 SHALL 为`.agents/prompts`；Agent 项目目标目录 SHALL 是该 Agent 用于承载多个 prompt catalog 的父级目录，例如 Claude Code 的`.claude/commands`。系统 SHALL 避免为单个 catalog 创建`.claude/commands/opsx`这类子目录软链。

#### Scenario: Claude Code prompts 使用父级 commands 目录软链

- **WHEN** 开发者运行`make agents.prompts.link agent=claude-code`
- **THEN** 命令在仓库根创建`.claude/commands`软链
- **AND** 软链目标为相对路径`../.agents/prompts`
- **AND** `.agents/prompts/opsx`通过`.claude/commands/opsx`自然可见
- **AND** 命令不得创建`.claude/commands/opsx`软链

#### Scenario: 已存在受管父级软链时视为正常

- **WHEN** 仓库中已存在`.claude/commands -> ../.agents/prompts`
- **THEN** 开发者运行`make agents.prompts.link agent=claude-code`时，该资源结果应为`ok`
- **AND** 命令不得因为`.claude/commands/opsx`可通过父级软链访问而报告`conflict`

### Requirement: Agent 选择器应支持名称归一化与小写参数

系统 SHALL 将命令行传入的 Agent 名称归一化为注册表使用的标准`kebab-case`标识。聚合入口和各资源子命令 SHALL 共享同一归一化规则。Make 入口 SHALL 优先支持小写参数`agent`、`action`和`force`，并继续兼容历史大写变量`AGENT`、`ACTION`和`FORCE`。

#### Scenario: 聚合命令接受 CamelCase Agent 名称

- **WHEN** 开发者运行`make agents agent=ClaudeCode`
- **THEN** 系统应将`ClaudeCode`归一化为`claude-code`
- **AND** 执行效果应与`make agents agent=claude-code`一致
- **AND** 输出中的 Agent 展示名仍应为`Claude Code`

#### Scenario: 选择器接受常见分隔形式

- **WHEN** 开发者在聚合命令或资源子命令中传入`Claude Code`、`claude_code`或`claude-code`
- **THEN** 系统应将这些值统一解析为`claude-code`
- **AND** 资源子命令的逗号列表也应对每个 Agent token 应用同一归一化规则

#### Scenario: Make 参数优先使用小写并兼容大写

- **WHEN** 开发者运行`make agents agent=claude-code action=unlink force=1`
- **THEN** Makefile 应向`linactl`传递`agent=claude-code action=unlink force=1`
- **AND** 开发者运行历史写法`make agents AGENT=claude-code ACTION=unlink FORCE=1`时仍应得到兼容行为
