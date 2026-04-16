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
    fieldName: 'userName',
    label: '用户账号',
  },
  {
    component: 'Input',
    fieldName: 'ip',
    label: 'IP地址',
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '登录状态',
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'RangePicker',
    fieldName: 'loginTime',
    label: '登录日期',
    componentProps: {
      valueFormat: 'YYYY-MM-DD',
    },
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'userName',
    title: '用户账号',
    minWidth: 120,
  },
  {
    field: 'ip',
    title: 'IP地址',
    minWidth: 130,
  },
  {
    field: 'browser',
    title: '浏览器',
    minWidth: 120,
  },
  {
    field: 'os',
    title: '操作系统',
    minWidth: 140,
  },
  {
    field: 'status',
    title: '登录状态',
    minWidth: 100,
    slots: {
      default: ({ row }) => {
        const dicts =
          dictStore.dictOptionsMap.get('sys_oper_status') || [];
        return h(DictTag, { dicts: dicts as any, value: row.status });
      },
    },
  },
  {
    field: 'msg',
    title: '提示信息',
    minWidth: 160,
  },
  {
    field: 'loginTime',
    title: '登录日期',
    minWidth: 180,
    sortable: true,
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
