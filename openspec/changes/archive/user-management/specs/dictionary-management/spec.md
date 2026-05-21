## ADDED Requirements

### Requirement: Platform Dict and Tenant Dict Visibility Rules
Tenant admin querying dict types/data SHALL only return: own tenant overrides and `tenant_id=0 AND allow_tenant_override` as fallback candidates.

### Requirement: Dict Write Permissions
Tenant admin SHALL only create/modify `tenant_id=current` dict data for `allow_tenant_override=true` types; cannot modify type metadata or platform defaults.

### Requirement: Dict Data SQL Orchestration
All dict seed SQL SHALL write `tenant_id=0`; mock data can optionally specify tenant.
