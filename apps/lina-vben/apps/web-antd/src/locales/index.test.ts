import { describe, expect, it, vi } from 'vitest';

const { mergeLocaleMessage, preferencesState, setRuntimeLocaleOptions } =
  vi.hoisted(() => ({
    mergeLocaleMessage: vi.fn(),
    preferencesState: {
      app: {
        locale: 'en-US',
      },
    },
    setRuntimeLocaleOptions: vi.fn(),
  }));

vi.mock('@vben/locales', () => ({
  $t: (key: string) => key,
  direction: { value: 'ltr' },
  i18n: {
    global: {
      mergeLocaleMessage,
    },
  },
  loadLocaleMessages: vi.fn(),
  setRuntimeLocaleOptions,
  setupI18n: vi.fn(),
}));

vi.mock('@vben/preferences', () => ({
  preferences: preferencesState,
  updatePreferences: vi.fn((updates) => {
    Object.assign(preferencesState.app, updates.app || {});
  }),
}));

vi.mock('#/runtime/runtime-i18n', () => ({
  loadRuntimeLocaleOptions: vi.fn(),
  loadRuntimeLocaleMessages: vi.fn(),
  mergeMessages: (
    target: Record<string, any>,
    source: Record<string, any>,
  ) => ({
    ...target,
    ...source,
  }),
}));

vi.mock('#/runtime/public-frontend', () => ({
  syncPublicFrontendSettings: vi.fn(),
}));

vi.mock('virtual:lina-app-third-party-locales', () => ({
  antdLocaleLoaders: {},
  dayjsLocaleLoaders: {},
}));

import { createLocaleMessagesLoader } from './index';

function makeRuntimeLocalesResult(options: any[] = []) {
  return {
    defaultLocale: 'zh-CN',
    enabled: true,
    locale: 'en-US',
    options,
  };
}

describe('web locale message loader', () => {
  it('uses runtime fallback semantics without blocking app locale loading', async () => {
    const notifyRuntimeFallback = vi.fn();
    const loader = createLocaleMessagesLoader({
      loadRuntimeLocales: vi.fn().mockResolvedValue(makeRuntimeLocalesResult()),
      loadRuntimeMessages: vi.fn().mockRejectedValue(new Error('unavailable')),
      loadThirdPartyMessages: vi.fn().mockResolvedValue(undefined),
      notifyRuntimeFallback,
      syncPublicSettings: vi.fn().mockResolvedValue(null),
    });

    const messages = await loader('en-US');

    expect(messages).toEqual(expect.any(Object));
    expect(notifyRuntimeFallback).toHaveBeenCalledTimes(1);
  });

  it('does not wait for public frontend settings sync', async () => {
    const loader = createLocaleMessagesLoader({
      loadRuntimeLocales: vi.fn().mockResolvedValue(makeRuntimeLocalesResult()),
      loadRuntimeMessages: vi.fn().mockResolvedValue({
        runtime: {
          title: 'Runtime',
        },
      }),
      loadThirdPartyMessages: vi.fn().mockResolvedValue(undefined),
      notifyRuntimeFallback: vi.fn(),
      syncPublicSettings: vi.fn().mockReturnValue(new Promise(() => {})),
    });

    await expect(loader('en-US')).resolves.toEqual(
      expect.objectContaining({
        runtime: {
          title: 'Runtime',
        },
      }),
    );
  });

  it('applies runtime language switch visibility metadata', async () => {
    setRuntimeLocaleOptions.mockClear();
    const options = [
      {
        isDefault: true,
        label: '简体中文',
        nativeName: '简体中文',
        value: 'zh-CN',
      },
    ];
    const loader = createLocaleMessagesLoader({
      loadRuntimeLocales: vi.fn().mockResolvedValue({
        defaultLocale: 'zh-CN',
        enabled: false,
        locale: 'zh-CN',
        options,
      }),
      loadRuntimeMessages: vi.fn().mockResolvedValue({}),
      loadThirdPartyMessages: vi.fn().mockResolvedValue(undefined),
      notifyRuntimeFallback: vi.fn(),
      syncPublicSettings: vi.fn().mockResolvedValue(null),
    });

    await loader('zh-CN');

    expect(setRuntimeLocaleOptions).toHaveBeenCalledWith(options, {
      enabled: false,
    });
  });

  it('waits for third-party locale packages before returning messages', async () => {
    let resolveThirdParty!: () => void;
    let resolved = false;
    const loader = createLocaleMessagesLoader({
      loadRuntimeLocales: vi.fn().mockResolvedValue(makeRuntimeLocalesResult()),
      loadRuntimeMessages: vi.fn().mockResolvedValue({}),
      loadThirdPartyMessages: vi.fn(
        () =>
          new Promise<void>((resolve) => {
            resolveThirdParty = resolve;
          }),
      ),
      notifyRuntimeFallback: vi.fn(),
      syncPublicSettings: vi.fn().mockResolvedValue(null),
    });

    const loading = loader('en-US').then(() => {
      resolved = true;
    });
    await Promise.resolve();

    expect(resolved).toBe(false);
    resolveThirdParty();
    await loading;
    expect(resolved).toBe(true);
  });

  it('rejects when third-party locale packages fail to load', async () => {
    const loader = createLocaleMessagesLoader({
      loadRuntimeLocales: vi.fn().mockResolvedValue(makeRuntimeLocalesResult()),
      loadRuntimeMessages: vi.fn().mockResolvedValue({}),
      loadThirdPartyMessages: vi
        .fn()
        .mockRejectedValue(new Error('third-party failed')),
      notifyRuntimeFallback: vi.fn(),
      syncPublicSettings: vi.fn().mockResolvedValue(null),
    });

    await expect(loader('en-US')).rejects.toThrow('third-party failed');
  });
});
