import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:monitor',
      order: 20,
      title: '系统监控',
    },
    name: 'Monitor',
    path: '/monitor',
    children: [
      {
        name: 'OnlineUser',
        path: '/monitor/online',
        component: () => import('#/views/monitor/online/index.vue'),
        meta: {
          icon: 'lucide:users',
          title: '在线用户',
        },
      },
      {
        name: 'ServerMonitor',
        path: '/monitor/server',
        component: () => import('#/views/monitor/server/index.vue'),
        meta: {
          icon: 'lucide:activity',
          title: '服务监控',
        },
      },
      {
        name: 'OperLog',
        path: '/monitor/operlog',
        component: () => import('#/views/monitor/operlog/index.vue'),
        meta: {
          icon: 'lucide:file-text',
          title: '操作日志',
        },
      },
      {
        name: 'LoginLog',
        path: '/monitor/loginlog',
        component: () => import('#/views/monitor/loginlog/index.vue'),
        meta: {
          icon: 'lucide:log-in',
          title: '登录日志',
        },
      },
    ],
  },
];

export default routes;
