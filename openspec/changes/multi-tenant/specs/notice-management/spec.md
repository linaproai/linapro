## ADDED Requirements

### Requirement: 通知消息与投递租户化
`sys_notify_message` 与 `sys_notify_delivery` SHALL 加 `tenant_id`;CRUD 按租户过滤;跨租户通知由平台管理员通过 `/platform/notify/*` 接口发起,显式指定目标租户。`sys_notify_channel` 是平台全局通道目录,不携带 `tenant_id`;租户级渠道覆盖若后续落地,应使用独立租户配置表表达,不得破坏全局 `channel_key` 语义。

#### Scenario: 租户内通知
- **WHEN** 租户 A 管理员发送通知给本租户用户
- **THEN** 写入 `tenant_id=A`
- **AND** 仅租户 A 用户可见

#### Scenario: 平台广播通知
- **WHEN** 平台管理员发送系统通知给所有租户
- **THEN** 通过 `/platform/notify/broadcast` 显式指定 `target_tenant_ids=[A,B,C]`
- **AND** 各租户独立写入一条消息

### Requirement: 通知发送渠道保持平台全局配置
通知渠道目录(`sys_notify_channel`)SHALL 保持平台全局配置,内置 `inbox` 通道由全局 `channel_key` 唯一识别。租户管理员不得直接修改全局通道配置;若未来需要短信/邮件/webhook 的租户覆盖,必须通过单独的租户配置或覆盖表实现。

#### Scenario: 内置站内信通道全局唯一
- **WHEN** 任意租户发送站内信
- **THEN** 系统读取全局 `channel_key=inbox` 通道配置
- **AND** 消息与投递记录仍按发送上下文写入对应 `tenant_id`

### Requirement: 通知投递日志按租户记录
`sys_notify_delivery` SHALL 加 `tenant_id`;查询接口 MUST 按租户隔离,租户管理员仅可见本租户日志。平台管理员仅通过 `/platform/notify/deliveries` 管理平台接口查看全量;impersonation 模式下仍仅可见目标租户投递日志。

#### Scenario: 跨租户日志不可见
- **WHEN** 租户 A 管理员查询投递日志
- **THEN** 仅返回 `tenant_id=A` 行
- **AND** 不见租户 B 的投递记录
