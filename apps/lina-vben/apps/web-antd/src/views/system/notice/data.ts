import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'title',
    label: '公告标题',
  },
  {
    component: 'Select',
    fieldName: 'type',
    label: '公告类型',
    componentProps: {
      options: [
        { label: '通知', value: 1 },
        { label: '公告', value: 2 },
      ],
    },
  },
  {
    component: 'Input',
    fieldName: 'createdBy',
    label: '创建人',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'title',
    title: '公告标题',
    minWidth: 200,
  },
  {
    field: 'type',
    title: '公告类型',
    minWidth: 100,
    slots: { default: 'type' },
  },
  {
    field: 'status',
    title: '状态',
    minWidth: 100,
    slots: { default: 'status' },
  },
  {
    field: 'createdByName',
    title: '创建人',
    minWidth: 120,
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
