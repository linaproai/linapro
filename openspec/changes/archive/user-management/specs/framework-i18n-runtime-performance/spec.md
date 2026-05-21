## ADDED Requirements

### Requirement: Runtime Translation Cache Maintains Global Delivery Resource Dimension
Runtime translation package cache SHALL continue using `locale`, `sector` and plugin identifier as cache dimensions; must not introduce `tenant_id` bucketing without actual data source. Current iteration does not land tenant-level i18n override.

### Requirement: Runtime Translation Invalidation Must Explicitly Limit Resource Scope
Runtime translation package cache invalidation SHALL use explicit `InvalidateScope`, limited by `locale`, `sector` and plugin identifier.

### Requirement: Dict and Config Tenant Override Cache Handles Tenant Dimension
Dict cache and config cache SHALL carry `tenant_id` or equivalent tenant scope; tenant write override only invalidates corresponding tenant dimension.
