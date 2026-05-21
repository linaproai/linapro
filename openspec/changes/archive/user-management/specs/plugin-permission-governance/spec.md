## ADDED Requirements

### Requirement: Permission Points Do Not Encode Platform/Tenant Boundary
Permission point strings SHALL only express functional actions, not encode platform/tenant boundary through `platform:*` or `tenant:*` prefix. Host and official multi-tenant plugin SHALL unified use `system:*` permission points; platform/tenant boundaries determined by route plane, current tenant context, data permissions, plugin enable state and interface-level validation.

### Requirement: Plugin Platform Governance Permissions Must Be Separated from Tenant Plugin Self-Service
System SHALL distinguish plugin platform governance permissions from tenant plugin self-service permissions. Platform plugin governance includes upload, sync, install, uninstall, enable, disable, upgrade, install mode switch, new tenant auto-enable strategy and platform plugin state governance; these operations MUST require platform context. Tenant plugin self-service can only act on current tenant's `tenant_scoped` plugin enable state through tenant plugin interfaces.

### Requirement: Permission Resolution Filters by Current Tenant Enabled Plugins
Permission resolution SHALL exclude "plugin not enabled in current tenant" permission points (even if user assigned); avoid showing permissions user cannot actually operate.

### Requirement: Menu Button Permissions Project by Page Actual Buttons
Plugin menu manifest button permissions SHALL only project corresponding management page's actually displayed or triggerable buttons/entries.
