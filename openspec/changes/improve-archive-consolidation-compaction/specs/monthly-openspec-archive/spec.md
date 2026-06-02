## MODIFIED Requirements

### Requirement: Monthly workflow must consolidate only after new archive changes

系统 SHALL 仅在本次 monthly 工具运行时自动归档产生 OpenSpec 文件变更后执行`lina-openspec-archive-consolidate`技能，避免无新增归档时重复重写聚合归档文档。月度归档聚合 SHALL 使用`lina-openspec-archive-consolidate`定义的高价值摘要压缩、语义覆盖门禁和原始归档清理规则；当无人值守流程无法确认聚合输出已覆盖输入归档的高价值语义时，workflow MUST 失败并停止后续 PR 写回。

#### Scenario: Archive produced changes

- **WHEN** tool-specific auto archive 执行后`openspec/`下存在新的文件变更
- **THEN** workflow 调用`lina-openspec-archive-consolidate`聚合已归档变更
- **AND** workflow 使用技能定义的摘要压缩规则保留背景、设计、规范、反馈、验证和审查治理证据
- **AND** workflow 只有在语义覆盖门禁通过后才允许清理本次参与聚合的日期前缀原始归档目录
- **AND** workflow 在聚合后执行临时变更清理检查和 OpenSpec 校验
- **AND** workflow stops before PR finalization if archive consolidation, semantic coverage validation, temporary change cleanup, or OpenSpec validation fails

#### Scenario: Archive produced no changes

- **WHEN** tool-specific auto archive 执行完成
- **AND** `openspec/`下没有新的文件变更
- **THEN** workflow 跳过`lina-openspec-archive-consolidate`
- **AND** workflow 不创建或更新归档 PR

### Requirement: Monthly workflow must share prompt files across AI tools

系统 SHALL 将 monthly OpenSpec 自动归档和归档聚合提示词维护为`.github/prompts/`下的公共文件，并 SHALL 让所有工具专属 reusable workflow 引用同一份自动归档提示词和同一份聚合提示词内容。共享归档聚合提示词 SHALL 明确要求执行`lina-openspec-archive-consolidate`的高价值摘要压缩和语义覆盖门禁；不得在工具专属 workflow 中维护绕过压缩门禁的重复提示词正文。

#### Scenario: Shared archive consolidate prompt

- **WHEN** 任一工具专属 reusable workflow 执行`lina-openspec-archive-consolidate`
- **THEN** workflow 从`.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`读取提示词
- **AND** 共享提示词要求遵守`lina-openspec-archive-consolidate`技能中的摘要压缩、原始归档保护和失败优先规则
- **AND** workflow 不在工具专属 workflow 中内联维护重复的归档聚合提示词正文

#### Scenario: Shared auto archive prompt

- **WHEN** 任一工具专属 reusable workflow 执行 base auto archive
- **THEN** workflow 从`.github/prompts/monthly-openspec-auto-archive.zh-CN.md`读取提示词
- **AND** workflow 通过当前已选择的 Codex、Claude Code 或 GitHub Copilot CLI 运行时执行该提示词
