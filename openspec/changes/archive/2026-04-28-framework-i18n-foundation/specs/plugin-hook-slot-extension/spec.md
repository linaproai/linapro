## MODIFIED Requirements

### Requirement: Dynamic plugin routing management metadata is concentrated in `g.Meta`

The system SHALL require the dynamic plugin to centrally define backend dynamic-routing management metadata in `g.Meta` of the API-layer request structure, avoiding a second scattered routing-management configuration source. Host-governed route fields SHALL remain explicit, while plugin-defined route metadata SHALL be transported through the generic route-contract `meta` map without host interpretation.

#### Scenario: Dynamic plugin declares minimum governance fields

- **WHEN** a developer defines a dynamic plugin backend interface
- **THEN** the interface can declare `access` and `permission` in `g.Meta`
- **AND** plugin-specific route declarations can be added as extra `g.Meta` tags and will be preserved in route-contract `meta`
- **AND** `access` supports only `public` and `login`
- **AND** missing `access` is treated as `login`

#### Scenario: Public routing governance boundaries are limited

- **WHEN** a developer declares a `public` dynamic route
- **THEN** the route MUST NOT declare `permission`
- **AND** the route MUST NOT rely on host login state injection
- **AND** illegal host-governance configurations are rejected during the host loading phase
