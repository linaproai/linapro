<script setup lang="ts">
import type { SysUser } from '#/api/system/user';

import { ref } from 'vue';

import { useVbenModal, z } from '@vben/common-ui';

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
        placeholder: '请输入新的密码，密码长度为5-20',
      },
      fieldName: 'password',
      label: '新的密码',
      rules: z
        .string()
        .min(5, { message: '密码长度为5-20' })
        .max(20, { message: '密码长度为5-20' }),
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
    message.success('重置密码成功');
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
    title="重置密码"
  >
    <div class="flex flex-col gap-[12px]">
      <Descriptions v-if="currentUser" size="small" :column="1" bordered>
        <DescriptionsItem label="用户ID">
          {{ currentUser.id }}
        </DescriptionsItem>
        <DescriptionsItem label="用户名">
          {{ currentUser.username }}
        </DescriptionsItem>
        <DescriptionsItem label="昵称">
          {{ currentUser.nickname }}
        </DescriptionsItem>
      </Descriptions>
      <BasicForm />
    </div>
  </BasicModal>
</template>
