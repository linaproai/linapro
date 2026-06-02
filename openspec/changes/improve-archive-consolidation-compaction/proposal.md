## Why

随着迭代增加，`openspec/changes/archive/`中的归档内容会持续膨胀。现有`lina-openspec-archive-consolidate`已经能按功能领域做归档聚合，但对`tasks.md`中的反馈闭环、根因、验证证据、审查结论和低价值任务流水缺少明确的语义压缩策略，长期会导致归档文件越来越大且维护者难以快速提取关键历史。

## What Changes

- 增强`lina-openspec-archive-consolidate`技能，将归档聚合扩展为“语义聚合 + 高价值摘要压缩”流程。
- 明确`tasks.md`压缩规则：保留反馈、根因、验证、审查、治理影响和关键实现阶段，裁剪普通 checklist 流水。
- 明确`proposal.md`、`design.md`、`specs/`和`tasks.md`之间的归档信息承载边界，避免把所有历史过程无限复制到聚合结果中。
- 增加压缩安全门禁：只有在高价值信息已经迁移到聚合归档文档后，才允许清理原始日期前缀归档目录。
- 增加压缩报告要求，输出保留内容、裁剪内容、未压缩原因和验证结果，便于审查和后续维护。
- 不在本变更中实际压缩现有`openspec/changes/archive/`历史目录。

## Capabilities

### New Capabilities

- `archive-consolidation-compaction`：定义`lina-openspec-archive-consolidate`在归档聚合过程中对历史内容进行高价值摘要压缩、信息承载分层、安全裁剪和结果验证的能力边界。

### Modified Capabilities

- `monthly-openspec-archive`：月度归档聚合使用的共享提示词和工具执行结果需要遵循增强后的归档摘要压缩门禁，避免自动化流程产生无界增长的聚合归档。

## Impact

- 影响`.agents/skills/lina-openspec-archive-consolidate/SKILL.md`的技能说明、输入模式、语义压缩规则、清理门禁和输出报告。
- 可能影响`.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`，使月度自动归档调用同一套压缩语义。
- 影响 OpenSpec 归档治理文档，不涉及运行时代码、HTTP API、数据库、缓存、前端 UI 或用户可见运行时文案。
- 验证方式以`openspec validate improve-archive-consolidation-compaction --strict`、技能说明静态检查和示例归档压缩策略审查为主。
