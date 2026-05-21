## ADDED Requirements

### Requirement: Dict Tables Add tenant_id and Override Switch
`sys_dict_type` and `sys_dict_data` SHALL add `tenant_id INT NOT NULL DEFAULT 0`; `sys_dict_type` also adds `allow_tenant_override BOOL NOT NULL DEFAULT FALSE`.

### Requirement: Dict Read Path Platform Fallback
Dict data read SHALL through `tenantcap.ReadWithPlatformFallback` implement "tenant override priority + PLATFORM fallback" semantics.

### Requirement: Dict Cache Key Carries Tenant
Dict cache key SHALL be `dict:tenant=<id>:type=<type>`; tenant write triggers own tenant invalidation; platform default write triggers all-tenant invalidation.

### Requirement: Tenant Dict Fallback Row Metadata
Tenant context querying dict types and dict data SHALL for platform default fallback rows return source and action metadata including `sourceTenantId`, `isFallback`, `canEdit`, `canOverride`, `overrideMode`.

### Requirement: Dict Fallback Actions Must Avoid Must-Fail Detail Requests
Frontend SHALL use dict row action metadata to determine operation buttons; for `isFallback=true` and `canEdit=false` rows, frontend must not show direct edit entry that would trigger not-found detail request.
