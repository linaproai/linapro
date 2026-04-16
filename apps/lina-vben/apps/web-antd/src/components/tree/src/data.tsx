import type { VxeGridProps } from '#/adapter/vxe-table';
import type { MenuTreeNode } from '#/api/system/menu';

import { h, markRaw } from 'vue';

import { createIconifyIcon } from '@vben/icons';

export interface Permission {
  checked: boolean;
  id: number;
  label: string;
}

export interface MenuPermissionOption extends MenuTreeNode {
  permissions: Permission[];
}

// 使用 Iconify 图标
const FolderIcon = createIconifyIcon('lucide:folder');
const MenuIcon = createIconifyIcon('lucide:menu');
const CheckCircleIcon = createIconifyIcon('lucide:check-circle');

const menuTypes: Record<string, { icon: any; value: string }> = {
  B: { icon: markRaw(CheckCircleIcon), value: '按钮' },
  D: { icon: markRaw(FolderIcon), value: '目录' },
  M: { icon: markRaw(MenuIcon), value: '菜单' },
};

export const nodeOptions = [
  { label: '节点关联', value: true },
  { label: '节点独立', value: false },
];

export const columns: VxeGridProps['columns'] = [
  {
    type: 'checkbox',
    title: '菜单名称',
    field: 'label',
    treeNode: true,
    headerAlign: 'left',
    align: 'left',
    width: 230,
  },
  {
    title: '图标',
    field: 'icon',
    width: 80,
    slots: {
      default: ({ row }) => {
        if (row?.icon === '#') {
          return '';
        }
        if (row?.icon) {
          return h('span', { class: 'flex justify-center' }, h(createIconifyIcon(row.icon)));
        }
        return '';
      },
    },
  },
  {
    title: '类型',
    field: 'type',
    width: 80,
    slots: {
      default: ({ row }) => {
        const current = menuTypes[row.type as string];
        if (!current) {
          return '未知';
        }
        return h('span', { class: 'flex items-center justify-center gap-1' }, [
          h(current.icon, { class: 'size-[18px]' }),
          h('span', current.value),
        ]);
      },
    },
  },
  {
    title: '权限标识',
    field: 'permissions',
    headerAlign: 'left',
    align: 'left',
    slots: {
      default: 'permissions',
    },
  },
];