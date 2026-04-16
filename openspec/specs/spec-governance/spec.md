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
