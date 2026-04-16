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
    fieldName: 'title',
    label: '模块名称',
  },
  {
    component: 'Input',
    fieldName: 'operName',
    label: '操作人员',
  },
  {
    component: 'Select',
    fieldName: 'operType',
    label: '操作类型',
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '操作结果',
    componentProps: {
      options: [] as { label: string; value: string }[],
    },
  },
  {
    component: 'RangePicker',
    fieldName: 'operTime',
    label: '操作时间',
    componentProps: {
      valueFormat: 'YYYY-MM-DD',
    },
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'id',
    title: '日志编号',
    minWidth: 100,
  },
  {
    field: 'title',
    title: '模块名称',
    minWidth: 120,
  },
  {
    field: 'operSummary',
    title: '操作名称',
    minWidth: 140,
  },
  {
    field: 'operType',
    title: '操作类型',
    minWidth: 100,
    slots: {
      default: ({ row }) => {
        const dicts =
          dictStore.dictOptionsMap.get('sys_oper_type') || [];
        return h(DictTag, { dicts: dicts as any, value: row.operType });
      },
    },
  },
  {
    field: 'operName',
    title: '操作人员',
    minWidth: 120,
  },
  {
    field: 'operIp',
    title: 'IP地址',
    minWidth: 130,
  },
  {
    field: 'status',
    title: '操作结果',
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
    field: 'operTime',
    title: '操作日期',
    minWidth: 180,
    sortable: true,
  },
  {
    field: 'costTime',
    title: '操作耗时',
    minWidth: 100,
    sortable: true,
    formatter({ cellValue }) {
      return `${cellValue} ms`;
    },
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

/** 请求方法标签颜色映射 */
export function getMethodTagColor(method: string): string {
  const map: Record<string, string> = {
    GET: 'green',
    POST: 'blue',
    PUT: 'orange',
    DELETE: 'red',
    PATCH: 'cyan',
  };
  return map[method?.toUpperCase()] || 'default';
}

export function getMethodLabel(method: string): string {
  const map: Record<string, string> = {
    DELETE: '删除',
    GET: '查询',
    PATCH: '局部更新',
    POST: '新增',
    PUT: '修改',
  };
  return map[method?.toUpperCase()] || method;
}
