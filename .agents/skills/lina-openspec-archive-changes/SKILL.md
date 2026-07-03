---
name: lina-openspec-archive-changes
description: >-
  扫描并归档 LinaPro 仓库中已经完成的所有 OpenSpec 活跃变更；若任务已全部完成但存在可判定的归档异常，先自动修复并复验，再继续归档。
  必须用户手动触发，禁止自动触发该技能。
compatibility: 依赖 OpenSpec CLI，要求在 LinaPro 仓库根目录执行。
---

# Lina 自动归档

扫描 `openspec/changes/` 根目录下的活跃 OpenSpec 变更，将已经完成的变更执行归档，并清晰展示本次处理结果。

## 适用场景

该技能需要手动调用触发。

## 核心原则

1. **只处理活跃变更**：只扫描 `openspec/changes/` 根目录下的一级子目录，并明确排除 `openspec/changes/archive/`。
2. **只自动归档已完成变更**：变更必须满足任务全部完成，且在自动修复可判定异常后通过 OpenSpec 完成态与校验复验，才允许归档。
3. **未完成的一律跳过**：只要存在未完成任务、状态为 `in-progress`、状态无法读取且不能通过可判定异常修复恢复、任务文件无法安全判定或归档命令失败，就不要移动该变更目录。
4. **不要人工放行未完成项**：本技能是自动归档流程，不对未完成变更发起“是否强制归档”的确认，也不要使用 `--no-validate` 绕过校验。
5. **保留可追溯结果**：结束时必须列出成功归档的变更，以及未归档变更和具体原因。
6. **尊重现有工作区**：执行前查看 `git status --short`，了解是否已有用户改动。不要还原、整理或修改与归档无关的文件。
7. **先全量预检与修复再归档**：在执行任何 `openspec archive` 前，必须先完成全部候选变更的完成态、任务清单、增量规范 header 匹配和插件规范目录预检。若任务已全部完成但存在可机械判定的异常，先自动修复、复验，再放入可归档列表；不可安全修复的只跳过该变更并记录原因，不阻塞其他已完成变更继续归档。
8. **插件规范目录必须带插件前缀**：涉及用户在`apps/lina-plugins/<plugin-id>/`下开发的具体业务插件能力时，归档前必须确认该变更的`specs/`能力目录以`<plugin-id>`开头；若能唯一识别插件 ID 和目标目录，应自动重命名修复，避免归档后把插件规范写入主框架通用能力目录。
9. **修复后必须复验**：任何自动修复都必须重新运行该变更的 OpenSpec 校验和状态检查；复验失败时不得归档该变更，也不得使用 `--no-validate` 或手动移动目录绕过。

## 插件相关 OpenSpec 命名要求

“插件相关内容”指用户在`apps/lina-plugins/<plugin-id>/`下开发的具体业务插件功能及其 OpenSpec 规范，例如`john-content-cms`插件提供的 CMS 功能。不包括`apps/lina-core`中的插件宿主框架、插件生命周期、插件治理、`pluginbridge`、host service 或源码/动态插件运行时等主框架通用能力。

归档预检时必须按以下规则处理插件相关内容：

- 优先从`proposal.md`、`design.md`、`tasks.md`、增量规范正文、涉及路径或`plugin.yaml`中的插件`id`识别`<plugin-id>`。
- 插件相关变更中的每个`openspec/changes/<change-name>/specs/<capability>/`目录必须以`<plugin-id>`开头。允许直接使用`<plugin-id>`作为插件根能力目录，也允许使用`<plugin-id>-<capability>`承载插件内更细的能力；禁止使用`cms`、`article-management`、`settings`这类不带插件前缀的通用目录名。
- 同一个变更涉及多个插件时，必须分别为每个插件使用对应的`<plugin-id>`前缀；不得把多个插件的业务规范合并进同一个无插件前缀的能力目录。
- 归档后写入`openspec/specs/`的插件规范目录必须仍然以`<plugin-id>`开头，例如`openspec/specs/john-content-cms/spec.md`或`openspec/specs/john-content-cms-article/spec.md`。
- 主框架的插件基础设施规范仍使用主框架能力名，例如`plugin-framework`、`plugin-governance`或`plugin-host-service-extension`，不要强行加某个业务插件 ID。
- 若完成态候选变更违反插件前缀规则，且能唯一识别`<plugin-id>`、当前`specs/<capability>/`不含插件前缀、目标目录不存在或可确认属于同一变更，则自动将目录重命名为`specs/<plugin-id>-<capability>/`或`specs/<plugin-id>/`，再重新检查 header 目标和 OpenSpec 校验。
- 若插件 ID 无法唯一识别、同一能力可能归属多个插件、目标目录存在冲突，或重命名会覆盖用户已有改动，记录跳过原因`插件规范目录缺少插件前缀且无法安全自动修复：<capability>`；只跳过该变更，不停止其他变更归档。

## 完成判定

一个变更只有在以下条件全部成立时，才视为可自动归档：

1. 变更目录位于 `openspec/changes/<change-name>/`，且 `<change-name>` 不是 `archive`。
2. `openspec status --change "<change-name>" --json` 在自动修复后能正常执行。
3. `openspec status` 返回的所有 artifact 都处于完成状态。常见完成值包括 `done`、`complete`、`completed`；若返回结构不包含 artifact 明细，则使用 `openspec list --json` 中该变更的 `status` 字段辅助判断，只有 `complete`、`completed`、`done` 视为完成。若状态异常仅由可修复的增量规范问题导致，应先修复再复验，而不是直接阻塞全局归档。
4. `tasks.md` 中不存在未完成任务标记：
   - `- [ ]`
   - `- [未完成]`（若项目中出现中文任务标记）
5. 若 `openspec list --json` 提供 `completedTasks` 和 `totalTasks`，优先要求二者相等；若二者不等但`tasks.md`逐项扫描确认无未完成任务，记录`CLI 任务统计与 tasks.md 不一致`，并在该变更通过严格校验后继续归档。
6. 变更中的所有 `MODIFIED` 和 `REMOVED` requirement header 必须能在当前 `openspec/specs/<capability>/spec.md` 中找到。缺失时进入自动修复流程；修复后仍缺失时说明 `OpenSpec header 不匹配且无法安全自动修复：<capability>/<requirement>`，并跳过该变更。
7. 变更涉及具体业务插件能力时，所有插件相关`specs/<capability>/`目录必须满足“插件相关 OpenSpec 命名要求”，确保归档目标为`openspec/specs/<plugin-id>.../`而不是主框架通用目录。缺少插件前缀时优先按本技能的自动修复规则处理。

如果 `tasks.md` 不存在，应将该变更标记为“无法判定任务完成情况”，并跳过自动归档。LinaPro 的 OpenSpec 变更应维护任务清单，缺失任务文件不适合自动处理。

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
- `pluginIds`：识别到的业务插件 ID 列表（如有）
- `pluginSpecNaming`：插件相关`specs/`目录是否满足`<plugin-id>`前缀要求
- `repairableAnomalies`：任务已完成但可尝试自动修复的异常列表
- `repairActions`：已执行的自动修复动作和涉及文件
- `postRepairValidation`：修复后的校验、状态检查结果
- `skipReason`：若不可归档，记录明确原因

推荐记录写法：

- `任务未完成：79/113`
- `tasks.md 中仍有 34 个未完成任务`
- `OpenSpec 状态未完成：in-progress`
- `OpenSpec 状态读取失败`
- `缺少 tasks.md，无法自动判定任务完成情况`
- `artifact 未完成：proposal, design`
- `插件规范目录缺少插件前缀：cms 应以 john-content-cms 开头`
- `已自动修复插件规范目录：cms → john-content-cms-cms`
- `已自动修复 OpenSpec header：article-management/文章管理 MODIFIED → ADDED`
- `CLI 任务统计与 tasks.md 不一致，已按 tasks.md 复验继续归档`
- `OpenSpec header 不匹配且无法安全自动修复：<capability>/<requirement>`
- `归档失败：<错误摘要>`

### 4. 归档前全量预检与自动修复

执行任何归档命令前，必须先检查全部候选变更，并形成以下列表：

- `ready`：已完成且无需修复的变更。
- `repair-required`：`tasks.md`已全部完成，但存在可机械判定异常的变更。
- `skipped`：任务未完成、无法判定完成态或异常无法安全修复的变更。

预检内容包括：

- `openspec status --change "<change-name>" --json` 能读取且所有 artifact 完成。
- `tasks.md` 存在且未包含未完成任务标记。
- `openspec list --json` 的 `completedTasks` 与 `totalTasks` 一致（如 CLI 提供）。
- `specs/**/spec.md` 中所有 `MODIFIED`/`REMOVED` requirement header 都存在于当前主规范对应 capability 的 `openspec/specs/<capability>/spec.md`。
- 插件相关增量规范的`specs/<capability>/`目录均以对应`<plugin-id>`开头；检查`MODIFIED`/`REMOVED` header 时也只能使用带插件前缀的`openspec/specs/<capability>/spec.md`作为目标，不得用历史遗留的无前缀插件目录放行。

当`tasks.md`已全部完成但预检失败时，按以下顺序自动修复可判定异常：

1. **插件规范目录缺少前缀**：若能唯一识别插件 ID，且目标目录不会覆盖无关文件，将`specs/<capability>/`重命名为`specs/<plugin-id>-<capability>/`；若`<capability>`已经等于插件根能力语义，也可以重命名为`specs/<plugin-id>/`。修复后重新按新 capability 检查 header。
2. **`MODIFIED` header 找不到基线 requirement**：先检查是否是 capability 重命名、大小写、空格或标点导致的精确匹配问题；能确定目标 requirement 时，修正增量规范中的 capability 或 header 文本。若目标规范中确实不存在该 requirement，且该变更块表达的是新增能力、当前目标规范不存在同名 requirement，则将该块从 `MODIFIED` 调整为 `ADDED`。
3. **`REMOVED` header 找不到基线 requirement**：若目标规范已经不存在该 requirement，且删除块没有迁移说明以外的有效新增语义，将该 `REMOVED` 块视为已完成的空操作并移除；若无法判断是否误删或目标 capability 错误，则跳过该变更。
4. **CLI 任务统计与`tasks.md`不一致**：当`tasks.md`扫描确认无未完成任务，且该变更严格校验通过时，记录差异并继续归档，不因 CLI 统计差异阻塞。

自动修复必须满足以下安全条件：

- 不修改`tasks.md`来伪造完成态，不把未完成任务改成完成。
- 不覆盖用户已有未提交内容；若目标文件已有用户改动且无法精确合并，跳过该变更。
- 不为了通过归档删除有效需求内容；只允许做 capability 命名、header 精确匹配、`MODIFIED`转`ADDED`或无效`REMOVED`空操作清理这类可解释修复。
- 每个变更最多进行两轮修复与复验；超过后仍失败，记录为不可安全自动修复。

每个修复后的变更必须运行：

```bash
openspec validate "<change-name>" --strict
openspec status --change "<change-name>" --json
```

只有复验通过且完成态仍成立的变更才能进入归档列表。某个已完成变更修复失败时，只将该变更放入`skipped`并报告原因；除非出现 OpenSpec CLI 缺失、仓库根目录错误、`openspec/changes`缺失或全局命令无法执行这类环境级问题，否则不得停止其他已完成变更归档。

### 5. 执行归档

仅对通过完成判定和修复后复验的变更执行：

```bash
openspec archive -y "<change-name>"
```

说明：

- 使用 `-y` 跳过交互确认，因为本技能已经完成自动检查。
- 不使用 `--no-validate`。
- 不默认使用 `--skip-specs`。只有当 OpenSpec CLI 明确提示该变更不需要同步 specs，或用户事先要求跳过 specs 时，才可以使用 `--skip-specs`。
- 每次归档后重新确认目标变更已不在 `openspec/changes/<change-name>/`，并记录归档路径。常见归档路径为 `openspec/changes/archive/YYYY-MM-DD-<change-name>/`。

如果归档命令失败，记录为未归档，并保留错误摘要；不要手动 `mv` 目录来绕过 OpenSpec CLI。单个变更归档失败不应阻塞后续已通过预检和复验的变更继续归档，除非失败原因表明 OpenSpec CLI 或仓库状态整体不可用。

### 6. 输出结果报告

执行完成后，用中文输出结构化结果。必须包含：

1. 扫描到的活跃变更总数
2. 自动修复的变更列表、修复动作和复验结果
3. 成功归档的变更列表
4. 未归档的变更列表和原因
5. 若没有任何可归档变更，明确说明“本次没有归档任何变更”

推荐格式：

```markdown
**自动归档结果**

扫描到 3 个活跃变更，自动修复 1 个，成功归档 2 个，跳过 1 个。

自动修复：
- `change-a`：`specs/cms/` → `specs/john-content-cms-cms/`，复验通过

成功归档：
- `change-a` → `openspec/changes/archive/2026-05-12-change-a/`
- `change-b` → `openspec/changes/archive/2026-05-12-change-b/`

未归档：
- `change-c`：任务未完成：5/8
```

若发生环境错误：

```markdown
无法执行自动归档：未找到 OpenSpec CLI。
请先安装或修复 `openspec` 命令后再运行 `lina-openspec-archive-changes`。
```

## 边界情况

- **没有活跃变更**：报告“未发现可处理的活跃变更”，不报错。
- **只有未完成变更**：报告每个未归档原因，不执行归档。
- **目录已在 archive 下**：不纳入扫描，也不在未归档列表中展示。
- **CLI 状态与 tasks.md 不一致**：若`tasks.md`存在未完成项，以未完成为准并跳过；若`tasks.md`确认全部完成且严格校验通过，记录统计差异并继续归档。
- **已完成变更存在可修复异常**：自动修复、复验并继续归档，不因单个候选的可修复异常阻塞整体流程。
- **已完成变更存在不可修复异常**：只跳过该变更并报告原因；继续处理其他已完成且可安全归档的变更。
- **归档后产生工作区变更**：这是 OpenSpec 归档的预期结果，报告归档路径即可；不要自动提交。
- **存在用户未提交改动**：可以继续自动归档，但不要改动无关文件；若用户改动位于同一个待归档变更目录且该变更满足完成条件，先提醒该目录有本地改动，再继续依赖 OpenSpec CLI 归档。

## 验证建议

归档结束后，根据实际情况运行轻量验证：

```bash
openspec list --json
openspec validate --all
```

若 `openspec validate --all` 因仓库中既有未完成变更失败，不要把失败归咎于本次归档；在报告中说明失败范围和相关变更名。
