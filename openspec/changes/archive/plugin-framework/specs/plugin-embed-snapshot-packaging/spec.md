## Requirements

### Requirement: Dynamic Plugins Support go:embed Resource Declaration

The system SHALL allow dynamic plugins to declare `plugin.yaml`, `frontend`, `manifest` resources via `go:embed`, matching the source plugin author experience.

#### Scenario: Dynamic plugin declares embedded resources
- **WHEN** a dynamic plugin needs to deliver manifest, frontend, or SQL resources with `.wasm`
- **THEN** the plugin declares these via `go:embed`
- **AND** declaration paths must match directory conventions
- **AND** the builder reads from the embedded filesystem

### Requirement: Builder Converts Embedded Resources to Host Snapshots

The system SHALL have the builder convert `go:embed` resources into host-governable WASM custom section snapshots.

#### Scenario: Generate snapshots from embedded resources
- **WHEN** the builder processes a dynamic plugin with embedded resources
- **THEN** it reads manifest, frontend, and SQL from the embedded filesystem
- **AND** generates the host-recognized manifest, frontend, and SQL custom sections
- **AND** the resulting artifact is directly consumable by host governance

### Requirement: Directory Scan Fallback Preserved for Migration

The system SHALL continue allowing dynamic plugins without `go:embed` to build via directory scan during the migration period.

#### Scenario: Legacy dynamic plugin without embedded resources
- **WHEN** the builder processes a dynamic plugin without `go:embed`
- **THEN** it falls back to directory scan for manifest, frontend, and SQL
- **AND** the host governance logic remains unchanged

### Requirement: Build Output Converges to Unified Directory

The system SHALL write dynamic plugin build artifacts to `temp/output/` or caller-specified output, not back to plugin source directories.

#### Scenario: Default output directory
- **WHEN** building via standard entry
- **THEN** the artifact writes to `temp/output/<plugin-id>.wasm`
- **AND** intermediate WASM files only go to the unified output directory
