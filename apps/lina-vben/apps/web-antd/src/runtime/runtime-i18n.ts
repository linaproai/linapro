import { ref, shallowRef } from 'vue';

import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';

const runtimeI18nFetchInit: RequestInit = {
  cache: 'no-store',
  credentials: 'same-origin',
  method: 'GET',
};

const runtimeI18nLocale = ref('');
const runtimeI18nVersion = ref(0);
const runtimeLocaleMessages = shallowRef<Record<string, any>>({});

function resolveRuntimeI18nEndpoint() {
  const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
  return `${apiURL.replace(/\/$/, '')}/i18n/runtime/messages`;
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

async function loadRuntimeLocaleMessages(
  locale?: string,
  options: {
    force?: boolean;
  } = {},
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

  try {
    const response = await fetch(resolveRuntimeI18nEndpoint(), {
      ...runtimeI18nFetchInit,
      headers: {
        'Accept-Language': activeLocale,
      },
    });
    if (!response.ok) {
      return {};
    }

    const payload = await response.json();
    const runtimeMessages = payload?.data?.messages ?? payload?.messages ?? {};
    const normalizedMessages = isPlainObject(runtimeMessages)
      ? runtimeMessages
      : {};
    setRuntimeLocaleMessages(activeLocale, normalizedMessages);
    return normalizedMessages;
  } catch {
    return {};
  }
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
  runtimeI18nVersion,
};
