## ADDED Requirements

### Requirement: Plugin manifest SQL resources must classify mock-data as a separate asset type

When scanning a plugin `manifest/sql/` directory, the host SHALL distinguish install, uninstall, and mock-data SQL assets without overlap. Mock SQL files MUST NOT appear in install or uninstall asset lists. Source plugins and dynamic plugins MUST use the same scanning logic.

#### Scenario: Install asset list excludes mock-data files
- **WHEN** the host resolves install SQL assets for a plugin that also contains mock SQL
- **THEN** the install asset list excludes files under `manifest/sql/mock-data/`

#### Scenario: Mock asset scan returns mock-data files only
- **WHEN** the host resolves mock SQL assets for the same plugin
- **THEN** the returned asset list contains only files under `manifest/sql/mock-data/`
- **AND** files are sorted by file name ascending

### Requirement: Dynamic plugin packaging must preserve the mock-data directory convention

Dynamic plugin packaging SHALL preserve `manifest/sql/mock-data/` in the artifact file-system view and use the same runtime scanning method as source plugins.

#### Scenario: Dynamic plugin upgrade preserves mock-data visibility
- **WHEN** a dynamic plugin adds or changes mock-data SQL files in a new version
- **AND** the host upgrades to the new artifact
- **THEN** mock SQL scanning reflects the new version contents
