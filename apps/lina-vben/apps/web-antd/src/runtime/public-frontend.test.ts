import { beforeEach, describe, expect, it, vi } from 'vitest';

const { preferencesState } = vi.hoisted(() => ({
  preferencesState: {
    app: {
      locale: 'zh-CN',
    },
  },
}));

const updatePreferences = vi.fn();
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
  },
  updatePreferences,
}));

describe('public frontend runtime settings', () => {
  beforeEach(() => {
    vi.resetModules();
    preferencesState.app.locale = 'zh-CN';
    updatePreferences.mockReset();
    getInitialPreferences.mockClear();
    vi.stubGlobal('fetch', vi.fn());
  });

  it('bypasses browser cache and applies the latest server theme', async () => {
    vi.mocked(fetch).mockResolvedValue({
      json: async () => ({
        data: {
          app: {
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
    expect(settings?.auth.panelLayout).toBe('panel-right');
    expect(settings?.user.defaultAvatar).toBe('/avatar.webp');
    expect(settings?.ui.themeMode).toBe('dark');
    expect(updatePreferences).toHaveBeenCalledWith(
      expect.objectContaining({
        app: expect.objectContaining({
          authPageLayout: 'panel-right',
          defaultAvatar: '/avatar.webp',
          name: 'LinaPro Dark',
        }),
        theme: expect.objectContaining({
          builtinType: 'default',
          colorPrimary: '#1677ff',
          mode: 'dark',
        }),
      }),
    );
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
          defaultAvatar: '/avatar.webp',
        }),
      }),
    );
  });
});
