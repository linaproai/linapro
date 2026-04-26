import { defineOverridesPreferences } from '@vben/preferences';

/**
 * @description 项目配置文件
 * 只需要覆盖项目中的一部分配置，不需要的配置不用覆盖，会自动使用默认配置
 * !!! 更改配置后请清空缓存，否则可能不生效
 */
export const overridesPreferences = defineOverridesPreferences({
  // overrides
  app: {
    accessMode: 'backend',
    authPageLayout: 'panel-right',
    enableRefreshToken: false,
    name: import.meta.env.VITE_APP_TITLE,
  },
  copyright: {
    companyName: 'LinaPro.AI',
    companySiteLink: 'https://linapro.ai',
    date: String(new Date().getFullYear()),
    enable: true,
    settingShow: false,
  },
  footer: {
    enable: true,
    settingShow: false,
  },
  theme: {
    mode: 'light',
  },
});
