# plugin-notify-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
## Requirements
### Requirement: 宿主通知域与通知公告内容管理解耦

系统 SHALL 将宿主通知域设计为独立于 `sys_notice` 的统一通知发送与投递模型；`sys_notice` 继续负责通知公告内容管理，消息中心与插件 `notify` 能力统一基于新的通知域表实现，不再继续使用 `sys_user_message`。

#### Scenario: 发布通知公告时走统一通知域

- **WHEN** 宿主将一条 `sys_notice` 从草稿发布为生效状态
- **THEN** 宿主通过统一的 `notify` 服务创建消息主记录与 inbox 投递记录
- **AND** 宿主不得继续直接写入 `sys_user_message`

#### Scenario: 用户消息中心继续保留现有预览语义

- **WHEN** 当前用户在消息中心查看一条由通知公告产生的站内消息
- **THEN** 宿主仍然返回可用于预览公告的 `sourceType/sourceId` 语义
- **AND** 前端可以继续据此打开通知公告预览

### Requirement: 动态插件通过命名通知通道发送宿主通知

系统 SHALL 为动态插件提供受治理的通知服务，插件只能通过宿主授权的通知通道发送站内信、邮件、Webhook 等通知。

#### Scenario: 插件使用授权通知通道

- **WHEN** 插件调用通知服务向已授权的`host-notify-channel`发送通知
- **THEN** 宿主校验通道权限、模板或消息体约束
- **AND** 宿主按对应通知通道完成发送

#### Scenario: 插件尝试使用未授权通知通道

- **WHEN** 插件调用一个未授权的通知通道
- **THEN** 宿主拒绝该调用
- **AND** 宿主不向 guest 暴露宿主通知后端实现细节

