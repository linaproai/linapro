## MODIFIED Requirements

### Requirement: The host must provide structured plugin auto-enable entries in the main config file

The host SHALL provide `plugin.autoEnable` as a list of structured objects with required `id` and optional `withMockData`. Bare string entries MUST be rejected. `withMockData` defaults to `false` and only loads mock data during first-time startup installation.

#### Scenario: Parse a valid structured auto-enable list
- **WHEN** config contains entries with `id` and optional `withMockData`
- **THEN** the host parses each entry into an internal `(id, withMockData)` tuple
- **AND** startup loads mock data only for entries with `withMockData=true`

#### Scenario: Reject invalid auto-enable config
- **WHEN** config contains an empty id, missing id, non-boolean `withMockData`, or a bare string entry
- **THEN** startup fails with an error identifying the invalid entry

### Requirement: The auto-enable list must implicitly include install and enable semantics

For each auto-enable entry, startup SHALL install the plugin if needed and enable it. If `withMockData=true`, first-time installation MUST execute mock SQL through the same transactional path as manual installation. Already installed plugins MUST NOT reload mock data.

#### Scenario: Auto-enable without mock data
- **WHEN** an entry has no `withMockData` or has `withMockData=false`
- **THEN** startup installs and enables the plugin without scanning mock-data SQL

#### Scenario: Auto-enable with mock data opt-in
- **WHEN** an entry has `withMockData=true` and the plugin is not installed
- **THEN** startup executes mock-data SQL transactionally after install SQL succeeds

### Requirement: Any failure for a listed auto-enable plugin must block host startup

Any auto-enable failure MUST block host startup. Mock-data failures MUST roll back the mock transaction and surface plugin ID, failed file, and cause.

#### Scenario: Mock SQL failure during auto-enable causes startup failure
- **WHEN** install SQL succeeds but a mock SQL file fails
- **THEN** the host rolls back mock data and fails startup
- **AND** the error includes the plugin ID, failed mock SQL file, and failure cause
