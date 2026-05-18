## Requirements

### Requirement: Demo Read-Only Mode Controlled by Plugin Enabled State

The system SHALL treat the `demo-control` plugin's enabled state as the runtime switch for demo read-only mode. `plugin.autoEnable` only controls startup auto-enable, not independent activation.

#### Scenario: Default config does not auto-enable demo-control
- **WHEN** `plugin.autoEnable` does not contain `demo-control`
- **THEN** the system does not apply extra write interception

#### Scenario: Manual enablement activates demo mode
- **WHEN** an administrator enables `demo-control` through governance
- **THEN** demo read-only mode takes effect immediately

### Requirement: Demo-Control Blocks Write Operations When Enabled

The system MUST block `POST`, `PUT`, `DELETE` requests when `demo-control` is enabled, with clear demo-mode error messages.
