import type { Locale } from 'ant-design-vue/es/locale';

import type { App } from 'vue';

import type { LocaleSetupOptions, SupportedLanguagesType } from '@vben/locales';
import type { RuntimeLocaleMessagesLoadOptions } from '#/runtime/runtime-i18n';

import { ref } from 'vue';

import {
  $t,
  direction as localeDirection,
  i18n,
  loadLocaleMessages,
  setRuntimeLocaleOptions,
  setupI18n as coreSetup,
} from '@vben/locales';
import { preferences, updatePreferences } from '@vben/preferences';

import { message } from 'ant-design-vue';
import antdDefaultLocale from 'ant-design-vue/es/locale/zh_CN';
import dayjs from 'dayjs';
import {
  antdLocaleLoaders,
  dayjsLocaleLoaders,
} from 'virtual:lina-app-third-party-locales';

import { syncPublicFrontendSettings } from '#/runtime/public-frontend';
import {
  clearRuntimeLocaleMessagesCache,
  loadRuntimeLocaleOptions,
  loadRuntimeLocaleMessages,
  mergeMessages,
  type RuntimeLocaleOptionsResult,
} from '#/runtime/runtime-i18n';

const antdLocale = ref<Locale>(antdDefaultLocale);

const localeModules = import.meta.glob('./langs/**/*.json', {
  eager: true,
  import: 'default',
});

function buildAppLocalesMap(
  modules: Record<string, Record<string, any>>,
): Record<string, Record<string, any>> {
  const messagesMap: Record<string, Record<string, any>> = {};

  for (const [path, content] of Object.entries(modules)) {
    const match = path.match(/\.\/langs\/([^/]+)\/(.*)\.json$/);
    if (!match?.[1] || !match[2]) {
      continue;
    }

    const locale = match[1];
    const messageNamespace = match[2];
    if (!messagesMap[locale]) {
      messagesMap[locale] = {};
    }
    messagesMap[locale][messageNamespace] = content;
  }

  return messagesMap;
}

const appLocalesMap = buildAppLocalesMap(
  localeModules as Record<string, Record<string, any>>,
);

type RuntimeMessagesLoader = (
  lang: SupportedLanguagesType,
  options?: RuntimeLocaleMessagesLoadOptions,
) => Promise<Record<string, any>>;

type RuntimeLocalesLoader = (
  lang: SupportedLanguagesType,
) => Promise<RuntimeLocaleOptionsResult>;

type LocaleMessagesLoaderDependencies = {
  loadRuntimeLocales: RuntimeLocalesLoader;
  loadRuntimeMessages: RuntimeMessagesLoader;
  loadThirdPartyMessages: (lang: SupportedLanguagesType) => Promise<void>;
  notifyRuntimeFallback: () => void;
  syncPublicSettings: (lang: SupportedLanguagesType) => Promise<unknown>;
};

function notifyRuntimeBundleFallback() {
  message.warning($t('common.runtimeI18nLoadFailed'));
}

function mergeBackgroundRuntimeMessages(
  lang: SupportedLanguagesType,
  messages: Record<string, any>,
) {
  if (preferences.app.locale !== lang) {
    return;
  }
  i18n.global.mergeLocaleMessage(lang, messages);
}

function createLocaleMessagesLoader(
  dependencies: LocaleMessagesLoaderDependencies,
) {
  return async (lang: SupportedLanguagesType) => {
    void dependencies.syncPublicSettings(lang).catch(() => null);
    const runtimeLocalesPromise = dependencies
      .loadRuntimeLocales(lang)
      .then((result) => {
        setRuntimeLocaleOptions(result.options, {
          enabled: result.enabled,
        });
        return result;
      })
      .catch(() => null);

    const runtimeMessagesPromise = dependencies
      .loadRuntimeMessages(lang, {
        onBackgroundRefresh: (messages) =>
          mergeBackgroundRuntimeMessages(lang, messages),
        onFallback: dependencies.notifyRuntimeFallback,
      })
      .catch(() => {
        dependencies.notifyRuntimeFallback();
        return {};
      });

    await dependencies.loadThirdPartyMessages(lang);
    await runtimeLocalesPromise;
    const runtimeMessages = await runtimeMessagesPromise;

    return mergeMessages(appLocalesMap[lang] || {}, runtimeMessages);
  };
}

async function resolveStartupLocale(lang: SupportedLanguagesType) {
  try {
    const runtimeLocales = await loadRuntimeLocaleOptions(lang);
    setRuntimeLocaleOptions(runtimeLocales.options, {
      enabled: runtimeLocales.enabled,
    });
    return runtimeLocales.locale || runtimeLocales.defaultLocale || lang;
  } catch {
    return lang;
  }
}

/**
 * 加载应用特有的语言包
 * 这里也可以改造为从服务端获取翻译数据
 * @param lang
 */
const loadMessages = createLocaleMessagesLoader({
  loadRuntimeLocales: loadRuntimeLocaleOptions,
  loadRuntimeMessages: loadRuntimeLocaleMessages,
  loadThirdPartyMessages: loadThirdPartyMessage,
  notifyRuntimeFallback: notifyRuntimeBundleFallback,
  syncPublicSettings: syncPublicFrontendSettings,
});

async function reloadActiveLocaleMessages(
  lang: SupportedLanguagesType = preferences.app.locale,
) {
  clearRuntimeLocaleMessagesCache();
  await loadLocaleMessages(lang);
}

/**
 * 加载第三方组件库的语言包
 * @param lang
 */
async function loadThirdPartyMessage(lang: SupportedLanguagesType) {
  await Promise.all([loadAntdLocale(lang), loadDayjsLocale(lang)]);
}

function uniqueLocaleCandidates(candidates: string[]) {
  return [...new Set(candidates.map((item) => item.trim()).filter(Boolean))];
}

function splitLocaleCode(lang: SupportedLanguagesType) {
  const segments = String(lang).trim().split('-').filter(Boolean);
  const language = String(segments[0] || '').toLowerCase();
  const region = String(segments[segments.length - 1] || '').toUpperCase();
  return { language, region };
}

function buildDayjsLocaleCandidates(lang: SupportedLanguagesType) {
  const normalized = String(lang).trim().toLowerCase();
  const { language } = splitLocaleCode(lang);
  return uniqueLocaleCandidates([normalized, language, 'en']);
}

function buildUnderscoreLocaleCandidates(
  lang: SupportedLanguagesType,
  modules: Record<string, () => Promise<{ default?: unknown }>>,
) {
  const { language, region } = splitLocaleCode(lang);
  return uniqueLocaleCandidates([
    language && region ? `${language}_${region}` : '',
    findLanguageLocaleCandidate(modules, language, '_'),
    'en_US',
  ]);
}

function findLanguageLocaleCandidate(
  modules: Record<string, () => Promise<{ default?: unknown }>>,
  language: string,
  separator: string,
) {
  if (!language) {
    return '';
  }
  const languagePrefix = `${language}${separator}`;
  return (
    Object.keys(modules)
      .toSorted()
      .find((candidate) => {
        const normalizedCandidate = candidate.toLowerCase();
        return (
          normalizedCandidate === language ||
          normalizedCandidate.startsWith(languagePrefix)
        );
      }) || ''
  );
}

async function loadLocaleModule(
  modules: Record<string, () => Promise<{ default?: unknown }>>,
  candidates: string[],
) {
  for (const candidate of candidates) {
    const loader = modules[candidate];
    if (!loader) {
      continue;
    }
    const module = await loader();
    return { candidate, module };
  }
  return null;
}

/**
 * 加载dayjs的语言包
 * @param lang
 */
async function loadDayjsLocale(lang: SupportedLanguagesType) {
  const loadedLocale = await loadLocaleModule(
    dayjsLocaleLoaders,
    buildDayjsLocaleCandidates(lang),
  );
  dayjs.locale(loadedLocale?.candidate || 'en');
  if (!loadedLocale) {
    console.warn(`Failed to load dayjs locale for ${lang}; fallback to en`);
  }
}

/**
 * 加载antd的语言包
 * @param lang
 */
async function loadAntdLocale(lang: SupportedLanguagesType) {
  const loadedLocale = await loadLocaleModule(
    antdLocaleLoaders,
    buildUnderscoreLocaleCandidates(lang, antdLocaleLoaders),
  );
  antdLocale.value =
    (loadedLocale?.module.default as Locale | undefined) || antdDefaultLocale;
}

async function setupI18n(app: App, options: LocaleSetupOptions = {}) {
  const defaultLocale = await resolveStartupLocale(preferences.app.locale);
  if (preferences.app.locale !== defaultLocale) {
    updatePreferences({
      app: {
        locale: defaultLocale,
      },
    });
  }
  await coreSetup(app, {
    defaultLocale,
    loadMessages,
    missingWarn: !import.meta.env.PROD,
    ...options,
  });
}

export { $t, antdLocale, localeDirection, setupI18n };
export {
  createLocaleMessagesLoader,
  loadMessages,
  reloadActiveLocaleMessages,
  resolveStartupLocale,
};
