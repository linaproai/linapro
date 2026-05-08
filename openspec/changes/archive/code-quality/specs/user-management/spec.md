## ADDED Requirements

### Requirement: User batch delete

The system SHALL provide a RESTful batch delete endpoint to remove multiple users in a single request, sharing the same protection rules and atomicity as single-user delete.

#### Scenario: Successful batch delete

- **WHEN** a caller with `user:remove` permission invokes `DELETE /api/v1/user?ids=2,3,4`
- **AND** none of the ids match the built-in admin or the current logged-in user
- **THEN** the system soft-deletes all specified users in a single transaction
- **AND** clears their organization assignments and `sys_user_role` associations atomically
- **AND** returns success
- **AND** access topology is notified once after the transaction commits

#### Scenario: Batch delete rejects built-in admin id

- **WHEN** the caller invokes `DELETE /api/v1/user?ids=1&ids=2&ids=3`
- **AND** id `1` belongs to the built-in admin
- **THEN** the entire batch MUST be rejected with `bizerr` `CodeUserBuiltinAdminDeleteDenied`
- **AND** no user is deleted, no association is cleaned

#### Scenario: Batch delete rejects current user id

- **WHEN** the caller invokes `DELETE /api/v1/user?ids=...`
- **AND** the id list contains the current logged-in user's id
- **THEN** the entire batch MUST be rejected with `bizerr` `CodeUserCurrentDeleteDenied`
- **AND** no user is deleted

#### Scenario: Empty id list rejected at validation

- **WHEN** the caller invokes `DELETE /api/v1/user?ids=`
- **THEN** the system MUST reject the request with a validation error
- **AND** no transaction is started

### Requirement: sys_user table must carry common query indexes

The system SHALL maintain `KEY idx_status (status)`, `KEY idx_phone (phone)`, and `KEY idx_created_at (created_at)` on the `sys_user` table so that user list queries filtering by status, phone, or created-time range avoid full table scans.

#### Scenario: sys_user indexes present after init

- **WHEN** `make init` finishes initializing the database
- **THEN** `SHOW INDEX FROM sys_user` returns entries `idx_status`, `idx_phone`, and `idx_created_at` in addition to the existing primary key and `username` unique key

## MODIFIED Requirements

### Requirement: Delete user

The system SHALL support deleting a single user with full transactional cleanup of all associated data. Soft-deleting the user record, removing organization assignments (when `org-center` is installed and enabled), and removing entries in `sys_user_role` MUST occur within a single database transaction. Any failure in associated cleanup MUST cause the entire deletion to roll back. Access topology change notification MUST be issued only after the transaction successfully commits.

#### Scenario: Delete user atomically cleans associated data

- **WHEN** the caller deletes a user
- **AND** all cleanup steps succeed
- **THEN** the system soft-deletes the user record
- **AND** when `org-center` is installed and enabled, removes department/position assignments
- **AND** removes the matching `sys_user_role` rows
- **AND** notifies access topology change after commit

#### Scenario: Association cleanup failure rolls back user deletion

- **WHEN** the caller deletes a user
- **AND** organization or `sys_user_role` cleanup fails inside the transaction
- **THEN** the user soft-delete MUST be rolled back
- **AND** the operation returns the underlying error
- **AND** no access topology notification is issued
