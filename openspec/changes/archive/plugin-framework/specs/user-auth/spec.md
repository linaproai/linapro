## Requirements

### Requirement: Authentication Lifecycle Events Available for Plugin Subscription

The system SHALL publish login/logout success events as hooks for enabled plugins.

#### Scenario: Login success hook
- **WHEN** a user logs in successfully
- **THEN** the host dispatches `auth.login.succeeded` to subscribed plugins

### Requirement: Plugin Auth Extension Failure Does Not Affect Auth Result

Plugin failures during auth hooks do not change login/logout outcomes.
