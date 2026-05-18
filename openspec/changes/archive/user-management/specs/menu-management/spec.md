## ADDED Requirements

### Requirement: Menu List Query

The system MUST support tree-structured menu list queries returning all menus with their hierarchy.

#### Scenario: Query menu tree list
- **WHEN** a user visits the menu management page
- **THEN** the system returns a tree-structured menu list including all directory, menu, and button types
- **AND** each menu node contains id, parentId, name, path, icon, type, sort, visible, status, etc.
- **AND** child menus are nested in the parent menu's children field

#### Scenario: Filter menus by condition
- **WHEN** a user enters a menu name for search
- **THEN** the system returns menus whose name contains the keyword (fuzzy match)
- **AND** the results maintain tree structure

### Requirement: Menu Detail Query

The system MUST support querying menu details by menu ID.

#### Scenario: Query existing menu detail
- **WHEN** a user clicks the edit button in the menu list
- **THEN** the system returns the complete menu information including parent menu name

#### Scenario: Query non-existent menu
- **WHEN** a user requests a non-existent menu ID
- **THEN** the system returns an error message "menu does not exist"

### Requirement: Create Menu

The system MUST support creating new menus including directory, menu, and button types.

#### Scenario: Create directory type menu
- **WHEN** a user fills in directory information (name, icon, sort, visibility, status) and submits
- **THEN** the system creates a directory type menu with type "D"
- **AND** the system automatically sets created_at and updated_at
- **AND** the directory type's path field is used for route grouping

#### Scenario: Create menu type menu
- **WHEN** a user fills in menu information (name, route address, component path, permission key, icon, sort, visibility, cache flag, status) and submits
- **THEN** the system creates a menu type menu with type "M"
- **AND** menu type must have path and component fields

#### Scenario: Create button type menu
- **WHEN** a user fills in button information (name, permission key, parent menu) and submits
- **THEN** the system creates a button type menu with type "B"
- **AND** button type has no path, component, or icon fields

#### Scenario: Create external link menu
- **WHEN** a user creates a menu and sets is_frame to 1
- **THEN** the system treats the path as an external link address
- **AND** clicking the menu in the frontend opens the external link in a new window

#### Scenario: Duplicate menu name
- **WHEN** a user creates a menu with an existing name
- **THEN** the system returns an error message "menu name already exists"

### Requirement: Update Menu

The system MUST support updating menu information.

#### Scenario: Update menu information
- **WHEN** a user modifies menu information and submits
- **THEN** the system updates the menu's updated_at timestamp
- **AND** the system updates all editable fields of the menu

#### Scenario: Update non-existent menu
- **WHEN** a user attempts to update a non-existent menu
- **THEN** the system returns an error message "menu does not exist"

#### Scenario: Parent menu selection restriction when editing
- **WHEN** a user opens the parent menu selector while editing a menu
- **THEN** the system grays out and disables the current menu and all its descendants in the tree
- **AND** disabled nodes cannot be selected
- **AND** no nodes are disabled when adding a new menu

#### Scenario: Parent menu selection when adding
- **WHEN** a user opens the parent menu selector while adding a new menu
- **THEN** all menu nodes are selectable
- **AND** no nodes are disabled

### Requirement: Delete Menu

The system MUST support deleting menus with cascade deletion of child menus.

#### Scenario: Delete menu without children
- **WHEN** a user deletes a menu that has no children
- **THEN** the system soft-deletes the menu (sets deleted_at)
- **AND** the system deletes the menu's association records from sys_role_menu

#### Scenario: Delete menu with children
- **WHEN** a user deletes a menu that has children
- **THEN** the system prompts "child menus exist, cascade delete?"
- **AND** after user confirmation, deletes the menu and all its children
- **AND** deletes all related role-menu associations

### Requirement: Menu Dropdown Tree

The system MUST provide a menu dropdown tree endpoint for role menu assignment.

#### Scenario: Get menu dropdown tree
- **WHEN** role menu assignment requests the menu tree
- **THEN** the system returns a tree-structured menu list
- **AND** filters out button type (type="B") menus
- **AND** each node contains id, parentId, label, children

### Requirement: Role Menu Tree

The system MUST provide an endpoint to get a specific role's menu tree for displaying assigned menus during role editing.

#### Scenario: Get role menu permissions
- **WHEN** editing a role and requesting its menu tree
- **THEN** the system returns the complete menu tree
- **AND** returns the role's assigned menu ID list (checkedKeys)
- **AND** assigned menus are shown as checked in the menu tree

### Requirement: Menu Status Control

The system MUST support enabling/disabling menu status.

#### Scenario: Disable menu
- **WHEN** a user changes a menu's status to disabled
- **THEN** the menu does not appear in the frontend menu bar
- **AND** direct URL access to the menu route is rejected by the frontend

#### Scenario: Hide menu
- **WHEN** a user sets a menu's visible field to hidden
- **THEN** the menu does not appear in the frontend menu bar
- **AND** the user can still access the menu via direct URL
