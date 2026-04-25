import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { DictEnum } from '@vben/constants';
import { FolderIcon, MenuIcon, OkButtonIcon, VbenIcon } from '@vben/icons';
import { getPopupContainer } from '@vben/utils';

import { z } from '#/adapter/form';
import { $t } from '#/locales';
import { getDictOptions } from '#/utils/dict';
import { renderDict } from '#/utils/render';

/** 查询表单配置 */
export function querySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('pages.system.menu.fields.menuName'),
      componentProps: {
        placeholder: $t('pages.system.menu.placeholders.menuName'),
      },
    },
    {
      component: 'Select',
      componentProps: {
        getPopupContainer,
        options: getDictOptions(DictEnum.SYS_NORMAL_DISABLE),
      },
      fieldName: 'status',
      label: $t('pages.system.menu.fields.menuStatus'),
    },
    {
      component: 'Select',
      componentProps: {
        getPopupContainer,
        options: getDictOptions(DictEnum.SYS_SHOW_HIDE),
      },
      fieldName: 'visible',
      label: $t('pages.system.menu.fields.visibleStatus'),
    },
  ];
}

const menuTypes = {
  D: { icon: FolderIcon, value: $t('pages.system.menu.type.directory') },
  M: { icon: MenuIcon, value: $t('pages.system.menu.type.menu') },
  B: { icon: OkButtonIcon, value: $t('pages.system.menu.type.button') },
};

/** 表格列配置 */
export const columns: VxeGridProps['columns'] = [
  {
    title: $t('pages.system.menu.fields.menuName'),
    field: 'name',
    treeNode: true,
    width: 200,
    align: 'left',
  },
  {
    title: $t('pages.system.menu.fields.icon'),
    field: 'icon',
    width: 80,
    slots: {
      default: ({ row }) => {
        if (!row?.icon || row.icon === '#') {
          return '';
        }
        return h('span', { class: 'flex justify-center' }, [
          h(VbenIcon, { icon: row.icon }),
        ]);
      },
    },
  },
  {
    title: $t('pages.fields.sort'),
    field: 'sort',
    width: 80,
  },
  {
    title: $t('pages.system.menu.fields.componentType'),
    field: 'type',
    width: 120,
    slots: {
      default: ({ row }) => {
        const current = menuTypes[row.type as 'D' | 'M' | 'B'];
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
    field: 'perms',
  },
  {
    title: $t('pages.system.menu.fields.componentPath'),
    field: 'component',
  },
  {
    title: $t('pages.common.status'),
    field: 'status',
    width: 100,
    slots: {
      default: ({ row }) => {
        return renderDict(row.status, DictEnum.SYS_NORMAL_DISABLE);
      },
    },
  },
  {
    title: $t('pages.system.menu.fields.visible'),
    field: 'visible',
    width: 100,
    slots: {
      default: ({ row }) => {
        return renderDict(row.visible, DictEnum.SYS_SHOW_HIDE);
      },
    },
  },
  {
    title: $t('pages.common.createdAt'),
    field: 'createdAt',
  },
  {
    field: 'action',
    fixed: 'right',
    slots: { default: 'action' },
    title: $t('pages.common.actions'),
    resizable: false,
    width: 'auto',
  },
];

const yesNoOptions = [
  { label: $t('pages.common.yes'), value: 1 },
  { label: $t('pages.common.no'), value: 0 },
];

/** 抽屉表单配置 */
export function drawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      dependencies: {
        show: () => false,
        triggerFields: [''],
      },
      fieldName: 'id',
    },
    {
      component: 'TreeSelect',
      defaultValue: 0,
      fieldName: 'parentId',
      label: $t('pages.system.menu.fields.parentMenu'),
      rules: 'selectRequired',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('pages.system.menu.type.directory'), value: 'D' },
          { label: $t('pages.system.menu.type.menu'), value: 'M' },
          { label: $t('pages.system.menu.type.button'), value: 'B' },
        ],
        optionType: 'button',
      },
      defaultValue: 'D',
      dependencies: {
        componentProps: () => {
          return {};
        },
        triggerFields: ['type'],
      },
      fieldName: 'type',
      label: $t('pages.system.menu.fields.menuType'),
    },
    {
      component: 'Input',
      dependencies: {
        show: (values) => values.type !== 'B',
        triggerFields: ['type'],
      },
      renderComponentContent: (model) => ({
        addonBefore: () =>
          model.icon ? h(VbenIcon, { icon: model.icon }) : null,
        addonAfter: () =>
          h(
            'a',
            { href: 'https://icon-sets.iconify.design/', target: '_blank' },
            $t('pages.system.menu.actions.searchIcons'),
          ),
      }),
      fieldName: 'icon',
      help: $t('pages.system.menu.help.icon'),
      label: $t('pages.system.menu.fields.menuIcon'),
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('pages.system.menu.fields.menuName'),
      componentProps: {
        placeholder: $t('pages.system.menu.placeholders.menuName'),
      },
      help: $t('pages.system.menu.help.name'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.menu.placeholders.permissionKey'),
      },
      dependencies: {
        rules: (model) => {
          if (model.type === 'M' || model.type === 'B') {
            return z
              .string({
                message: $t('pages.system.menu.validation.permissionRequired'),
              })
              .min(1, $t('pages.system.menu.validation.permissionRequired'));
          }
          return z.string().optional();
        },
        show: (values) => values.type !== 'D',
        triggerFields: ['type'],
      },
      fieldName: 'perms',
      help: $t('pages.system.menu.help.permission'),
      label: $t('pages.system.menu.fields.permissionKey'),
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      help: $t('pages.system.menu.help.sort'),
      label: $t('pages.system.menu.fields.displaySort'),
      defaultValue: 0,
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: (model) => {
        const placeholder =
          model.isFrame === 1
            ? $t('pages.system.menu.placeholders.externalPath')
            : $t('pages.system.menu.placeholders.routePath');
        return {
          placeholder,
        };
      },
      dependencies: {
        rules: (model) => {
          if (model.isFrame !== 1) {
            return z
              .string({
                message: $t('pages.system.menu.validation.routeRequired'),
              })
              .min(1, $t('pages.system.menu.validation.routeRequired'));
          }
          return z
            .string({
              message: $t('pages.system.menu.validation.linkRequired'),
            })
            .regex(/^https?:\/\//, {
              message: $t('pages.system.menu.validation.linkInvalid'),
            });
        },
        show: (values) => values?.type !== 'B',
        triggerFields: ['isFrame', 'type'],
      },
      fieldName: 'path',
      help: $t('pages.system.menu.help.path'),
      label: $t('pages.system.menu.fields.routePath'),
    },
    {
      component: 'Input',
      componentProps: (model) => {
        return {
          disabled: model.isFrame === 1,
        };
      },
      defaultValue: '',
      dependencies: {
        rules: (model) => {
          if (model.path && !/^https?:\/\//.test(model.path)) {
            return z
              .string()
              .min(1, {
                message: $t(
                  'pages.system.menu.validation.componentRequired',
                ),
              })
              .refine((val) => !val.startsWith('/') && !val.endsWith('/'), {
                message: $t(
                  'pages.system.menu.validation.componentNoSlash',
                ),
              });
          }
          return z.string().optional();
        },
        show: (values) => values.type === 'M',
        triggerFields: ['type', 'path'],
      },
      fieldName: 'component',
      help: $t('pages.system.menu.help.component'),
      label: $t('pages.system.menu.fields.componentPath'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: yesNoOptions,
        optionType: 'button',
      },
      defaultValue: 0,
      dependencies: {
        show: (values) => values.type !== 'B',
        triggerFields: ['type'],
      },
      fieldName: 'isFrame',
      help: $t('pages.system.menu.help.externalLink'),
      label: $t('pages.system.menu.fields.isExternalLink'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('pages.system.menu.visible.shown'), value: 1 },
          { label: $t('pages.system.menu.visible.hidden'), value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      dependencies: {
        show: (values) => values.type !== 'B',
        triggerFields: ['type'],
      },
      fieldName: 'visible',
      help: $t('pages.system.menu.help.visible'),
      label: $t('pages.system.menu.fields.visible'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('pages.status.enabled'), value: 1 },
          { label: $t('pages.status.disabled'), value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      dependencies: {
        show: (values) => values.type !== 'B',
        triggerFields: ['type'],
      },
      fieldName: 'status',
      help: $t('pages.system.menu.help.status'),
      label: $t('pages.system.menu.fields.menuStatus'),
    },
    {
      component: 'Input',
      componentProps: (model) => ({
        disabled: model.isFrame === 1,
        placeholder: $t('pages.system.menu.placeholders.queryParam'),
      }),
      dependencies: {
        show: (values) => values.type === 'M',
        triggerFields: ['type'],
      },
      fieldName: 'queryParam',
      help: $t('pages.system.menu.help.queryParam'),
      label: $t('pages.system.menu.fields.routeParams'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: yesNoOptions,
        optionType: 'button',
      },
      defaultValue: 0,
      dependencies: {
        show: (values) => values.type === 'M',
        triggerFields: ['type'],
      },
      fieldName: 'isCache',
      help: $t('pages.system.menu.help.cache'),
      label: $t('pages.system.menu.fields.cache'),
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: $t('pages.system.menu.placeholders.remark'),
        rows: 3,
      },
      fieldName: 'remark',
      formItemClass: 'col-span-2',
      label: $t('pages.common.remark'),
    },
  ];
}
