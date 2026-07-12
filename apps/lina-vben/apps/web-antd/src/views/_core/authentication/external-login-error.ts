/**
 * Maps external-login redirect `message` query values (stable bizerr codes)
 * to host authentication i18n keys. Protocol plugins only put machine codes
 * on the redirect URL for safety; the SPA must never surface those codes as
 * the primary user-facing description.
 */

/** Derives the same runtime i18n key as lina-core pkg/bizerr.MessageKey. */
export function businessErrorMessageKey(errorCode: string): string {
  const segments = errorCode
    .trim()
    .toLowerCase()
    .split(/[_\-.\s]+/)
    .filter(Boolean);
  if (segments.length === 0) {
    return '';
  }
  return `error.${segments.join('.')}`;
}

const CONFIG_MISSING_CODES = new Set([
  'PLUGIN_OIDC_GOOGLE_CONFIG_MISSING',
  'PLUGIN_OIDC_DISCORD_CONFIG_MISSING',
  'PLUGIN_OIDC_GENERIC_CONFIG_MISSING',
]);

const DISCOVERY_FAILED_CODES = new Set([
  'PLUGIN_OIDC_GENERIC_DISCOVERY_FAILED',
]);

export type ExternalLoginErrorResolveOptions = {
  /**
   * Looks up a runtime i18n key (host or plugin messages). Return the
   * translated string, or the key itself / empty when missing.
   */
  translate: (key: string) => string;
  /** Host fallback when no specific mapping or plugin translation exists. */
  fallbackLoginFailed: string;
  configMissing: string;
  discoveryFailed: string;
  externalLoginFailed: string;
};

/**
 * Resolves a redirect `message` into a human-readable description for the
 * login-page notification.
 */
export function resolveExternalLoginErrorMessage(
  message: string,
  options: ExternalLoginErrorResolveOptions,
): string {
  const normalized = message.trim();
  if (!normalized) {
    return options.fallbackLoginFailed;
  }
  if (CONFIG_MISSING_CODES.has(normalized)) {
    return options.configMissing;
  }
  if (DISCOVERY_FAILED_CODES.has(normalized)) {
    return options.discoveryFailed;
  }

  const messageKey = businessErrorMessageKey(normalized);
  if (messageKey) {
    const localized = options.translate(messageKey);
    if (localized && localized !== messageKey) {
      return localized;
    }
  }

  // Machine-style codes (PLUGIN_*, EXTERNAL_LOGIN_FAILED, …) must not be shown raw.
  if (/^[A-Z][A-Z0-9_]*$/.test(normalized)) {
    return options.externalLoginFailed;
  }

  return normalized;
}
