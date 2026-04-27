import type { Locale } from 'ant-design-vue/es/locale';

import type { App } from 'vue';

import type { LocaleSetupOptions, SupportedLanguagesType } from '@vben/locales';
import type { RuntimeLocaleMessagesLoadOptions } from '#/runtime/runtime-i18n';

import { ref } from 'vue';

import {
  $t,
  i18n,
  loadLocaleMessages,
  setupI18n as coreSetup,
} from '@vben/locales';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';
import antdEnLocale from 'ant-design-vue/es/locale/en_US';
import antdDefaultLocale from 'ant-design-vue/es/locale/zh_CN';
import dayjs from 'dayjs';

import { syncPublicFrontendSettings } from '#/runtime/public-frontend';
import {
  clearRuntimeLocaleMessagesCache,
  loadRuntimeLocaleMessages,
  mergeMessages,
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

type LocaleMessagesLoaderDependencies = {
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
    const runtimeMessages = await runtimeMessagesPromise;

    return mergeMessages(appLocalesMap[lang] || {}, runtimeMessages);
  };
}

/**
 * 加载应用特有的语言包
 * 这里也可以改造为从服务端获取翻译数据
 * @param lang
 */
const loadMessages = createLocaleMessagesLoader({
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

/**
 * 加载dayjs的语言包
 * @param lang
 */
async function loadDayjsLocale(lang: SupportedLanguagesType) {
  let locale;
  switch (lang) {
    case 'en-US': {
      locale = await import('dayjs/locale/en');
      break;
    }
    case 'zh-CN': {
      locale = await import('dayjs/locale/zh-cn');
      break;
    }
    // 默认使用英语
    default: {
      locale = await import('dayjs/locale/en');
    }
  }
  if (locale) {
    dayjs.locale(locale);
  } else {
    console.error(`Failed to load dayjs locale for ${lang}`);
  }
}

/**
 * 加载antd的语言包
 * @param lang
 */
async function loadAntdLocale(lang: SupportedLanguagesType) {
  switch (lang) {
    case 'en-US': {
      antdLocale.value = antdEnLocale;
      break;
    }
    case 'zh-CN': {
      antdLocale.value = antdDefaultLocale;
      break;
    }
  }
}

async function setupI18n(app: App, options: LocaleSetupOptions = {}) {
  await coreSetup(app, {
    defaultLocale: preferences.app.locale,
    loadMessages,
    missingWarn: !import.meta.env.PROD,
    ...options,
  });
}

export { $t, antdLocale, setupI18n };
export { createLocaleMessagesLoader, loadMessages, reloadActiveLocaleMessages };
