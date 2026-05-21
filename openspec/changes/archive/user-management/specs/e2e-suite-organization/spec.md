## ADDED Requirements

### Requirement: Multi-Tenant e2e Modular Grouping
`apps/lina-plugins/multi-tenant/hack/tests/e2e/` SHALL maintain multi-tenant plugin's own e2e module with scenario groups: tenant-lifecycle, tenant-isolation, tenant-resolution, tenant-switching, platform-admin, plugin-governance, lifecycle-guard, tenant-config-override, org-center-tenancy.

### Requirement: Cross-Tenant Isolation Matrix Coverage 100%
Each tenancy-aware table SHALL have at least one case verifying "tenant A cannot see tenant B data" in list/get/update/delete anti-examples.

### Requirement: Multi-Tenant Enable/Disable Dual Scenario Validation
e2e SHALL verify multi-tenant enabled and not enabled two states of critical paths.
