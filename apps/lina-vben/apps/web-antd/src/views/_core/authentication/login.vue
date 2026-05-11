<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, reactive } from 'vue';

import {
  AuthenticationLogin,
  useVbenForm,
  VbenButton,
  z,
} from '@vben/common-ui';
import { $t } from '@vben/locales';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { pluginSlotKeys } from '#/plugins/plugin-slots';
import { publicFrontendSettings } from '#/runtime/public-frontend';
import { useAuthStore, useTenantStore } from '#/store';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const tenantStore = useTenantStore();
const tenantOptions = computed(() =>
  tenantStore.tenants.map((tenant) => ({
    code: tenant.code,
    label: `${tenant.name} (${tenant.code})`,
    name: tenant.name,
    value: String(tenant.id),
  })),
);
const loginSubtitle = computed(
  () =>
    publicFrontendSettings.auth.loginSubtitle ||
    $t('authentication.loginSubtitle'),
);

const tenantSubtitle = computed(() =>
  $t('pages.multiTenant.login.selectTenantSubtitle'),
);

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z.string().min(1, { message: $t('authentication.usernameTip') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.passwordTip'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      rules: z.string().min(1, { message: $t('authentication.passwordTip') }),
    },
  ];
});

const tenantFormSchema = computed((): VbenFormSchema[] => [
  {
    component: 'VbenSelect',
    componentProps: {
      class: 'h-11',
      options: tenantOptions.value,
      placeholder: $t('pages.multiTenant.login.selectTenant'),
    },
    fieldName: 'tenantId',
    label: $t('pages.multiTenant.login.selectTenant'),
    rules: 'selectRequired',
  },
]);

const [TenantForm, tenantFormApi] = useVbenForm(
  reactive({
    commonConfig: {
      hideLabel: true,
      hideRequiredMark: true,
    },
    schema: tenantFormSchema,
    showDefaultActions: false,
  }),
);

async function handleSubmit(values: Record<string, any>) {
  const result = await authStore.authLogin(values);
  if (result.requiresTenantSelection && result.tenants?.[0]) {
    await tenantFormApi.setFieldValue(
      'tenantId',
      String(result.tenants[0].id),
    );
  }
}

async function handleSelectTenant() {
  const { valid } = await tenantFormApi.validate();
  if (!valid) {
    return;
  }
  const values = await tenantFormApi.getValues<{ tenantId?: string }>();
  const tenantId = Number(values.tenantId);
  if (!Number.isFinite(tenantId) || tenantId <= 0) {
    return;
  }
  await authStore.selectTenant(tenantId);
}
</script>

<template>
  <div>
    <AuthenticationLogin
      v-if="!authStore.pendingPreToken"
      :form-schema="formSchema"
      :loading="authStore.loginLoading"
      :show-code-login="false"
      :show-forget-password="false"
      :show-qrcode-login="false"
      :show-register="false"
      :show-third-party-login="false"
      :sub-title="loginSubtitle"
      @submit="handleSubmit"
    />
    <div
      v-else
      data-testid="login-tenant-selector"
      @keydown.enter.prevent="handleSelectTenant"
    >
      <div class="mb-7 sm:mx-auto sm:w-full sm:max-w-md">
        <h2
          class="mb-3 text-3xl/9 font-bold tracking-tight text-foreground lg:text-4xl"
        >
          {{ $t('pages.multiTenant.login.selectTenant') }}
        </h2>
        <p class="lg:text-md text-sm text-muted-foreground">
          {{ tenantSubtitle }}
        </p>
      </div>
      <TenantForm class="mb-8" data-testid="login-tenant-form" />
      <VbenButton
        :class="{
          'cursor-wait': authStore.loginLoading,
        }"
        :loading="authStore.loginLoading"
        aria-label="select tenant"
        class="w-full"
        data-testid="login-tenant-confirm"
        @click="handleSelectTenant"
      >
        {{ $t('pages.multiTenant.login.enterTenant') }}
      </VbenButton>
    </div>
    <PluginSlotOutlet :slot-key="pluginSlotKeys.authLoginAfter" class="mt-4" />
  </div>
</template>
