<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationForgetPassword, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { notification } from 'ant-design-vue';

import { forgetPasswordApi } from '#/api/core/auth';
import { publicFrontendSettings } from '#/runtime/public-frontend';

defineOptions({ name: 'ForgetPassword' });

const router = useRouter();
const loading = ref(false);

const forgetPasswordEnabled = computed(
  () => publicFrontendSettings.auth.forgetPasswordEnabled !== false,
);

function redirectWhenDisabled() {
  if (!forgetPasswordEnabled.value) {
    void router.replace('/auth/login');
  }
}

onMounted(redirectWhenDisabled);
watch(forgetPasswordEnabled, redirectWhenDisabled);

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: 'example@example.com',
      },
      fieldName: 'email',
      label: $t('authentication.email'),
      rules: z
        .string()
        .min(1, { message: $t('authentication.emailTip') })
        .email($t('authentication.emailValidErrorTip')),
    },
  ];
});

async function handleSubmit(value: Recordable<any>) {
  loading.value = true;
  try {
    await forgetPasswordApi({ email: String(value.email ?? '') });
    notification.success({
      description: $t('authentication.forgetPasswordSentDesc'),
      duration: 6,
      message: $t('authentication.forgetPasswordSentTitle'),
    });
  } catch {
    // requestClient already surfaces the localized API error toast.
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <AuthenticationForgetPassword
    v-if="forgetPasswordEnabled"
    :form-schema="formSchema"
    :loading="loading"
    :sub-title="$t('authentication.forgetPasswordSubtitle')"
    data-testid="forget-password-page"
    @submit="handleSubmit"
  />
</template>
