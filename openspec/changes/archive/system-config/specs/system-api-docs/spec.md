## MODIFIED Requirements

### Requirement: System API Documentation Uses a Host-Managed OpenAPI Document
The system SHALL provide the system API documentation page by rendering a host-managed `/api.json` OpenAPI v3 document instead of relying directly on GoFrame's built-in output.

#### Scenario: User opens the system API documentation page
- **WHEN** a user navigates to the system API documentation page
- **THEN** the page renders the host-managed `/api.json`
- **AND** the document contains host static APIs plus the currently enabled plugin APIs that the host chooses to project

### Requirement: Enabled Dynamic Plugin Routes Are Projected Into the Host Document
The system SHALL project currently enabled dynamic-plugin routes into the host-managed OpenAPI document.

#### Scenario: Dynamic plugin route disappears after disable or uninstall
- **WHEN** a dynamic plugin is disabled, uninstalled, or switched to a release that no longer exposes a route
- **THEN** the host removes that route projection from the OpenAPI document

## ADDED Requirements

### Requirement: Enabled Source Plugin Routes Are Projected by Ownership Binding
The system SHALL project only the currently enabled source-plugin routes that were captured during registration.

#### Scenario: Enabled source plugin route appears in system API docs
- **WHEN** a source plugin is enabled and its DTO route binding was captured during registration
- **THEN** the system API document shows that route with the metadata declared in DTO `g.Meta`

#### Scenario: Disabled source plugin route is removed from system API docs
- **WHEN** a source plugin is disabled
- **THEN** the host removes that source plugin route from the projected OpenAPI document even if the underlying route still exists in the route table

### Requirement: Internal and Non-Business Routes Are Excluded From the System API Document
The host SHALL exclude internal routing infrastructure and non-business routes from the system API document.

#### Scenario: Host scans the real route table
- **WHEN** the host builds the OpenAPI document from the live route table
- **THEN** it excludes static fallbacks, dynamic plugin dispatch entry routes, the host-managed `/api.json` handler, and other internal non-business routes
