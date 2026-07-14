<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, h, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationRegister, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { Modal, notification } from 'ant-design-vue';

import { registerApi } from '#/api/core/auth';
import { publicFrontendSettings } from '#/runtime/public-frontend';

defineOptions({ name: 'Register' });

const router = useRouter();
const loading = ref(false);
const privacyOpen = ref(false);
const termsOpen = ref(false);

const registerEnabled = computed(
  () => publicFrontendSettings.auth.registerEnabled !== false,
);

const privacyPolicyContent = computed(
  () =>
    publicFrontendSettings.auth.privacyPolicy ||
    $t('authentication.privacyPolicyDefaultBody'),
);

const termsOfServiceContent = computed(
  () =>
    publicFrontendSettings.auth.termsOfService ||
    $t('authentication.termsOfServiceDefaultBody'),
);

function redirectWhenDisabled() {
  if (!registerEnabled.value) {
    void router.replace('/auth/login');
  }
}

function openPrivacyPolicy(event: Event) {
  event.preventDefault();
  event.stopPropagation();
  privacyOpen.value = true;
}

function openTermsOfService(event: Event) {
  event.preventDefault();
  event.stopPropagation();
  termsOpen.value = true;
}

onMounted(redirectWhenDisabled);
watch(registerEnabled, redirectWhenDisabled);

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(2, { message: $t('authentication.usernameTip') }),
    },
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
    {
      component: 'VbenCheckbox',
      fieldName: 'agreePolicy',
      renderComponentContent: () => ({
        default: () =>
          h('span', { class: 'text-sm' }, [
            $t('authentication.agree'),
            h(
              'a',
              {
                class: 'vben-link ml-1',
                href: 'javascript:void(0)',
                'data-testid': 'register-privacy-link',
                onClick: openPrivacyPolicy,
              },
              $t('authentication.privacyPolicy'),
            ),
            h('span', { class: 'mx-1' }, ' & '),
            h(
              'a',
              {
                class: 'vben-link',
                href: 'javascript:void(0)',
                'data-testid': 'register-terms-link',
                onClick: openTermsOfService,
              },
              $t('authentication.terms'),
            ),
          ]),
      }),
      rules: z.boolean().refine((value) => !!value, {
        message: $t('authentication.agreeTip'),
      }),
    },
  ];
});

async function handleSubmit(value: Recordable<any>) {
  loading.value = true;
  try {
    await registerApi({
      email: String(value.email ?? ''),
      password: String(value.password ?? ''),
      username: String(value.username ?? ''),
    });
    notification.success({
      description: $t('authentication.registerSuccessDesc'),
      duration: 5,
      message: $t('authentication.registerSuccessTitle'),
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
  <div>
    <AuthenticationRegister
      v-if="registerEnabled"
      :form-schema="formSchema"
      :loading="loading"
      :sub-title="$t('authentication.signUpSubtitle')"
      data-testid="register-page"
      @submit="handleSubmit"
    />

    <Modal
      v-model:open="privacyOpen"
      :footer="null"
      :title="$t('authentication.privacyPolicy')"
      destroy-on-close
      width="640px"
    >
      <div
        class="max-h-[60vh] overflow-y-auto whitespace-pre-wrap text-sm leading-6 text-foreground"
        data-testid="register-privacy-modal-body"
      >
        {{ privacyPolicyContent }}
      </div>
    </Modal>

    <Modal
      v-model:open="termsOpen"
      :footer="null"
      :title="$t('authentication.terms')"
      destroy-on-close
      width="640px"
    >
      <div
        class="max-h-[60vh] overflow-y-auto whitespace-pre-wrap text-sm leading-6 text-foreground"
        data-testid="register-terms-modal-body"
      >
        {{ termsOfServiceContent }}
      </div>
    </Modal>
  </div>
</template>
