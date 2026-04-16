<script setup lang="ts">
import type { VbenFormSchema } from '#/adapter/form';

import { z } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { updateProfile } from '#/api/system/user';

const emit = defineEmits<{ updated: [] }>();

const formSchema: VbenFormSchema[] = [
  {
    fieldName: 'oldPassword',
    label: '旧密码',
    component: 'VbenInputPassword',
    componentProps: {
      placeholder: '请输入旧密码',
    },
    rules: z
      .string({ required_error: '请输入旧密码' })
      .min(1, { message: '请输入旧密码' }),
  },
  {
    fieldName: 'newPassword',
    label: '新密码',
    component: 'VbenInputPassword',
    componentProps: {
      passwordStrength: true,
      placeholder: '请输入新密码',
    },
    rules: z
      .string({ required_error: '请输入新密码' })
      .min(5, { message: '密码长度至少5个字符' }),
  },
  {
    fieldName: 'confirmPassword',
    label: '确认密码',
    component: 'VbenInputPassword',
    componentProps: {
      placeholder: '请再次输入新密码',
    },
    dependencies: {
      rules(values) {
        const { newPassword } = values;
        return z
          .string({ required_error: '请再次输入新密码' })
          .min(1, { message: '请再次输入新密码' })
          .refine((value) => value === newPassword, {
            message: '两次输入的密码不一致',
          });
      },
      triggerFields: ['newPassword'],
    },
  },
];

function buttonLoading(loading: boolean) {
  formApi.setState({ submitButtonOptions: { loading } });
}

const [Form, formApi] = useVbenForm({
  schema: formSchema,
  commonConfig: {
    labelWidth: 90,
    componentProps: {
      class: 'w-full',
    },
  },
  wrapperClass: 'grid-cols-1',
  resetButtonOptions: { show: false },
  submitButtonOptions: { content: '修改密码' },
  async handleSubmit(values) {
    buttonLoading(true);
    try {
      await updateProfile({ password: values.newPassword });
      message.success('密码修改成功');
      formApi.resetForm();
      emit('updated');
    } finally {
      buttonLoading(false);
    }
  },
});
</script>
<template>
  <div class="mt-[16px] md:w-full lg:w-1/2 2xl:w-2/5">
    <Form />
  </div>
</template>
