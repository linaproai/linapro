## Requirements

### Requirement: Notification Domain Decoupled from Notice Content Management

The system SHALL separate the notification domain from `sys_notice` content management, replacing `sys_user_message` with unified notification domain tables.

#### Scenario: Publishing notice uses unified notification domain
- **WHEN** a `sys_notice` is published
- **THEN** the host creates message and delivery records via the notify service
- **AND** does not write directly to `sys_user_message`

#### Scenario: Message center preserves preview semantics
- **WHEN** a user views a notice-originated inbox message
- **THEN** the host returns `sourceType/sourceId` for notice preview

### Requirement: Dynamic Plugins Send Notifications via Authorized Channels

The system SHALL provide notification service where plugins send through authorized notification channels.

#### Scenario: Plugin uses authorized channel
- **WHEN** a plugin calls notification service send
- **THEN** the host validates channel permissions and message constraints

#### Scenario: Unauthorized channel is rejected
- **WHEN** a plugin calls an unauthorized notification channel
- **THEN** the host rejects the call
