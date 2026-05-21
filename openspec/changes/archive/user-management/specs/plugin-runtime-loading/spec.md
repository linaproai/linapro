## ADDED Requirements

### Requirement: Route Global Mount + Request-Time Tenant Filtering
`tenant_scoped` plugin HTTP routes SHALL be globally mounted at framework startup; middleware decides by `(tenant_id, plugin_id)` enable state whether to enter handler.

### Requirement: Menu Resolution Filters by Tenant Enable State
`menu-management` resolution of current user's menus SHALL exclude `(plugin_id, tenant_id)` disabled plugin menu items; even if user assigned corresponding permissions.

### Requirement: Enable State Change Immediately Effective
Plugin enable/disable changes SHALL through `pluginruntimecache` invalidation broadcast + cluster message make all nodes immediately aware; no restart needed, expected delay < 1s.
