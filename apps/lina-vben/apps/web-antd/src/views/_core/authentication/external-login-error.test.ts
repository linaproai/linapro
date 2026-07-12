import { describe, expect, it } from 'vitest';

import {
  businessErrorMessageKey,
  resolveExternalLoginErrorMessage,
} from './external-login-error';

const options = {
  configMissing: 'config-missing-text',
  discoveryFailed: 'discovery-failed-text',
  externalLoginFailed: 'external-login-failed-text',
  fallbackLoginFailed: 'login-failed-text',
  translate: (key: string) => {
    if (key === 'error.plugin.oidc.generic.identity.verify.failed') {
      return 'identity-verify-text';
    }
    return key;
  },
};

describe('businessErrorMessageKey', () => {
  it('matches bizerr MessageKey derivation for plugin codes', () => {
    expect(businessErrorMessageKey('PLUGIN_OIDC_GENERIC_DISCOVERY_FAILED')).toBe(
      'error.plugin.oidc.generic.discovery.failed',
    );
  });
});

describe('resolveExternalLoginErrorMessage', () => {
  it('returns fallback for empty message', () => {
    expect(resolveExternalLoginErrorMessage('  ', options)).toBe(
      'login-failed-text',
    );
  });

  it('maps config-missing codes to the host config message', () => {
    expect(
      resolveExternalLoginErrorMessage(
        'PLUGIN_OIDC_GENERIC_CONFIG_MISSING',
        options,
      ),
    ).toBe('config-missing-text');
    expect(
      resolveExternalLoginErrorMessage(
        'PLUGIN_OIDC_GOOGLE_CONFIG_MISSING',
        options,
      ),
    ).toBe('config-missing-text');
  });

  it('maps discovery-failed to a clear host message instead of the raw code', () => {
    expect(
      resolveExternalLoginErrorMessage(
        'PLUGIN_OIDC_GENERIC_DISCOVERY_FAILED',
        options,
      ),
    ).toBe('discovery-failed-text');
  });

  it('prefers runtime plugin i18n when available', () => {
    expect(
      resolveExternalLoginErrorMessage(
        'PLUGIN_OIDC_GENERIC_IDENTITY_VERIFY_FAILED',
        options,
      ),
    ).toBe('identity-verify-text');
  });

  it('never surfaces raw PLUGIN machine codes to the user', () => {
    expect(
      resolveExternalLoginErrorMessage(
        'PLUGIN_OIDC_GENERIC_EXTERNAL_LOGIN_FAILED',
        options,
      ),
    ).toBe('external-login-failed-text');
  });

  it('passes through non-code free-text when present', () => {
    expect(
      resolveExternalLoginErrorMessage('Something went wrong at IdP', options),
    ).toBe('Something went wrong at IdP');
  });
});
