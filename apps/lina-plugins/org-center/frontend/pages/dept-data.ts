import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { h } from 'vue';

import { $t } from '#/locales';
import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

function resolveDictOptions(dictType: string) {
  return useDictStore().dictOptionsMap.get(dictType) || [];
}

/** 查询表单schema */
export function buildQuerySchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('plugin.org-center.dept.fields.name'),
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('plugin.org-center.dept.fields.status'),
    },
  ];
}

/** 表格列定义 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    {
      field: 'name',
      title: $t('plugin.org-center.dept.fields.name'),
      treeNode: true,
      minWidth: 200,
    },
    {
      field: 'code',
      title: $t('plugin.org-center.dept.fields.code'),
      minWidth: 120,
    },
    {
      field: 'orderNum',
      title: $t('pages.fields.sort'),
      width: 100,
    },
    {
      field: 'status',
      title: $t('pages.common.status'),
      width: 120,
      slots: {
        default: ({ row }) => {
          const dicts = resolveDictOptions('sys_normal_disable');
          return h(DictTag, { dicts: dicts as any, value: row.status });
        },
      },
    },
    {
      field: 'createdAt',
      title: $t('pages.common.createdAt'),
      minWidth: 180,
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
export function drawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'TreeSelect',
      fieldName: 'parentId',
      label: $t('plugin.org-center.dept.fields.parentDept'),
      rules: 'selectRequired',
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('plugin.org-center.dept.fields.name'),
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: $t('plugin.org-center.dept.fields.code'),
    },
    {
      component: 'InputNumber',
      fieldName: 'orderNum',
      label: $t('plugin.org-center.dept.fields.sortOrder'),
      rules: 'required',
      defaultValue: 0,
    },
    {
      component: 'Select',
      componentProps: {
        allowClear: true,
      },
      fieldName: 'leader',
      label: $t('plugin.org-center.dept.fields.leader'),
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: $t('plugin.org-center.dept.fields.phone'),
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: $t('pages.fields.email'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
      defaultValue: 1,
      fieldName: 'status',
      label: $t('pages.common.status'),
    },
  ];
}
