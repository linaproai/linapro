# Menu Management

## Purpose

The menu management function is responsible for the creation, editing, deletion, tree display and permission association management of system menus, ensuring that frontend routing, menu permissions and host menu configurations are consistent.
## Requirements
### Requirement: Menu list query

System SHALL supports tree query of menu list, returning all menus and their hierarchical relationships.

#### Scenario: Query menu tree list
- **WHEN** User accesses the menu management page
- **THEN** The system returns a menu tree list, including all directories, menus, and button types
- **AND** Each menu node contains information such as id, parentId, name, path, icon, type, sort, visible, status, etc.
- **AND** The submenu is nested in the children field of the parent menu

#### Scenario: Filter menu by criteria
- **WHEN** User enters menu name to search
- **THEN** The system returns a menu list whose name contains keywords (fuzzy matching)
- **AND** The returned result still maintains the tree structure

### Requirement: Menu details query

System SHALL supports querying menu details based on menu ID.

#### Scenario: Query existing menu details
- **WHEN** User clicks the edit button in the menu list
- **THEN** The system returns the complete information of the menu, including the name of the parent menu

#### Scenario: Query a menu that does not exist
- **WHEN** The user requested a menu ID that does not exist
- **THEN** The system returns the error message "Menu does not exist"

### Requirement: Create menu

System SHALL supports the creation of new menus, including three types: directory, menu, and button.

#### Scenario: Create catalog type menu
- **WHEN** The user fills in the directory information (name, icon, sorting, whether to display, status) and submits
- **THEN** The system creates directory type menu, type is "D"
- **AND** The system automatically sets created_at and updated_at
- **AND** path field of directory type is used for routing grouping

#### Scenario: Create menu type menu
- **WHEN** The user fills in the menu information (name, routing address, component path, permission identifier, icon, sorting, whether to display, whether to cache, status) and submit
- **THEN** The system creates menu type menu, type is "M"
- **AND** menu type MUST have path and component fields

#### Scenario: Create button type menu
- **WHEN** The user fills in the button information (name, permission ID, upper-level menu) and submits
- **THEN** The system creates a button type menu, type is "B"
- **AND** button type has no path, component, icon and other fields

#### Scenario: Create external link menu
- **WHEN** User creates menu and sets is_frame to 1
- **THEN** The system treats path as an external link address
- **AND** When the frontend clicks on the menu, a new window opens the external link

#### Scenario: Duplicate menu name
- **WHEN** Use the existing menu name when the user creates the menu
- **THEN** The system returns the error message "Menu name already exists"

### Requirement: Update menu

System SHALL supports updating menu information.

#### Scenario: Update menu information
- **WHEN** The user modifies the menu information and submits it
- **THEN** updated_at timestamp of system update menu
- **AND** All editable fields of the system update menu

#### Scenario: Update non-existent menu
- **WHEN** The user attempted to update a menu that does not exist
- **THEN** The system returns the error message "Menu does not exist"

#### Scenario: Upper menu selection restrictions when editing menus
- **WHEN** Open the upper-level menu selector when the user edits the menu
- **THEN** The system will gray out the current menu and all its descendant menus in the tree list and disable it.
- **AND** Disabled nodes cannot be selected
- **AND** There is no disabled node when adding a menu

#### Scenario: Select the upper level menu when adding a new menu
- **WHEN** Open the upper-level menu selector when the user adds a menu
- **THEN** All menu nodes are selectable
- **AND** No disabled nodes

### Requirement: Delete menu

System SHALL supports deleting menus and cascading deletion of submenus.

#### Scenario: Remove menus without submenus
- **WHEN** User deletes a menu that has no submenus
- **THEN** The system soft deletes this menu (set deleted_at)
- **AND** Synchronously delete the associated records of this menu in sys_role_menu

#### Scenario: Delete menu with submenus
- **WHEN** The user deletes a menu with submenus
- **THEN** The system prompts "Submenu exists, do you want to delete it cascaded?"
- **AND** After user confirmation, delete this menu and all its submenus
- **AND** Synchronously delete all associated character menu relationships

### Requirement: Get the menu drop-down tree

System SHALL provides a menu drop-down tree interface for selection in the role assignment menu.

#### Scenario: Get menu drop-down tree
- **WHEN** Request menu tree when role assignment menu
- **THEN** The system returns to the tree menu list
- **AND** Filter out menus with button type (type="B")
- **AND** Each node contains id, parentId, label, children

### Requirement: Get the role's menu tree

System SHALL provides an interface to obtain the menu tree of a specified role, which is used to display the assigned menu when editing the role.

#### Scenario: Get the menu permissions of the character
- **WHEN** Request the character's menu tree when editing a character
- **THEN** The system returns to all menu trees
- **AND** Returns the list of menu IDs assigned to this role (checkedKeys)
- **AND** When the menu tree is displayed, the assigned menu is checked

### Requirement: Menu status control

System SHALL supports menu enable/disable status switching.

#### Scenario: Disable menu
- **WHEN** User changes menu status to disabled
- **THEN** This menu will not appear in the frontend menu bar
- **AND** When accessing this menu route directly via URL, the frontend should deny access

#### Scenario: Hide menu
- **WHEN** The user sets the visible field of the menu to hidden
- **THEN** This menu will not appear in the frontend menu bar
- **AND** Users can still access the menu directly via the URL

### Requirement: Menu management displays plugin menu ownership information

System SHALL displays the ownership and life cycle status of plugin menus in the menu management capability to facilitate unified management by administrators.

#### Scenario: View plugin menu ownership

- **WHEN** The administrator views a menu registered by the plugin on the menu management page
- **THEN** The system can identify the plugin to which this menu belongs, the source type and the current plugin status
- **AND** Administrators can differentiate between host menu and plugin menu

### Requirement: Plug-in status linkage menu visibility and accessibility

System SHALL jointly controls the display and access behavior of the plugin's menu when it is deactivated, uninstalled, or the upgrade does not take effect.

#### Scenario: Plugin disabled

- **WHEN** A plugin is disabled
- **THEN** The menu registered by this plugin will not appear in the frontend menu bar
- **AND** Users will receive controlled unavailability feedback when accessing the plugin page route directly.

#### Scenario: Plugin re-enabled

- **WHEN** A disabled plugin is re-enabled
- **THEN** The original menu can reappear in the frontend menu bar according to the authorization relationship
- **AND** Administrators do not need to recreate menu records

#### Scenario: Administrator modifies menu display status

- **WHEN** The administrator changes a menu visible to the current user to hidden on the menu management page
- **THEN** The left navigation of the currently logged in user will remove the menu immediately after this operation is completed.
- **AND** If the administrator restores the display again, the left navigation will immediately restore the menu under the same refresh link

#### Scenario: Menu or plugin state changes occur again during refresh

- **WHEN** The host is refreshing the accessible menus and dynamic routes for the currently logged in user, and during this period it receives new menu visibility or plugin status changes.
- **THEN** The host will make up a round of the latest status immediately after the current refresh is completed.
- **AND** The left navigation finally converges to the menu result after the latest operation instead of staying in the old state

### Requirement: The plugin management menu uses stable business identifiers

System SHALL provides a stable business identifier `menu_key` for menus governed by plugins, and uses this as the main anchor for menu life cycle management.

#### Scenario: Identify plugin menu ownership

- **WHEN** The host needs to determine whether a menu belongs to a certain plugin
- **THEN** Host priority is based on `sys_menu.menu_key` to parse menu ownership
- **AND** `remark` is only used as a remark display field, not as a formal function identifier

#### Scenario: Plug-in menu installation or upgrade

- **WHEN** Dynamic plugin installer or source plugin synchronization process creates or updates menus based on manifest menu metadata
- **THEN** The host presses `menu_key` to perform idempotent writing
- **AND** The host parses the real `parent_id` based on `parent_key`
- **AND** The same `menu_key` remains unique within the host scope

#### Scenario: Plug-in uninstallation menu metadata cleanup menu

- **WHEN** Dynamic plugin uninstallation or host execution plugin menu cleaning
- **THEN** The host deletes the corresponding menu and role association based on the `menu_key` declared in the manifest
- **AND** The host does not rely on plugin uninstallation and SQL manual maintenance menu deletion statements.

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

### Requirement: Menu capability must return localized titles for the current language
The system SHALL return localized menu titles in menu trees, parent menu display, role menu trees, and dynamic route projections according to the current request language. Menu localization MUST derive translation keys from stable `menu_key` anchors. When the current language lacks a translation, the system MUST fall back to the default language or the menu default name. The directly editable `name` field in menu edit forms MUST keep the database value to avoid writing display titles back into governance data during language switching.

#### Scenario: Menu tree returns English titles
- **WHEN** a user requests the menu tree or current-user dynamic routes with `en-US`
- **THEN** menu titles in the response use localized values for that language
- **AND** the same menu remains consistent between tree node names and route `meta.title`
- **AND** English menu titles use concise, natural product wording instead of literal translations from Chinese names

#### Scenario: Button titles under resource menus use short action words
- **WHEN** a user views button menus in the menu management page, parent menu tree, or role menu tree with `en-US`
- **THEN** button titles under resource menus use short action words such as `Query`, `Create`, `Update`, `Delete`, and `Export`
- **AND** button titles avoid repeating the parent resource menu name, such as `Users / Create` instead of `Users / Create User`
- **AND** buttons that cannot be clearly expressed by generic actions may keep short business actions such as `Reset Password`, `Force Logout`, or `Run Now`

#### Scenario: Missing menu translations fall back to default names
- **WHEN** a menu has no translation for the current language
- **THEN** the system falls back to the default-language title or menu default name
- **AND** menu structure, permissions, and sort order remain unaffected

#### Scenario: Administrator edits menus in English
- **WHEN** an administrator opens menu detail or edit drawer in an `en-US` environment
- **THEN** the menu list tree, parent menu display name, and parent menu selector tree show localized titles for the current language
- **AND** the editable `name` field in the form keeps the original database value

### Requirement: Menu management must support stable business keys for i18n copy
The system SHALL preserve stable business keys as i18n anchors in menu governance and allow host and plugin resources to maintain menu titles through unified translation resources instead of hard-coding the same menu copy in multiple pages.

#### Scenario: Plugin menus integrate with i18n resources
- **WHEN** a plugin declares a menu with a stable `menu_key` and provides matching translation resources
- **THEN** the host uses the menu's localized title in the menu management page, left navigation, and role authorization tree
- **AND** administrators do not need to configure multiple translation mappings for the same plugin menu

#### Scenario: Administrator views menu detail
- **WHEN** an administrator views a menu detail in the current language
- **THEN** the system returns the localized parent menu display title and related read-only copy for the current language
- **AND** the menu's editable fields continue to keep database values while the stable business key remains available for later translation maintenance and diagnostics

