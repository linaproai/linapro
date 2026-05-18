## Requirements

### Requirement: Inactive Session Auto-Cleanup

The system SHALL clean up inactive sessions via scheduled task. Timeout and interval use duration strings (`session.timeout`, `session.cleanupInterval`).

#### Scenario: Throttled active time writes
- **WHEN** a protected request arrives within the throttle window
- **THEN** authentication still checks session validity
- **AND** the system may skip the `last_active_time` database update
