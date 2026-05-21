## ADDED Requirements

### Requirement: Invalidation Messages Carry tenant_id
`cachecoord` invalidation broadcast messages SHALL include `tenant_id` field; each invalidation message must explicitly declare scope.

### Requirement: Prohibit Unjustified All-Tenant Clear
Normal business paths SHALL not allow initiating `tenant_id = -1` all-clear invalidation; only platform admin explicit operations can, with operlog audit.

### Requirement: Invalidation Message Idempotent
Nodes receiving invalidation messages SHALL process idempotently (retryable without error); duplicate messages do not affect final consistency.
