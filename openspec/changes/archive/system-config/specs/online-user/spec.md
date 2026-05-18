## ADDED Requirements

### Requirement: Runtime-Configured Session Timeout
The system SHALL allow `sys.session.timeout` to control the online-session timeout threshold at runtime and SHALL fall back to static configuration when no runtime override exists.

#### Scenario: Runtime timeout value becomes effective
- **WHEN** an administrator maintains `sys.session.timeout=24h`
- **THEN** the host uses that duration as the effective online-session timeout threshold

### Requirement: Authentication Checks Session Timeout on Every Protected Request
The system SHALL evaluate session timeout during authentication instead of relying only on scheduled cleanup.

#### Scenario: Expired session is rejected during authentication
- **WHEN** a protected request carries a token whose session `last_active_time` exceeds the effective timeout threshold
- **THEN** the authentication chain rejects the request with `401`
- **AND** the host cleans up the corresponding online-session record
