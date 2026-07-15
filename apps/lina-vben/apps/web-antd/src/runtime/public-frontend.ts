import type { AuthPageLayoutType } from '@vben/types';

import { reactive, readonly } from 'vue';

import { useAppConfig } from '@vben/hooks';
import {
  preferences,
  preferencesManager,
  updatePreferences,
} from '@vben/preferences';

interface PublicFrontendAppSettings {
  logo: string;
  logoDark: string;
  name: string;
}

interface PublicFrontendAuthSettings {
  forgetPasswordEnabled: boolean;
  loginSubtitle: string;
  panelLayout: AuthPageLayoutType;
  pageDesc: string;
  pageTitle: string;
  privacyPolicy: string;
  registerEnabled: boolean;
  sloganImage: string;
  termsOfService: string;
}

interface PublicFrontendUserSettings {
  defaultAvatar: string;
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

interface PublicFrontendWorkspaceSettings {
  basePath: string;
}

interface PublicFrontendSettings {
  app: PublicFrontendAppSettings;
  auth: PublicFrontendAuthSettings;
  cron: PublicFrontendCronSettings;
  user: PublicFrontendUserSettings;
  ui: PublicFrontendUISettings;
  workspace: PublicFrontendWorkspaceSettings;
}

const publicFrontendFetchInit: RequestInit = {
  // Public frontend settings are managed by sys_config. Force each sync to bypass
  // the browser HTTP cache so the same browser immediately sees the latest theme
  // and branding values after backend updates.
  cache: 'no-store',
  credentials: 'same-origin',
  headers: {},
  method: 'GET',
};

const publicFrontendState = reactive<PublicFrontendSettings>({
  app: {
    logo: '',
    logoDark: '',
    name: '',
  },
  auth: {
    forgetPasswordEnabled: true,
    loginSubtitle: '',
    panelLayout: 'panel-center',
    pageDesc: '',
    pageTitle: '',
    privacyPolicy: '',
    registerEnabled: true,
    sloganImage: '',
    termsOfService: '',
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
  user: {
    defaultAvatar: '',
  },
  ui: {
    layout: '',
    themeMode: '',
    watermarkContent: '',
    watermarkEnabled: false,
  },
  workspace: {
    basePath: '/admin',
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

function normalizeWorkspaceBasePath(value: unknown): string {
  const cleaned = normalizeString(value)
    .replaceAll('\\', '/')
    .replace(/\/+/g, '/');
  if (cleaned === '/') {
    return '/';
  }
  const normalized = cleaned.replace(/\/+$/, '');
  if (
    !normalized ||
    normalized === '/' ||
    normalized.includes('*') ||
    normalized.includes('?') ||
    normalized.includes('#') ||
    normalized.includes('://') ||
    !normalized.startsWith('/')
  ) {
    return '/admin';
  }
  if (normalized === '/') {
    return '/';
  }

  const reservedPrefixes = ['/api', '/api/v1', '/x', '/x-assets'];
  if (
    reservedPrefixes.some(
      (prefix) => normalized === prefix || normalized.startsWith(`${prefix}/`),
    )
  ) {
    return '/admin';
  }
  return normalized;
}

function resolveWorkspaceRouterBase() {
  const basePath = normalizeWorkspaceBasePath(publicFrontendState.workspace.basePath);
  return basePath === '/' ? '/' : `${basePath}/`;
}

function splitURLSuffix(value: string): [string, string] {
  const queryIndex = value.indexOf('?');
  const hashIndex = value.indexOf('#');
  const suffixIndex = [queryIndex, hashIndex]
    .filter((index) => index >= 0)
    .sort((left, right) => left - right)[0];
  if (suffixIndex === undefined) {
    return [value, ''];
  }
  return [value.slice(0, suffixIndex), value.slice(suffixIndex)];
}

function isAbsoluteOrSpecialURL(value: string): boolean {
  return (
    value.startsWith('//') ||
    value.startsWith('#') ||
    value.startsWith('?') ||
    /^[a-z][a-z\d+.-]*:/i.test(value)
  );
}

function resolveWorkspaceAssetURL(value: unknown): string {
  const cleaned = normalizeString(value);
  if (!cleaned || isAbsoluteOrSpecialURL(cleaned)) {
    return cleaned;
  }

  const [rawPath, suffix] = splitURLSuffix(cleaned);
  const normalizedPath = rawPath
    .replaceAll('\\', '/')
    .replace(/^\.\//, '')
    .replace(/\/+/g, '/');
  if (!normalizedPath || normalizedPath === '/') {
    return cleaned;
  }

  const rootedPath = normalizedPath.startsWith('/')
    ? normalizedPath
    : `/${normalizedPath}`;
  if (
    rootedPath === '/api.json' ||
    ['/api', '/x', '/x-assets'].some(
      (prefix) => rootedPath === prefix || rootedPath.startsWith(`${prefix}/`),
    )
  ) {
    return `${rootedPath}${suffix}`;
  }

  const workspaceBasePath = normalizeWorkspaceBasePath(
    publicFrontendState.workspace.basePath,
  );
  if (
    workspaceBasePath === '/' ||
    rootedPath === workspaceBasePath ||
    rootedPath.startsWith(`${workspaceBasePath}/`)
  ) {
    return `${rootedPath}${suffix}`;
  }

  return `${workspaceBasePath}${rootedPath}${suffix}`;
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
  const user = payload?.user ?? {};
  const ui = payload?.ui ?? {};
  const workspace = payload?.workspace ?? {};

  return {
    app: {
      logo: normalizeString(app.logo),
      logoDark: normalizeString(app.logoDark),
      name: normalizeString(app.name),
    },
    auth: {
      forgetPasswordEnabled:
        auth?.forgetPasswordEnabled === undefined
          ? true
          : normalizeBoolean(auth.forgetPasswordEnabled),
      loginSubtitle: normalizeString(auth.loginSubtitle),
      panelLayout: normalizeAuthPanelLayout(auth.panelLayout),
      pageDesc: normalizeString(auth.pageDesc),
      pageTitle: normalizeString(auth.pageTitle),
      privacyPolicy: normalizeString(auth.privacyPolicy),
      registerEnabled:
        auth?.registerEnabled === undefined
          ? true
          : normalizeBoolean(auth.registerEnabled),
      sloganImage: normalizeString(auth.sloganImage),
      termsOfService: normalizeString(auth.termsOfService),
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
    user: {
      defaultAvatar: normalizeString(user.defaultAvatar),
    },
    ui: {
      layout: normalizeString(ui.layout),
      themeMode: normalizeString(ui.themeMode),
      watermarkContent: normalizeString(ui.watermarkContent),
      watermarkEnabled: normalizeBoolean(ui.watermarkEnabled),
    },
    workspace: {
      basePath: normalizeWorkspaceBasePath(workspace.basePath),
    },
  };
}

function applyPublicFrontendPreferences(settings: PublicFrontendSettings) {
  const initial = preferencesManager.getInitialPreferences();
  const logoSource = resolveWorkspaceAssetURL(
    settings.app.logo || initial.logo.source,
  );
  const logoSourceDark =
    resolveWorkspaceAssetURL(settings.app.logoDark || initial.logo.sourceDark) ||
    logoSource;
  const themePreference = {
    builtinType: initial.theme.builtinType,
    colorPrimary: initial.theme.colorPrimary,
    ...(!preferencesManager.hasUserThemePreference()
      ? { mode: (settings.ui.themeMode || initial.theme.mode) as any }
      : {}),
  };

  updatePreferences(
    {
      app: {
        authPageLayout: settings.auth.panelLayout,
        defaultAvatar: resolveWorkspaceAssetURL(
          settings.user.defaultAvatar || initial.app.defaultAvatar,
        ),
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
      theme: themePreference,
    },
    { markUserThemePreference: false },
  );
}

async function syncPublicFrontendSettings(locale?: string) {
  try {
    const activeLocale = locale || preferences.app.locale;
    const requestInit: RequestInit = {
      ...publicFrontendFetchInit,
      headers: {
        'Accept-Language': activeLocale,
      },
    };
    const response = await fetch(resolvePublicFrontendEndpoint(), requestInit);
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
    Object.assign(publicFrontendState.user, settings.user);
    Object.assign(publicFrontendState.ui, settings.ui);
    Object.assign(publicFrontendState.workspace, settings.workspace);
    applyPublicFrontendPreferences(settings);

    return settings;
  } catch {
    return null;
  }
}

export {
  normalizeWorkspaceBasePath,
  resolveWorkspaceAssetURL,
  resolveWorkspaceRouterBase,
  syncPublicFrontendSettings,
};
export const publicFrontendSettings = readonly(publicFrontendState);
export type { PublicFrontendSettings, PublicFrontendWorkspaceSettings };
