import { effectScope, nextTick } from 'vue';

import { describe, expect, it, vi } from 'vitest';

import { setupVbenVxeTable } from './init';

const { localeLoaders, preferencesMock, vxeUiMock } = await vi.hoisted(
  async () => {
    const { ref } = await import('vue');

    return {
      localeLoaders: {
        enUS: vi.fn(async () => ({ default: { language: 'en-US' } })),
        zhCN: vi.fn(async () => ({ default: { language: 'zh-CN' } })),
      },
      preferencesMock: {
        isDark: ref(false),
        locale: ref('en-US'),
      },
      vxeUiMock: {
        component: vi.fn(),
        setI18n: vi.fn(),
        setLanguage: vi.fn(),
        setTheme: vi.fn(),
      },
    };
  },
);

vi.mock('@vben/preferences', () => ({
  usePreferences: () => preferencesMock,
}));

vi.mock('@vben-core/form-ui', () => ({
  useVbenForm: vi.fn(),
}));

vi.mock('virtual:lina-vxe-locales', () => ({
  vxeLocaleLoaders: {
    'en-US': localeLoaders.enUS,
    'zh-CN': localeLoaders.zhCN,
  },
}));

vi.mock('vxe-pc-ui/lib/language/en-US', () => ({
  default: { language: 'en-US' },
}));

vi.mock('vxe-pc-ui', () => ({
  VxeButton: {},
  VxeCheckbox: {},
  VxeIcon: {},
  VxeInput: {},
  VxeLoading: {},
  VxeModal: {},
  VxeNumberInput: {},
  VxePager: {},
  VxeRadioGroup: {},
  VxeSelect: {},
  VxeTooltip: {},
  VxeUI: vxeUiMock,
  VxeUpload: {},
}));

vi.mock('vxe-table', () => ({
  VxeColgroup: {},
  VxeColumn: {},
  VxeGrid: {},
  VxeTable: {},
  VxeToolbar: {},
}));

vi.mock('./extends', () => ({
  extendsDefaultFormatter: vi.fn(),
}));

describe('setupVbenVxeTable', () => {
  it('keeps theme and locale updates independent', async () => {
    const scope = effectScope();

    try {
      scope.run(() => {
        setupVbenVxeTable({
          configVxeTable: vi.fn(),
          useVbenForm: vi.fn() as never,
        });
      });

      await vi.waitFor(() => {
        expect(vxeUiMock.setLanguage).toHaveBeenCalledWith('en-US');
      });
      expect(vxeUiMock.setTheme).toHaveBeenCalledWith('light');

      vi.clearAllMocks();
      preferencesMock.isDark.value = true;
      await nextTick();

      expect(vxeUiMock.setTheme).toHaveBeenCalledOnce();
      expect(vxeUiMock.setTheme).toHaveBeenCalledWith('dark');
      expect(localeLoaders.enUS).not.toHaveBeenCalled();
      expect(localeLoaders.zhCN).not.toHaveBeenCalled();
      expect(vxeUiMock.setI18n).not.toHaveBeenCalled();
      expect(vxeUiMock.setLanguage).not.toHaveBeenCalled();

      vi.clearAllMocks();
      preferencesMock.locale.value = 'zh-CN';
      await nextTick();

      await vi.waitFor(() => {
        expect(vxeUiMock.setLanguage).toHaveBeenCalledWith('zh-CN');
      });
      expect(localeLoaders.zhCN).toHaveBeenCalledOnce();
      expect(vxeUiMock.setI18n).toHaveBeenCalledWith('zh-CN', {
        language: 'zh-CN',
      });
      expect(vxeUiMock.setTheme).not.toHaveBeenCalled();
    } finally {
      scope.stop();
      preferencesMock.isDark.value = false;
      preferencesMock.locale.value = 'en-US';
    }
  });
});
