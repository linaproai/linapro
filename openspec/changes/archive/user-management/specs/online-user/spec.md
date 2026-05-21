## ADDED Requirements

### Requirement: sys_online_session Adds Tenant Field
`sys_online_session` SHALL add `tenant_id` column; primary key uses globally unique `token_id`; `tenant_id` only as session ownership, list filtering, data permissions and request-time claim/session consistency validation dimension; index covers `(tenant_id, user_id)` and `(tenant_id, login_time)`.

### Requirement: Online Session Query Tenant Filtered
`GET /online/list` SHALL pass `tenantcap.Apply` filtering; tenant admin only sees own tenant sessions.

### Requirement: Kick Interface Tenant Validated
`POST /online/{token_id}/kick` SHALL first locate target session by globally unique `token_id`, then validate target session's visibility in current operator's data scope; mismatch returns 403.

### Requirement: Session Cleanup Task Tenant-Aware
Session expired cleanup scheduled task SHALL scan by `last_active_time`, use `(tenant_id, last_active_time)` index when needed; cleanup granularity per-tenant independent statistics, but task itself is platform-level (`tenant_id=0`).
