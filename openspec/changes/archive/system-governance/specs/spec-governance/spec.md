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

#### Scenario: Deterministic 归档与 AI 聚合分离
- **WHEN** monthly OpenSpec 归档流程执行基础归档
- **THEN** 流程使用确定性 OpenSpec CLI 命令直接归档，不依赖 AI 工具执行基础归档步骤
- **AND** AI Coding 工具仅用于确定性归档产生变更后的可选归档聚合增强

#### Scenario: 归档阶段化失败
- **WHEN** monthly OpenSpec 归档流程的确定性归档阶段存在无法归档的 completed active change
- **THEN** 流程先为成功的归档结果创建或更新归档 PR
- **AND** 流程在所有候选处理完成后失败退出并输出剩余变更信息
- **AND** AI 归档聚合阶段的失败 MUST NOT 阻塞已通过校验的确定性归档 PR
