import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: $t('pages.system.config.fields.name'),
  },
  {
    component: 'Input',
    fieldName: 'key',
    label: $t('pages.system.config.fields.key'),
  },
  {
    component: 'RangePicker',
    fieldName: 'createTime',
    label: $t('pages.common.createdAt'),
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: $t('pages.system.config.fields.name'),
    field: 'name',
  },
  {
    title: $t('pages.system.config.fields.key'),
    field: 'key',
  },
  {
    title: $t('pages.system.config.fields.value'),
    field: 'value',
  },
  {
    title: $t('pages.common.remark'),
    field: 'remark',
  },
  {
    title: $t('pages.common.updatedAt'),
    field: 'updatedAt',
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

/** 新增/编辑表单schema */
export const modalSchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: $t('pages.system.config.fields.name'),
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'key',
    label: $t('pages.system.config.fields.key'),
    rules: 'required',
  },
  {
    component: 'Textarea',
    fieldName: 'value',
    label: $t('pages.system.config.fields.value'),
    rules: 'required',
    componentProps: {
      autoSize: true,
    },
    formItemClass: 'items-start',
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    label: $t('pages.common.remark'),
    componentProps: {
      rows: 3,
    },
    formItemClass: 'items-start',
  },
];
