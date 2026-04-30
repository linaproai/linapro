## ADDED Requirements

### Requirement: Plugin list query is side-effect free

The system SHALL treat plugin list queries as side-effect-free read operations. The list query may read discovered source manifests, dynamic plugin registry data, release snapshots, and governance projections, but MUST NOT create, update, or delete plugin governance table data. Plugin scanning and governance synchronization MUST be triggered only by explicit synchronization actions.

#### Scenario: Query plugin list from management page
- **WHEN** an administrator opens plugin management and calls `GET /api/v1/plugins`
- **THEN** the system returns the plugin list and current governance state
- **AND** the GET request does not write plugin registry, release, resource, menu, or role-menu governance tables

#### Scenario: Synchronize plugins explicitly
- **WHEN** an administrator triggers plugin synchronization with `POST /api/v1/plugins/sync`
- **THEN** the system scans source plugins and dynamic plugin artifacts
- **AND** it may synchronize registry, release snapshot, resource index, menu, and permission governance data

### Requirement: Plugin host-service metadata lookup must avoid schema probing errors

The system SHALL read host database metadata for plugin list host-service projections through read-only metadata queries. This lookup MUST NOT trigger incorrect business-table schema probing for `information_schema.TABLES`; if the database does not support the metadata lookup or the lookup fails, the plugin list API SHALL degrade to returning raw table names.

#### Scenario: Resolve data table comments for dynamic plugin permissions
- **WHEN** a plugin list item declares `data.resources.tables`
- **THEN** the system attempts to read table comments for permission review display
- **AND** the lookup does not emit schema probing errors against `information_schema.TABLES`

#### Scenario: Metadata lookup unavailable
- **WHEN** the current database dialect does not support host table comment lookup or the lookup fails
- **THEN** the plugin list API still returns successfully
- **AND** hostServices permission display uses raw table names as fallback information
