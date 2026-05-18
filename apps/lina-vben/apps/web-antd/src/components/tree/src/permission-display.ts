import { $t, $te } from '@vben/locales';
import { preferences } from '@vben/preferences';

const dynamicRoutePermissionSourcePrefix = 'Dynamic Route Permission:';
const permissionDisplayI18nKeyPrefix = 'pages.tree.permissionDisplay';

function getActiveLocale() {
  if (typeof document !== 'undefined' && document.documentElement.lang) {
    return document.documentElement.lang;
  }
  return preferences.app.locale;
}

function isEnglishLocale() {
  return getActiveLocale().startsWith('en');
}

function toTitleCase(rawValue: string) {
  if (!rawValue) {
    return '';
  }

  return rawValue
    .split(/\s+/)
    .filter(Boolean)
    .map((token) => token.slice(0, 1).toUpperCase() + token.slice(1))
    .join(' ');
}

function translateWithFallback(
  key: string,
  fallback: string,
  values?: Record<string, string>,
) {
  if (!$te(key)) {
    return fallback;
  }
  const translated = values ? $t(key, values) : $t(key);
  return translated && translated !== key ? translated : fallback;
}

function humanizePermissionSegment(rawValue: string) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  const tokens = normalized.split(/[-_/]+/).filter(Boolean);
  if (tokens.length === 0) {
    return normalized;
  }

  const transformed = tokens.map((token) => {
    const lowerToken = token.toLowerCase();
    const fallback = isEnglishLocale() ? toTitleCase(token) : token;
    return translateWithFallback(
      `${permissionDisplayI18nKeyPrefix}.segments.${lowerToken}`,
      fallback,
    );
  });

  return isEnglishLocale() ? transformed.join(' ') : transformed.join('');
}

function extractDynamicRoutePermission(rawValue: string) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  if (normalized.startsWith(dynamicRoutePermissionSourcePrefix)) {
    return normalized.slice(dynamicRoutePermissionSourcePrefix.length).trim();
  }

  const parts = normalized.split(':');
  const pluginSegment = parts[0] ?? '';
  const resourceSegment = parts[1] ?? '';
  const actionSegment = parts[2] ?? '';
  if (
    parts.length === 3 &&
    pluginSegment.trim() !== '' &&
    resourceSegment.trim() !== '' &&
    actionSegment.trim() !== '' &&
    /^plugin[-_]/.test(pluginSegment.trim())
  ) {
    return normalized;
  }

  return '';
}

export function formatMenuPermissionLabel(rawValue: null | string | undefined) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  const permission = extractDynamicRoutePermission(normalized);
  if (!permission) {
    return normalized;
  }

  const parts = permission.split(':');
  if (parts.length !== 3) {
    return translateWithFallback(
      `${permissionDisplayI18nKeyPrefix}.dynamicRoutePermission`,
      normalized,
    );
  }

  const resourceLabel = humanizePermissionSegment(parts[1] ?? '');
  const actionLabel = humanizePermissionSegment(parts[2] ?? '');

  return translateWithFallback(
    `${permissionDisplayI18nKeyPrefix}.dynamicRoutePermissionLabel`,
    permission,
    {
      action: actionLabel,
      resource: resourceLabel,
    },
  );
}

export function formatMenuPermissionShortLabel(
  rawValue: null | string | undefined,
) {
  const normalized = String(rawValue || '').trim();
  if (!normalized) {
    return '';
  }

  const permission = extractDynamicRoutePermission(normalized);
  if (!permission) {
    return normalized;
  }

  const parts = permission.split(':');
  if (parts.length !== 3) {
    return translateWithFallback(
      `${permissionDisplayI18nKeyPrefix}.dynamicRoutePermission`,
      normalized,
    );
  }

  const resourceLabel = humanizePermissionSegment(parts[1] ?? '');
  const actionLabel = humanizePermissionSegment(parts[2] ?? '');
  const labels = [resourceLabel, actionLabel].filter(Boolean);

  return isEnglishLocale() ? labels.join(' ') : labels.join('');
}
