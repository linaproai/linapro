import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

/** Query form schema. */
export function querySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('pages.monitor.online.fields.userAccount'),
    },
    {
      component: 'Input',
      fieldName: 'ip',
      label: $t('pages.monitor.online.fields.ip'),
    },
  ];
}

/** Table column configuration. */
export function columns(): VxeGridProps['columns'] {
  return [
    {
      title: $t('pages.monitor.online.fields.loginAccount'),
      field: 'username',
    },
    {
      title: $t('pages.monitor.online.fields.deptName'),
      field: 'deptName',
    },
    {
      title: $t('pages.monitor.online.fields.ip'),
      field: 'ip',
    },
    {
      title: $t('pages.monitor.online.fields.browser'),
      field: 'browser',
    },
    {
      title: $t('pages.monitor.online.fields.os'),
      field: 'os',
    },
    {
      title: $t('pages.monitor.online.fields.loginTime'),
      field: 'loginTime',
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      resizable: false,
      width: 120,
    },
  ];
}
