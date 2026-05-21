## ADDED Requirements

### Requirement: Tenant Lifecycle Domain Orchestration
`multi-tenant` plugin SHALL execute current version's explicitly needed domain orchestration during tenant create, suspend, resume, delete, but SHALL NOT create `plugin_multi_tenant_event_outbox` table or declare unimplemented cross-plugin event bus.

#### Scenario: Tenant creation executes built-in side effects
- **WHEN** Platform admin creates tenant T
- **THEN** `multi-tenant` plugin first persists `plugin_multi_tenant_tenant`
- **AND** Then directly executes new tenant plugin enable strategy domain service
- **AND** Does not write outbox events, does not promise async re-delivery

### Requirement: Explicit Plugin Governance Orchestration
New tenant default-enabled tenant-scoped plugin strategy SHALL be completed by `multi-tenant` plugin explicitly calling plugin governance service during tenant creation flow, not through lifecycle event subscription.

#### Scenario: New tenant default enable strategy
- **WHEN** Tenant T created successfully
- **THEN** System queries installed, enabled, `tenant_aware`, `tenant_scoped` and platform strategy allows new tenant default-enabled plugins
- **AND** Writes corresponding tenant plugin state for tenant T

### Requirement: Delete Outbox Placeholder Table
System SHALL NOT use `plugin_multi_tenant_event_outbox` placeholder table that only has writes but lacks subscription registration, consumer claiming, retry, dead letter and per-subscriber delivery state.

### Requirement: Delete Pre-Veto and Cleanup Order
Tenant deletion SHALL execute soft delete only after all plugins' `CanTenantDelete` hooks all pass. Current version does not trigger cross-plugin cleanup via `tenant.deleted` event; when cross-plugin delete cleanup needed, must first design complete reliable event or lifecycle orchestration mechanism.

### Requirement: Suspend/Resume Events Do Not Trigger Data Cleanup
Tenant suspend and resume SHALL only modify tenant status, not trigger any data cleanup or event dispatch; all tenant data must be preserved for recovery.
