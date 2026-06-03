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

const defaultApiURL = '/api/v1';
const { apiURL: configuredApiURL } = useAppConfig(
  import.meta.env,
  import.meta.env.PROD,
);
const apiURL = normalizeApiBaseURL(configuredApiURL);
const pluginApiOrigin = resolvePluginApiOrigin(apiURL);

type RuntimeErrorResponse = {
  code?: number;
  error?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
};

type RefreshTokenEnvelope = RuntimeErrorResponse & {
  data?: {
    accessToken?: string;
    refreshToken?: string;
  };
};

type SessionExpiredError = Error & {
  response: {
    data: RuntimeErrorResponse;
    status: number;
  };
};

const sessionExpiredMessageKey = 'ui.fallback.http.unauthorized';

const refreshRequestClient = new RequestClient({ baseURL: apiURL });

function normalizeApiBaseURL(baseURL: unknown) {
  const normalized = typeof baseURL === 'string' ? baseURL.trim() : '';
  return normalized || defaultApiURL;
}

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

function createSessionExpiredError(
  responseData: RuntimeErrorResponse = {},
): SessionExpiredError {
  const data = {
    ...responseData,
    messageKey: responseData.messageKey || sessionExpiredMessageKey,
  };
  const message =
    resolveRuntimeErrorMessage(data) ||
    $t(sessionExpiredMessageKey) ||
    'Session expired';
  const error = new Error(message) as SessionExpiredError;
  error.response = {
    data,
    status: 401,
  };
  return error;
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
    accessStore.setRefreshToken(null);
    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      await authStore.clearSession();
    }
  }

  async function doRefreshToken() {
    const accessStore = useAccessStore();
    const refreshToken = accessStore.refreshToken;
    if (!refreshToken) {
      throw createSessionExpiredError();
    }

    const response =
      await refreshRequestClient.instance.post<RefreshTokenEnvelope>(
        '/auth/refresh',
        { refreshToken },
        {
          headers: {
            'Accept-Language': resolveRequestLocale(),
          },
        },
      );
    const responseData = response.data;
    const nextAccessToken = responseData?.data?.accessToken;
    if (responseData?.code !== 0 || !nextAccessToken) {
      throw createSessionExpiredError(responseData);
    }

    accessStore.setAccessToken(nextAccessToken);
    accessStore.setRefreshToken(
      responseData.data?.refreshToken || refreshToken,
    );
    return nextAccessToken;
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
      enableRefreshToken: preferences.app.enableRefreshToken,
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

function resolvePluginApiOrigin(baseURL: string) {
  const normalized = baseURL.trim();
  if (normalized.startsWith('/')) {
    return typeof window === 'undefined' ? '' : window.location.origin;
  }
  if (!normalized) {
    return '';
  }
  return normalized.replace(/\/api\/v1\/?$/, '').replace(/\/$/, '');
}

export function pluginApiPath(pluginId: string, pathName: string) {
  const normalizedPluginID = pluginId.trim().replace(/^\/+|\/+$/g, '');
  const normalizedPath = pathName.replace(/^\/+/, '');
  return `${pluginApiOrigin}/x/${normalizedPluginID}/api/v1/${normalizedPath}`;
}

export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

export const baseRequestClient = new RequestClient({ baseURL: apiURL });
