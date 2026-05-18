## ADDED Requirements

### Requirement: Role List Query

The system MUST support paginated role queries.

#### Scenario: Query role list
- **WHEN** a user visits the role management page
- **THEN** the system returns a paginated role list
- **AND** each role contains id, name, key, sort, dataScope, status, remark, createdAt, etc.
- **AND** supports filtering by role name, permission key, and status

#### Scenario: Filter roles by condition
- **WHEN** a user enters a role name or selects a status for search
- **THEN** the system returns matching role list

### Requirement: Role Detail Query

The system MUST support querying role details by role ID.

#### Scenario: Query existing role detail
- **WHEN** a user clicks the edit button in the role list
- **THEN** the system returns the complete role information
- **AND** returns the role's assigned menu ID list

#### Scenario: Query non-existent role
- **WHEN** a user requests a non-existent role ID
- **THEN** the system returns an error message "role does not exist"

### Requirement: Create Role

The system MUST support creating new roles.

#### Scenario: Create role
- **WHEN** a user fills in role information (name, permission key, sort, data scope, status, remark) and submits
- **THEN** the system creates a role record
- **AND** the system automatically sets created_at and updated_at

#### Scenario: Duplicate role name
- **WHEN** a user creates a role with an existing name
- **THEN** the system returns an error message "role name already exists"

#### Scenario: Duplicate permission key
- **WHEN** a user creates a role with an existing permission key
- **THEN** the system returns an error message "permission key already exists"

### Requirement: Update Role

The system MUST support updating role information.

#### Scenario: Update role basic information
- **WHEN** a user modifies role information and submits
- **THEN** the system updates the role's updated_at timestamp
- **AND** the system updates all editable fields of the role

#### Scenario: Update role menu permissions
- **WHEN** a user modifies role menu assignment and submits
- **THEN** the system deletes old sys_role_menu association records
- **AND** the system inserts new sys_role_menu association records
- **AND** menu assignment uses parent-child linkage mode

#### Scenario: Update non-existent role
- **WHEN** a user attempts to update a non-existent role
- **THEN** the system returns an error message "role does not exist"

### Requirement: Delete Role

The system MUST support deleting roles.

#### Scenario: Delete unassigned role
- **WHEN** a user deletes a role that is not assigned to any user
- **THEN** the system soft-deletes the role (sets deleted_at)
- **AND** deletes sys_role_menu and sys_user_role association records

#### Scenario: Delete assigned role
- **WHEN** a user attempts to delete a role assigned to users
- **THEN** the system prompts "this role is assigned to X users, confirm deletion?"
- **AND** after user confirmation, deletes the role and removes user assignments

#### Scenario: Delete super admin role
- **WHEN** a user attempts to delete the admin role
- **THEN** the system returns an error message "cannot delete super admin role"

### Requirement: Role Status Control

The system MUST support enabling/disabling role status.

#### Scenario: Disable role
- **WHEN** a user changes a role's status to disabled
- **THEN** users assigned this role cannot get its menu permissions on login

#### Scenario: Enable role
- **WHEN** a user changes a disabled role's status to enabled
- **THEN** users assigned this role can get its menu permissions after re-login

### Requirement: Role Dropdown Options

The system MUST provide a role dropdown options endpoint for user management role selection.

#### Scenario: Get role dropdown options
- **WHEN** the user management page loads the user form
- **THEN** the system returns all enabled roles
- **AND** each option contains id, name, key fields

### Requirement: Role User Assignment

The system MUST support assigning users to roles.

#### Scenario: View role's user list
- **WHEN** a user clicks the role's "assign" button
- **THEN** the system navigates to the role user management page
- **AND** displays users assigned to this role
- **AND** the user list supports pagination and search

#### Scenario: Remove user authorization
- **WHEN** a user clicks "remove authorization" in the role user list
- **THEN** the system deletes the corresponding sys_user_role record
- **AND** the user list automatically refreshes

#### Scenario: Batch authorize users
- **WHEN** a user selects multiple unassigned users and clicks "authorize"
- **THEN** the system batch-inserts sys_user_role records
- **AND** these users can get this role's menu permissions after re-login

### Requirement: Data Permission Scope

The system MUST support three simplified data-scope levels.

#### Scenario: Set all data permissions
- **WHEN** a role's dataScope is set to 1
- **THEN** the role can view all data (actual filtering logic not implemented yet)

#### Scenario: Set department data permissions
- **WHEN** a role's dataScope is set to 2
- **THEN** the role can only view own department data (actual filtering logic not implemented yet)

#### Scenario: Set self-only data permissions
- **WHEN** a role's dataScope is set to 3
- **THEN** the role can only view self-created data (actual filtering logic not implemented yet)
