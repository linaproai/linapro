import { i18n } from '@vben/locales';

import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import enUSPages from '#/locales/langs/en-US/pages.json';
import zhCNPages from '#/locales/langs/zh-CN/pages.json';

import {
  buildPluginDetailHostServiceCards,
  formatServiceLabel,
  knownHostServiceLabels,
} from './plugin-host-service-view';

describe('formatServiceLabel', () => {
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

  it('localizes all known host service labels in Chinese', () => {
    setActiveLocale('zh-CN');

    expect(formatServiceLabel('data')).toBe('数据');
    expect(formatServiceLabel('tenant')).toBe('租户');
    expect(formatServiceLabel('org')).toBe('组织');
    expect(formatServiceLabel('users')).toBe('用户');
    expect(formatServiceLabel('auth')).toBe('认证授权');
    expect(formatServiceLabel('ai')).toBe('AI');
    expect(formatServiceLabel('cache')).toBe('缓存');
    expect(formatServiceLabel('hostconfig')).toBe('宿主配置');
  });

  it('localizes all known host service labels in English', () => {
    setActiveLocale('en-US');

    expect(formatServiceLabel('data')).toBe('Data');
    expect(formatServiceLabel('tenant')).toBe('Tenant');
    expect(formatServiceLabel('org')).toBe('Organization');
    expect(formatServiceLabel('users')).toBe('Users');
    expect(formatServiceLabel('auth')).toBe('Auth');
    expect(formatServiceLabel('ai')).toBe('AI');
    expect(formatServiceLabel('cache')).toBe('Cache');
    expect(formatServiceLabel('hostconfig')).toBe('Host Config');
  });

  it('uses a consistent unknown-service fallback instead of raw wire names only', () => {
    setActiveLocale('zh-CN');
    expect(formatServiceLabel('custom-domain')).toBe(
      '未知服务 (custom-domain)',
    );

    setActiveLocale('en-US');
    expect(formatServiceLabel('custom-domain')).toBe(
      'Unknown Service (custom-domain)',
    );
  });

  it('returns empty string for blank service identifiers', () => {
    setActiveLocale('zh-CN');
    expect(formatServiceLabel('')).toBe('');
    expect(formatServiceLabel('   ')).toBe('');
  });

  it('keeps knownHostServiceLabels aligned with bilingual i18n service keys', () => {
    const zhServiceKeys = Object.keys(
      zhCNPages.system.plugin.hostServices.service,
    ).filter((key) => key !== 'unknown');
    const enServiceKeys = Object.keys(
      enUSPages.system.plugin.hostServices.service,
    ).filter((key) => key !== 'unknown');

    expect([...knownHostServiceLabels].sort()).toEqual(
      [...zhServiceKeys].sort(),
    );
    expect([...knownHostServiceLabels].sort()).toEqual(
      [...enServiceKeys].sort(),
    );
  });

  it('builds detail cards with localized titles for previously unmapped services', () => {
    setActiveLocale('zh-CN');

    const cards = buildPluginDetailHostServiceCards(
      [
        { methods: ['tenant.directory.list'], service: 'tenant' },
        { methods: ['list'], service: 'data', tables: ['demo_table'] },
      ],
      [
        { methods: ['tenant.directory.list'], service: 'tenant' },
        { methods: ['list'], service: 'data', tables: ['demo_table'] },
      ],
    );

    expect(cards.map((card) => card.title)).toEqual(['数据', '租户']);
    expect(cards.map((card) => card.service)).toEqual(['data', 'tenant']);
  });
});
