## ADDED Requirements

### Requirement: 技能必须手动生成版本更新日志

`lina-community-release-changelog`技能 SHALL 作为手动触发的仓库级技能，用于从当前`LinaPro`仓库生成版本更新日志，并 SHALL 将结果写入项目根目录`temp/changelog.md`。该技能 SHALL NOT 自动创建`GitHub Release`、推送标签、提交文件或修改任何`CI`工作流。

#### Scenario: 手动执行生成临时更新日志

- **WHEN** 用户手动调用`lina-community-release-changelog`
- **THEN** 技能在仓库根目录生成或覆盖`temp/changelog.md`
- **AND** 生成内容使用`Markdown`格式
- **AND** 技能不修改`.github/workflows/`下任何文件
- **AND** 技能不执行`git commit`、`git push`、`git tag`或`gh release create`

#### Scenario: 技能声明不接入自动化发布

- **WHEN** `.agents/skills/lina-community-release-changelog/SKILL.md`被读取
- **THEN** 技能说明其当前阶段仅用于手动执行
- **AND** 技能说明稳定前不接入`CI`或`GitHub Actions`
- **AND** 技能说明输出路径为`temp/changelog.md`

### Requirement: 技能必须支持默认和显式比较范围

`lina-community-release-changelog`技能 SHALL 支持在用户未指定范围时生成最近可达发布标签到当前`HEAD`的更新日志，并 SHALL 支持用户显式指定两个`Git`引用来生成历史区间更新日志。技能 SHALL 将最终比较范围规范化为旧引用到新引用的`<from>..<to>`形式。

#### Scenario: 未指定范围时使用最近发布标签到 HEAD

- **WHEN** 用户调用技能且未提供比较范围
- **THEN** 技能定位当前`HEAD`可达的最近发布标签作为`from`
- **AND** 技能使用当前`HEAD`作为`to`
- **AND** 技能在`temp/changelog.md`中写入规范化的来源范围

#### Scenario: 用户指定两个版本标签

- **WHEN** 用户要求比较`v0.1.0`和`v0.2.0`之间的历史更新日志
- **THEN** 技能验证两个`Git`引用均存在
- **AND** 技能将较旧引用规范化为`from`
- **AND** 技能将较新引用规范化为`to`
- **AND** 技能使用`v0.1.0..v0.2.0`作为来源范围生成`temp/changelog.md`

#### Scenario: 用户输入顺序与时间方向相反

- **WHEN** 用户要求比较`v0.2.0`和`v0.1.0`
- **THEN** 技能仍然按旧版本到新版本输出
- **AND** 技能在来源范围中写入`v0.1.0..v0.2.0`

#### Scenario: 无法安全判断比较方向

- **WHEN** 用户指定的两个引用都存在但技能无法通过祖先关系或可验证证据判断旧引用和新引用
- **THEN** 技能停止生成
- **AND** 技能说明无法安全判断比较范围方向
- **AND** 技能不写入误导性的`temp/changelog.md`

### Requirement: 技能必须基于仓库证据整理内容

`lina-community-release-changelog`技能 SHALL 根据`Git`历史记录、源码差异和`OpenSpec`内容整理版本更新日志。技能 MUST NOT 要求`PR`中新增关键标识、发布说明字段、`label`或其他额外流程数据。

#### Scenario: 收集 Git 历史和源码差异

- **WHEN** 技能已经确定比较范围
- **THEN** 技能读取该范围内的提交历史
- **AND** 技能读取该范围内的文件变更清单和差异统计
- **AND** 技能按受影响目录识别关键变更区域
- **AND** 技能对关键源码或配置变更进行必要抽样阅读

#### Scenario: 收集 OpenSpec 语义证据

- **WHEN** 比较范围内存在`OpenSpec`相关文件变更
- **THEN** 技能读取相关`proposal.md`、`design.md`、`tasks.md`和`specs/**/spec.md`
- **AND** 技能优先使用`OpenSpec`内容提炼功能语义、治理目标和验收范围
- **AND** 技能不得仅凭提交标题替代`OpenSpec`内容分析

#### Scenario: 不要求 PR 标识

- **WHEN** 比较范围内的提交没有`PR`编号、`label`或发布说明字段
- **THEN** 技能仍然通过`Git`历史、源码差异和`OpenSpec`内容生成更新日志
- **AND** 技能不要求用户补充`PR`关键标识后才能继续

### Requirement: 技能必须生成固定双语 Markdown 模板

`lina-community-release-changelog`技能 SHALL 使用固定`Markdown`模板生成更新日志，英文内容位于上半部分，中文内容位于下半部分，中间使用模板中的分割线。技能 SHALL NOT 在固定模板外新增其他章节。

#### Scenario: 生成英文和中文固定章节

- **WHEN** 技能写入`temp/changelog.md`
- **THEN** 文件以`# LinaPro <to> Release Notes`开头
- **AND** 英文部分包含`Source range: <from>..<to>`、`Highlights`、`Improvements`、`Bug Fixes`和`Tooling and Experience`章节
- **AND** 中文部分包含`# LinaPro <to> 更新日志`、`来源范围：<from>..<to>`、`主要亮点`、`功能改进`、`Bug 修复`和`开发体验与工具链`章节
- **AND** 英文部分和中文部分之间存在模板要求的分割线

#### Scenario: 不新增额外章节

- **WHEN** 技能生成更新日志
- **THEN** 技能不新增`Known Issues`、`OpenSpec and Governance`或其他模板外章节
- **AND** 技能将治理、规范或工具相关变化归入`Tooling and Experience`或`开发体验与工具链`

#### Scenario: 章节无证据时写明无相关变更

- **WHEN** 某个固定章节没有从当前比较范围证据中识别到相关变更
- **THEN** 英文章节写入`No changes identified from the available evidence.`
- **AND** 对应中文章节写入`根据现有证据未识别到相关变更。`

### Requirement: 技能必须保障内容详尽性和双语一致性

`lina-community-release-changelog`技能 SHALL 生成详尽的发布日志，覆盖比较范围内的关键内容，不得只输出提交摘要。英文和中文内容 SHALL 分别完整成文且语义一致，不得中英文交叉混写。

#### Scenario: 关键变更不被遗漏

- **WHEN** 比较范围内存在多个功能、修复、治理或工具链变更
- **THEN** 技能按章节覆盖所有关键变更类别
- **AND** 每个关键条目说明变化内容和发布价值
- **AND** 技能不把多个高影响功能压缩为单个泛泛条目

#### Scenario: 双语内容保持语义一致

- **WHEN** 技能生成英文部分和中文部分
- **THEN** 中文部分覆盖英文部分中的全部事实
- **AND** 英文部分覆盖中文部分中的全部事实
- **AND** 英文章节不包含中文正文
- **AND** 中文章节不包含英文正文，除非内容是路径、命令、标识符或产品名

#### Scenario: 内容必须保持证据约束

- **WHEN** 技能无法从`Git`、源码或`OpenSpec`证据确认某项变化
- **THEN** 技能不得把该变化写成确定事实
- **AND** 技能可以省略该变化或使用明确的保守表述
