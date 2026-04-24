import { beforeEach, describe, expect, it, vi } from 'vitest';

const { preferencesState } = vi.hoisted(() => ({
  preferencesState: {
    app: {
      locale: 'zh-CN',
    },
  },
}));

vi.mock('@vben/hooks', () => ({
  useAppConfig: () => ({
    apiURL: 'http://127.0.0.1:8080/api/v1',
  }),
}));

vi.mock('@vben/preferences', () => ({
  preferences: preferencesState,
}));

import {
  clearRuntimeLocaleMessagesCache,
  getRuntimeLocaleMessagesSnapshot,
  loadRuntimeLocaleMessages,
  lookupRuntimeMessageString,
  mergeMessages,
  reloadRuntimeLocaleMessages,
  runtimeI18nVersion,
} from '../runtime-i18n';

describe('runtime-i18n', () => {
  beforeEach(() => {
    clearRuntimeLocaleMessagesCache();
    preferencesState.app.locale = 'zh-CN';
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('merges nested runtime messages without dropping existing branches', () => {
    expect(
      mergeMessages(
        {
          menu: {
            dashboard: {
              title: '工作台',
            },
          },
        },
        {
          menu: {
            extension: {
              title: '扩展中心',
            },
          },
        },
      ),
    ).toEqual({
      menu: {
        dashboard: {
          title: '工作台',
        },
        extension: {
          title: '扩展中心',
        },
      },
    });
  });

  it('loads and caches runtime locale messages for the active locale', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        data: {
          messages: {
            plugin: {
              demo: {
                title: 'Runtime Demo',
              },
            },
          },
        },
      }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const firstMessages = await loadRuntimeLocaleMessages('en-US');
    const secondMessages = await loadRuntimeLocaleMessages('en-US');

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(firstMessages).toEqual(secondMessages);
    expect(getRuntimeLocaleMessagesSnapshot()).toEqual(firstMessages);
    expect(lookupRuntimeMessageString(firstMessages, 'plugin.demo.title')).toBe(
      'Runtime Demo',
    );
    expect(runtimeI18nVersion.value).toBeGreaterThan(0);
  });

  it('forces a reload when runtime plugin messages change', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          data: {
            messages: {
              plugin: {
                demo: {
                  title: 'Version One',
                },
              },
            },
          },
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          data: {
            messages: {
              plugin: {
                demo: {
                  title: 'Version Two',
                },
              },
            },
          },
        }),
      });
    vi.stubGlobal('fetch', fetchMock);

    await loadRuntimeLocaleMessages('en-US');
    const versionAfterFirstLoad = runtimeI18nVersion.value;
    const reloadedMessages = await reloadRuntimeLocaleMessages('en-US');

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(versionAfterFirstLoad).toBeLessThan(runtimeI18nVersion.value);
    expect(
      lookupRuntimeMessageString(reloadedMessages, 'plugin.demo.title'),
    ).toBe('Version Two');
  });
});
