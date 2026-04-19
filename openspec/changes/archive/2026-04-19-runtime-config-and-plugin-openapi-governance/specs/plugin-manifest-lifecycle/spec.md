## MODIFIED Requirements

### Requirement: `plugin.yaml` Remains Minimal and May Declare Menus
The system SHALL keep `plugin.yaml` focused on stable plugin metadata and SHALL not require source plugins to declare backend route inventories in the manifest.

#### Scenario: Source plugin backend routes are not duplicated in the manifest
- **WHEN** the host parses a source plugin `plugin.yaml`
- **THEN** the manifest does not need to list backend routes
- **AND** backend route registration code plus DTO `g.Meta` remains the only source of truth for source-plugin routes
- **AND** the host captures route ownership and documentation metadata during registration instead of reading a second route declaration model from `plugin.yaml`
