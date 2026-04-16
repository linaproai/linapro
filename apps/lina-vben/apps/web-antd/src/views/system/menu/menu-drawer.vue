<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { cloneDeep, getPopupContainer } from '@vben/utils';

import { Skeleton } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { menuAdd, menuInfo, menuList, menuUpdate } from '#/api/system/menu';
import { addFullName, getDescendantIds, listToTree, treeToList } from '#/utils/tree';
import { defaultFormValueGetter, useBeforeCloseDiff } from '#/utils/popup';

import { drawerSchema } from './data';

interface ModalProps {
  id?: number | string;
  parentId?: number;
  update: boolean;
}

const emit = defineEmits<{ reload: [] }>();

const isUpdate = ref(false);
const title = computed(() => {
  return isUpdate.value ? $t('pages.common.edit') : $t('pages.common.add');
});
const loading = ref(false);

const [BasicForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-2',
    labelWidth: 90,
  },
  schema: drawerSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

async function setupMenuSelect(currentId?: number | string) {
  // menu API returns tree structure
  const menuTree = await menuList();
  /**
   * 过滤掉按钮类型
   * 不允许在按钮下添加数据
   * 需要先展平树形结构，过滤后再重建树
   */
  const flatList = treeToList(menuTree, { childProp: 'children' });
  const filteredList = flatList.filter((item) => item.type !== 'B');
  const rebuiltTree = listToTree(filteredList, { id: 'id', pid: 'parentId' });

  // 获取需要禁用的节点ID（当前菜单及其子孙节点）
  const disabledIds = currentId ? getDescendantIds(rebuiltTree, currentId) : [];

  // 为需要禁用的节点设置 disabled 属性
  const disabledIdSet = new Set(disabledIds.map(String));
  const setDisabled = (nodes: any[]) => {
    for (const node of nodes) {
      if (disabledIdSet.has(String(node.id))) {
        node.disabled = true;
      }
      if (node.children && node.children.length > 0) {
        setDisabled(node.children);
      }
    }
  };
  setDisabled(rebuiltTree);

  const fullMenuTree = [
    {
      id: 0,
      name: '根菜单',
      children: rebuiltTree,
    },
  ];
  addFullName(fullMenuTree, 'name', ' / ');

  formApi.updateSchema([
    {
      componentProps: {
        fieldNames: {
          label: 'name',
          value: 'id',
        },
        getPopupContainer,
        // 设置弹窗滚动高度 默认256
        listHeight: 300,
        showSearch: true,
        treeData: fullMenuTree,
        treeDefaultExpandAll: false,
        // 默认展开的树节点
        treeDefaultExpandedKeys: [0],
        treeLine: { showLeafIcon: false },
        // 筛选的字段
        treeNodeFilterProp: 'name',
        treeNodeLabelProp: 'fullName',
      },
      fieldName: 'parentId',
    },
  ]);
}

const { onBeforeClose, markInitialized, resetInitialized } = useBeforeCloseDiff(
  {
    initializedGetter: defaultFormValueGetter(formApi),
    currentGetter: defaultFormValueGetter(formApi),
  },
);

const [BasicDrawer, drawerApi] = useVbenDrawer({
  onBeforeClose,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  async onOpenChange(isOpen) {
    if (!isOpen) {
      return null;
    }
    drawerApi.setState({ loading: true });
    loading.value = true;

    const data = drawerApi.getData() as ModalProps;
    isUpdate.value = data?.update ?? false;

    // 加载菜单树选择，编辑时传入当前菜单ID以禁用自身及子孙节点
    const currentId = data?.update && data.id ? Number(data.id) : undefined;
    await setupMenuSelect(currentId);

    if (data?.update && data.id) {
      const record = await menuInfo(Number(data.id));
      await formApi.setValues(record);
    } else {
      await formApi.resetForm();
      // 新增子菜单时，设置上级菜单（必须在 resetForm 之后设置，否则会被重置）
      if (data?.parentId) {
        await formApi.setFieldValue('parentId', data.parentId);
      }
    }
    await markInitialized();

    drawerApi.setState({ loading: false });
    loading.value = false;
  },
});

async function handleConfirm() {
  try {
    drawerApi.setState({ loading: true });
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const data = cloneDeep(await formApi.getValues());
    await (isUpdate.value ? menuUpdate(data.id, data) : menuAdd(data));
    resetInitialized();
    emit('reload');
    drawerApi.close();
  } catch (error) {
    console.error(error);
  } finally {
    drawerApi.setState({ loading: false });
  }
}

async function handleClosed() {
  await formApi.resetForm();
  resetInitialized();
}
</script>

<template>
  <BasicDrawer :title="title" class="w-[600px]">
    <Skeleton active v-if="loading" />
    <BasicForm v-show="!loading" />
  </BasicDrawer>
</template>
