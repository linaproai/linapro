这是无人值守的 nightly CI 归档聚合任务。请使用仓库内 `.agents/skills/lina-archive-consolidate/SKILL.md` 的规则执行归档聚合：
1. 未指定变更列表时，只处理 `openspec/changes/archive/` 下目录名以 `YYYY-MM-DD-` 开头的原始归档变更。
2. 不要把既有非日期前缀聚合目录再次作为默认输入。
3. 生成或更新聚合归档文档时保持变更范围在 `openspec/**` 内。
4. 不要等待人工交互；如果无法在 CI 中安全完成，请失败并说明原因。
5. 运行结束后输出中文结果摘要。
