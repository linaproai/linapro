import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'original',
    label: $t('pages.system.file.fields.originalName'),
  },
  {
    component: 'Select',
    fieldName: 'suffix',
    label: $t('pages.system.file.fields.fileType'),
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'Select',
    fieldName: 'scene',
    label: $t('pages.system.file.fields.scene'),
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'RangePicker',
    fieldName: 'createTime',
    label: $t('pages.system.file.fields.uploadedAt'),
  },
];

/** Supported image extensions for preview */
export const supportImageList = ['jpg', 'jpeg', 'png', 'gif', 'webp'];

export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: $t('pages.system.file.fields.originalName'),
    field: 'original',
    showOverflow: true,
  },
  {
    title: $t('pages.system.file.fields.fileType'),
    field: 'suffix',
    width: 100,
  },
  {
    title: $t('pages.system.file.fields.scene'),
    field: 'scene',
    width: 120,
    slots: { default: 'scene' },
  },
  {
    title: $t('pages.system.file.fields.preview'),
    field: 'url',
    showOverflow: true,
    slots: { default: 'url' },
  },
  {
    title: $t('pages.system.file.fields.size'),
    field: 'size',
    width: 120,
    sortable: true,
    slots: { default: 'size' },
  },
  {
    title: $t('pages.system.file.fields.uploadedAt'),
    field: 'createdAt',
    sortable: true,
    width: 180,
  },
  {
    title: $t('pages.system.file.fields.uploader'),
    field: 'createdByName',
    width: 120,
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
