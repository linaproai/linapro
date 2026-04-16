import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

const dictStore = useDictStore();

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'code',
    label: '岗位编码',
  },
  {
    component: 'Input',
    fieldName: 'name',
    label: '岗位名称',
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '状态',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'code',
    title: '岗位编码',
    minWidth: 120,
  },
  {
    field: 'name',
    title: '岗位名称',
    minWidth: 120,
  },
  {
    field: 'sort',
    title: '排序',
    minWidth: 80,
  },
  {
    field: 'status',
    title: '状态',
    minWidth: 100,
    slots: {
      default: ({ row }) => {
        const dicts = dictStore.dictOptionsMap.get('sys_normal_disable') || [];
        return h(DictTag, { dicts: dicts as any, value: row.status });
      },
    },
  },
  {
    field: 'createdAt',
    title: '创建时间',
    minWidth: 180,
  },
  {
    field: 'action',
    slots: { default: 'action' },
    title: '操作',
    fixed: 'right',
    resizable: false,
    width: 'auto',
  },
];

/** 新增/编辑表单schema */
export const drawerSchema: VbenFormSchema[] = [
  {
    component: 'TreeSelect',
    fieldName: 'deptId',
    label: '所属部门',
    rules: 'selectRequired',
    formItemClass: 'col-span-2',
  },
  {
    component: 'Input',
    fieldName: 'name',
    label: '岗位名称',
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'code',
    label: '岗位编码',
    rules: 'required',
  },
  {
    component: 'InputNumber',
    fieldName: 'sort',
    label: '岗位排序',
    defaultValue: 0,
  },
  {
    component: 'RadioGroup',
    fieldName: 'status',
    label: '状态',
    defaultValue: 1,
    componentProps: {
      buttonStyle: 'solid',
      optionType: 'button',
    },
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    label: '备注',
    formItemClass: 'col-span-2',
    componentProps: {
      rows: 3,
    },
  },
];
