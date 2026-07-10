---
name: lina-openspec-archive-changes
description: >-
  扫描并归档已完成的 OpenSpec 活跃变更。
  必须用户手动触发，禁止自动触发。
compatibility: 依赖 OpenSpec CLI，在 LinaPro 仓库根目录执行。
---

# Lina 自动归档

归档 `openspec/changes/` 下已完成的活跃变更（排除 `archive/`），输出成功/跳过清单。

## 硬规则

1. 只扫活跃一级目录，不扫 `archive/`。
2. 任务未完成、校验失败或不可安全修复 → 跳过，不强制归档，不用 `--no-validate`，不手动 `mv`。
3. **先全量预检/修复，再归档**；单个失败不阻塞其他可归档项（环境级故障除外）。
4. 不改无关工作区文件；不伪造 `tasks.md` 完成态；不空写 `design.md`。
5. 修复后必须 `openspec validate <name> --strict` + `openspec status --change <name> --json` 复验。
6. 结束时列出：修复项、成功归档路径、跳过原因。

## 完成门禁（全部满足才可归档）

| 条件 | 要求 |
|---|---|
| 位置 | `openspec/changes/<name>/`，`name ≠ archive` |
| status 可读 | `openspec status --change <name> --json` 成功 |
| 必选 artifact | `proposal` / `specs` / `tasks`（若存在）为 `done`/`complete`/`completed` |
| **design（可选）** | **缺失 `design.md`、或 `design=ready`、或已 done → 均不阻塞**。`isComplete=false` 仅因 design 未写 → 仍可继续。禁止为归档生成占位 design |
| tasks.md | 存在；无 `- [ ]` / `- [未完成]` |
| 任务统计 | `completedTasks == totalTasks`；不等则以 tasks.md 为准，记录差异后继续 |
| MODIFIED/REMOVED | header 须命中 `openspec/specs/<capability>/spec.md`；否则走修复 |
| 插件业务 specs | 能力目录必须以 `<plugin-id>` 开头（见下） |
| 校验 | `openspec validate <name> --strict` 通过 |

无 `tasks.md` → 跳过（无法判定）。缺 `design.md` → **不跳过**。

无 artifact 明细时，用 `openspec list --json` 的 `status ∈ {complete,completed,done}` 辅助判断。

## 插件业务规范前缀

**业务插件** = `apps/lina-plugins/<plugin-id>/` 下的具体插件能力。  
**不含** 宿主框架 / lifecycle / governance / pluginbridge / host service / 动态运行时。

- 从 proposal/design/tasks/specs/路径/`plugin.yaml` 识别 `plugin-id`。
- 插件相关 `specs/<capability>/` 必须以 `<plugin-id>` 或 `<plugin-id>-...` 命名；归档后 `openspec/specs/` 同理。
- 多插件分目录，不混用无前缀通用名（如 `cms`）。
- 主框架能力保持主框架名（如 `plugin-framework`）。
- 可安全重命名时：`specs/<cap>/` → `specs/<plugin-id>-<cap>/` 或 `specs/<plugin-id>/`，再复验。
- 无法唯一识别 / 冲突 → 跳过：`插件规范目录缺少插件前缀且无法安全自动修复：<capability>`。

## 流程

### 1. 环境

```bash
pwd && test -d openspec/changes && openspec --version && git status --short
```

CLI 不可用、非仓库根或无 `openspec/changes` → 停止并说明。

### 2. 候选

```bash
openspec list --json
find openspec/changes -mindepth 1 -maxdepth 1 -type d ! -name archive -exec basename {} \; | sort
```

CLI + 文件系统合并去重，字母序处理。

### 3. 逐项检查

```bash
openspec status --change "<name>" --json
# 读 openspec/changes/<name>/tasks.md；扫 specs/**
```

记录：任务完成度、artifact、插件 id、可修复异常、skipReason。

常用原因：`任务未完成：a/b` · `缺少 tasks.md` · `artifact 未完成：proposal/specs/tasks` · `design 可选，已继续` · `header 不匹配…` · `插件前缀…` · `归档失败：…`

### 4. 预检分类

- **ready**：门禁全过
- **repair-required**：任务完成但有可修异常
- **skipped**：未完成 / 不可修

**可自动修复（最多 2 轮，每轮后 strict 复验）：**

1. 插件 specs 缺前缀 → 安全重命名  
2. MODIFIED header 不匹配 → 对齐主规范文本；确属新增且主规范无同名 → 改 `ADDED`  
3. REMOVED 主规范已无该条且无有效新增语义 → 删除空 REMOVED 块  
4. CLI 任务数 ≠ tasks.md 但 tasks 已全部完成 + strict 通过 → 记录差异并继续  
5. design 缺失/ready → **不修复**，直接可选放行  

禁止：改 tasks 勾选、覆盖用户未提交冲突文件、删除有效需求以混过校验。

### 5. 归档

```bash
openspec archive -y "<name>"
```

- 用 `-y`；默认不用 `--skip-specs` / `--no-validate`  
- 确认目录已迁至 `openspec/changes/archive/YYYY-MM-DD-<name>/`  
- 单条失败记原因，继续下一条  

### 6. 报告

```markdown
**自动归档结果**

扫描到 N 个活跃变更，自动修复 A 个，成功归档 B 个，跳过 C 个。

自动修复：
- `name`：动作摘要，复验通过

成功归档：
- `name` → `openspec/changes/archive/YYYY-MM-DD-name/`

未归档：
- `name`：原因
```

无任何可归档项时写明「本次没有归档任何变更」。环境错误示例：未找到 OpenSpec CLI。

### 7. 轻量验证（可选）

```bash
openspec list --json
openspec validate --all
```

`validate --all` 因其他未完成变更失败时，勿归咎本次归档，注明范围即可。

## 边界速查

| 情况 | 处理 |
|---|---|
| 无活跃变更 | 报告即可，不报错 |
| 仅未完成 | 列原因，不归档 |
| 已在 archive/ | 不扫描 |
| 缺 design / design=ready | **归档**，记「design 可选」 |
| 工作区已有本地改动 | 可继续；不动无关文件；待归档目录有改动时先提示 |
| 归档产生 diff | 预期结果，不自动 commit |
