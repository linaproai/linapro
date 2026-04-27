/**
 * @zh_CN 登录页面 url 地址
 */
export const LOGIN_PATH = '/auth/login';

export interface LanguageOption {
  label: string;
  value: string;
}

/**
 * Static language fallback kept for legacy imports. Runtime applications should
 * consume locale options from @vben/locales so new languages are resource-driven.
 */
export const SUPPORT_LANGUAGES: LanguageOption[] = [];
