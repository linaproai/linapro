## ADDED Requirements

### Requirement: Login Log Adds Tenant and Impersonation Fields
`monitor-loginlog` table SHALL add: `tenant_id INT NOT NULL DEFAULT 0` (login target tenant), `acting_user_id INT` (actual operator, = platform admin during impersonation), `on_behalf_of_tenant_id INT` (impersonation target tenant), `is_impersonation BOOL`.

### Requirement: Login Log Query Tenant Isolated
Login log query interface SHALL pass `tenantcap.Apply` filtering; tenant admin only sees own tenant logs.

### Requirement: Log List Shows Impersonation Mark
Tenant admin view, impersonation records SHALL show "platform admin acting on behalf" badge; platform admin view shows acting_user and on_behalf_of_tenant details.
