## ADDED Requirements

### Requirement: Notice Message and Delivery Tenantized
`sys_notify_message` and `sys_notify_delivery` SHALL add `tenant_id`; CRUD filtered by tenant; cross-tenant notices by platform admin through `/platform/notify/*` interface. `sys_notify_channel` is platform global channel directory, no `tenant_id`.

### Requirement: Notice Delivery Log Tenant Recorded
`sys_notify_delivery` SHALL add `tenant_id`; query MUST be tenant isolated.
