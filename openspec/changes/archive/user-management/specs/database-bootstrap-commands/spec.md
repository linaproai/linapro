## ADDED Requirements

### Requirement: make init Default PLATFORM
`make init` executing seed SQL, all target tables carrying `tenant_id` seed data SHALL write `tenant_id = 0` (PLATFORM).

### Requirement: make mock Supports Specified Tenant
`make mock` defaults writes PLATFORM; optional parameter `make mock tenant=acme` MUST write mock data to specified tenant.

### Requirement: Multi-Tenant Capability Switch Awareness
After `make init`, system startup SHALL detect if `multi-tenant` plugin installed and enabled; not enabled short-circuits tenancy middleware.
