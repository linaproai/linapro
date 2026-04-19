import { beforeEach, describe, expect, it, vi } from 'vitest';

const updatePreferences = vi.fn();
const getInitialPreferences = vi.fn(() => ({
  app: {
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
  preferencesManager: {
    getInitialPreferences,
  },
  updatePreferences,
}));

describe('public frontend runtime settings', () => {
  beforeEach(() => {
    vi.resetModules();
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
          auth: {},
          cron: {
            shell: {
              disabledReason: '',
              enabled: true,
              supported: true,
            },
          },
          ui: {
            themeMode: 'dark',
          },
        },
      }),
      ok: true,
    } as Response);

    const { publicFrontendSettings, syncPublicFrontendSettings } = await import(
      './public-frontend'
    );
    const settings = await syncPublicFrontendSettings();

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/config/public/frontend',
      expect.objectContaining({
        cache: 'no-store',
        credentials: 'same-origin',
        method: 'GET',
      }),
    );
    expect(publicFrontendSettings.cron.shell.enabled).toBe(true);
    expect(publicFrontendSettings.ui.themeMode).toBe('dark');
    expect(settings?.ui.themeMode).toBe('dark');
    expect(updatePreferences).toHaveBeenCalledWith(
      expect.objectContaining({
        app: expect.objectContaining({
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
});
