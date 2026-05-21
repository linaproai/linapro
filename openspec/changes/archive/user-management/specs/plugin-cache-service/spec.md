## ADDED Requirements

### Requirement: Plugin Cache Keys Default Carry Tenant Dimension
Plugin cache read/write interfaces through host service SHALL default prefix `tenant=<id>:` on cache key; plugin can explicitly declare `scope=platform` in platform mode to access cross-tenant shared cache (requires platform admin permission with audit).

### Requirement: Invalidation Broadcast Tenant-Scoped
Plugin calling `cache.Invalidate(scope, key)` SHALL carry current tenant in distributed-cache-coordination invalidation message; `cache.InvalidatePlatform(...)` broadcasts all-tenant cascade invalidation.

### Requirement: Platform Shared Cache Audit Requirements
Any write to `tenant=0` SHALL be recorded in operlog with `oper_type='other'` marking `platform_cache_write`.
