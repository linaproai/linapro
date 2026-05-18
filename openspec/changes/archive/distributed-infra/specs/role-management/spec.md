## ADDED Requirements

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
