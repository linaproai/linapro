import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';

const runtimeI18nFetchInit: RequestInit = {
  cache: 'no-store',
  credentials: 'same-origin',
  method: 'GET',
};

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

async function loadRuntimeLocaleMessages(locale?: string) {
  const activeLocale = locale || preferences.app.locale;

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
    return isPlainObject(runtimeMessages) ? runtimeMessages : {};
  } catch {
    return {};
  }
}

export { loadRuntimeLocaleMessages, mergeMessages };
