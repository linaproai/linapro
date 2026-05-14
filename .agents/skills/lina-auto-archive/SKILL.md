---
name: lina-auto-archive
description: >-
  自动扫描并归档 LinaPro 仓库中已经完成的 OpenSpec 活跃变更。用户要求自动归档、批量归档、清理已完成变更、
  扫描 openspec/changes 中可归档项目，或提到 lina-auto-archive 时必须使用本技能。技能会跳过仍有未完成任务、
  OpenSpec 状态未完成、已归档目录或无法安全判定的变更，并在结束后用中文汇总成功归档列表与未归档原因。
compatibility: 依赖 OpenSpec CLI，要求在 LinaPro 仓库根目录执行。
---

# Lina 自动归档

自动扫描 `openspec/changes/` 根目录下的活跃 OpenSpec 变更，将已经完成的变更执行归档，并清晰展示本次处理结果。

## 适用场景

该技能需要手动调用触发。

## 核心原则

1. **只处理活跃变更**：只扫描 `openspec/changes/` 根目录下的一级子目录，并明确排除 `openspec/changes/archive/`。
2. **只自动归档已完成变更**：变更必须同时满足 OpenSpec 状态完成、任务全部完成，才允许归档。
3. **未完成的一律跳过**：只要存在未完成任务、状态为 `in-progress`、状态无法读取、任务文件无法安全判定或归档命令失败，就不要移动该变更目录。
4. **不要人工放行未完成项**：本技能是自动归档流程，不对未完成变更发起“是否强制归档”的确认，也不要使用 `--no-validate` 绕过校验。
5. **保留可追溯结果**：结束时必须列出成功归档的变更，以及未归档变更和具体原因。
6. **尊重现有工作区**：执行前查看 `git status --short`，了解是否已有用户改动。不要还原、整理或修改与归档无关的文件。

---

## 完成判定

一个变更只有在以下条件全部成立时，才视为可自动归档：

1. 变更目录位于 `openspec/changes/<change-name>/`，且 `<change-name>` 不是 `archive`。
2. `openspec status --change "<change-name>" --json` 能正常执行。
3. `openspec status` 返回的所有 artifact 都处于完成状态。常见完成值包括 `done`、`complete`、`completed`；若返回结构不包含 artifact 明细，则使用 `openspec list --json` 中该变更的 `status` 字段辅助判断，只有 `complete`、`completed`、`done` 视为完成。
4. `tasks.md` 中不存在未完成任务标记：
   - `- [ ]`
   - `- [未完成]`（若项目中出现中文任务标记）
5. 若 `openspec list --json` 提供 `completedTasks` 和 `totalTasks`，则二者必须相等。

如果 `tasks.md` 不存在，应将该变更标记为“无法判定任务完成情况”，并跳过自动归档。LinaPro 的 OpenSpec 变更应维护任务清单，缺失任务文件不适合自动处理。

---

## 执行流程

### 1. 确认环境

在仓库根目录执行：

```bash
pwd
test -d openspec/changes
openspec --version
git status --short
```

若 `openspec` 不可用、当前目录不是仓库根目录，或 `openspec/changes` 不存在，停止执行并说明原因。

### 2. 收集候选变更

优先使用 OpenSpec CLI 获取状态摘要：

```bash
openspec list --json
```

同时用文件系统扫描作为兜底，确保不会漏掉 CLI 未列出的活跃目录：

```bash
find openspec/changes -mindepth 1 -maxdepth 1 -type d ! -name archive -exec basename {} \; | sort
```

合并两边得到的变更名，去重后按字母序处理。不要扫描 `openspec/changes/archive/` 内部目录。

### 3. 逐个检查完成状态

对每个候选变更执行：

```bash
openspec status --change "<change-name>" --json
```

并读取：

```text
openspec/changes/<change-name>/tasks.md
```

记录以下信息：

- `changeName`：变更名
- `listStatus`：`openspec list --json` 中的状态（如有）
- `statusArtifacts`：`openspec status --json` 中 artifact 的完成情况（如有）
- `completedTasks` / `totalTasks`：任务统计（如 CLI 提供）
- `uncheckedTasks`：从 `tasks.md` 扫描到的未完成任务数量
- `skipReason`：若不可归档，记录明确原因

推荐跳过原因写法：

- `任务未完成：79/113`
- `tasks.md 中仍有 34 个未完成任务`
- `OpenSpec 状态未完成：in-progress`
- `OpenSpec 状态读取失败`
- `缺少 tasks.md，无法自动判定任务完成情况`
- `artifact 未完成：proposal, design`
- `归档失败：<错误摘要>`

### 4. 执行归档

仅对通过完成判定的变更执行：

```bash
openspec archive -y "<change-name>"
```

说明：

- 使用 `-y` 跳过交互确认，因为本技能已经完成自动检查。
- 不使用 `--no-validate`。
- 不默认使用 `--skip-specs`。只有当 OpenSpec CLI 明确提示该变更不需要同步 specs，或用户事先要求跳过 specs 时，才可以使用 `--skip-specs`。
- 每次归档后重新确认目标变更已不在 `openspec/changes/<change-name>/`，并记录归档路径。常见归档路径为 `openspec/changes/archive/YYYY-MM-DD-<change-name>/`。

如果归档命令失败，记录为未归档，并保留错误摘要；不要手动 `mv` 目录来绕过 OpenSpec CLI。

### 5. 输出结果报告

执行完成后，用中文输出结构化结果。必须包含：

1. 扫描到的活跃变更总数
2. 成功归档的变更列表
3. 未归档的变更列表和原因
4. 若没有任何可归档变更，明确说明“本次没有归档任何变更”

推荐格式：

```markdown
**自动归档结果**

扫描到 3 个活跃变更，成功归档 1 个，跳过 2 个。

成功归档：
- `change-a` → `openspec/changes/archive/2026-05-12-change-a/`

未归档：
- `change-b`：任务未完成：5/8
- `change-c`：OpenSpec 状态未完成：in-progress
```

若发生环境错误：

```markdown
无法执行自动归档：未找到 OpenSpec CLI。
请先安装或修复 `openspec` 命令后再运行 `lina-auto-archive`。
```

---

## 边界情况

- **没有活跃变更**：报告“未发现可处理的活跃变更”，不报错。
- **只有未完成变更**：报告每个未归档原因，不执行归档。
- **目录已在 archive 下**：不纳入扫描，也不在未归档列表中展示。
- **CLI 状态与 tasks.md 不一致**：以更保守的结果为准；任一来源显示未完成即跳过，并说明状态不一致。
- **归档后产生工作区变更**：这是 OpenSpec 归档的预期结果，报告归档路径即可；不要自动提交。
- **存在用户未提交改动**：可以继续自动归档，但不要改动无关文件；若用户改动位于同一个待归档变更目录且该变更满足完成条件，先提醒该目录有本地改动，再继续依赖 OpenSpec CLI 归档。

---

## 验证建议

归档结束后，根据实际情况运行轻量验证：

```bash
openspec list --json
openspec validate --all
```

若 `openspec validate --all` 因仓库中既有未完成变更失败，不要把失败归咎于本次归档；在报告中说明失败范围和相关变更名。
