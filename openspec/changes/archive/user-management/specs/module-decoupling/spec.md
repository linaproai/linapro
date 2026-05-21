## ADDED Requirements

### Requirement: Multi-Tenant Plugin Disable Linked Hide
When `multi-tenant` plugin not enabled, frontend SHALL completely hide: login page tenant selector and input; workbench header tenant identifier and switcher; user management page tenant column and filter; all `/platform/*` and `/tenant/*` entry menus.

### Requirement: Enable Linked Display
After enable SHALL linked display corresponding UI; state switch via `IsEnabled(ctx, 'multi-tenant')`.

### Requirement: Cross-Tenant UI Element Bidirectional Hide
Tenant admin view hides all platform-level UI; platform admin view hides tenant-level private UI.
