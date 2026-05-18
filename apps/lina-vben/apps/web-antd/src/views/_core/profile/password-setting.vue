<script setup lang="ts">
import type { VbenFormSchema } from '#/adapter/form';

import { z } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { updateProfile } from '#/api/system/user';
import { $t } from '#/locales';

const emit = defineEmits<{ updated: [] }>();

const formSchema: VbenFormSchema[] = [
  {
    fieldName: 'oldPassword',
    label: $t('pages.profile.password.oldPassword'),
    component: 'VbenInputPassword',
    componentProps: {
      placeholder: $t('pages.profile.password.placeholders.oldPassword'),
    },
    rules: z
      .string({
        required_error: $t('pages.profile.password.validation.oldPassword'),
      })
      .min(1, { message: $t('pages.profile.password.validation.oldPassword') }),
  },
  {
    fieldName: 'newPassword',
    label: $t('pages.profile.password.newPassword'),
    component: 'VbenInputPassword',
    componentProps: {
      passwordStrength: true,
      placeholder: $t('pages.profile.password.placeholders.newPassword'),
    },
    rules: z
      .string({
        required_error: $t('pages.profile.password.validation.newPassword'),
      })
      .min(5, { message: $t('pages.profile.password.validation.passwordLength') }),
  },
  {
    fieldName: 'confirmPassword',
    label: $t('pages.profile.password.confirmPassword'),
    component: 'VbenInputPassword',
    componentProps: {
      placeholder: $t('pages.profile.password.placeholders.confirmPassword'),
    },
    dependencies: {
      rules(values) {
        const { newPassword } = values;
        return z
          .string({
            required_error: $t('pages.profile.password.validation.confirmPassword'),
          })
          .min(1, {
            message: $t('pages.profile.password.validation.confirmPassword'),
          })
          .refine((value) => value === newPassword, {
            message: $t('pages.profile.password.validation.passwordMismatch'),
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
    labelWidth: 140,
    componentProps: {
      class: 'w-full',
    },
  },
  wrapperClass: 'grid-cols-1',
  resetButtonOptions: { show: false },
  submitButtonOptions: { content: $t('pages.profile.password.submit') },
  async handleSubmit(values) {
    buttonLoading(true);
    try {
      await updateProfile({ password: values.newPassword });
      message.success($t('pages.profile.password.success'));
      formApi.resetForm();
      emit('updated');
    } finally {
      buttonLoading(false);
    }
  },
});
</script>
<template>
  <div data-testid="profile-password-form" class="mt-[16px] w-full max-w-[30rem]">
    <Form />
  </div>
</template>
