## Requirements

### Requirement: Dynamic Plugins Acquire Named Lock Resources

The system SHALL provide lock service reusing host distributed lock with ticket-based renew/release.

#### Scenario: Plugin acquires authorized lock
- **WHEN** a plugin calls lock service acquire
- **THEN** the host applies lease and timeout policies
- **AND** binds the logical lock name to plugin isolation
- **AND** returns a lock ticket

#### Scenario: Plugin renews or releases lock
- **WHEN** a plugin renews or releases a held lock
- **THEN** the host validates ticket and lock-resource match
- **AND** only operates on the current plugin's valid locks
