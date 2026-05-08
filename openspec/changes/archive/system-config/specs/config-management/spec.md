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

### Requirement: The default upload size must be unified at 20 MB
The system SHALL set the platform default value of `sys.upload.maxSize` to `20`, and database initialization, config-template defaults, and runtime upload fallbacks SHALL all use that same value unless an administrator explicitly overrides it.

#### Scenario: Host initialization writes the 20 MB default
- **WHEN** an administrator runs the host initialization SQL
- **THEN** the default value of `sys.upload.maxSize` in `sys_config` is `20`
- **AND** the default value read by config management for that built-in parameter is also `20`

#### Scenario: Runtime default remains 20 MB when no override is provided
- **WHEN** the host handles a `multipart` upload request without any administrator override for the upload-size setting
- **THEN** file-upload validation enforces a 20 MB limit
- **AND** the friendly error message triggered by the default limit returns wording equivalent to "file size cannot exceed 20 MB"

### Requirement: All default upload-size sources must stay consistent
The system SHALL keep the database seed value, config-template default, and host static fallback value for `sys.upload.maxSize` consistent so different startup paths do not expose different default upload limits.

#### Scenario: The host starts from the default template
- **WHEN** an operator generates runtime config from the host default `config.template.yaml` and does not change the upload limit separately
- **THEN** the host reads a default upload size of 20 MB
- **AND** that default matches the `sys.upload.maxSize` default written by the host initialization SQL

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

### Requirement: The config-management component must have a unit-test coverage gate
The system SHALL maintain repeatable unit tests for the `apps/lina-core/internal/service/config` config-management component, and SHALL use package-level coverage verification as a delivery gate before that component is considered ready.

#### Scenario: Package-level coverage meets the delivery bar
- **WHEN** a maintainer runs `go test ./internal/service/config -cover` from `apps/lina-core`
- **THEN** the command succeeds
- **AND** the reported package-level statement coverage is not lower than `80%`

### Requirement: Critical config-management branches must have automated regression protection
The system SHALL add automated unit tests for critical helper logic inside the config-management component, including high-risk branches around defaults and fallbacks, cache or snapshot reuse, and invalid input or error propagation.

#### Scenario: Plugin and public-frontend config helper logic changes
- **WHEN** a change touches plugin dynamic storage paths, protected public-frontend config key checks, or the shared validation entry point
- **THEN** unit tests cover the normal read path
- **AND** cover default-value or compatibility-fallback behavior
- **AND** cover invalid input or empty-value defensive behavior

#### Scenario: Runtime-parameter cache and revision synchronization logic changes
- **WHEN** a change touches runtime-parameter snapshot caches, the revision controller, or shared-KV synchronization logic
- **THEN** unit tests cover cache-hit or local-reuse behavior
- **AND** cover rebuilds after revision changes
- **AND** cover error propagation and defensive behavior for shared-KV read failures, invalid cached values, or equivalent exceptional cases
