## ADDED Requirements

### Requirement: Menu subtree checks must use in-memory traversal

The system SHALL determine whether `targetId` belongs to the subtree of `parentId` (`isDescendant`) by loading the menu parent-child mapping once and running BFS/DFS in memory. It MUST NOT issue a separate SQL query for each `parent_id` level inside a loop.

#### Scenario: Subtree check does not create per-level SQL round trips

- **WHEN** the service calls `isDescendant(ctx, parentId, targetId)`
- **THEN** the whole check MUST issue at most one `dao.SysMenu` query to load the parent-child mapping
- **AND** all subsequent traversal MUST happen in memory

### Requirement: sys_role_menu must include menu_id reverse index

The system SHALL maintain `KEY idx_menu_id (menu_id)` on the `sys_role_menu` table to support query paths that batch delete role-menu associations by `menu_id` during menu cascade deletion.

#### Scenario: sys_role_menu reverse index exists

- **WHEN** `make init` completes database initialization
- **THEN** `SHOW INDEX FROM sys_role_menu` MUST include `idx_menu_id` on column `menu_id`

## MODIFIED Requirements

### Requirement: Delete menu

The system SHALL support deleting menus and cascading deletion of submenus. The deletion process MUST run inside a single transaction; any failure in deleting `sys_role_menu` associations MUST cause the entire operation to roll back, rather than only logging a warning and continuing to delete the menu itself.

#### Scenario: Remove menus without submenus

- **WHEN** the user deletes a menu that has no submenus
- **THEN** the system soft deletes this menu (set deleted_at) inside a transaction
- **AND** synchronously deletes the associated records of this menu in `sys_role_menu`

#### Scenario: Delete menu with submenus

- **WHEN** the user deletes a menu with submenus
- **THEN** the system asks for cascade-delete confirmation
- **AND** after user confirmation, deletes this menu and all its submenus inside a single transaction
- **AND** synchronously deletes all associated role-menu relationships

#### Scenario: Association cleanup failure rolls back menu deletion

- **WHEN** the user deletes a menu
- **AND** `sys_role_menu` association cleanup returns an error inside the transaction
- **THEN** the menu soft delete MUST also roll back
- **AND** the operation MUST return the underlying error rather than swallowing it as a warning
