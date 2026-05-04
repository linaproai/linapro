# Role Management

## Purpose

Define role query, maintenance, permission assignment, status control, user assignment, plugin menu authorization, localized seed display behavior, batch deletion, and distributed permission cache invalidation.

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

The system SHALL support deleting roles and toggling enabled or disabled status while preserving protected built-in role rules. Deletion MUST be atomic through a transaction, and any associated-record deletion failure inside the transaction MUST roll back the whole operation. The service MUST NOT only log a warning and continue deleting the role itself. Access topology change notification MUST be emitted after the transaction commits.

#### Scenario: Delete unassigned role

- **WHEN** a user deletes a role that is not assigned to users
- **THEN** the system soft-deletes the role inside the transaction by setting `deleted_at`
- **AND** synchronously deletes related records in `sys_role_menu` and `sys_user_role`
- **AND** notifies access topology change only after transaction commit

#### Scenario: Delete role assigned to users

- **WHEN** a user attempts to delete a role assigned to users
- **THEN** the system asks for confirmation that the role is assigned and will be removed from those users
- **AND** after confirmation, the system deletes the role and cancels those role assignments inside the transaction

#### Scenario: Protected administrator role cannot be deleted

- **WHEN** a user attempts to delete the built-in administrator role
- **THEN** the system rejects the deletion

#### Scenario: Association cleanup failure rolls back the whole operation

- **WHEN** a user deletes a role
- **AND** cleanup for `sys_role_menu` or `sys_user_role` returns an error inside the transaction
- **THEN** role soft delete MUST roll back as well
- **AND** the operation returns the underlying error rather than only logging a warning
- **AND** no access topology change notification is emitted

#### Scenario: Disable role

- **WHEN** a role status changes to disabled
- **THEN** users associated with the role no longer obtain that role's menu permissions after login

### Requirement: Role option and user assignment capabilities

The system SHALL provide role option APIs for user management and role-user assignment workflows. Batch grant and revoke operations MUST commit inside one transaction. All insert or delete operations MUST succeed together or roll back together. The system MUST NOT continue processing later items after logging a warning for a failed per-row insert.

#### Scenario: Get role options

- **WHEN** the user form loads
- **THEN** the system returns enabled role options with `id`, `name`, and `key`

#### Scenario: View users assigned to a role

- **WHEN** a user clicks role assignment
- **THEN** the system shows assigned users with pagination and search

#### Scenario: Revoke a user's role authorization

- **WHEN** a user clicks revoke authorization in the role-user list
- **THEN** the system deletes the corresponding `sys_user_role` record inside a transaction
- **AND** the user list refreshes automatically

#### Scenario: Batch authorization commits atomically

- **WHEN** a user selects multiple users not assigned to the role and clicks authorize
- **THEN** the system batch-inserts `sys_user_role` association records inside one transaction
- **AND** any insert failure MUST roll back the whole operation
- **AND** those users can obtain the role menu permissions after signing in again

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

### Requirement: Permission topology cache must reliably invalidate across nodes

After role, menu, user-role, plugin permission menu, or permission resource relationships change, the system SHALL reliably invalidate token permission snapshots on all nodes through the unified cache coordination mechanism.

#### Scenario: Role menu permission changes

- **WHEN** an administrator updates role menu or button permissions
- **THEN** the system commits role permission relationship changes
- **AND** reliably publishes a permission topology cache revision
- **AND** all nodes discard old token permission snapshots within the staleness window allowed by the permission cache domain

#### Scenario: Menu permission identifier changes

- **WHEN** an administrator creates, updates, deletes, or disables menu permissions
- **THEN** the system publishes a permission topology cache revision
- **AND** later protected API permission checks MUST NOT keep using old menu permission topology indefinitely

#### Scenario: Plugin permission topology changes

- **WHEN** plugin install, enable, disable, uninstall, or synchronization changes plugin menus or button permissions
- **THEN** the system publishes a permission topology cache revision
- **AND** affected permissions participate in authorization decisions on all nodes according to the latest plugin state

### Requirement: Permission topology invalidation publish failure must not silently succeed

Critical permission topology write paths MUST return structured errors or roll back business changes when they cannot publish a permission cache revision. This prevents cluster nodes from continuing to use old authorization snapshots.

#### Scenario: Permission revision publishing fails

- **WHEN** a role, menu, or user-role write path needs to publish a permission topology revision but publishing fails
- **THEN** the system returns a structured business error
- **AND** callers MUST NOT receive a success response claiming the permission change is fully effective
- **AND** the system records the failure reason for retry or repair

#### Scenario: Protected API sees stale permission cache

- **WHEN** a protected API validates permissions, cannot confirm local permission snapshot freshness, and exceeds the failure window
- **THEN** the system rejects the request according to fail-closed policy
- **AND** the system MUST NOT continue allowing uncertain permissions because of an old local permission snapshot
