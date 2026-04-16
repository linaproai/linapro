<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { postAdd, postDeptTree, postInfo, postUpdate } from '#/api/system/post';
import { useDictStore } from '#/store/dict';

import { drawerSchema } from './data';

const emit = defineEmits<{ reload: [] }>();

const dictStore = useDictStore();

const isUpdate = ref(false);
const postId = ref<number>(0);
const title = computed(() => (isUpdate.value ? '编辑岗位' : '新增岗位'));

const [Form, formApi] = useVbenForm({
  commonConfig: {
    formItemClass: 'col-span-1',
    componentProps: {
      class: 'w-full',
    },
    labelWidth: 80,
  },
  schema: drawerSchema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

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

async function setupDeptSelect() {
  const deptTree = await postDeptTree();
  addFullName(deptTree, 'label', ' / ');
  formApi.updateSchema([
    {
      componentProps: {
        fieldNames: { label: 'label', value: 'id' },
        treeData: deptTree,
        treeDefaultExpandAll: true,
        treeLine: { showLeafIcon: false },
        treeNodeLabelProp: 'fullName',
      },
      fieldName: 'deptId',
    },
  ]);
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (open) {
      drawerApi.setState({ loading: true });
      const data = drawerApi.getData<{ id?: number }>();
      isUpdate.value = !!data?.id;

      await setupDeptSelect();

      // 加载字典：状态选项
      const statusOptions =
        await dictStore.getDictOptionsAsync('sys_normal_disable');
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

      if (isUpdate.value && data?.id) {
        postId.value = data.id;
        const record = await postInfo(data.id);
        await formApi.setValues(record);
      } else {
        postId.value = 0;
        await formApi.resetForm();
      }
      drawerApi.setState({ loading: false });
    }
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const { valid } = await formApi.validate();
      if (!valid) {
        return;
      }
      const values = await formApi.getValues();
      if (isUpdate.value) {
        await postUpdate(postId.value, values);
        message.success('更新成功');
      } else {
        await postAdd(values);
        message.success('创建成功');
      }
      emit('reload');
      drawerApi.close();
    } catch (error) {
      console.error(error);
    } finally {
      drawerApi.lock(false);
    }
  },
  onClosed() {
    formApi.resetForm();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[600px]">
    <Form />
  </Drawer>
</template>
