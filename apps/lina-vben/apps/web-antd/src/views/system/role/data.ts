import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import { getDictOptions } from '#/utils/dict';
import { renderDict } from '#/utils/render';

export const DATA_SCOPE_DICT_TYPE = 'sys_data_scope';
export const DATA_SCOPE_ALL = 1;
export const DATA_SCOPE_TENANT = 2;
export const DATA_SCOPE_DEPT = 3;
export const DATA_SCOPE_SELF = 4;

/** 数据权限选项 */
export function getDataScopeOptions(orgEnabled = true, tenantEnabled = true) {
  let options = getDictOptions(DATA_SCOPE_DICT_TYPE).map((item) => ({
    ...item,
    value: Number(item.value),
  }));
  if (!tenantEnabled) {
    options = options.filter((item) => item.value !== DATA_SCOPE_TENANT);
  }
  if (!orgEnabled) {
    options = options.filter((item) => item.value !== DATA_SCOPE_DEPT);
  }
  return options;
}

/** 数据权限默认值 */
export function getDefaultDataScope(tenantEnabled = true) {
  return tenantEnabled ? DATA_SCOPE_TENANT : DATA_SCOPE_ALL;
}

/** 规范化当前上下文不可用的数据权限值 */
export function normalizeDataScopeValue(
  value: number,
  orgEnabled = true,
  tenantEnabled = true,
) {
  if (!tenantEnabled && value === DATA_SCOPE_TENANT) {
    return DATA_SCOPE_ALL;
  }
  if (!orgEnabled && value === DATA_SCOPE_DEPT) {
    return DATA_SCOPE_SELF;
  }
  const allowedValues = new Set(
    getDataScopeOptions(orgEnabled, tenantEnabled).map((item) => item.value),
  );
  return allowedValues.has(value) ? value : getDefaultDataScope(tenantEnabled);
}

/** 查询表单schema */
export function querySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.roleName'),
      },
      fieldName: 'name',
      label: $t('pages.system.role.fields.roleName'),
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.permissionKey'),
      },
      fieldName: 'key',
      label: $t('pages.system.role.fields.permissionKey'),
    },
    {
      component: 'Select',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.status'),
        options: [],
      },
      fieldName: 'status',
      label: $t('pages.common.status'),
    },
  ];
}

/** 表格列定义 */
export function columns(
  orgEnabled = true,
  tenantEnabled = true,
): VxeGridProps['columns'] {
  return [
    { type: 'checkbox', width: 60 },
    {
      title: $t('pages.system.role.fields.roleName'),
      field: 'name',
      minWidth: 120,
    },
    {
      title: $t('pages.system.role.fields.permissionKey'),
      field: 'key',
      minWidth: 120,
      slots: {
        default: ({ row }) => {
          return h(Tag, { color: 'processing' }, () => row.key);
        },
      },
    },
    {
      title: $t('pages.system.role.fields.dataScope'),
      field: 'dataScope',
      minWidth: 120,
      slots: {
        default: ({ row }) => {
          return renderDict(
            normalizeDataScopeValue(
              Number(row.dataScope),
              orgEnabled,
              tenantEnabled,
            ),
            DATA_SCOPE_DICT_TYPE,
          );
        },
      },
    },
    {
      title: $t('pages.fields.sort'),
      field: 'sort',
      width: 80,
    },
    {
      title: $t('pages.common.status'),
      field: 'status',
      width: 100,
      slots: { default: 'status' },
    },
    {
      title: $t('pages.common.createdAt'),
      field: 'createdAt',
      width: 160,
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

/** 新增/编辑表单schema */
export function getDrawerSchema(
  orgEnabled = true,
  tenantEnabled = true,
): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'id',
      label: $t('pages.system.role.fields.roleId'),
      dependencies: {
        show: () => false,
        triggerFields: [''],
      },
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.roleName'),
      },
      fieldName: 'name',
      label: $t('pages.system.role.fields.roleName'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.permissionExample'),
      },
      fieldName: 'key',
      help: $t('pages.system.role.placeholders.permissionHelp'),
      label: $t('pages.system.role.fields.permissionKey'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.sort'),
        min: 0,
        style: { width: '100%' },
      },
      fieldName: 'sort',
      label: $t('pages.system.role.fields.roleSort'),
      rules: 'required',
      defaultValue: 0,
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: $t('pages.status.enabled'), value: 1 },
          { label: $t('pages.status.disabled'), value: 0 },
        ],
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      help: $t('pages.system.role.help.status'),
      label: $t('pages.system.role.fields.roleStatus'),
      rules: 'required',
    },
    {
      component: 'Select',
      fieldName: 'dataScope',
      label: $t('pages.system.role.fields.dataScope'),
      help: $t('pages.system.role.help.dataScope'),
      rules: 'required',
      defaultValue: getDefaultDataScope(tenantEnabled),
      componentProps: {
        allowClear: false,
        options: getDataScopeOptions(orgEnabled, tenantEnabled),
        placeholder: $t('pages.system.role.placeholders.dataScope'),
      },
    },
    {
      component: 'Input',
      fieldName: 'menuCheckStrictly',
      label: $t('pages.system.role.fields.menuPermissions'),
      dependencies: {
        show: () => false,
        triggerFields: [''],
      },
    },
    {
      component: 'Input',
      defaultValue: [],
      fieldName: 'menuIds',
      label: $t('pages.system.role.fields.menuPermissions'),
      formItemClass: 'col-span-2',
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.remark'),
        rows: 3,
      },
      defaultValue: '',
      fieldName: 'remark',
      formItemClass: 'col-span-2',
      label: $t('pages.common.remark'),
    },
  ];
}
