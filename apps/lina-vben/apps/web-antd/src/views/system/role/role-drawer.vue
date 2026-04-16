<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { cloneDeep } from '@vben/utils';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { menuTreeSelect, roleMenuTreeSelect } from '#/api/system/menu';
import { roleAdd, roleInfo, roleUpdate } from '#/api/system/role';
import { MenuSelectTable } from '#/components/tree';
import { defaultFormValueGetter, useBeforeCloseDiff } from '#/utils/popup';

import { getDrawerSchema } from './data';

const emit = defineEmits<{ reload: [] }>();

const isUpdate = ref(false);
const title = computed(() => {
  return isUpdate.value ? $t('pages.common.edit') : $t('pages.common.add');
});

const [BasicForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
  },
  layout: 'vertical',
  schema: getDrawerSchema(),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2 gap-x-4',
});

const menuTree = ref<any[]>([]);
async function setupMenuTree(id?: number) {
  if (id) {
    const resp = await roleMenuTreeSelect(id);
    menuTree.value = resp.menus;
    await nextTick();
    await formApi.setFieldValue('menuIds', resp.checkedKeys);
  } else {
    const resp = await menuTreeSelect();
    menuTree.value = resp;
    await nextTick();
    await formApi.setFieldValue('menuIds', []);
  }
}

async function customFormValueGetter() {
  const v = await defaultFormValueGetter(formApi)();
  const menuIds = menuSelectRef.value?.getCheckedKeys?.() ?? [];
  return v + menuIds.join(',');
}

const { onBeforeClose, markInitialized, resetInitialized } = useBeforeCloseDiff(
  {
    initializedGetter: customFormValueGetter,
    currentGetter: customFormValueGetter,
  },
);

const [BasicDrawer, drawerApi] = useVbenDrawer({
  onBeforeClose,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  destroyOnClose: true,
  async onOpenChange(isOpen) {
    if (!isOpen) {
      return null;
    }
    drawerApi.setState({ loading: true });

    const { id } = drawerApi.getData() as { id?: number };
    isUpdate.value = !!id;

    if (isUpdate.value && id) {
      const record = await roleInfo(id);
      await formApi.setValues(record);
    } else {
      // 新增模式：调用 resetForm 以应用 schema 中定义的 defaultValue
      await formApi.resetForm();
    }
    await setupMenuTree(id);
    await markInitialized();

    drawerApi.setState({ loading: false });
  },
});

const menuSelectRef = ref<InstanceType<typeof MenuSelectTable>>();
async function handleConfirm() {
  try {
    drawerApi.setState({ loading: true });

    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const menuIds = menuSelectRef.value?.getCheckedKeys?.() ?? [];
    const data = cloneDeep(await formApi.getValues());
    data.menuIds = menuIds;
    await (isUpdate.value ? roleUpdate(data.id, data) : roleAdd(data));
    message.success(isUpdate.value ? '更新成功' : '创建成功');
    emit('reload');
    resetInitialized();
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

function handleMenuCheckStrictlyChange(value: boolean) {
  formApi.setFieldValue('menuCheckStrictly', value);
}
</script>

<template>
  <BasicDrawer :title="title" class="w-[800px]">
    <BasicForm>
      <template #menuIds="slotProps">
        <div class="h-[600px] w-full">
          <MenuSelectTable
            ref="menuSelectRef"
            :checked-keys="slotProps.value"
            :association="formApi.form.values.menuCheckStrictly"
            :menus="menuTree"
            @update:association="handleMenuCheckStrictlyChange"
          />
        </div>
      </template>
    </BasicForm>
  </BasicDrawer>
</template>