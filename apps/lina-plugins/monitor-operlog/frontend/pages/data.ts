import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { $t } from '#/locales';
import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

function resolveDictOptions(dictType: string) {
  return useDictStore().getDictOptions(dictType);
}

/** 查询表单schema */
export function buildQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'title',
      label: $t('plugin.monitor-operlog.fields.moduleName'),
    },
    {
      component: 'Input',
      fieldName: 'operName',
      label: $t('plugin.monitor-operlog.fields.operator'),
    },
    {
      component: 'Select',
      fieldName: 'operType',
      label: $t('plugin.monitor-operlog.fields.operType'),
      componentProps: {
        options: [] as { label: string; value: string }[],
      },
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('plugin.monitor-operlog.fields.operResult'),
      componentProps: {
        options: [] as { label: string; value: string }[],
      },
    },
    {
      component: 'RangePicker',
      fieldName: 'operTime',
      label: $t('plugin.monitor-operlog.fields.operTime'),
      componentProps: {
        valueFormat: 'YYYY-MM-DD',
      },
    },
  ];
}

/** 表格列定义 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    { type: 'checkbox', width: 60 },
    {
      field: 'id',
      title: $t('plugin.monitor-operlog.fields.logId'),
      minWidth: 100,
    },
    {
      field: 'title',
      title: $t('plugin.monitor-operlog.fields.moduleName'),
      minWidth: 120,
    },
    {
      field: 'operSummary',
      title: $t('plugin.monitor-operlog.fields.operSummary'),
      minWidth: 140,
    },
    {
      field: 'operType',
      title: $t('plugin.monitor-operlog.fields.operType'),
      minWidth: 100,
      slots: {
        default: ({ row }) => {
          const dicts = resolveDictOptions('sys_oper_type');
          return h(DictTag, { dicts: dicts as any, value: row.operType });
        },
      },
    },
    {
      field: 'operName',
      title: $t('plugin.monitor-operlog.fields.operator'),
      minWidth: 120,
    },
    {
      field: 'operIp',
      title: $t('plugin.monitor-operlog.fields.ipAddress'),
      minWidth: 130,
    },
    {
      field: 'status',
      title: $t('plugin.monitor-operlog.fields.operResult'),
      minWidth: 100,
      slots: {
        default: ({ row }) => {
          const dicts = resolveDictOptions('sys_oper_status');
          return h(DictTag, { dicts: dicts as any, value: row.status });
        },
      },
    },
    {
      field: 'operTime',
      title: $t('plugin.monitor-operlog.fields.operDate'),
      minWidth: 180,
      sortable: true,
    },
    {
      field: 'costTime',
      title: $t('plugin.monitor-operlog.fields.duration'),
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
      title: $t('pages.common.actions'),
      resizable: false,
      width: 'auto',
    },
  ];
}

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
    DELETE: $t('pages.common.delete'),
    GET: $t('pages.common.search'),
    PATCH: $t('plugin.monitor-operlog.method.patch'),
    POST: $t('pages.common.add'),
    PUT: $t('pages.common.edit'),
  };
  return map[method?.toUpperCase()] || method;
}
