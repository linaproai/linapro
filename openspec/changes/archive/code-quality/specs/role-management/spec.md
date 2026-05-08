## ADDED Requirements

### Requirement: Role batch delete

The system SHALL provide `DELETE /api/v1/role?ids=...` to delete roles in batch, reuse all single-delete protection rules, and guarantee whole-operation transaction atomicity.

#### Scenario: Delete multiple roles that are not assigned to users

- **WHEN** a caller with `system:role:remove` permission calls `DELETE /api/v1/role?ids=2,3,4`
- **AND** none of the target roles are assigned to users and none are the super administrator role
- **THEN** the system soft-deletes all roles inside one transaction
- **AND** synchronously deletes related `sys_role_menu` and `sys_user_role` association records
- **AND** triggers one access topology change notification after commit

#### Scenario: Batch containing the super administrator role is rejected

- **WHEN** a caller calls `DELETE /api/v1/role?ids=1&ids=2&ids=3`
- **AND** id `1` is the super administrator role
- **THEN** the entire batch MUST be rejected
- **AND** no role is deleted

#### Scenario: Empty id list is rejected during validation

- **WHEN** a caller calls `DELETE /api/v1/role?ids=`
- **THEN** the system returns a parameter validation error
- **AND** no transaction is started

## MODIFIED Requirements

### Requirement: Delete role

The system SHALL support deleting roles. Deletion MUST be atomic through a transaction, and any associated-record deletion failure inside the transaction MUST roll back the whole operation. The service MUST NOT only log a warning and continue deleting the role itself. Access topology change notification MUST be emitted after the transaction commits.

#### Scenario: Delete role not assigned to users

- **WHEN** a user deletes a role that is not assigned to any user
- **THEN** the system soft-deletes the role inside the transaction by setting `deleted_at`
- **AND** synchronously deletes related records in `sys_role_menu` and `sys_user_role`
- **AND** notifies access topology change only after transaction commit

#### Scenario: Delete role assigned to users

- **WHEN** a user attempts to delete a role assigned to users
- **THEN** the system asks for confirmation that the role is assigned and will be removed from those users
- **AND** after confirmation, the system deletes the role and cancels those role assignments inside the transaction

#### Scenario: Delete super administrator role

- **WHEN** a user attempts to delete the admin role
- **THEN** the system returns an error that the super administrator role cannot be deleted

#### Scenario: Association cleanup failure rolls back the whole operation

- **WHEN** a user deletes a role
- **AND** cleanup for `sys_role_menu` or `sys_user_role` returns an error inside the transaction
- **THEN** role soft delete MUST roll back as well
- **AND** the operation returns the underlying error rather than only logging a warning
- **AND** no access topology change notification is emitted

### Requirement: Assign users to role

The system SHALL support assigning users to roles. Batch grant and revoke operations MUST commit inside one transaction. All insert or delete operations MUST succeed together or roll back together. The system MUST NOT continue processing later items after logging a warning for a failed per-row insert.

#### Scenario: View users assigned to a role

- **WHEN** a user clicks the assign action for a role
- **THEN** the system navigates to the role-user management page
- **AND** displays the list of users assigned to that role
- **AND** the user list supports pagination and search

#### Scenario: Revoke a user's role authorization

- **WHEN** a user clicks revoke authorization in the role-user list
- **THEN** the system deletes the corresponding `sys_user_role` record inside a transaction
- **AND** the user list refreshes automatically

#### Scenario: Batch authorization commits atomically

- **WHEN** a user selects multiple users not assigned to the role and clicks authorize
- **THEN** the system batch-inserts `sys_user_role` association records inside one transaction
- **AND** any insert failure MUST roll back the whole operation
- **AND** those users can obtain the role menu permissions after signing in again
