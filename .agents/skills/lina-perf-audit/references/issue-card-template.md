# Issue Card Template

Use this template for persistent cards under `perf-issues/`. One file equals one audit finding, including performance issues and read-request side-effect violations. Cards are cross-run assets and must not be placed under `temp/`.

Card body descriptions must be written in Chinese. Keep API paths, SQL excerpts,
Trace IDs, fingerprints, frontmatter field names, and status enum values as
machine-readable originals.

## File Name

```text
perf-issues/<severity>-<module>-<slug>.md
```

Examples:

- `perf-issues/HIGH-user-n-plus-1-list.md`
- `perf-issues/HIGH-usermsg-read-write-side-effect.md`
- `perf-issues/MEDIUM-dict-unbounded-list.md`
- `perf-issues/LOW-notice-app-filter.md`

## Template

```markdown
---
id: <severity>-<module>-<slug>
severity: <HIGH|MEDIUM|LOW>
module: <module>
endpoint: <METHOD PATH>
status: open
first_seen_run: <run-id>
last_seen_run: <run-id>
seen_count: 1
fingerprint: <sha256>
---

# <severity> - <module> - <中文问题标题>

## 问题描述

<用中文描述审查观察到的性能问题或读接口写入副作用,说明影响范围、风险原因以及受影响接口。>

## 复现方式

1. `make init confirm=init rebuild=true && make mock confirm=mock`
2. `bash .agents/skills/lina-perf-audit/scripts/setup-audit-env.sh --run-id <run-id>`
3. `bash .agents/skills/lina-perf-audit/scripts/prepare-builtin-plugins.sh --run-dir temp/lina-perf-audit/<run-id>`
4. `bash .agents/skills/lina-perf-audit/scripts/stress-fixture.sh --run-dir temp/lina-perf-audit/<run-id>`
5. `curl -i -H "Authorization: Bearer <token>" "<base-url><path>"`
6. 从响应头获取 `Trace-ID`，并检查 `temp/lina-perf-audit/<run-id>/server.log` 与 `backend-nohup.log` 中同一请求链路的 SQL。
7. 预期可复现现象：<SQL 数量、耗时或读接口写入副作用模式，例如"SQL 数量随返回行数增长"或"GET trace 包含 UPDATE sys_notify_delivery">。

## 证据

- Trace-ID：`<trace-id or fallback marker>`
- 审计文件：`temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`
- 源码位置：`<relative-source-file>:<line>`
- SQL 总数：`<count>`
- 写入 SQL 数：`<count when applicable>`

```sql
<short SQL excerpts only>
```

## 改进方案

1. <中文描述具体修复步骤,例如使用 WHERE IN 批量加载、增加分页、增加索引或把 GET 写入拆成显式动作。>
2. <中文描述验证步骤,例如复跑 lina-perf-audit 并确认 SQL 数量保持稳定或写入 SQL 数为 0。>

## 历史记录

- <run-id>：本次审查发现，审计文件 `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`，SQL 总数 `<count>`。
```

## Status Rules

- `open`：问题仍待处理，尚未进入修复。
- `in-progress`：后续 OpenSpec 变更正在修复该问题。
- `fixed`：后续审计或人工验证确认问题已修复。
- `obsolete`：接口、模块或代码路径已不存在。

当 `status: fixed` 或 `status: obsolete` 的问题再次被观察到时，将状态改回 `open`，并在「历史记录」中追加"被再次观察到（回归）"条目。

## Content Rules

- 「复现方式」必须能在干净本地环境独立执行。
- 证据必须使用仓库相对路径和简短 SQL 片段。
- 「改进方案」必须描述后续修复工作；审计 skill 本身不修改生产代码。
- 不得依赖原始 run 中临时存在的 shell 变量。
- 持久卡片不得使用本机绝对路径。
