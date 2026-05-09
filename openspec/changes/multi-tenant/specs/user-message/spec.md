## ADDED Requirements

### Requirement: 用户消息表租户化
`sys_user_message`(若存在)SHALL 加 `tenant_id`;消息 inbox 仅返回当前租户的消息。

#### Scenario: 跨租户消息隔离
- **WHEN** 用户 U 在租户 A 收到 3 条消息,在租户 B 收到 5 条
- **THEN** U 持租户 A token 时 inbox 仅 3 条
- **AND** 切到租户 B 后 inbox 仅 5 条

### Requirement: 消息读取权限
用户 SHALL 仅可读取自己 + 当前租户的消息;跨租户访问被拒。

#### Scenario: 跨租户访问被拒
- **WHEN** 用户用 token A 尝试通过 message_id 读取实际属租户 B 的消息
- **THEN** 返回 404(避免泄露存在性)
