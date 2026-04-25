<script setup lang="ts">
import type { SysUser } from '#/api/system/user';

import { ref } from 'vue';

import { useVbenModal, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { Descriptions, DescriptionsItem, message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { userResetPassword } from '#/api/system/user';

const emit = defineEmits<{ reload: [] }>();

const currentUser = ref<null | SysUser>(null);

const [BasicForm, formApi] = useVbenForm({
  schema: [
    {
      component: 'InputPassword',
      componentProps: {
        placeholder: $t('pages.system.user.resetPassword.placeholder'),
      },
      fieldName: 'password',
      label: $t('pages.system.user.resetPassword.newPassword'),
      rules: z
        .string()
        .min(5, { message: $t('pages.system.user.messages.passwordLength') })
        .max(20, { message: $t('pages.system.user.messages.passwordLength') }),
    },
  ],
  showDefaultActions: false,
  commonConfig: {
    labelWidth: 80,
  },
});

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleSubmit,
  onOpenChange: handleOpenChange,
});

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ record: SysUser }>();
  if (data?.record) {
    currentUser.value = data.record;
  }
  await formApi.resetForm();
}

async function handleSubmit() {
  if (!currentUser.value) {
    return;
  }
  try {
    modalApi.lock(true);
    const { valid } = await formApi.validate();
    if (!valid) {
      return;
    }
    const values = await formApi.getValues();
    await userResetPassword(currentUser.value.id, values.password);
    message.success($t('pages.system.user.messages.resetPasswordSuccess'));
    emit('reload');
    handleClosed();
  } catch (error) {
    console.error(error);
  } finally {
    modalApi.lock(false);
  }
}

async function handleClosed() {
  modalApi.close();
  await formApi.resetForm();
  currentUser.value = null;
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    :title="$t('pages.system.user.resetPassword.title')"
  >
    <div class="flex flex-col gap-[12px]">
      <Descriptions v-if="currentUser" size="small" :column="1" bordered>
        <DescriptionsItem :label="$t('pages.system.user.resetPassword.userId')">
          {{ currentUser.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.fields.username')">
          {{ currentUser.username }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.fields.nickname')">
          {{ currentUser.nickname }}
        </DescriptionsItem>
      </Descriptions>
      <BasicForm />
    </div>
  </BasicModal>
</template>
