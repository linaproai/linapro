import { ref, shallowRef } from 'vue';

import { preferences } from '@vben/preferences';
import type { RequestResponse } from '@vben/request';

import { requestClient } from '#/api/request';

const runtimeI18nLocale = ref('');
const runtimeI18nVersion = ref(0);
const runtimeLocaleMessages = shallowRef<Record<string, any>>({});
const runtimePersistentCachePrefix = 'linapro:i18n:runtime:';
const runtimePersistentCacheTTL = 7 * 24 * 60 * 60 * 1000;
const runtimeRequestMaxAttempts = 2;

type RuntimeMessagesPayload = {
  messages?: Record<string, any>;
};

type RuntimeMessagesResponse = RuntimeMessagesPayload | {
  data?: RuntimeMessagesPayload;
};

type RuntimePersistentCacheEntry = {
  etag: string;
  messages: Record<string, any>;
  savedAt: number;
};

type RuntimeFallbackReason = 'empty' | 'network';

type RuntimeLocaleMessagesLoadOptions = {
  force?: boolean;
  onBackgroundRefresh?: (messages: Record<string, any>) => void;
  onFallback?: (reason: RuntimeFallbackReason) => void;
};

function getRuntimePersistentCacheKey(locale: string) {
  return `${runtimePersistentCachePrefix}${locale}`;
}

function isPlainObject(value: unknown): value is Record<string, any> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function mergeMessages(
  target: Record<string, any>,
  source: Record<string, any>,
): Record<string, any> {
  const output: Record<string, any> = { ...target };

  for (const [key, value] of Object.entries(source)) {
    if (isPlainObject(value) && isPlainObject(output[key])) {
      output[key] = mergeMessages(output[key], value);
      continue;
    }
    output[key] = value;
  }

  return output;
}

function setRuntimeLocaleMessages(
  locale: string,
  messages: Record<string, any>,
) {
  runtimeI18nLocale.value = locale;
  runtimeLocaleMessages.value = messages;
  runtimeI18nVersion.value += 1;
}

function getRuntimeLocaleMessagesSnapshot() {
  return runtimeLocaleMessages.value;
}

function clearRuntimeLocaleMessagesCache() {
  setRuntimeLocaleMessages('', {});
}

function getRuntimeStorage(): null | Storage {
  if (typeof window === 'undefined' || !window.localStorage) {
    return null;
  }
  return window.localStorage;
}

function readPersistentRuntimeLocaleMessages(
  locale: string,
): null | RuntimePersistentCacheEntry {
  try {
    const storage = getRuntimeStorage();
    const rawEntry = storage?.getItem(getRuntimePersistentCacheKey(locale));
    if (!rawEntry) {
      return null;
    }

    const entry = JSON.parse(rawEntry) as Partial<RuntimePersistentCacheEntry>;
    if (
      typeof entry.etag !== 'string' ||
      typeof entry.savedAt !== 'number' ||
      !isPlainObject(entry.messages)
    ) {
      return null;
    }

    return {
      etag: entry.etag,
      messages: entry.messages,
      savedAt: entry.savedAt,
    };
  } catch {
    return null;
  }
}

function writePersistentRuntimeLocaleMessages(
  locale: string,
  entry: RuntimePersistentCacheEntry,
) {
  try {
    getRuntimeStorage()?.setItem(
      getRuntimePersistentCacheKey(locale),
      JSON.stringify(entry),
    );
  } catch {
    // Storage quota or privacy-mode failures should not block runtime i18n.
  }
}

function isPersistentRuntimeCacheFresh(
  entry: null | RuntimePersistentCacheEntry,
) {
  return !!entry && Date.now() - entry.savedAt <= runtimePersistentCacheTTL;
}

function lookupRuntimeMessageString(
  messages: Record<string, any>,
  key: string,
): string {
  const segments = key.split('.').filter(Boolean);
  if (!segments.length) {
    return '';
  }

  let current: any = messages;
  for (const segment of segments) {
    if (!isPlainObject(current) || !(segment in current)) {
      return '';
    }
    current = current[segment];
  }
  return typeof current === 'string' ? current : '';
}

function normalizeRuntimeMessagesPayload(payload?: RuntimeMessagesResponse) {
  const payloadObject: Record<string, any> = isPlainObject(payload)
    ? payload
    : {};
  const dataObject = isPlainObject(payloadObject.data) ? payloadObject.data : {};
  const runtimeMessages = dataObject.messages ?? payloadObject.messages ?? {};
  return isPlainObject(runtimeMessages) ? runtimeMessages : {};
}

function readResponseHeader(
  response: RequestResponse<RuntimeMessagesResponse>,
  name: string,
) {
  const headers = response.headers as any;
  if (!headers) {
    return '';
  }
  if (typeof headers.get === 'function') {
    return headers.get(name) ?? headers.get(name.toLowerCase()) ?? '';
  }
  return headers[name] ?? headers[name.toLowerCase()] ?? '';
}

async function requestRuntimeLocaleMessages(
  locale: string,
  etag?: string,
): Promise<RequestResponse<RuntimeMessagesResponse>> {
  let lastError: unknown;

  for (let attempt = 1; attempt <= runtimeRequestMaxAttempts; attempt += 1) {
    try {
      return await requestClient.get<RequestResponse<RuntimeMessagesResponse>>(
        '/i18n/runtime/messages',
        {
          headers: {
            'Accept-Language': locale,
            ...(etag ? { 'If-None-Match': etag } : {}),
          },
          responseReturn: 'raw',
          validateStatus: (status) => status >= 200 && status < 400,
        },
      );
    } catch (error) {
      lastError = error;
    }
  }

  throw lastError;
}

async function refreshRuntimeLocaleMessages(
  locale: string,
  persistentEntry: null | RuntimePersistentCacheEntry,
  options: RuntimeLocaleMessagesLoadOptions,
  useConditionalRequest = true,
) {
  try {
    const response = await requestRuntimeLocaleMessages(
      locale,
      useConditionalRequest ? persistentEntry?.etag : undefined,
    );

    if (response.status === 304) {
      if (persistentEntry) {
        return persistentEntry.messages;
      }
      return {};
    }

    const runtimeMessages = normalizeRuntimeMessagesPayload(response.data);
    const etag = String(readResponseHeader(response, 'etag') || '');
    setRuntimeLocaleMessages(locale, runtimeMessages);
    writePersistentRuntimeLocaleMessages(locale, {
      etag,
      messages: runtimeMessages,
      savedAt: Date.now(),
    });
    return runtimeMessages;
  } catch {
    if (persistentEntry) {
      setRuntimeLocaleMessages(locale, persistentEntry.messages);
      options.onFallback?.('network');
      return persistentEntry.messages;
    }
    options.onFallback?.('empty');
    return {};
  }
}

async function loadRuntimeLocaleMessages(
  locale?: string,
  options: RuntimeLocaleMessagesLoadOptions = {},
) {
  const activeLocale = locale || preferences.app.locale;
  if (
    !options.force &&
    runtimeI18nLocale.value === activeLocale &&
    isPlainObject(runtimeLocaleMessages.value) &&
    Object.keys(runtimeLocaleMessages.value).length > 0
  ) {
    return runtimeLocaleMessages.value;
  }

  const persistentEntry = readPersistentRuntimeLocaleMessages(activeLocale);
  const persistentEntryIsFresh =
    isPersistentRuntimeCacheFresh(persistentEntry);

  if (!options.force && persistentEntryIsFresh && persistentEntry) {
    setRuntimeLocaleMessages(activeLocale, persistentEntry.messages);
    void refreshRuntimeLocaleMessages(
      activeLocale,
      persistentEntry,
      options,
    ).then((messages) => {
      if (messages !== persistentEntry.messages) {
        options.onBackgroundRefresh?.(messages);
      }
    });
    return persistentEntry.messages;
  }

  return await refreshRuntimeLocaleMessages(
    activeLocale,
    persistentEntry,
    options,
    !options.force && persistentEntryIsFresh,
  );
}

async function reloadRuntimeLocaleMessages(locale?: string) {
  return await loadRuntimeLocaleMessages(locale, { force: true });
}

export {
  clearRuntimeLocaleMessagesCache,
  getRuntimeLocaleMessagesSnapshot,
  loadRuntimeLocaleMessages,
  lookupRuntimeMessageString,
  mergeMessages,
  reloadRuntimeLocaleMessages,
  runtimePersistentCacheTTL,
  runtimeI18nVersion,
};
export type { RuntimeLocaleMessagesLoadOptions };
