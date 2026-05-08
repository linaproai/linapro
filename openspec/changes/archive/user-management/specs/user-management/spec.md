## MODIFIED Requirements

### Requirement: User List Query

The system MUST support paginated user queries including role information.

#### Scenario: Query user list
- **WHEN** a user visits the user management page
- **THEN** the system returns a paginated user list
- **AND** each user contains id, username, nickname, email, phone, sex, avatar, status, deptId, deptName, roleIds, roleNames, etc.

#### Scenario: User list includes role information
- **WHEN** a user requests the user list
- **THEN** the system queries user-role associations from sys_user_role
- **AND** the system queries role names from sys_role
- **AND** roleIds is a role ID array
- **AND** roleNames is a role name array

### Requirement: User Detail Query

The system MUST return user details with associated role information.

#### Scenario: Query user detail
- **WHEN** a user requests user details
- **THEN** the system returns user basic information
- **AND** the system returns deptId, deptName
- **AND** the system returns postIds (post ID array)
- **AND** the system returns roleIds (role ID array)

### Requirement: Create User

The system MUST support creating users with role association.

#### Scenario: Create user with role association
- **WHEN** a user submits the create user form
- **THEN** the system creates the user record
- **AND** if deptId is provided, the system inserts association in sys_user_dept
- **AND** if postIds are provided, the system inserts associations in sys_user_post
- **AND** if roleIds are provided, the system inserts associations in sys_user_role

#### Scenario: Create user without role assignment
- **WHEN** a user submits the create user form without selecting any roles
- **THEN** the system creates the user record
- **AND** no records are inserted in sys_user_role

### Requirement: Update User

The system MUST support updating users with role association modification.

#### Scenario: Update user role association
- **WHEN** a user submits the update user form
- **THEN** the system updates user basic information
- **AND** if deptId is provided, the system updates sys_user_dept association
- **AND** if postIds are provided, the system updates sys_user_post associations
- **AND** if roleIds are provided, the system deletes old sys_user_role records and inserts new ones

#### Scenario: Clear user roles
- **WHEN** a user removes all role selections and submits
- **THEN** the system updates user basic information
- **AND** the system deletes all sys_user_role records for this user

#### Scenario: Update user transaction handling
- **WHEN** an error occurs during user update
- **THEN** the system rolls back all operations (user info, dept association, post association, role association)

### Requirement: Delete User

The system MUST support deleting users with cleanup of all association data.

#### Scenario: Delete user cleanup associations
- **WHEN** a user deletes a user
- **THEN** the system soft-deletes the user record
- **AND** the system deletes sys_user_dept association records
- **AND** the system deletes sys_user_post association records
- **AND** the system deletes sys_user_role association records

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
