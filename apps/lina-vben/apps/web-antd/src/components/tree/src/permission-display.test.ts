import { i18n } from '@vben/locales';

import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import enUSPages from '#/locales/langs/en-US/pages.json';
import zhCNPages from '#/locales/langs/zh-CN/pages.json';

import {
  formatMenuPermissionLabel,
  formatMenuPermissionShortLabel,
} from './permission-display';

describe('formatMenuPermissionLabel', () => {
  beforeEach(() => {
    i18n.global.setLocaleMessage('en-US', { pages: enUSPages });
    i18n.global.setLocaleMessage('zh-CN', { pages: zhCNPages });
  });

  afterEach(() => {
    document.documentElement.lang = '';
  });

  function setActiveLocale(locale: 'en-US' | 'zh-CN') {
    document.documentElement.lang = locale;
    i18n.global.locale.value = locale;
  }

  it('formats English-prefixed dynamic route permissions in Chinese locale', () => {
    setActiveLocale('zh-CN');

    expect(
      formatMenuPermissionLabel(
        'Dynamic Route Permission:plugin-dynamic-host-auth-ui:review:query',
      ),
    ).toBe('动态路由权限（资源：审核，动作：查询）');
  });

  it('formats raw dynamic route permissions in English locale', () => {
    setActiveLocale('en-US');

    expect(
      formatMenuPermissionLabel('plugin-dynamic-host-auth-ui:audit:query'),
    ).toBe('Dynamic Route Permission (resource: Audit, action: Query)');
  });

  it('falls back to readable unknown segments without translation entries', () => {
    setActiveLocale('en-US');

    expect(
      formatMenuPermissionLabel(
        'plugin-dynamic-host-auth-ui:report-center:read',
      ),
    ).toBe('Dynamic Route Permission (resource: Report Center, action: Read)');
  });

  it('formats dynamic route permission buttons as short English menu names', () => {
    setActiveLocale('en-US');

    expect(
      formatMenuPermissionShortLabel(
        'Dynamic Route Permission:plugin-demo-dynamic:record:create',
      ),
    ).toBe('Record Create');
  });

  it('formats dynamic route permission buttons as short Chinese menu names', () => {
    setActiveLocale('zh-CN');

    expect(
      formatMenuPermissionShortLabel(
        'Dynamic Route Permission:plugin-demo-dynamic:backend:view',
      ),
    ).toBe('后端查看');
  });
});
