<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, ref } from 'vue';

import { AuthenticationLogin, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { Button, Card, Radio, RadioGroup } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { pluginSlotKeys } from '#/plugins/plugin-slots';
import { publicFrontendSettings } from '#/runtime/public-frontend';
import { useAuthStore, useTenantStore } from '#/store';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const tenantStore = useTenantStore();
const selectedTenantId = ref<number>();
const loginSubtitle = computed(
  () =>
    publicFrontendSettings.auth.loginSubtitle ||
    $t('authentication.loginSubtitle'),
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

async function handleSubmit(values: Record<string, any>) {
  const result = await authStore.authLogin(values);
  if (result.requiresTenantSelection && result.tenants?.[0]) {
    selectedTenantId.value = result.tenants[0].id;
  }
}

async function handleSelectTenant() {
  if (!selectedTenantId.value) {
    return;
  }
  await authStore.selectTenant(selectedTenantId.value);
}
</script>

<template>
  <div>
    <AuthenticationLogin
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
    <Card
      v-if="authStore.pendingPreToken"
      class="mt-4"
      data-testid="login-tenant-selector"
      :title="$t('pages.multiTenant.login.selectTenant')"
      :bordered="false"
    >
      <RadioGroup v-model:value="selectedTenantId" class="w-full">
        <div class="grid gap-2">
          <Radio
            v-for="tenant in tenantStore.tenants"
            :key="tenant.id"
            :value="tenant.id"
            class="m-0 rounded border p-3"
          >
            <span class="font-medium">{{ tenant.name }}</span>
            <span class="ml-2 text-xs text-[var(--ant-color-text-secondary)]">
              {{ tenant.code }}
            </span>
          </Radio>
        </div>
      </RadioGroup>
      <Button
        class="mt-4 w-full"
        type="primary"
        data-testid="login-tenant-confirm"
        @click="handleSelectTenant"
      >
        {{ $t('pages.multiTenant.login.enterTenant') }}
      </Button>
    </Card>
    <PluginSlotOutlet :slot-key="pluginSlotKeys.authLoginAfter" class="mt-4" />
  </div>
</template>
