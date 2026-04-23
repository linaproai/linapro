import type { AuthPageLayoutType } from '@vben/types';

import { reactive, readonly } from 'vue';

import { useAppConfig } from '@vben/hooks';
import { preferencesManager, updatePreferences } from '@vben/preferences';

interface PublicFrontendAppSettings {
  logo: string;
  logoDark: string;
  name: string;
}

interface PublicFrontendAuthSettings {
  loginSubtitle: string;
  panelLayout: AuthPageLayoutType;
  pageDesc: string;
  pageTitle: string;
}

interface PublicFrontendUISettings {
  layout: string;
  themeMode: string;
  watermarkContent: string;
  watermarkEnabled: boolean;
}

interface PublicFrontendCronShellSettings {
  disabledReason: string;
  enabled: boolean;
  supported: boolean;
}

interface PublicFrontendCronLogRetentionSettings {
  mode: string;
  value: number;
}

interface PublicFrontendCronTimezoneSettings {
  current: string;
}

interface PublicFrontendCronSettings {
  logRetention: PublicFrontendCronLogRetentionSettings;
  shell: PublicFrontendCronShellSettings;
  timezone: PublicFrontendCronTimezoneSettings;
}

interface PublicFrontendSettings {
  app: PublicFrontendAppSettings;
  auth: PublicFrontendAuthSettings;
  cron: PublicFrontendCronSettings;
  ui: PublicFrontendUISettings;
}

const publicFrontendFetchInit: RequestInit = {
  // Public frontend settings are managed by sys_config. Force each sync to bypass
  // the browser HTTP cache so the same browser immediately sees the latest theme
  // and branding values after backend updates.
  cache: 'no-store',
  credentials: 'same-origin',
  method: 'GET',
};

const publicFrontendState = reactive<PublicFrontendSettings>({
  app: {
    logo: '',
    logoDark: '',
    name: '',
  },
  auth: {
    loginSubtitle: '',
    panelLayout: 'panel-center',
    pageDesc: '',
    pageTitle: '',
  },
  cron: {
    logRetention: {
      mode: 'days',
      value: 30,
    },
    shell: {
      disabledReason: '',
      enabled: false,
      supported: true,
    },
    timezone: {
      current: 'Asia/Shanghai',
    },
  },
  ui: {
    layout: '',
    themeMode: '',
    watermarkContent: '',
    watermarkEnabled: false,
  },
});

function normalizeString(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function normalizeBoolean(value: unknown): boolean {
  return value === true || value === 'true';
}

function normalizeNumber(value: unknown, fallback: number): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function normalizeAuthPanelLayout(value: unknown): AuthPageLayoutType {
  const normalized = normalizeString(value);
  switch (normalized) {
    case 'panel-left':
    case 'panel-center':
    case 'panel-right':
      return normalized;
    default:
      return 'panel-center';
  }
}

function resolvePublicFrontendEndpoint(): string {
  const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
  return `${apiURL.replace(/\/$/, '')}/config/public/frontend`;
}

function normalizeCronLogRetentionSettings(payload: any) {
  const mode = normalizeString(payload?.mode) || 'days';
  const fallbackValue = mode === 'none' ? 0 : 30;
  const value = normalizeNumber(payload?.value, fallbackValue);

  if (mode === 'none') {
    return {
      mode,
      value: 0,
    };
  }

  return {
    mode,
    value: value > 0 ? value : fallbackValue,
  };
}

function normalizeCronTimezoneSettings(payload: any) {
  return {
    current: normalizeString(payload?.current) || 'Asia/Shanghai',
  };
}

function normalizePublicFrontendSettings(payload: any): PublicFrontendSettings {
  const app = payload?.app ?? {};
  const auth = payload?.auth ?? {};
  const cron = payload?.cron ?? {};
  const logRetention = cron?.logRetention ?? {};
  const shell = cron?.shell ?? {};
  const timezone = cron?.timezone ?? {};
  const ui = payload?.ui ?? {};

  return {
    app: {
      logo: normalizeString(app.logo),
      logoDark: normalizeString(app.logoDark),
      name: normalizeString(app.name),
    },
    auth: {
      loginSubtitle: normalizeString(auth.loginSubtitle),
      panelLayout: normalizeAuthPanelLayout(auth.panelLayout),
      pageDesc: normalizeString(auth.pageDesc),
      pageTitle: normalizeString(auth.pageTitle),
    },
    cron: {
      logRetention: normalizeCronLogRetentionSettings(logRetention),
      shell: {
        disabledReason: normalizeString(shell.disabledReason),
        enabled: normalizeBoolean(shell.enabled),
        supported:
          shell?.supported === undefined
            ? true
            : normalizeBoolean(shell.supported),
      },
      timezone: normalizeCronTimezoneSettings(timezone),
    },
    ui: {
      layout: normalizeString(ui.layout),
      themeMode: normalizeString(ui.themeMode),
      watermarkContent: normalizeString(ui.watermarkContent),
      watermarkEnabled: normalizeBoolean(ui.watermarkEnabled),
    },
  };
}

function applyPublicFrontendPreferences(settings: PublicFrontendSettings) {
  const initial = preferencesManager.getInitialPreferences();
  const logoSource = settings.app.logo || initial.logo.source;
  const logoSourceDark =
    settings.app.logoDark || initial.logo.sourceDark || logoSource;

  updatePreferences({
    app: {
      authPageLayout: settings.auth.panelLayout,
      layout: (settings.ui.layout || initial.app.layout) as any,
      name: settings.app.name || initial.app.name,
      watermark: settings.ui.watermarkEnabled,
      watermarkContent:
        settings.ui.watermarkContent || initial.app.watermarkContent,
    },
    logo: {
      source: logoSource,
      sourceDark: logoSourceDark,
    },
    theme: {
      builtinType: initial.theme.builtinType,
      colorPrimary: initial.theme.colorPrimary,
      mode: (settings.ui.themeMode || initial.theme.mode) as any,
    },
  });
}

async function syncPublicFrontendSettings() {
  try {
    const response = await fetch(
      resolvePublicFrontendEndpoint(),
      publicFrontendFetchInit,
    );
    if (!response.ok) {
      return null;
    }

    const payload = await response.json();
    const settings = normalizePublicFrontendSettings(payload?.data ?? payload);

    Object.assign(publicFrontendState.app, settings.app);
    Object.assign(publicFrontendState.auth, settings.auth);
    Object.assign(
      publicFrontendState.cron.logRetention,
      settings.cron.logRetention,
    );
    Object.assign(publicFrontendState.cron.shell, settings.cron.shell);
    Object.assign(publicFrontendState.cron.timezone, settings.cron.timezone);
    Object.assign(publicFrontendState.ui, settings.ui);
    applyPublicFrontendPreferences(settings);

    return settings;
  } catch {
    return null;
  }
}

export { syncPublicFrontendSettings };
export const publicFrontendSettings = readonly(publicFrontendState);
export type { PublicFrontendSettings };
