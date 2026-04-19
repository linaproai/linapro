## ADDED Requirements

### Requirement: Built-in Runtime Parameter Metadata
The system SHALL provide built-in metadata records for host-consumed runtime parameters so administrators can manage effective host behavior directly from config management.

#### Scenario: Initialize built-in runtime parameters
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains `sys.jwt.expire`, `sys.session.timeout`, `sys.upload.maxSize`, and `sys.login.blackIPList`
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Built-in Runtime Parameter Protection
The system SHALL validate built-in runtime parameter values and SHALL protect stable host-owned keys from rename or deletion.

#### Scenario: Reject invalid built-in runtime parameter values
- **WHEN** a user creates, updates, or imports one of the built-in runtime parameters with an invalid value format
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of built-in runtime parameter keys
- **WHEN** a user attempts to rename or delete a built-in runtime parameter key already consumed by the host
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Upload Size Parameter Must Drive Host Runtime Behavior
The system SHALL ensure that `sys.upload.maxSize` is enforced by the host upload chain instead of existing only as editable metadata.

#### Scenario: Upload size change takes effect immediately
- **WHEN** an administrator updates `sys.upload.maxSize` to `1`
- **THEN** subsequent upload requests are validated against a 1 MB limit
- **AND** uploads above the configured limit are rejected

### Requirement: Multi-Instance Runtime Parameter Cache Synchronization
The system SHALL use a local snapshot plus shared revision strategy for protected runtime parameter reads so hot paths do not query `sys_config` on every request.

#### Scenario: Runtime reads hit the local snapshot
- **WHEN** a node repeatedly reads protected runtime parameters while the shared revision has not changed
- **THEN** the node reuses its local in-memory snapshot
- **AND** it does not need to query `sys_config` on every read

#### Scenario: Parameter changes propagate to other instances
- **WHEN** a protected runtime parameter changes on one instance
- **THEN** the writing instance clears its local snapshot and bumps the shared revision
- **AND** other instances rebuild their local snapshots during the next synchronization cycle

### Requirement: Public Frontend Setting Metadata
The system SHALL provide built-in metadata for safe public frontend settings used by branding, login-page presentation, and workspace theme bootstrap.

#### Scenario: Initialize public frontend settings
- **WHEN** an administrator executes the host initialization SQL
- **THEN** `sys_config` contains the built-in public frontend setting keys used by the login page and workspace bootstrap
- **AND** each record includes a readable name, a default value, and a format description

### Requirement: Public Frontend Setting Protection
The system SHALL validate built-in public frontend setting values and SHALL protect their stable keys from rename or deletion.

#### Scenario: Reject invalid public frontend setting values
- **WHEN** a user creates, updates, or imports a built-in public frontend setting with an invalid enum, boolean, or required-text value
- **THEN** the system rejects the change and returns a validation error

#### Scenario: Reject rename or deletion of public frontend setting keys
- **WHEN** a user attempts to rename or delete a built-in public frontend setting key already consumed by the login page or admin workspace
- **THEN** the system rejects the operation and keeps the parameter record intact

### Requirement: Login and Workspace Consume Public Frontend Settings
The system SHALL expose a safe whitelist endpoint for public frontend settings and SHALL let the login page and admin workspace consume that contract.

#### Scenario: Public frontend settings are available before login
- **WHEN** a browser loads the login page without an authenticated session
- **THEN** the frontend can read the whitelisted branding and presentation settings through the public endpoint
- **AND** the endpoint does not expose arbitrary `sys_config` keys

#### Scenario: Updated branding is reflected after refresh
- **WHEN** an administrator updates public frontend settings and a user refreshes the login page or workspace
- **THEN** the refreshed UI shows the updated branding, copy, and theme defaults
