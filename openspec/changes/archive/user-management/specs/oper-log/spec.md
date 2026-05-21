## ADDED Requirements

### Requirement: Operation Log Adds Tenant and Impersonation Fields
`monitor-operlog` table SHALL add `tenant_id`, `acting_user_id`, `on_behalf_of_tenant_id`, `is_impersonation` fields. All protected operations including lifecycle guard, force action, tenant lifecycle must be recorded.

### Requirement: Force Operations Separate Audit Class
Platform admin `--force` channel operations SHALL write `oper_type='other'`; payload marks `platform_force_action` with all bypassed reasons and context.

### Requirement: LifecycleGuard Calls Separate Audit Class
All LifecycleGuard hook calls SHALL be recorded in operlog with `oper_type='other'` marking `lifecycle_guard`.

### Requirement: Operation Log Query Tenant Isolated
Operation log query interface SHALL pass `tenantcap.Apply` filtering; tenant admin only sees own tenant logs. Platform admin only through `/platform/oper-log` management interface views full with filtering support.
