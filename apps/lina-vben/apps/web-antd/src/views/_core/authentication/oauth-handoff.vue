<script lang="ts" setup>
import type { LoginTenant } from '#/api/tenant/model';

import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';

import { useAuthStore } from '#/store';

defineOptions({ name: 'OAuthHandoff' });

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const status = ref<'failed' | 'finished' | 'pending'>('pending');
const message = ref('Finishing authentication, please wait...');

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
  switch (code) {
    case 'AUTH_EXTERNAL_IDENTITY_INVALID': {
      return 'The external provider did not return a valid identity. Please try again.';
    }
    case 'AUTH_EXTERNAL_LOGIN_FAILED': {
      return 'OAuth login failed. Please try again or contact your administrator.';
    }
    case 'AUTH_EXTERNAL_USER_NOT_PROVISIONED': {
      return 'No local account is linked to this external identity. Please contact your administrator to grant access.';
    }
    case 'AUTH_IP_BLACKLISTED': {
      return 'Login from this network is currently blocked.';
    }
    case 'AUTH_TENANT_UNAVAILABLE': {
      return 'No active tenant is available for this account.';
    }
    case 'AUTH_USER_DISABLED': {
      return 'This account has been disabled.';
    }
    case 'code_exchange_failed': {
      return 'Failed to exchange the authorization code. Please try again.';
    }
    case 'email_not_verified': {
      return 'The provider did not return a verified email address.';
    }
    case 'empty_login_result': {
      return 'OAuth login returned an empty result. Please try again.';
    }
    case 'invalid_state': {
      return 'OAuth state validation failed. Please restart the login flow.';
    }
    case 'missing_code_or_state': {
      return 'OAuth callback is missing required parameters.';
    }
    case 'provider_disabled': {
      return 'This authentication provider is currently disabled.';
    }
    case 'userinfo_failed': {
      return 'Failed to fetch the external user profile. Please try again.';
    }
    default: {
      return `OAuth login failed: ${code}`;
    }
  }
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
    message.value = 'OAuth login finished.';
  } catch (error) {
    status.value = 'failed';
    message.value =
      error instanceof Error
        ? error.message
        : 'OAuth handoff failed for an unknown reason.';
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
        OAuth Login
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
