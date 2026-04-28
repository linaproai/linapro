## MODIFIED Requirements

### Requirement: Dynamic plugin runtime artifacts carry governable route contracts

The system SHALL allow dynamic plugins to carry backend dynamic route contracts in runtime artifacts. After loading an artifact, the host can restore route paths, methods, and minimum governance metadata without scanning source directories again at request time. Custom route declarations consumed by the plugin or plugin middleware SHALL pass through the generic `meta` field. The host SHALL NOT define or validate business-plugin-specific route fields.

#### Scenario: Build phase extracts dynamic route contracts

- **WHEN** a dynamic plugin runtime artifact is built
- **THEN** the builder extracts dynamic route metadata from request structure `g.Meta` values under `backend/api/**/*.go`
- **AND** writes paths, methods, documentation, and host governance fields into a dedicated section of the runtime artifact
- **AND** writes plugin custom declarations that are not host route-contract fields into route-contract `meta`
- **AND** the host can restore them as dynamic plugin `manifest.Routes` after loading the artifact

#### Scenario: Host validates dynamic route contracts

- **WHEN** the host loads a dynamic plugin route contract
- **THEN** the host validates internal path, method, `access`, and `permission`
- **AND** missing `access` is treated as `login`
- **AND** `public` routes MUST NOT declare `permission`
- **AND** the host does not validate or interpret plugin custom route declarations in `meta`
- **AND** an invalid contract causes artifact loading to fail
