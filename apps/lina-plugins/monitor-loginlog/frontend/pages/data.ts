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
      fieldName: 'userName',
      label: $t('plugin.monitor-loginlog.fields.userName'),
    },
    {
      component: 'Input',
      fieldName: 'ip',
      label: $t('plugin.monitor-loginlog.fields.ipAddress'),
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('plugin.monitor-loginlog.fields.status'),
      componentProps: {
        options: [] as { label: string; value: string }[],
      },
    },
    {
      component: 'RangePicker',
      fieldName: 'loginTime',
      label: $t('plugin.monitor-loginlog.fields.loginDate'),
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
      field: 'userName',
      title: $t('plugin.monitor-loginlog.fields.userName'),
      minWidth: 120,
    },
    {
      field: 'ip',
      title: $t('plugin.monitor-loginlog.fields.ipAddress'),
      minWidth: 130,
    },
    {
      field: 'browser',
      title: $t('plugin.monitor-loginlog.fields.browser'),
      minWidth: 120,
    },
    {
      field: 'os',
      title: $t('plugin.monitor-loginlog.fields.os'),
      minWidth: 140,
    },
    {
      field: 'status',
      title: $t('plugin.monitor-loginlog.fields.status'),
      minWidth: 100,
      slots: {
        default: ({ row }) => {
          const dicts = resolveDictOptions('sys_login_status');
          return h(DictTag, { dicts: dicts as any, value: row.status });
        },
      },
    },
    {
      field: 'msg',
      title: $t('plugin.monitor-loginlog.fields.message'),
      minWidth: 160,
    },
    {
      field: 'loginTime',
      title: $t('plugin.monitor-loginlog.fields.loginDate'),
      minWidth: 180,
      sortable: true,
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
