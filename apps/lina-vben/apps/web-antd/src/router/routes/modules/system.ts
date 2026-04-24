import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:shield-check',
      order: 10,
      title: 'page.routes.system.accessManagement',
    },
    name: 'IAM',
    path: '/iam',
    children: [
      {
        name: 'UserManagement',
        path: '/system/user',
        component: () => import('#/views/system/user/index.vue'),
        meta: {
          icon: 'ant-design:user-outlined',
          title: 'page.routes.system.userManagement',
        },
      },
      {
        name: 'RoleManagement',
        path: '/system/role',
        component: () => import('#/views/system/role/index.vue'),
        meta: {
          icon: 'lucide:shield',
          title: 'page.routes.system.roleManagement',
        },
      },
      {
        name: 'MenuManagement',
        path: '/system/menu',
        component: () => import('#/views/system/menu/index.vue'),
        meta: {
          icon: 'lucide:menu',
          title: 'page.routes.system.menuManagement',
        },
      },
      {
        name: 'RoleAuthUser',
        path: '/system/role-auth/user/:id',
        component: () => import('#/views/system/role-auth/index.vue'),
        meta: {
          hideInMenu: true,
          title: 'page.routes.system.roleUserManagement',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:settings-2',
      order: 20,
      title: 'page.routes.system.systemSettings',
    },
    name: 'Setting',
    path: '/setting',
    children: [
      {
        name: 'DictManagement',
        path: '/system/dict',
        component: () => import('#/views/system/dict/index.vue'),
        meta: {
          icon: 'lucide:book-a',
          title: 'page.routes.system.dictManagement',
        },
      },
      {
        name: 'ConfigManagement',
        path: '/system/config',
        component: () => import('#/views/system/config/index.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: 'page.routes.system.configManagement',
        },
      },
      {
        name: 'FileManagement',
        path: '/system/file',
        component: () => import('#/views/system/file/index.vue'),
        meta: {
          icon: 'lucide:folder-open',
          title: 'page.routes.system.fileManagement',
        },
      },
      {
        name: 'MessageList',
        path: '/system/message',
        component: () => import('#/views/system/message/index.vue'),
        meta: {
          hideInMenu: true,
          title: 'page.routes.system.messageList',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:calendar-range',
      order: 30,
      title: 'page.routes.system.scheduler',
    },
    name: 'Scheduler',
    path: '/scheduler',
    children: [
      {
        name: 'JobManagement',
        path: '/system/job',
        component: () => import('#/views/system/job/index.vue'),
        meta: {
          icon: 'lucide:clock-3',
          title: 'page.routes.system.jobManagement',
        },
      },
      {
        name: 'JobGroupManagement',
        path: '/system/job-group',
        component: () => import('#/views/system/job-group/index.vue'),
        meta: {
          icon: 'lucide:blocks',
          title: 'page.routes.system.jobGroupManagement',
        },
      },
      {
        name: 'JobLogManagement',
        path: '/system/job-log',
        component: () => import('#/views/system/job-log/index.vue'),
        meta: {
          icon: 'lucide:scroll-text',
          title: 'page.routes.system.jobLogManagement',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:puzzle',
      order: 40,
      title: 'page.routes.system.extensionCenter',
    },
    name: 'Extension',
    path: '/extension',
    children: [
      {
        name: 'PluginManagement',
        path: '/system/plugin',
        component: () => import('#/views/system/plugin/index.vue'),
        meta: {
          icon: 'lucide:plug',
          title: 'page.routes.system.pluginManagement',
        },
      },
    ],
  },
  {
    name: 'Profile',
    path: '/profile',
    component: () => import('#/views/_core/profile/index.vue'),
    meta: {
      icon: 'lucide:user',
      hideInMenu: true,
      title: 'page.auth.profile',
    },
  },
];

export default routes;
