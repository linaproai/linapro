## ADDED Requirements

### Requirement: User Message Table Tenantized
`sys_user_message` SHALL add `tenant_id`; message inbox only returns current tenant's messages.

### Requirement: Message Read Permission
User SHALL only read own + current tenant messages; cross-tenant access rejected (returns 404 to avoid leaking existence).
