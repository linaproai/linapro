import type { RouteRecordRaw } from 'vue-router';

// Tenant member administration is handled from the host user-management page.
// Keep this module empty so the sidebar does not expose a separate tenant
// workbench directory.
const routes: RouteRecordRaw[] = [];

export default routes;
