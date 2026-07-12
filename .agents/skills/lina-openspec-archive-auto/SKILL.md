---
name: lina-openspec-archive-auto
description: >-
  先执行 lina-openspec-archive-changes 归档活跃变更，再执行 lina-openspec-archive-consolidate 做归档摘要。
  必须用户手动触发，禁止自动触发。
---

# Lina OpenSpec 归档自动化

按顺序串联两个技能：

1. 读取并执行 `.agents/skills/lina-openspec-archive-changes/SKILL.md`
2. 读取并执行 `.agents/skills/lina-openspec-archive-consolidate/SKILL.md`

不复制子技能细则；门禁、修复、聚合与清理规则以各自 `SKILL.md` 为准。

## 硬规则

1. **仅手动触发** — 禁止被其他技能、CI、钩子或模糊意图自动调用。
2. **先归档，后聚合** — 不得并行；阶段一环境故障则停止。
3. 部分跳过或无可归档项时仍继续阶段二。
4. 待聚合集合为空时跳过阶段二并说明。
5. 不自动 commit / push / 开 PR。

## 报告

结束后汇总两阶段结果：归档成功/跳过清单，以及聚合分组、清理与语义覆盖结论。
