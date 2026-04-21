# spec-governance Specification

## Purpose
规范 OpenSpec 主规范的结构、归档残留治理与归档前校验要求，保障规格资产可持续维护。
## Requirements
### Requirement: 主规范结构统一
系统 SHALL 将 `openspec/specs/` 下的主规范统一为当前 OpenSpec schema 要求的标准结构，至少包含 `## Purpose` 和 `## Requirements` 两个核心章节。

#### Scenario: 校验主规范结构
- **WHEN** 开发者对任一主规范执行 OpenSpec 校验或查看命令
- **THEN** 该主规范使用当前 schema 可识别的章节结构
- **AND** 不因缺少 `Purpose`、`Requirements` 等必需章节而失败

### Requirement: 归档残留治理
系统 SHALL 在归档失败或中断后识别并治理半成品主规范更新，避免后续归档因为重复新增或残留文件而阻塞。

#### Scenario: 存在半成品主规范文件
- **WHEN** 归档过程异常中断并遗留半成品主规范文件
- **THEN** 系统或维护流程能够识别该残留
- **AND** 在下一次归档前完成清理或对齐，避免产生重复 capability 写入

### Requirement: 归档前主规范可验证
系统 SHALL 在执行归档前确保将被更新的主规范通过当前 OpenSpec schema 的基础验证。

#### Scenario: 执行变更归档
- **WHEN** 开发者归档一个会更新主规范的变更
- **THEN** 被影响的主规范均能通过结构和 requirement 级校验
- **AND** 归档流程不会因为历史主规范格式不兼容而中断

### Requirement: New Active Change Artifacts Follow User Language
The system SHALL generate new active change artifacts in the user's current request language unless the user explicitly requests another language.

#### Scenario: Chinese request context creates a new active change
- **WHEN** the user requests a new active change primarily in Simplified Chinese
- **THEN** the generated proposal, design, tasks, and delta specs use Simplified Chinese

#### Scenario: English request context creates a new active change
- **WHEN** the user requests a new active change primarily in English
- **THEN** the generated proposal, design, tasks, and delta specs use English

### Requirement: Archived Change Documents Use English
The system SHALL archive change documents and archived delta specs in English, regardless of the current conversation language.

#### Scenario: Archive is executed from a Chinese conversation
- **WHEN** a completed change is archived from a Chinese conversation context
- **THEN** the archived proposal, design, tasks, and archived delta specs are written in English
- **AND** any synced baseline spec updates introduced by the archive are written in English

### Requirement: Active Change Status Is Determined By Archive State
The system SHALL treat every unarchived change directory under `openspec/changes/` as an active change, regardless of whether its implementation tasks are already complete.

#### Scenario: Completed but unarchived change still receives feedback
- **WHEN** a change directory still exists under `openspec/changes/` and has not been moved into `openspec/changes/archive/`
- **AND** the change reports all tasks completed or `openspec list --json` shows `status: complete`
- **THEN** the workflow still treats that change as active
- **AND** new feedback MUST be appended to that existing change instead of creating a new change
