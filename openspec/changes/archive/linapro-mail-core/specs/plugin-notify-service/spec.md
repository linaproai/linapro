## ADDED Requirements

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

### Requirement: notify 与 mail-core 的依赖边界

系统 SHALL 将 notify 对邮件出站建模为运行时可选能力：通过进程内 email delivery 契约（如 `notifycap.EmailDelivery`）接收 mail-core 提供的实现。mail-core 未提供实现时 email 通道 MUST 不可用。宿主 notify 模块 MUST NOT 将协议插件列为编译期硬依赖。

#### Scenario: 桥接仅绑定 owner

- **WHEN** mail-core 启用并注册 email delivery 实现
- **THEN** notify 发送路径 MUST 仅通过该契约调用
- **AND** MUST NOT 出现对 `linapro-mail-smtp`/`imap`/`pop3` 的直接 import 或硬编码插件 ID 发送路径
