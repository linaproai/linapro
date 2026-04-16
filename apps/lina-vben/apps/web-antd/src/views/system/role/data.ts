import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

/** 数据权限选项 */
export const dataScopeOptions = [
  { color: 'green', label: '全部数据权限', value: 1 },
  { color: 'default', label: '本部门数据权限', value: 2 },
  { color: 'error', label: '仅本人数据权限', value: 3 },
];

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    componentProps: {
      placeholder: '请输入角色名称',
    },
    fieldName: 'name',
    label: '角色名称',
  },
  {
    component: 'Input',
    componentProps: {
      placeholder: '请输入权限字符',
    },
    fieldName: 'key',
    label: '权限字符',
  },
  {
    component: 'Select',
    componentProps: {
      placeholder: '请选择状态',
      options: [],
    },
    fieldName: 'status',
    label: '状态',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: '角色名称',
    field: 'name',
    minWidth: 120,
  },
  {
    title: '权限字符',
    field: 'key',
    minWidth: 120,
    slots: {
      default: ({ row }) => {
        return h(Tag, { color: 'processing' }, () => row.key);
      },
    },
  },
  {
    title: '数据权限',
    field: 'dataScope',
    minWidth: 120,
    slots: {
      default: ({ row }) => {
        const found = dataScopeOptions.find((item) => item.value === row.dataScope);
        if (found) {
          return h(Tag, { color: found.color }, () => found.label);
        }
        return h(Tag, {}, () => row.dataScope);
      },
    },
  },
  {
    title: '排序',
    field: 'sort',
    width: 80,
  },
  {
    title: '状态',
    field: 'status',
    width: 100,
    slots: { default: 'status' },
  },
  {
    title: '创建时间',
    field: 'createdAt',
    width: 160,
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

/** 新增/编辑表单schema */
export function getDrawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'id',
      label: '角色ID',
      dependencies: {
        show: () => false,
        triggerFields: [''],
      },
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入角色名称',
      },
      fieldName: 'name',
      label: '角色名称',
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '如: admin, user等',
      },
      fieldName: 'key',
      help: '如: admin, simpleUser等',
      label: '权限标识',
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        placeholder: '请输入排序',
        min: 0,
        style: { width: '100%' },
      },
      fieldName: 'sort',
      label: '角色排序',
      rules: 'required',
      defaultValue: 0,
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
      fieldName: 'status',
      help: '修改后, 拥有该角色的用户将自动下线',
      label: '角色状态',
      rules: 'required',
    },
    {
      component: 'RadioGroup',
      fieldName: 'dataScope',
      label: '数据权限',
      help: '更改后需要用户重新登录才能生效',
      rules: 'required',
      defaultValue: 1,
      componentProps: {
        optionType: 'button',
        buttonStyle: 'solid',
        options: dataScopeOptions,
      },
    },
    {
      component: 'Input',
      fieldName: 'menuCheckStrictly',
      label: '菜单权限',
      dependencies: {
        show: () => false,
        triggerFields: [''],
      },
    },
    {
      component: 'Input',
      defaultValue: [],
      fieldName: 'menuIds',
      label: '菜单权限',
      formItemClass: 'col-span-2',
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: '请输入备注',
        rows: 3,
      },
      defaultValue: '',
      fieldName: 'remark',
      formItemClass: 'col-span-2',
      label: '备注',
    },
  ];
}