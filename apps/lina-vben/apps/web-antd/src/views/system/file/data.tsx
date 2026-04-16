import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'original',
    label: '原始文件名',
  },
  {
    component: 'Select',
    fieldName: 'suffix',
    label: '文件类型',
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'Select',
    fieldName: 'scene',
    label: '使用场景',
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'RangePicker',
    fieldName: 'createTime',
    label: '上传时间',
  },
];

/** Supported image extensions for preview */
export const supportImageList = ['jpg', 'jpeg', 'png', 'gif', 'webp'];

export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: '原始文件名',
    field: 'original',
    showOverflow: true,
  },
  {
    title: '文件类型',
    field: 'suffix',
    width: 100,
  },
  {
    title: '使用场景',
    field: 'scene',
    width: 120,
    slots: { default: 'scene' },
  },
  {
    title: '文件预览',
    field: 'url',
    showOverflow: true,
    slots: { default: 'url' },
  },
  {
    title: '文件大小',
    field: 'size',
    width: 120,
    sortable: true,
    slots: { default: 'size' },
  },
  {
    title: '上传时间',
    field: 'createdAt',
    sortable: true,
    width: 180,
  },
  {
    title: '上传者',
    field: 'createdByName',
    width: 120,
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
