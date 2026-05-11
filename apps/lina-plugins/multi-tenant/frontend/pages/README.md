# Multi Tenant Frontend Pages

This directory owns the source-plugin frontend pages for the official `multi-tenant` plugin.

- `tenant-management.vue` mounts the platform tenant-management route at `/platform/tenants`.
- `components/tenant-modal.vue` contains the tenant create/edit modal used by the page.
- `tenant-client.ts` contains page-local API calls for tenant CRUD.

Host-owned tenant runtime features, such as login tenant selection, the header tenant switcher, route guards, and the tenant store, remain in `apps/lina-vben` because they are workbench runtime integration points rather than page-local plugin UI.
