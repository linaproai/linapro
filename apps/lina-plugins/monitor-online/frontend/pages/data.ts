import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

/** 查询表单schema */
export function buildQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('plugin.monitor-online.page.fields.userAccount'),
    },
    {
      component: 'Input',
      fieldName: 'ip',
      label: $t('plugin.monitor-online.page.fields.ipAddress'),
    },
  ];
}

/** 表格列配置 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    {
      title: $t('plugin.monitor-online.page.fields.loginAccount'),
      field: 'username',
    },
    {
      title: $t('plugin.monitor-online.page.fields.departmentName'),
      field: 'deptName',
    },
    {
      title: $t('plugin.monitor-online.page.fields.ipAddress'),
      field: 'ip',
    },
    {
      title: $t('plugin.monitor-online.page.fields.browser'),
      field: 'browser',
    },
    {
      title: $t('plugin.monitor-online.page.fields.os'),
      field: 'os',
    },
    {
      title: $t('plugin.monitor-online.page.fields.loginTime'),
      field: 'loginTime',
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('plugin.monitor-online.page.fields.actions'),
      resizable: false,
      width: 120,
    },
  ];
}
