<script lang="ts" setup>
import type { LoginTenant } from '#/api/tenant/model';

import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { $t } from '@vben/locales';

import { useAuthStore } from '#/store';

defineOptions({ name: 'OAuthHandoff' });

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const status = ref<'failed' | 'finished' | 'pending'>('pending');
const message = ref($t('authentication.oauthHandoff.pending'));

const oauthErrorMessageKeyByCode: Record<string, string> = {
  AUTH_EXTERNAL_IDENTITY_INVALID: 'externalIdentityInvalid',
  AUTH_EXTERNAL_LOGIN_FAILED: 'externalLoginFailed',
  AUTH_EXTERNAL_USER_NOT_PROVISIONED: 'externalUserNotProvisioned',
  AUTH_IP_BLACKLISTED: 'ipBlacklisted',
  AUTH_TENANT_UNAVAILABLE: 'tenantUnavailable',
  AUTH_USER_DISABLED: 'userDisabled',
  authorize_url_failed: 'authorizeUrlFailed',
  code_exchange_failed: 'codeExchangeFailed',
  email_not_verified: 'emailNotVerified',
  empty_login_result: 'emptyLoginResult',
  handoff_failed: 'handoffFailed',
  invalid_state: 'invalidState',
  missing_code_or_state: 'missingCodeOrState',
  provider_disabled: 'providerDisabled',
  settings_unavailable: 'settingsUnavailable',
  userinfo_failed: 'userinfoFailed',
};

/**
 * Reads the OAuth handoff payload from the current route query.
 *
 * Source-plugin callbacks redirect to /oauth-handoff with the host login
 * outcome encoded as query parameters; vue-router exposes them as
 * route.query whether the workspace runs in hash or history mode.
 */
function readQuery(): Record<string, string> {
  const result: Record<string, string> = {};
  const source = route.query as Record<string, unknown>;
  for (const key of Object.keys(source)) {
    const value = source[key];
    if (typeof value === 'string') {
      result[key] = value;
    } else if (Array.isArray(value) && typeof value[0] === 'string') {
      result[key] = value[0];
    }
  }
  return result;
}

/**
 * Decodes the base64url-encoded tenant list shipped by callbacks for
 * multi-tenant users so the SPA can show the tenant picker without
 * re-querying the backend.
 */
function decodeTenants(value: string | undefined): LoginTenant[] {
  if (!value) {
    return [];
  }
  try {
    const padded = value + '==='.slice((value.length + 3) % 4);
    const normalized = padded.replaceAll('-', '+').replaceAll('_', '/');
    const decoded = atob(normalized);
    const parsed = JSON.parse(decoded) as Array<Record<string, unknown>>;
    return parsed
      .map((item) => ({
        code: typeof item.code === 'string' ? item.code : '',
        id: typeof item.id === 'number' ? item.id : Number(item.id ?? 0),
        name: typeof item.name === 'string' ? item.name : '',
        status: typeof item.status === 'string' ? item.status : '',
      }))
      .filter((tenant): tenant is LoginTenant => tenant.id > 0);
  } catch {
    return [];
  }
}

/**
 * Maps a stable OAuth error code (host bizerr RuntimeCode or callback
 * pipeline string) to a localized friendly message. Unknown codes fall back
 * to a generic shape so operators see the raw code for diagnostics.
 */
function describeOAuthError(code: string): string {
  const messageKey = oauthErrorMessageKeyByCode[code];
  if (!messageKey) {
    return $t('authentication.oauthHandoff.defaultError', [code]);
  }
  const key = `authentication.oauthHandoff.errors.${messageKey}`;
  const translated = $t(key);
  return translated === key
    ? $t('authentication.oauthHandoff.defaultError', [code])
    : translated;
}

onMounted(async () => {
  const payload = readQuery();

  if (payload.error) {
    status.value = 'failed';
    message.value = describeOAuthError(payload.error);
    await router.replace({
      path: LOGIN_PATH,
      query: { oauthError: payload.error },
    });
    return;
  }

  try {
    await authStore.completeOAuthHandoff({
      accessToken: payload.accessToken,
      preToken: payload.preToken,
      redirect: payload.redirect,
      refreshToken: payload.refreshToken,
      tenants: decodeTenants(payload.tenants),
    });
    status.value = 'finished';
    message.value = $t('authentication.oauthHandoff.finished');
  } catch (error) {
    status.value = 'failed';
    message.value =
      error instanceof Error
        ? error.message
        : $t('authentication.oauthHandoff.unknownFailure');
    await router.replace({
      path: LOGIN_PATH,
      query: { oauthError: 'handoff_failed' },
    });
  }
});
</script>

<template>
  <div
    style="
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 100%;
      padding: 24px;
    "
  >
    <div
      style="
        width: min(100%, 420px);
        padding: 32px;
        background: rgba(255, 255, 255, 0.95);
        border: 1px solid #e5e7eb;
        border-radius: 16px;
        text-align: center;
      "
    >
      <div
        v-if="status === 'pending'"
        aria-hidden="true"
        style="
          margin: 0 auto 16px;
          width: 32px;
          height: 32px;
          border-radius: 9999px;
          border: 2px solid #e5e7eb;
          border-top-color: #1677ff;
          animation: oauthHandoffSpin 0.8s linear infinite;
        "
      ></div>
      <h1 style="margin: 0 0 8px; font-size: 18px; font-weight: 600">
        {{ $t('authentication.oauthHandoff.title') }}
      </h1>
      <p style="margin: 0; color: #4b5563; line-height: 1.6">
        {{ message }}
      </p>
    </div>
  </div>
</template>

<style scoped>
@keyframes oauthHandoffSpin {
  to {
    transform: rotate(360deg);
  }
}
</style>
