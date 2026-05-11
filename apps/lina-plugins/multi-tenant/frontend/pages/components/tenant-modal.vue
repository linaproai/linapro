<script setup lang="ts">
import type { PlatformTenant } from '../tenant-client';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm, z } from '#/adapter/form';
import { $t } from '#/locales';

import {
  platformTenantCreate,
  platformTenantUpdate,
} from '../tenant-client';

const emit = defineEmits<{ success: [] }>();

interface TenantFormValues {
  code: string;
  name: string;
  remark: string;
}

const recordId = ref(0);
const formValues: TenantFormValues = {
  code: '',
  name: '',
  remark: '',
};

const isEdit = computed(() => recordId.value > 0);
const title = computed(() =>
  isEdit.value
    ? $t('pages.multiTenant.tenant.actions.edit')
    : $t('pages.multiTenant.tenant.actions.create'),
);

const [TenantForm, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    labelWidth: 90,
  },
  schema: [
    {
      component: 'Input',
      componentProps: {
        'data-testid': 'tenant-code-input',
      },
      fieldName: 'code',
      label: $t('pages.multiTenant.fields.code'),
      rules: z.string().regex(/^[a-z0-9-]{2,32}$/, {
        message: $t('pages.multiTenant.messages.codeRule'),
      }),
    },
    {
      component: 'Input',
      componentProps: {
        'data-testid': 'tenant-name-input',
      },
      fieldName: 'name',
      label: $t('pages.multiTenant.fields.name'),
      rules: z.string().min(1, {
        message: $t('pages.multiTenant.messages.nameRequired'),
      }),
    },
    {
      component: 'Textarea',
      componentProps: {
        'data-testid': 'tenant-remark-input',
        rows: 3,
      },
      fieldName: 'remark',
      formItemClass: 'items-start',
      label: $t('pages.common.remark'),
    },
  ],
  showDefaultActions: false,
});

const [Modal, modalApi] = useVbenModal({
  fullscreenButton: false,
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const values = await formApi.getValues<TenantFormValues>();
    if (isEdit.value) {
      await platformTenantUpdate(recordId.value, {
        name: values.name,
        remark: values.remark,
      });
      message.success($t('pages.common.updateSuccess'));
    } else {
      await platformTenantCreate({
        code: values.code,
        name: values.name,
        remark: values.remark,
      });
      message.success($t('pages.common.createSuccess'));
    }
    emit('success');
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<Partial<PlatformTenant>>();
  recordId.value = data?.id ?? 0;
  await formApi.setValues({
    code: data?.code ?? formValues.code,
    name: data?.name ?? formValues.name,
    remark: data?.remark ?? formValues.remark,
  });
  formApi.updateSchema([
    {
      componentProps: {
        'data-testid': 'tenant-code-input',
        disabled: isEdit.value,
      },
      fieldName: 'code',
    },
  ]);
}

async function handleClosed() {
  recordId.value = 0;
  await formApi.resetForm();
  formApi.updateSchema([
    {
      componentProps: {
        'data-testid': 'tenant-code-input',
        disabled: false,
      },
      fieldName: 'code',
    },
  ]);
}
</script>

<template>
  <Modal :title="title">
    <div data-testid="tenant-form">
      <TenantForm />
    </div>
  </Modal>
</template>
