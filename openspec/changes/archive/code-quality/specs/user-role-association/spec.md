## ADDED Requirements

### Requirement: sys_user_role must include role_id reverse index

The system SHALL maintain `KEY idx_role_id (role_id)` on the `sys_user_role` table to support access paths such as querying users by role and deleting associations by role in batch. This avoids full table scans when only the `(user_id, role_id)` composite primary key exists.

#### Scenario: sys_user_role reverse index exists

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_user_role` MUST include `idx_role_id` on column `role_id`

#### Scenario: Query users by role uses the index

- **WHEN** the service executes queries of the form `WHERE role_id = ?`
- **THEN** the database MUST select `idx_role_id` to avoid a full table scan

## MODIFIED Requirements

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
