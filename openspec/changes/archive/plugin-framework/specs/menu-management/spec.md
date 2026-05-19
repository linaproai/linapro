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

### Requirement: Plugin Menu Semantic Mount Declared by Plugin Manifest

The system SHALL treat the plugin menu's semantic mount position as the plugin manifest's product assembly declaration, not as a host runtime strategy hardcoded by plugin ID.

#### Scenario: Plugin declares semantic mount directory
- **WHEN** a plugin needs to sync menus to the host
- **THEN** the plugin declares the corresponding `parent_key` in its own `plugin.yaml`
- **AND** the host resolves the parent menu per that declaration, rather than maintaining a hardcoded mapping of official plugin IDs to parent directories

#### Scenario: Plugin declares non-existent parent menu
- **WHEN** a plugin manifest declares a `parent_key` that does not exist in the host menu records
- **THEN** the host fails to sync the menu
- **AND** the error states the `parent_key` cannot be resolved

### Requirement: Default Backend Uses Stable First-Level Directory Structure

The system SHALL provide a stable first-level directory structure for the default management backend oriented toward the project management backend main scenario.

#### Scenario: Query default backend menu skeleton
- **WHEN** the host projects the default backend menu for the current user
- **THEN** first-level directories are organized as: `Dashboard`, `Permission Management`, `Organization Management`, `System Settings`, `Content Management`, `System Monitoring`, `Task Scheduling`, `Extension Center`, `Development Center`
- **AND** the corresponding host stable parent `menu_key` values are `dashboard`, `iam`, `org`, `setting`, `content`, `monitor`, `scheduler`, `extension`, `developer`

#### Scenario: First-level directories exist as host stable directory records
- **WHEN** the host initializes or synchronizes the default backend menu skeleton
- **THEN** these first-level directories are created and owned by the host
- **AND** plugins can reference these host directories via `parent_key`, but the host must not force specific plugins to only mount to one fixed directory

#### Scenario: Default backend extends business modules
- **WHEN** developers continuously add business modules or official source plugins to the project
- **THEN** new menus will preferentially be placed in existing stable directories
- **AND** there is no need to frequently refactor first-level navigation naming and structure
