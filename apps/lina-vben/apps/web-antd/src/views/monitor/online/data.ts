import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'username',
    label: '用户账号',
  },
  {
    component: 'Input',
    fieldName: 'ip',
    label: 'IP地址',
  },
];

/** 表格列配置 */
export const columns: VxeGridProps['columns'] = [
  {
    title: '登录账号',
    field: 'username',
  },
  {
    title: '部门名称',
    field: 'deptName',
  },
  {
    title: 'IP地址',
    field: 'ip',
  },
  {
    title: '浏览器',
    field: 'browser',
  },
  {
    title: '操作系统',
    field: 'os',
  },
  {
    title: '登录时间',
    field: 'loginTime',
  },
  {
    field: 'action',
    fixed: 'right',
    slots: { default: 'action' },
    title: '操作',
    resizable: false,
    width: 120,
  },
];
