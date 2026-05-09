## ADDED Requirements

### Requirement: 通知表租户化
`sys_notify_*`(channel/delivery/message)SHALL 加 `tenant_id`;CRUD 按租户过滤;跨租户通知由平台管理员通过 `/platform/notify/*` 接口发起,显式指定目标租户。

#### Scenario: 租户内通知
- **WHEN** 租户 A 管理员发送通知给本租户用户
- **THEN** 写入 `tenant_id=A`
- **AND** 仅租户 A 用户可见

#### Scenario: 平台广播通知
- **WHEN** 平台管理员发送系统通知给所有租户
- **THEN** 通过 `/platform/notify/broadcast` 显式指定 `target_tenant_ids=[A,B,C]`
- **AND** 各租户独立写入一条消息

### Requirement: 通知发送渠道按租户配置
通知渠道(短信/邮件/webhook)SHALL 支持按租户独立配置(覆盖平台默认);租户管理员仅可改本租户配置。

#### Scenario: 租户覆盖邮件渠道
- **WHEN** 租户 A 管理员设置自己的 SMTP
- **THEN** 该租户的通知走自己 SMTP,fallback 时才走平台默认

### Requirement: 通知投递日志按租户记录
`sys_notify_delivery` SHALL 加 `tenant_id`;查询接口 MUST 按租户隔离,租户管理员仅可见本租户日志。平台管理员仅通过 `/platform/notify/deliveries` 管理平台接口查看全量;impersonation 模式下仍仅可见目标租户投递日志。

#### Scenario: 跨租户日志不可见
- **WHEN** 租户 A 管理员查询投递日志
- **THEN** 仅返回 `tenant_id=A` 行
- **AND** 不见租户 B 的投递记录
