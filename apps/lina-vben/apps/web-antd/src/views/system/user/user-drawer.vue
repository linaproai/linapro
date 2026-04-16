<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { postOptionSelect } from '#/api/system/post';
import { roleOptions } from '#/api/system/role';
import { getDeptTree, userAdd, userInfo, userUpdate } from '#/api/system/user';
import { useDictStore } from '#/store/dict';

import { drawerSchema } from './data';

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const userId = ref<number>(0);
const title = computed(() => (isEdit.value ? '编辑用户' : '新增用户'));

const dictStore = useDictStore();

function addFullName(
  tree: any[],
  labelField = 'label',
  separator = ' / ',
  parentName = '',
) {
  for (const node of tree) {
    node.fullName = parentName
      ? `${parentName}${separator}${node[labelField]}`
      : node[labelField];
    if (node.children?.length) {
      addFullName(node.children, labelField, separator, node.fullName);
    }
  }
  return tree;
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    formItemClass: 'col-span-2',
    componentProps: {
      class: 'w-full',
    },
    labelWidth: 80,
  },
  schema: drawerSchema(false),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

/**
 * 岗位的加载
 */
async function setupPostOptions(deptId: number | string) {
  const postList = await postOptionSelect(Number(deptId));
  const options = postList.map((item) => ({
    label: item.postName,
    value: item.postId,
  }));
  const placeholder = options.length > 0 ? '请选择' : '该部门下暂无岗位';
  formApi.updateSchema([
    {
      componentProps: { options, placeholder },
      fieldName: 'postIds',
    },
  ]);
}

/**
 * 角色的加载
 */
async function setupRoleOptions() {
  const roleList = await roleOptions();
  const options = roleList.map((item) => ({
    label: item.name,
    value: item.id,
  }));
  formApi.updateSchema([
    {
      componentProps: { options, placeholder: '请选择角色' },
      fieldName: 'roleIds',
    },
  ]);
}

/**
 * 初始化部门选择
 */
async function setupDeptSelect() {
  const deptTree = await getDeptTree();
  addFullName(deptTree, 'label', ' / ');
  formApi.updateSchema([
    {
      componentProps: (formModel: any) => ({
        class: 'w-full',
        fieldNames: {
          key: 'id',
          value: 'id',
          children: 'children',
        },
        async onSelect(deptId: number | string) {
          /** 根据部门ID加载岗位 */
          await setupPostOptions(deptId);
          /** 变化后需要重新选择岗位 */
          formModel.postIds = [];
        },
        placeholder: '请选择',
        showSearch: true,
        treeData: deptTree,
        treeDefaultExpandAll: true,
        treeLine: { showLeafIcon: false },
        treeNodeFilterProp: 'label',
        treeNodeLabelProp: 'fullName',
      }),
      fieldName: 'deptId',
    },
  ]);
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      // 需要重置岗位选择
      formApi.updateSchema([
        {
          componentProps: { options: [], placeholder: '请先选择部门' },
          fieldName: 'postIds',
        },
      ]);
      return;
    }

    drawerApi.setState({ loading: true });

    const data = drawerApi.getData<{ isEdit: boolean; row?: any }>();
    isEdit.value = data?.isEdit ?? false;

    // Update schema based on mode
    formApi.setState({
      schema: drawerSchema(isEdit.value),
    });

    // 加载部门树
    await setupDeptSelect();

    // 加载角色选项
    await setupRoleOptions();

    // 加载字典：状态选项
    const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
    formApi.updateSchema([
      {
        fieldName: 'status',
        componentProps: {
          options: statusOptions.map((d) => ({
            label: d.label,
            value: Number(d.value),
          })),
        },
      },
    ]);

    if (isEdit.value && data?.row) {
      userId.value = data.row.id;
      // Load user info
      const user = await userInfo(data.row.id);
      await formApi.setValues({
        username: user.username,
        nickname: user.nickname,
        email: user.email,
        phone: user.phone,
        sex: user.sex,
        status: user.status,
        remark: user.remark,
        deptId: user.deptId,
        postIds: user.postIds,
        roleIds: user.roleIds,
      });
      // 编辑时加载该部门下的岗位
      if (user.deptId) {
        await setupPostOptions(user.deptId);
      }
    } else {
      userId.value = 0;
      await formApi.resetForm();
    }

    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    const values = await formApi.getValues();

    if (isEdit.value) {
      await userUpdate({
        id: userId.value,
        ...values,
      });
      message.success('更新成功');
    } else {
      await userAdd(values as any);
      message.success('创建成功');
    }

    emit('success');
    drawerApi.close();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[600px]">
    <Form />
  </Drawer>
</template>
