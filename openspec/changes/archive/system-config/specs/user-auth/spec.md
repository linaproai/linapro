## ADDED Requirements

### Requirement: Runtime-Configured JWT Expiry
The system SHALL allow `sys.jwt.expire` to control the lifetime of newly issued JWT tokens at runtime and SHALL fall back to static configuration when no runtime override exists.

#### Scenario: Runtime JWT expiry takes effect
- **WHEN** an administrator maintains `sys.jwt.expire=24h`
- **THEN** newly issued JWT tokens use that duration as their effective expiry time

### Requirement: Runtime-Configured Login IP Blacklist
The system SHALL allow `sys.login.blackIPList` to control login IP blacklisting at runtime.

#### Scenario: Login request is denied by the configured blacklist
- **WHEN** a login request originates from an IP or CIDR range matched by `sys.login.blackIPList`
- **THEN** the system rejects the login attempt
- **AND** the login log records the failure reason that the login IP is forbidden
