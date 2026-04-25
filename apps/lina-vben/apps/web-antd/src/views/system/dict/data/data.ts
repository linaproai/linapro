import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';
import type { DictData } from '#/api/system/dict/dict-data-model';

import { h } from 'vue';
import { $t } from '#/locales';

import { Tag } from 'ant-design-vue';

import { tagTypes } from '#/components/dict';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'label',
    label: $t('pages.system.dict.data.fields.label'),
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: $t('pages.system.dict.data.fields.label'),
    field: 'label',
    slots: {
      default: ({ row }) => {
        const { label, tagStyle, cssClass } = row as DictData;
        if (!tagStyle) {
          return h('span', { class: cssClass }, label);
        }
        const isDefault = Reflect.has(tagTypes, tagStyle);
        const color = isDefault ? tagTypes[tagStyle]!.color : tagStyle;
        return h(Tag, { color, class: cssClass }, () => label);
      },
    },
  },
  {
    title: $t('pages.system.dict.data.fields.value'),
    field: 'value',
  },
  {
    title: $t('pages.fields.sort'),
    field: 'sort',
  },
  {
    title: $t('pages.common.remark'),
    field: 'remark',
  },
  {
    title: $t('pages.common.createdAt'),
    field: 'createdAt',
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

/** 新增/编辑表单schema */
export const drawerSchema: VbenFormSchema[] = [
  {
    component: 'Input',
    componentProps: {
      disabled: true,
    },
    fieldName: 'dictType',
    label: $t('pages.system.dict.type.fields.type'),
  },
  {
    component: 'Input',
    fieldName: 'tagStyle',
    label: $t('pages.system.dict.data.fields.tagStyle'),
  },
  {
    component: 'Input',
    fieldName: 'label',
    label: $t('pages.system.dict.data.fields.dataLabel'),
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'value',
    label: $t('pages.system.dict.data.fields.dataValue'),
    rules: 'required',
  },
  {
    component: 'InputNumber',
    fieldName: 'sort',
    label: $t('pages.system.dict.data.fields.sortOrder'),
    defaultValue: 0,
  },
  {
    component: 'Input',
    fieldName: 'cssClass',
    label: $t('pages.system.dict.data.fields.cssClass'),
    help: $t('pages.system.dict.data.help.cssClass'),
  },
  {
    component: 'RadioGroup',
    fieldName: 'status',
    label: $t('pages.common.status'),
    defaultValue: 1,
    componentProps: {
      buttonStyle: 'solid',
      optionType: 'button',
      options: [
        { label: $t('pages.status.enabled'), value: 1 },
        { label: $t('pages.status.disabled'), value: 0 },
      ],
    },
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    formItemClass: 'items-start col-span-2',
    label: $t('pages.common.remark'),
    componentProps: {
      rows: 3,
    },
  },
];
