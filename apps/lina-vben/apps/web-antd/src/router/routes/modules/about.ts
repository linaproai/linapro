import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:flask-conical',
      order: 50,
      title: 'page.routes.about.developerCenter',
    },
    name: 'Developer',
    path: '/developer',
    children: [
      {
        name: 'ApiDocs',
        path: '/about/api-docs',
        component: () => import('#/views/about/api-docs/index.vue'),
        meta: {
          icon: 'lucide:file-code',
          title: 'page.routes.about.apiDocs',
        },
      },
      {
        name: 'SystemInfo',
        path: '/about/system-info',
        component: () => import('#/views/about/system-info/index.vue'),
        meta: {
          icon: 'lucide:server',
          title: 'page.routes.about.systemInfo',
        },
      },
    ],
  },
];

export default routes;
