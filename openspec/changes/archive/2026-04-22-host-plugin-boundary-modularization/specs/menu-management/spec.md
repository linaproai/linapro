## MODIFIED Requirements

### Requirement: By default, the background uses a stable first-level directory structure

The system SHALL provides the default management backend with a stable first-level directory structure oriented to the main scene of the project management backend.

#### Scenario: Query the default background menu skeleton
- **WHEN** The host projects the default background menu for the current user
- **THEN** The first-level directory is organized according to the following structure: `Workbench`, `Permission Management`, `Organization Management`, `System Settings`, `Content Management`, `System Monitoring`, `Task Scheduling`, `Extension Center`, `Development Center`
- **AND** The stable host parent `menu_key` corresponding to these first-level directories is exactly `dashboard`, `iam`, `org`, `setting`, `content`, `monitor`, `scheduler`, `extension`, `developer`

#### Scenario: The first-level directory exists as a host stable directory record
- **WHEN** The host initializes or synchronizes the default background menu skeleton
- **THEN** These first-level directories are created and owned by the host
- **AND** The plugin can only mount submenus to these directories instead of creating new sibling first-level directories.

#### Scenario: Default background extension business module
- **WHEN** Developers continue to add business modules or official source plugins to the project
- **THEN** New menus will be placed in existing stable directories first.
- **AND** No need to frequently restructure first-level navigation naming and structure

### Requirement: The plugin menu is semantically mounted to the host directory

The system SHALL requires the plugin menu to fall in the semantically corresponding host directory, rather than being unified into the plugin management directory.

#### Scenario: Organize plugin mounting menu
- **WHEN** `org-center` synchronizes its menu to the host
- **THEN** Its menu is mounted to `Organization Management`
- **AND** Do not mount to `Extension Center`

#### Scenario: Content plugin mounting menu
- **WHEN** `content-notice` synchronizes its menu to the host
- **THEN** whose menu is mounted to `Content Management`
- **AND** Do not mount to `Extension Center`

#### Scenario: Monitoring plugin mounting menu
- **WHEN** `monitor-online`, `monitor-server`, `monitor-operlog` or `monitor-loginlog` sync menu to host
- **THEN** These menus are mounted to `System Monitor`
- **AND** `Extension Center/Plug-in Management` is still responsible for installation and start-up and stop management

### Requirement: Empty parent directories are automatically hidden.

The system SHALL automatically hides a directory at a certain level when there is no visible submenu to avoid empty shell navigation in the default background.

#### Scenario: No visible menu for content management
- **WHEN** `content-notice` is not installed, not enabled, or the current user does not have access to its menu
- **THEN** `Content Management` does not appear in the left navigation

#### Scenario: Some system monitoring plugins are missing
- **WHEN** Only some monitoring plugins are installed and visible under `System Monitoring`
- **THEN** The left navigation only displays visible monitoring submenus
- **AND** The parent directory `System Monitor` continues to be retained
- **AND** If all monitoring submenus are invisible, the parent directory will be hidden as well.
