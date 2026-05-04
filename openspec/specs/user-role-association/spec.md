# User Role Association Specification

## Purpose

Define user-role association management, role assignment, and role option behavior so user permission ownership is consistently maintained in user-management workflows.

## Requirements

### Requirement: User list displays role information

The system SHALL display role information associated with each user in the user list.

#### Scenario: User list contains role column

- **WHEN** a user opens user management
- **THEN** the user list includes a Role column
- **AND** the Role column displays all associated role names separated by commas
- **AND** users without assigned roles display an empty value or an unassigned state

#### Scenario: User details include role information

- **WHEN** a user views user details
- **THEN** the system returns associated role ID list `roleIds` and role name list `roleNames`

### Requirement: Assign roles when creating users

The system SHALL support selecting associated roles when creating a user.

#### Scenario: Create user and assign roles

- **WHEN** a user selects one or more roles in the create-user form and submits
- **THEN** the system creates the user record
- **AND** inserts user-role association records into `sys_user_role`

#### Scenario: Create user without assigning roles

- **WHEN** a user submits the create-user form without selecting roles
- **THEN** the system creates the user record
- **AND** no association record for that user exists in `sys_user_role`

### Requirement: Modify roles when updating users

The system SHALL support modifying associated roles when updating a user.

#### Scenario: Update user roles

- **WHEN** a user changes role selection in the edit-user form and submits
- **THEN** the system updates user base information
- **AND** deletes old `sys_user_role` association records
- **AND** inserts new `sys_user_role` association records

#### Scenario: Clear user roles

- **WHEN** a user clears all role selection in the edit-user form and submits
- **THEN** the system updates user base information
- **AND** deletes all `sys_user_role` association records for that user

### Requirement: Clean role associations when deleting users

The system SHALL clean role association data when deleting users. User soft delete and `sys_user_role` association cleanup MUST run inside the same transaction, and any failure MUST roll back the whole operation.

#### Scenario: User deletion transactionally cleans role associations

- **WHEN** a user deletes another user
- **AND** all cleanup steps succeed
- **THEN** the system soft-deletes the user record inside the transaction
- **AND** deletes all `sys_user_role` associations for that user
- **AND** notifies access topology change only after transaction commit

#### Scenario: Association cleanup failure rolls back the whole operation

- **WHEN** a user deletes another user
- **AND** `sys_user_role` association cleanup returns an error inside the transaction
- **THEN** user soft delete MUST roll back as well
- **AND** the operation returns the underlying error

### Requirement: Role option loading

The system SHALL provide role options in user forms.

#### Scenario: Load role options

- **WHEN** a user opens the create-user or edit-user form
- **THEN** the system requests role options
- **AND** the role selector displays all enabled roles

#### Scenario: Select multiple roles

- **WHEN** a user selects multiple roles in the role selector
- **THEN** the selector displays selected roles as tags
- **AND** the user can remove selected roles

### Requirement: sys_user_role must include role_id reverse index

The system SHALL maintain `KEY idx_role_id (role_id)` on the `sys_user_role` table to support access paths such as querying users by role and deleting associations by role in batch. This avoids full table scans when only the `(user_id, role_id)` composite primary key exists.

#### Scenario: sys_user_role reverse index exists

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_user_role` MUST include `idx_role_id` on column `role_id`

#### Scenario: Query users by role uses the index

- **WHEN** the service executes queries of the form `WHERE role_id = ?`
- **THEN** the database MUST select `idx_role_id` to avoid a full table scan
