## ADDED Requirements

### Requirement: 菜单子树判定必须使用内存遍历
系统 SHALL 在判定 `targetId` 是否属于 `parentId` 子树（`isDescendant`）时一次性加载菜单父子映射，并在内存中执行 BFS/DFS 遍历，禁止在循环内对每一层 `parent_id` 发起独立 SQL 查询。

#### Scenario: 子树判定不产生按层 SQL 往返
- **WHEN** Service 调用 `isDescendant(ctx, parentId, targetId)`
- **THEN** 整个判定 MUST 至多产生一次 `dao.SysMenu` 查询用于加载父子映射
- **AND** 后续遍历 MUST 全程在内存中完成

### Requirement: sys_role_menu 必须包含 menu_id 反向索引
系统 SHALL 在 `sys_role_menu` 表上维护 `KEY idx_menu_id (menu_id)` 反向索引，以支撑级联删除菜单时按 `menu_id` 批量删除关联记录的查询路径。

#### Scenario: sys_role_menu 反向索引存在
- **WHEN** `make init` 完成数据库初始化
- **THEN** `SHOW INDEX FROM sys_role_menu` 结果中 MUST 出现 `idx_menu_id`，索引列为 `menu_id`

## MODIFIED Requirements

### Requirement: Delete menu

System SHALL support deleting menus and cascading deletion of submenus. The deletion process MUST run inside a single transaction; any failure in deleting `sys_role_menu` associations MUST cause the entire operation to roll back, rather than only logging a warning and continuing to delete the menu itself.

#### Scenario: Remove menus without submenus
- **WHEN** User deletes a menu that has no submenus
- **THEN** The system soft deletes this menu (set deleted_at) inside a transaction
- **AND** Synchronously delete the associated records of this menu in sys_role_menu

#### Scenario: Delete menu with submenus
- **WHEN** The user deletes a menu with submenus
- **THEN** The system prompts "Submenu exists, do you want to delete it cascaded?"
- **AND** After user confirmation, delete this menu and all its submenus inside a single transaction
- **AND** Synchronously delete all associated role-menu relationships

#### Scenario: Association cleanup failure rolls back menu deletion
- **WHEN** User deletes a menu
- **AND** `sys_role_menu` association cleanup returns an error inside the transaction
- **THEN** The menu soft delete MUST also roll back
- **AND** The operation MUST return the underlying error rather than swallowing it as a warning
