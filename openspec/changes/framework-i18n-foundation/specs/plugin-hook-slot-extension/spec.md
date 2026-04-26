## MODIFIED Requirements

### Requirement: Dynamic plugin routing management metadata is concentrated in `g.Meta`

The system SHALL require the dynamic plugin to centrally define backend dynamic-routing management metadata in `g.Meta` of the API-layer request structure, avoiding a second scattered routing-management configuration source. Host-governed route fields SHALL remain explicit, while plugin-defined route metadata SHALL be transported through the generic route-contract `meta` map without host interpretation.

#### Scenario: Dynamic plugin declares minimum governance fields

- **WHEN** Developer defines a dynamic plugin backend interface
- **THEN** This interface can declare `access` and `permission` in `g.Meta`
- **AND** plugin-specific route declarations can be added as extra `g.Meta` tags and will be preserved in route-contract `meta`
- **AND** `access` only supports `public` and `login`
- **AND** If `access` is not declared, it will be processed as `login`

#### Scenario: Public routing governance boundaries are limited

- **WHEN** Developer declares a `public` dynamic route
- **THEN** This route MUST not declare `permission`
- **AND** This route MUST not rely on host login state injection
- **AND** Illegal host-governance configurations will be rejected during the host loading phase
