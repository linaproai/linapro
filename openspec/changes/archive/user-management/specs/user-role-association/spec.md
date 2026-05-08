## ADDED Requirements

### Requirement: User List Displays Role Information

The system MUST display each user's associated role information in the user list.

#### Scenario: User list contains role column
- **WHEN** a user visits the user management page
- **THEN** the user list contains a "role" column
- **AND** the role column shows all associated role names separated by commas
- **AND** users with no roles show empty or "unassigned"

#### Scenario: User detail contains role information
- **WHEN** a user views user details
- **THEN** the system returns roleIds (role ID array) and roleNames (role name array)

### Requirement: Assign Roles During User Creation

The system MUST support selecting associated roles when creating a user.

#### Scenario: Create user and assign roles
- **WHEN** a user selects one or more roles in the create user form and submits
- **THEN** the system creates the user record
- **AND** the system inserts user-role association records in sys_user_role

#### Scenario: Create user without role assignment
- **WHEN** a user does not select any roles in the create user form and submits
- **THEN** the system creates the user record
- **AND** no records are inserted in sys_user_role

### Requirement: Modify Roles During User Update

The system MUST support modifying associated roles when updating a user.

#### Scenario: Update user roles
- **WHEN** a user modifies role selection in the edit user form and submits
- **THEN** the system updates user basic information
- **AND** the system deletes old sys_user_role association records
- **AND** the system inserts new sys_user_role association records

#### Scenario: Clear user roles
- **WHEN** a user removes all role selections in the edit user form and submits
- **THEN** the system updates user basic information
- **AND** the system deletes all sys_user_role records for this user

### Requirement: Clean Role Associations on User Deletion

The system MUST clean role association data when deleting a user.

#### Scenario: Delete user cleanup role associations
- **WHEN** a user deletes a user
- **THEN** the system soft-deletes the user record
- **AND** the system deletes all sys_user_role records for this user

### Requirement: Role Dropdown Options Loading

The system MUST provide role dropdown options in the user form.

#### Scenario: Load role dropdown options
- **WHEN** a user opens the create/edit user form
- **THEN** the system requests the role dropdown options endpoint
- **AND** the form's role selector shows all enabled roles

#### Scenario: Role multi-select
- **WHEN** a user selects multiple roles in the role selector
- **THEN** the selector displays selected roles as tags
- **AND** the user can remove selected roles
