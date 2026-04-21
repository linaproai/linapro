import type { RouteRecordRaw } from 'vue-router';

// Plugin-owned monitor pages are injected through backend menus and the plugin
// dynamic page host. Keep this module empty so local static routes do not
// bypass plugin install/enable state.
const routes: RouteRecordRaw[] = [];

export default routes;
