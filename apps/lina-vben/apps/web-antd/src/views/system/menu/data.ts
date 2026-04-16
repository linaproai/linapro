import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { DictEnum } from '@vben/constants';
import { FolderIcon, MenuIcon, OkButtonIcon, VbenIcon } from '@vben/icons';
import { getPopupContainer } from '@vben/utils';

import { z } from '#/adapter/form';
import { getDictOptions } from '#/utils/dict';
import { renderDict } from '#/utils/render';

/** 查询表单配置 */
export function querySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: '菜单名称',
      componentProps: {
        placeholder: '请输入菜单名称',
      },
    },
    {
      component: 'Select',
      componentProps: {
        getPopupContainer,
        options: getDictOptions(DictEnum.SYS_NORMAL_DISABLE),
      },
      fieldName: 'status',
      label: '菜单状态',
    },
    {
      component: 'Select',
      componentProps: {
        getPopupContainer,
        options: getDictOptions(DictEnum.SYS_SHOW_HIDE),
      },
      fieldName: 'visible',
      label: '显示状态',
    },
  ];
}

// 菜单类型映射（D=目录 M=菜单 B=按钮）
const menuTypes = {
  D: { icon: FolderIcon, value: '目录' },
  M: { icon: MenuIcon, value: '菜单' },
  B: { icon: OkButtonIcon, value: '按钮' },
};

/** 表格列配置 */
export const columns: VxeGridProps['columns'] = [
  {
    title: '菜单名称',
    field: 'name',
    treeNode: true,
    width: 200,
    align: 'left',
  },
  {
    title: '图标',
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
    title: '排序',
    field: 'sort',
    width: 80,
  },
  {
    title: '组件类型',
    field: 'type',
    width: 120,
    slots: {
      default: ({ row }) => {
        const current = menuTypes[row.type as 'D' | 'M' | 'B'];
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
    field: 'perms',
  },
  {
    title: '组件路径',
    field: 'component',
  },
  {
    title: '状态',
    field: 'status',
    width: 100,
    slots: {
      default: ({ row }) => {
        return renderDict(row.status, DictEnum.SYS_NORMAL_DISABLE);
      },
    },
  },
  {
    title: '显示',
    field: 'visible',
    width: 100,
    slots: {
      default: ({ row }) => {
        return renderDict(row.visible, DictEnum.SYS_SHOW_HIDE);
      },
    },
  },
  {
    title: '创建时间',
    field: 'createdAt',
  },
  {
    field: 'action',
    fixed: 'right',
    slots: { default: 'action' },
    title: '操作',
    resizable: false,
    width: 'auto',
  },
];

// 是否选项
const yesNoOptions = [
  { label: '是', value: 1 },
  { label: '否', value: 0 },
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
    label: '上级菜单',
    rules: 'selectRequired',
  },
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: [
        { label: '目录', value: 'D' },
        { label: '菜单', value: 'M' },
        { label: '按钮', value: 'B' },
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
    label: '菜单类型',
  },
  {
    component: 'Input',
    dependencies: {
      // 类型不为按钮时显示
      show: (values) => values.type !== 'B',
      triggerFields: ['type'],
    },
    renderComponentContent: (model) => ({
      addonBefore: () => (model.icon ? h(VbenIcon, { icon: model.icon }) : null),
      addonAfter: () =>
        h(
          'a',
          { href: 'https://icon-sets.iconify.design/', target: '_blank' },
          '搜索图标',
        ),
    }),
    fieldName: 'icon',
    help: '点击搜索图标跳转到iconify & 粘贴',
    label: '菜单图标',
  },
  {
    component: 'Input',
    fieldName: 'name',
    label: '菜单名称',
    componentProps: {
      placeholder: '请输入菜单名称',
    },
    help: '支持i18n写法, 如: menu.system.user',
    rules: 'required',
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: '请输入权限标识',
    },
    dependencies: {
      rules: (model) => {
        if (model.type === 'M' || model.type === 'B') {
          return z
            .string({ message: '请输入权限标识' })
            .min(1, '请输入权限标识');
        }
        return z.string().optional();
      },
      // 类型为菜单/按钮时显示
      show: (values) => values.type !== 'D',
      triggerFields: ['type'],
    },
    fieldName: 'perms',
    help: `控制器中定义的权限字符\n 如: @SaCheckPermission("system:user:import")`,
    label: '权限标识',
  },
  {
    component: 'InputNumber',
    fieldName: 'sort',
    help: '排序, 数字越小越靠前',
    label: '显示排序',
    defaultValue: 0,
    rules: 'required',
  },
  {
    component: 'Input',
    componentProps: (model) => {
      const placeholder =
        model.isFrame === 1
          ? '填写链接地址http(s)://  使用新页面打开'
          : '填写`路由地址`或者`链接地址`  链接默认使用内部iframe内嵌打开';
      return {
        placeholder,
      };
    },
    dependencies: {
      rules: (model) => {
        if (model.isFrame !== 1) {
          return z
            .string({ message: '请输入路由地址' })
            .min(1, '请输入路由地址')
            .refine((val) => !val.startsWith('/'), {
              message: '路由地址不需要带/',
            });
        }
        // 为链接
        return z
          .string({ message: '请输入链接地址' })
          .regex(/^https?:\/\//, { message: '请输入正确的链接地址' });
      },
      // 类型不为按钮时显示
      show: (values) => values?.type !== 'B',
      triggerFields: ['isFrame', 'type'],
    },
    fieldName: 'path',
    help: `路由地址不带/, 如: menu, user\n 链接为http(s)://开头\n 链接默认使用内部iframe打开, 可通过{是否外链}控制打开方式`,
    label: '路由地址',
  },
  {
    component: 'Input',
    componentProps: (model) => {
      return {
        // 为链接时组件disabled
        disabled: model.isFrame === 1,
      };
    },
    defaultValue: '',
    dependencies: {
      rules: (model) => {
        // 非链接时为必填项
        if (model.path && !/^https?:\/\//.test(model.path)) {
          return z
            .string()
            .min(1, { message: '非链接时必填组件路径' })
            .refine((val) => !val.startsWith('/') && !val.endsWith('/'), {
              message: '组件路径开头/末尾不需要带/',
            });
        }
        // 为链接时非必填
        return z.string().optional();
      },
      // 类型为菜单时显示
      show: (values) => values.type === 'M',
      triggerFields: ['type', 'path'],
    },
    fieldName: 'component',
    help: '填写./src/views下的组件路径, 如system/menu/index',
    label: '组件路径',
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
      // 类型不为按钮时显示
      show: (values) => values.type !== 'B',
      triggerFields: ['type'],
    },
    fieldName: 'isFrame',
    help: '外链为http(s)://开头\n 选择是时, 使用新窗口打开页面, 否则iframe从内部打开页面',
    label: '是否外链',
  },
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: [
        { label: '显示', value: 1 },
        { label: '隐藏', value: 0 },
      ],
      optionType: 'button',
    },
    defaultValue: 1,
    dependencies: {
      // 类型不为按钮时显示
      show: (values) => values.type !== 'B',
      triggerFields: ['type'],
    },
    fieldName: 'visible',
    help: '隐藏后不会出现在菜单栏, 但仍然可以访问',
    label: '是否显示',
  },
  {
    component: 'RadioGroup',
    componentProps: {
      buttonStyle: 'solid',
      options: [
        { label: '正常', value: 1 },
        { label: '停用', value: 0 },
      ],
      optionType: 'button',
    },
    defaultValue: 1,
    dependencies: {
      // 类型不为按钮时显示
      show: (values) => values.type !== 'B',
      triggerFields: ['type'],
    },
    fieldName: 'status',
    help: '停用后不会出现在菜单栏, 也无法访问',
    label: '菜单状态',
  },
  {
    component: 'Input',
    componentProps: (model) => ({
      // 为链接时组件disabled
      disabled: model.isFrame === 1,
      placeholder: '必须为json字符串格式',
    }),
    dependencies: {
      // 类型为菜单时显示
      show: (values) => values.type === 'M',
      triggerFields: ['type'],
    },
    fieldName: 'queryParam',
    help: 'vue-router中的query属性\n 如{"name": "xxx", "age": 16}',
    label: '路由参数',
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
      // 类型为菜单时显示
      show: (values) => values.type === 'M',
      triggerFields: ['type'],
    },
    fieldName: 'isCache',
    help: '路由的keepAlive属性',
    label: '是否缓存',
  },
  {
    component: 'Textarea',
    componentProps: {
      placeholder: '请输入备注',
      rows: 3,
    },
    fieldName: 'remark',
    formItemClass: 'col-span-2',
    label: '备注',
  },
  ];
}
