import type { TenantAwareLoginResult } from '#/api/tenant/model';

import { pluginApiPath, requestClient } from '#/api/request';

/**
 * Stable owner plugin id for the external-identity domain (managed plugin).
 * Must match apps/lina-plugins/linapro-extlogin-core/plugin.yaml id.
 */
export const EXTLOGIN_CORE_PLUGIN_ID = 'linapro-extlogin-core';

/**
 * Public relative path under the plugin API prefix for SPA handoff exchange.
 * Must match backend/api/handoff/v1 path:"/handoff/exchange".
 */
export const EXTLOGIN_HANDOFF_EXCHANGE_PATH = 'handoff/exchange';

/**
 * Exchange a one-time external-login handoff code for session tokens.
 * Owned by managed plugin linapro-extlogin-core (not host /auth).
 */
export async function exchangeExternalLoginHandoffApi(handoff: string) {
  return requestClient.post<TenantAwareLoginResult>(
    pluginApiPath(EXTLOGIN_CORE_PLUGIN_ID, EXTLOGIN_HANDOFF_EXCHANGE_PATH),
    { handoff },
  );
}
