import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { z } from '#/adapter/form';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: '字典名称',
  },
  {
    component: 'Input',
    fieldName: 'type',
    label: '字典类型',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: '字典名称',
    field: 'name',
  },
  {
    title: '字典类型',
    field: 'type',
  },
  {
    title: '备注',
    field: 'remark',
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

/** 新增/编辑表单schema */
export const modalSchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: '字典名称',
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'type',
    label: '字典类型',
    help: '使用英文/数字/下划线命名, 如:sys_normal_disable',
    rules: z
      .string()
      .regex(/^[a-z0-9_]+$/i, { message: '字典类型只能使用英文/数字/下划线命名' }),
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    label: '备注',
    componentProps: {
      rows: 3,
    },
  },
];
