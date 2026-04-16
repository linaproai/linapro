import type { VbenFormSchema } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';
import type { DictData } from '#/api/system/dict/dict-data-model';

import { h } from 'vue';

import { Tag } from 'ant-design-vue';

import { tagTypes } from '#/components/dict';

/** 查询表单schema */
export const querySchema: VbenFormSchema[] = [
  {
    component: 'Input',
    fieldName: 'label',
    label: '字典标签',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    title: '字典标签',
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
    title: '字典键值',
    field: 'value',
  },
  {
    title: '排序',
    field: 'sort',
  },
  {
    title: '备注',
    field: 'remark',
  },
  {
    title: '创建时间',
    field: 'createdAt',
  },
  {
    field: 'action',
    fixed: 'right',
    slots: { default: 'action' },
    title: '操作',
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
    label: '字典类型',
  },
  {
    component: 'Input',
    fieldName: 'tagStyle',
    label: '标签样式',
  },
  {
    component: 'Input',
    fieldName: 'label',
    label: '数据标签',
    rules: 'required',
  },
  {
    component: 'Input',
    fieldName: 'value',
    label: '数据键值',
    rules: 'required',
  },
  {
    component: 'InputNumber',
    fieldName: 'sort',
    label: '显示排序',
    defaultValue: 0,
  },
  {
    component: 'Input',
    fieldName: 'cssClass',
    label: 'CSS类名',
    help: '标签的css样式, 可添加已经编译的css类名',
  },
  {
    component: 'RadioGroup',
    fieldName: 'status',
    label: '状态',
    defaultValue: 1,
    componentProps: {
      buttonStyle: 'solid',
      optionType: 'button',
      options: [
        { label: '正常', value: 1 },
        { label: '停用', value: 0 },
      ],
    },
  },
  {
    component: 'Textarea',
    fieldName: 'remark',
    formItemClass: 'items-start col-span-2',
    label: '备注',
    componentProps: {
      rows: 3,
    },
  },
];
