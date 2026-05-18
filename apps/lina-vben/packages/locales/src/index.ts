import {
  direction,
  i18n,
  loadLocaleMessages,
  loadLocalesMap,
  loadLocalesMapFromDir,
  runtimeLocaleSwitchEnabled,
  runtimeLocaleOptions,
  setupI18n,
  setRuntimeLocaleOptions,
} from './i18n';

const $t = i18n.global.t;
const $te = i18n.global.te;

export {
  $t,
  $te,
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
export {
  type ImportLocaleFn,
  type LocaleDirection,
  type LocaleSetupOptions,
  type RuntimeLocaleOption,
  type RuntimeLocaleSwitchConfig,
  type SupportedLanguagesType,
} from './typing';
export type { CompileError } from '@intlify/core-base';

export { useI18n } from 'vue-i18n';

export type { Locale } from 'vue-i18n';
