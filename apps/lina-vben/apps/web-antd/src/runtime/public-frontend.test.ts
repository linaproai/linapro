import { beforeEach, describe, expect, it, vi } from 'vitest';

const { preferencesState } = vi.hoisted(() => ({
  preferencesState: {
    app: {
      locale: 'zh-CN',
    },
  },
}));

const updatePreferences = vi.fn();
const hasUserThemePreference = vi.fn(() => false);
const getInitialPreferences = vi.fn(() => ({
  app: {
    authPageLayout: 'panel-right',
    defaultAvatar: '/avatar.webp',
    layout: 'sidebar-nav',
    name: 'LinaPro',
    watermarkContent: '',
  },
  logo: {
    source: '/logo.svg',
    sourceDark: '/logo-dark.svg',
  },
  theme: {
    builtinType: 'default',
    colorPrimary: '#1677ff',
    mode: 'light',
  },
}));

vi.mock('@vben/hooks', () => ({
  useAppConfig: () => ({
    apiURL: '/api/v1',
  }),
}));

vi.mock('@vben/preferences', () => ({
  preferences: preferencesState,
  preferencesManager: {
    getInitialPreferences,
    hasUserThemePreference,
  },
  updatePreferences,
}));

describe('public frontend runtime settings', () => {
  beforeEach(() => {
    vi.resetModules();
    preferencesState.app.locale = 'zh-CN';
    updatePreferences.mockReset();
    getInitialPreferences.mockClear();
    hasUserThemePreference.mockReset();
    hasUserThemePreference.mockReturnValue(false);
    vi.stubGlobal('fetch', vi.fn());
  });

  it('bypasses browser cache and applies the latest server theme', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {
            logo: '/logo.webp',
            logoDark: '/logo-dark.webp',
            name: 'LinaPro Dark',
          },
          auth: {
            panelLayout: 'panel-right',
          },
          cron: {
            logRetention: {
              mode: 'count',
              value: 120,
            },
            shell: {
              disabledReason: '',
              enabled: true,
              supported: true,
            },
            timezone: {
              current: 'UTC',
            },
          },
          user: {
            defaultAvatar: '/avatar.webp',
          },
          ui: {
            themeMode: 'dark',
          },
          workspace: {
            basePath: '/admin',
          },
        },
      }),
      ok: true,
    } as Response);

    const { publicFrontendSettings, syncPublicFrontendSettings } =
      await import('./public-frontend');
    const settings = await syncPublicFrontendSettings();

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/config/public/frontend',
      expect.objectContaining({
        cache: 'no-store',
        credentials: 'same-origin',
        headers: {
          'Accept-Language': 'zh-CN',
        },
        method: 'GET',
      }),
    );
    expect(publicFrontendSettings.cron.logRetention.mode).toBe('count');
    expect(publicFrontendSettings.cron.logRetention.value).toBe(120);
    expect(publicFrontendSettings.cron.shell.enabled).toBe(true);
    expect(publicFrontendSettings.cron.timezone.current).toBe('UTC');
    expect(publicFrontendSettings.auth.panelLayout).toBe('panel-right');
    expect(publicFrontendSettings.user.defaultAvatar).toBe('/avatar.webp');
    expect(publicFrontendSettings.ui.themeMode).toBe('dark');
    expect(publicFrontendSettings.workspace.basePath).toBe('/admin');
    expect(settings?.auth.panelLayout).toBe('panel-right');
    expect(settings?.user.defaultAvatar).toBe('/avatar.webp');
    expect(settings?.ui.themeMode).toBe('dark');
    expect(settings?.workspace.basePath).toBe('/admin');
    expect(updatePreferences).toHaveBeenCalledWith(
      expect.objectContaining({
        app: expect.objectContaining({
          authPageLayout: 'panel-right',
          defaultAvatar: '/admin/avatar.webp',
          name: 'LinaPro Dark',
        }),
        logo: expect.objectContaining({
          source: '/admin/logo.webp',
          sourceDark: '/admin/logo-dark.webp',
        }),
        theme: expect.objectContaining({
          builtinType: 'default',
          colorPrimary: '#1677ff',
          mode: 'dark',
        }),
      }),
      { markUserThemePreference: false },
    );
  });

  it('keeps an explicit user theme preference over the server default', async () => {
    hasUserThemePreference.mockReturnValue(true);
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {
            name: 'LinaPro Light',
          },
          auth: {
            panelLayout: 'panel-right',
          },
          cron: {},
          ui: {
            themeMode: 'light',
          },
        },
      }),
      ok: true,
    } as Response);

    const { publicFrontendSettings, syncPublicFrontendSettings } =
      await import('./public-frontend');
    const settings = await syncPublicFrontendSettings();
    const [preferenceUpdate, options] = updatePreferences.mock.calls[0] ?? [];

    expect(settings?.ui.themeMode).toBe('light');
    expect(publicFrontendSettings.ui.themeMode).toBe('light');
    expect(preferenceUpdate.theme).toEqual(
      expect.objectContaining({
        builtinType: 'default',
        colorPrimary: '#1677ff',
      }),
    );
    expect(preferenceUpdate.theme).not.toHaveProperty('mode');
    expect(options).toEqual({ markUserThemePreference: false });
  });

  it('falls back to panel-right when the server omits auth panel layout', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {},
          auth: {},
          cron: {},
          ui: {},
        },
      }),
      ok: true,
    } as Response);

    const { publicFrontendSettings, syncPublicFrontendSettings } =
      await import('./public-frontend');
    const settings = await syncPublicFrontendSettings();

    expect(publicFrontendSettings.auth.panelLayout).toBe('panel-right');
    expect(publicFrontendSettings.user.defaultAvatar).toBe('');
    expect(settings?.auth.panelLayout).toBe('panel-right');
    expect(settings?.user.defaultAvatar).toBe('');
    expect(updatePreferences).toHaveBeenCalledWith(
      expect.objectContaining({
        app: expect.objectContaining({
          authPageLayout: 'panel-right',
          defaultAvatar: '/admin/avatar.webp',
        }),
      }),
      { markUserThemePreference: false },
    );
  });

  it('normalizes the startup workspace base path exposed to the router', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {},
          auth: {},
          cron: {},
          ui: {},
          workspace: {
            basePath: '///console///',
          },
        },
      }),
      ok: true,
    } as Response);

    const {
      normalizeWorkspaceBasePath,
      publicFrontendSettings,
      resolveWorkspaceRouterBase,
      syncPublicFrontendSettings,
    } = await import('./public-frontend');
    const settings = await syncPublicFrontendSettings();

    expect(settings?.workspace.basePath).toBe('/console');
    expect(publicFrontendSettings.workspace.basePath).toBe('/console');
    expect(resolveWorkspaceRouterBase()).toBe('/console/');
    expect(normalizeWorkspaceBasePath('/')).toBe('/');
    expect(normalizeWorkspaceBasePath('/x')).toBe('/admin');
    expect(normalizeWorkspaceBasePath('/x-assets/plugin')).toBe('/admin');
    expect(normalizeWorkspaceBasePath('/plugin-assets')).toBe('/plugin-assets');
  });

  it('allows a root workspace base path for a dedicated admin domain', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {},
          auth: {},
          cron: {},
          ui: {},
          workspace: {
            basePath: '/',
          },
        },
      }),
      ok: true,
    } as Response);

    const {
      publicFrontendSettings,
      resolveWorkspaceRouterBase,
      syncPublicFrontendSettings,
    } = await import('./public-frontend');
    const settings = await syncPublicFrontendSettings();

    expect(settings?.workspace.basePath).toBe('/');
    expect(publicFrontendSettings.workspace.basePath).toBe('/');
    expect(resolveWorkspaceRouterBase()).toBe('/');
  });

  it('resolves workspace static assets under the configured workspace base path', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {},
          auth: {},
          cron: {},
          ui: {},
          workspace: {
            basePath: '/admin',
          },
        },
      }),
      ok: true,
    } as Response);

    const { resolveWorkspaceAssetURL, syncPublicFrontendSettings } =
      await import('./public-frontend');
    await syncPublicFrontendSettings();

    expect(resolveWorkspaceAssetURL('/logo.webp')).toBe('/admin/logo.webp');
    expect(resolveWorkspaceAssetURL('stoplight/apidocs.html?lang=zh-CN')).toBe(
      '/admin/stoplight/apidocs.html?lang=zh-CN',
    );
    expect(resolveWorkspaceAssetURL('/admin/logo.webp')).toBe(
      '/admin/logo.webp',
    );
    expect(resolveWorkspaceAssetURL('/api.json?lang=zh-CN')).toBe(
      '/api.json?lang=zh-CN',
    );
    expect(resolveWorkspaceAssetURL('/api/v1/user')).toBe('/api/v1/user');
    expect(resolveWorkspaceAssetURL('/x-assets/plugin/v0.1.0/app.js')).toBe(
      '/x-assets/plugin/v0.1.0/app.js',
    );
    expect(resolveWorkspaceAssetURL('/plugin-assets/logo.webp')).toBe(
      '/admin/plugin-assets/logo.webp',
    );
    expect(resolveWorkspaceAssetURL('https://cdn.example.com/logo.webp')).toBe(
      'https://cdn.example.com/logo.webp',
    );
  });
});
