## ADDED Requirements

### Requirement: OpenSpec 文档必须分层承载当前契约与历史信息

OpenSpec 文档治理 SHALL 将`openspec/specs`作为当前能力契约的唯一事实来源，将`openspec/changes/<active-change>`作为本次增量变更承载位置，并将`openspec/changes/archive`作为历史摘要、设计演进、反馈闭环和验证证据承载位置。

#### Scenario: 读取当前能力契约

- **WHEN** 维护者或 AI 需要理解某能力当前应满足的需求
- **THEN** 首先读取`openspec/specs/<capability>/spec.md`
- **AND** 不把`openspec/changes/archive/**/specs/<capability>/spec.md`中的重复历史副本作为当前契约事实来源

#### Scenario: 读取历史设计原因

- **WHEN** 维护者或 AI 需要理解某能力为何形成当前设计
- **THEN** 读取该能力归档 owner 分组中的`proposal.md`、`design.md`、`tasks.md`和必要历史规范摘要
- **AND** 归档内容提供背景、决策、演进、反馈和验证证据，而不是重复保存所有当前最终契约全文

### Requirement: 归档能力规范必须具备唯一 owner

OpenSpec 归档治理 SHALL 为跨分组重复出现的能力规范建立唯一归档 owner。每个能力最多由一个`openspec/changes/archive/<domain>`分组长期保存该能力的历史规范摘要或无法迁移的历史约束。

#### Scenario: 能力跨多个归档分组重复出现

- **WHEN** 静态扫描发现同一`specs/<capability>/spec.md`出现在多个归档分组中
- **THEN** 压缩任务为该能力选择一个主要交付物归属分组作为归档 owner
- **AND** owner 判定依据记录在任务记录或压缩报告中

#### Scenario: 归档 owner 保留能力历史

- **WHEN** 某归档分组被判定为能力 owner
- **THEN** 该分组可以保留对应能力的历史规范摘要、关键设计演进或无法安全迁移的验收约束
- **AND** 当前最终契约仍以`openspec/specs/<capability>/spec.md`为准

### Requirement: 非 owner 归档分组不得长期保存重复能力全文规范

OpenSpec 归档治理 SHALL 将非 owner 分组中的重复能力全文规范迁移为交叉影响摘要，避免多个归档分组长期重复保存同一能力的完整 requirement 与 scenario。

#### Scenario: 非 owner 分组存在重复能力规范

- **WHEN** 压缩任务确认某归档分组不是`<capability>`的归档 owner
- **THEN** 该分组不得长期保留`specs/<capability>/spec.md`完整全文
- **AND** 若该分组对该能力存在历史影响，则在该分组`design.md`中保留交叉影响摘要
- **AND** 摘要必须指向当前契约位置和历史 owner 分组

#### Scenario: 归档规范与主规范完全相同

- **WHEN** 静态校验发现归档`spec.md`与`openspec/specs/<capability>/spec.md`内容完全相同
- **THEN** 该归档副本默认视为可删除重复全文
- **AND** 如需保留，必须在任务记录中说明该副本承载的额外历史价值

### Requirement: 归档压缩必须先执行样板分组验证

OpenSpec 归档治理 SHALL 在批量压缩高体量归档分组前，先选择一个高体量或高重复度分组作为样板，验证压缩规则、owner 映射和语义覆盖记录可行。

#### Scenario: 开始批量压缩前

- **WHEN** 实施者准备压缩多个既有归档分组
- **THEN** 先选择`plugin-framework`、`user-auth`或其他高体量高重复度分组作为样板
- **AND** 完整处理样板分组的`proposal.md`、`design.md`、`tasks.md`和`specs/`
- **AND** 在样板验证通过前不得批量删除其他分组的重复规范

#### Scenario: 样板分组验证通过

- **WHEN** 样板分组完成压缩
- **THEN** OpenSpec 校验、重复能力扫描、Markdown 格式检查和语义覆盖审查均通过
- **AND** 后续批量压缩必须复用样板中验证通过的规则

### Requirement: 归档文件压缩必须按文件职责裁剪

OpenSpec 归档治理 SHALL 按`proposal.md`、`design.md`、`tasks.md`和`specs/`的不同职责裁剪历史信息，避免同一事实在多个归档文件中重复保存。

#### Scenario: 压缩 proposal

- **WHEN** 压缩归档`proposal.md`
- **THEN** 只保留背景、目标、范围和影响
- **AND** 删除普通实施过程、重复能力列表和已由`design.md`或`tasks.md`承载的细节

#### Scenario: 压缩 design

- **WHEN** 压缩归档`design.md`
- **THEN** 保留架构决策、方案演进、废弃方案原因、关键约束和交叉影响摘要
- **AND** 删除已经由主规范或代码事实承载的低价值执行流水

#### Scenario: 压缩 tasks

- **WHEN** 压缩归档`tasks.md`
- **THEN** 保留`FB-*`、根因、修复、验证、审查和治理影响的最小维护摘要
- **AND** 裁剪普通 checklist、重复命令、逐文件记录和已被其他文件覆盖的执行流水

### Requirement: 归档压缩结果必须输出量化验证报告

OpenSpec 归档治理 SHALL 在压缩任务完成时输出量化验证报告，说明压缩前后体量、规范数量、重复能力数量、语义覆盖和未压缩原因。

#### Scenario: 输出压缩验证报告

- **WHEN** 完成样板分组或批量归档压缩
- **THEN** 报告包含压缩前后`openspec/changes/archive`体量、归档 spec 文件数量、跨分组重复能力数量和完全重复主规范副本数量
- **AND** 报告列出保留的高价值信息类别、裁剪的低价值信息类别、未压缩或未删除原因
- **AND** 报告记录`openspec validate --all`或对应严格校验结果

#### Scenario: 无法确认语义覆盖

- **WHEN** 实施者无法确认某段归档内容已被`proposal.md`、`design.md`、`tasks.md`、owner 历史摘要或主规范覆盖
- **THEN** 不得删除该内容的唯一副本
- **AND** 必须在报告中记录阻断原因和后续处理建议
