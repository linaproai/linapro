/**
 * This file can be adjusted according to application request behavior.
 */
import type { RequestClientOptions } from '@vben/request';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  defaultResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';
import { useAccessStore } from '@vben/stores';

import { message } from 'ant-design-vue';

import { $t } from '#/locales';
import { useAuthStore } from '#/store';
import { useTenantStore } from '#/store/tenant';

const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);

type RuntimeErrorResponse = {
  error?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
};

function resolveRequestLocale() {
  if (typeof document === 'undefined') {
    return preferences.app.locale;
  }
  return document.documentElement.lang || preferences.app.locale;
}

function resolveRuntimeErrorMessage(responseData: RuntimeErrorResponse) {
  const messageKey = responseData?.messageKey?.trim();
  if (messageKey) {
    const localized = $t(messageKey, responseData.messageParams || {});
    if (localized && localized !== messageKey) {
      return localized;
    }
  }
  return responseData?.error || responseData?.message || '';
}

function createRequestClient(baseURL: string, options?: RequestClientOptions) {
  const client = new RequestClient({
    ...options,
    baseURL,
  });

  /**
   * Re-authentication flow.
   */
  async function doReAuthenticate() {
    console.warn('Access token is invalid or expired.');
    const accessStore = useAccessStore();
    const authStore = useAuthStore();
    accessStore.setAccessToken(null);
    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      await authStore.clearSession();
    }
  }

  /**
   * Token refresh is not supported; re-authenticate directly.
   */
  async function doRefreshToken() {
    return '';
  }

  function formatToken(token: null | string) {
    return token ? `Bearer ${token}` : null;
  }

  // Request header handling.
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();
      const tenantStore = useTenantStore();

      config.headers.Authorization = formatToken(accessStore.accessToken);
      config.headers['Accept-Language'] = resolveRequestLocale();
      if (tenantStore.enabled && tenantStore.currentTenant?.code) {
        config.headers['X-Tenant-Code'] = tenantStore.currentTenant.code;
      }
      return config;
    },
  });

  // Normalize response data.
  client.addResponseInterceptor(
    defaultResponseInterceptor({
      codeField: 'code',
      dataField: 'data',
      successCode: 0,
    }),
  );

  // Token expiration handling.
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: false,
      formatToken,
    }),
  );

  // Generic error handling.
  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      const responseData = (error?.response?.data ??
        {}) as RuntimeErrorResponse;
      const errorMessage = resolveRuntimeErrorMessage(responseData);
      message.error(errorMessage || msg);
    }),
  );

  return client;
}

export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

export const baseRequestClient = new RequestClient({ baseURL: apiURL });
