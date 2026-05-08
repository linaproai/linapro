## Requirements

### Requirement: Dynamic Plugins Access Data via Host-Authorized Tables

The system SHALL provide table-level data service where plugins access data through authorized tables, not database connections.

#### Scenario: Plugin queries authorized table
- **WHEN** a plugin calls data service to query a table
- **THEN** the host validates the table is in the release authorization snapshot
- **AND** only allows access to authorized method sets

#### Scenario: Raw SQL is not a public capability
- **WHEN** a developer tries to declare raw SQL capability
- **THEN** the builder or host rejects the declaration

### Requirement: Data Service Reuses User Permission and Data Scope

The system SHALL apply current user identity, role permissions, and data scope to plugin data service calls in request-bound context.

#### Scenario: Logged-in user triggers data service
- **WHEN** a logged-in user triggers a plugin route that calls data service
- **THEN** the host applies user ID, role permissions, and data scope

#### Scenario: Sensitive data call without user context
- **WHEN** a data method requiring user context is called from a hook or cron
- **THEN** the host rejects or limits to system-level resources

### Requirement: Data Service Executes via DAO and ORM Contracts

The system SHALL execute data requests through controlled DAO and GoFrame `gdb` ORM components, not raw SQL.

#### Scenario: Host executes authorized data query
- **WHEN** a plugin queries an authorized table
- **THEN** the host resolves to a DAO operation plan
- **AND** assembles query conditions, projections, sorting, and pagination through controlled DAO/Model

### Requirement: DoCommit Interception for Data Governance

The system SHALL intercept data service execution at the `DoCommit` layer for permission control, audit, and transaction governance.

### Requirement: plugindb Guest SDK for Data Access

The system SHALL provide `pkg/plugindb` as the recommended guest-side ORM-style SDK, with strong-typed enums for actions, filters, sorting, mutations, and access modes.
