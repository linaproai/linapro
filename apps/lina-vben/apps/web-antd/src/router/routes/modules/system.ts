import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:settings',
      order: 10,
      title: '系统管理',
    },
    name: 'System',
    path: '/system',
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
        name: 'DeptManagement',
        path: '/system/dept',
        component: () => import('#/views/system/dept/index.vue'),
        meta: {
          icon: 'lucide:network',
          title: '部门管理',
        },
      },
      {
        name: 'PostManagement',
        path: '/system/post',
        component: () => import('#/views/system/post/index.vue'),
        meta: {
          icon: 'lucide:briefcase',
          title: '岗位管理',
        },
      },
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
        name: 'NoticeManagement',
        path: '/system/notice',
        component: () => import('#/views/system/notice/index.vue'),
        meta: {
          icon: 'lucide:megaphone',
          title: '通知公告',
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
        name: 'ScheduledJobCatalog',
        path: '/system/scheduled-job',
        redirect: '/system/job',
        meta: {
          icon: 'lucide:calendar-range',
          title: '定时任务',
        },
      },
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
      {
        name: 'PluginManagement',
        path: '/system/plugin',
        component: () => import('#/views/system/plugin/index.vue'),
        meta: {
          icon: 'lucide:puzzle',
          title: '插件管理',
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
