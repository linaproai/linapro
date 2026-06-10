---
name: lina-community-release-changelog
description: >-
  手动生成 LinaPro 版本更新日志。用户要求生成 changelog、release notes、版本更新日志、发布说明，或提到
  lina-community-release-changelog 时必须使用本技能。技能会基于 Git 历史、源码差异和 OpenSpec 内容整理详尽的双语
  Markdown 更新日志，涉及数据库变更时必须单独列出，支持默认比较范围和用户指定两个版本/标签/提交进行比较，固定写入
  localdocs/changelog.md。
---

# Lina Release Changelog

手动生成 `LinaPro` 版本更新日志。当前阶段仅用于人工执行和质量验证，不接入 `CI`、`GitHub Actions` 或自动发布流程。

## 核心原则

1. **只手动执行**：不得自动创建 `GitHub Release`、推送标签、提交文件或修改任何 `.github/workflows/` 文件。
2. **固定输出路径**：生成结果必须写入仓库根目录 `localdocs/changelog.md`。
3. **固定模板**：英文内容在上，中文内容在下，中间使用模板分割线；不得新增模板外章节。
4. **证据优先**：必须基于 `Git` 历史、源码差异和 `OpenSpec` 内容整理，不要求也不等待 `PR` 标识、`label` 或发布说明字段。
5. **详尽覆盖**：输出面向发布人员和用户，不是提交摘要；必须覆盖比较范围内关键功能、修复、数据库结构或初始化数据变化，以及工具链体验变化。
6. **双语一致**：英文和中文分别完整成文，事实覆盖一致，不交叉混写。
7. **表达自然**：英文和中文都必须地道、清晰、流利、易懂；按目标语言重新组织句子，不做生硬直译，不使用少见、拗口或容易误解的词语。
8. **尊重工作区**：执行前查看 `git status --short`。除 `localdocs/changelog.md` 外，不修改无关文件；不要还原用户已有改动。

## 输入范围

支持两种用法：

| 用法 | 行为 |
| --- | --- |
| 未指定范围 | 使用当前 `HEAD` 作为目标引用，选择目标引用可达且早于目标引用的最近发布标签作为起点 |
| 指定两个引用 | 比较用户给出的两个版本、标签、提交或分支，规范化为旧引用到新引用的 `<from>..<to>` |

用户可能使用以下表达：

- `生成当前版本更新日志`
- `生成 v0.1.0 到 v0.2.0 的 changelog`
- `比较 v0.2.0 和 v0.1.0`
- `from=v0.1.0 to=v0.2.0`
- `v0.1.0..v0.2.0`

## 执行流程

### 1. 确认环境

在仓库根目录执行只读检查：

```bash
pwd
git status --short
git tag --list
```

如果当前目录不是 `LinaPro` 仓库根目录，或 `git` 不可用，停止并说明原因。

### 2. 解析比较范围

#### 默认范围

未指定范围时：

1. 将 `to` 设为 `HEAD`。
2. 从 `git tag --merged HEAD` 中找出可达发布标签，发布标签通常匹配 `vMAJOR.MINOR.PATCH` 或 `vMAJOR.MINOR.PATCH-prerelease`。
3. 选择早于 `to` 的最近发布标签作为 `from`。如果 `HEAD` 正好带有发布标签，不要把同一个标签作为 `from`；应选择它之前的最近可达发布标签。
4. 如果找不到可用 `from`，停止并说明无法确定默认比较范围，要求用户显式指定两个引用。

#### 显式双引用范围

用户指定两个引用时：

1. 使用 `git rev-parse --verify "<ref>^{commit}"` 验证两个引用都存在。
2. 如果用户明确提供 `from=<ref>` 和 `to=<ref>`，优先尊重该方向，但必须确认该方向可验证：
   - `from` 是 `to` 的祖先时，方向有效。
   - 两者都是语义化发布标签且 `from` 的版本号早于 `to` 时，方向有效，即使仓库历史经过压缩导致二者不是线性祖先关系。
   - 如果 `from` 的版本号晚于 `to`，停止并说明用户给出的方向与版本顺序相反，建议改用无方向的比较表达或交换参数。
   - 如果既不能通过祖先关系验证，也不能通过语义化发布标签顺序验证，停止并说明范围不安全。
3. 如果用户只说“比较 A 和 B”，则通过祖先关系规范化方向：
   - `A` 是 `B` 的祖先时，使用 `A..B`。
   - `B` 是 `A` 的祖先时，使用 `B..A`。
   - 两者互不为祖先时，只有在二者都是可比较的语义化发布标签且顺序明确时才按版本号排序；否则停止并说明无法安全判断方向。
4. 使用 `git rev-list --count <from>..<to>` 确认范围非空；如果为 `0`，停止并说明没有可生成的变更。

#### 标题版本

标题中的版本使用以下优先级：

1. 如果 `to` 是发布标签，使用该标签，例如 `v0.2.0`。
2. 如果 `to` 是 `HEAD` 且当前 `HEAD` 有发布标签，使用该标签。
3. 否则读取 `apps/lina-core/manifest/config/metadata.yaml` 中的 `framework.version`。
4. 如果仍无法确定，使用 `Unreleased`。

### 3. 收集证据

不要只读取 `git log --oneline`。至少收集以下信息：

```bash
git log --first-parent --decorate --date=short --format="%h %ad %s" <from>..<to>
git log --decorate --date=short --format="%h %ad %s" <from>..<to>
git diff --stat <from>..<to>
git diff --name-status <from>..<to>
```

然后按变更路径分组，重点阅读这些目录中的关键文件或差异：

- `.agents/skills/`
- `.github/`
- `openspec/`
- `hack/tools/`
- `hack/makefiles/`
- `apps/lina-core/manifest/sql/`
- `apps/lina-core/`
- `apps/lina-vben/`
- `apps/lina-plugins/*/manifest/sql/`
- `apps/lina-plugins/`
- `README.md` 和 `README.zh-CN.md`

对于重要提交，使用 `git show --stat <commit>` 和必要的源码片段确认真实行为。提交标题只能作为线索，不能作为关键内容的唯一证据。

#### 插件 submodule 证据

`apps/lina-plugins/`是父仓库中的`submodule`时，父仓库的`git diff`只能显示`gitlink`指针变化，不能展示插件仓库内部提交和文件差异。只要`git diff --name-status <from>..<to>`包含`apps/lina-plugins`，或`git ls-files -s apps/lina-plugins`显示该路径的`mode`为`160000`，必须解析`submodule`两端`commit`并进入插件仓库收集证据：

```bash
git ls-tree <from> apps/lina-plugins
git ls-tree <to> apps/lina-plugins
git -C apps/lina-plugins rev-parse --verify "<old-plugin-commit>^{commit}"
git -C apps/lina-plugins rev-parse --verify "<new-plugin-commit>^{commit}"
git -C apps/lina-plugins log --decorate --date=short --format="%h %ad %s" <old-plugin-commit>..<new-plugin-commit>
git -C apps/lina-plugins diff --stat <old-plugin-commit>..<new-plugin-commit>
git -C apps/lina-plugins diff --name-status <old-plugin-commit>..<new-plugin-commit>
```

如果本地`submodule`缺少比较端`commit`，先执行只更新插件仓库`Git`对象的读取性拉取，再重试验证：

```bash
git -C apps/lina-plugins fetch --tags --prune origin
```

如果拉取失败或`commit`仍不可用，不得断言插件无变化；必须在证据来源摘要和保守处理说明中记录无法确认的`submodule`范围。如果`old-plugin-commit`与`new-plugin-commit`相同，记录父仓库比较范围内没有已提交的插件指针变化；`git -C apps/lina-plugins status --short`只能作为工作区状态提示，不能把未提交的插件工作区改动写入发布变更。

不要把父仓库中的`M apps/lina-plugins`或`gitlink`指针本身写成发布功能。插件发布说明必须来自插件仓库内部提交、文件差异、`OpenSpec`或源码事实。

### 4. 读取 OpenSpec 语义

如果比较范围涉及 `openspec/`，优先读取相关文件：

- `proposal.md`
- `design.md`
- `tasks.md`
- `specs/**/spec.md`

包括活跃变更和 `openspec/changes/archive/` 下落入比较范围的归档变更。用 `OpenSpec` 内容提炼功能语义、治理目标、验收范围和用户可见价值。若 `Git` 历史和 `OpenSpec` 表述存在差异，以源码和最终 `OpenSpec` 状态为准，并使用保守描述。

如果`apps/lina-plugins/`的`submodule`内部差异涉及插件内容，必须按`apps/lina-plugins/<plugin-id>/`分组梳理语义，重点确认插件清单、前后端入口、菜单或路由、权限标签、`pluginbridge`路由声明、`hostServices`声明、安装或卸载`SQL`、资源产物、生命周期扫描、宿主能力调用和插件间能力调用边界。每个插件只写对发布读者有价值的用户可见能力、行为修复、运行时契约或开发体验变化；纯内部重排只有在影响发布风险、使用方式或维护方式时才写入。

### 5. 识别数据库变更

必须单独判断比较范围是否涉及数据库变更，不能只把它写进功能描述。出现以下任一证据时，应进入`Database Changes`/`数据库变更`章节：

- 新增、修改或删除`apps/lina-core/manifest/sql/**`或`apps/lina-plugins/<plugin-id>/manifest/sql/**`。
- 新增、修改或删除安装、卸载、迁移、初始化、Seed、Mock 数据等`SQL`文件。
- 新增、修改或删除由表结构变化引起的`DAO`、`DO`、`Entity`、模型字段或索引相关代码。
- `OpenSpec`、提交记录或源码差异明确说明新增表、删除表、字段变更、索引变更、软删除字段、时间字段、初始化数据或插件安装卸载数据变更。

写入数据库章节时，必须说明对发布或升级人员有用的信息：受影响模块或插件、表名或资源名、变更类型（新增表、字段调整、索引调整、初始化数据、Mock 数据、安装/卸载`SQL`等）、对应路径或证据来源，以及是否意味着用户除了更新代码外还需要更新数据库结构或重新执行初始化/迁移脚本。

如果某个功能条目同时带来数据库结构变化，可以在`Highlights`或`Improvements`中描述用户价值，同时仍必须在`Database Changes`/`数据库变更`中单独列出升级影响。数据库章节用于识别发布操作风险，不受“不要重复堆叠”的限制。

### 6. 分类规则

将证据归入固定章节：

| 章节 | 收录内容 |
| --- | --- |
| `Highlights` / `主要亮点` | 本次范围最重要、最值得发布人员优先说明的能力或架构变化 |
| `Improvements` / `功能改进` | 功能增强、产品能力补充、运行时行为改进、治理能力增强 |
| `Bug Fixes` / `Bug 修复` | 明确修复问题的提交、反馈修复、回归修复、测试修复 |
| `Database Changes`/`数据库变更` | 表结构、索引、迁移、初始化数据、Seed/Mock 数据、插件安装或卸载`SQL`变化，以及由这些变化引起的升级操作要求 |
| `Tooling and Experience` / `开发体验与工具链` | `CI`、构建、发布、`OpenSpec`治理、技能、开发命令、测试效率、文档维护体验 |

如果某条变化同时属于多个非数据库章节，放入对发布读者最有价值的章节，不要重复堆叠。治理、规范和`OpenSpec`流程变化默认放入`Tooling and Experience`，除非它也是主要发布亮点。涉及数据库的升级影响必须额外进入数据库章节。

### 7. 生成 Markdown

必须使用以下模板：

```markdown
## Highlights

## Improvements

## Bug Fixes

## Database Changes

## Tooling and Experience

---

## 主要亮点

## 功能改进

## Bug 修复

## 数据库变更

## 开发体验与工具链

```

正文写法：

- 使用 `Markdown` 列表。
- 每个重要条目以加粗短标题开头，随后说明具体变化和发布价值。
- 重要功能不要压缩成一句泛泛描述；必要时一个条目可以包含多句。
- 不要在英文正文中写中文，不要在中文正文中写英文句子；路径、命令、标识符、产品名可以保留原文。
- 英文部分和中文部分必须覆盖同一组事实。中文不能新增英文没有的事实，英文也不能遗漏中文事实。
- 英文必须使用自然的发布说明表达，优先使用常见动词和短句，避免中式英语、罕见词、复杂从句和不必要的抽象名词。
- 中文必须使用自然的产品更新表达，优先选择常用、明确、读者容易理解的说法，避免逐词翻译英文术语。
- 遇到容易产生生硬译法的技术词时，按上下文翻译含义。例如，不要把软件语境中的 `seam` 写成“接缝”，可改为“扩展点”“衔接点”或直接描述具体边界；不要把测试语境中的 `fixture` 写成“夹具”，可改为“测试数据”“测试准备逻辑”“测试基线”或更具体的事实描述。
- `Database Changes`/`数据库变更`章节如果有内容，应优先写清楚表结构或初始化数据的升级影响，不要只写“更新了`SQL`文件”。
- 如果某个章节没有证据，写入：
  - 英文：`No changes identified from the available evidence.`
  - 中文：`根据现有证据未识别到相关变更。`
- 如果数据库章节没有证据，优先使用更明确的表述：
  - 英文：`No database schema or seed data changes identified from the available evidence.`
  - 中文：`根据现有证据未识别到数据库结构或初始化数据变更。`

### 示例条目

英文：

```markdown
- **Release metadata command**: Added a version update command that changes `framework.version` and refreshes README image cache keys together, reducing manual release preparation errors.
```

中文：

```markdown
- **发布元数据命令**：新增版本更新命令，可同步修改`framework.version`并刷新`README`图片缓存参数，降低发布准备过程中的人工遗漏风险。
```

### 写入文件

确保 `localdocs/` 目录存在，然后将完整内容写入：

```text
localdocs/changelog.md
```

`localdocs/` 已被 `.gitignore` 忽略。不要把生成的 `localdocs/changelog.md` 添加到版本控制。

### 8. 自检

写入后必须重新读取 `localdocs/changelog.md`，检查：

1. 只包含固定模板章节。
2. 中间分割线存在。
3. 中文标题和来源范围存在。
4. 英文和中文事实覆盖一致。
5. 英文和中文表达地道、清晰、流利、易懂，没有生硬直译、罕见词或拗口表达。
6. 没有把软件语境中的 `seam`、`fixture` 等词机械翻译成“接缝”“夹具”等不自然说法。
7. 每个章节要么有证据支持的内容，要么写明未识别到相关变更。
8. 若证据涉及`manifest/sql/`、安装/卸载`SQL`、Seed/Mock 数据、表结构相关`DAO`/模型字段或`OpenSpec`数据库语义，`Database Changes`/`数据库变更`已单独列出受影响表或资源、变更类型、证据路径和升级操作影响；若没有数据库证据，数据库章节已明确写明未识别到数据库结构或初始化数据变更。
9. 若`apps/lina-plugins/`是`submodule`且指针变化，已包含插件仓库内部`old/new commit`、`log`、`diff`证据和按插件 ID 的语义总结；若无法获取历史，已明确保守处理。
10. `.github/workflows/` 没有被修改。
11. 没有执行提交、推送、打标签或创建 `GitHub Release`。

结束时用中文汇报：

- 规范化比较范围。
- 标题版本。
- 输出文件路径。
- 证据来源摘要。
- 数据库变更摘要，包括是否涉及表结构、索引、初始化/Mock 数据或插件安装卸载`SQL`。
- 插件`submodule`处理情况。
- 是否存在无法确认而被保守处理的内容。
