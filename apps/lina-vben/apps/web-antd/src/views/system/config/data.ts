import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: '参数名称',
  },
  {
    component: 'Input',
    fieldName: 'key',
    label: '参数键名',
  },
  {
    component: 'RangePicker',
    fieldName: 'createTime',
    label: '创建时间',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: '参数名称',
    field: 'name',
  },
  {
    title: '参数键名',
    field: 'key',
  },
  {
    title: '参数键值',
    field: 'value',
  },
  {
    title: '备注',
    field: 'remark',
  },
  {
    title: '修改时间',
    field: 'updatedAt',
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
    label: '参数名称',
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'key',
    label: '参数键名',
    rules: 'required',
  },
  {
    component: 'Textarea',
    fieldName: 'value',
    label: '参数键值',
    rules: 'required',
    componentProps: {
      autoSize: true,
    },
    formItemClass: 'items-start',
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    label: '备注',
    componentProps: {
      rows: 3,
    },
    formItemClass: 'items-start',
  },
];
