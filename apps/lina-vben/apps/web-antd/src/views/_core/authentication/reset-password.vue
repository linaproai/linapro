<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { AuthenticationForgetPassword, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { notification } from 'ant-design-vue';

import { resetPasswordApi } from '#/api/core/auth';
import { publicFrontendSettings } from '#/runtime/public-frontend';

defineOptions({ name: 'ResetPassword' });

const route = useRoute();
const router = useRouter();
const loading = ref(false);

const forgetPasswordEnabled = computed(
  () => publicFrontendSettings.auth.forgetPasswordEnabled !== false,
);

const token = computed(() => {
  const raw = route.query.token;
  return typeof raw === 'string' ? raw.trim() : '';
});

function redirectWhenDisabled() {
  if (!forgetPasswordEnabled.value) {
    void router.replace('/auth/login');
  }
}

onMounted(() => {
  redirectWhenDisabled();
  if (forgetPasswordEnabled.value && !token.value) {
    notification.error({
      description: $t('authentication.resetPasswordTokenMissingDesc'),
      message: $t('authentication.resetPasswordTokenMissingTitle'),
    });
    void router.replace('/auth/forget-password');
  }
});
watch(forgetPasswordEnabled, redirectWhenDisabled);

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInputPassword',
      componentProps: {
        passwordStrength: true,
        placeholder: $t('authentication.password'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      renderComponentContent() {
        return {
          strengthText: () => $t('authentication.passwordStrength'),
        };
      },
      rules: z.string().min(6, { message: $t('authentication.passwordTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.confirmPassword'),
      },
      dependencies: {
        rules(values) {
          const { password } = values;
          return z
            .string({ required_error: $t('authentication.passwordTip') })
            .min(6, { message: $t('authentication.passwordTip') })
            .refine((value) => value === password, {
              message: $t('authentication.confirmPasswordTip'),
            });
        },
        triggerFields: ['password'],
      },
      fieldName: 'confirmPassword',
      label: $t('authentication.confirmPassword'),
    },
  ];
});

async function handleSubmit(value: Recordable<any>) {
  if (!token.value) {
    return;
  }
  loading.value = true;
  try {
    await resetPasswordApi({
      password: String(value.password ?? ''),
      token: token.value,
    });
    notification.success({
      description: $t('authentication.resetPasswordSuccessDesc'),
      duration: 5,
      message: $t('authentication.resetPasswordSuccessTitle'),
    });
    await router.push('/auth/login');
  } catch {
    // requestClient already surfaces the localized API error toast.
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <AuthenticationForgetPassword
    v-if="forgetPasswordEnabled && token"
    :form-schema="formSchema"
    :loading="loading"
    :sub-title="$t('authentication.resetPasswordSubtitle')"
    :submit-button-text="$t('authentication.resetPasswordSubmit')"
    :title="$t('authentication.resetPasswordTitle')"
    data-testid="reset-password-page"
    @submit="handleSubmit"
  />
</template>
