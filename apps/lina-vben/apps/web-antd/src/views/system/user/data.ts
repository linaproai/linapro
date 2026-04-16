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
    fieldName: 'nickname',
    label: '用户昵称',
  },
  {
    component: 'Input',
    fieldName: 'phone',
    label: '手机号码',
  },
  {
    component: 'Select',
    fieldName: 'status',
    label: '用户状态',
  },
  {
    component: 'RangePicker',
    fieldName: 'createdAt',
    label: '创建时间',
  },
];

/** 表格列定义 */
export const columns: VxeGridProps['columns'] = [
  { type: 'checkbox', width: 60 },
  {
    field: 'username',
    title: '名称',
    minWidth: 120,
    sortable: true,
  },
  {
    field: 'avatar',
    title: '头像',
    slots: { default: 'avatar' },
    minWidth: 80,
  },
  {
    field: 'nickname',
    title: '昵称',
    minWidth: 120,
    sortable: true,
  },
  {
    field: 'deptName',
    title: '部门',
    minWidth: 120,
    formatter({ cellValue }) {
      return cellValue || '未分配部门';
    },
  },
  {
    field: 'roleNames',
    title: '角色',
    minWidth: 120,
    formatter({ cellValue }) {
      return cellValue || '未分配角色';
    },
  },
  {
    field: 'phone',
    title: '手机号码',
    formatter({ cellValue }) {
      return cellValue || '暂无';
    },
    minWidth: 130,
    sortable: true,
  },
  {
    field: 'sex',
    title: '性别',
    minWidth: 80,
    formatter({ cellValue }) {
      const map: Record<number, string> = { 0: '未知', 1: '男', 2: '女' };
      return map[cellValue as number] ?? '未知';
    },
  },
  {
    field: 'email',
    title: '邮箱',
    minWidth: 160,
    sortable: true,
  },
  {
    field: 'status',
    title: '状态',
    minWidth: 100,
    slots: { default: 'status' },
    sortable: true,
  },
  {
    field: 'createdAt',
    title: '创建时间',
    minWidth: 180,
    sortable: true,
  },
  {
    field: 'action',
    slots: { default: 'action' },
    title: '操作',
    fixed: 'right',
    resizable: false,
    width: 'auto',
  },
];

/** 新增/编辑表单schema */
export function drawerSchema(isEdit: boolean): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'username',
      label: '用户名',
      rules: 'required',
      componentProps: {
        placeholder: '请输入用户名',
        disabled: isEdit,
      },
    },
    {
      component: 'InputPassword',
      fieldName: 'password',
      label: '密码',
      rules: isEdit ? undefined : 'required',
      componentProps: {
        placeholder: isEdit ? '留空则不修改' : '请输入密码',
      },
    },
    {
      component: 'Input',
      fieldName: 'nickname',
      label: '昵称',
      rules: 'required',
      componentProps: {
        placeholder: '请输入昵称',
      },
    },
    {
      component: 'Input',
      fieldName: 'email',
      label: '邮箱',
      componentProps: {
        placeholder: '请输入邮箱',
      },
    },
    {
      component: 'Input',
      fieldName: 'phone',
      label: '手机号码',
      componentProps: {
        placeholder: '请输入手机号码',
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'sex',
      label: '性别',
      defaultValue: 0,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
        options: [
          { label: '未知', value: 0 },
          { label: '男', value: 1 },
          { label: '女', value: 2 },
        ],
      },
    },
    {
      component: 'RadioGroup',
      fieldName: 'status',
      label: '状态',
      defaultValue: 1,
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
    },
    {
      component: 'TreeSelect',
      defaultValue: undefined,
      fieldName: 'deptId',
      label: '部门',
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
        placeholder: '请选择',
      },
    },
    {
      component: 'Select',
      fieldName: 'postIds',
      label: '岗位',
      help: '选择部门后, 将自动加载该部门下所有的岗位',
      componentProps: {
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: '请先选择部门',
      },
    },
    {
      component: 'Select',
      fieldName: 'roleIds',
      label: '角色',
      help: '可分配多个角色给该用户',
      componentProps: {
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: '请选择角色',
      },
    },
    {
      component: 'Textarea',
      fieldName: 'remark',
      label: '备注',
      componentProps: {
        placeholder: '请输入备注',
        rows: 3,
      },
    },
  ];
}
