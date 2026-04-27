import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

/** 查询表单schema */
export function buildQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'title',
      label: $t('plugin.content-notice.fields.title'),
    },
    {
      component: 'Select',
      fieldName: 'type',
      label: $t('plugin.content-notice.fields.type'),
      componentProps: {
        options: [] as { label: string; value: number }[],
      },
    },
    {
      component: 'Input',
      fieldName: 'createdBy',
      label: $t('plugin.content-notice.fields.createdBy'),
    },
  ];
}

/** 表格列定义 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    { type: 'checkbox', width: 60 },
    {
      field: 'title',
      title: $t('plugin.content-notice.fields.title'),
      minWidth: 200,
    },
    {
      field: 'type',
      title: $t('plugin.content-notice.fields.type'),
      minWidth: 100,
      slots: { default: 'type' },
    },
    {
      field: 'status',
      title: $t('pages.common.status'),
      minWidth: 100,
      slots: { default: 'status' },
    },
    {
      field: 'createdByName',
      title: $t('plugin.content-notice.fields.createdBy'),
      minWidth: 120,
    },
    {
      field: 'createdAt',
      title: $t('pages.common.createdAt'),
      minWidth: 180,
    },
    {
      field: 'action',
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      fixed: 'right',
      resizable: false,
      width: 'auto',
    },
  ];
}
