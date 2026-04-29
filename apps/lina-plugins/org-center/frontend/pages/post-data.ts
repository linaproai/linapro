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
      fieldName: 'code',
      label: $t('plugin.org-center.post.fields.code'),
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('plugin.org-center.post.fields.name'),
    },
    {
      component: 'Select',
      fieldName: 'status',
      label: $t('pages.common.status'),
    },
  ];
}

/** 表格列定义 */
export function buildColumns(): VxeGridProps['columns'] {
  return [
    { type: 'checkbox', width: 60 },
    {
      field: 'code',
      title: $t('plugin.org-center.post.fields.code'),
      minWidth: 120,
    },
    {
      field: 'name',
      title: $t('plugin.org-center.post.fields.name'),
      minWidth: 120,
    },
    {
      field: 'sort',
      title: $t('pages.fields.sort'),
      minWidth: 80,
    },
    {
      field: 'status',
      title: $t('pages.common.status'),
      minWidth: 100,
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
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      fixed: 'right',
      resizable: false,
      width: 'auto',
    },
  ];
}

/** 新增/编辑表单schema */
export function buildDrawerSchema(): VbenFormSchema[] {
  return [
    {
      component: 'TreeSelect',
      fieldName: 'deptId',
      label: $t('plugin.org-center.post.fields.dept'),
      rules: 'selectRequired',
      formItemClass: 'col-span-2',
    },
    {
      component: 'Input',
      fieldName: 'name',
      label: $t('plugin.org-center.post.fields.name'),
      rules: 'required',
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: $t('plugin.org-center.post.fields.code'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      fieldName: 'sort',
      label: $t('plugin.org-center.post.fields.sortOrder'),
      defaultValue: 0,
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: $t('pages.common.status'),
      defaultValue: 1,
      formItemClass: 'col-span-2',
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: $t('pages.common.remark'),
      formItemClass: 'col-span-2',
      componentProps: {
        rows: 3,
      },
    },
  ];
}
