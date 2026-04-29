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

#### Scenario: Shell trigger remains available when shell editing is blocked
- **WHEN** Shell 任务因为环境开关或 Shell 附加权限不足无法新增或编辑
- **THEN** 任务列表仍 SHALL 为可手动触发的 Shell 任务展示可点击的“立即执行”操作
- **AND** 点击后 SHALL 先展示二次确认弹窗，而不是被 Shell 编辑限制直接置灰
- **AND** 操作列 SHALL 只展示一个编辑入口，避免禁用态编辑和普通编辑同时出现

#### Scenario: Shell jobs are enabled by default
- **WHEN** 系统初始化 `cron.shell.enabled` 运行参数或该参数缺失后使用内置 fallback
- **THEN** 默认值 SHALL 为 `true`
- **AND** 不影响平台不支持 Shell 任务时自动禁用的安全保护
