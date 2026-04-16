<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  dictTypeAdd,
  dictTypeInfo,
  dictTypeUpdate,
} from '#/api/system/dict/dict-type';

import { modalSchema } from './data';

const emit = defineEmits<{ reload: [] }>();

const isEdit = ref(false);
const recordId = ref<number>(0);
const title = computed(() => (isEdit.value ? '编辑字典类型' : '新增字典类型'));

const [BasicForm, formApi] = useVbenForm({
  schema: modalSchema,
  showDefaultActions: false,
});

const [BasicModal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen) => {
    if (!isOpen) {
      return;
    }
    modalApi.setState({ loading: true });

    const { id } = modalApi.getData() as { id?: number };
    isEdit.value = !!id;
    recordId.value = id || 0;

    if (isEdit.value && id) {
      const record = await dictTypeInfo(id);
      await formApi.setValues(record);
    }

    modalApi.setState({ loading: false });
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const data = await formApi.getValues();
    if (isEdit.value) {
      await dictTypeUpdate(recordId.value, data);
      message.success('更新成功');
    } else {
      await dictTypeAdd(data);
      message.success('创建成功');
    }
    emit('reload');
    modalApi.close();
  } catch (error) {
    console.error(error);
  } finally {
    modalApi.lock(false);
  }
}

async function handleClosed() {
  await formApi.resetForm();
}
</script>

<template>
  <BasicModal :title="title">
    <BasicForm />
  </BasicModal>
</template>
