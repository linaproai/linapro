import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:info',
      order: 30,
      title: '系统信息',
    },
    name: 'About',
    path: '/about',
    children: [
      {
        name: 'ApiDocs',
        path: '/about/api-docs',
        component: () => import('#/views/about/api-docs/index.vue'),
        meta: {
          icon: 'lucide:file-code',
          title: '系统接口',
        },
      },
      {
        name: 'SystemInfo',
        path: '/about/system-info',
        component: () => import('#/views/about/system-info/index.vue'),
        meta: {
          icon: 'lucide:server',
          title: '版本信息',
        },
      },
    ],
  },
];

export default routes;
