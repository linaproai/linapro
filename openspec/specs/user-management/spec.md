# User Management

## Purpose

Define the query, maintenance, role association and collaboration rules for the `org-center` optional organizational capabilities of the host user management module to ensure that user management can work stably when the organization plugin is enabled or missing.
## Requirements
### Requirement: User list query
The system SHALL provides a paging query interface for user lists, supporting multi-field sorting, enhanced conditional filtering, and role information aggregation. When `org-center` is installed and enabled, the system additionally supports filtering by department and returns department fields; when the plugin is missing, the host ignores the organization extended filtering and keeps the user list main function available.

#### Scenario: Filter user list by department when organization plugin is available
- **WHEN** `org-center` is installed and enabled, and `deptId` is passed in when querying
- **THEN** The system filters users belonging to this department through the organizational relationship provided by the organization plugin.
- **AND** The returned user data can contain the `deptId` and `deptName` fields

#### Scenario: Query the user list when the organization plugin is missing
- **WHEN** `org-center` is not installed or enabled, and the user list is queried
- **THEN** The system still returns the user paginated list and role information
- **AND** Department-related filters and fields are safely ignored or omitted

### Requirement: Create user
The system SHALL provides a user interface for creation and always supports role association; when `org-center` is installed and enabled, the system additionally supports associated departments and positions; when the plugin is missing, these organization extension fields do not block user creation.

#### Scenario: Create users when organization plugin is missing
- **WHEN** `org-center` is not installed or enabled and the administrator created the user
- **THEN** The system still successfully created the user and processed the role association
- **AND** Lack of department and position information will not cause creation failure

### Requirement: Update user information
The system SHALL provides an interface for updating user information and always supports role association; when `org-center` is installed and enabled, the system additionally supports updating department and position associations; when the plugin is missing, these organization extension fields do not block user updates.

#### Scenario: Update users when organization plugin is missing
- **WHEN** `org-center` is not installed or enabled and the administrator updates the user
- **THEN** The system still successfully updated the user's basic information and role association
- **AND** Fields related to departments and positions are safely ignored

### Requirement: View user details
The system SHALL provides user details query interface. When `org-center` is installed and enabled, the associated department and position information is returned; when the plugin is missing, basic user information and role information are still returned.

#### Scenario: Query user details when the organization plugin is missing
- **WHEN** `org-center` is not installed or enabled and calling `GET /api/v1/user/{id}`
- **THEN** The system returns the complete basic information (excluding password) and role information of the user
- **AND** `deptId`, `deptName`, `postIds` and other organization extension fields are omitted, set to zero values or set to empty sets

### Requirement: Delete user

System SHALL supports clearing all associated data when deleting a user.

#### Scenario: Delete user to clean associated data
- **WHEN** User deletes a user
- **THEN** System soft delete user records
- **AND** When `org-center` is installed and enabled, the system deletes department and position related records in organizational capabilities
- **AND** The system deletes the associated records in sys_user_role

### Requirement: User department tree interface
The system SHALL provides a department tree interface for user management left filtering when `org-center` is installed and enabled; when the plugin is missing, the host no longer exposes the organization extension interface.

#### Scenario: Get user department tree when organization plugin is available
- **WHEN** `org-center` is installed and enabled, and calls `GET /api/v1/user/dept-tree`
- **THEN** The system returns department tree structure data, each node contains id, label, children, userCount
- **AND** The first level of the tree can still contain `Unassigned` virtual nodes

#### Scenario: User department tree is unavailable when the organization plugin is missing
- **WHEN** `org-center` is not installed or not enabled
- **THEN** The host no longer exposes `GET /api/v1/user/dept-tree` as the default user management dependency interface
- **AND** The user management main process does not depend on this interface to work properly.

### Requirement: User management frontend department tree filtering
The system SHALL only displays the `DeptTree` filter area on the left side of the user management page when `org-center` is installed and enabled; when the plugin is missing, the page degrades to a full-width user list.

#### Scenario: Page layout degraded when organization plugin is missing
- **WHEN** `org-center` is not installed or enabled, and the administrator opens the user management page
- **THEN** The page does not display the `DeptTree` component
- **AND** The user list area is displayed in a single-column full-width layout

### Requirement: User edit form to add department and position fields
The system SHALL only displays department selection and position multi-select fields in user edit forms when `org-center` is installed and enabled; these fields are hidden when the plugin is missing.

#### Scenario: Hide the department position field when the organization plugin is missing
- **WHEN** `org-center` is not installed or enabled and the administrator opens the user edit drawer
- **THEN** Department fields and position fields are not displayed in the form
- **AND** Users can still complete editing of basic information and role information

### Requirement: Add a department name column to the user list
The system SHALL only displays the department name column in the user list table when `org-center` is installed and enabled; the column is hidden when the plugin is missing.

#### Scenario: Hide department column when organization plugin is missing
- **WHEN** `org-center` is not installed or enabled and the administrator views the user list table
- **THEN** The `Department` column is not displayed in the table
- **AND** The remaining core user columns continue to display normally

### Requirement: User list role names must match backend-localized role display
The user management list SHALL use role display names returned by the backend and keep built-in role display consistent with role management in the same language.

#### Scenario: User list shows administrator role in English
- **WHEN** an administrator opens user management in `en-US`
- **THEN** the `admin` user's associated administrator role displays the same English name as role management
- **AND** the frontend does not maintain extra mappings based on Chinese role names or role keys

#### Scenario: Role selector keeps governance semantics
- **WHEN** an administrator opens the user create or edit form
- **THEN** the role selector continues to use backend role option data
- **AND** saving user-role relationships still submits stable role IDs rather than localized display text

