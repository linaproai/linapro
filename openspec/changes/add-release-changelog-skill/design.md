## Context

`LinaPro`已经在`.agents/skills/`下维护项目自有技能，并通过`OpenSpec`记录技能能力契约。现有发布流程中，`create-release-tag.yml`只负责创建版本标签，`release-test-and-build.yml`在标签触发后创建`GitHub Release`，但本次需求明确要求现阶段不接入`CI`，先通过手动技能验证版本更新日志质量。

`temp/`目录已被仓库忽略，适合作为人工生成和审阅发布日志的临时输出位置。技能需要在没有`PR`关键标识、`label`或发布说明字段的情况下，综合`Git`历史、源码差异和`OpenSpec`内容整理详尽的双语`Markdown`发布日志。

## Goals / Non-Goals

**Goals:**

- 新增`lina-community-release-changelog`技能，用于手动生成`temp/changelog.md`。
- 支持默认比较范围和用户显式指定的两个`Git`引用比较范围。
- 固定输出用户确认的双语`Markdown`模板，英文在上、中文在下，中间保持模板分割线。
- 要求技能读取`Git`提交历史、源码差异和`OpenSpec`内容，形成详尽、可追溯、面向发布人员的变更说明。
- 明确在技能输出中记录比较范围，并保证范围方向从旧版本到新版本。

**Non-Goals:**

- 不修改`.github/workflows/create-release-tag.yml`、`.github/workflows/release-test-and-build.yml`或任何其他`CI`配置。
- 不创建或更新`GitHub Release`。
- 不要求`PR`新增关键标识、`label`、`Release-Note`字段或模板约束。
- 不新增后端、前端、数据库、插件运行时或`HTTP API`能力。
- 不新增长期维护脚本或`linactl`命令；如果后续需要确定性辅助工具，应另起变更设计跨平台`Go`实现。

## Decisions

### 手动技能优先于自动化发布集成

先将能力实现为`.agents/skills/lina-community-release-changelog/SKILL.md`，不接入`CI`。这样发布人员可以在不同版本区间上人工试运行、审阅和修订技能提示，待生成质量稳定后再设计自动发布集成。

替代方案是立即集成到`create-release-tag.yml`或`release-test-and-build.yml`。该方案会更快进入发布链路，但当前生成质量尚未验证，且`temp/changelog.md`跨工作流传递还需要额外设计，因此暂不采用。

### 比较范围由技能规范化

技能支持两类输入：

- 未指定范围时，默认使用最近可达发布标签到当前`HEAD`。
- 指定两个版本、标签、提交或分支时，技能先验证两个引用存在，再通过祖先关系或提交日期推断旧版本和新版本，最终以`<from>..<to>`形式输出。

如果两个引用无法安全排序或比较范围为空，技能必须停止并说明原因，避免生成方向错误或空洞的发布日志。

### 证据来源以仓库内容为准

技能不得要求`PR`标识作为输入。生成内容必须来自以下证据：

- `git log`、`git show`、`git diff --name-status`、`git diff --stat`等历史和差异信息。
- 变更范围内相关源码文件的抽样阅读，特别是`.agents/skills/`、`.github/`、`openspec/`、`hack/tools/`、`apps/lina-core/`、`apps/lina-vben/`和`apps/lina-plugins/`。
- `OpenSpec`的`proposal.md`、`design.md`、`tasks.md`和`specs/**/spec.md`，包括活跃变更和归档变更中落在比较范围内的内容。

提交信息可以用于定位线索，但不得作为唯一依据生成关键变更说明。对重要功能区，技能需要用源码或`OpenSpec`内容补足语义，降低遗漏和误判。

### 输出模板固定

技能必须写入`temp/changelog.md`，并使用以下章节：

- 英文：`Highlights`、`Improvements`、`Bug Fixes`、`Tooling and Experience`
- 中文：`主要亮点`、`功能改进`、`Bug 修复`、`开发体验与工具链`

不额外添加`Known Issues`、`OpenSpec and Governance`等章节，避免发布人员拿到的格式随执行者变化。若某章节没有证据，技能在该章节写明没有从现有证据识别到相关变更。

### 详尽性优先

技能生成的发布日志应面向发布人员和用户，不只是提交列表。每个章节可以使用多个`Markdown`列表项，重要条目需要说明变化内容和用户价值。技能应覆盖本次比较范围内的关键能力、治理、修复和工具链变化，不把大块功能压缩成一句泛泛描述。

## Risks / Trade-offs

- `Git`历史和`OpenSpec`内容语义不一致 → 技能优先采用源码和最终`OpenSpec`状态，并在无法确认时写保守描述。
- 提交数量很多导致上下文压力较大 → 技能先用`git diff --name-status/stat`分组，再按高影响目录抽样阅读，必要时分批整理章节。
- 没有`PR`标识会降低自动分类精度 → 技能通过提交类型、文件路径、`OpenSpec`任务和源码语义综合分类；不能确认的内容放入更保守的`Improvements`或`Tooling and Experience`。
- 双语内容可能信息不一致 → 技能先完成英文结构，再生成语义等价的中文内容，中文不得新增英文未覆盖的事实，英文也不得遗漏中文事实。
- 技能写入`temp/changelog.md`会覆盖已有临时输出 → 技能执行前应说明输出路径，并在需要时提醒用户先保留旧文件；`temp/`本身不参与版本控制。
