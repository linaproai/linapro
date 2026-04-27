import type { App } from 'vue';
import type { Locale } from 'vue-i18n';

import type {
  ImportLocaleFn,
  LoadMessageFn,
  LocaleDirection,
  LocaleSetupOptions,
  RuntimeLocaleOption,
  RuntimeLocaleSwitchConfig,
  SupportedLanguagesType,
} from './typing';

import { ref, shallowRef, unref } from 'vue';
import { createI18n } from 'vue-i18n';

import { useSimpleLocale } from '@vben-core/composables';

const i18n = createI18n({
  globalInjection: true,
  legacy: false,
  locale: '',
  messages: {},
});

const direction = ref<LocaleDirection>('ltr');
const runtimeLocaleOptions = shallowRef<RuntimeLocaleOption[]>([]);
const runtimeLocaleSwitchEnabled = ref(true);

const modules = import.meta.glob('./langs/**/*.json');

const { setMessages: setSimpleLocaleMessages, setSimpleLocale } =
  useSimpleLocale();

const localesMap = loadLocalesMapFromDir(
  /\.\/langs\/([^/]+)\/(.*)\.json$/,
  modules,
);
runtimeLocaleOptions.value = buildStaticLocaleOptions(Object.keys(localesMap));
let loadMessages: LoadMessageFn;

/**
 * Load locale modules
 * @param modules
 */
function loadLocalesMap(modules: Record<string, () => Promise<unknown>>) {
  const localesMap: Record<Locale, ImportLocaleFn> = {};

  for (const [path, loadLocale] of Object.entries(modules)) {
    const key = path.match(/([\w-]*)\.(json)/)?.[1];
    if (key) {
      localesMap[key] = loadLocale as ImportLocaleFn;
    }
  }
  return localesMap;
}

/**
 * Load locale modules with directory structure
 * @param regexp - Regular expression to match language and file names
 * @param modules - The modules object containing paths and import functions
 * @returns A map of locales to their corresponding import functions
 */
function loadLocalesMapFromDir(
  regexp: RegExp,
  modules: Record<string, () => Promise<unknown>>,
): Record<Locale, ImportLocaleFn> {
  const localesRaw: Record<Locale, Record<string, () => Promise<unknown>>> = {};
  const localesMap: Record<Locale, ImportLocaleFn> = {};

  // Iterate over the modules to extract language and file names
  for (const path in modules) {
    const match = path.match(regexp);
    if (match) {
      const [_, locale, fileName] = match;
      if (locale && fileName) {
        if (!localesRaw[locale]) {
          localesRaw[locale] = {};
        }
        if (modules[path]) {
          localesRaw[locale][fileName] = modules[path];
        }
      }
    }
  }

  // Convert raw locale data into async import functions
  for (const [locale, files] of Object.entries(localesRaw)) {
    localesMap[locale] = async () => {
      const messages: Record<string, any> = {};
      for (const [fileName, importFn] of Object.entries(files)) {
        messages[fileName] = ((await importFn()) as any)?.default;
      }
      return { default: messages };
    };
  }

  return localesMap;
}

/**
 * Build fallback language options from bundled locale directories.
 * @param locales
 */
function buildStaticLocaleOptions(locales: string[]): RuntimeLocaleOption[] {
  return [...new Set(locales.map((locale) => locale.trim()).filter(Boolean))]
    .sort()
    .map((locale) => ({
      label: locale,
      value: locale,
    }));
}

/**
 * Apply runtime locale descriptors returned by the host.
 * @param options
 */
function setRuntimeLocaleOptions(
  options: RuntimeLocaleOption[],
  config: RuntimeLocaleSwitchConfig = {},
) {
  const normalizedOptions = options
    .map((option): RuntimeLocaleOption | null => {
      const value = String(option.value || '').trim();
      if (!value) {
        return null;
      }
      return {
        isDefault: option.isDefault === true,
        label: String(option.label || option.nativeName || value).trim(),
        nativeName: option.nativeName,
        value,
      } satisfies RuntimeLocaleOption;
    })
    .filter((option): option is RuntimeLocaleOption => option !== null);

  if (!normalizedOptions.length) {
    return;
  }

  runtimeLocaleOptions.value = normalizedOptions;
  runtimeLocaleSwitchEnabled.value =
    config.enabled !== false && normalizedOptions.length > 1;
}

/**
 * Set i18n language
 * @param locale
 */
function setI18nLanguage(locale: Locale) {
  i18n.global.locale.value = locale;

  direction.value = 'ltr';

  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('lang', locale);
    document.documentElement.setAttribute('dir', 'ltr');
  }
}

async function setupI18n(app: App, options: LocaleSetupOptions = {}) {
  const { defaultLocale = 'zh-CN' } = options;
  // app可以自行扩展一些第三方库和组件库的国际化
  loadMessages = options.loadMessages || (async () => ({}));
  app.use(i18n);
  await loadLocaleMessages(defaultLocale);

  // 在控制台打印警告
  i18n.global.setMissingHandler((locale, key) => {
    if (options.missingWarn && key.includes('.')) {
      console.warn(
        `[intlify] Not found '${key}' key in '${locale}' locale messages.`,
      );
    }
  });
}

/**
 * Load locale messages
 * @param lang
 */
async function loadLocaleMessages(lang: SupportedLanguagesType) {
  if (unref(i18n.global.locale) === lang) {
    return setI18nLanguage(lang);
  }
  setSimpleLocale(lang);

  const message = await localesMap[lang]?.();

  if (message?.default) {
    i18n.global.setLocaleMessage(lang, message.default);
    setSimpleLocaleMessages(
      lang,
      (message.default as Record<string, any>).common || {},
    );
  }

  const mergeMessage = await loadMessages(lang);
  i18n.global.mergeLocaleMessage(lang, mergeMessage);

  return setI18nLanguage(lang);
}

export {
  direction,
  i18n,
  loadLocaleMessages,
  loadLocalesMap,
  loadLocalesMapFromDir,
  runtimeLocaleSwitchEnabled,
  runtimeLocaleOptions,
  setupI18n,
  setRuntimeLocaleOptions,
};
