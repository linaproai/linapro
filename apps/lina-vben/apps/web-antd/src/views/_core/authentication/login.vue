<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import type { AuthApi } from '#/api/core/auth';

import { computed, onMounted, reactive, ref } from 'vue';

import {
  AuthenticationLogin,
  useVbenForm,
  VbenButton,
  z,
} from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { $t } from '@vben/locales';

import { listAuthProvidersApi } from '#/api/core/auth';
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
const tenantTransitionTitle = computed(() =>
  $t('pages.multiTenant.login.enteringTenant'),
);
const tenantTransitionSubtitle = computed(() =>
  $t('pages.multiTenant.login.enteringTenantSubtitle'),
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

// Third-party login provider entries discovered via /auth/providers. The
// host filters out providers whose owning plugin is disabled, so this list
// reflects exactly the entries the workbench should render.
const authProviders = ref<AuthApi.ProviderEntity[]>([]);

/**
 * Composes the URL the browser visits when an operator clicks the
 * provider's login button. When backend redirect routing is enabled the
 * provider id is appended as a state parameter so the OAuth callback can
 * resolve the matching post-login redirect rule.
 */
function buildLoginEntryURL(provider: AuthApi.ProviderEntity) {
  if (!provider.backendRedirectEnabled) {
    return provider.entryUrl;
  }
  return `${provider.entryUrl}?state=${encodeURIComponent(provider.providerId)}`;
}

function handleProviderClick(provider: AuthApi.ProviderEntity) {
  window.location.href = buildLoginEntryURL(provider);
}

onMounted(() => {
  void (async () => {
    try {
      const res = await listAuthProvidersApi();
      authProviders.value = Array.isArray(res.providers)
        ? res.providers
            .slice()
            .sort((a, b) => (a.displayOrder ?? 0) - (b.displayOrder ?? 0))
        : [];
    } catch {
      authProviders.value = [];
    }
  })();
});
</script>

<template>
  <div>
    <div
      v-if="authStore.tenantLoginTransitioning"
      aria-live="polite"
      class="flex min-h-72 flex-col items-center justify-center text-center"
      data-testid="login-tenant-transition"
      role="status"
    >
      <div
        aria-hidden="true"
        class="border-primary/20 border-t-primary mb-8 size-10 animate-spin rounded-full border-2"
      ></div>
      <h2
        class="mb-3 text-2xl/8 font-bold tracking-tight text-foreground lg:text-3xl"
      >
        {{ tenantTransitionTitle }}
      </h2>
      <p class="max-w-sm text-sm text-muted-foreground lg:text-base">
        {{ tenantTransitionSubtitle }}
      </p>
    </div>
    <AuthenticationLogin
      v-else-if="!authStore.pendingPreToken"
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
    <div
      v-if="authProviders.length > 0 && !authStore.pendingPreToken && !authStore.tenantLoginTransitioning"
      class="mt-4 space-y-3"
    >
      <a-button
        v-for="provider in authProviders"
        :key="provider.providerId"
        block
        size="large"
        @click="handleProviderClick(provider)"
      >
        <template #icon>
          <IconifyIcon :icon="provider.icon || 'ant-design:login-outlined'" />
        </template>
        Continue with {{ provider.name }}
      </a-button>
    </div>
    <PluginSlotOutlet :slot-key="pluginSlotKeys.authLoginAfter" class="mt-4" />
  </div>
</template>
