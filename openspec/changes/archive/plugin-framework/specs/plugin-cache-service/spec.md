## Requirements

### Requirement: Dynamic Plugins Access Distributed Cache via Named Spaces

The system SHALL provide cache service via MySQL MEMORY table with namespace isolation and strict length validation.

#### Scenario: Plugin accesses authorized cache space
- **WHEN** a plugin calls cache service get/set/delete/incr/expire
- **THEN** the host only allows access to authorized `host-cache` resources
- **AND** data is stored in the shared MEMORY table

#### Scenario: Oversized value is rejected with explicit error
- **WHEN** a plugin writes data exceeding namespace/key/value length limits
- **THEN** the host returns an explicit error without truncating or partial write

#### Scenario: Unauthorized cache space is rejected
- **WHEN** a plugin calls an unauthorized cache space
- **THEN** the host rejects the call
