## Requirements

### Requirement: Dynamic Plugins Make Outbound HTTP via Authorized URL Patterns

The system SHALL provide controlled outbound HTTP where plugins access only authorized URL patterns.

#### Scenario: Plugin calls authorized URL
- **WHEN** a plugin sends an outbound HTTP request
- **THEN** the target URL must match an authorized pattern
- **AND** scheme must match exactly
- **AND** host supports glob matching
- **AND** path matches by prefix after normalization

#### Scenario: Query and fragment do not participate in authorization
- **WHEN** a URL differs only in query or fragment
- **THEN** the host does not reject based on those differences

#### Scenario: Default deny on any dimension mismatch
- **WHEN** scheme, host, port, or path does not match
- **THEN** the host rejects the request

#### Scenario: Upstream failure is isolated
- **WHEN** an upstream call fails
- **THEN** the host returns a structured failure result
- **AND** the failure does not affect other requests
