## Requirements

### Requirement: Plugin List Query is Side-Effect Free

The system SHALL treat plugin list queries as read-only. Governance table writes only happen via explicit sync actions.

#### Scenario: Query plugin list
- **WHEN** an administrator calls `GET /api/v1/plugins`
- **THEN** the system returns the plugin list without writing governance tables

### Requirement: Plugin Host-Service Metadata Lookup Avoids Schema Probing

Table comment lookup uses safe read-only metadata queries. Failures degrade to raw table names.

#### Scenario: Resolve data table comments
- **WHEN** a plugin list declares `data.resources.tables`
- **THEN** the system reads table comments without schema probing errors

#### Scenario: Metadata lookup unavailable
- **WHEN** the database doesn't support comment lookup
- **THEN** the API returns successfully with raw table names as fallback
