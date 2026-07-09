import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  accessStore,
  appConfig,
  authStore,
  requestClientInstances,
  tenantStore,
  translations,
} = vi.hoisted(() => {
  const accessStore = {
    accessToken: null as null | string,
    isAccessChecked: false,
    refreshToken: null as null | string,
    setAccessToken: vi.fn((token: null | string) => {
      accessStore.accessToken = token;
    }),
    setLoginExpired: vi.fn(),
    setRefreshToken: vi.fn((token: null | string) => {
      accessStore.refreshToken = token;
    }),
  };
  return {
    accessStore,
    appConfig: {
      apiURL: undefined as string | undefined,
    },
    authStore: {
      clearSession: vi.fn(),
    },
    requestClientInstances: [] as Array<{
      addRequestInterceptor: ReturnType<typeof vi.fn>;
      addResponseInterceptor: ReturnType<typeof vi.fn>;
      instance: {
        post: ReturnType<typeof vi.fn>;
      };
      options?: unknown;
    }>,
    tenantStore: {
      currentTenant: null as null | { code?: string },
      enabled: false,
    },
    translations: {
      'ui.fallback.http.unauthorized': '登录认证过期，请重新登录后继续。',
    } as Record<string, string>,
  };
});

vi.mock('@vben/hooks', () => ({
  useAppConfig: () => appConfig,
}));

vi.mock('@vben/preferences', () => ({
  preferences: {
    app: {
      enableRefreshToken: false,
      locale: 'zh-CN',
      loginExpiredMode: 'modal',
    },
  },
}));

vi.mock('@vben/request', () => {
  class MockRequestClient {
    public instance = {
      post: vi.fn(),
    };

    public addRequestInterceptor = vi.fn();

    public addResponseInterceptor = vi.fn();

    public constructor(public readonly options?: unknown) {
      requestClientInstances.push(this);
    }
  }

  return {
    authenticateResponseInterceptor: vi.fn(() => vi.fn()),
    defaultResponseInterceptor: vi.fn(() => vi.fn()),
    errorMessageResponseInterceptor: vi.fn(() => vi.fn()),
    RequestClient: MockRequestClient,
  };
});

vi.mock('@vben/stores', () => ({
  useAccessStore: () => accessStore,
}));

vi.mock('ant-design-vue', () => ({
  message: {
    error: vi.fn(),
  },
}));

vi.mock('#/locales', () => ({
  $t: (key: string) => translations[key] ?? key,
}));

vi.mock('#/store', () => ({
  useAuthStore: () => authStore,
}));

vi.mock('#/store/tenant', () => ({
  useTenantStore: () => tenantStore,
}));

async function importRequestWithApiURL(apiURL?: string) {
  appConfig.apiURL = apiURL;
  vi.resetModules();
  return await import('./request');
}

describe('request API helpers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    requestClientInstances.length = 0;
    accessStore.accessToken = null;
    accessStore.isAccessChecked = false;
    accessStore.refreshToken = null;
    appConfig.apiURL = undefined;
    tenantStore.currentTenant = null;
    tenantStore.enabled = false;
  });

  it('uses the default host API base when the test environment omits VITE_GLOB_API_URL', async () => {
    const { pluginApiPath } = await importRequestWithApiURL();

    expect(pluginApiPath('linapro-tenant-core', '/auth/login-tenants')).toBe(
      `${window.location.origin}/x/linapro-tenant-core/api/v1/auth/login-tenants`,
    );
  });

  it('derives plugin API paths from an absolute host API base', async () => {
    const { pluginApiPath } = await importRequestWithApiURL(
      'https://example.test/api/v1/',
    );

    expect(pluginApiPath('/linapro-tenant-core/', 'platform/tenants')).toBe(
      'https://example.test/x/linapro-tenant-core/api/v1/platform/tenants',
    );
  });

  it('marks refresh authorization failures as a localized session-expired error', async () => {
    accessStore.refreshToken = 'stored-refresh-token';
    await importRequestWithApiURL();

    const refreshClient = requestClientInstances[0];
    refreshClient?.instance.post.mockResolvedValue({
      data: {
        code: 61,
        message: 'Not Authorized',
      },
    });

    const { authenticateResponseInterceptor } = await import('@vben/request');
    const authenticateOptions = vi.mocked(authenticateResponseInterceptor).mock
      .calls[0]?.[0] as
      | {
          doRefreshToken: () => Promise<string>;
        }
      | undefined;

    await expect(authenticateOptions?.doRefreshToken()).rejects.toMatchObject({
      message: '登录认证过期，请重新登录后继续。',
      response: {
        data: {
          code: 61,
          message: 'Not Authorized',
          messageKey: 'ui.fallback.http.unauthorized',
        },
        status: 401,
      },
    });
    expect(refreshClient?.instance.post).toHaveBeenCalledWith(
      '/auth/refresh',
      { refreshToken: 'stored-refresh-token' },
      {
        headers: {
          'Accept-Language': 'zh-CN',
        },
      },
    );
    expect(accessStore.setAccessToken).not.toHaveBeenCalled();
  });
});
