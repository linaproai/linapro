# Role Management

## Purpose

Define role query, maintenance, permission assignment, status control, user assignment, plugin menu authorization, and localized seed display behavior.

## Requirements

### Requirement: Role list and detail queries

The system SHALL provide paginated role list and detail APIs that return role identity, permission key, sort order, data scope, status, remark, timestamps, and assigned menu IDs where applicable.

#### Scenario: Query role list
- **WHEN** a user opens role management
- **THEN** the system returns paginated roles
- **AND** filtering by role name, permission key, and status is supported

#### Scenario: Query role details
- **WHEN** a user opens an existing role for editing
- **THEN** the system returns the role details and assigned menu IDs

### Requirement: Role create and update operations

The system SHALL support creating and updating roles while enforcing unique role names and permission keys.

#### Scenario: Create role
- **WHEN** a user submits valid role name, permission key, sort, data scope, status, and remark
- **THEN** the system creates the role and timestamps it

#### Scenario: Reject duplicate role identity
- **WHEN** a submitted role name or permission key already exists
- **THEN** the system returns a validation or business error

#### Scenario: Update role menu permissions
- **WHEN** a user updates menu assignments for a role
- **THEN** old `sys_role_menu` records are replaced by the new selection
- **AND** parent-child menu assignment semantics are preserved

### Requirement: Role deletion and status control

The system SHALL support deleting roles and toggling enabled or disabled status while preserving protected built-in role rules.

#### Scenario: Delete unassigned role
- **WHEN** a user deletes a role that is not assigned to users
- **THEN** the system soft-deletes the role
- **AND** related role-menu and user-role records are removed

#### Scenario: Delete role assigned to users
- **WHEN** a user deletes a role assigned to users
- **THEN** the system requires confirmation and then removes those assignments

#### Scenario: Protected administrator role cannot be deleted
- **WHEN** a user attempts to delete the built-in administrator role
- **THEN** the system rejects the deletion

#### Scenario: Disable role
- **WHEN** a role status changes to disabled
- **THEN** users associated with the role no longer obtain that role's menu permissions after login

### Requirement: Role option and user assignment capabilities

The system SHALL provide role option APIs for user management and role-user assignment workflows.

#### Scenario: Get role options
- **WHEN** the user form loads
- **THEN** the system returns enabled role options with `id`, `name`, and `key`

#### Scenario: View users assigned to a role
- **WHEN** a user clicks role assignment
- **THEN** the system shows assigned users with pagination and search

#### Scenario: Cancel or grant user authorization
- **WHEN** users are removed from or granted to a role
- **THEN** `sys_user_role` records are updated accordingly
- **AND** affected users receive permissions according to the new assignments after re-login

### Requirement: Data scope options must be supported

The system SHALL support simplified data scope values for all data, current department data, and own created data.

#### Scenario: Configure data scope
- **WHEN** a role data scope is set to one of the supported values
- **THEN** the system stores that value for later governance use

### Requirement: Roles must support plugin menu and permission authorization

The system SHALL allow administrators to assign plugin menus and plugin button permissions in the role authorization flow, and SHALL preserve authorization relationships while plugins are disabled.

#### Scenario: Assign plugin menus to a role
- **WHEN** an administrator opens the role menu authorization tree and a plugin is installed and enabled
- **THEN** plugin menus and button permissions appear in the assignable tree

#### Scenario: Plugin disabled keeps role authorization records
- **WHEN** a plugin with existing role authorizations is disabled
- **THEN** those roles temporarily stop receiving the plugin menus and permissions
- **AND** the system preserves the original authorization records for later recovery

### Requirement: Built-in role display must be localized consistently

Built-in protected roles and default seed roles SHALL display according to current language in framework-delivered pages. User management and role management MUST show the same localized display value for the same built-in role.

#### Scenario: Default user role displays in English
- **WHEN** an administrator opens role management in `en-US`
- **THEN** the default user role displays an English name
- **AND** it does not remain as the original Chinese seed value

#### Scenario: User management role display matches role management
- **WHEN** an administrator opens user management in `en-US`
- **THEN** the administrator user's role name matches the English role-management display
- **AND** the frontend does not maintain extra mappings based on Chinese names or role keys

