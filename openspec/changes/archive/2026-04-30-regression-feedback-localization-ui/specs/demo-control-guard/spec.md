## ADDED Requirements

### Requirement: The demo-control plugin must reject plugin-governance write operations when enabled

When `demo-control` is enabled, the system SHALL reject plugin governance writes, including synchronization, dynamic package upload, installation, uninstallation, enablement, and disablement.

#### Scenario: Plugin governance writes are rejected
- **WHEN** `demo-control` is enabled
- **AND** a caller sends a plugin governance write request
- **THEN** demo-control rejects the request with a read-only demo message

#### Scenario: Plugin management reads stay allowed
- **WHEN** `demo-control` is enabled
- **AND** a caller sends a plugin management read request
- **THEN** demo-control allows the request to continue
