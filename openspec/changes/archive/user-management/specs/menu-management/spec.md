## ADDED Requirements

### Requirement: Global Menu Governance Write Operations Must Require Platform Context
System SHALL treat `sys_menu` as global permission topology and menu governance resource. Menu create, update, delete, status change and other write operations MUST require platform context and corresponding permission strings. Tenant context cannot directly write global menu model.

#### Scenario: Tenant context create menu rejected
- **WHEN** Tenant user with historical `system:menu:add` permission calls menu create interface
- **THEN** System MUST return structured permission error
- **AND** Does not write any record to `sys_menu`

### Requirement: Role Menu Tree Must Filter by Authorization Context
Role authorization menu tree SHALL filter to assignable set by current operator context. Tenant authorization tree does not return platform directory; platform authorization tree retains platform directory.

### Requirement: Menu Write Success Must Publish Permission Topology Revision
When platform admin successfully creates, updates or deletes menu, system reliably publishes access topology cache revision number.
