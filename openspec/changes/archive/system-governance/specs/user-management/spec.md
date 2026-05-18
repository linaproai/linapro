## MODIFIED Requirements

### Requirement: User List Query

The system SHALL provide a paginated user list query interface, supporting multi-field sorting, enhanced condition filtering, and role information aggregation. When `org-center` is installed and enabled, the system additionally supports filtering by department and returns department fields; when the plugin is missing, the host ignores organization extension filtering and keeps the user list main function available. The user list SHALL integrate host data permission governance: all-data scope can query all users; department-data scope only queries users within the current user's department scope; self-only scope only queries the current user's own user record.

#### Scenario: Filter User List by Department When Org Plugin Available

- **WHEN** `org-center` is installed and enabled, and `deptId` is passed in the query
- **THEN** the system filters users belonging to that department through the org plugin's organizational relationships
- **AND** returned user data may contain `deptId` and `deptName` fields

#### Scenario: Query User List When Org Plugin Missing

- **WHEN** `org-center` is not installed or not enabled, and user list is queried
- **THEN** the system still returns user pagination list and role information
- **AND** department-related filters and fields are safely ignored or omitted

#### Scenario: Department-Data Scope Restricts User List

- **WHEN** a normal user's role data scope is department data
- **AND** `org-center` is installed and enabled
- **THEN** the user list only returns users within the current user's department scope
- **AND** even if the request parameters do not include `deptId`, the system still applies this data permission boundary

#### Scenario: Self-Only Scope Restricts User List

- **WHEN** a normal user's role data scope is self-only data
- **THEN** the user list only returns the current logged-in user's own record

#### Scenario: Department Scope Overlaps Explicit Department Filter

- **WHEN** a normal user's role data scope is department data
- **AND** the query parameter passes a `deptId` not within the current user's visible scope
- **THEN** the system returns an empty list
- **AND** does not return users from other departments

### Requirement: View User Detail

The system SHALL provide a user detail query interface. When `org-center` is installed and enabled, it returns associated department and post information; when the plugin is missing, it still returns basic user information and role information. User detail SHALL integrate host data permission governance; the caller can only view target users within their data permission scope.

#### Scenario: Query User Detail When Org Plugin Missing

- **WHEN** `org-center` is not installed or not enabled and `GET /api/v1/user/{id}` is called
- **THEN** the system returns the user's complete basic information (excluding password) and role information
- **AND** `deptId`, `deptName`, `postIds` and other organization extension fields are omitted, set to zero values, or set to empty sets

#### Scenario: Reject Viewing Out-of-Scope User Detail

- **WHEN** a normal user's role data scope is self-only data
- **AND** calls `GET /api/v1/user/{id}` to query another user
- **THEN** the system returns a structured data-not-visible error
- **AND** does not return the target user's detail

#### Scenario: Department User Detail Visible

- **WHEN** a normal user's role data scope is department data
- **AND** the target user is within the current user's visible department scope
- **THEN** the system returns the target user's detail

### Requirement: Update User Information

The system SHALL provide an update user information interface that always supports role association; when `org-center` is installed and enabled, the system additionally supports updating department and post associations; when the plugin is missing, these organization extension fields do not block user update. User update, status change, password reset, and role association change SHALL integrate host data permission governance; the caller can only modify target users within their data permission scope; existing built-in admin, current user deletion protection, and transaction rules continue to apply.

#### Scenario: Update User When Org Plugin Missing

- **WHEN** `org-center` is not installed or not enabled and admin updates a user
- **THEN** the system still successfully updates basic user information and role association
- **AND** department and post related fields are safely ignored

#### Scenario: Reject Updating Out-of-Scope User

- **WHEN** a normal user's role data scope is self-only data
- **AND** the user attempts to update another user's basic information, status, password, or role association
- **THEN** the system returns a structured data-not-visible error
- **AND** the target user record and associations remain unchanged

#### Scenario: Department Scope Allows Updating Department Users

- **WHEN** a normal user's role data scope is department data
- **AND** the target user is within the current user's visible department scope
- **THEN** the system can update that user after existing function permission and protection rules pass

### Requirement: Delete User

The system SHALL support deleting a single user with complete transactional cleanup of all associated data. Soft-deleting the user record, removing organizational assignments (when `org-center` is installed and enabled), and removing entries in `sys_user_role` must occur within a single database transaction. Any failure in associated cleanup must cause the entire delete to roll back. Access topology change notification must only be issued after the transaction commits successfully. Delete user SHALL integrate host data permission governance; the caller can only delete target users within their data permission scope, and still cannot delete built-in admin or current logged-in user.

#### Scenario: Delete User with Atomic Cleanup of Associated Data

- **WHEN** caller deletes a user
- **AND** all cleanup steps succeed
- **THEN** the system soft-deletes the user record
- **AND** when `org-center` is installed and enabled, removes department/post assignments
- **AND** removes matching `sys_user_role` rows
- **AND** notifies access topology change after commit

#### Scenario: Rollback User Delete on Associated Cleanup Failure

- **WHEN** caller deletes a user
- **AND** organization or `sys_user_role` cleanup fails within the transaction
- **THEN** the user soft-delete must be rolled back
- **AND** the operation returns the underlying error
- **AND** no access topology notification is issued

#### Scenario: Reject Deleting Out-of-Scope User

- **WHEN** a normal user's role data scope is department data
- **AND** the target user is not within the current user's visible department scope
- **THEN** the system rejects the delete
- **AND** does not clean up any associated data for the target user

### Requirement: User Batch Delete

The system SHALL provide a RESTful batch delete endpoint that deletes multiple users in a single request, sharing the same protection rules, data permission boundaries, and atomicity as single user delete.

#### Scenario: Batch Delete Success

- **WHEN** caller with `user:remove` permission calls `DELETE /api/v1/user?ids=2,3,4`
- **AND** no id matches built-in admin or current logged-in user
- **AND** all target users are within the current caller's data permission scope
- **THEN** the system soft-deletes all specified users in a single transaction
- **AND** atomically cleans up their organizational assignments and `sys_user_role` associations
- **AND** returns success
- **AND** notifies access topology once after transaction commit

#### Scenario: Batch Delete Rejects Built-in Admin id

- **WHEN** caller calls `DELETE /api/v1/user?ids=1&ids=2&ids=3`
- **AND** id `1` belongs to built-in admin
- **THEN** the entire batch must be rejected with `bizerr` `CodeUserBuiltinAdminDeleteDenied`
- **AND** no user is deleted, no associations are cleaned up

#### Scenario: Batch Delete Rejects Current User id

- **WHEN** caller calls `DELETE /api/v1/user?ids=...`
- **AND** the id list contains the current logged-in user's id
- **THEN** the entire batch must be rejected with `bizerr` `CodeUserCurrentDeleteDenied`
- **AND** no user is deleted

#### Scenario: Empty id List Rejected at Validation

- **WHEN** caller calls `DELETE /api/v1/user?ids=`
- **THEN** the system must reject the request with a validation error
- **AND** no transaction is started

#### Scenario: Batch Delete Rejects Out-of-Scope Users

- **WHEN** caller batch deletes multiple users
- **AND** any target user is not within the current caller's data permission scope
- **THEN** the entire batch delete is rejected
- **AND** no user is deleted
