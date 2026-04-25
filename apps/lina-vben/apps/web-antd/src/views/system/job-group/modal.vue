<script setup lang="ts">
import type { JobGroupPayload, JobGroupRecord } from '#/api/system/jobGroup/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  jobGroupCreate,
  jobGroupUpdate,
} from '#/api/system/jobGroup';

const emit = defineEmits<{ reload: [] }>();

const currentRecord = ref<JobGroupRecord | null>(null);
const title = computed(() =>
  currentRecord.value
    ? $t('pages.system.jobGroup.drawer.editTitle')
    : $t('pages.system.jobGroup.drawer.createTitle'),
);

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-1',
    labelWidth: 88,
  },
  schema: [
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.jobGroup.placeholders.code'),
      },
      fieldName: 'code',
      label: $t('pages.system.jobGroup.fields.code'),
      rules: 'required',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.jobGroup.placeholders.name'),
      },
      fieldName: 'name',
      label: $t('pages.system.jobGroup.fields.name'),
      rules: 'required',
    },
    {
      component: 'InputNumber',
      componentProps: {
        min: 0,
        precision: 0,
        style: { width: '100%' },
      },
      defaultValue: 0,
      fieldName: 'sortOrder',
      label: $t('pages.system.jobGroup.fields.sortOrder'),
      rules: 'required',
    },
    {
      component: 'Textarea',
      componentProps: {
        placeholder: $t('pages.system.jobGroup.placeholders.remark'),
        rows: 3,
      },
      fieldName: 'remark',
      formItemClass: 'col-span-2',
      label: $t('pages.common.remark'),
    },
  ],
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: async () => {
    currentRecord.value = null;
    await formApi.resetForm();
  },
  onConfirm: handleConfirm,
  onOpenChange: async (open) => {
    if (!open) {
      return;
    }
    modalApi.setState({ loading: true });
    const data = modalApi.getData<{ record?: JobGroupRecord }>();
    currentRecord.value = data?.record ?? null;
    if (currentRecord.value) {
      await formApi.setValues({
        code: currentRecord.value.code,
        name: currentRecord.value.name,
        remark: currentRecord.value.remark,
        sortOrder: currentRecord.value.sortOrder,
      });
      formApi.updateSchema([
        {
          componentProps: {
            disabled: currentRecord.value.isDefault === 1,
            placeholder: $t('pages.system.jobGroup.placeholders.code'),
          },
          fieldName: 'code',
        },
      ]);
    } else {
      formApi.updateSchema([
        {
          componentProps: {
            disabled: false,
            placeholder: $t('pages.system.jobGroup.placeholders.code'),
          },
          fieldName: 'code',
        },
      ]);
      await formApi.setValues({
        code: '',
        name: '',
        remark: '',
        sortOrder: 0,
      });
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
    const values = await formApi.getValues<JobGroupPayload>();
    if (currentRecord.value) {
      await jobGroupUpdate(currentRecord.value.id, values);
      message.success($t('pages.common.updateSuccess'));
    } else {
      await jobGroupCreate(values);
      message.success($t('pages.common.createSuccess'));
    }
    emit('reload');
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal
    :title="title"
    class="w-[640px]"
    data-testid="job-group-modal"
  >
    <Form />
  </Modal>
</template>
