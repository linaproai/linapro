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

interface PublicFrontendCronSettings {
  shell: PublicFrontendCronShellSettings;
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
    pageDesc: '',
    pageTitle: '',
  },
  cron: {
    shell: {
      disabledReason: '',
      enabled: false,
      supported: true,
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

function resolvePublicFrontendEndpoint(): string {
  const { apiURL } = useAppConfig(import.meta.env, import.meta.env.PROD);
  return `${apiURL.replace(/\/$/, '')}/config/public/frontend`;
}

function normalizePublicFrontendSettings(payload: any): PublicFrontendSettings {
  const app = payload?.app ?? {};
  const auth = payload?.auth ?? {};
  const cron = payload?.cron ?? {};
  const shell = cron?.shell ?? {};
  const ui = payload?.ui ?? {};

  return {
    app: {
      logo: normalizeString(app.logo),
      logoDark: normalizeString(app.logoDark),
      name: normalizeString(app.name),
    },
    auth: {
      loginSubtitle: normalizeString(auth.loginSubtitle),
      pageDesc: normalizeString(auth.pageDesc),
      pageTitle: normalizeString(auth.pageTitle),
    },
    cron: {
      shell: {
        disabledReason: normalizeString(shell.disabledReason),
        enabled: normalizeBoolean(shell.enabled),
        supported:
          shell?.supported === undefined
            ? true
            : normalizeBoolean(shell.supported),
      },
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
    Object.assign(publicFrontendState.cron.shell, settings.cron.shell);
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
