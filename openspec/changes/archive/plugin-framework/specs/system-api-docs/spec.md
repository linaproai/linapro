## Requirements

### Requirement: System API Auto-Merges Dynamic Plugin Route Documentation

The system SHALL project enabled dynamic plugin route contracts into the OpenAPI documentation.

#### Scenario: Dynamic plugin routes appear in system API
- **WHEN** a dynamic plugin is enabled with route contracts
- **THEN** routes appear as `/api/v1/extensions/{pluginId}/...` in OpenAPI
- **AND** each route includes method, tags, summary, description, and security

#### Scenario: Executable routes show real response semantics
- **WHEN** a route has an executable bridge
- **THEN** the doc shows 200 and 500 responses

#### Scenario: Non-executable routes show 501 placeholder
- **WHEN** a route lacks an executable bridge
- **THEN** the doc shows a 501 placeholder description

#### Scenario: Disabled plugin routes removed from docs
- **WHEN** a dynamic plugin is disabled or uninstalled
- **THEN** its routes are removed from OpenAPI
