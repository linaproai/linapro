# 插件通知服务规范

## Purpose
待定 - 由归档变更 dynamic-plugin-host-service-extension 创建。归档后更新目的。
## Requirements
### Requirement: 宿主通知域与通知公告内容管理解耦

系统 SHALL 将宿主通知域设计为独立于 `linapro-content-notice` 插件中 `plugin_linapro_content_notice` 公告内容模型的统一通知发送与投递模型；通知公告内容管理继续由该插件负责，消息中心与插件 `notify` 能力统一基于新的通知域表实现，不再继续使用 `sys_user_message`。

#### Scenario: 发布通知公告时走统一通知域

- **WHEN** `linapro-content-notice` 将一条 `plugin_linapro_content_notice` 从草稿发布为生效状态
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

### Requirement: 通知消息读取必须返回类型化投影
系统 SHALL 将插件可见通知消息读取收敛到稳定`MessageProjection`或等价类型化 DTO。动态`messages.batch_get` MUST 返回类型化消息投影，不得使用未治理`map[string]any`作为长期协议载荷。

#### Scenario: 批量读取通知消息
- **WHEN** 插件批量读取消息 ID
- **THEN** 系统返回当前 actor 可见的类型化消息投影
- **AND** 不暴露通知内部存储结构或任意扩展字段 map

### Requirement: 通知必须支持按业务来源批量读取
系统 SHALL 提供`Notifications.BatchGetBySource`和动态`messages.by_source.batch_get`，按`SourceType + SourceIDs`集合化读取当前 actor 可见消息。

#### Scenario: 按来源读取消息
- **WHEN** 插件按一组业务来源 ID 读取通知消息
- **THEN** 系统批量返回这些来源下当前 actor 可见的消息投影
- **AND** 不得对每个来源执行一次消息列表查询作为常规实现

### Requirement: 通知必须支持可见性校验
系统 SHALL 提供`Notifications.EnsureVisible`和动态`messages.visible.ensure`，用于写入关联或执行动作前校验消息目标可见。任一不可见、不存在或未授权消息 MUST 整体拒绝。

#### Scenario: 消息可见性校验失败
- **WHEN** 插件校验的消息集合包含不可见消息
- **THEN** 系统返回结构化拒绝错误
- **AND** 不暴露该消息是不存在还是不可见

### Requirement: 邮件通知通道必须通过 linapro-mail-core 出站

系统 SHALL 在宿主通知域处理 `channel_type=email`（或等价邮件通道）发送时，通过 `linapro-mail-core` 发布的邮件能力完成实际出站，MUST NOT 直接依赖或调用 `linapro-mail-smtp` 及其他协议插件实现。通道配置 MUST 能解析到 mail-core 的 Account（例如 `config_json.accountId` 或平台默认 Account）。

#### Scenario: 邮件通道成功发送

- **WHEN** 邮件通道已启用、mail-core 可用，且通道配置指向合法 Account（或存在可用默认 Account）
- **THEN** 宿主 notify MUST 创建消息与 delivery 记录并委托 mail-core 发送
- **AND** 发送路径 MUST NOT import 或直连具体 SMTP 协议插件

#### Scenario: mail-core 不可用

- **WHEN** 调用方通过 notify 向 email 通道发送，但 `linapro-mail-core` 未安装、未启用或无法解析可用 Account
- **THEN** 系统 MUST fail-closed 并返回明确错误
- **AND** MUST NOT 静默降级为 inbox 或丢弃发送意图

### Requirement: 邮件投递状态必须可观测

系统 SHALL 在邮件通道发送路径中维护 delivery 状态（至少 pending/succeeded/failed 语义与现有 delivery 模型对齐）。发送失败 MUST 可反映为失败状态，供运维与审计观察。

#### Scenario: SMTP 失败标记 delivery

- **WHEN** mail-core 出站因协议或配置失败
- **THEN** 对应 notify delivery MUST 记录失败语义（或等价可查询状态）
- **AND** 调用方 MUST 能获得结构化错误而非仅超时

