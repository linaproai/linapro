import type { VxeGridProps } from '#/adapter/vxe-table';
import type { MenuTreeNode } from '#/api/system/menu';

import { h, markRaw } from 'vue';

import { createIconifyIcon } from '@vben/icons';

import { $t } from '#/locales';

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
  B: { icon: markRaw(CheckCircleIcon), value: $t('pages.system.menu.type.button') },
  D: { icon: markRaw(FolderIcon), value: $t('pages.system.menu.type.directory') },
  M: { icon: markRaw(MenuIcon), value: $t('pages.system.menu.type.menu') },
};

export const nodeOptions = [
  { label: $t('pages.tree.association.linked'), value: true },
  { label: $t('pages.tree.association.independent'), value: false },
];

export const columns: VxeGridProps['columns'] = [
  {
    type: 'checkbox',
    title: $t('pages.system.menu.fields.menuName'),
    field: 'label',
    treeNode: true,
    headerAlign: 'left',
    align: 'left',
    width: 230,
  },
  {
    title: $t('pages.system.menu.fields.icon'),
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
    title: $t('pages.system.menu.fields.type'),
    field: 'type',
    width: 80,
    slots: {
      default: ({ row }) => {
        const current = menuTypes[row.type as string];
        if (!current) {
          return $t('pages.status.unknown');
        }
        return h('span', { class: 'flex items-center justify-center gap-1' }, [
          h(current.icon, { class: 'size-[18px]' }),
          h('span', current.value),
        ]);
      },
    },
  },
  {
    title: $t('pages.system.menu.fields.permissionKey'),
    field: 'permissions',
    headerAlign: 'left',
    align: 'left',
    slots: {
      default: 'permissions',
    },
  },
];
