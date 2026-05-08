## Requirements

### Requirement: Installation Dialog Supports Install-and-Enable Shortcut

The system SHALL provide both "Install Only" and "Install and Enable" in the installation dialog.

#### Scenario: Show shortcut with both permissions
- **WHEN** the user has both `plugin:install` and `plugin:enable`
- **THEN** the dialog shows both actions

#### Scenario: Hide shortcut with install-only permission
- **WHEN** the user has only `plugin:install`
- **THEN** only "Install Only" is shown

#### Scenario: Show real state on partial failure
- **WHEN** install succeeds but enable fails
- **THEN** the page shows `installed but disabled`
