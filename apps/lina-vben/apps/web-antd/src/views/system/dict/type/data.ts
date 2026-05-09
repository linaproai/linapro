import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

import { z } from '#/adapter/form';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'name',
    label: $t('pages.system.dict.type.fields.name'),
  },
  {
    component: 'Input',
    fieldName: 'type',
    label: $t('pages.system.dict.type.fields.type'),
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: $t('pages.system.dict.type.fields.name'),
    field: 'name',
  },
  {
    title: $t('pages.system.dict.type.fields.type'),
    field: 'type',
  },
  {
    title: $t('pages.common.remark'),
    field: 'remark',
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
    label: $t('pages.system.dict.type.fields.name'),
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'type',
    label: $t('pages.system.dict.type.fields.type'),
    help: $t('pages.system.dict.type.help.typeRule'),
    rules: z
      .string()
      .regex(/^[a-z0-9_]+$/i, {
        message: $t('pages.system.dict.type.messages.typeRule'),
      }),
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    label: $t('pages.common.remark'),
    componentProps: {
      rows: 3,
    },
  },
];
