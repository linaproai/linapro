## ADDED Requirements

### Requirement: 新建变更文档语言跟随用户上下文

系统 SHALL 在新建 active change 并首次生成 `proposal.md`、`design.md`、`tasks.md` 及 `specs/` 下增量规范时，跟随用户输入的上下文语言编写内容。

#### Scenario: 用户以中文上下文创建新迭代

- **WHEN** 用户使用中文描述需求，或明确要求以中文生成新迭代文档
- **THEN** 生成的 `proposal.md`、`design.md`、`tasks.md` 与增量规范使用简体中文编写
- **AND** 标题、段落、场景与任务项等内容保持中文表达

#### Scenario: 用户以英文上下文创建新迭代

- **WHEN** 用户使用英文描述需求，或明确要求以英文生成新迭代文档
- **THEN** 生成的 `proposal.md`、`design.md`、`tasks.md` 与增量规范使用英文编写
- **AND** 同一 active change 内的新增文档内容保持一致语言，除非用户明确要求整体改写语言

### Requirement: 归档文档统一使用英文

系统 SHALL 在归档变更时使用英文编写归档目录中的文档内容，并确保同步到 `openspec/specs/` 的主规范也使用英文，而不是跟随当前对话上下文语言。

#### Scenario: 中文上下文下执行归档

- **WHEN** 开发者在中文对话上下文中归档一个变更
- **THEN** 归档后的 `proposal.md`、`design.md`、`tasks.md` 与增量规范使用英文
- **AND** 同步到 `openspec/specs/` 的主规范内容也使用英文

#### Scenario: 归档产物用于国际化协作

- **WHEN** 一个变更完成交付并进入长期归档与社区协作阶段
- **THEN** 归档文档默认以英文作为统一语言
- **AND** 后续国际化支持和社区贡献者可以直接基于归档英文文档协作
