## MODIFIED Requirements

### Requirement: Plugin list query has no side effects

The system SHALL treat plugin list queries as read-only operations with no side effects. List queries may read discovered source manifests, dynamic plugin registry data, release snapshots, and governance projections, but MUST NOT create, update, or delete plugin governance table data. Plugin scanning and governance synchronization MUST only be triggered by explicit synchronization operations or host startup synchronization. Host startup synchronization SHALL also be difference-driven: when plugin registry, release snapshot, menus, permissions, and resource reference projections all match, the system MUST NOT open transactions, write to the database, or perform post-write reads.

#### Scenario: Querying plugin list from management page

- **WHEN** an administrator opens plugin management and calls `GET /api/v1/plugins`
- **THEN** the system returns the plugin list and current governance state
- **AND** the GET request does not write to `sys_plugin`, `sys_plugin_release`, `sys_plugin_resource_ref`, `sys_menu`, or `sys_role_menu`

#### Scenario: Explicit plugin synchronization

- **WHEN** an administrator triggers plugin synchronization via `POST /api/v1/plugins/sync`
- **THEN** the system scans source plugins and dynamic plugin artifacts
- **AND** the system may synchronize registry, release snapshot, resource index, menu, and permission governance data from manifests

#### Scenario: Startup synchronization produces no database side effects when no differences exist

- **WHEN** host startup synchronization discovers plugin manifests are fully consistent with existing governance projections
- **THEN** the system MUST NOT write to `sys_plugin`, `sys_plugin_release`, `sys_plugin_resource_ref`, `sys_menu`, or `sys_role_menu` for that plugin
- **AND** the system MUST NOT open empty transactions or repeatedly post-read the same governance rows to refresh startup snapshots
