import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { $t } from '#/locales';

type GridColumns = NonNullable<VxeGridProps['columns']>;
type GridColumn = GridColumns[number];

/** 查询表单schema */
export function querySchema(tenantEnabled = false): VbenFormSchema[] {
  const schema: VbenFormSchema[] = [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('pages.system.user.labels.userAccount'),
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: $t('pages.system.user.labels.userNickname'),
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: $t('pages.fields.phone'),
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('pages.system.user.labels.userStatus'),
    },
    {
      component: 'RangePicker',
      fieldName: 'createdAt',
      label: $t('pages.common.createdAt'),
    },
  ];

  if (tenantEnabled) {
    schema.splice(3, 0, {
      component: 'Select',
      fieldName: 'tenantId',
      label: $t('pages.system.user.labels.tenantMemberships'),
      componentProps: {
        'data-testid': 'user-tenant-filter',
        optionFilterProp: 'label',
        placeholder: $t('pages.multiTenant.placeholders.selectTenant'),
        showSearch: true,
      },
    });
  }

  return schema;
}

function buildDeptColumn(): GridColumn {
  return {
    field: 'deptName',
    title: $t('pages.fields.dept'),
    minWidth: 120,
    formatter({ cellValue }: { cellValue?: string }) {
      return cellValue || $t('pages.system.user.labels.unassignedDept');
    },
  };
}

function buildTenantColumn(): GridColumn {
  return {
    field: 'tenantNames',
    title: $t('pages.system.user.labels.tenantMemberships'),
    minWidth: 180,
    formatter({ cellValue, row }: { cellValue?: string[]; row?: any }) {
      if (Array.isArray(cellValue) && cellValue.length > 0) {
        return cellValue.join(', ');
      }
      if (row?.tenantName) {
        return row.tenantName;
      }
      if (row?.tenantId === 0) {
        return $t('pages.multiTenant.platformScope');
      }
      return $t('pages.system.user.labels.na');
    },
  };
}

/** 表格列定义 */
export function buildColumns(
  orgEnabled: boolean,
  tenantEnabled = false,
): GridColumns {
  const columns: GridColumns = [
    { type: 'checkbox', width: 60 },
    {
      field: 'username',
      title: $t('pages.system.user.labels.account'),
      minWidth: 180,
      sortable: true,
    },
    {
      field: 'avatar',
      title: $t('pages.fields.avatar'),
      slots: { default: 'avatar' },
      minWidth: 80,
    },
    {
      field: 'nickname',
      title: $t('pages.fields.nickname'),
      minWidth: 160,
      sortable: true,
    },
    {
      field: 'roleNames',
      title: $t('pages.fields.roles'),
      minWidth: 220,
      formatter({ cellValue }) {
        return cellValue
          ? String(cellValue)
          : $t('pages.system.user.labels.unassignedRole');
      },
    },
    {
      field: 'phone',
      title: $t('pages.fields.phone'),
      formatter({ cellValue }) {
        return cellValue || $t('pages.system.user.labels.na');
      },
      minWidth: 130,
      sortable: true,
    },
    {
      field: 'sex',
      title: $t('pages.fields.sex'),
      minWidth: 80,
      formatter({ cellValue }) {
        const map: Record<number, string> = {
          0: $t('pages.status.unknown'),
          1: $t('pages.status.male'),
          2: $t('pages.status.female'),
        };
        return map[cellValue as number] ?? $t('pages.status.unknown');
      },
    },
    {
      field: 'email',
      title: $t('pages.fields.email'),
      minWidth: 240,
      sortable: true,
    },
    {
      field: 'status',
      title: $t('pages.common.status'),
      minWidth: 100,
      slots: { default: 'status' },
      sortable: true,
    },
    {
      field: 'createdAt',
      title: $t('pages.common.createdAt'),
      minWidth: 180,
      sortable: true,
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

  if (orgEnabled) {
    columns.splice(4, 0, buildDeptColumn());
  }
  if (tenantEnabled) {
    const roleColumnIndex = columns.findIndex(
      (column) => 'field' in column && column.field === 'roleNames',
    );
    columns.splice(
      roleColumnIndex >= 0 ? roleColumnIndex + 1 : 5,
      0,
      buildTenantColumn(),
    );
  }

  return columns;
}

/** 新增/编辑表单schema */
export function drawerSchema(
  isEdit: boolean,
  orgEnabled: boolean,
  tenantEnabled = false,
  tenantReadonly = false,
): VbenFormSchema[] {
  const schema: VbenFormSchema[] = [
    {
      component: 'Input',
      fieldName: 'username',
      label: $t('pages.system.user.labels.account'),
      rules: 'required',
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.username'),
        disabled: isEdit,
      },
    },
    {
      component: 'InputPassword',
      fieldName: 'password',
      label: $t('pages.fields.password'),
      rules: isEdit ? undefined : 'required',
      componentProps: {
        placeholder: isEdit
          ? $t('pages.system.user.placeholders.passwordKeep')
          : $t('pages.system.user.placeholders.password'),
      },
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: $t('pages.fields.nickname'),
      rules: 'required',
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.nickname'),
      },
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: $t('pages.fields.email'),
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.email'),
      },
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: $t('pages.fields.phone'),
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.phone'),
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'sex',
      label: $t('pages.fields.sex'),
      defaultValue: 0,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: [
          { label: $t('pages.status.unknown'), value: 0 },
          { label: $t('pages.status.male'), value: 1 },
          { label: $t('pages.status.female'), value: 2 },
        ],
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: $t('pages.common.status'),
      defaultValue: 1,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
    },
  ];

  if (orgEnabled) {
    schema.push(
      {
        component: 'TreeSelect',
        defaultValue: undefined,
        fieldName: 'deptId',
        label: $t('pages.fields.dept'),
        componentProps: {
          fieldNames: {
            key: 'id',
            value: 'id',
            children: 'children',
          },
          showSearch: true,
          treeDefaultExpandAll: true,
          treeNodeLabelProp: 'fullName',
          treeLine: { showLeafIcon: false },
          treeNodeFilterProp: 'label',
          placeholder: $t('pages.system.user.placeholders.selectDept'),
        },
      },
      {
        component: 'Select',
        fieldName: 'postIds',
        label: $t('pages.system.user.labels.positions'),
        help: $t('pages.system.user.help.positions'),
        componentProps: {
          mode: 'multiple',
          optionFilterProp: 'label',
          placeholder: $t('pages.system.user.placeholders.selectDeptFirst'),
        },
      },
    );
  }

  if (tenantEnabled) {
    schema.push({
      component: 'Select',
      fieldName: 'tenantIds',
      label: $t('pages.system.user.labels.tenantMemberships'),
      help: $t('pages.system.user.help.tenantMemberships'),
      componentProps: {
        'data-testid': 'user-drawer-tenant-select',
        allowClear: true,
        disabled: tenantReadonly,
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: $t('pages.multiTenant.placeholders.selectTenant'),
        showSearch: true,
      },
    });
  }

  schema.push(
    {
      component: 'Select',
      fieldName: 'roleIds',
      label: $t('pages.fields.roles'),
      help: $t('pages.system.user.help.roles'),
      componentProps: {
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: $t('pages.system.user.placeholders.selectRole'),
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('pages.common.remark'),
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.remark'),
        rows: 3,
      },
    },
  );

  return schema;
}
