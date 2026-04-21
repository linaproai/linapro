import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:shield-check',
      order: 10,
      title: '权限管理',
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
          title: '用户管理',
        },
      },
      {
        name: 'RoleManagement',
        path: '/system/role',
        component: () => import('#/views/system/role/index.vue'),
        meta: {
          icon: 'lucide:shield',
          title: '角色管理',
        },
      },
      {
        name: 'MenuManagement',
        path: '/system/menu',
        component: () => import('#/views/system/menu/index.vue'),
        meta: {
          icon: 'lucide:menu',
          title: '菜单管理',
        },
      },
      {
        name: 'RoleAuthUser',
        path: '/system/role-auth/user/:id',
        component: () => import('#/views/system/role-auth/index.vue'),
        meta: {
          hideInMenu: true,
          title: '角色授权用户',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:settings-2',
      order: 20,
      title: '系统设置',
    },
    name: 'Setting',
    path: '/setting',
    children: [
      {
        name: 'DictManagement',
        path: '/system/dict',
        component: () => import('#/views/system/dict/index.vue'),
        meta: {
          icon: 'lucide:book-open',
          title: '字典管理',
        },
      },
      {
        name: 'ConfigManagement',
        path: '/system/config',
        component: () => import('#/views/system/config/index.vue'),
        meta: {
          icon: 'lucide:sliders-horizontal',
          title: '参数设置',
        },
      },
      {
        name: 'FileManagement',
        path: '/system/file',
        component: () => import('#/views/system/file/index.vue'),
        meta: {
          icon: 'lucide:folder-open',
          title: '文件管理',
        },
      },
      {
        name: 'MessageList',
        path: '/system/message',
        component: () => import('#/views/system/message/index.vue'),
        meta: {
          hideInMenu: true,
          title: '消息列表',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:calendar-range',
      order: 30,
      title: '任务调度',
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
          title: '任务管理',
        },
      },
      {
        name: 'JobGroupManagement',
        path: '/system/job-group',
        component: () => import('#/views/system/job-group/index.vue'),
        meta: {
          icon: 'lucide:blocks',
          title: '分组管理',
        },
      },
      {
        name: 'JobLogManagement',
        path: '/system/job-log',
        component: () => import('#/views/system/job-log/index.vue'),
        meta: {
          icon: 'lucide:scroll-text',
          title: '执行日志',
        },
      },
    ],
  },
  {
    meta: {
      icon: 'lucide:puzzle',
      order: 40,
      title: '扩展中心',
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
          title: '插件管理',
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
      title: $t('page.auth.profile'),
    },
  },
];

export default routes;
