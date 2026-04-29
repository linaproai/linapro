## ADDED Requirements

### Requirement: Manual job trigger must require confirmation
定时任务“立即执行”操作 SHALL 在真正触发任务前展示二次确认弹窗，防止误触发运维任务。

#### Scenario: Trigger action asks for confirmation
- **WHEN** 管理员在定时任务列表点击“立即执行”
- **THEN** 前端展示确认弹窗，说明确认后会立即触发该任务
- **AND** 管理员确认前不得调用触发接口

#### Scenario: Trigger confirmation uses execution styling
- **WHEN** 确认弹窗展示“立即执行”操作
- **THEN** 弹窗复用与删除按钮一致的确认组件模式
- **AND** 确认按钮使用区分执行动作的样式和文案，而不是删除动作样式

#### Scenario: Canceling trigger does nothing
- **WHEN** 管理员取消“立即执行”确认弹窗
- **THEN** 前端不调用 `POST /job/{id}/trigger`
- **AND** 任务列表状态保持不变
