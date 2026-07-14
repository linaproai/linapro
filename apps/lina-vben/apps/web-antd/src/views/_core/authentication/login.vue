<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed, onMounted, reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  AuthenticationLogin,
  useVbenForm,
  VbenButton,
  z,
} from '@vben/common-ui';
import { $t } from '@vben/locales';

import { notification } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { pluginSlotKeys } from '#/plugins/plugin-slots';
import { publicFrontendSettings } from '#/runtime/public-frontend';
import { useAuthStore, useTenantStore } from '#/store';

import { resolveExternalLoginErrorMessage as mapExternalLoginError } from './external-login-error';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const tenantStore = useTenantStore();
const route = useRoute();
const router = useRouter();

/**
 * consumeExternalLogin handles the `?externalLogin=1` outcome a protocol
 * plugin callback encodes into the login-page query. Successful outcomes carry
 * a single-use handoff code (never raw JWTs); the SPA exchanges it with the
 * host and then reuses the password-login session bootstrap. Errors use a
 * safe message code. The query is stripped afterwards so refreshes cannot
 * replay the handoff.
 */
function resolveExternalLoginErrorMessage(message: string) {
  return mapExternalLoginError(message, {
    configMissing: $t('authentication.externalLoginConfigMissing'),
    discoveryFailed: $t('authentication.externalLoginDiscoveryFailed'),
    externalLoginFailed: $t('authentication.externalLoginFailed'),
    fallbackLoginFailed: $t('authentication.loginFailed'),
    translate: (key) => $t(key),
  });
}

async function consumeExternalLogin() {
  const query = route.query;
  if (query.externalLogin !== '1') {
    return;
  }
  const status = String(query.status ?? '');
  const message = String(query.message ?? '');
  const handoff = typeof query.handoff === 'string' ? query.handoff : '';
  const redirectPath =
    typeof query.redirect === 'string' ? query.redirect : '';
  await router.replace({ path: route.path, query: {} });
  if (status === 'error') {
    notification.error({
      description: resolveExternalLoginErrorMessage(message),
      duration: 5,
      message: $t('authentication.loginFailed'),
    });
    return;
  }
  if (!handoff) {
    notification.error({
      description: $t('authentication.externalLoginHandoffInvalid'),
      duration: 5,
      message: $t('authentication.loginFailed'),
    });
    return;
  }
  try {
    await authStore.completeExternalLoginFromHandoff(handoff, redirectPath);
  } catch {
    notification.error({
      description: $t('authentication.externalLoginHandoffInvalid'),
      duration: 5,
      message: $t('authentication.loginFailed'),
    });
  }
}

onMounted(consumeExternalLogin);
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

/** Host-managed public switches; default true when public config has not loaded. */
const forgetPasswordEnabled = computed(
  () => publicFrontendSettings.auth.forgetPasswordEnabled !== false,
);
const registerEnabled = computed(
  () => publicFrontendSettings.auth.registerEnabled !== false,
);

function goToRegister() {
  router.push('/auth/register');
}

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
      :show-forget-password="forgetPasswordEnabled"
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
    <!--
      协议 / 目录登录（通用 OIDC、LDAP 等）：全宽单行按钮纵向排列。
      无插件注入时通过 :has(.plugin-slot-outlet) 整块隐藏。
    -->
    <div
      class="login-external-auth w-full sm:mx-auto md:max-w-md"
      data-testid="login-external-auth-region"
    >
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.authLoginAfter"
        class="mt-4 flex w-full flex-col gap-3"
        data-testid="login-external-auth-slot"
      />
    </div>
    <!--
      第三方平台账号（Google / Discord / QQ 等）：Vben 同构分隔线 + 横向图标行。
      无插件注入时整块隐藏。
    -->
    <div
      class="login-social-auth w-full sm:mx-auto md:max-w-md"
      data-testid="login-social-auth-region"
    >
      <div class="mt-4 flex items-center justify-between">
        <span class="w-[35%] border-b border-input dark:border-gray-600"></span>
        <span class="text-center text-xs uppercase text-muted-foreground">
          {{ $t('authentication.thirdPartyLogin') }}
        </span>
        <span class="w-[35%] border-b border-input dark:border-gray-600"></span>
      </div>
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.authLoginSocial"
        class="mt-4 flex flex-wrap justify-center"
        data-testid="login-social-auth-slot"
      />
    </div>
    <!--
      创建账号入口放在「其他登录方式」之后，对齐 Vben 登录页截图顺序。
      是否展示由系统参数 sys.auth.registerEnabled 控制。
    -->
    <div
      v-if="registerEnabled && !authStore.pendingPreToken && !authStore.tenantLoginTransitioning"
      class="mt-3 w-full text-center text-sm sm:mx-auto md:max-w-md"
      data-testid="login-create-account"
    >
      {{ $t('authentication.accountTip') }}
      <span
        class="vben-link text-sm font-normal"
        data-testid="login-create-account-link"
        @click="goToRegister"
      >
        {{ $t('authentication.createAccount') }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.login-external-auth:not(:has(.plugin-slot-outlet)),
.login-social-auth:not(:has(.plugin-slot-outlet)) {
  display: none;
}
</style>
