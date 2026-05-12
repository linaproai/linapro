import { beforeEach, describe, expect, it, vi } from 'vitest';

const { preferencesState, requestClientGet } = vi.hoisted(() => ({
  preferencesState: {
    app: {
      locale: 'zh-CN',
    },
  },
  requestClientGet: vi.fn(),
}));

vi.mock('@vben/preferences', () => ({
  preferences: preferencesState,
}));

vi.mock('#/api/request', () => ({
  requestClient: {
    get: requestClientGet,
  },
}));

import {
  clearRuntimeLocaleMessagesCache,
  getRuntimeLocaleMessagesSnapshot,
  loadRuntimeLocaleMessages,
  lookupRuntimeMessageString,
  mergeMessages,
  normalizeRuntimeLocalesPayload,
  reloadRuntimeLocaleMessages,
  runtimeI18nVersion,
  runtimePersistentCacheTTL,
} from '../runtime-i18n';

function makeRuntimeResponse(
  messages: Record<string, any>,
  etag = '"en-US-1"',
) {
  return {
    data: {
      data: {
        messages,
      },
    },
    headers: {
      etag,
    },
    status: 200,
  };
}

function seedPersistentRuntimeMessages(
  locale: string,
  entry: {
    etag?: string;
    messages: Record<string, any>;
    savedAt?: number;
  },
) {
  window.localStorage.setItem(
    `linapro:i18n:runtime:${locale}`,
    JSON.stringify({
      etag: entry.etag ?? '"persisted"',
      messages: entry.messages,
      savedAt: entry.savedAt ?? Date.now(),
    }),
  );
}

describe('runtime-i18n', () => {
  beforeEach(() => {
    clearRuntimeLocaleMessagesCache();
    preferencesState.app.locale = 'zh-CN';
    requestClientGet.mockReset();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    window.localStorage.clear();
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

  it('normalizes runtime locale descriptors without exposing direction metadata', () => {
    expect(
      normalizeRuntimeLocalesPayload({
        enabled: false,
        items: [
          {
            isDefault: true,
            locale: 'zh-CN',
            name: '简体中文',
            nativeName: '简体中文',
          },
        ],
        locale: 'zh-CN',
      }),
    ).toEqual({
      defaultLocale: 'zh-CN',
      enabled: false,
      locale: 'zh-CN',
      options: [
        {
          isDefault: true,
          label: '简体中文',
          nativeName: '简体中文',
          value: 'zh-CN',
        },
      ],
    });
  });

  it('loads and caches runtime locale messages for the active locale', async () => {
    requestClientGet.mockResolvedValue(
      makeRuntimeResponse({
        plugin: {
          demo: {
            title: 'Runtime Demo',
          },
        },
      }),
    );

    const firstMessages = await loadRuntimeLocaleMessages('en-US');
    const secondMessages = await loadRuntimeLocaleMessages('en-US');

    expect(requestClientGet).toHaveBeenCalledTimes(1);
    expect(requestClientGet).toHaveBeenCalledWith(
      '/i18n/runtime/messages',
      expect.objectContaining({
        headers: expect.objectContaining({
          'Accept-Language': 'en-US',
        }),
        params: {
          lang: 'en-US',
        },
        responseReturn: 'raw',
      }),
    );
    expect(firstMessages).toEqual(secondMessages);
    expect(getRuntimeLocaleMessagesSnapshot()).toEqual(firstMessages);
    expect(lookupRuntimeMessageString(firstMessages, 'plugin.demo.title')).toBe(
      'Runtime Demo',
    );
    expect(runtimeI18nVersion.value).toBeGreaterThan(0);
  });

  it('renders immediately from fresh persistent cache and revalidates in the background', async () => {
    const persistedMessages = {
      plugin: {
        demo: {
          title: 'Persisted Demo',
        },
      },
    };
    seedPersistentRuntimeMessages('en-US', {
      etag: '"en-US-persisted"',
      messages: persistedMessages,
    });
    requestClientGet.mockReturnValue(new Promise(() => {}));

    const messages = await loadRuntimeLocaleMessages('en-US');

    expect(messages).toEqual(persistedMessages);
    expect(getRuntimeLocaleMessagesSnapshot()).toEqual(persistedMessages);
    expect(requestClientGet).toHaveBeenCalledWith(
      '/i18n/runtime/messages',
      expect.objectContaining({
        headers: expect.objectContaining({
          'If-None-Match': '"en-US-persisted"',
        }),
        params: {
          lang: 'en-US',
        },
      }),
    );
  });

  it('keeps persistent cache unchanged when background revalidation returns 304', async () => {
    const persistedMessages = {
      plugin: {
        demo: {
          title: 'Persisted Demo',
        },
      },
    };
    seedPersistentRuntimeMessages('en-US', {
      etag: '"en-US-persisted"',
      messages: persistedMessages,
    });
    requestClientGet.mockResolvedValue({
      data: '',
      headers: {},
      status: 304,
    });

    const messages = await loadRuntimeLocaleMessages('en-US');

    expect(messages).toEqual(persistedMessages);
    await vi.waitFor(() => expect(requestClientGet).toHaveBeenCalledTimes(1));
    expect(
      JSON.parse(
        window.localStorage.getItem('linapro:i18n:runtime:en-US') || '{}',
      ),
    ).toEqual(
      expect.objectContaining({
        etag: '"en-US-persisted"',
        messages: persistedMessages,
      }),
    );
  });

  it('refreshes expired persistent cache before returning messages', async () => {
    const staleMessages = {
      plugin: {
        demo: {
          title: 'Stale Demo',
        },
      },
    };
    const freshMessages = {
      plugin: {
        demo: {
          title: 'Fresh Demo',
        },
      },
    };
    seedPersistentRuntimeMessages('en-US', {
      etag: '"en-US-stale"',
      messages: staleMessages,
      savedAt: Date.now() - runtimePersistentCacheTTL - 1000,
    });
    requestClientGet.mockResolvedValue(makeRuntimeResponse(freshMessages));

    const messages = await loadRuntimeLocaleMessages('en-US');

    expect(messages).toEqual(freshMessages);
    expect(requestClientGet).toHaveBeenCalledWith(
      '/i18n/runtime/messages',
      expect.objectContaining({
        headers: expect.not.objectContaining({
          'If-None-Match': '"en-US-stale"',
        }),
      }),
    );
  });

  it('falls back to persistent messages when refresh fails', async () => {
    const fallback = vi.fn();
    const persistedMessages = {
      plugin: {
        demo: {
          title: 'Persisted Demo',
        },
      },
    };
    seedPersistentRuntimeMessages('en-US', {
      messages: persistedMessages,
      savedAt: Date.now() - runtimePersistentCacheTTL - 1000,
    });
    requestClientGet.mockRejectedValue(new Error('network failed'));

    const messages = await loadRuntimeLocaleMessages('en-US', {
      onFallback: fallback,
    });

    expect(messages).toEqual(persistedMessages);
    expect(fallback).toHaveBeenCalledWith('network');
  });

  it('forces a reload when runtime plugin messages change', async () => {
    requestClientGet
      .mockResolvedValueOnce(
        makeRuntimeResponse({
          plugin: {
            demo: {
              title: 'Version One',
            },
          },
        }),
      )
      .mockResolvedValueOnce(
        makeRuntimeResponse(
          {
            plugin: {
              demo: {
                title: 'Version Two',
              },
            },
          },
          '"en-US-2"',
        ),
      );

    await loadRuntimeLocaleMessages('en-US');
    const versionAfterFirstLoad = runtimeI18nVersion.value;
    const reloadedMessages = await reloadRuntimeLocaleMessages('en-US');

    expect(requestClientGet).toHaveBeenCalledTimes(2);
    expect(versionAfterFirstLoad).toBeLessThan(runtimeI18nVersion.value);
    expect(
      lookupRuntimeMessageString(reloadedMessages, 'plugin.demo.title'),
    ).toBe('Version Two');
  });

  it('notifies callers when no runtime messages can be loaded', async () => {
    const fallback = vi.fn();
    requestClientGet.mockRejectedValue(new Error('network failed'));

    const messages = await loadRuntimeLocaleMessages('en-US', {
      onFallback: fallback,
    });

    expect(messages).toEqual({});
    expect(fallback).toHaveBeenCalledWith('empty');
  });

  it('stores the refreshed background messages and notifies the caller', async () => {
    const refreshed = vi.fn();
    const persistedMessages = {
      plugin: {
        demo: {
          title: 'Persisted Demo',
        },
      },
    };
    const freshMessages = {
      plugin: {
        demo: {
          title: 'Fresh Demo',
        },
      },
    };
    seedPersistentRuntimeMessages('en-US', {
      etag: '"en-US-persisted"',
      messages: persistedMessages,
    });
    requestClientGet.mockResolvedValue(
      makeRuntimeResponse(freshMessages, '"en-US-fresh"'),
    );

    const messages = await loadRuntimeLocaleMessages('en-US', {
      onBackgroundRefresh: refreshed,
    });

    expect(messages).toEqual(persistedMessages);
    await vi.waitFor(() =>
      expect(refreshed).toHaveBeenCalledWith(freshMessages),
    );
    expect(getRuntimeLocaleMessagesSnapshot()).toEqual(freshMessages);
    expect(
      JSON.parse(
        window.localStorage.getItem('linapro:i18n:runtime:en-US') || '{}',
      ),
    ).toEqual(
      expect.objectContaining({
        etag: '"en-US-fresh"',
        messages: freshMessages,
      }),
    );
  });

  it('replaces stale same-version persistent cache when the server emits a content fingerprint ETag', async () => {
    const refreshed = vi.fn();
    const staleMessages = {
      app: {
        sample: {
          metrics: {
            total: '总数',
          },
          overview: {
            kicker: 'Content operations',
          },
        },
      },
    };
    const freshMessages = {
      app: {
        sample: {
          metrics: {
            total: '总数',
            summary: '摘要',
          },
          overview: {
            kicker: '内容运营',
          },
        },
      },
    };
    seedPersistentRuntimeMessages('zh-CN', {
      etag: '"zh-CN-1"',
      messages: staleMessages,
    });
    requestClientGet.mockResolvedValue(
      makeRuntimeResponse(
        freshMessages,
        '"zh-CN-1-0123456789abcdef0123456789abcdef"',
      ),
    );

    const messages = await loadRuntimeLocaleMessages('zh-CN', {
      onBackgroundRefresh: refreshed,
    });

    expect(messages).toEqual(staleMessages);
    await vi.waitFor(() =>
      expect(refreshed).toHaveBeenCalledWith(freshMessages),
    );
    const stored = JSON.parse(
      window.localStorage.getItem('linapro:i18n:runtime:zh-CN') || '{}',
    );
    expect(stored).toEqual(
      expect.objectContaining({
        etag: '"zh-CN-1-0123456789abcdef0123456789abcdef"',
        messages: freshMessages,
      }),
    );
    expect(getRuntimeLocaleMessagesSnapshot()).toEqual(freshMessages);
  });
});
