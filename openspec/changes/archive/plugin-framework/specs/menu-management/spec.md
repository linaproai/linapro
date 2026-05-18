## Requirements

### Requirement: Menu Management Shows Plugin Menu Ownership

The system SHALL display plugin menu ownership and lifecycle status in menu management.

### Requirement: Plugin State Links to Menu Visibility

The system SHALL control menu visibility based on plugin enable/disable state, with immediate refresh on changes.

#### Scenario: Plugin disabled
- **WHEN** a plugin is disabled
- **THEN** its menus disappear from navigation
- **AND** direct route access returns controlled feedback

#### Scenario: Menu visibility change triggers immediate refresh
- **WHEN** an administrator changes menu visibility
- **THEN** the current user's navigation updates immediately

### Requirement: Plugin Governance Uses Stable Menu Key

The system SHALL use `menu_key` as the stable menu identifier for plugin governance. `remark` is display-only.
