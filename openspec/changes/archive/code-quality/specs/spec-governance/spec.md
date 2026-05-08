## ADDED Requirements

### Requirement: Main spec structure standardization

The system SHALL standardize main specs under `openspec/specs/` to the current OpenSpec schema structure, including at least `## Purpose` and `## Requirements` core sections.

#### Scenario: Validating main spec structure

- **WHEN** a developer runs OpenSpec validation or inspection on any main spec
- **THEN** the spec uses the current schema-recognizable section structure
- **AND** does not fail due to missing required sections such as `Purpose` or `Requirements`

### Requirement: Archive residual governance

The system SHALL identify and govern half-finished main spec updates after failed or interrupted archival, preventing subsequent archives from being blocked by duplicate additions or residual files.

#### Scenario: Half-finished main spec files exist

- **WHEN** an archival process was interrupted abnormally and left half-finished main spec files
- **THEN** the system or maintenance workflow can identify the residual
- **AND** clean up or align it before the next archival, avoiding duplicate capability writes

### Requirement: Main spec verifiable before archival

The system SHALL ensure that main specs about to be updated pass current OpenSpec schema basic validation before executing archival.

#### Scenario: Executing change archival

- **WHEN** a developer archives a change that will update main specs
- **THEN** all affected main specs pass structure and requirement-level validation
- **AND** the archival process does not break due to historical main spec format incompatibility
