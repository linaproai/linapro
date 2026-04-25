<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { configAdd, configInfo, configUpdate } from '#/api/system/config';
import { syncPublicFrontendSettings } from '#/runtime/public-frontend';

import { modalSchema } from './data';

const emit = defineEmits<{ reload: [] }>();

const isEdit = ref(false);
const recordId = ref<number>(0);
const title = computed(() =>
  isEdit.value
    ? $t('pages.system.config.drawer.editTitle')
    : $t('pages.system.config.drawer.createTitle'),
);

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
      const record = await configInfo(id);
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
      await configUpdate(recordId.value, data);
      await syncPublicFrontendSettings();
      message.success($t('pages.common.updateSuccess'));
    } else {
      await configAdd(data);
      await syncPublicFrontendSettings();
      message.success($t('pages.common.createSuccess'));
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
