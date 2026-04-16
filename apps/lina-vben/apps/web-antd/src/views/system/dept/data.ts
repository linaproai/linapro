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
    fieldName: 'name',
    label: '部门名称',
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '部门状态',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  {
    field: 'name',
    title: '部门名称',
    treeNode: true,
    minWidth: 200,
  },
  {
    field: 'code',
    title: '部门编码',
    minWidth: 120,
  },
  {
    field: 'orderNum',
    title: '排序',
    width: 100,
  },
  {
    field: 'status',
    title: '状态',
    width: 120,
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
    fixed: 'right',
    slots: { default: 'action' },
    title: '操作',
    resizable: false,
    width: 'auto',
  },
];

/** 新增/编辑表单schema */
export function drawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'TreeSelect',
      fieldName: 'parentId',
      label: '上级部门',
      rules: 'selectRequired',
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: '部门名称',
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '部门编码',
    },
    {
      component: 'InputNumber',
      fieldName: 'orderNum',
      label: '显示排序',
      rules: 'required',
      defaultValue: 0,
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
      },
      fieldName: 'leader',
      label: '负责人',
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: '联系电话',
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: '邮箱',
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: '状态',
    },
  ];
}
