## MODIFIED Requirements

### Requirement: 归档残留治理
系统 SHALL 在归档失败或中断后识别并治理半成品主规范更新，避免后续归档因为重复新增或残留文件而阻塞。系统 SHALL 允许通过受控 monthly 自动化流程执行已完成变更的归档治理，但该流程必须在创建或更新归档 PR 前执行 OpenSpec 校验并保护变更范围。

#### Scenario: 存在半成品主规范文件
- **WHEN** 归档过程异常中断并遗留半成品主规范文件
- **THEN** 系统或维护流程能够识别该残留
- **AND** 在下一次归档前完成清理或对齐，避免产生重复 capability 写入

#### Scenario: Monthly 自动归档治理
- **WHEN** monthly OpenSpec 归档流程自动归档已完成变更
- **THEN** 流程在创建或更新归档 PR 前执行 OpenSpec 校验
- **AND** 流程拒绝将归档治理允许范围外的文件变更写入 PR
