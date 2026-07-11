import type { TenantAwareLoginResult } from '#/api/tenant/model';

import { pluginApiPath, requestClient } from '#/api/request';

/**
 * Stable owner plugin id for the external-identity domain (managed plugin).
 * Must match apps/lina-plugins/linapro-extid-core/plugin.yaml id.
 */
export const EXTID_CORE_PLUGIN_ID = 'linapro-extid-core';

/**
 * Public relative path under the plugin API prefix for SPA handoff exchange.
 * Must match backend/api/handoff/v1 path:"/handoff/exchange".
 */
export const EXTID_HANDOFF_EXCHANGE_PATH = 'handoff/exchange';

/**
 * Exchange a one-time external-login handoff code for session tokens.
 * Owned by managed plugin linapro-extid-core (not host /auth).
 */
export async function exchangeExternalLoginHandoffApi(handoff: string) {
  return requestClient.post<TenantAwareLoginResult>(
    pluginApiPath(EXTID_CORE_PLUGIN_ID, EXTID_HANDOFF_EXCHANGE_PATH),
    { handoff },
  );
}
