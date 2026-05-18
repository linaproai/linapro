import type { SetupVxeTable } from './types';

import { defineComponent, watch } from 'vue';

import { usePreferences } from '@vben/preferences';

import { useVbenForm } from '@vben-core/form-ui';

import {
  VxeButton,
  VxeCheckbox,

  // VxeFormGather,
  // VxeForm,
  // VxeFormItem,
  VxeIcon,
  VxeInput,
  VxeLoading,
  VxeModal,
  VxeNumberInput,
  VxePager,
  // VxeList,
  // VxeModal,
  // VxeOptgroup,
  // VxeOption,
  // VxePulldown,
  // VxeRadio,
  // VxeRadioButton,
  VxeRadioGroup,
  VxeSelect,
  VxeTooltip,
  VxeUI,
  VxeUpload,
  // VxeSwitch,
  // VxeTextarea,
} from 'vxe-pc-ui';
import enUS from 'vxe-pc-ui/lib/language/en-US';
import { vxeLocaleLoaders } from 'virtual:lina-vxe-locales';
import {
  VxeColgroup,
  VxeColumn,
  VxeGrid,
  VxeTable,
  VxeToolbar,
} from 'vxe-table';

import { extendsDefaultFormatter } from './extends';

// 是否加载过
let isInit = false;

// eslint-disable-next-line import/no-mutable-exports
export let useTableForm: typeof useVbenForm;

type VxeLocale = Parameters<typeof VxeUI.setLanguage>[0];

// 部分组件，如果没注册，vxe-table 会报错，这里实际没用组件，只是为了不报错，同时可以减少打包体积
const createVirtualComponent = (name = '') => {
  return defineComponent({
    name,
  });
};

function uniqueLocaleCandidates(candidates: string[]) {
  return [...new Set(candidates.map((item) => item.trim()).filter(Boolean))];
}

function splitLocaleCode(locale: string) {
  const segments = String(locale).trim().split('-').filter(Boolean);
  const language = String(segments[0] || '').toLowerCase();
  const region = String(segments[segments.length - 1] || '').toUpperCase();
  return { language, region };
}

function buildVxeLocaleCandidates(locale: string) {
  const { language, region } = splitLocaleCode(locale);
  return uniqueLocaleCandidates([
    language && region ? `${language}-${region}` : '',
    findLanguageLocaleCandidate(language),
    'en-US',
  ]);
}

function findLanguageLocaleCandidate(language: string) {
  if (!language) {
    return '';
  }
  const languagePrefix = `${language}-`;
  return (
    Object.keys(vxeLocaleLoaders)
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

async function loadVxeLocale(locale: string) {
  for (const candidate of buildVxeLocaleCandidates(locale)) {
    const loader = vxeLocaleLoaders[candidate];
    if (!loader) {
      continue;
    }
    const module = await loader();
    return {
      locale: candidate as VxeLocale,
      messages: module.default || module,
    };
  }

  return {
    locale: 'en-US' as VxeLocale,
    messages: enUS,
  };
}

export function initVxeTable() {
  if (isInit) {
    return;
  }

  VxeUI.component(VxeTable);
  VxeUI.component(VxeColumn);
  VxeUI.component(VxeColgroup);
  VxeUI.component(VxeGrid);
  VxeUI.component(VxeToolbar);

  VxeUI.component(VxeButton);
  // VxeUI.component(VxeButtonGroup);
  VxeUI.component(VxeCheckbox);
  // VxeUI.component(VxeCheckboxGroup);
  VxeUI.component(createVirtualComponent('VxeForm'));
  // VxeUI.component(VxeFormGather);
  // VxeUI.component(VxeFormItem);
  VxeUI.component(VxeIcon);
  VxeUI.component(VxeInput);
  // VxeUI.component(VxeList);
  VxeUI.component(VxeLoading);
  VxeUI.component(VxeModal);
  VxeUI.component(VxeNumberInput);
  // VxeUI.component(VxeOptgroup);
  // VxeUI.component(VxeOption);
  VxeUI.component(VxePager);
  // VxeUI.component(VxePulldown);
  // VxeUI.component(VxeRadio);
  // VxeUI.component(VxeRadioButton);
  VxeUI.component(VxeRadioGroup);
  VxeUI.component(VxeSelect);
  // VxeUI.component(VxeSwitch);
  // VxeUI.component(VxeTextarea);
  VxeUI.component(VxeTooltip);
  VxeUI.component(VxeUpload);

  isInit = true;
}

export function setupVbenVxeTable(setupOptions: SetupVxeTable) {
  const { configVxeTable, useVbenForm } = setupOptions;

  initVxeTable();
  useTableForm = useVbenForm;

  const { isDark, locale } = usePreferences();

  watch(
    [() => isDark.value, () => locale.value],
    ([isDarkValue, localeValue]) => {
      VxeUI.setTheme(isDarkValue ? 'dark' : 'light');
      void loadVxeLocale(localeValue).then((vxeLocale) => {
        VxeUI.setI18n(vxeLocale.locale, vxeLocale.messages);
        VxeUI.setLanguage(vxeLocale.locale);
      });
    },
    {
      immediate: true,
    },
  );

  extendsDefaultFormatter(VxeUI);

  configVxeTable(VxeUI);
}
